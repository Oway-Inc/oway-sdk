# @oway/sdk

Official Oway SDK for JavaScript and TypeScript.

[![npm version](https://img.shields.io/npm/v/@oway/sdk.svg)](https://www.npmjs.com/package/@oway/sdk)

## Installation

```bash
npm install @oway/sdk
```

## Quick Start

```typescript
import Oway, { OwayEnvironments } from '@oway/sdk';

// 1. Initialize with M2M credentials
const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  apiKey: process.env.OWAY_API_KEY,       // Optional: default company API key
  baseUrl: OwayEnvironments.SANDBOX,
});

// 2. Get a shipping quote
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

// 3. Create a shipment from the quote
const shipment = await oway.shipments.create({
  quoteId: quote.id,
  pickupAddress: { /* same as above */ },
  deliveryAddress: { /* same as above */ },
  orderComponents: [
    { palletCount: 2, poundsWeight: 1000, palletDimensions: [48, 40, 48] },
  ],
  description: 'Electronics - fragile',
});

console.log(`Order: ${shipment.orderNumber} (${shipment.orderStatus})`);

// 4. Confirm and track
await oway.shipments.confirm(shipment.orderNumber!);
const tracking = await oway.shipments.tracking(shipment.orderNumber!);
console.log(`Status: ${tracking.orderStatus}`);
```

## Getting Credentials

### M2M Credentials (Required)
Contact Oway Sales Engineering (support@oway.io) to obtain a `clientId` and `clientSecret`.

### Company API Key (Optional)
Self-service at [app.oway.io/settings/api](https://app.oway.io/settings/api).
- Shipper keys: `oway_sk_test_...` (test) / `oway_sk_live_...` (production)
- Carrier keys: `oway_ck_...`

## Authentication

The SDK handles authentication automatically:
- **M2M Token** (`Authorization: Bearer`) - Auto-refreshed JWT from clientId/clientSecret
- **API Key** (`x-oway-api-key`) - Company-specific key for authorization

You provide credentials once at initialization - token management is automatic.

## API Reference

### Quotes

```typescript
// Request a quote
const quote = await oway.quotes.create({
  pickupAddress: { name, address1, city, state, zipCode, phoneNumber, contactPerson },
  deliveryAddress: { name, address1, city, state, zipCode, phoneNumber, contactPerson },
  orderComponents: [{ palletCount: 2, poundsWeight: 1000, palletDimensions: [48, 40, 48] }],
});

// Retrieve a quote by ID
const quote = await oway.quotes.retrieve('507f1f77bcf86cd799439011');
```

### Shipments

```typescript
// Create a shipment (optionally from a quote)
const shipment = await oway.shipments.create({
  quoteId: quote.id,           // Optional: lock in quoted price
  pickupAddress: { ... },
  deliveryAddress: { ... },
  orderComponents: [{ palletCount: 2, poundsWeight: 1000, palletDimensions: [48, 40, 48] }],
  description: 'Palletized freight',
  poNumber: 'PO-2024-12345',  // Optional
});

// Retrieve a shipment
const shipment = await oway.shipments.retrieve('ZKYQ5');

// Confirm a shipment
await oway.shipments.confirm('ZKYQ5');

// Cancel a shipment
await oway.shipments.cancel('ZKYQ5');
```

### Tracking

```typescript
const tracking = await oway.shipments.tracking('ZKYQ5');
console.log(tracking.orderStatus);          // 'IN_TRANSIT', 'DELIVERED', etc.
console.log(tracking.estimatedDeliveryDate);
console.log(tracking.actualDeliveryDate);
```

### Invoices

```typescript
const invoice = await oway.shipments.invoice('ZKYQ5');
console.log(`Total: $${(invoice.totalChargesInCents ?? 0) / 100}`);
console.log(`Line items: ${invoice.lineItems?.length}`);
```

### Documents

```typescript
const { url } = await oway.shipments.document('ZKYQ5', 'BILL_OF_LADING');
// Available types: 'BILL_OF_LADING', 'INVOICE', 'SHIPPING_LABEL'
```

### Multi-Tenant (Per-Request API Key)

All methods accept an optional `companyApiKey` as the last argument:

```typescript
const shipment = await oway.shipments.create(request, 'oway_sk_...');
const tracking = await oway.shipments.tracking('ZKYQ5', 'oway_sk_...');
const invoice  = await oway.shipments.invoice('ZKYQ5', 'oway_sk_...');
```

## Configuration

```typescript
const oway = new Oway({
  clientId: string,        // Required: M2M client ID
  clientSecret: string,    // Required: M2M client secret
  apiKey?: string,         // Optional: Default company API key
  baseUrl?: string,        // Optional: Defaults to sandbox
  tokenUrl?: string,       // Optional: Custom token endpoint
  maxRetries?: number,     // Optional: Default 3
  timeout?: number,        // Optional: Default 30000ms
  debug?: boolean,         // Optional: Enable debug logging
});
```

## Environments

| Environment | Constant | URL |
|-------------|----------|-----|
| Sandbox | `OwayEnvironments.SANDBOX` | `https://rest-api.sandbox.oway.io` |
| Production | `OwayEnvironments.PRODUCTION` | `https://rest-api.oway.io` |

## Error Handling

```typescript
import { OwayError } from '@oway/sdk';

try {
  const quote = await oway.quotes.create({ ... });
} catch (error) {
  if (error instanceof OwayError) {
    console.error({
      message: error.message,
      code: error.code,             // Error code
      statusCode: error.statusCode, // HTTP status
      requestId: error.requestId,   // For support
    });

    if (error.isRetryable()) {
      // Retry logic - SDK retries automatically up to maxRetries
    }
  }
}
```

## TypeScript Support

Full type definitions included:

```typescript
import type { Quote, Shipment, Tracking, Invoice, QuoteRequest, ShipmentRequest, Address } from '@oway/sdk';
```

## Support

- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)
- **API Reference**: [rest-api.oway.io/api-docs](https://rest-api.oway.io/api-docs)
- **Email**: support@oway.io

## License

MIT
