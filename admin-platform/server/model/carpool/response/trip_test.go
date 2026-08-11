package response

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ride-hailing/admin-server/model/carpool"
)

func TestTripResponseSerializesLargeIDsAsStrings(t *testing.T) {
	trip := carpool.Trip{
		ID:              2086799305091444736,
		PublisherID:     2086799305091444737,
		AuditOperatorID: 888,
		MatchedOrderID:  2086799305091444738,
		OriginName:      "A",
		DestName:        "B",
		CreatedAt:       time.Date(2026, 8, 10, 9, 0, 0, 0, time.Local),
		UpdatedAt:       time.Date(2026, 8, 10, 9, 0, 0, 0, time.Local),
	}

	data, err := json.Marshal(NewTripResponse(trip))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(data)

	for _, want := range []string{
		`"id":"2086799305091444736"`,
		`"publisherId":"2086799305091444737"`,
		`"auditOperatorId":"888"`,
		`"matchedOrderId":"2086799305091444738"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"id":2086799305091444736`) {
		t.Fatalf("response JSON still exposes numeric id: %s", body)
	}
}

func TestTripResponsesConvertEachItem(t *testing.T) {
	items := []carpool.Trip{
		{ID: 2086799305091444736, PublisherID: 4},
		{ID: 2086799305091444737, PublisherID: 5},
	}

	responses := NewTripResponses(items)

	if len(responses) != 2 {
		t.Fatalf("len = %d, want 2", len(responses))
	}
	if responses[0].ID != "2086799305091444736" || responses[1].ID != "2086799305091444737" {
		t.Fatalf("ids not preserved as strings: %+v", responses)
	}
}
