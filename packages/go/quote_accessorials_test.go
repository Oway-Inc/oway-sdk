package oway

import (
	"encoding/json"
	"testing"
)

func TestQuoteCarriesTheAccessorialsEachStopWasPricedFor(t *testing.T) {
	body := `{"id":"quote_123","quotedPriceInCents":125000,"transitDays":2,"estimatedDeliveryDate":null,
		"pickupAccessorials":{"appointmentRequired":false,"liftgateRequired":false},
		"deliveryAccessorials":{"appointmentRequired":true,"liftgateRequired":true,"callAheadRequired":false}}`
	var quote Quote
	if err := json.Unmarshal([]byte(body), &quote); err != nil {
		t.Fatal(err)
	}
	var delivery *QuoteStopAccessorials = quote.DeliveryAccessorials
	if delivery == nil || delivery.AppointmentRequired == nil || !*delivery.AppointmentRequired {
		t.Fatalf("delivery appointment not decoded: %+v", delivery)
	}
	if quote.PickupAccessorials == nil || quote.PickupAccessorials.AppointmentRequired == nil || *quote.PickupAccessorials.AppointmentRequired {
		t.Fatalf("pickup appointment not decoded as false: %+v", quote.PickupAccessorials)
	}
}

func TestQuoteWithoutRecordedAccessorialsLeavesThemNil(t *testing.T) {
	var quote Quote
	if err := json.Unmarshal([]byte(`{"id":"quote_123","transitDays":null,"estimatedDeliveryDate":null}`), &quote); err != nil {
		t.Fatal(err)
	}
	if quote.PickupAccessorials != nil || quote.DeliveryAccessorials != nil {
		t.Fatalf("accessorials = %+v / %+v, want nil", quote.PickupAccessorials, quote.DeliveryAccessorials)
	}
}
