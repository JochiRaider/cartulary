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
		request, admissionFailure := AdmitCreateJSON(TaskRequestsViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-task-admission",
			"task.title":"  Collect   endpoint logs  ",
			"task.task_kind":"collection"
		}`))
		if admissionFailure != nil {
			t.Fatalf("admit task create: %v", admissionFailure)
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
		if _, admissionFailure := AdmitCreateJSON("cartulary.view.notes.v1", strings.NewReader(`{}`)); admissionFailure == nil {
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
		request, admissionFailure := AdmitPatchJSON(strings.NewReader(body))
		if admissionFailure != nil {
			t.Fatalf("admit decision patch: %v", admissionFailure)
		}
		change := request.Changes[0]
		if change.Collection == nil || *change.Collection.Actions[0].LinkedRecordID != recordID {
			t.Fatalf("unexpected collection admission: %#v", change)
		}
		claims := ConflictClaims{
			RecordID: uuid.New(), ViewSchemaID: DecisionsViewSchemaID,
			FieldKey: "decision.rationale", CurrentRowVersion: 3,
		}
		conflict, admissionFailure := AdmitConflictResolveJSON(
			strings.NewReader(`{
				"conflict_token":"opaque-token",
				"resolution_kind":"merged_value",
				"client_txn_id":"txn-decision-resolve",
				"resolved_value":"  Combined rationale.  "
			}`),
			"opaque-token",
			claims,
		)
		if admissionFailure != nil {
			t.Fatalf("admit decision conflict: %v", admissionFailure)
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
		request, admissionFailure := AdmitSupersedeJSON(bytes.NewReader(body))
		if admissionFailure != nil {
			t.Fatalf("admit decision supersede: %v", admissionFailure)
		}
		if request.Reason != "Use the better-supported decision." || *request.ReplacementRecordID != replacement {
			t.Fatalf("unexpected supersede admission: %#v", request)
		}
		if len(SupersedeRequestHash(request)) != 32 {
			t.Fatal("supersede hash is not SHA-256")
		}
	})

	t.Run("admission failures expose closed semantic detail variants", func(t *testing.T) {
		_, plain := AdmitCreateJSON("cartulary.view.notes.v1", strings.NewReader(`{}`))
		requireAdmissionFailure(t, plain, "view_schema_id", "unknown_view_schema", "", 0, 0, false)
		if got := plain.Error(); got != "tasksdecisions: invalid mutation admission" {
			t.Fatalf("safe admission error = %q", got)
		}

		emptyCollection := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":1,
			"client_txn_id":"txn-empty-collection",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1","actions":[]
			}}]
		}`
		_, empty := AdmitPatchJSON(strings.NewReader(emptyCollection))
		requireAdmissionFailure(
			t, empty, "decision.support_refs.actions", "empty_collection_actions",
			"decision.support_refs", 0, 0, false,
		)

		patchLimit := `{
			"view_schema_id":"cartulary.view.task_requests.v1",
			"base_row_version":1,
			"client_txn_id":"txn-patch-limit",
			"changes":[` + repeatedJSONObjectMembers(33) + `]
		}`
		_, patchCount := AdmitPatchJSON(strings.NewReader(patchLimit))
		requireAdmissionFailure(t, patchCount, "changes", "change_count_exceeded", "", 33, 32, true)

		collectionLimit := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":1,
			"client_txn_id":"txn-collection-limit",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1","actions":[` + repeatedJSONObjectMembers(65) + `]
			}}]
		}`
		_, collectionCount := AdmitPatchJSON(strings.NewReader(collectionLimit))
		requireAdmissionFailure(
			t, collectionCount, "decision.support_refs.actions", "collection_action_count_exceeded",
			"decision.support_refs", 65, 64, true,
		)
	})
}

func requireAdmissionFailure(
	t testing.TB,
	failure *AdmissionFailure,
	wantField string,
	wantReasonCode string,
	wantCollectionFieldKey string,
	wantRequestedCount int,
	wantMaxCount int,
	wantCountLimit bool,
) {
	t.Helper()
	if failure == nil {
		t.Fatal("expected admission failure")
	}
	if failure.Field() != wantField || failure.ReasonCode() != wantReasonCode {
		t.Fatalf(
			"admission identity = (%q, %q), want (%q, %q)",
			failure.Field(), failure.ReasonCode(), wantField, wantReasonCode,
		)
	}
	fieldKey, hasFieldKey := failure.CollectionFieldKey()
	if hasFieldKey != (wantCollectionFieldKey != "") || fieldKey != wantCollectionFieldKey {
		t.Fatalf(
			"collection field key = (%q, %t), want (%q, %t)",
			fieldKey, hasFieldKey, wantCollectionFieldKey, wantCollectionFieldKey != "",
		)
	}
	requestedCount, maxCount, hasCountLimit := failure.CountLimit()
	if hasCountLimit != wantCountLimit || requestedCount != wantRequestedCount || maxCount != wantMaxCount {
		t.Fatalf(
			"count limit = (%d, %d, %t), want (%d, %d, %t)",
			requestedCount, maxCount, hasCountLimit, wantRequestedCount, wantMaxCount, wantCountLimit,
		)
	}
}

func repeatedJSONObjectMembers(count int) string {
	return strings.TrimSuffix(strings.Repeat(`{},`, count), ",")
}
