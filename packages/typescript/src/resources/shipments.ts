import type { HttpClient } from '../client';
import type { ShipmentRequest, Shipment, Tracking, Invoice } from '../types';

export class Shipments {
  constructor(private client: HttpClient) {}

  async create(params: ShipmentRequest, companyApiKey?: string): Promise<Shipment> {
    return this.client.post<Shipment>('/v1/shipper/shipment', params, companyApiKey);
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

  async tracking(orderNumber: string, companyApiKey?: string): Promise<Tracking> {
    return this.client.get<Tracking>(`/v1/shipper/shipment/${orderNumber}/tracking`, undefined, companyApiKey);
  }

  async document(orderNumber: string, documentType: 'BOL' | 'INVOICE' | 'LABEL', companyApiKey?: string): Promise<{ url: string }> {
    return this.client.get<{ url: string }>(`/v1/shipper/shipment/${orderNumber}/document/${documentType}`, undefined, companyApiKey);
  }

  async invoice(orderNumber: string, companyApiKey?: string): Promise<Invoice> {
    return this.client.get<Invoice>(`/v1/shipper/shipment/${orderNumber}/invoice`, undefined, companyApiKey);
  }
}
