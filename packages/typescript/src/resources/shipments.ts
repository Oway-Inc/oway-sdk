import { randomId, type HttpClient } from '../client';
import type { ShipmentRequest, Shipment, Tracking, Invoice } from '../types';

export interface CreateShipmentOptions {
  /**
   * Sent as the Idempotency-Key header. Sending the same request with the same key within 24
   * hours returns the original shipment instead of creating another; reusing the key with a
   * different request is refused with 409 `idempotency_key_conflict`. Derive it from your own
   * stable identity for the shipment so a retry after a lost response is recognized. Defaults to
   * a new key per call, as does an empty string.
   */
  idempotencyKey?: string;
}

export class Shipments {
  constructor(private client: HttpClient) {}

  /**
   * Create a shipment. Every attempt, the SDK's own retries included, sends one Idempotency-Key,
   * so a retry never creates a second shipment.
   *
   * @example
   * const shipment = await oway.shipments.create(request, undefined, { idempotencyKey: 'order-4711' });
   */
  async create(
    params: ShipmentRequest,
    companyApiKey?: string,
    options: CreateShipmentOptions = {}
  ): Promise<Shipment> {
    const idempotencyKey = options.idempotencyKey || randomId();
    return this.client.post<Shipment>('/v1/shipper/shipment', params, companyApiKey, {
      'Idempotency-Key': idempotencyKey,
    });
  }

  async retrieve(orderNumber: string, companyApiKey?: string): Promise<Shipment> {
    return this.client.get<Shipment>(`/v1/shipper/shipment/${orderNumber}`, undefined, companyApiKey);
  }

  async confirm(orderNumber: string, companyApiKey?: string): Promise<Shipment> {
    return this.client.put<Shipment>(`/v1/shipper/shipment/${orderNumber}/confirm`, undefined, companyApiKey);
  }

  async cancel(orderNumber: string, companyApiKey?: string): Promise<void> {
    return this.client.put<void>(`/v1/shipper/shipment/${orderNumber}/cancel`, undefined, companyApiKey);
  }

  /**
   * Track a shipment. The live position estimate (GPS center, uncertainty
   * radius, last-event time, delay flags) is included by default. Pass
   * `{ includeLocation: false }` for lightweight status-only polling.
   *
   * @param orderNumber The 5-character PRO number.
   * @param opts.includeLocation Embed the position estimate. Defaults to true.
   * @param companyApiKey Optional per-request company API key.
   *
   * @example
   * // Default: includes the live position estimate
   * const tracking = await oway.shipments.tracking('ZKYQ5');
   * console.log(tracking.location?.center, tracking.orderStatus);
   *
   * @example
   * // Lightweight status-only polling (skips the position computation)
   * const status = await oway.shipments.tracking('ZKYQ5', { includeLocation: false });
   */
  async tracking(
    orderNumber: string,
    opts: { includeLocation?: boolean } = {},
    companyApiKey?: string,
  ): Promise<Tracking> {
    const query = (opts.includeLocation ?? true) ? { include: 'location' } : undefined;
    return this.client.get<Tracking>(
      `/v1/shipper/shipment/${orderNumber}/tracking`,
      query,
      companyApiKey,
    );
  }

  async document(orderNumber: string, documentType: 'BOL' | 'INVOICE' | 'LABEL', companyApiKey?: string): Promise<{ url: string }> {
    return this.client.get<{ url: string }>(`/v1/shipper/shipment/${orderNumber}/document/${documentType}`, undefined, companyApiKey);
  }

  async invoice(orderNumber: string, companyApiKey?: string): Promise<Invoice> {
    return this.client.get<Invoice>(`/v1/shipper/shipment/${orderNumber}/invoice`, undefined, companyApiKey);
  }
}
