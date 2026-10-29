import { describe, it, expect, vi } from 'vitest';
import { Quotes } from '../resources/quotes';
import type { HttpClient } from '../client';
import type { Quote, QuoteStopAccessorials } from '../types';

describe('Quotes Resource', () => {
  it('should call POST with correct path and params', async () => {
    const mockClient = {
      post: vi.fn().mockResolvedValue({ quoteId: 'quote_123' }),
    } as unknown as HttpClient;

    const quotes = new Quotes(mockClient);
    await quotes.create({ origin: {}, destination: {}, items: [] } as any);

    expect(mockClient.post).toHaveBeenCalledWith(
      '/v1/shipper/quote',
      expect.any(Object),
      undefined
    );
  });

  it('should support per-company API key', async () => {
    const mockClient = {
      post: vi.fn().mockResolvedValue({ quoteId: 'quote_123' }),
    } as unknown as HttpClient;

    const quotes = new Quotes(mockClient);
    await quotes.create({} as any, 'oway_sk_company_xyz');

    expect(mockClient.post).toHaveBeenCalledWith(
      '/v1/shipper/quote',
      expect.any(Object),
      'oway_sk_company_xyz'
    );
  });

  it('returns the accessorials each stop was priced for', async () => {
    const quote: Quote = {
      id: 'quote_123',
      quotedPriceInCents: 125000,
      transitDays: 2,
      estimatedDeliveryDate: '2026-10-31T21:00:00Z',
      pickupAccessorials: { appointmentRequired: false, liftgateRequired: false },
      deliveryAccessorials: { appointmentRequired: true, liftgateRequired: true },
    };
    const mockClient = {
      post: vi.fn().mockResolvedValue(quote),
    } as unknown as HttpClient;

    const result = await new Quotes(mockClient).create({} as any);

    const delivery: QuoteStopAccessorials | undefined = result.deliveryAccessorials;
    expect(delivery?.appointmentRequired).toBe(true);
    expect(result.pickupAccessorials?.appointmentRequired).toBe(false);
  });
});
