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
	// Initialize client
	client, err := oway.New(oway.Config{
		APIKey: os.Getenv("OWAY_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Request a quote
	quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{
		Origin: oway.Address{
			Name:          "Warehouse LA",
			Address1:      "123 Warehouse Rd",
			City:          "Los Angeles",
			State:         "CA",
			ZipCode:       "90210",
			PhoneNumber:   "555-0123",
			ContactPerson: "John Doe",
		},
		Destination: oway.Address{
			Name:          "Distribution NYC",
			Address1:      "456 Distribution Ave",
			City:          "New York",
			State:         "NY",
			ZipCode:       "10001",
			PhoneNumber:   "555-5678",
			ContactPerson: "Jane Smith",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Quote: $%.2f\n", float64(quote.QuotedPriceInCents.Value)/100)

	// Schedule shipment
	shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{
		QuoteId: quote.Id,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Order Number: %s\n", shipment.OrderNumber)
}
```

## Getting Your API Key

**For Shippers:**
1. Sign up at [app.oway.io](https://app.oway.io)
2. Navigate to Settings → API Keys  
3. Create a test API key (`oway_sk_test_...`)

**For Carriers:**
Contact support@oway.io for carrier credentials (`oway_ck_...`)

## Authentication

The SDK handles authentication automatically:
- **API Key** - Your company-specific key
- **Bearer Token** - Auto-refreshed JWT token

Features:
- Thread-safe token caching with `sync.RWMutex`
- Automatic refresh 5 minutes before expiry
- Double-check locking (prevents thundering herd)

## API Methods

### Quotes

```go
// Request quote
quote, err := client.RequestQuote(ctx, &oway.QuoteRequest{
    Origin:      oway.Address{...},
    Destination: oway.Address{...},
})
```

### Shipments

```go
// Create shipment
shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{
    QuoteId: quote.Id,
})

// Access full ogen client for advanced operations
fullClient := client.GetClient()
tracking, err := fullClient.TrackShipmentByOrderNumber(ctx, orderNumber)
```

## Configuration

```go
oway.New(oway.Config{
    APIKey:     "oway_sk_...",              // Required
    BaseURL:    "https://rest-api.oway.io",     // Optional
    TokenURL:   "...",                      // Optional
    HTTPClient: &http.Client{},             // Optional
    Debug:      true,                       // Optional
})
```

## Error Handling

```go
shipment, err := client.CreateShipment(ctx, request)
if err != nil {
    var owayErr *oway.Error
    if errors.As(err, &owayErr) {
        fmt.Printf("Code: %s, Status: %d\n", owayErr.Code, owayErr.StatusCode)
        
        if owayErr.IsRetryable() {
            // Retry logic
        }
    }
}
```

## Context Support

All methods accept `context.Context` for timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

quote, err := client.RequestQuote(ctx, request)
```

## Type Aliases

Clean, intuitive names:

| Clean Name | Internal Type |
|------------|---------------|
| `oway.Quote` | `client.ExternalRequestQuoteResponse` |
| `oway.Shipment` | `client.ExternalOrder` |
| `oway.Address` | `client.ExternalAddress` |

## Support

- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)
- **API Reference**: [rest-api.oway.io/api-docs](https://rest-api.oway.io/api-docs)
- **Email**: support@oway.io

## License

MIT
