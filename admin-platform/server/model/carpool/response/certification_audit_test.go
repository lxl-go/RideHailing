package response

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ride-hailing/admin-server/model/carpool"
)

func TestCertificationAuditResponseSerializesLargeIDsAsStrings(t *testing.T) {
	audit := carpool.CertificationAudit{
		ID:         2084116829910999040,
		UserID:     3,
		ReviewerID: 0,
		CreatedAt:  time.Date(2026, 8, 3, 11, 19, 8, 0, time.Local),
		UpdatedAt:  time.Date(2026, 8, 3, 11, 51, 0, 0, time.Local),
	}

	data, err := json.Marshal(NewCertificationAuditResponse(audit))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	body := string(data)

	for _, want := range []string{
		`"id":"2084116829910999040"`,
		`"userId":"3"`,
		`"reviewerId":"0"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"id":2084116829910999040`) {
		t.Fatalf("response JSON still exposes numeric id: %s", body)
	}
}

func TestCertificationAuditResponsesConvertEachItem(t *testing.T) {
	items := []carpool.CertificationAudit{
		{ID: 2084116829910999040, UserID: 3},
		{ID: 2084116829910999041, UserID: 4},
	}

	responses := NewCertificationAuditResponses(items)

	if len(responses) != 2 {
		t.Fatalf("len = %d, want 2", len(responses))
	}
	if responses[0].ID != "2084116829910999040" || responses[1].ID != "2084116829910999041" {
		t.Fatalf("ids not preserved as strings: %+v", responses)
	}
}
