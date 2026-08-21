package tasksdecisions

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestTaskDecisionMutationAdmissionAndReplayHashing(t *testing.T) {
	t.Run("create admission owns fixed surfaces normalization and hash", func(t *testing.T) {
		request, apiErr := DecodeCreateRequest(TaskRequestsViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-task-admission",
			"task.title":"  Collect   endpoint logs  ",
			"task.task_kind":"collection"
		}`))
		if apiErr != nil {
			t.Fatalf("decode task create: %v", apiErr)
		}
		if got := *request.Values["task.title"].Text; got != "Collect   endpoint logs" {
			t.Fatalf("normalized title = %q", got)
		}
		reordered := CreateRequest{
			ViewSchemaID: TaskRequestsViewSchemaID, ClientTxnID: "ignored-by-hash",
			Values: map[string]FieldValue{
				"task.task_kind": request.Values["task.task_kind"],
				"task.title":     request.Values["task.title"],
			},
			Collections: map[string]CollectionActionPayload{},
		}
		if !slices.Equal(CreateRequestHash(request), CreateRequestHash(reordered)) {
			t.Fatal("create hash depends on map order or client transaction ID")
		}
		if _, apiErr := DecodeCreateRequest("cartulary.view.notes.v1", strings.NewReader(`{}`)); apiErr == nil {
			t.Fatal("Task/Decision admission accepted a foreign view")
		}
	})

	t.Run("patch and conflict admission own record-reference collections", func(t *testing.T) {
		recordID := uuid.New()
		body := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":2,
			"client_txn_id":"txn-decision-patch",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"add_record_ref","linked_record_id":"` + recordID.String() + `"}]
			}}]
		}`
		request, apiErr := DecodePatchRequest(strings.NewReader(body))
		if apiErr != nil {
			t.Fatalf("decode decision patch: %v", apiErr)
		}
		change := request.Changes[0]
		if change.Collection == nil || *change.Collection.Actions[0].LinkedRecordID != recordID {
			t.Fatalf("unexpected collection admission: %#v", change)
		}
		claims := ConflictClaims{
			RecordID: uuid.New(), ViewSchemaID: DecisionsViewSchemaID,
			FieldKey: "decision.rationale", CurrentRowVersion: 3,
		}
		conflict, apiErr := DecodeConflictResolveRequest(
			strings.NewReader(`{
				"conflict_token":"opaque-token",
				"resolution_kind":"merged_value",
				"client_txn_id":"txn-decision-resolve",
				"resolved_value":"  Combined rationale.  "
			}`),
			"opaque-token",
			claims,
		)
		if apiErr != nil {
			t.Fatalf("decode decision conflict: %v", apiErr)
		}
		if conflict.Patch == nil || *conflict.Patch.Changes[0].Value.Text != "Combined rationale." {
			t.Fatalf("unexpected conflict admission: %#v", conflict)
		}
		if len(ConflictResolveRequestHash(claims, conflict)) != 32 || len(PatchRequestHash(request)) != 32 {
			t.Fatal("owner hashes are not SHA-256 values")
		}
	})

	t.Run("decision supersede admission and hash are owner local", func(t *testing.T) {
		replacement := uuid.New()
		body := []byte(`{
			"base_row_version":4,
			"client_txn_id":"txn-decision-supersede",
			"reason":"  Use the better-supported decision.  ",
			"replacement_record_id":"` + replacement.String() + `"
		}`)
		request, apiErr := DecodeSupersedeRequest(bytes.NewReader(body))
		if apiErr != nil {
			t.Fatalf("decode decision supersede: %v", apiErr)
		}
		if request.Reason != "Use the better-supported decision." || *request.ReplacementRecordID != replacement {
			t.Fatalf("unexpected supersede admission: %#v", request)
		}
		if len(SupersedeRequestHash(request)) != 32 {
			t.Fatal("supersede hash is not SHA-256")
		}
	})
}
