package response

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ride-hailing/admin-server/model/carpool"
)

func TestVehicleInfoResponseSerializesLargeIDsAsStrings(t *testing.T) {
	info := carpool.VehicleInfo{
		ID:         2084282472140513280,
		DriverID:   4,
		ReviewerID: 0,
		CreatedAt:  time.Date(2026, 8, 3, 22, 17, 20, 0, time.Local),
		UpdatedAt:  time.Date(2026, 8, 3, 22, 17, 20, 0, time.Local),
	}

	data, err := json.Marshal(NewVehicleInfoResponse(info))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(data)

	for _, want := range []string{
		`"id":"2084282472140513280"`,
		`"driverId":"4"`,
		`"reviewerId":"0"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"id":2084282472140513280`) {
		t.Fatalf("response JSON still exposes numeric id: %s", body)
	}
}

func TestVehicleInfoResponsesConvertEachItem(t *testing.T) {
	items := []carpool.VehicleInfo{
		{ID: 2084282472140513280, DriverID: 4},
		{ID: 2084282472140513281, DriverID: 5},
	}

	responses := NewVehicleInfoResponses(items)

	if len(responses) != 2 {
		t.Fatalf("len = %d, want 2", len(responses))
	}
	if responses[0].ID != "2084282472140513280" || responses[1].ID != "2084282472140513281" {
		t.Fatalf("ids not preserved as strings: %+v", responses)
	}
}
