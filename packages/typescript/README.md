# @oway/sdk

Official Oway SDK for JavaScript and TypeScript.

[![npm version](https://img.shields.io/npm/v/@oway/sdk.svg)](https://www.npmjs.com/package/@oway/sdk)

## Installation

```bash
npm install @oway/sdk
```

## 5-Minute Quickstart

```typescript
import Oway from '@oway/sdk';

// 1. Initialize with your API key
const oway = new Oway({
  apiKey: 'oway_sk_test_...',
});

// 2. Get a shipping quote
const quote = await oway.quotes.create({
  origin: {
    addressLine1: '123 Warehouse Rd',
    city: 'Los Angeles',
    state: 'CA',
    zipCode: '90210',
    country: 'US',
  },
  destination: {
    addressLine1: '456 Distribution Center',
    city: 'New York',
    state: 'NY',
    zipCode: '10001',
    country: 'US',
  },
  items: [{
    description: 'Pallets',
    weight: 500,
    weightUnit: 'LBS',
    quantity: 2,
  }],
});

console.log(`Quote: $${quote.totalCharge / 100}`);

// 3. Schedule the shipment
const shipment = await oway.shipments.create({
  quoteId: quote.quoteId,
  pickup: {
    contactName: 'John Doe',
    contactPhone: '555-1234',
    contactEmail: 'pickup@example.com',
  },
  delivery: {
    contactName: 'Jane Smith',
    contactPhone: '555-5678',
    contactEmail: 'delivery@example.com',
  },
});

// 4. Confirm and track
await oway.shipments.confirm(shipment.orderNumber);
const tracking = await oway.shipments.tracking(shipment.orderNumber);
console.log(`Status: ${tracking.status}`);
```

## Getting Your API Key

**For Shippers:**
1. Log in to [app.oway.io](https://app.oway.io)
2. Go to Settings → API Keys
3. Create a new API key
4. Use test keys (`oway_sk_test_...`) for development

**For Carriers:**
Contact support@oway.io for carrier API credentials (`oway_ck_...`)

## Authentication

The SDK automatically handles authentication:
- **API Key** (`x-oway-api-key` header) - Your company-specific key
- **Access Token** (`Authorization: Bearer`) - Auto-refreshed JWT

You only provide your API key - token management is automatic.

## API Reference

### Quotes

#### Create Quote
```typescript
const quote = await oway.quotes.create({
  origin: {
    addressLine1: string,
    city: string,
    state: string,
    zipCode: string,
    country: string,
  },
  destination: { /* same as origin */ },
  items: [{
    description: string,
    weight: number,
    weightUnit: 'LBS' | 'KG',
    quantity: number,
  }],
  pickupDate?: string,     // Optional: ISO 8601 date
  accessorials?: string[], // Optional: ['LIFTGATE', 'RESIDENTIAL']
});
```

#### Retrieve Quote
```typescript
const quote = await oway.quotes.retrieve('quote_abc123');
```

### Shipments

#### Create Shipment
```typescript
const shipment = await oway.shipments.create({
  quoteId: string,
  pickup: {
    contactName: string,
    contactPhone: string,
    contactEmail: string,
    instructions?: string,
  },
  delivery: { /* same as pickup */ },
  reference?: string, // Your PO number
});
```

#### Retrieve Shipment
```typescript
const shipment = await oway.shipments.retrieve('ORD-123456');
```

#### Confirm Shipment
```typescript
await oway.shipments.confirm('ORD-123456');
```

#### Cancel Shipment
```typescript
await oway.shipments.cancel('ORD-123456');
```

#### Track Shipment
```typescript
const tracking = await oway.shipments.tracking('ORD-123456');
console.log(tracking.status);        // 'IN_TRANSIT', 'DELIVERED', etc.
console.log(tracking.estimatedDelivery);
```

#### Get Document
```typescript
const { url } = await oway.shipments.document('ORD-123456', 'BOL');
// Download BOL, INVOICE, or LABEL
```

## Configuration

```typescript
const oway = new Oway({
  apiKey: string,         // Required
  baseUrl?: string,       // Optional: Default 'https://rest-api.sandbox.oway.io'
  maxRetries?: number,    // Optional: Default 3
  timeout?: number,       // Optional: Default 30000ms
  debug?: boolean,        // Optional: Enable logging
});
```

## Error Handling

```typescript
import { OwayError } from '@oway/sdk';

try {
  const quote = await oway.quotes.create({...});
} catch (error) {
  if (error instanceof OwayError) {
    console.error({
      message: error.message,
      code: error.code,           // Error code
      statusCode: error.statusCode, // HTTP status
      requestId: error.requestId,   // For support
    });
  }
}
```

## TypeScript Support

Full type definitions included:

```typescript
import type { Quote, Shipment, QuoteRequest } from '@oway/sdk';

const request: QuoteRequest = {
  // TypeScript will autocomplete all fields
};
```

## Environment Variables

Store API keys securely:

```typescript
const oway = new Oway({
  apiKey: process.env.OWAY_API_KEY!,
});
```

Example `.env`:
```
OWAY_API_KEY=oway_sk_test_abc123...
```

## Support

- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)
- **API Reference**: [rest-api.oway.io/api-docs](https://rest-api.oway.io/api-docs)
- **Support**: support@oway.io

## License

MIT
