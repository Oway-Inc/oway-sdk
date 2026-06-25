package oway

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// deprecatedShipmentRequestFields lists json tags on client.CreateShipmentRequest
// intentionally not mirrored on the SDK-local ShipmentRequest. Empty today:
// the wrapper exposes the full request body. Adding an entry is a deliberate
// decision to hide a field from the public SDK surface.
var deprecatedShipmentRequestFields = map[string]bool{}

// TestShipmentRequestMirrorsClient asserts every non-deprecated field on the
// generated client.CreateShipmentRequest has a sibling on the SDK-local
// oway.ShipmentRequest. Without this, adding a field to the OpenAPI spec (and
// the regenerated client) silently never reaches the wire, because toClient
// only copies whatever the SDK-local type happens to expose.
func TestShipmentRequestMirrorsClient(t *testing.T) {
	clientTags := jsonTags(reflect.TypeOf(client.CreateShipmentRequest{}))
	sdkTags := jsonTags(reflect.TypeOf(ShipmentRequest{}))

	for tag := range clientTags {
		if deprecatedShipmentRequestFields[tag] {
			continue
		}
		if !sdkTags[tag] {
			t.Errorf(
				"client.CreateShipmentRequest.%q is not mirrored on oway.ShipmentRequest. "+
					"Either add the field to ShipmentRequest + toClient, or, if it's "+
					"intentionally hidden, add it to deprecatedShipmentRequestFields with a comment.",
				tag,
			)
		}
	}
}

// TestShipmentRequestToClientCopiesEveryField populates the SDK-local
// ShipmentRequest with a distinct non-zero value per field, runs it through
// toClient, JSON-marshals both sides, and compares per-tag. Catches both a
// field the mapper drops (missing on the client side) and a field copied from
// the wrong source (present but differing from the SDK source).
func TestShipmentRequestToClientCopiesEveryField(t *testing.T) {
	sdk := fullyPopulatedShipmentRequest()
	got := sdk.toClient()

	sdkMap := marshalToMap(t, sdk)
	gotMap := marshalToMap(t, got)

	clientType := reflect.TypeOf(got)
	for i := 0; i < clientType.NumField(); i++ {
		tag := jsonTagName(clientType.Field(i).Tag.Get("json"))
		if tag == "" || tag == "-" || deprecatedShipmentRequestFields[tag] {
			continue
		}
		gotVal, gotOK := gotMap[tag]
		if !gotOK {
			t.Errorf(
				"toClient did not copy %q to client.CreateShipmentRequest. "+
					"Populate it in fullyPopulatedShipmentRequest and copy it in toClient.",
				tag,
			)
			continue
		}
		sdkVal, sdkOK := sdkMap[tag]
		if !sdkOK {
			t.Errorf(
				"client.CreateShipmentRequest.%q has a value but the SDK source doesn't "+
					"expose the same JSON tag. Add the field to oway.ShipmentRequest or to "+
					"deprecatedShipmentRequestFields.",
				tag,
			)
			continue
		}
		// Compare semantically: oway.OrderComponent and client.OrderComponent
		// hold the same data but serialize keys in different orders, so a raw
		// byte compare would false-positive on nested composite fields.
		if !jsonSemanticEqual(t, sdkVal, gotVal) {
			t.Errorf(
				"toClient miscopied %q: source=%s, output=%s. Suggests the mapper read "+
					"from the wrong field.",
				tag, sdkVal, gotVal,
			)
		}
	}
}

// jsonSemanticEqual reports whether two JSON fragments are equal ignoring
// object key ordering.
func jsonSemanticEqual(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		t.Fatalf("unmarshal source %s: %v", a, err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		t.Fatalf("unmarshal output %s: %v", b, err)
	}
	return reflect.DeepEqual(av, bv)
}

// fullyPopulatedShipmentRequest sets every field on the SDK-local
// ShipmentRequest to a distinct non-zero value so the miscopy detection in
// TestShipmentRequestToClientCopiesEveryField can spot a swapped assignment.
// Pickup and delivery carry different values so a swap is detectable.
func fullyPopulatedShipmentRequest() ShipmentRequest {
	quoteID := "QT-PARITY1"
	po := "PO-PARITY"
	ref := "REF-PARITY"
	pickupAt := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	deliverBy := time.Date(2026, 6, 3, 17, 0, 0, 0, time.UTC)
	dispatchEmail := openapi_types.Email("dispatch@example.com")
	dispatchPhone := "+15555550100"
	scac := "ACME"
	return ShipmentRequest{
		PickupAddress:      client.Address{Address1: "1 Pickup St", City: "Shafter", State: "CA", ZipCode: "93263"},
		DeliveryAddress:    client.Address{Address1: "2 Delivery Ave", City: "Lake Havasu City", State: "AZ", ZipCode: "86404"},
		Description:        "Parity probe cargo",
		OrderComponents:    []OrderComponent{fullyPopulatedSDKComponent()},
		QuoteID:            &quoteID,
		PoNumber:           &po,
		RefNumber:          &ref,
		RequiredPickupDate: &pickupAt,
		RequiredDeliveryBy: &deliverBy,
		Appointments: &Appointments{
			Delivery: &AppointmentRequirement{Channel: AppointmentChannelEmail},
		},
		ShipperDispatch: &ShipperDispatch{
			Name:  "Acme Logistics",
			Email: &dispatchEmail,
			Phone: &dispatchPhone,
			Scac:  &scac,
		},
	}
}
