package oway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// Config holds configuration for the Oway client
type Config struct {
	// ClientID for M2M authentication (REQUIRED - from Sales Engineering)
	ClientID string

	// ClientSecret for M2M authentication (REQUIRED - from Sales Engineering)
	ClientSecret string

	// APIKey is the default company API key (optional)
	// Single-company: Set this to your company's key
	// Multi-company: Provide per-request
	APIKey string

	// BaseURL is the Oway API base URL
	BaseURL string

	// TokenURL is the M2M token endpoint
	TokenURL string

	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client

	// Debug enables debug logging
	Debug bool
}

// Client is the main Oway SDK client
type Client struct {
	config      Config
	client      *client.ClientWithResponses
	token       string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// New creates a new Oway client
func New(config Config) (*Client, error) {
	// M2M credentials are REQUIRED
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("clientId and clientSecret are required (contact Oway Sales Engineering)")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://rest-api.sandbox.oway.io"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://rest-api.sandbox.oway.io/v1/auth/token"
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	c := &Client{config: config}

	authHTTPClient := &http.Client{
		Timeout: config.HTTPClient.Timeout,
		Transport: &authenticatedTransport{
			client:    c,
			transport: config.HTTPClient.Transport,
		},
	}

	oapiClient, err := client.NewClientWithResponses(config.BaseURL, client.WithHTTPClient(authHTTPClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	c.client = oapiClient

	if config.Debug {
		fmt.Printf("[Oway] SDK initialized (M2M, default API key: %v)\n", config.APIKey != "")
	}

	return c, nil
}

// GetClient returns the underlying oapi-codegen client
func (c *Client) GetClient() *client.ClientWithResponses {
	return c.client
}

type authenticatedTransport struct {
	client    *Client
	transport http.RoundTripper
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.client.getAccessToken(req.Context())
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)

	// Add company API key if present in request context or default
	apiKey := req.Context().Value(companyAPIKeyContextKey{})
	if apiKey == nil {
		apiKey = t.client.config.APIKey
	}
	if apiKey != nil && apiKey != "" {
		req.Header.Set("x-oway-api-key", apiKey.(string))
	}

	req.Header.Set("x-request-id", fmt.Sprintf("%d", time.Now().UnixNano()))

	transport := t.transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(req)
}

// companyAPIKeyContextKey is used to pass per-request API keys via context
type companyAPIKeyContextKey struct{}

// WithCompanyAPIKey returns a context with the specified company API key
func WithCompanyAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, companyAPIKeyContextKey{}, apiKey)
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.tokenMutex.RLock()
	if c.token != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry) {
		token := c.token
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.token != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry) {
		return c.token, nil
	}

	token, expiry, err := c.refreshToken(ctx)
	if err != nil {
		return "", err
	}

	c.token = token
	c.tokenExpiry = expiry
	return token, nil
}

func (c *Client) refreshToken(ctx context.Context) (string, time.Time, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"clientId":     c.config.ClientID,
		"clientSecret": c.config.ClientSecret,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.TokenURL, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(reqBody))

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("M2M token request failed: %d %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	json.NewDecoder(resp.Body).Decode(&tokenResp)

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiry, nil
}

