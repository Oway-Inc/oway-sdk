package oway

import (
	"time"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// Quote is the response payload returned by RequestQuote and GetQuote.
type Quote = client.QuoteResponse

// Shipment is the response payload returned by shipment endpoints.
type Shipment = client.Shipment

// Tracking is the response payload returned by TrackShipment.
type Tracking = client.Tracking

// Invoice is the response payload returned by GetInvoice.
type Invoice = client.InvoiceResponse

// Address is a pickup or delivery address.
type Address = client.Address

// Dimensions describes pallet dimensions in inches.
type Dimensions = client.Dimensions

// PalletDims returns a *Dimensions with the provided non-zero sides
// populated. It exists so callers can write Dimensions inline without
// taking the address of each int32. Pass zero for any side that should
// be left unspecified (the API default applies for those).
func PalletDims(length, width, height int32) *Dimensions {
	d := &Dimensions{}
	if length > 0 {
		d.Length = &length
	}
	if width > 0 {
		d.Width = &width
	}
	if height > 0 {
		d.Height = &height
	}
	return d
}

// Document is a downloadable shipment document (BOL, label, invoice, POD).
type Document = client.DocumentResponse

// DocumentType identifies a kind of shipment document.
type DocumentType = client.GetDocumentParamsDocumentType

// OrderComponent describes a pallet group within a shipment. The SDK exposes
// only the modern `Dimensions` shape; the deprecated array form on the
// underlying generated type is not surfaced.
type OrderComponent struct {
	PalletCount   int32       `json:"palletCount"`
	PoundsWeight  int32       `json:"poundsWeight"`
	Dimensions    *Dimensions `json:"dimensions,omitempty"`
	Description   *string     `json:"description,omitempty"`
	NmfcCode      *string     `json:"nmfcCode,omitempty"`
	PackagingType *string     `json:"packagingType,omitempty"`
	PieceCount    *int32      `json:"pieceCount,omitempty"`
}

// QuoteRequest mirrors the generated client.QuoteRequest but uses the
// SDK-local OrderComponent (no deprecated fields).
type QuoteRequest struct {
	PickupAddress      Address          `json:"pickupAddress"`
	DeliveryAddress    Address          `json:"deliveryAddress"`
	OrderComponents    []OrderComponent `json:"orderComponents"`
	RequiredPickupDate *time.Time       `json:"requiredPickupDate,omitempty"`
}

// ShipmentRequest mirrors the generated client.CreateShipmentRequest but
// uses the SDK-local OrderComponent (no deprecated fields).
type ShipmentRequest struct {
	PickupAddress      Address          `json:"pickupAddress"`
	DeliveryAddress    Address          `json:"deliveryAddress"`
	Description        string           `json:"description"`
	OrderComponents    []OrderComponent `json:"orderComponents"`
	QuoteID            *string          `json:"quoteId,omitempty"`
	PoNumber           *string          `json:"poNumber,omitempty"`
	RefNumber          *string          `json:"refNumber,omitempty"`
	RequiredPickupDate *time.Time       `json:"requiredPickupDate,omitempty"`
	RequiredDeliveryBy *time.Time       `json:"requiredDeliveryBy,omitempty"`
}

// Document type constants
const (
	DocumentTypeBOL           DocumentType = "BILL_OF_LADING"
	DocumentTypeInvoice       DocumentType = "INVOICE"
	DocumentTypeShippingLabel DocumentType = "SHIPPING_LABEL"
	DocumentTypePOD           DocumentType = "POD"
)

func toClientOrderComponents(in []OrderComponent) []client.OrderComponent {
	out := make([]client.OrderComponent, len(in))
	for i, c := range in {
		out[i] = client.OrderComponent{
			PalletCount:   c.PalletCount,
			PoundsWeight:  c.PoundsWeight,
			Dimensions:    c.Dimensions,
			Description:   c.Description,
			NmfcCode:      c.NmfcCode,
			PackagingType: c.PackagingType,
			PieceCount:    c.PieceCount,
		}
	}
	return out
}

func (q *QuoteRequest) toClient() client.QuoteRequest {
	return client.QuoteRequest{
		PickupAddress:      q.PickupAddress,
		DeliveryAddress:    q.DeliveryAddress,
		OrderComponents:    toClientOrderComponents(q.OrderComponents),
		RequiredPickupDate: q.RequiredPickupDate,
	}
}

func (s *ShipmentRequest) toClient() client.CreateShipmentRequest {
	return client.CreateShipmentRequest{
		PickupAddress:      s.PickupAddress,
		DeliveryAddress:    s.DeliveryAddress,
		Description:        s.Description,
		OrderComponents:    toClientOrderComponents(s.OrderComponents),
		QuoteId:            s.QuoteID,
		PoNumber:           s.PoNumber,
		RefNumber:          s.RefNumber,
		RequiredPickupDate: s.RequiredPickupDate,
		RequiredDeliveryBy: s.RequiredDeliveryBy,
	}
}
