# AI Agent & Copilot Guide

This SDK is designed to work seamlessly with AI coding assistants and agents.

## GitHub Copilot / Cursor / Codeium

The SDK is optimized for AI assistants through:
- ✅ Strong TypeScript types
- ✅ Resource-based API (`oway.quotes.create()`)
- ✅ Self-documenting method names

### Tips

**Start with a comment:**
```typescript
// Get a quote for shipping from LA to NYC for 500 lbs
```
Copilot will suggest the full API call.

**Use type imports:**
```typescript
import type { Quote, QuoteRequest } from '@oway/sdk';

const request: QuoteRequest = {
  // Copilot autocompletes all fields
```

## Claude Code / Aider

Claude can read the SDK source and understand the API:

```typescript
import Oway from '@oway/sdk';

const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  apiKey: process.env.OWAY_API_KEY, // Optional default
});

// Claude understands the full API
const quote = await oway.quotes.create({...});
```

## Model Context Protocol (MCP)

Expose Oway as native AI tools for Claude Desktop, etc.

### MCP Server Setup

**Add to `~/.claude/claude_desktop_config.json`:**
```json
{
  "mcpServers": {
    "oway": {
      "command": "npx",
      "args": ["@oway/sdk-mcp-server"],
      "env": {
        "OWAY_M2M_CLIENT_ID": "client_...",
        "OWAY_M2M_CLIENT_SECRET": "secret_...",
        "OWAY_API_KEY": "oway_sk_..."
      }
    }
  }
}
```

### Available MCP Tools

- `oway_get_quote` - Get shipping quotes
- `oway_create_shipment` - Schedule shipments
- `oway_track_shipment` - Track shipments
- `oway_confirm_shipment` - Confirm scheduled shipments
- `oway_cancel_shipment` - Cancel shipments

### Usage

```
Human: Get me a shipping quote from Los Angeles to New York for 500 pounds

Claude: I'll use the oway_get_quote tool...
[Executes API automatically]

The quote is $450 (Quote ID: quote_abc123). Would you like me to schedule this shipment?
```

## OpenAI GPTs / Custom GPTs

### Create an Oway Shipping GPT

1. Go to OpenAI → Create a GPT
2. Add Actions (import OpenAPI spec):
   ```
   https://rest-api.oway.io/api-docs/all-v1
   ```
3. Configure Authentication:
   - M2M credentials in GPT settings
   - Company API key per conversation

## LangChain / LlamaIndex

```typescript
import { Oway } from '@oway/sdk';
import { DynamicTool } from 'langchain/tools';

const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
});

const getQuoteTool = new DynamicTool({
  name: 'get_shipping_quote',
  description: 'Get freight quote. Input: JSON with origin, destination, items',
  func: async (input: string) => {
    const params = JSON.parse(input);
    const quote = await oway.quotes.create(params);
    return JSON.stringify(quote);
  },
});
```

## Best Practices for AI Agents

### 1. Add Examples

Create `examples/` directory with real use cases - AI tools learn from these:

```typescript
// examples/quote-and-schedule.ts
import Oway from '@oway/sdk';

const oway = new Oway({
  clientId: process.env.OWAY_M2M_CLIENT_ID!,
  clientSecret: process.env.OWAY_M2M_CLIENT_SECRET!,
  apiKey: 'oway_sk_test_123',
});

const quote = await oway.quotes.create({
  origin: { zipCode: '90210', country: 'US' },
  destination: { zipCode: '10001', country: 'US' },
  items: [{ weight: 500, weightUnit: 'LBS' }],
});
```

### 2. Use Clean Type Names

```typescript
import type { Quote, Shipment } from '@oway/sdk';
// AI understands these better than complex paths
```

### 3. Structured Error Handling

```typescript
try {
  await oway.quotes.create(params);
} catch (error) {
  if (error instanceof OwayError) {
    console.log(error.code, error.requestId);
    // AI can parse structured errors
  }
}
```

## Resources

- **MCP Protocol**: [modelcontextprotocol.io](https://modelcontextprotocol.io)
- **GitHub Copilot**: [docs.github.com/copilot](https://docs.github.com/copilot)
- **OpenAI GPTs**: [platform.openai.com/docs/actions](https://platform.openai.com/docs/actions)
