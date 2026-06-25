import { OwayError, type HttpClient } from '../client';
import type {
  Offer,
  AcceptOfferRequest,
  RejectOfferRequest,
  CarrierShipment,
  PickupConfirmationRequest,
  DeliveryConfirmationRequest,
  OfferLocationUpdate,
  LocationAcknowledgment,
  ExceptionReportRequest,
  ExceptionResponse,
  CarrierTripRequest,
  CarrierTracking,
} from '../types';

function requireIdentifier(id: string): void {
  if (!id) {
    throw new OwayError({ message: 'identifier is required', code: 'CARRIER_MISSING_IDENTIFIER' });
  }
}

/**
 * Carrier-side API: discover and respond to offers, drive a shipment
 * through pickup and delivery, stream GPS positions, report exceptions,
 * and submit completed trips. Mirrors the shipper resources' thin-wrapper
 * style; auth, retries, and error handling come from the shared client.
 */
export class Carrier {
  constructor(private client: HttpClient) {}

  /** List all pending offers visible to the authenticated carrier. */
  async offers(companyApiKey?: string): Promise<Offer[]> {
    return this.client.get<Offer[]>('/v1/carrier/offers', undefined, companyApiKey);
  }

  /** Fetch a single offer by identifier. */
  async offer(identifier: string, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.get<Offer>(`/v1/carrier/offers/${identifier}`, undefined, companyApiKey);
  }

  /** Accept a pending offer. */
  async acceptOffer(identifier: string, req: AcceptOfferRequest, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.put<Offer>(`/v1/carrier/offers/${identifier}/accept`, req, companyApiKey);
  }

  /** Reject a pending offer with a typed reason. */
  async rejectOffer(identifier: string, req: RejectOfferRequest, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.put<Offer>(`/v1/carrier/offers/${identifier}/reject`, req, companyApiKey);
  }

  /** Fetch the carrier-view shipment for an accepted offer. */
  async shipment(identifier: string, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.get<CarrierShipment>(`/v1/carrier/shipments/${identifier}`, undefined, companyApiKey);
  }

  /** Confirm pickup for a shipment. */
  async confirmPickup(identifier: string, req: PickupConfirmationRequest, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.put<CarrierShipment>(`/v1/carrier/shipments/${identifier}/pickup`, req, companyApiKey);
  }

  /** Confirm delivery for a shipment. */
  async confirmDelivery(identifier: string, req: DeliveryConfirmationRequest, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.put<CarrierShipment>(`/v1/carrier/shipments/${identifier}/deliver`, req, companyApiKey);
  }

  /** Post a GPS location update for an in-progress shipment. */
  async submitLocation(identifier: string, loc: OfferLocationUpdate, companyApiKey?: string): Promise<LocationAcknowledgment> {
    requireIdentifier(identifier);
    return this.client.post<LocationAcknowledgment>(`/v1/carrier/shipments/${identifier}/location`, loc, companyApiKey);
  }

  /** Report an exception (weather, mechanical, traffic, shipper delay) during transit. */
  async reportException(identifier: string, req: ExceptionReportRequest, companyApiKey?: string): Promise<ExceptionResponse> {
    requireIdentifier(identifier);
    return this.client.post<ExceptionResponse>(`/v1/carrier/shipments/${identifier}/exception`, req, companyApiKey);
  }

  /** Submit one or more completed trip summaries to the carrier coverage API. */
  async submitTrips(trips: CarrierTripRequest[], companyApiKey?: string): Promise<void> {
    if (!trips || trips.length === 0) {
      throw new OwayError({ message: 'at least one trip is required', code: 'CARRIER_EMPTY_TRIPS' });
    }
    return this.client.post<void>('/v1/carrier/coverage/trips', trips, companyApiKey);
  }

  /** Get GPS tracking history for a carrier shipment. */
  async tracking(identifier: string, companyApiKey?: string): Promise<CarrierTracking> {
    requireIdentifier(identifier);
    return this.client.get<CarrierTracking>(`/v1/carrier/shipments/${identifier}/tracking`, undefined, companyApiKey);
  }
}
