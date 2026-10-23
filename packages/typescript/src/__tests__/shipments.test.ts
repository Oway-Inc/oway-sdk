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

describe('Shipments.create', () => {
  const request = {} as Parameters<Shipments['create']>[0];

  it('sends a new idempotency key per call', async () => {
    const post = vi.fn().mockResolvedValue({});
    const shipments = new Shipments({ post } as unknown as HttpClient);

    await shipments.create(request);
    await shipments.create(request);

    const keys = post.mock.calls.map((call) => call[3]['Idempotency-Key']);
    expect(keys[0]).toMatch(/^[0-9a-f-]{32,36}$/);
    expect(keys[1]).not.toBe(keys[0]);
  });

  it("sends the caller's key and company API key", async () => {
    const post = vi.fn().mockResolvedValue({});
    await new Shipments({ post } as unknown as HttpClient).create(request, 'oway_sk_company_xyz', {
      idempotencyKey: 'order-4711',
    });

    expect(post).toHaveBeenCalledWith('/v1/shipper/shipment', request, 'oway_sk_company_xyz', {
      'Idempotency-Key': 'order-4711',
    });
  });
});
