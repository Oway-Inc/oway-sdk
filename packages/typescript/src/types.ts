/**
 * Clean type aliases for the Oway SDK
 */

import type { paths, components } from './generated/schema';

// Request types
export type QuoteRequest = paths['/v1/shipper/quote']['post']['requestBody']['content']['application/json'];
export type ShipmentRequest = paths['/v1/shipper/shipment']['post']['requestBody']['content']['application/json'];

// Response types
export type Quote = paths['/v1/shipper/quote']['post']['responses']['200']['content']['application/hal+json'];
export type Shipment = paths['/v1/shipper/shipment']['post']['responses']['200']['content']['application/hal+json'];
export type Tracking = paths['/v1/shipper/shipment/{orderNumber}/tracking']['get']['responses']['200']['content']['application/hal+json'];
export type Invoice = paths['/v1/shipper/shipment/{orderNumber}/invoice']['get']['responses']['200']['content']['application/hal+json'];

// Common types
export type Address = components['schemas']['ExternalAddress'];

// Re-export for advanced usage
export type { paths, components } from './generated/schema';
