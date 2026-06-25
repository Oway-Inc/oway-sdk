package oway

// Carrier-side convenience wrappers for carrier integrations: poll for
// offers, accept or reject them, drive a shipment through pickup and
// delivery, stream GPS positions, report exceptions, and submit trips.
// The generated client (`packages/go/client/generated.go`) already
// exposes the underlying methods; these wrappers add the same retry and
// decode semantics as the shipper-side wrappers in oway.go.
//
// Type aliases are re-exported so consumers don't have to import the
// generated client package directly.

import (
	"context"
	"errors"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// Carrier-side type aliases re-exported from the generated client so callers
// don't need to import the client package directly. The aliased types carry
// their own doc comments in the generated code.
//
//nolint:revive // re-exports; doc lives on the generated types
type (
	Offer                       = client.Offer
	AcceptOfferRequest          = client.AcceptOfferRequest
	RejectOfferRequest          = client.RejectOfferRequest
	RejectOfferReason           = client.RejectOfferRequestReason
	CarrierShipment             = client.CarrierShipment
	PickupConfirmation          = client.PickupConfirmation
	PickupConfirmationRequest   = client.PickupConfirmationRequest
	DeliveryConfirmation        = client.DeliveryConfirmation
	DeliveryConfirmationRequest = client.DeliveryConfirmationRequest
	OfferLocationUpdate         = client.OfferLocationUpdate
	LocationAcknowledgment      = client.LocationAcknowledgment
	ExceptionReportRequest      = client.ExceptionReportRequest
	ExceptionType               = client.ExceptionReportRequestExceptionType
	ExceptionResponse           = client.ExceptionResponse
	CarrierTracking             = client.TrackingResponse
	CarrierDocumentResponse     = client.CarrierDocumentResponse
	CarrierTripRequest          = client.CarrierTripRequest
)

// ListCarrierOffers returns all pending offers visible to the
// authenticated carrier.
func (c *Client) ListCarrierOffers(ctx context.Context) ([]Offer, error) {
	var out []Offer
	err := c.retry(ctx, func() error {
		r, err := c.api.GetOffersWithResponse(ctx)
		if err != nil {
			return err
		}
		decoded, derr := decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		if derr != nil {
			return derr
		}
		if decoded != nil {
			out = *decoded
		}
		return nil
	})
	return out, err
}

// GetCarrierOffer fetches a single offer by identifier.
func (c *Client) GetCarrierOffer(ctx context.Context, identifier string) (*Offer, error) {
	if identifier == "" {
		return nil, errors.New("oway: offer identifier is required")
	}
	var out *Offer
	err := c.retry(ctx, func() error {
		r, err := c.api.GetOfferWithResponse(ctx, identifier)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// AcceptCarrierOffer accepts a pending offer. The req carries driver
// info + carrier reference + any required acceptance metadata.
func (c *Client) AcceptCarrierOffer(ctx context.Context, identifier string, req AcceptOfferRequest) (*Offer, error) {
	if identifier == "" {
		return nil, errors.New("oway: offer identifier is required")
	}
	var out *Offer
	err := c.retry(ctx, func() error {
		r, err := c.api.AcceptOfferWithResponse(ctx, identifier, client.AcceptOfferJSONRequestBody(req))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// RejectCarrierOffer rejects a pending offer with a typed reason.
func (c *Client) RejectCarrierOffer(ctx context.Context, identifier string, req RejectOfferRequest) (*Offer, error) {
	if identifier == "" {
		return nil, errors.New("oway: offer identifier is required")
	}
	var out *Offer
	err := c.retry(ctx, func() error {
		r, err := c.api.RejectOfferWithResponse(ctx, identifier, client.RejectOfferJSONRequestBody(req))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// GetCarrierShipment fetches the carrier-view shipment for an accepted
// offer. Carrier-side endpoints return CarrierShipment (vs. shipper-side
// Shipment), schemas are currently identical but kept separate to
// match the spec. The generated client suffixes the carrier-side
// method with "1" to disambiguate from the shipper-side GetShipment.
func (c *Client) GetCarrierShipment(ctx context.Context, identifier string) (*CarrierShipment, error) {
	if identifier == "" {
		return nil, errors.New("oway: shipment identifier is required")
	}
	var out *CarrierShipment
	err := c.retry(ctx, func() error {
		r, err := c.api.GetShipment1WithResponse(ctx, identifier)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// ConfirmPickup posts a pickup confirmation for a shipment. The req
// carries lat/lng + signature + any POP attachments by reference.
func (c *Client) ConfirmPickup(ctx context.Context, identifier string, req PickupConfirmationRequest) (*CarrierShipment, error) {
	if identifier == "" {
		return nil, errors.New("oway: shipment identifier is required")
	}
	var out *CarrierShipment
	err := c.retry(ctx, func() error {
		r, err := c.api.ConfirmPickupWithResponse(ctx, identifier, client.ConfirmPickupJSONRequestBody(req))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// ConfirmDelivery posts a delivery confirmation for a shipment. The req
// carries lat/lng + signature + POD attachments by reference.
func (c *Client) ConfirmDelivery(ctx context.Context, identifier string, req DeliveryConfirmationRequest) (*CarrierShipment, error) {
	if identifier == "" {
		return nil, errors.New("oway: shipment identifier is required")
	}
	var out *CarrierShipment
	err := c.retry(ctx, func() error {
		r, err := c.api.ConfirmDeliveryWithResponse(ctx, identifier, client.ConfirmDeliveryJSONRequestBody(req))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// SubmitLocation posts a GPS location update for an accepted offer
// (during DRIVING_TO_PICKUP / IN_TRANSIT). The simulator calls this on
// the configured GpsUpdateInterval.
func (c *Client) SubmitLocation(ctx context.Context, identifier string, loc OfferLocationUpdate) (*LocationAcknowledgment, error) {
	if identifier == "" {
		return nil, errors.New("oway: offer identifier is required")
	}
	var out *LocationAcknowledgment
	err := c.retry(ctx, func() error {
		r, err := c.api.SubmitLocationWithResponse(ctx, identifier, client.SubmitLocationJSONRequestBody(loc))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// ReportException reports an exception (weather, mechanical, traffic,
// shipper delay) during transit. The simulator uses this to inject
// realistic interruptions per the BehaviorConfig's ExceptionRate.
func (c *Client) ReportException(ctx context.Context, identifier string, req ExceptionReportRequest) (*ExceptionResponse, error) {
	if identifier == "" {
		return nil, errors.New("oway: shipment identifier is required")
	}
	var out *ExceptionResponse
	err := c.retry(ctx, func() error {
		r, err := c.api.ReportExceptionWithResponse(ctx, identifier, client.ReportExceptionJSONRequestBody(req))
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}

// SubmitCarrierTrips pushes one or more trip summaries to the
// carrier-side coverage API. Called by the simulator after DELIVERED
// to seed coverage growth tier with synthetic ELD-shaped trip data.
//
// AddTrips returns an empty 200 body with no JSON shape to unwrap.
func (c *Client) SubmitCarrierTrips(ctx context.Context, trips []CarrierTripRequest) error {
	if len(trips) == 0 {
		return errors.New("oway: at least one trip is required")
	}
	return c.retry(ctx, func() error {
		r, err := c.api.AddTripsWithResponse(ctx, client.AddTripsJSONRequestBody(trips))
		if err != nil {
			return err
		}
		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			requestID := ""
			if r.HTTPResponse != nil {
				requestID = r.HTTPResponse.Header.Get("x-request-id")
			}
			return parseHTTPError(r.StatusCode(), requestID, r.Body)
		}
		return nil
	})
}

// GetCarrierTracking fetches GPS tracking history for a carrier
// shipment by identifier (offerId, orderNumber, or carrierReference).
func (c *Client) GetCarrierTracking(ctx context.Context, identifier string) (*CarrierTracking, error) {
	if identifier == "" {
		return nil, errors.New("oway: shipment identifier is required")
	}
	var out *CarrierTracking
	err := c.retry(ctx, func() error {
		r, err := c.api.GetShipmentTrackingWithResponse(ctx, identifier)
		if err != nil {
			return err
		}
		out, err = decode(r.StatusCode(), r.Body, r.HTTPResponse, r.JSON200)
		return err
	})
	return out, err
}
