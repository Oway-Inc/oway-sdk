# Oway SDK - Project Guide

This is the official Oway SDK mono-repo for multiple programming languages.

## Project Structure

```
oway-sdk/
├── packages/
│   ├── typescript/          # TypeScript/JavaScript SDK (main SDK)
│   │   ├── src/
│   │   │   ├── client.ts    # HTTP client with auth
│   │   │   ├── index.ts     # Main entry point
│   │   │   ├── resources/   # API resources (quotes, shipments)
│   │   │   ├── generated/   # Auto-generated from OpenAPI spec
│   │   │   └── ai/          # MCP server for AI agents
│   │   ├── examples/        # Code examples for AI training
│   │   └── README.md        # SDK documentation
│   ├── go/                  # Go SDK (future)
│   └── java/                # Java SDK (future)
├── openapi/
│   └── spec.json           # Source of truth - OpenAPI 3.0.1 spec
└── scripts/                # Generation scripts

## Key Concepts

### 1. Two-Layer Architecture

Each SDK has two layers:
- **Generated layer**: Type-safe client from OpenAPI spec (in `src/generated/`)
- **Wrapper layer**: Hand-crafted DX with resource-based API (in `src/resources/`)

### 2. Authentication

Oway API requires TWO credentials (handled automatically by SDK):
- **API Key**: `x-oway-api-key` header (company-specific)
- **Access Token**: `Authorization: Bearer` header (auto-refreshed JWT)

The SDK manages token refresh automatically.

### 3. Resource-Based API

Like Stripe's SDK:
```typescript
oway.quotes.create()       // Not: oway.createQuote()
oway.shipments.tracking()  // Not: oway.getShipmentTracking()
```

This is more intuitive for developers and AI agents.

### 4. AI Agent Ready

The SDK is designed for AI coding assistants:
- Strong TypeScript types for autocomplete
- JSDoc with examples in every method
- MCP server for native AI tool integration
- Example files for AI training

## Common Tasks

### Generate Types from OpenAPI Spec

```bash
cd packages/typescript
npm run generate
```

This fetches the latest OpenAPI spec and regenerates TypeScript types.

### Build the SDK

```bash
npm run build
```

Outputs to `dist/` with CJS, ESM, and TypeScript definitions.

### Add a New Resource

1. Create `packages/typescript/src/resources/new-resource.ts`
2. Follow the pattern in `quotes.ts` or `shipments.ts`
3. Export it from `src/index.ts`
4. Add JSDoc examples for AI agents

### Add a New Language SDK

1. Create `packages/{language}/` directory
2. Set up code generation (oapi-codegen for Go, openapi-generator for Java)
3. Add wrapper layer on top of generated code
4. Update root `package.json` scripts

## OpenAPI Spec

The spec is the single source of truth at `openapi/spec.json`.

**Update it:**
```bash
curl https://rest-api.oway.io/api-docs/all-v1 -o openapi/spec.json
```

**Key endpoints:**
- `POST /v1/shipper/quote` - Get quote
- `POST /v1/shipper/shipment` - Create shipment
- `PUT /v1/shipper/shipment/{orderNumber}/confirm` - Confirm
- `GET /v1/shipper/shipment/{orderNumber}/tracking` - Track

## Development Workflow

1. **Update OpenAPI spec** (if API changed)
2. **Regenerate types**: `npm run generate:typescript`
3. **Update wrapper code** if needed (resources, client logic)
4. **Build**: `npm run build`
5. **Test**: `npm run test`
6. **Publish**: Update version, then `npm publish`

## Publishing

Each language SDK is independently versioned:
- TypeScript: `@oway/sdk` on npm
- Go: `github.com/Oway-Inc/oway-sdk-go` (Go modules)
- Java: `io.oway:oway-sdk` (Maven Central)

## AI Agent Integration

See `packages/typescript/AI_AGENT_GUIDE.md` for details on:
- GitHub Copilot optimization
- MCP server setup for Claude
- OpenAI GPT actions
- LangChain/LlamaIndex tools

## References

- **API Docs**: https://docs.shipoway.com
- **OpenAPI Spec**: https://rest-api.oway.io/api-docs/all-v1
- **Main Repo**: https://github.com/Oway-Inc/oway-sdk

## Important Notes

- Never commit API keys to the repo
- Always regenerate types after updating OpenAPI spec
- Keep wrapper layer thin - most logic should be in resources
- Write JSDoc examples - they train AI coding assistants

## Terminology

- **"Shipment"** not "order" in customer-facing text (though API uses orderNumber)

## Security Best Practices

1. **Never log sensitive data:**
   ```typescript
   // ✅ Good
   logger.info('Token refreshed', { apiKeyPrefix: key.substring(0, 12) + '...' });

   // ❌ Bad
   logger.info('Token refreshed', { apiKey: key });
   ```

2. **Sanitize error messages:**
   - Error responses might contain PII
   - Log request IDs, not full error bodies
   - Use structured logging with safe metadata

3. **Token management:**
   - Cache tokens with 5-minute buffer before expiry
   - Use mutex/locks for concurrent token refresh
   - Never refresh on every request (thundering herd)
