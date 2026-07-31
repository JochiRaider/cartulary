package workbook

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestWorkbookWritableConflictCapabilityDerivedFromRegistry_Unit(t *testing.T) {
	allowedClasses := []string{"atomic_replace", "collection_review", "text_compare_merge"}
	recordTypes := map[string]int{}
	for _, resource := range viewschema.ListPublicResources() {
		schema, ok := viewschema.Lookup(resource.ViewSchemaID)
		if !ok {
			t.Fatalf("public resource %s has no internal schema", resource.ViewSchemaID)
		}
		writableFields := 0
		for fieldKey, field := range schema.Fields() {
			if !field.Writable {
				continue
			}
			writableFields++
			if !slices.Contains(allowedClasses, field.ConflictResolutionClass) {
				t.Fatalf("%s field %s has unsupported conflict class %q", resource.ViewSchemaID, fieldKey, field.ConflictResolutionClass)
			}
			switch field.WriteKind {
			case "direct_value":
				if field.WriteTarget == nil || strings.TrimSpace(*field.WriteTarget) == "" {
					t.Fatalf("%s field %s is directly writable without an authoritative write target", resource.ViewSchemaID, fieldKey)
				}
			case "action_payload":
				if field.WriteAction == nil || strings.TrimSpace(*field.WriteAction) == "" {
					t.Fatalf("%s field %s is action-writable without an authoritative write action", resource.ViewSchemaID, fieldKey)
				}
			default:
				t.Fatalf("%s field %s has unsupported write kind %q", resource.ViewSchemaID, fieldKey, field.WriteKind)
			}
		}
		if writableFields == 0 {
			continue
		}
		if len(resource.SourceRecordTypes) != 1 {
			t.Fatalf("conflict-capable surface %s has source record types %v, want exactly one", resource.ViewSchemaID, resource.SourceRecordTypes)
		}
		recordTypes[resource.SourceRecordTypes[0]] += writableFields
		if !slices.Contains(expectedPatchSurfaces()[resource.SourceRecordTypes[0]], resource.ViewSchemaID) {
			t.Fatalf(
				"writable surface %s is absent from the registry-derived patch requirements",
				resource.ViewSchemaID,
			)
		}
	}
	if len(recordTypes) == 0 {
		t.Fatal("machine registry produced no conflict-capable record types")
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

func TestWorkbookMutationDecoderDistinguishesKnownNonWritableAndUnknownFields(t *testing.T) {
	tests := []struct {
		name            string
		fieldKey        string
		wantDetailField string
	}{
		{
			name:            "known non-writable field uses canonical key",
			fieldKey:        "assessment.assessment_state",
			wantDetailField: "assessment.assessment_state",
		},
		{
			name:            "unknown field retains generic key",
			fieldKey:        "assessment.not_a_field",
			wantDetailField: "field_key",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"view_schema_id":"cartulary.view.assessments.v1","base_row_version":1,"client_txn_id":"txn-field-error","changes":[{"field_key":"` + tc.fieldKey + `","value":"cleared"}]}`
			_, apiErr := DecodePatchRequest(strings.NewReader(body))
			if apiErr == nil {
				t.Fatalf("expected %s to be rejected", tc.fieldKey)
			}
			if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected error for %s: %#v", tc.fieldKey, apiErr)
			}
			if got := apiErr.Details["field"]; got != tc.wantDetailField {
				t.Fatalf("detail field for %s = %#v, want %q", tc.fieldKey, got, tc.wantDetailField)
			}
			if got := apiErr.Details["reason_code"]; got != "unsupported_field_key" {
				t.Fatalf("reason for %s = %#v, want unsupported_field_key", tc.fieldKey, got)
			}
		})
	}
}

func TestWorkbookMutationDecoderAdmitsStructurallyValidOwnerSpecificCollectionOps(t *testing.T) {
	stableID := "11111111-2222-3333-4444-555555555555"
	tests := []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		action       string
	}{
		{
			name:         "record refs reject party ref op",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"add_party_ref","party_id":"` + stableID + `"}`,
		},
		{
			name:         "party refs reject record ref op",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.audience_party_ids",
			action:       `{"op":"add_record_ref","linked_record_id":"` + stableID + `"}`,
		},
		{
			name:         "tag collections reject record ref op",
			viewSchemaID: NotesViewSchemaID,
			fieldKey:     "note.tags",
			action:       `{"op":"add_record_ref","linked_record_id":"` + stableID + `"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"view_schema_id":"` + tc.viewSchemaID + `","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"` + tc.fieldKey + `","action_payload":{"kind":"collection_actions_v1","actions":[` + tc.action + `]}}]}`
			if _, apiErr := DecodePatchRequest(strings.NewReader(body)); apiErr != nil {
				t.Fatalf("structurally valid action should reach the source owner for %s: %#v", tc.fieldKey, apiErr)
			}
		})
	}
}

