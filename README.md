# Oway SDK

Official SDKs for integrating with the [Oway](https://shipoway.com) freight shipping platform. Get quotes, book shipments, track deliveries, and retrieve documents — all with a few lines of code.

## Available SDKs

| Language | Package |
|----------|---------|
| **TypeScript/JavaScript** | [`@oway/sdk`](https://www.npmjs.com/package/@oway/sdk) |
| **Go** | `github.com/Oway-Inc/oway-sdk/packages/go` |

## Quick Start (TypeScript/JavaScript)

### 1. Install

```bash
npm install @oway/sdk
```

### 2. Set up credentials

```bash
export OWAY_M2M_CLIENT_ID="your-client-id"
export OWAY_M2M_CLIENT_SECRET="your-client-secret"
export OWAY_API_KEY="oway_sk_test_..."  # optional — from app.oway.io/settings/api
```

### 3. Get a quote, book a shipment, and track it

```typescript
import Oway, { OwayEnvironments } from '@oway/sdk';

// Initialize the client
const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  apiKey: process.env.OWAY_API_KEY,
  baseUrl: OwayEnvironments.SANDBOX,
});

// Get a shipping quote
const quote = await oway.quotes.create({
  pickupAddress: {
    name: 'Warehouse LA',
    address1: '123 Warehouse Rd',
    city: 'Los Angeles',
    state: 'CA',
    zipCode: '90210',
    phoneNumber: '+15550123456',
    contactPerson: 'John Doe',
  },
  deliveryAddress: {
    name: 'Distribution NYC',
    address1: '456 Distribution Ave',
    city: 'New York',
    state: 'NY',
    zipCode: '10001',
    phoneNumber: '+15555678901',
    contactPerson: 'Jane Smith',
  },
  orderComponents: [
    { palletCount: 2, poundsWeight: 1000, palletDimensions: [48, 40, 48] },
  ],
});

console.log(`Quote: $${(quote.quotedPriceInCents ?? 0) / 100}`);

// Book the shipment
const shipment = await oway.shipments.create({
  quoteId: quote.id,
  pickupAddress: quote.pickupAddress,
  deliveryAddress: quote.deliveryAddress,
  orderComponents: [
    { palletCount: 2, poundsWeight: 1000, palletDimensions: [48, 40, 48] },
  ],
  description: 'Electronics - fragile',
});

console.log(`Order: ${shipment.orderNumber}`);

// Confirm the shipment
await oway.shipments.confirm(shipment.orderNumber!);

// Track the shipment
const tracking = await oway.shipments.tracking(shipment.orderNumber!);
console.log(`Status: ${tracking.orderStatus}`);
```

## Quick Start (Go)

### 1. Install

```bash
go get github.com/Oway-Inc/oway-sdk/packages/go
```

### 2. Set up credentials

```bash
export OWAY_M2M_CLIENT_ID="your-client-id"
export OWAY_M2M_CLIENT_SECRET="your-client-secret"
export OWAY_API_KEY="oway_sk_test_..."  # optional — from app.oway.io/settings/api
```

### 3. Get a quote, book a shipment, and track it

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
	// Initialize the client
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

	// Get a shipping quote
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

	// Book the shipment
	shipment, err := client.CreateShipment(ctx, &oway.ShipmentRequest{
		QuoteId: quote.Id,
		PickupAddress: oway.Address{
			Name: "Warehouse LA", Address1: "123 Warehouse Rd",
			City: "Los Angeles", State: "CA", ZipCode: "90210",
			PhoneNumber: "+15550123456", ContactPerson: "John Doe",
		},
		DeliveryAddress: oway.Address{
			Name: "Distribution NYC", Address1: "456 Distribution Ave",
			City: "New York", State: "NY", ZipCode: "10001",
			PhoneNumber: "+15555678901", ContactPerson: "Jane Smith",
		},
		OrderComponents: []oway.OrderComponent{
			{PalletCount: 2, PoundsWeight: 1000, PalletDimensions: []int32{48, 40, 48}},
		},
		Description: "Electronics - fragile",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Order: %s\n", *shipment.OrderNumber)

	// Confirm the shipment
	_, err = client.ConfirmShipment(ctx, *shipment.OrderNumber)
	if err != nil {
		log.Fatal(err)
	}

	// Track the shipment
	tracking, err := client.TrackShipment(ctx, *shipment.OrderNumber)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %s\n", *tracking.OrderStatus)
}
```

## Environments

| Environment | Constant | URL |
|-------------|----------|-----|
| **Sandbox** | `OwayEnvironments.SANDBOX` | `https://rest-api.sandbox.oway.io` |
| **Production** | `OwayEnvironments.PRODUCTION` | `https://rest-api.oway.io` |

Start with Sandbox for development and testing. Switch to Production when you're ready for live shipments. See [ENVIRONMENTS.md](./ENVIRONMENTS.md) for details.

## Getting Credentials

### M2M Credentials (Required)
Contact Oway Sales Engineering at **support@oway.io** to get your `clientId` and `clientSecret`. These authenticate your application with the Oway platform.

### Company API Keys (Optional)
Generate API keys self-service at [app.oway.io/settings/api](https://app.oway.io/settings/api):
- **Test keys** (`oway_sk_test_...`) — for sandbox development
- **Live keys** (`oway_sk_live_...`) — for production shipments

See [AUTH_GUIDE.md](./AUTH_GUIDE.md) for full authentication details.

## What You Can Do

| Feature | Method |
|---------|--------|
| Get a shipping quote | `oway.quotes.create(request)` |
| Retrieve a quote | `oway.quotes.retrieve(quoteId)` |
| Book a shipment | `oway.shipments.create(request)` |
| Retrieve a shipment | `oway.shipments.retrieve(orderNumber)` |
| Confirm a shipment | `oway.shipments.confirm(orderNumber)` |
| Cancel a shipment | `oway.shipments.cancel(orderNumber)` |
| Track a shipment | `oway.shipments.tracking(orderNumber)` |
| Get documents (BOL, label) | `oway.shipments.document(orderNumber, type)` |
| Get invoice | `oway.shipments.invoice(orderNumber)` |

## Documentation

- **[TypeScript SDK Reference](./packages/typescript/README.md)** — full API reference, error handling, configuration
- **[Go SDK Reference](./packages/go/README.md)** — Go client documentation
- **[API Documentation](https://docs.shipoway.com)** — platform docs and guides
- **[AI Agent Guide](./AI_AGENT_GUIDE.md)** — integrate with Copilot, Cursor, Claude Code, and more

## Requirements

- **Node.js** >= 18.0.0 (TypeScript/JavaScript)
- **Go** >= 1.21 (Go)

## Support

- **Email**: support@oway.io
- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)

## License

MIT
