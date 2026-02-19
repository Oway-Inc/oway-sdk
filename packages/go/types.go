package oway

import "github.com/Oway-Inc/oway-sdk/packages/go/client"

// Clean type aliases for public API

// Request types
type (
	QuoteRequest    = client.ExternalRequestQuoteRequest
	ShipmentRequest = client.ExternalCreateShipmentRequest
)

// Response types
type (
	Quote    = client.ExternalRequestQuoteResponse
	Shipment = client.ExternalOrder
	Tracking = client.ExternalTracking
	Invoice  = client.ExternalInvoiceResponse
)

// Common types
type (
	Address = client.ExternalAddress
)
