package workbook

import (
	"bytes"
	"strings"
	"testing"
)

func TestWorkbookMutationDecoderValidation(t *testing.T) {
	tests := []struct {
		name         string
		viewSchemaID string
		body         string
	}{
		{
			name:         "party requires display name",
			viewSchemaID: PartiesViewSchemaID,
			body:         `{"client_txn_id":"txn-party","party.party_kind":"organization"}`,
		},
		{
			name:         "comm log requires summary",
			viewSchemaID: CommLogViewSchemaID,
			body:         `{"client_txn_id":"txn-comm","comm_log.comm_type":"briefing","comm_log.audience":"leadership","comm_log.channel_or_meeting":"Bridge"}`,
		},
		{
			name:         "handoff requires incoming owner",
			viewSchemaID: HandoffViewSchemaID,
			body:         `{"client_txn_id":"txn-handoff","handoff.current_state_summary":"state"}`,
		},
		{
			name:         "status review requires summary",
			viewSchemaID: StatusReviewViewSchemaID,
			body:         `{"client_txn_id":"txn-status","status_review.active_risks_summary":"risk"}`,
		},
		{
			name:         "lesson requires summary",
			viewSchemaID: LessonViewSchemaID,
			body:         `{"client_txn_id":"txn-lesson","lesson.closure_state":"open"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request, apiErr := DecodeCreateRequest(tc.viewSchemaID, strings.NewReader(tc.body))
			if apiErr != nil {
				t.Fatalf("decode create request: %#v", apiErr)
			}
			if err := validateCreateRequest(request); err == nil {
				t.Fatalf("expected required-field validation failure")
			}
		})
	}
}

func TestWorkbookMutationDecoderRejectsCollectionReplacement(t *testing.T) {
	rawArray := `{"view_schema_id":"cartulary.view.comm_log.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"comm_log.decision_ids","value":[]}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(rawArray)); apiErr == nil {
		t.Fatalf("expected raw array collection patch to fail")
	}

	rawNull := `{"view_schema_id":"cartulary.view.comm_log.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"comm_log.decision_ids","value":null}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(rawNull)); apiErr == nil {
		t.Fatalf("expected raw null collection patch to fail")
	}

	unknownAction := `{"view_schema_id":"cartulary.view.comm_log.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"comm_log.decision_ids","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"replace_all","linked_record_id":"00000000-0000-0000-0000-000000000001"}]}}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(unknownAction)); apiErr == nil {
		t.Fatalf("expected unknown collection action to fail")
	}
}

