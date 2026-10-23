package oway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

func TestHazmatCapturePreservesPartialMaterialsAndExplicitFalse(t *testing.T) {
	id, source, unit := "material-one", "item-1", "kg"
	no := false
	quantity := 12.5
	entries := []HazmatEntry{
		{EntryId: &id, SourceItemReference: &source, LimitedQuantity: &no},
		{TotalQuantity: &quantity, QuantityUnit: &unit},
	}
	yes := true
	classification := HazmatClassification(client.LIMITEDQUANTITY)
	paper := &HazmatPaper{Classification: &classification, Entries: &entries}
	quote := (&QuoteRequest{Hazmat: &yes, HazmatPaper: paper}).toClient()
	shipment := (&ShipmentRequest{Hazmat: &yes, HazmatPaper: paper}).toClient()
	for _, request := range []any{quote, shipment} {
		raw, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "null") {
			t.Fatalf("unknown facts must be omitted, not sent as null: %s", raw)
		}
		var roundTrip struct {
			HazmatPaper HazmatPaper `json:"hazmatPaper"`
		}
		if err := json.Unmarshal(raw, &roundTrip); err != nil {
			t.Fatal(err)
		}
		if c := roundTrip.HazmatPaper.Classification; c == nil || *c != client.LIMITEDQUANTITY {
			t.Fatalf("classification lost: %s", raw)
		}
		got := *roundTrip.HazmatPaper.Entries
		if len(got) != 2 || got[0].LimitedQuantity == nil || *got[0].LimitedQuantity {
			t.Fatalf("partial rows or explicit false lost: %s", raw)
		}
		if got[0].PackageCount != nil || got[1].PackingGroup != nil || roundTrip.HazmatPaper.Certification != nil {
			t.Fatalf("unknown facts or consent invented: %s", raw)
		}
	}
}

func TestLegacyRequestsOmitHazmatFields(t *testing.T) {
	for _, request := range []any{(&QuoteRequest{}).toClient(), (&ShipmentRequest{}).toClient()} {
		body := marshalToMap(t, request)
		for _, key := range []string{"hazmat", "hazmatPaper"} {
			if _, exists := body[key]; exists {
				t.Fatalf("legacy request unexpectedly contains %s", key)
			}
		}
	}
}
