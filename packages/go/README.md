# Oway Go SDK

Official Go SDK for the Oway freight shipping platform.

## Installation

```bash
go get github.com/Oway-Inc/oway-sdk/packages/go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	oway "github.com/Oway-Inc/oway-sdk/packages/go"
)

func main() {
	client, err := oway.New(oway.Config{
		ClientID:     os.Getenv("OWAY_M2M_CLIENT_ID"),
		ClientSecret: os.Getenv("OWAY_M2M_CLIENT_SECRET"),
		APIKey:       os.Getenv("OWAY_API_KEY"),
		BaseURL:      oway.EnvironmentSandbox,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Request a quote
	quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{
		PickupAddress: oway.Address{
			Name:          "Warehouse LA",
			Address1:      "123 Warehouse Rd",
			City:          "Los Angeles",
			State:         "CA",
			ZipCode:       "90210",
			PhoneNumber:   "+15550123456",
			ContactPerson: "John Doe",
		},
		DeliveryAddress: oway.Address{
			Name:          "Distribution NYC",
			Address1:      "456 Distribution Ave",
			City:          "New York",
			State:         "NY",
			ZipCode:       "10001",
			PhoneNumber:   "+15555678901",
			ContactPerson: "Jane Smith",
		},
		OrderComponents: []oway.OrderComponent{
			{PalletCount: 2, PoundsWeight: 1000, PalletDimensions: []int32{48, 40, 48}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Quote: $%.2f\n", float64(*quote.QuotedPriceInCents)/100)

	// Create shipment from quote
	shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{
		QuoteId:         quote.Id,
		PickupAddress:   oway.Address{/* same as above */},
		DeliveryAddress: oway.Address{/* same as above */},
		OrderComponents: []oway.OrderComponent{
			{PalletCount: 2, PoundsWeight: 1000, PalletDimensions: []int32{48, 40, 48}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Order: %s (status: %s)\n", *shipment.OrderNumber, *shipment.OrderStatus)

	// Confirm, track, and get invoice
	shipment, _ = client.ConfirmShipment(ctx, *shipment.OrderNumber)
	tracking, _ := client.TrackShipment(ctx, *shipment.OrderNumber)
	fmt.Printf("Status: %s\n", *tracking.OrderStatus)
}
```

## Getting Credentials

### M2M Credentials (Required)
Contact Oway Sales Engineering (support@oway.io) to obtain a `ClientID` and `ClientSecret`.

### Company API Key (Optional)
Self-service at [app.oway.io/settings/api](https://app.oway.io/settings/api).
- Shipper keys: `oway_sk_test_...` (test) / `oway_sk_live_...` (production)
- Carrier keys: `oway_ck_...`

## Authentication

The SDK handles authentication automatically:
- **M2M Token** (`Authorization: Bearer`) - Auto-refreshed JWT from ClientID/ClientSecret
- **API Key** (`x-oway-api-key`) - Company-specific key for authorization

Features:
- Thread-safe token caching with `sync.RWMutex`
- Automatic refresh 5 minutes before expiry
- Double-check locking (prevents thundering herd)

## API Methods

All methods have a `ForCompany` variant that accepts a per-request API key for multi-tenant integrations.

### Quotes

```go
// Request a quote
quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{...})

// Retrieve a quote by ID
quote, err := client.GetQuoteByID(ctx, quoteID)

// Multi-tenant: specify company API key per request
quote, err := client.RequestQuoteForCompany(ctx, &oway.QuoteRequest{...}, "oway_sk_...")
```

### Shipments

```go
// Create a shipment
shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{...})

// Retrieve a shipment by order number
shipment, err := client.GetShipment(ctx, orderNumber)

// Confirm a shipment
shipment, err := client.ConfirmShipment(ctx, orderNumber)

// Cancel a shipment
shipment, err := client.CancelShipment(ctx, orderNumber)
```

### Tracking

```go
tracking, err := client.TrackShipment(ctx, orderNumber)
fmt.Printf("Status: %s\n", *tracking.OrderStatus)
fmt.Printf("ETA: %v\n", tracking.EstimatedDeliveryDate)
```

### Invoices

```go
invoice, err := client.GetInvoice(ctx, orderNumber)
fmt.Printf("Total: $%.2f\n", float64(*invoice.TotalChargesInCents)/100)
```

### Documents

```go
doc, err := client.GetDocument(ctx, orderNumber, oway.DocumentTypeBOL)
fmt.Printf("Download: %s\n", *doc.DownloadLink)
```

Available document types: `oway.DocumentTypeBOL`, `oway.DocumentTypeInvoice`, `oway.DocumentTypeShippingLabel`

## Configuration

```go
oway.New(oway.Config{
    ClientID:     "...",                   // Required: M2M client ID
    ClientSecret: "...",                   // Required: M2M client secret
    APIKey:       "oway_sk_...",           // Optional: Default company API key
    BaseURL:      oway.EnvironmentSandbox, // Optional: defaults to sandbox
    TokenURL:     "...",                   // Optional: custom token endpoint
    HTTPClient:   &http.Client{},          // Optional: custom HTTP client
    Debug:        true,                    // Optional: enable debug logging
})
```

## Environments

| Environment | Constant | URL |
|-------------|----------|-----|
| Sandbox | `oway.EnvironmentSandbox` | `https://rest-api.sandbox.oway.io` |
| Production | `oway.EnvironmentProduction` | `https://rest-api.oway.io` |

## Type Aliases

Clean names mapped from generated types:

| SDK Type | Generated Type |
|----------|----------------|
| `oway.QuoteRequest` | `client.QuoteRequest` |
| `oway.ShipmentRequest` | `client.CreateShipmentRequest` |
| `oway.Quote` | `client.QuoteResponse` |
| `oway.Shipment` | `client.Shipment` |
| `oway.Tracking` | `client.Tracking` |
| `oway.Invoice` | `client.InvoiceResponse` |
| `oway.Address` | `client.Address` |
| `oway.OrderComponent` | `client.OrderComponent` |
| `oway.Document` | `client.DocumentResponse` |
| `oway.DocumentType` | `client.GetDocumentByOrderNumberParamsDocumentType` |

## Context Support

All methods accept `context.Context` for timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

quote, err := client.RequestQuote(ctx, request)
```

## Advanced Usage

Access the underlying oapi-codegen client directly:

```go
raw := client.GetClient()
resp, err := raw.GetShipmentByOrderNumberWithResponse(ctx, orderNumber)
```

## Support

- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)
- **API Reference**: [rest-api.oway.io/api-docs](https://rest-api.oway.io/api-docs)
- **Email**: support@oway.io

## License

MIT
