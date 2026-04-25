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
