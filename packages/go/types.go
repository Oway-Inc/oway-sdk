package oway

import "github.com/Oway-Inc/oway-sdk/packages/go/client"

// Clean type aliases for public API

// Request types
type (
	QuoteRequest    = client.QuoteRequest
	ShipmentRequest = client.CreateShipmentRequest
)

// Response types
type (
	Quote    = client.QuoteResponse
	Shipment = client.Shipment
	Tracking = client.Tracking
	Invoice  = client.InvoiceResponse
)

// Common types
type (
	Address        = client.Address
	OrderComponent = client.OrderComponent
	Document       = client.DocumentResponse
	DocumentType   = client.GetDocumentByOrderNumberParamsDocumentType
)

// Document type constants
const (
	DocumentTypeBOL           DocumentType = "BILL_OF_LADING"
	DocumentTypeInvoice       DocumentType = "INVOICE"
	DocumentTypeShippingLabel DocumentType = "SHIPPING_LABEL"
)

// Additional aliases for EDI Gateway compatibility
type (
	CreateShipmentRequest = ShipmentRequest
)
