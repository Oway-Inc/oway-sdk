package oway

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	mrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// DefaultBaseURL is used when Config.BaseURL is empty. SDK consumers must
// opt in to a non-sandbox environment by setting BaseURL explicitly (use
// EnvironmentProduction or EnvironmentSandbox).
const DefaultBaseURL = EnvironmentSandbox

// DefaultTimeout bounds every HTTP request the SDK makes when the caller
// does not pass a deadline via context.
const DefaultTimeout = 30 * time.Second

// DefaultMaxRetries is the number of retry attempts (in addition to the
// initial attempt) applied to transient failures.
const DefaultMaxRetries = 3

// Config controls a Client. All fields except ClientID/ClientSecret are
// optional and have sensible defaults.
type Config struct {
	// ClientID and ClientSecret authenticate the SDK with the Oway token
	// endpoint. Both are required.
	ClientID     string
	ClientSecret string

	// APIKey is the default company API key sent on every request as
	// `x-oway-api-key`. For multi-tenant integrations, leave this empty
	// and pass a per-request key via WithCompanyAPIKey.
	APIKey string

	// BaseURL is the API root. Defaults to EnvironmentSandbox.
	BaseURL string

	// TokenURL overrides the M2M token endpoint. Defaults to
	// BaseURL + "/v1/auth/token".
	TokenURL string

	// HTTPClient is the underlying http.Client. When nil, the SDK creates
	// one with DefaultTimeout. A caller-supplied client's Timeout is
	// preserved; the SDK only wraps the Transport.
	HTTPClient *http.Client

	// MaxRetries bounds the number of retry attempts on transient errors.
	// Zero means use DefaultMaxRetries; pass -1 to disable retries.
	MaxRetries int

	// Logger receives structured debug/info/warn/error events. When nil,
	// the SDK is silent.
	Logger *slog.Logger
}

// Client is the high-level Oway SDK. Construct one with New and reuse it for
// the lifetime of your process. All methods are safe for concurrent use.
type Client struct {
	cfg     Config
	api     *client.ClientWithResponses
	rawHTTP *http.Client

	tokenMu     sync.Mutex
	token       string
	tokenExpiry time.Time
}

// New creates a Client. It returns an error if required credentials are
// missing or the underlying generated client cannot be constructed.
func New(cfg Config) (*Client, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("oway: ClientID and ClientSecret are required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = cfg.BaseURL + "/v1/auth/token"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: DefaultTimeout}
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = DefaultMaxRetries
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}

	c := &Client{cfg: cfg, rawHTTP: cfg.HTTPClient}

	authedHTTP := &http.Client{
		Timeout: cfg.HTTPClient.Timeout,
		Transport: &authedTransport{
			client: c,
			next:   cfg.HTTPClient.Transport,
		},
	}

	api, err := client.NewClientWithResponses(cfg.BaseURL, client.WithHTTPClient(authedHTTP))
	if err != nil {
		return nil, fmt.Errorf("oway: build generated client: %w", err)
	}
	c.api = api

	c.log("debug", "sdk initialized", "baseURL", cfg.BaseURL, "hasDefaultAPIKey", cfg.APIKey != "")
	return c, nil
}

// GeneratedClient returns the underlying oapi-codegen client for use cases
// not covered by the high-level wrapper. Most callers should not need this.
func (c *Client) GeneratedClient() *client.ClientWithResponses { return c.api }

// companyAPIKeyContextKey carries a per-request company API key through
// context values.
type companyAPIKeyContextKey struct{}

// WithCompanyAPIKey returns a child context that carries a per-request
// company API key. Use this for multi-tenant integrations where a single
// SDK Client serves several companies.
func WithCompanyAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, companyAPIKeyContextKey{}, apiKey)
}

func companyAPIKeyFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(companyAPIKeyContextKey{}).(string); ok {
		return v
	}
	return ""
}

