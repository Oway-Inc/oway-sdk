# Oway Go SDK

Official Go SDK for the Oway freight shipping platform.

## Installation

```bash
go get github.com/Oway-Inc/oway-sdk/packages/go
```

Requires Go 1.22 or newer.

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

	shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{
		QuoteId:         quote.Id,
		PickupAddress:   oway.Address{ /* ... */ },
		DeliveryAddress: oway.Address{ /* ... */ },
		OrderComponents: []oway.OrderComponent{
			{PalletCount: 2, PoundsWeight: 1000, PalletDimensions: []int32{48, 40, 48}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Order: %s\n", *shipment.OrderNumber)
}
```

## Authentication

The SDK manages both authentication tokens automatically:

- **M2M JWT** (`Authorization: Bearer`) refreshed five minutes before expiry.
- **Company API key** (`x-oway-api-key`) attached to every request.

Set the default key on the client, or supply a per-request key via context for multi-tenant use:

```go
ctx := oway.WithCompanyAPIKey(context.Background(), "oway_sk_tenant_xyz")
quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{ /* ... */ })
```

## Error Handling

Every method returns `*oway.Error` on a non-2xx response. The error carries the parsed RFC 9457 `ProblemDetail`, the server-issued request id, and per-field validation failures when present.

```go
shipment, err := client.CreateShipment(ctx, req)
if err != nil {
	if oe, ok := oway.AsError(err); ok {
		// Programmatic branching:
		switch oe.Code {
		case "no_coverage":
			// Lane not within Oway's coverage area.
		case "account_restriction":
			// Account does not have the requested service enabled.
		case "daily_trip_limit":
			// Trip cap hit for the pickup date.
		}

		// Validation failures expose per-field reasons:
		for _, v := range oe.Violations {
			fmt.Printf("  %s: %s\n", v.Field, v.Reason)
		}

		// Request id to quote when reporting an issue:
		fmt.Println("requestId:", oe.RequestID)
	}
	return err
}
```

`Error.IsRetryable()` is true for 408, 429, 500, 502, 503, and 504. The SDK already retries those for you with full-jitter exponential backoff; the helper is exposed so callers can decide whether to surface failures up.

## Retries

Transient failures are retried automatically using full-jitter exponential backoff capped at 30 seconds. Configure with `Config.MaxRetries`:

```go
oway.New(oway.Config{
	// ...
	MaxRetries: 5, // default 3; pass -1 to disable
})
```

Retries respect `context.Context` cancellation.

## Logging

The SDK is silent by default. Supply an `*slog.Logger` to receive structured events:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
oway.New(oway.Config{ /* ... */, Logger: logger })
```

## API Methods

### Quotes

```go
quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{ /* ... */ })
quote, err := client.GetQuote(ctx, quoteID)
```

### Shipments

```go
shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{ /* ... */ })
shipment, err := client.GetShipment(ctx, orderNumber)
shipment, err := client.ConfirmShipment(ctx, orderNumber)
shipment, err := client.CancelShipment(ctx, orderNumber)
```

### Tracking, Invoices, Documents

```go
tracking, err := client.TrackShipment(ctx, orderNumber)
invoice, err  := client.GetInvoice(ctx, orderNumber)
doc, err      := client.GetDocument(ctx, orderNumber, oway.DocumentTypeBOL)
```

Document types: `DocumentTypeBOL`, `DocumentTypeInvoice`, `DocumentTypeShippingLabel`, `DocumentTypePOD`.

## Configuration

```go
oway.New(oway.Config{
	ClientID:     "...",                   // required
	ClientSecret: "...",                   // required
	APIKey:       "oway_sk_...",           // default company key (optional)
	BaseURL:      oway.EnvironmentSandbox, // default
	TokenURL:     "...",                   // default: BaseURL + /v1/auth/token
	HTTPClient:   &http.Client{},          // optional override
	MaxRetries:   3,                       // 0 -> default 3, -1 disables
	Logger:       slog.Default(),          // optional
})
```

## Environments

| Environment | Constant | URL |
|---|---|---|
| Sandbox | `oway.EnvironmentSandbox` | `https://api.sandbox.oway.io` |
| Production | `oway.EnvironmentProduction` | `https://api.oway.io` |

## Type Aliases

Clean public names mapped from generated types:

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
| `oway.DocumentType` | `client.GetDocumentParamsDocumentType` |

## Advanced

The underlying oapi-codegen client is exposed for use cases not covered by the wrapper:

```go
raw := client.GeneratedClient()
resp, err := raw.GetShipmentWithResponse(ctx, orderNumber)
```

## Support

- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)
- **API Reference**: [api.oway.io/api-docs](https://api.oway.io/api-docs)
- **Email**: support@oway.io

## License

MIT
