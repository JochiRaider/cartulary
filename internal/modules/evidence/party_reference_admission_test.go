package evidence_test

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
)

func TestDirectPartyReferencesAcceptOnlyExactStableIDs_Unit(t *testing.T) {
	stablePartyID := "11111111-2222-3333-4444-555555555555"
	valid := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":"` + stablePartyID + `"}]}`
	request, apiErr := evidence.DecodePatchRequest(strings.NewReader(valid))
	if apiErr != nil {
		t.Fatalf("expected exact stable party id to decode: %#v", apiErr)
	}
	if got := request.Changes[0].Value.UUID.String(); got != stablePartyID {
		t.Fatalf("unexpected decoded party id: got %s want %s", got, stablePartyID)
	}

	clear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":null}]}`
	request, apiErr = evidence.DecodePatchRequest(strings.NewReader(clear))
	if apiErr != nil {
		t.Fatalf("expected direct null clear to decode: %#v", apiErr)
	}
	if request.Changes[0].Value == nil || request.Changes[0].CanonicalValue != nil {
		t.Fatalf("expected direct null clear value, got %#v", request.Changes[0].Value)
	}

	invalidValues := []string{
		`" 11111111-2222-3333-4444-555555555555"`,
		`"11111111-2222-3333-4444-555555555555 "`,
		`"collector@example.test"`,
		`"Incident Commander"`,
		`"party:11111111-2222-3333-4444-555555555555"`,
		`""`,
		`[]`,
		`{}`,
		`true`,
		`42`,
	}
	for _, rawValue := range invalidValues {
		body := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":` + rawValue + `}]}`
		if _, apiErr := evidence.DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
			t.Fatalf("expected invalid direct party ref %s to fail", rawValue)
		}
	}

	actionPayloadClear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"remove_party_ref","item_ref":"party_ref:` + stablePartyID + `"}]}}]}`
	if _, apiErr := evidence.DecodePatchRequest(strings.NewReader(actionPayloadClear)); apiErr == nil {
		t.Fatalf("expected non-direct clear shape to fail")
	}
}
