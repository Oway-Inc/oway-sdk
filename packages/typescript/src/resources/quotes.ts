import type { HttpClient } from '../client';
import type { QuoteRequest, Quote } from '../types';

export class Quotes {
  constructor(private client: HttpClient) {}

  /**
   * Request a shipping quote
   * @param params Quote parameters
   * @param companyApiKey Optional: Specify company API key for multi-tenant integrations
   */
  async create(params: QuoteRequest, companyApiKey?: string): Promise<Quote> {
    return this.client.post<Quote>('/v1/shipper/quote', params, companyApiKey);
  }

  /**
   * Retrieve a quote by ID
   * @param quoteId Quote ID
   * @param companyApiKey Optional: Specify company API key for multi-tenant integrations
   */
  async retrieve(quoteId: string, companyApiKey?: string): Promise<Quote> {
    return this.client.get<Quote>(`/v1/shipper/quote/${quoteId}`, undefined, companyApiKey);
  }
}
