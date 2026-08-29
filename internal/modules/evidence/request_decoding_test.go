package evidence

import (
	"strings"
	"testing"
)

func TestEvidenceHandleRequestRejectsUnknownMembers(t *testing.T) {
	body := strings.NewReader(`{"client_txn_id":"forbidden"}`)
	if apiErr := decodeHandleIssueRequest(body); apiErr == nil || apiErr.Code != "invalid_evidence_handle_request" {
		t.Fatalf("expected invalid_evidence_handle_request for handle issuance client_txn_id, got %#v", apiErr)
	}
}
