import { HttpClient, type OwayConfig } from './client';
import { Quotes } from './resources/quotes';
import { Shipments } from './resources/shipments';

/**
 * Official Oway SDK for JavaScript/TypeScript
 */
export class Oway {
  private client: HttpClient;

  public readonly quotes: Quotes;
  public readonly shipments: Shipments;

  constructor(config: OwayConfig) {
    this.client = new HttpClient(config);
    this.quotes = new Quotes(this.client);
    this.shipments = new Shipments(this.client);
  }
}

// Export clean types
export type {
  QuoteRequest,
  ShipmentRequest,
  Quote,
  Shipment,
  Tracking,
  Invoice,
  Address,
} from './types';

// Export errors and config
export { OwayError, type OwayConfig } from './client';

// Export generated types for advanced usage
export type { paths, components } from './generated/schema';

// Default export
export default Oway;

// Export environment constants
export { OwayEnvironments } from './environments';
