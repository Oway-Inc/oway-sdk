import { describe, it, expect, vi } from 'vitest';
import { Carrier } from '../resources/carrier';
import { OwayError, type HttpClient } from '../client';

function mockClient() {
  return {
    get: vi.fn().mockResolvedValue({}),
    post: vi.fn().mockResolvedValue({}),
    put: vi.fn().mockResolvedValue({}),
  } as unknown as HttpClient;
}

describe('Carrier Resource', () => {
  it('offers() lists carrier offers', async () => {
    const c = mockClient();
    await new Carrier(c).offers('oway_sk_co');
    expect(c.get).toHaveBeenCalledWith('/v1/carrier/offers', undefined, 'oway_sk_co');
  });

  it('offer(id) fetches one offer', async () => {
    const c = mockClient();
    await new Carrier(c).offer('off-1');
    expect(c.get).toHaveBeenCalledWith('/v1/carrier/offers/off-1', undefined, undefined);
  });

  it('acceptOffer posts to the accept path', async () => {
    const c = mockClient();
    await new Carrier(c).acceptOffer('off-1', { carrierReference: 'R1' } as any);
    expect(c.put).toHaveBeenCalledWith('/v1/carrier/offers/off-1/accept', { carrierReference: 'R1' }, undefined);
  });

  it('rejectOffer posts to the reject path', async () => {
    const c = mockClient();
    await new Carrier(c).rejectOffer('off-1', { reason: 'NO_CAPACITY' } as any);
    expect(c.put).toHaveBeenCalledWith('/v1/carrier/offers/off-1/reject', { reason: 'NO_CAPACITY' }, undefined);
  });

  it('shipment(id) fetches the carrier shipment', async () => {
    const c = mockClient();
    await new Carrier(c).shipment('off-1');
    expect(c.get).toHaveBeenCalledWith('/v1/carrier/shipments/off-1', undefined, undefined);
  });

  it('confirmPickup puts to the pickup path', async () => {
    const c = mockClient();
    await new Carrier(c).confirmPickup('off-1', { latitude: 1, longitude: 2 } as any);
    expect(c.put).toHaveBeenCalledWith('/v1/carrier/shipments/off-1/pickup', { latitude: 1, longitude: 2 }, undefined);
  });

  it('confirmDelivery puts to the deliver path', async () => {
    const c = mockClient();
    await new Carrier(c).confirmDelivery('off-1', { latitude: 1, longitude: 2 } as any);
    expect(c.put).toHaveBeenCalledWith('/v1/carrier/shipments/off-1/deliver', { latitude: 1, longitude: 2 }, undefined);
  });

  it('submitLocation posts a GPS update', async () => {
    const c = mockClient();
    await new Carrier(c).submitLocation('off-1', { latitude: 1, longitude: 2 } as any);
    expect(c.post).toHaveBeenCalledWith('/v1/carrier/shipments/off-1/location', { latitude: 1, longitude: 2 }, undefined);
  });

  it('reportException posts an exception', async () => {
    const c = mockClient();
    await new Carrier(c).reportException('off-1', { exceptionType: 'WEATHER' } as any);
    expect(c.post).toHaveBeenCalledWith('/v1/carrier/shipments/off-1/exception', { exceptionType: 'WEATHER' }, undefined);
  });

  it('submitTrips posts the trip array', async () => {
    const c = mockClient();
    await new Carrier(c).submitTrips([{ scac: 'OWAY' }] as any);
    expect(c.post).toHaveBeenCalledWith('/v1/carrier/coverage/trips', [{ scac: 'OWAY' }], undefined);
  });

  it('tracking(id) fetches carrier tracking', async () => {
    const c = mockClient();
    await new Carrier(c).tracking('off-1');
    expect(c.get).toHaveBeenCalledWith('/v1/carrier/shipments/off-1/tracking', undefined, undefined);
  });

  it('throws on empty identifier without calling the client', async () => {
    const c = mockClient();
    await expect(new Carrier(c).offer('')).rejects.toBeInstanceOf(OwayError);
    expect(c.get).not.toHaveBeenCalled();
  });

  it('throws on empty trips without calling the client', async () => {
    const c = mockClient();
    await expect(new Carrier(c).submitTrips([])).rejects.toBeInstanceOf(OwayError);
    expect(c.post).not.toHaveBeenCalled();
  });
});

describe('carrier uploaded documents', () => {
  const metadata = { document_type: 'PROOF_OF_DELIVERY' as const, content_type: 'application/pdf' as const, file_name: 'pod.pdf' };
  it('preserves snake_case metadata, company identity and encoded path segments', async () => {
    const c = mockClient(); const carrier = new Carrier(c);
    await carrier.requestUploadUrl('ref/one', metadata, 'company-key');
    expect(c.post).toHaveBeenCalledWith('/v1/carrier/shipments/ref%2Fone/upload', metadata, 'company-key');
    await carrier.confirmUpload('ref/one', 'document-id', metadata, 'company-key');
    expect(c.put).toHaveBeenCalledWith('/v1/carrier/shipments/ref%2Fone/uploads/document-id/confirm', metadata, 'company-key');
    await carrier.uploadedDocument('ref/one', 'document-id', 'company-key');
    expect(c.get).toHaveBeenCalledWith('/v1/carrier/shipments/ref%2Fone/documents/document-id', undefined, 'company-key');
  });
  it('rejects missing document identity before any request', async () => {
    const c = mockClient(); const carrier = new Carrier(c);
    await expect(carrier.confirmUpload('offer-id', '', metadata)).rejects.toBeInstanceOf(OwayError);
    await expect(carrier.uploadedDocument('offer-id', '')).rejects.toBeInstanceOf(OwayError);
    expect(c.put).not.toHaveBeenCalled(); expect(c.get).not.toHaveBeenCalled();
  });
});
