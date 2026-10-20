import { describe, expect, it, vi } from 'vitest';
import { Quotes } from '../resources/quotes';
import { Shipments } from '../resources/shipments';
import type { HttpClient } from '../client';
import type { HazmatPaper, QuoteRequest, ShipmentRequest } from '../types';

describe('hazmat capture', () => {
  it('preserves partial rows, explicit false and units in quote and shipment requests', async () => {
    const paper: HazmatPaper = {
      entries: [
        { entryId: 'material-one', sourceItemReference: 'item-1', limitedQuantity: false },
        { entryId: 'material-two', totalQuantity: 12.5, quantityUnit: 'kg' },
      ],
    };
    const quote: QuoteRequest = {
      pickupAddress: { zipCode: '10001' }, deliveryAddress: { zipCode: '90001' },
      orderComponents: [{ palletCount: 1, poundsWeight: 400 }],
      hazmat: true, hazmatPaper: paper,
    };
    const shipment: ShipmentRequest = { ...quote, description: 'Example mixed load' };
    const post = vi.fn().mockResolvedValue({});
    const client = { post } as unknown as HttpClient;
    await new Quotes(client).create(quote);
    await new Shipments(client).create(shipment);
    for (const [, body] of post.mock.calls) {
      expect(body.hazmatPaper).toEqual(paper);
      expect(body.hazmatPaper.certification).toBeUndefined();
      expect(body.hazmatPaper.entries[0].packageCount).toBeUndefined();
    }
  });

  it('still accepts legacy request shapes without hazmat fields', async () => {
    const legacy: QuoteRequest = {
      pickupAddress: { zipCode: '10001' }, deliveryAddress: { zipCode: '90001' },
      orderComponents: [{ palletCount: 1, poundsWeight: 400 }],
    };
    const post = vi.fn().mockResolvedValue({});
    await new Quotes({ post } as unknown as HttpClient).create(legacy);
    expect(post.mock.calls[0][1]).not.toHaveProperty('hazmat');
  });
});
