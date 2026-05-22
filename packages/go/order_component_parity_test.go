package oway

import (
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
// OrderComponent with non-zero values for every exposed field, runs it through
// toClientOrderComponents, and asserts the corresponding client.OrderComponent
// fields are also non-zero. Catches the "added the field on the wrapper but
// forgot to copy it in toClientOrderComponents" mistake — the second half of
// the gap the wrapper introduces.
func TestToClientOrderComponentsCopiesEveryField(t *testing.T) {
	sdk := fullyPopulatedSDKComponent()
	out := toClientOrderComponents([]OrderComponent{sdk})
	if len(out) != 1 {
		t.Fatalf("expected 1 client component, got %d", len(out))
	}
	got := out[0]

	clientType := reflect.TypeOf(got)
	clientVal := reflect.ValueOf(got)
	for i := 0; i < clientType.NumField(); i++ {
		f := clientType.Field(i)
		tag := jsonTagName(f.Tag.Get("json"))
		if tag == "" || tag == "-" || deprecatedClientFields[tag] {
			continue
		}
		if isZero(clientVal.Field(i)) {
			t.Errorf(
				"toClientOrderComponents did not copy %q to client.OrderComponent. "+
					"Either populate it in the test's fullyPopulatedSDKComponent "+
					"(if the SDK type already exposes it) or update toClientOrderComponents.",
				tag,
			)
		}
	}
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

func isZero(v reflect.Value) bool {
	if v.Kind() == reflect.Pointer {
		return v.IsNil()
	}
	return v.IsZero()
}

// fullyPopulatedSDKComponent sets every field on the SDK-local
// OrderComponent to a non-zero value. New fields should be added here as
// they're added to OrderComponent so the copy-coverage test stays accurate.
func fullyPopulatedSDKComponent() OrderComponent {
	desc := "Stone NOI"
	nmfc := "90500-04"
	pkg := "Pallets"
	pieces := int32(0)
	return OrderComponent{
		PalletCount:   2,
		PoundsWeight:  3000,
		Dimensions:    PalletDims(48, 48, 48),
		Description:   &desc,
		NmfcCode:      &nmfc,
		PackagingType: &pkg,
		PieceCount:    &pieces,
	}
}
