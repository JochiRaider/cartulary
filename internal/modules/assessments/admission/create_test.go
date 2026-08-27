package admission

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCreateRequestAdmissionCompatibility(t *testing.T) {
	t.Parallel()
	input, apiErr := DecodeCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-assessment-owner",
		"assessment.subject_ref":"10000000-0000-4000-8000-000000000001",
		"assessment.subject_type":" host ",
		"assessment.assessment_state":" confirmed ",
		"assessment.confidence_score":73,
		"assessment.rationale":"  First\r\nSecond  ",
		"assessment.assessor":"40000000-0000-4000-8000-000000000004",
		"assessment.assessed_at":"2026-08-20T12:34:56.123456789-04:00",
		"assessment.support_refs":{
			"kind":"collection_actions_v1",
			"actions":[
				{"op":"add_record_ref","linked_record_id":"30000000-0000-4000-8000-000000000003"},
				{"op":"add_record_ref","linked_record_id":"20000000-0000-4000-8000-000000000002"}
			]
		}
	}`))
	if apiErr != nil {
		t.Fatalf("decode Assessment create request: %#v", apiErr)
	}
	if input.SubjectType != "host" || input.AssessmentState != "confirmed" || input.Rationale != "First\nSecond" {
		t.Fatalf("normalized Assessment input = %#v", input)
	}
	if input.ConfidenceScore == nil || *input.ConfidenceScore != 73 {
		t.Fatalf("confidence score = %#v", input.ConfidenceScore)
	}
	wantTime := time.Date(2026, 8, 20, 16, 34, 56, 123456789, time.UTC)
	if input.AssessedAt == nil || !input.AssessedAt.Equal(wantTime) {
		t.Fatalf("assessed_at = %#v, want %s", input.AssessedAt, wantTime)
	}
	if len(input.SupportRefs) != 2 || input.SupportRefs[0].String() != "30000000-0000-4000-8000-000000000003" {
		t.Fatalf("support refs = %v", input.SupportRefs)
	}
}

func TestCreateRequestAdmissionEnforcesSupportReferenceLimit(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		count      int
		wantReason string
	}{
		{name: "maximum accepted", count: 64},
		{name: "over maximum rejected", count: 65, wantReason: "collection_action_count_exceeded"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input, apiErr := DecodeCreateRequest(strings.NewReader(supportReferencesBody(test.count)))
			if test.wantReason == "" {
				if apiErr != nil || len(input.SupportRefs) != test.count {
					t.Fatalf("support reference admission = input:%#v error:%#v", input, apiErr)
				}
				return
			}
			if apiErr == nil || apiErr.Details["reason_code"] != test.wantReason || apiErr.Details["max_count"] != 64 {
				t.Fatalf("support reference limit error = %#v", apiErr)
			}
		})
	}
}

func supportReferencesBody(count int) string {
	actions := make([]string, 0, count)
	for index := 0; index < count; index++ {
		actions = append(actions, fmt.Sprintf(
			`{"op":"add_record_ref","linked_record_id":"20000000-0000-4000-8000-%012d"}`,
			index+1,
		))
	}
	return `{"client_txn_id":"txn-limit","assessment.support_refs":{"kind":"collection_actions_v1","actions":[` +
		strings.Join(actions, ",") + `]}}`
}

func TestCreateRequestAdmissionRejectsNonContractualShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       string
		wantField  string
		wantReason string
	}{
		{
			name:       "unknown field",
			body:       `{"client_txn_id":"txn","assessment.future":true}`,
			wantField:  "assessment.future",
			wantReason: "unknown_field",
		},
		{
			name:       "noncanonical subject UUID",
			body:       `{"client_txn_id":"txn","assessment.subject_ref":"10000000-0000-4000-8000-00000000000A"}`,
			wantField:  "assessment.subject_ref",
			wantReason: "invalid_value",
		},
		{
			name:       "string confidence",
			body:       `{"client_txn_id":"txn","assessment.confidence_score":"73"}`,
			wantField:  "assessment.confidence_score",
			wantReason: "invalid_value",
		},
		{
			name: "remove support reference",
			body: `{"client_txn_id":"txn","assessment.support_refs":{"kind":"collection_actions_v1","actions":[` +
				`{"op":"remove_record_ref","linked_record_id":"20000000-0000-4000-8000-000000000002"}]}}`,
			wantField:  "assessment.support_refs",
			wantReason: "invalid_value",
		},
		{
			name:       "empty support references",
			body:       `{"client_txn_id":"txn","assessment.support_refs":{"kind":"collection_actions_v1","actions":[]}}`,
			wantField:  "assessment.support_refs.actions",
			wantReason: "empty_collection_actions",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, apiErr := DecodeCreateRequest(strings.NewReader(test.body))
			if apiErr == nil {
				t.Fatal("expected invalid mutation payload")
			}
			if apiErr.Code != "invalid_mutation_payload" || apiErr.Details["field"] != test.wantField || apiErr.Details["reason_code"] != test.wantReason {
				t.Fatalf("error = %#v", apiErr)
			}
		})
	}
}
