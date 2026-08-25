package workbookassembly

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
)

func TestTaskDecisionAdmissionFailureMapping_Unit(t *testing.T) {
	tests := []struct {
		name       string
		admit      func() *tasksdecisions.AdmissionFailure
		wantField  string
		wantReason string
	}{
		{
			name: "plain",
			admit: func() *tasksdecisions.AdmissionFailure {
				_, failure := tasksdecisions.AdmitCreateJSON("cartulary.view.notes.v1", strings.NewReader(`{}`))
				return failure
			},
			wantField:  "view_schema_id",
			wantReason: "unknown_view_schema",
		},
		{
			name: "collection context without counts",
			admit: func() *tasksdecisions.AdmissionFailure {
				_, failure := tasksdecisions.AdmitPatchJSON(strings.NewReader(`{
					"view_schema_id":"cartulary.view.decisions.v1",
					"base_row_version":1,
					"client_txn_id":"txn-empty-collection",
					"changes":[{"field_key":"decision.support_refs","action_payload":{
						"kind":"collection_actions_v1","actions":[]
					}}]
				}`))
				return failure
			},
			wantField:  "decision.support_refs.actions",
			wantReason: "empty_collection_actions",
		},
		{
			name: "patch count limit",
			admit: func() *tasksdecisions.AdmissionFailure {
				_, failure := tasksdecisions.AdmitPatchJSON(strings.NewReader(`{
					"view_schema_id":"cartulary.view.task_requests.v1",
					"base_row_version":1,
					"client_txn_id":"txn-patch-limit",
					"changes":[` + repeatedObjects(33) + `]
				}`))
				return failure
			},
			wantField:  "changes",
			wantReason: "change_count_exceeded",
		},
		{
			name: "collection count limit",
			admit: func() *tasksdecisions.AdmissionFailure {
				_, failure := tasksdecisions.AdmitPatchJSON(strings.NewReader(`{
					"view_schema_id":"cartulary.view.decisions.v1",
					"base_row_version":1,
					"client_txn_id":"txn-collection-limit",
					"changes":[{"field_key":"decision.support_refs","action_payload":{
						"kind":"collection_actions_v1","actions":[` + repeatedObjects(65) + `]
					}}]
				}`))
				return failure
			},
			wantField:  "decision.support_refs.actions",
			wantReason: "collection_action_count_exceeded",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			admissionFailure := test.admit()
			if admissionFailure == nil {
				t.Fatal("admission unexpectedly succeeded")
			}
			mapped := taskDecisionAdmissionFailure(admissionFailure)
			if mapped.Kind() != workbook.MutationFailureInvalidPayload {
				t.Fatalf("failure kind = %q", mapped.Kind())
			}
			field, reason, ok := mapped.InvalidPayloadDetail()
			if !ok || field != test.wantField || reason != test.wantReason {
				t.Fatalf(
					"mapped identity = (%q, %q, %t), want (%q, %q, true)",
					field, reason, ok, test.wantField, test.wantReason,
				)
			}
		})
	}
}

func repeatedObjects(count int) string {
	return strings.TrimSuffix(strings.Repeat(`{},`, count), ",")
}