func TestWorkbookMutationDecoderDefersReservedEvidenceStorageRefToOwner(t *testing.T) {
	body := `{"client_txn_id":"txn-reserved-ref","evidence.title":"Reserved ref","evidence.storage_ref":"object://00000000-0000-0000-0000-000000210004"}`
	create, apiErr := DecodeCreateRequest(EvidenceViewSchemaID, strings.NewReader(body))
	if apiErr != nil {
		t.Fatalf("generic create decoder rejected owner-specific storage ref: %#v", apiErr)
	}
	if value := create.Values["evidence.storage_ref"]; value.Text == nil || *value.Text != "object://00000000-0000-0000-0000-000000210004" {
		t.Fatalf("generic create decoder changed storage ref: %#v", value)
	}

	patch := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn-reserved-ref-patch","changes":[{"field_key":"evidence.storage_ref","value":"object://00000000-0000-0000-0000-000000210005"}]}`
	decodedPatch, apiErr := DecodePatchRequest(strings.NewReader(patch))
	if apiErr != nil {
		t.Fatalf("generic patch decoder rejected owner-specific storage ref: %#v", apiErr)
	}
	value := decodedPatch.Changes[0].Value
	if value == nil || value.Text == nil || *value.Text != "object://00000000-0000-0000-0000-000000210005" {
		t.Fatalf("generic patch decoder changed storage ref: %#v", value)
	}
}

func TestWorkbookMutationDecoderCollectionIDsAreExactLexicalTokens(t *testing.T) {
	stableID := "11111111-2222-3333-4444-555555555555"
	tests := []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		action       string
	}{
		{
			name:         "add record ref rejects leading whitespace",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"add_record_ref","linked_record_id":" ` + stableID + `"}`,
		},
		{
			name:         "add record ref rejects trailing whitespace",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"add_record_ref","linked_record_id":"` + stableID + ` "}`,
		},
		{
			name:         "add record ref rejects noncanonical casing",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"add_record_ref","linked_record_id":"11111111-2222-3333-4444-AAAAAAAAAAAA"}`,
		},
		{
			name:         "remove record ref rejects padded item ref",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"remove_record_ref","item_ref":"record_ref:` + stableID + ` "}`,
		},
		{
			name:         "remove record ref rejects noncanonical item ref suffix",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"remove_record_ref","item_ref":"record_ref:11111111-2222-3333-4444-AAAAAAAAAAAA"}`,
		},
		{
			name:         "add party ref rejects padded party id",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.audience_party_ids",
			action:       `{"op":"add_party_ref","party_id":"` + stableID + ` "}`,
		},
		{
			name:         "remove party ref rejects padded item ref",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.audience_party_ids",
			action:       `{"op":"remove_party_ref","item_ref":" party_ref:` + stableID + `"}`,
		},
		{
			name:         "remove risk ref rejects padded item ref",
			viewSchemaID: HandoffViewSchemaID,
			fieldKey:     "handoff.open_risk_refs",
			action:       `{"op":"remove_risk_ref","item_ref":"risk_ref:` + stableID + ` "}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"view_schema_id":"` + tc.viewSchemaID + `","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"` + tc.fieldKey + `","action_payload":{"kind":"collection_actions_v1","actions":[` + tc.action + `]}}]}`
			if _, apiErr := DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
				t.Fatalf("expected inexact identifier to fail for %s", tc.fieldKey)
			} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected error for %s: %#v", tc.fieldKey, apiErr)
			}
		})
	}
}

func TestWorkbookMutationDecoderCollectionRemovalRequiresItemRef(t *testing.T) {
	stableID := "11111111-2222-3333-4444-555555555555"
	tests := []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		action       string
	}{
		{
			name:         "record ref removal rejects linked record id",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			action:       `{"op":"remove_record_ref","linked_record_id":"` + stableID + `"}`,
		},
		{
			name:         "party ref removal rejects party id",
			viewSchemaID: CommLogViewSchemaID,
			fieldKey:     "comm_log.audience_party_ids",
			action:       `{"op":"remove_party_ref","party_id":"` + stableID + `"}`,
		},
		{
			name:         "risk ref removal rejects risk ref id",
			viewSchemaID: HandoffViewSchemaID,
			fieldKey:     "handoff.open_risk_refs",
			action:       `{"op":"remove_risk_ref","risk_ref_id":"` + stableID + `"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"view_schema_id":"` + tc.viewSchemaID + `","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"` + tc.fieldKey + `","action_payload":{"kind":"collection_actions_v1","actions":[` + tc.action + `]}}]}`
			if _, apiErr := DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
				t.Fatalf("expected removal without item_ref to fail for %s", tc.fieldKey)
			} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected error for %s: %#v", tc.fieldKey, apiErr)
			}
		})
	}
}

