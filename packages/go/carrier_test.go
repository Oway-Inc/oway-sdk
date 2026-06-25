package oway

import (
	"context"
	"net/http"
	"testing"
)

// TestListCarrierOffers verifies the convenience wrapper unwraps the
// generated client's response correctly + propagates the API key /
// auth headers as the shipper-side wrappers do.
func TestListCarrierOffers(t *testing.T) {
	c, srv := newTestServer(t, okToken,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/carrier/offers" {
				t.Errorf("unexpected path %q", r.URL.Path)
			}
			if got := r.Header.Get("x-oway-api-key"); got != "oway_sk_test" {
				t.Errorf("api key header = %q, want oway_sk_test", got)
			}
			writeJSON(w, http.StatusOK, `[{"id":"off-1"},{"id":"off-2"}]`)
		})
	defer srv.Close()

	offers, err := c.ListCarrierOffers(context.Background())
	if err != nil {
		t.Fatalf("ListCarrierOffers: %v", err)
	}
	if len(offers) != 2 {
		t.Errorf("len(offers) = %d, want 2", len(offers))
	}
}

func TestAcceptCarrierOffer_RequiresIdentifier(t *testing.T) {
	c, srv := newTestServer(t, okToken,
		func(http.ResponseWriter, *http.Request) {
			t.Fatal("API should not be called when identifier missing")
		})
	defer srv.Close()

	_, err := c.AcceptCarrierOffer(context.Background(), "", AcceptOfferRequest{})
	if err == nil {
		t.Fatal("expected error for empty identifier, got nil")
	}
}

func TestSubmitCarrierTrips_RequiresAtLeastOne(t *testing.T) {
	c, srv := newTestServer(t, okToken,
		func(http.ResponseWriter, *http.Request) {
			t.Fatal("API should not be called for empty trip slice")
		})
	defer srv.Close()

	if err := c.SubmitCarrierTrips(context.Background(), nil); err == nil {
		t.Error("expected error for nil trips, got nil")
	}
	if err := c.SubmitCarrierTrips(context.Background(), []CarrierTripRequest{}); err == nil {
		t.Error("expected error for empty trips, got nil")
	}
}

// TestGetCarrierShipment_UsesCarrierEndpoint guards against an easy
// mistake — calling the shipper-side GetShipmentWithResponse from the
// carrier wrapper. The generated client suffixes the carrier-side
// method "1"; this test verifies the wrapper hits /v1/carrier/shipments/{id}.
func TestGetCarrierShipment_UsesCarrierEndpoint(t *testing.T) {
	var hitPath string
	c, srv := newTestServer(t, okToken,
		func(w http.ResponseWriter, r *http.Request) {
			hitPath = r.URL.Path
			writeJSON(w, http.StatusOK, `{"id":"ship-1"}`)
		})
	defer srv.Close()

	_, err := c.GetCarrierShipment(context.Background(), "ship-1")
	if err != nil {
		t.Fatalf("GetCarrierShipment: %v", err)
	}
	if hitPath != "/v1/carrier/shipments/ship-1" {
		t.Errorf("hit path = %q, want /v1/carrier/shipments/ship-1", hitPath)
	}
}

func TestGetCarrierTracking(t *testing.T) {
	c, srv := newTestServer(t, okToken,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/carrier/shipments/off-1/tracking" {
				t.Errorf("unexpected path %q", r.URL.Path)
			}
			writeJSON(w, http.StatusOK, `{"order_number":"ZKYQ5"}`)
		})
	defer srv.Close()

	trk, err := c.GetCarrierTracking(context.Background(), "off-1")
	if err != nil {
		t.Fatalf("GetCarrierTracking: %v", err)
	}
	if trk == nil || trk.OrderNumber == nil || *trk.OrderNumber != "ZKYQ5" {
		t.Errorf("unexpected tracking %+v", trk)
	}
}

func TestGetCarrierTracking_RequiresIdentifier(t *testing.T) {
	c, srv := newTestServer(t, okToken,
		func(http.ResponseWriter, *http.Request) {
			t.Fatal("API should not be called when identifier missing")
		})
	defer srv.Close()

	if _, err := c.GetCarrierTracking(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty identifier, got nil")
	}
}
