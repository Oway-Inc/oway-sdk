package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	oway "github.com/Oway-Inc/oway-sdk/packages/go"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type runState struct {
	OfferID    string                         `json:"offerId"`
	DocumentID string                         `json:"documentId"`
	Metadata   oway.UploadConfirmationRequest `json:"metadata"`
}

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic("Set " + name)
	}
	return value
}
func value[T ~string](p *T) string {
	if p == nil {
		return ""
	}
	return string(*p)
}
func ptr[T any](v T) *T { return &v }
func save(path string, state runState) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(state)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func prepare(ctx context.Context, c *oway.Client, stateFile string) error {
	if _, err := os.Stat(stateFile); err == nil {
		return fmt.Errorf("state file already exists; reconcile it first")
	}
	id := required("OWAY_OFFER_ID")
	pod, err := os.Open(required("OWAY_POD_FILE"))
	if err != nil {
		return err
	}
	defer func() { _ = pod.Close() }()
	data, err := io.ReadAll(io.LimitReader(pod, 10*1024*1024+1))
	if err != nil {
		return err
	}
	if len(data) == 0 || len(data) > 10*1024*1024 {
		return fmt.Errorf("POD must be at most 10 MiB")
	}
	var location oway.OfferLocationUpdate
	locationData, err := os.ReadFile(required("OWAY_LOCATION_FILE"))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(locationData, &location); err != nil {
		return err
	}
	location.Timestamp = time.Now().UTC()
	fmt.Println("Retrieve assigned offer")
	offer, err := c.GetCarrierOffer(ctx, id)
	if err != nil {
		return err
	}
	if offer == nil {
		return fmt.Errorf("offer missing")
	}
	if value(offer.Status) == "PENDING" {
		fmt.Println("Accept offer")
		offer, err = c.AcceptCarrierOffer(ctx, id, oway.AcceptOfferRequest{})
		if err != nil {
			return err
		}
	}
	if offer == nil {
		return fmt.Errorf("acceptance response missing")
	}
	if value(offer.Status) == "ACCEPTED" {
		fmt.Println("Confirm pickup")
		shipment, err := c.ConfirmPickup(ctx, id, oway.PickupConfirmationRequest{PickupTime: time.Now().UTC()})
		if err != nil {
			return err
		}
		if shipment == nil || value(shipment.Status) != "PICKED_UP" {
			return fmt.Errorf("pickup did not complete")
		}
	} else if value(offer.Status) != "PICKED_UP" {
		return fmt.Errorf("offer is not available for this exercise")
	}
	fmt.Println("Submit location")
	if _, err = c.SubmitLocation(ctx, id, location); err != nil {
		return err
	}
	filename := filepath.Base(required("OWAY_POD_FILE"))
	fmt.Println("Request upload URL")
	staging, err := c.RequestCarrierUploadURL(ctx, id, oway.UploadUrlRequest{DocumentType: "PROOF_OF_DELIVERY", ContentType: "application/pdf", FileName: &filename})
	if err != nil {
		return err
	}
	if staging == nil || value(staging.DocumentId) == "" || value(staging.ContentType) != "application/pdf" {
		return fmt.Errorf("incomplete upload response")
	}
	uploadURL, err := url.Parse(value(staging.UploadUrl))
	if err != nil || uploadURL.Scheme != "https" {
		return fmt.Errorf("expected HTTPS upload URL")
	}
	state := runState{id, value(staging.DocumentId), oway.UploadConfirmationRequest{DocumentType: "PROOF_OF_DELIVERY", ContentType: "application/pdf", FileName: filename}}
	if err = save(stateFile, state); err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/pdf")
	uploadClient := &http.Client{Timeout: time.Minute, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	fmt.Println("Upload POD bytes")
	response, err := uploadClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("storage PUT failed")
	}
	fmt.Println("Complete upload registration")
	completed, err := c.ConfirmCarrierUpload(ctx, id, state.DocumentID, state.Metadata)
	if err != nil {
		return err
	}
	if completed == nil || value(completed.DocumentId) != state.DocumentID {
		return fmt.Errorf("completion identity mismatch")
	}
	fmt.Println("POD registered. Request authorized review, then run OWAY_STEP=deliver.")
	return nil
}
func run() error {
	c, err := oway.New(oway.Config{ClientID: required("OWAY_M2M_CLIENT_ID"), ClientSecret: required("OWAY_M2M_CLIENT_SECRET"), APIKey: required("OWAY_API_KEY"), BaseURL: oway.EnvironmentSandbox, MaxRetries: -1})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	path := required("OWAY_STATE_FILE")
	step := required("OWAY_STEP")
	if step == "prepare" {
		return prepare(ctx, c, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var state runState
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	if step == "confirm-upload" {
		fmt.Println("Complete upload registration")
		_, err = c.ConfirmCarrierUpload(ctx, state.OfferID, state.DocumentID, state.Metadata)
		return err
	}
	if step != "deliver" {
		return fmt.Errorf("step must be prepare, confirm-upload or deliver")
	}
	fmt.Println("Retrieve registered POD and review status")
	doc, err := c.GetCarrierUploadedDocument(ctx, state.OfferID, state.DocumentID)
	if err != nil {
		return err
	}
	if doc == nil || value(doc.ReviewStatus) != "APPROVED" {
		return fmt.Errorf("POD requires authorized review")
	}
	fmt.Println("Reconcile shipment status")
	current, err := c.GetCarrierShipment(ctx, state.OfferID)
	if err != nil {
		return err
	}
	if current != nil && value(current.Status) == "DELIVERED" {
		fmt.Println("Shipment is already delivered; POD retrieved.")
		return nil
	}
	if current == nil || value(current.Status) != "PICKED_UP" {
		return fmt.Errorf("shipment must be picked up first")
	}
	fmt.Println("Confirm delivery")
	delivered, err := c.ConfirmDelivery(ctx, state.OfferID, oway.DeliveryConfirmationRequest{DeliveryTime: time.Now().UTC(), ProofOfDeliveryDocumentId: ptr(state.DocumentID)})
	if err != nil {
		return err
	}
	if delivered == nil || value(delivered.Status) != "DELIVERED" {
		return fmt.Errorf("delivery did not complete")
	}
	fmt.Println("Retrieve registered POD and review status")
	_, err = c.GetCarrierUploadedDocument(ctx, state.OfferID, state.DocumentID)
	if err == nil {
		fmt.Println("Sandbox shipment delivered; registered POD retrieved.")
	}
	return err
}
func main() {
	if err := run(); err != nil {
		if apiErr, ok := oway.AsError(err); ok {
			fmt.Fprintf(os.Stderr, "API status: %d\n", apiErr.StatusCode)
		}
		fmt.Fprintln(os.Stderr, "Carrier journey stopped. Preserve the state file and reconcile the last operation; see recovery instructions.")
		os.Exit(1)
	}
}
