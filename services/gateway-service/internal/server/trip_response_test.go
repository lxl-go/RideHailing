package server

import (
	"encoding/json"
	"strings"
	"testing"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func TestMobileDriverTripListResponseSerializesLargeIDsAsStrings(t *testing.T) {
	reply := &tripv1.ListDriverTripsReply{
		Total: 1,
		Items: []*tripv1.TripItem{{
			Id:              2086799305091444736,
			DriverId:        2086799305091444737,
			AuditOperatorId: 888,
			Origin:          "A",
			Destination:     "B",
			Status:          10,
		}},
	}

	data, err := json.Marshal(mobileDriverTripListResponse(reply))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(data)

	for _, want := range []string{
		`"id":"2086799305091444736"`,
		`"driverId":"2086799305091444737"`,
		`"driver_id":"2086799305091444737"`,
		`"auditOperatorId":"888"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"id":2086799305091444736`) {
		t.Fatalf("response JSON still exposes numeric id: %s", body)
	}
}
