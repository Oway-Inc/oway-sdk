/**
 * Model Context Protocol (MCP) Server for Oway SDK
 *
 * Enables AI agents to interact with Oway API through natural language.
 *
 * Setup in ~/.claude/claude_desktop_config.json:
 * {
 *   "mcpServers": {
 *     "oway": {
 *       "command": "npx",
 *       "args": ["@oway/sdk-mcp-server"],
 *       "env": {
 *         "OWAY_M2M_CLIENT_ID": "client_...",
 *         "OWAY_M2M_CLIENT_SECRET": "secret_...",
 *         "OWAY_API_KEY": "oway_sk_..." 
 *       }
 *     }
 *   }
 * }
 */

import { Oway, OwayError } from '../index';

interface McpTool {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, any>;
    required?: string[];
  };
}

export class OwayMcpServer {
  private oway: Oway;

  constructor(config: { clientId: string; clientSecret: string; apiKey?: string }) {
    this.oway = new Oway({
      clientId: config.clientId,
      clientSecret: config.clientSecret,
      apiKey: config.apiKey,
    });
  }

  getTools(): McpTool[] {
    return [
      {
        name: 'oway_get_quote',
        description: 'Get a freight shipping quote. Returns quote ID and price.',
        inputSchema: {
          type: 'object',
          properties: {
            origin: {
              type: 'object',
              properties: {
                zipCode: { type: 'string' },
                city: { type: 'string' },
                state: { type: 'string' },
                country: { type: 'string' },
              },
              required: ['zipCode', 'country'],
            },
            destination: {
              type: 'object',
              properties: {
                zipCode: { type: 'string' },
                city: { type: 'string' },
                state: { type: 'string' },
                country: { type: 'string' },
              },
              required: ['zipCode', 'country'],
            },
            items: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  weight: { type: 'number' },
                  weightUnit: { type: 'string', enum: ['LBS', 'KG'] },
                },
              },
            },
            companyApiKey: {
              type: 'string',
              description: 'Optional: Company API key for multi-tenant integrations',
            },
          },
          required: ['origin', 'destination', 'items'],
        },
      },
      {
        name: 'oway_create_shipment',
        description: 'Schedule a shipment. Returns order number.',
        inputSchema: {
          type: 'object',
          properties: {
            quoteId: { type: 'string' },
            pickup: {
              type: 'object',
              properties: {
                contactName: { type: 'string' },
                contactPhone: { type: 'string' },
              },
            },
            delivery: {
              type: 'object',
              properties: {
                contactName: { type: 'string' },
                contactPhone: { type: 'string' },
              },
            },
            companyApiKey: { type: 'string', description: 'Optional: Company API key' },
          },
          required: ['quoteId', 'pickup', 'delivery'],
        },
      },
      {
        name: 'oway_track_shipment',
        description: 'Get tracking status for a shipment.',
        inputSchema: {
          type: 'object',
          properties: {
            orderNumber: { type: 'string' },
            companyApiKey: { type: 'string', description: 'Optional: Company API key' },
          },
          required: ['orderNumber'],
        },
      },
    ];
  }

  async executeTool(toolName: string, args: any): Promise<any> {
    try {
      const companyApiKey = args.companyApiKey;
      delete args.companyApiKey; // Remove from params

      switch (toolName) {
        case 'oway_get_quote':
          return await this.oway.quotes.create(args, companyApiKey);

        case 'oway_create_shipment':
          return await this.oway.shipments.create(args, companyApiKey);

        case 'oway_track_shipment':
          return await this.oway.shipments.tracking(args.orderNumber, companyApiKey);

        default:
          throw new Error(`Unknown tool: ${toolName}`);
      }
    } catch (error) {
      if (error instanceof OwayError) {
        return {
          error: error.message,
          code: error.code,
          requestId: error.requestId,
        };
      }
      throw error;
    }
  }
}

export async function startMcpServer() {
  const clientId = process.env.OWAY_M2M_CLIENT_ID;
  const clientSecret = process.env.OWAY_M2M_CLIENT_SECRET;
  const apiKey = process.env.OWAY_API_KEY;

  if (!clientId || !clientSecret) {
    console.error('OWAY_M2M_CLIENT_ID and OWAY_M2M_CLIENT_SECRET required');
    process.exit(1);
  }

  const server = new OwayMcpServer({ clientId, clientSecret, apiKey });

  process.stdin.on('data', async (data) => {
    try {
      const request = JSON.parse(data.toString());

      if (request.method === 'tools/list') {
        console.log(JSON.stringify({
          jsonrpc: '2.0',
          id: request.id,
          result: { tools: server.getTools() },
        }));
      } else if (request.method === 'tools/call') {
        const { name, arguments: args } = request.params;
        const result = await server.executeTool(name, args);
        console.log(JSON.stringify({
          jsonrpc: '2.0',
          id: request.id,
          result: { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] },
        }));
      }
    } catch (error) {
      console.error('MCP server error:', error);
    }
  });
}