func TestTaskDecisionRelationshipConfidenceRejected(t *testing.T) {
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

func TestDirectPartyReferenceDecoderAcceptsOnlyExactStableIDs(t *testing.T) {
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

func TestDirectDecisionReferenceDecoderAcceptsOnlyExactStableIDs(t *testing.T) {
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

func TestTaskOwnerNullClearRejectedAtDecoder(t *testing.T) {
	clear := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.owner_user_id","value":null}]}`
	if _, apiErr := DecodePatchRequest(strings.NewReader(clear)); apiErr == nil {
		t.Fatalf("expected task owner null clear to fail")
	} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("unexpected task owner null clear error: %#v", apiErr)
	} else if apiErr.Details["field"] != "task.owner_user_id" || apiErr.Details["reason_code"] != "field_not_nullable" {
		t.Fatalf("unexpected task owner null clear details: %#v", apiErr.Details)
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

	const blobID = "00000000-0000-4000-8000-000000210101"
	evidenceOmittedDefault, apiErr := DecodeCreateRequest(EvidenceViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-evidence-left","evidence.title":"Disk image","evidence.initial_object_blob_id":"`+blobID+`"}`))
	if apiErr != nil {
		t.Fatalf("decode Evidence create input: %#v", apiErr)
	}
	if got := evidenceOmittedDefault.Inputs["evidence.initial_object_blob_id"].UUID.String(); got != blobID {
		t.Fatalf("Evidence create input = %q, want %q", got, blobID)
	}
	evidenceExplicitDefault, apiErr := DecodeCreateRequest(EvidenceViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-evidence-right","evidence.lifecycle_state":"requested","evidence.title":"Disk image","evidence.initial_object_blob_id":"`+blobID+`"}`))
	if apiErr != nil {
		t.Fatalf("decode Evidence explicit default: %#v", apiErr)
	}
	if !bytes.Equal(CreateRequestHash(evidenceOmittedDefault), CreateRequestHash(evidenceExplicitDefault)) {
		t.Fatal("Evidence create hash must equate omitted lifecycle with explicit requested")
	}
	differentBlob, apiErr := DecodeCreateRequest(EvidenceViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-evidence-left","evidence.title":"Disk image","evidence.initial_object_blob_id":"00000000-0000-4000-8000-000000210102"}`))
	if apiErr != nil {
		t.Fatalf("decode divergent Evidence create input: %#v", apiErr)
	}
	if bytes.Equal(CreateRequestHash(evidenceOmittedDefault), CreateRequestHash(differentBlob)) {
		t.Fatal("Evidence create hash must include the initial blob identifier")
	}
	explicitRequestedAtClear, apiErr := DecodeCreateRequest(EvidenceViewSchemaID, strings.NewReader(`{"client_txn_id":"txn-evidence-clear","evidence.title":"Disk image","evidence.requested_at":null,"evidence.initial_object_blob_id":"`+blobID+`"}`))
	if apiErr != nil {
		t.Fatalf("decode explicit requested_at clear: %#v", apiErr)
	}
	if bytes.Equal(CreateRequestHash(evidenceOmittedDefault), CreateRequestHash(explicitRequestedAtClear)) {
		t.Fatal("explicit requested_at null must remain distinct from the omitted timestamp default")
	}
	for name, body := range map[string]string{
		"null":               `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":null}`,
		"malformed":          `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":"not-a-uuid"}`,
		"foreign view input": `{"client_txn_id":"txn","party.display_name":"Acme","evidence.initial_object_blob_id":"` + blobID + `"}`,
	} {
		viewSchemaID := EvidenceViewSchemaID
		if name == "foreign view input" {
			viewSchemaID = PartiesViewSchemaID
		}
		if _, apiErr := DecodeCreateRequest(viewSchemaID, strings.NewReader(body)); apiErr == nil {
			t.Fatalf("%s create input unexpectedly accepted", name)
		} else if apiErr.Details["field"] != "evidence.initial_object_blob_id" {
			t.Fatalf("%s error field = %#v", name, apiErr.Details)
		}
	}
}
