package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	oway "github.com/Oway-Inc/oway-sdk/packages/go"
)

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic("Set " + name + " before running this sandbox example")
	}
	return value
}

func readJSON(name string, value any) error {
	data, err := os.ReadFile(required(name))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func run() error {
	statePath := required("OWAY_STATE_FILE")
	if _, err := os.Stat(statePath); err == nil {
		return fmt.Errorf("state file exists; reconcile it before another run")
	}
	var quoteRequest oway.QuoteRequest
	var shipmentRequest oway.ShipmentRequest
	if err := readJSON("OWAY_QUOTE_FILE", &quoteRequest); err != nil {
		return err
	}
	if err := readJSON("OWAY_SHIPMENT_FILE", &shipmentRequest); err != nil {
		return err
	}
	c, err := oway.New(oway.Config{ClientID: required("OWAY_M2M_CLIENT_ID"), ClientSecret: required("OWAY_M2M_CLIENT_SECRET"), APIKey: required("OWAY_API_KEY"), BaseURL: oway.EnvironmentSandbox, MaxRetries: -1})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = oway.WithIdempotencyKey(ctx, required("OWAY_IDEMPOTENCY_KEY"))
	fmt.Println("Request quote")
	quote, err := c.RequestQuote(ctx, &quoteRequest)
	if err != nil {
		return err
	}
	if quote == nil || quote.Id == nil || *quote.Id == "" {
		return fmt.Errorf("no usable quote")
	}
	shipmentRequest.QuoteID = quote.Id
	state, err := os.OpenFile(statePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(state).Encode(struct {
		Request        oway.ShipmentRequest `json:"request"`
		IdempotencyKey string               `json:"idempotencyKey"`
	}{shipmentRequest, required("OWAY_IDEMPOTENCY_KEY")})
	closeErr := state.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Println("Create or replay saved shipment")
	shipment, err := c.CreateShipment(ctx, &shipmentRequest)
	if err != nil {
		return err
	}
	if shipment == nil || shipment.OrderNumber == nil {
		return fmt.Errorf("create returned no order number; reconcile before retrying")
	}
	number := *shipment.OrderNumber
	fmt.Println("Create or replay saved shipment")
	replay, err := c.CreateShipment(ctx, &shipmentRequest)
	if err != nil {
		return err
	}
	if replay == nil || replay.OrderNumber == nil || *replay.OrderNumber != number {
		return fmt.Errorf("create replay returned a different shipment")
	}
	fmt.Println("Confirm shipment")
	confirmed, err := c.ConfirmShipment(ctx, number)
	if err != nil {
		return err
	}
	if confirmed == nil || confirmed.OrderStatus == nil || string(*confirmed.OrderStatus) != "CONFIRMED" {
		return fmt.Errorf("shipment not confirmed")
	}
	fmt.Println("Retrieve tracking")
	if _, err = c.TrackShipment(ctx, number); err != nil {
		return err
	}
	fmt.Println("Retrieve bill of lading")
	doc, err := c.GetDocument(ctx, number, oway.DocumentTypeBOL)
	if err != nil {
		return err
	}
	if doc == nil || doc.DownloadLink == nil || *doc.DownloadLink == "" {
		return fmt.Errorf("BOL not yet available")
	}
	fmt.Printf("Sandbox shipment %s: confirmed, tracking and BOL retrieved\n", number)
	return nil
}

func main() {
	if err := run(); err != nil {
		if apiErr, ok := oway.AsError(err); ok {
			fmt.Fprintf(os.Stderr, "API status: %d\n", apiErr.StatusCode)
		}
		fmt.Fprintln(os.Stderr, "Sandbox journey stopped. Reconcile the last completed operation before rerunning; see the quickstart recovery guide.")
		os.Exit(1)
	}
}
