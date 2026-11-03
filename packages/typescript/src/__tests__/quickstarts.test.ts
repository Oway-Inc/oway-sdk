import { describe, expect, it, vi } from 'vitest';
import type Oway from '../index';
import type { QuoteRequest, ShipmentRequest } from '../index';
import { shipperQuickstart } from '../../examples/quote-and-schedule';

function client(status = 'CONFIRMED') {
  return {
    quotes: { create: vi.fn().mockResolvedValue({ id: 'quote-example' }) },
    shipments: {
      create: vi.fn().mockResolvedValue({ orderNumber: 'ABCDE' }),
      confirm: vi.fn().mockResolvedValue({ orderStatus: status }),
      tracking: vi.fn().mockResolvedValue({ orderStatus: status }),
      document: vi.fn().mockResolvedValue({ url: 'https://example.com/bol.pdf' }),
    },
  };
}

describe('shipper quickstart', () => {
  it('saves the exact body and key before an uncertain create failure', async () => {
    const api = client(); const save = vi.fn();
    api.shipments.create.mockImplementation(async () => {
      expect(save).toHaveBeenCalledWith({ request: { quoteId: 'quote-example' }, idempotencyKey: 'run-example' });
      throw new Error('connection lost');
    });
    await expect(shipperQuickstart(api as unknown as Oway, {} as QuoteRequest, {} as ShipmentRequest, 'run-example', save)).rejects.toThrow('connection lost');
    expect(api.shipments.confirm).not.toHaveBeenCalled();
  });
  it('carries returned identities and replays the exact creation before confirming', async () => {
    const api = client();
    expect(await shipperQuickstart(api as unknown as Oway, {} as QuoteRequest, { description: 'Synthetic freight' } as ShipmentRequest, 'run-example')).toBe('ABCDE');
    expect(api.shipments.create.mock.calls[0]).toEqual(api.shipments.create.mock.calls[1]);
    expect(api.shipments.create).toHaveBeenCalledWith({ description: 'Synthetic freight', quoteId: 'quote-example' }, undefined, { idempotencyKey: 'run-example' });
    expect(api.shipments.confirm).toHaveBeenCalledWith('ABCDE');
    expect(api.shipments.document).toHaveBeenCalledWith('ABCDE', 'BILL_OF_LADING');
  });

  it('does not report HTTP-successful but unconfirmed shipments as complete', async () => {
    const api = client('INITIALIZED');
    await expect(shipperQuickstart(api as unknown as Oway, {} as QuoteRequest, {} as ShipmentRequest, 'run-example')).rejects.toThrow('not CONFIRMED');
    expect(api.shipments.document).not.toHaveBeenCalled();
  });

  it('stops before confirmation when create replay does not match', async () => {
    const api = client();
    api.shipments.create.mockResolvedValueOnce({ orderNumber: 'ABCDE' }).mockResolvedValueOnce({ orderNumber: 'FGHIJ' });
    await expect(shipperQuickstart(api as unknown as Oway, {} as QuoteRequest, {} as ShipmentRequest, 'run-example')).rejects.toThrow('different shipment');
    expect(api.shipments.confirm).not.toHaveBeenCalled();
  });
});

import { prepareCarrier, deliverCarrier, type CarrierRun } from '../../examples/carrier-quickstart';

const run: CarrierRun = { offerId: 'offer-example', documentId: 'document-example', metadata: {
  document_type: 'PROOF_OF_DELIVERY', content_type: 'application/pdf', file_name: 'pod.pdf',
} };
function carrierClient() {
  return { carrier: {
    offer: vi.fn().mockResolvedValue({ status: 'PENDING' }),
    acceptOffer: vi.fn().mockResolvedValue({ status: 'ACCEPTED' }),
    confirmPickup: vi.fn().mockResolvedValue({ status: 'PICKED_UP' }),
    submitLocation: vi.fn().mockResolvedValue({}),
    requestUploadUrl: vi.fn().mockResolvedValue({ document_id: run.documentId, upload_url: 'https://example.com/upload', content_type: 'application/pdf' }),
    confirmUpload: vi.fn().mockResolvedValue({ document_id: run.documentId, review_status: 'PENDING' }),
    uploadedDocument: vi.fn().mockResolvedValue({ document_id: run.documentId, review_status: 'APPROVED' }),
    shipment: vi.fn().mockResolvedValue({ status: 'PICKED_UP' }),
    confirmDelivery: vi.fn().mockResolvedValue({ status: 'DELIVERED' }),
  } };
}
const location = { latitude: 35, longitude: -115, timestamp: '2026-11-03T12:00:00Z' };
describe('carrier quickstart', () => {
  it('saves returned identity before upload and registers only after successful PUT', async () => {
    const api = carrierClient();
    const events: string[] = [];
    const saved = vi.fn(() => { events.push('save'); });
    const upload = vi.fn(async () => { events.push('put'); expect(api.carrier.confirmUpload).not.toHaveBeenCalled(); });
    const result = await prepareCarrier(api as unknown as Oway, run.offerId, new Uint8Array([1]), run.metadata, location, saved, upload);
    expect(result).toEqual(run);
    expect(saved).toHaveBeenCalledWith(run);
    expect(events).toEqual(['save', 'put']);
    expect(api.carrier.confirmUpload).toHaveBeenCalledWith(run.offerId, run.documentId, run.metadata);
    expect(api.carrier.confirmDelivery).not.toHaveBeenCalled();
  });
  it('retains recovery identity and never confirms a failed storage PUT', async () => {
    const api = carrierClient(); const save = vi.fn();
    await expect(prepareCarrier(api as unknown as Oway, run.offerId, new Uint8Array([1]), run.metadata, location, save, async () => { throw new Error('upload failed'); })).rejects.toThrow('upload failed');
    expect(save).toHaveBeenCalledWith(run);
    expect(api.carrier.confirmUpload).not.toHaveBeenCalled();
  });
  it('does not repeat acceptance or pickup when recovering a picked-up offer', async () => {
    const api = carrierClient(); api.carrier.offer.mockResolvedValue({ status: 'PICKED_UP' });
    await prepareCarrier(api as unknown as Oway, run.offerId, new Uint8Array([1]), run.metadata, location, vi.fn(), async () => {});
    expect(api.carrier.acceptOffer).not.toHaveBeenCalled(); expect(api.carrier.confirmPickup).not.toHaveBeenCalled();
  });
  it('blocks pending review before any delivery side effects', async () => {
    const api = carrierClient(); api.carrier.uploadedDocument.mockResolvedValue({ document_id: run.documentId, review_status: 'PENDING' });
    await expect(deliverCarrier(api as unknown as Oway, run)).rejects.toThrow('review is pending');
    expect(api.carrier.confirmDelivery).not.toHaveBeenCalled();
  });
  it('uses the registered ID and never converts a PUT URL into proof of delivery', async () => {
    const api = carrierClient(); await deliverCarrier(api as unknown as Oway, run);
    expect(api.carrier.confirmDelivery).toHaveBeenCalledWith(run.offerId, { delivery_time: expect.any(String), proof_of_delivery_document_id: run.documentId });
    expect(api.carrier.uploadedDocument).toHaveBeenCalledTimes(2);
  });
  it('reconciles delivery already triggered by review without submitting it again', async () => {
    const api = carrierClient(); api.carrier.shipment.mockResolvedValue({ status: 'DELIVERED' });
    await deliverCarrier(api as unknown as Oway, run);
    expect(api.carrier.confirmDelivery).not.toHaveBeenCalled();
  });
});
