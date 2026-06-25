import { describe, it, expect, vi } from 'vitest';
import { Shipments } from '../resources/shipments';
import type { HttpClient } from '../client';

describe('Shipments.tracking', () => {
  it('sends include=location by default', async () => {
    const mockClient = { get: vi.fn().mockResolvedValue({}) } as unknown as HttpClient;
    const shipments = new Shipments(mockClient);

    await shipments.tracking('ZKYQ5');

    expect(mockClient.get).toHaveBeenCalledWith(
      '/v1/shipper/shipment/ZKYQ5/tracking',
      { include: 'location' },
      undefined
    );
  });

  it('omits the include query when includeLocation is false', async () => {
    const mockClient = { get: vi.fn().mockResolvedValue({}) } as unknown as HttpClient;
    const shipments = new Shipments(mockClient);

    await shipments.tracking('ZKYQ5', { includeLocation: false });

    expect(mockClient.get).toHaveBeenCalledWith(
      '/v1/shipper/shipment/ZKYQ5/tracking',
      undefined,
      undefined
    );
  });

  it('forwards a per-company API key', async () => {
    const mockClient = { get: vi.fn().mockResolvedValue({}) } as unknown as HttpClient;
    const shipments = new Shipments(mockClient);

    await shipments.tracking('ZKYQ5', {}, 'oway_sk_company_xyz');

    expect(mockClient.get).toHaveBeenCalledWith(
      '/v1/shipper/shipment/ZKYQ5/tracking',
      { include: 'location' },
      'oway_sk_company_xyz'
    );
  });
});
