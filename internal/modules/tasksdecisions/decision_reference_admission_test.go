package tasksdecisions

import (
	"strings"
	"testing"
)

func TestDirectDecisionReferenceDecoderAcceptsOnlyExactStableIDs_Unit(t *testing.T) {
	stableDecisionID := "11111111-2222-3333-4444-555555555555"
	valid := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":"` + stableDecisionID + `"}]}`
	request, apiErr := DecodePatchRequest(strings.NewReader(valid))
	if apiErr != nil {
		t.Fatalf("expected exact stable decision id to decode: %#v", apiErr)
	}
	if got := request.Changes[0].Value.UUID.String(); got != stableDecisionID {
		t.Fatalf("unexpected decoded decision id: got %s want %s", got, stableDecisionID)
	}

	clear := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":null}]}`
	request, apiErr = DecodePatchRequest(strings.NewReader(clear))
	if apiErr != nil {
		t.Fatalf("expected direct null clear to decode: %#v", apiErr)
	}
	if request.Changes[0].Value == nil || request.Changes[0].CanonicalValue != nil {
		t.Fatalf("expected direct null clear value, got %#v", request.Changes[0].Value)
	}

	invalidValues := []string{
		`" 11111111-2222-3333-4444-555555555555"`,
		`"11111111-2222-3333-4444-555555555555 "`,
		`"decision@example.test"`,
		`"Contain endpoint"`,
		`"decision:11111111-2222-3333-4444-555555555555"`,
		`""`,
		`[]`,
		`{}`,
		`true`,
		`42`,
	}
	for _, rawValue := range invalidValues {
		body := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":` + rawValue + `}]}`
		if _, apiErr := DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
			t.Fatalf("expected invalid direct decision ref %s to fail", rawValue)
		}
	}

	actionPayloadClear := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"remove_record_ref","item_ref":"record_ref:` + stableDecisionID + `"}]}}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(actionPayloadClear)); apiErr == nil {
		t.Fatalf("expected non-direct clear shape to fail")
	}
}
