import { describe, it, expect, vi } from 'vitest';
import { Quotes } from '../resources/quotes';
import type { HttpClient } from '../client';

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
});