// authedTransport adds the Bearer token, API key, and request ID to every
// outbound request. Token refresh is lazy and serialized via Client.tokenMu.
type authedTransport struct {
	client *Client
	next   http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.client.accessToken(req.Context())
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)

	if key := companyAPIKeyFromContext(req.Context()); key != "" {
		req.Header.Set("x-oway-api-key", key)
	} else if t.client.cfg.APIKey != "" {
		req.Header.Set("x-oway-api-key", t.client.cfg.APIKey)
	}

	if req.Header.Get("x-request-id") == "" {
		req.Header.Set("x-request-id", newRequestID())
	}

	next := t.next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}

// newRequestID returns a random 128-bit hex string. crypto/rand is used so
// concurrent requests cannot collide on the same id.
func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

// accessToken returns a valid bearer token, refreshing if the cached token
// is within five minutes of expiring.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

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
	reqBody, err := json.Marshal(map[string]string{
		"clientId":     c.cfg.ClientID,
		"clientSecret": c.cfg.ClientSecret,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("oway: marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("oway: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-request-id", newRequestID())

	resp, err := c.rawHTTP.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("oway: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", time.Time{}, fmt.Errorf("oway: read token response: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, parseHTTPError(resp.StatusCode, resp.Header.Get("x-request-id"), body)
	}

	var tokenResp struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("oway: decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" || tokenResp.ExpiresIn <= 0 {
		return "", time.Time{}, errors.New("oway: token response missing accessToken or expiresIn")
	}

	return tokenResp.AccessToken, time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second), nil
}

// retry runs op until it returns nil, exhausts MaxRetries, or hits a
// non-retryable error. Wait time uses full-jitter exponential backoff
// capped at 30s and respects ctx cancellation.
func (c *Client) retry(ctx context.Context, op func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		err := op()
		if err == nil {
			return nil
		}
		lastErr = err

		oe, ok := AsError(err)
		if !ok || !oe.IsRetryable() {
			return err
		}
		if attempt == c.cfg.MaxRetries {
			break
		}

		delay := backoffDelay(attempt)
		c.log("warn", "retrying", "attempt", attempt+1, "delay", delay, "status", oe.StatusCode)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return lastErr
}

// backoffDelay returns a randomized exponential wait time bounded at 30s.
// Full-jitter (random across [0, base)) spreads retries from clients that
// hit a 429 storm at the same wall-clock instant. The lower bound guards
// against overflow from very large attempt counts.
func backoffDelay(attempt int) time.Duration {
	base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if base <= 0 {
		base = time.Second
	} else if base > 30*time.Second {
		base = 30 * time.Second
	}
	return time.Duration(mrand.Int63n(int64(base) + 1))
}

func (c *Client) log(level, msg string, args ...any) {
	if c.cfg.Logger == nil {
		return
	}
	switch level {
	case "debug":
		c.cfg.Logger.Debug(msg, args...)
	case "info":
		c.cfg.Logger.Info(msg, args...)
	case "warn":
		c.cfg.Logger.Warn(msg, args...)
	case "error":
		c.cfg.Logger.Error(msg, args...)
	}
}

// decode is the chokepoint that converts a generated response into either a
// typed value or a typed *Error. It is generic on the success payload so
// callers do not have to repeat boilerplate.
func decode[T any](status int, body []byte, httpResp *http.Response, json200 *T) (*T, error) {
	if status != http.StatusOK {
		requestID := ""
		if httpResp != nil {
			requestID = httpResp.Header.Get("x-request-id")
			if requestID == "" {
				requestID = httpResp.Header.Get("X-Request-Id")
			}
		}
		return nil, parseHTTPError(status, requestID, body)
	}
	if json200 == nil {
		return nil, &Error{StatusCode: status, Message: "empty 200 response body"}
	}
	return json200, nil
}

// RequestQuote requests a shipping quote. Use WithCompanyAPIKey on the
// context to override the default API key for multi-tenant calls.
func (c *Client) RequestQuote(ctx context.Context, req *QuoteRequest) (*Quote, error) {
	if req == nil {
		return nil, errors.New("oway: QuoteRequest is required")
	}
	body := req.toClient()
	var out *Quote
	err := c.retry(ctx, func() error {
		r, err := c.api.RequestQuoteWithResponse(ctx, client.RequestQuoteJSONRequestBody(body))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// CreateShipment creates a shipment.
func (c *Client) CreateShipment(ctx context.Context, req *ShipmentRequest) (*Shipment, error) {
	if req == nil {
		return nil, errors.New("oway: ShipmentRequest is required")
	}
	body := req.toClient()
	var out *Shipment
	err := c.retry(ctx, func() error {
		// nil params: the generated client now accepts an optional Idempotency-Key
		// header on this endpoint. The SDK doesn't surface idempotency yet; callers
		// who need it can drop down to GeneratedClient() until we add a wrapper option.
		r, err := c.api.CreateShipmentWithResponse(ctx, nil, client.CreateShipmentJSONRequestBody(body))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// ConfirmShipment confirms a previously created shipment.
func (c *Client) ConfirmShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	var out *Shipment
	err := c.retry(ctx, func() error {
		r, err := c.api.ConfirmShipmentWithResponse(ctx, orderNumber)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// TrackShipment returns tracking information for a shipment. The live
// position estimate (GPS center, uncertainty radius, last-event time,
// delay flags) is included by default. Use TrackShipmentStatusOnly for
// lightweight status-only polling that omits the position computation.
func (c *Client) TrackShipment(ctx context.Context, orderNumber string) (*Tracking, error) {
	return c.trackShipment(ctx, orderNumber, true)
}

// TrackShipmentStatusOnly returns tracking information without the live
// position estimate, for lightweight status polling.
func (c *Client) TrackShipmentStatusOnly(ctx context.Context, orderNumber string) (*Tracking, error) {
	return c.trackShipment(ctx, orderNumber, false)
}

func (c *Client) trackShipment(ctx context.Context, orderNumber string, includeLocation bool) (*Tracking, error) {
	var params *client.TrackShipmentParams
	if includeLocation {
		location := "location"
		params = &client.TrackShipmentParams{Include: &location}
	}
	var out *Tracking
	err := c.retry(ctx, func() error {
		r, err := c.api.TrackShipmentWithResponse(ctx, orderNumber, params)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// GetShipment retrieves a shipment by order number.
func (c *Client) GetShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	var out *Shipment
	err := c.retry(ctx, func() error {
		r, err := c.api.GetShipmentWithResponse(ctx, orderNumber)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// CancelShipment cancels a shipment.
func (c *Client) CancelShipment(ctx context.Context, orderNumber string) (*Shipment, error) {
	var out *Shipment
	err := c.retry(ctx, func() error {
		r, err := c.api.CancelShipmentWithResponse(ctx, orderNumber)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// GetInvoice retrieves the invoice for a delivered shipment.
func (c *Client) GetInvoice(ctx context.Context, orderNumber string) (*Invoice, error) {
	var out *Invoice
	err := c.retry(ctx, func() error {
		r, err := c.api.GetInvoiceWithResponse(ctx, orderNumber)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// GetQuote retrieves a quote by ID.
func (c *Client) GetQuote(ctx context.Context, quoteID string) (*Quote, error) {
	var out *Quote
	err := c.retry(ctx, func() error {
		r, err := c.api.GetQuoteWithResponse(ctx, quoteID)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// GetDocument retrieves a document (BOL, invoice, label, POD) for a shipment.
func (c *Client) GetDocument(ctx context.Context, orderNumber string, documentType DocumentType) (*Document, error) {
	var out *Document
	err := c.retry(ctx, func() error {
		r, err := c.api.GetDocumentWithResponse(ctx, orderNumber, client.GetDocumentParamsDocumentType(documentType))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}
