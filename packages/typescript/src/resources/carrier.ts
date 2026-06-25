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

  /**
   * List all pending offers visible to the authenticated carrier.
   *
   * @example
   * const offers = await oway.carrier.offers();
   * console.log(`${offers.length} pending offers`);
   */
  async offers(companyApiKey?: string): Promise<Offer[]> {
    return this.client.get<Offer[]>('/v1/carrier/offers', undefined, companyApiKey);
  }

  /**
   * Fetch a single offer by identifier.
   *
   * @example
   * const offer = await oway.carrier.offer('off-123');
   */
  async offer(identifier: string, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.get<Offer>(`/v1/carrier/offers/${identifier}`, undefined, companyApiKey);
  }

  /**
   * Accept a pending offer.
   *
   * @example
   * const accepted = await oway.carrier.acceptOffer('off-123', {
   *   carrier_reference: 'YOUR-REF-123',
   *   driver_name: 'Jane Driver',
   * });
   */
  async acceptOffer(identifier: string, req: AcceptOfferRequest, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.put<Offer>(`/v1/carrier/offers/${identifier}/accept`, req, companyApiKey);
  }

  /**
   * Reject a pending offer with a typed reason.
   *
   * @example
   * await oway.carrier.rejectOffer('off-123', { reason: 'capacity_unavailable' });
   */
  async rejectOffer(identifier: string, req: RejectOfferRequest, companyApiKey?: string): Promise<Offer> {
    requireIdentifier(identifier);
    return this.client.put<Offer>(`/v1/carrier/offers/${identifier}/reject`, req, companyApiKey);
  }

  /**
   * Fetch the carrier-view shipment for an accepted offer.
   *
   * @example
   * const shipment = await oway.carrier.shipment('off-123');
   */
  async shipment(identifier: string, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.get<CarrierShipment>(`/v1/carrier/shipments/${identifier}`, undefined, companyApiKey);
  }

  /**
   * Confirm pickup for a shipment.
   *
   * @example
   * await oway.carrier.confirmPickup('off-123', {
   *   coordinates: { latitude: 34.05, longitude: -118.24 },
   * });
   */
  async confirmPickup(identifier: string, req: PickupConfirmationRequest, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.put<CarrierShipment>(`/v1/carrier/shipments/${identifier}/pickup`, req, companyApiKey);
  }

  /**
   * Confirm delivery for a shipment.
   *
   * @example
   * await oway.carrier.confirmDelivery('off-123', {
   *   coordinates: { latitude: 40.71, longitude: -74.0 },
   *   signed_by: 'J. Recipient',
   * });
   */
  async confirmDelivery(identifier: string, req: DeliveryConfirmationRequest, companyApiKey?: string): Promise<CarrierShipment> {
    requireIdentifier(identifier);
    return this.client.put<CarrierShipment>(`/v1/carrier/shipments/${identifier}/deliver`, req, companyApiKey);
  }

  /**
   * Post a GPS location update for an in-progress shipment.
   *
   * @example
   * // Call on your GPS update interval while in transit
   * await oway.carrier.submitLocation('off-123', { latitude: 36.17, longitude: -115.14 });
   */
  async submitLocation(identifier: string, loc: OfferLocationUpdate, companyApiKey?: string): Promise<LocationAcknowledgment> {
    requireIdentifier(identifier);
    return this.client.post<LocationAcknowledgment>(`/v1/carrier/shipments/${identifier}/location`, loc, companyApiKey);
  }

  /**
   * Report an exception (weather, mechanical, traffic, shipper delay) during transit.
   *
   * @example
   * await oway.carrier.reportException('off-123', { exception_type: 'weather_delay' });
   */
  async reportException(identifier: string, req: ExceptionReportRequest, companyApiKey?: string): Promise<ExceptionResponse> {
    requireIdentifier(identifier);
    return this.client.post<ExceptionResponse>(`/v1/carrier/shipments/${identifier}/exception`, req, companyApiKey);
  }

  /**
   * Submit one or more completed trip summaries to the carrier coverage API.
   *
   * @example
   * await oway.carrier.submitTrips([{ trip_no: 'TRIP-001', legs: [] }]);
   */
  async submitTrips(trips: CarrierTripRequest[], companyApiKey?: string): Promise<void> {
    if (!trips || trips.length === 0) {
      throw new OwayError({ message: 'at least one trip is required', code: 'CARRIER_EMPTY_TRIPS' });
    }
    return this.client.post<void>('/v1/carrier/coverage/trips', trips, companyApiKey);
  }

  /**
   * Get GPS tracking history for a carrier shipment.
   *
   * @example
   * const history = await oway.carrier.tracking('off-123');
   * console.log(`${history.points?.length ?? 0} GPS points`);
   */
  async tracking(identifier: string, companyApiKey?: string): Promise<CarrierTracking> {
    requireIdentifier(identifier);
    return this.client.get<CarrierTracking>(`/v1/carrier/shipments/${identifier}/tracking`, undefined, companyApiKey);
  }
}