// RequestQuote requests a shipping quote
func (c *Client) RequestQuote(ctx context.Context, req *QuoteRequest) (*Quote, error) {
	res, err := c.client.RequestQuoteWithResponse(ctx, client.RequestQuoteJSONRequestBody(*req))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request quote failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// RequestQuoteForCompany requests a quote for a specific company
func (c *Client) RequestQuoteForCompany(ctx context.Context, req *QuoteRequest, companyAPIKey string) (*Quote, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.RequestQuote(ctx, req)
}

// CreateShipment creates a shipment
func (c *Client) CreateShipment(ctx context.Context, req *ShipmentRequest) (*Shipment, error) {
	res, err := c.client.CreateShipmentWithResponse(ctx, client.CreateShipmentJSONRequestBody(*req))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create shipment failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// CreateShipmentForCompany creates a shipment for a specific company
func (c *Client) CreateShipmentForCompany(ctx context.Context, req *ShipmentRequest, companyAPIKey string) (*Shipment, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.CreateShipment(ctx, req)
}

// ConfirmShipment confirms a shipment by order number
func (c *Client) ConfirmShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	res, err := c.client.ConfirmShipmentByOrderNumberWithResponse(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("confirm shipment failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// ConfirmShipmentForCompany confirms a shipment for a specific company
func (c *Client) ConfirmShipmentForCompany(ctx context.Context, orderNumber string, companyAPIKey string) (*Shipment, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.ConfirmShipment(ctx, orderNumber)
}

// TrackShipment gets tracking information for a shipment
func (c *Client) TrackShipment(ctx context.Context, orderNumber string) (*Tracking, error) {
	res, err := c.client.TrackShipmentByOrderNumberWithResponse(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("track shipment failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// TrackShipmentForCompany gets tracking information for a specific company's shipment
func (c *Client) TrackShipmentForCompany(ctx context.Context, orderNumber string, companyAPIKey string) (*Tracking, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.TrackShipment(ctx, orderNumber)
}

// GetInvoice retrieves the invoice for a delivered shipment
func (c *Client) GetInvoice(ctx context.Context, orderNumber string) (*Invoice, error) {
	res, err := c.client.GetInvoiceWithResponse(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get invoice failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// GetInvoiceForCompany retrieves the invoice for a specific company's shipment
func (c *Client) GetInvoiceForCompany(ctx context.Context, orderNumber string, companyAPIKey string) (*Invoice, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.GetInvoice(ctx, orderNumber)
}

// GetShipment retrieves a shipment by order number
func (c *Client) GetShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	res, err := c.client.GetShipmentByOrderNumberWithResponse(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get shipment failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// GetShipmentForCompany retrieves a shipment for a specific company
func (c *Client) GetShipmentForCompany(ctx context.Context, orderNumber string, companyAPIKey string) (*Shipment, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.GetShipment(ctx, orderNumber)
}

// CancelShipment cancels a shipment by order number
func (c *Client) CancelShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	res, err := c.client.CancelShipmentByOrderNumberWithResponse(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("cancel shipment failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// CancelShipmentForCompany cancels a shipment for a specific company
func (c *Client) CancelShipmentForCompany(ctx context.Context, orderNumber string, companyAPIKey string) (*Shipment, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.CancelShipment(ctx, orderNumber)
}

// GetQuoteByID retrieves a quote by its ID
func (c *Client) GetQuoteByID(ctx context.Context, quoteID string) (*Quote, error) {
	res, err := c.client.GetQuoteWithResponse(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get quote failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// GetQuoteByIDForCompany retrieves a quote for a specific company
func (c *Client) GetQuoteByIDForCompany(ctx context.Context, quoteID string, companyAPIKey string) (*Quote, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.GetQuoteByID(ctx, quoteID)
}

// GetDocument retrieves a document for a shipment by order number and document type
func (c *Client) GetDocument(ctx context.Context, orderNumber string, documentType DocumentType) (*Document, error) {
	res, err := c.client.GetDocumentByOrderNumberWithResponse(ctx, orderNumber, client.GetDocumentByOrderNumberParamsDocumentType(documentType))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get document failed: status %d", res.StatusCode())
	}
	if res.HALJSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response body")
	}
	return res.HALJSON200, nil
}

// GetDocumentForCompany retrieves a document for a specific company's shipment
func (c *Client) GetDocumentForCompany(ctx context.Context, orderNumber string, documentType DocumentType, companyAPIKey string) (*Document, error) {
	ctx = WithCompanyAPIKey(ctx, companyAPIKey)
	return c.GetDocument(ctx, orderNumber, documentType)
}
