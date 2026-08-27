package oway

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// deprecatedClientFields lists json tags on client.OrderComponent that are
// intentionally hidden from the SDK-local OrderComponent wrapper. The wrapper's
// purpose (see the doc comment on OrderComponent in types.go) is to drop these
// from the public SDK surface; adding a new deprecation means adding an entry
// here and a deliberate decision to not mirror it on the wrapper.
var deprecatedClientFields = map[string]bool{
	"palletDimensions": true,
}

// TestOrderComponentMirrorsClient asserts that every non-deprecated field on
// the generated client.OrderComponent has a sibling on the SDK-local
// oway.OrderComponent. Without this, adding a field to the OpenAPI spec
// (and the regenerated client) silently zero-values on the wire because
// toClientOrderComponents only copies whatever the SDK-local type happens
// to expose.
func TestOrderComponentMirrorsClient(t *testing.T) {
	clientTags := jsonTags(reflect.TypeOf(client.OrderComponent{}))
	sdkTags := jsonTags(reflect.TypeOf(OrderComponent{}))

	for tag := range clientTags {
		if deprecatedClientFields[tag] {
			continue
		}
		if !sdkTags[tag] {
			t.Errorf(
				"client.OrderComponent.%q is not mirrored on oway.OrderComponent. "+
					"Either add the field to types.OrderComponent + toClientOrderComponents, "+
					"or, if it's intentionally hidden, add it to deprecatedClientFields with a comment.",
				tag,
			)
		}
	}
}

// TestToClientOrderComponentsCopiesEveryField populates the SDK-local
// OrderComponent with distinct non-zero values for every exposed field, runs
// it through toClientOrderComponents, then JSON-marshals both sides and
// compares per-tag values. Catches both halves of the wrapper gap: a field
// the mapper drops on the floor (value missing on the client side), and a
// field the mapper copies from the wrong source (value present but differs
// from the SDK source).
func TestToClientOrderComponentsCopiesEveryField(t *testing.T) {
	sdk := fullyPopulatedSDKComponent()
	out := toClientOrderComponents([]OrderComponent{sdk})
	if len(out) != 1 {
		t.Fatalf("expected 1 client component, got %d", len(out))
	}
	got := out[0]

	sdkMap := marshalToMap(t, sdk)
	gotMap := marshalToMap(t, got)

	clientType := reflect.TypeOf(got)
	for i := 0; i < clientType.NumField(); i++ {
		tag := jsonTagName(clientType.Field(i).Tag.Get("json"))
		if tag == "" || tag == "-" || deprecatedClientFields[tag] {
			continue
		}
		gotVal, gotOK := gotMap[tag]
		if !gotOK {
			t.Errorf(
				"toClientOrderComponents did not copy %q to client.OrderComponent. "+
					"Populate it in fullyPopulatedSDKComponent (if the SDK type exposes "+
					"it) or update toClientOrderComponents.",
				tag,
			)
			continue
		}
		sdkVal, sdkOK := sdkMap[tag]
		if !sdkOK {
			t.Errorf(
				"client.OrderComponent.%q has a value but the SDK source doesn't expose "+
					"the same JSON tag. Either add the field to oway.OrderComponent or to "+
					"deprecatedClientFields.",
				tag,
			)
			continue
		}
		if !bytes.Equal(sdkVal, gotVal) {
			t.Errorf(
				"toClientOrderComponents miscopied %q: source=%s, output=%s. "+
					"Suggests the mapper read from the wrong field.",
				tag,
				string(sdkVal),
				string(gotVal),
			)
		}
	}
}

func marshalToMap(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	out := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
	return out
}

func jsonTags(t reflect.Type) map[string]bool {
	out := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := jsonTagName(t.Field(i).Tag.Get("json"))
		if tag != "" && tag != "-" {
			out[tag] = true
		}
	}
	return out
}

func jsonTagName(tag string) string {
	if tag == "" {
		return ""
	}
	if i := strings.IndexByte(tag, ','); i >= 0 {
		return tag[:i]
	}
	return tag
}

// fullyPopulatedSDKComponent sets every field on the SDK-local
// OrderComponent to a distinct non-zero value. The miscopy detection in
// TestToClientOrderComponentsCopiesEveryField relies on values being distinct
// per field — new fields should be added here with values that don't collide
// with neighbors so a swapped assignment is detectable.
func fullyPopulatedSDKComponent() OrderComponent {
	desc := "Stone NOI"
	nmfc := "90500-04"
	pkg := "Pallets"
	pieces := int32(7)
	freightClass := "100"
	return OrderComponent{
		PalletCount:   2,
		PoundsWeight:  3000,
		Dimensions:    PalletDims(48, 48, 48),
		Description:   &desc,
		NmfcCode:      &nmfc,
		FreightClass:  &freightClass,
		PackagingType: &pkg,
		PieceCount:    &pieces,
	}
}
