package oway

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAppointmentExtensionWireSemantics(t *testing.T) {
	for _, extension := range []*string{nil, extensionString(""), extensionString("0042")} {
		contact := AppointmentContact{Name: "Dock desk", Phone: extensionString("+12125550123"), PhoneExtension: extension}
		request := ShipmentRequest{Appointments: &Appointments{Pickup: &AppointmentRequirement{Contacts: &[]AppointmentContact{contact}}}}
		body, err := json.Marshal(request.toClient())
		if err != nil {
			t.Fatal(err)
		}
		if extension == nil {
			if strings.Contains(string(body), "phoneExtension") {
				t.Fatalf("omitted extension was emitted: %s", body)
			}
		} else if !strings.Contains(string(body), `"phoneExtension":"`+*extension+`"`) {
			t.Fatalf("extension changed or disappeared: %s", body)
		}
		var decoded AppointmentContact
		encoded, _ := json.Marshal(contact)
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if extension != nil && (decoded.PhoneExtension == nil || *decoded.PhoneExtension != *extension) {
			t.Fatal("extension did not survive a round trip")
		}
	}
}

func extensionString(value string) *string { return &value }
