/**
 * Clean type aliases for the Oway SDK
 */

import type { paths, components } from './generated/schema';

// Request types
export type QuoteRequest = paths['/v1/shipper/quote']['post']['requestBody']['content']['application/json'];
export type ShipmentRequest = paths['/v1/shipper/shipment']['post']['requestBody']['content']['application/json'];

// Response types
export type Quote = paths['/v1/shipper/quote']['post']['responses']['200']['content']['application/json'];
export type Shipment = paths['/v1/shipper/shipment']['post']['responses']['200']['content']['application/json'];
export type Tracking = paths['/v1/shipper/shipment/{orderNumber}/tracking']['get']['responses']['200']['content']['application/json'];
export type Invoice = paths['/v1/shipper/shipment/{orderNumber}/invoice']['get']['responses']['200']['content']['application/json'];

// Common types
export type Address = components['schemas']['Address'];

// Carrier-side types
export type Offer = components['schemas']['Offer'];
export type AcceptOfferRequest = components['schemas']['AcceptOfferRequest'];
export type RejectOfferRequest = components['schemas']['RejectOfferRequest'];
export type CarrierShipment = components['schemas']['CarrierShipment'];
export type PickupConfirmationRequest = components['schemas']['PickupConfirmationRequest'];
export type DeliveryConfirmationRequest = components['schemas']['DeliveryConfirmationRequest'];
export type OfferLocationUpdate = components['schemas']['OfferLocationUpdate'];
export type LocationAcknowledgment = components['schemas']['LocationAcknowledgment'];
export type ExceptionReportRequest = components['schemas']['ExceptionReportRequest'];
export type ExceptionResponse = components['schemas']['ExceptionResponse'];
export type CarrierTripRequest = components['schemas']['CarrierTripRequest'];
export type CarrierTracking = components['schemas']['CarrierTracking'];

// Re-export for advanced usage
export type { paths, components } from './generated/schema';
