# Oway SDK

Official SDKs for integrating with the Oway freight shipping platform.

## Available SDKs

| Language | Package | Installation |
|----------|---------|--------------|
| **TypeScript/JavaScript** | `@oway/sdk` | `npm install @oway/sdk` |
| **Go** | `github.com/Oway-Inc/oway-sdk/packages/go` | `go get github.com/Oway-Inc/oway-sdk/packages/go` |

## Quick Start

### TypeScript/JavaScript

```typescript
import Oway, { OwayEnvironments } from '@oway/sdk';

const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  apiKey: process.env.OWAY_API_KEY,
  baseUrl: OwayEnvironments.SANDBOX, // or PRODUCTION
});

const quote = await oway.quotes.create({
  origin: { zipCode: '90210', country: 'US' },
  destination: { zipCode: '10001', country: 'US' },
  items: [{ weight: 500, weightUnit: 'LBS' }],
});
```

### Go

```go
import oway "github.com/Oway-Inc/oway-sdk/packages/go"

client, _ := oway.New(oway.Config{
    ClientID:     os.Getenv("OWAY_M2M_CLIENT_ID"),
    ClientSecret: os.Getenv("OWAY_M2M_CLIENT_SECRET"),
    APIKey:       os.Getenv("OWAY_API_KEY"),
    BaseURL:      oway.EnvironmentSandbox, // or EnvironmentProduction
})

quote, _ := client.RequestQuote(ctx, &oway.QuoteRequest{...})
```

## Environments

| Environment | URL | Use For |
|-------------|-----|---------|
| **Sandbox** | `https://rest-api.sandbox.oway.io` | Development, testing |
| **Production** | `https://rest-api.oway.io` | Live shipments |

See [ENVIRONMENTS.md](./ENVIRONMENTS.md) for detailed configuration.

## Getting Credentials

### M2M Credentials (Required)
**Source:** Contact Oway Sales Engineering (support@oway.io)  
**What you get:** ClientID + ClientSecret

### Company API Keys
**Source:** Self-service at [app.oway.io/settings/api](https://app.oway.io/settings/api)  
**Types:** Test keys (`oway_sk_test_...`) and Live keys (`oway_sk_live_...`)

See [AUTH_GUIDE.md](./AUTH_GUIDE.md) for authentication details.

## Documentation

- [TypeScript SDK](./packages/typescript/README.md)
- [Go SDK](./packages/go/README.md)
- [API Documentation](https://docs.shipoway.com)

## Support

- **Email**: support@oway.io
- **Documentation**: [docs.shipoway.com](https://docs.shipoway.com)

## License

MIT
