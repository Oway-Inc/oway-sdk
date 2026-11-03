package oway

import (
	"context"
	"encoding/json"
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

func TestCarrierDocumentCompletion(t *testing.T) {
	c, srv := newTestServer(t, okToken, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-oway-api-key") != "carrier-key" {
			t.Error("company identity was not propagated")
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("bearer token missing")
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/carrier/shipments/offer-1/upload":
			writeJSON(w, http.StatusOK, `{"document_id":"document-1","upload_url":"https://example.com/upload","content_type":"application/pdf"}`)
		case "PUT /v1/carrier/shipments/offer-1/uploads/document-1/confirm":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["document_type"] != "PROOF_OF_DELIVERY" || body["file_name"] != "pod.pdf" || body["content_type"] != "application/pdf" {
				t.Errorf("incorrect metadata: %v", body)
			}
			writeJSON(w, http.StatusOK, `{"document_id":"document-1","review_status":"PENDING"}`)
		case "GET /v1/carrier/shipments/offer-1/documents/document-1":
			writeJSON(w, http.StatusOK, `{"document_id":"document-1","review_status":"APPROVED","download_url":"https://example.com/pod.pdf"}`)
		default:
			t.Errorf("unexpected operation %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	})
	defer srv.Close()
	ctx := WithCompanyAPIKey(context.Background(), "carrier-key")
	staging, err := c.RequestCarrierUploadURL(ctx, "offer-1", UploadUrlRequest{DocumentType: "PROOF_OF_DELIVERY", ContentType: "application/pdf"})
	if err != nil || staging == nil || staging.DocumentId == nil || *staging.DocumentId != "document-1" {
		t.Fatalf("upload URL: %+v %v", staging, err)
	}
	completed, err := c.ConfirmCarrierUpload(ctx, "offer-1", *staging.DocumentId, UploadConfirmationRequest{DocumentType: "PROOF_OF_DELIVERY", ContentType: "application/pdf", FileName: "pod.pdf"})
	if err != nil || completed == nil || completed.ReviewStatus == nil || *completed.ReviewStatus != "PENDING" {
		t.Fatalf("completion: %+v %v", completed, err)
	}
	doc, err := c.GetCarrierUploadedDocument(ctx, "offer-1", *staging.DocumentId)
	if err != nil || doc == nil || doc.DownloadUrl == nil || *doc.DownloadUrl != "https://example.com/pod.pdf" {
		t.Fatalf("retrieval: %+v %v", doc, err)
	}
}

func TestCarrierDocumentErrors(t *testing.T) {
	c, srv := newTestServer(t, okToken, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusConflict, `{"reason":"upload_missing","detail":"Upload bytes before completion"}`)
	})
	defer srv.Close()
	if _, err := c.ConfirmCarrierUpload(context.Background(), "offer-1", "document-1", UploadConfirmationRequest{}); err == nil {
		t.Fatal("completion conflict must propagate")
	}
	if _, err := c.ConfirmCarrierUpload(context.Background(), "offer-1", "", UploadConfirmationRequest{}); err == nil {
		t.Fatal("empty document ID must fail")
	}
	if _, err := c.GetCarrierUploadedDocument(context.Background(), "", "document-1"); err == nil {
		t.Fatal("empty shipment ID must fail")
	}
}