func TestPhase9TaskDecisionRelationshipConfidenceRejected(t *testing.T) {
	recordID := "11111111-2222-3333-4444-555555555555"
	tests := []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		createBody   string
		patchBody    string
	}{
		{
			name:         "task linked records",
			viewSchemaID: TaskRequestsViewSchemaID,
			fieldKey:     "task.linked_record_ids",
			createBody: `{
				"client_txn_id":"txn-task-confidence-create",
				"task.title":"Collect logs",
				"task.task_kind":"collection",
				"task.linked_record_ids":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}
			}`,
			patchBody: `{
				"view_schema_id":"cartulary.view.task_requests.v1",
				"base_row_version":1,
				"client_txn_id":"txn-task-confidence-patch",
				"changes":[{"field_key":"task.linked_record_ids","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}}]
			}`,
		},
		{
			name:         "decision support refs",
			viewSchemaID: DecisionsViewSchemaID,
			fieldKey:     "decision.support_refs",
			createBody: `{
				"client_txn_id":"txn-decision-support-confidence-create",
				"decision.summary":"Contain endpoint",
				"decision.decision_type":"containment",
				"decision.rationale":"Evidence supports containment.",
				"decision.support_refs":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}
			}`,
			patchBody: `{
				"view_schema_id":"cartulary.view.decisions.v1",
				"base_row_version":1,
				"client_txn_id":"txn-decision-support-confidence-patch",
				"changes":[{"field_key":"decision.support_refs","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}}]
			}`,
		},
		{
			name:         "decision affected records",
			viewSchemaID: DecisionsViewSchemaID,
			fieldKey:     "decision.affected_record_ids",
			createBody: `{
				"client_txn_id":"txn-decision-affected-confidence-create",
				"decision.summary":"Contain endpoint",
				"decision.decision_type":"containment",
				"decision.rationale":"Containment affects the endpoint.",
				"decision.affected_record_ids":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}
			}`,
			patchBody: `{
				"view_schema_id":"cartulary.view.decisions.v1",
				"base_row_version":1,
				"client_txn_id":"txn-decision-affected-confidence-patch",
				"changes":[{"field_key":"decision.affected_record_ids","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"` + recordID + `","confidence":100}]}}]
			}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name+" create", func(t *testing.T) {
			if _, apiErr := DecodeCreateRequest(tc.viewSchemaID, strings.NewReader(tc.createBody)); apiErr == nil {
				t.Fatalf("expected create confidence on %s to fail", tc.fieldKey)
			} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected create error for %s: %#v", tc.fieldKey, apiErr)
			}
		})
		t.Run(tc.name+" patch", func(t *testing.T) {
			if _, apiErr := DecodePatchRequest(strings.NewReader(tc.patchBody)); apiErr == nil {
				t.Fatalf("expected patch confidence on %s to fail", tc.fieldKey)
			} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected patch error for %s: %#v", tc.fieldKey, apiErr)
			}
		})
	}
}

func TestSupportPhase9_DirectPartyReferenceDecoderAcceptsOnlyExactStableIDs(t *testing.T) {
	stablePartyID := "11111111-2222-3333-4444-555555555555"
	valid := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":"` + stablePartyID + `"}]}`
	request, apiErr := DecodePatchRequest(strings.NewReader(valid))
	if apiErr != nil {
		t.Fatalf("expected exact stable party id to decode: %#v", apiErr)
	}
	if got := request.Changes[0].Value.UUID.String(); got != stablePartyID {
		t.Fatalf("unexpected decoded party id: got %s want %s", got, stablePartyID)
	}

	clear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":null}]}`
	request, apiErr = DecodePatchRequest(strings.NewReader(clear))
	if apiErr != nil {
		t.Fatalf("expected direct null clear to decode: %#v", apiErr)
	}
	if request.Changes[0].Value == nil || request.Changes[0].Value.Kind != "null" {
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
		if _, apiErr := DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
			t.Fatalf("expected invalid direct party ref %s to fail", rawValue)
		}
	}

	actionPayloadClear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"remove_party_ref","item_ref":"party_ref:` + stablePartyID + `"}]}}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(actionPayloadClear)); apiErr == nil {
		t.Fatalf("expected non-direct clear shape to fail")
	}
}

func TestSupportPhase9_DirectDecisionReferenceDecoderAcceptsOnlyExactStableIDs(t *testing.T) {
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
	if request.Changes[0].Value == nil || request.Changes[0].Value.Kind != "null" {
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

func TestWorkbookMutationRequestHashNormalization(t *testing.T) {
	left, apiErr := DecodePatchRequest(strings.NewReader(`{"view_schema_id":"cartulary.view.comm_log.v1","base_row_version":1,"client_txn_id":"txn-left","changes":[{"field_key":"comm_log.summary","value":" Updated "},{"field_key":"comm_log.privilege_tag","value":"legal"}]}`))
	if apiErr != nil {
		t.Fatalf("decode left patch: %#v", apiErr)
	}
	right, apiErr := DecodePatchRequest(strings.NewReader(`{"view_schema_id":"cartulary.view.comm_log.v1","base_row_version":1,"client_txn_id":"txn-right","changes":[{"field_key":"comm_log.privilege_tag","value":"legal"},{"field_key":"comm_log.summary","value":"Updated"}]}`))
	if apiErr != nil {
		t.Fatalf("decode right patch: %#v", apiErr)
	}
	if !bytes.Equal(PatchRequestHash(left), PatchRequestHash(right)) {
		t.Fatalf("patch hash should ignore client_txn_id and outer change order while using normalized values")
	}

	createLeft, apiErr := DecodeCreateRequest(PartiesViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-left","party.display_name":" Acme ","party.party_kind":"organization"}`))
	if apiErr != nil {
		t.Fatalf("decode left create: %#v", apiErr)
	}
	createRight, apiErr := DecodeCreateRequest(PartiesViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-right","party.party_kind":"organization","party.display_name":"Acme"}`))
	if apiErr != nil {
		t.Fatalf("decode right create: %#v", apiErr)
	}
	if !bytes.Equal(CreateRequestHash(createLeft), CreateRequestHash(createRight)) {
		t.Fatalf("create hash should ignore client_txn_id and object member order while using normalized values")
	}
}
