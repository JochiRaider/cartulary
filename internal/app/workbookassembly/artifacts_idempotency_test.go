package workbookassembly

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
)

func TestArtifactStoredIdempotencyPayloadCompatibility(t *testing.T) {
	t.Parallel()
	if got := string(artifacts.OperationCreate); got != "workbook.rows.create" {
		t.Fatalf("create operation = %q", got)
	}
	if got := string(artifacts.OperationPatch); got != "workbook.records.patch" {
		t.Fatalf("patch operation = %q", got)
	}
	if got := string(artifacts.OperationConflictResolve); got != "workbook.records.conflicts.resolve" {
		t.Fatalf("conflict operation = %q", got)
	}
	if got := string(artifacts.OperationLinkedNoteCreate); got != "workbook.records.linked_notes.create" {
		t.Fatalf("linked-note operation = %q", got)
	}
	recordID := uuid.New()
	incidentID := uuid.New()
	changeSetID := uuid.New()
	sourceRecordID := uuid.New()
	base := map[string]any{
		"view_schema_id": artifacts.NotesViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row": map[string]any{
			"record_id":   recordID.String(),
			"row_version": float64(4),
		},
	}
	tests := []struct {
		name       string
		operation  artifacts.OperationID
		kind       artifacts.StoredMutationKind
		additional map[string]any
	}{
		{name: "create", operation: artifacts.OperationCreate, kind: artifacts.StoredMutationCreate},
		{name: "patch", operation: artifacts.OperationPatch, kind: artifacts.StoredMutationPatch},
		{name: "conflict patch", operation: artifacts.OperationConflictResolve, kind: artifacts.StoredMutationPatch},
		{name: "linked note", operation: artifacts.OperationLinkedNoteCreate, kind: artifacts.StoredMutationLinkedNote, additional: map[string]any{
			"source_record_id": sourceRecordID.String(),
			"link_type":        "references_artifact",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := cloneArtifactPayload(t, base)
			for key, value := range tc.additional {
				payload[key] = value
			}
			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal legacy payload: %v", err)
			}
			kind, ok := artifactStoredKindForOperation(tc.operation)
			if !ok || kind != tc.kind {
				t.Fatalf("operation %q kind = %q, %v; want %q", tc.operation, kind, ok, tc.kind)
			}
			decoded, err := decodeArtifactStoredResult(kind, data, incidentID)
			if err != nil {
				t.Fatalf("decode existing payload: %v", err)
			}
			encoded, err := encodeArtifactStoredResult(decoded)
			if err != nil {
				t.Fatalf("encode durable payload: %v", err)
			}
			if !reflect.DeepEqual(encoded, payload) {
				t.Fatalf("round trip payload = %#v, want %#v", encoded, payload)
			}
			stored, ok := decoded.Payload()
			if !ok || stored.IncidentID != incidentID || stored.RecordID != recordID ||
				stored.RowVersion != 4 || stored.ChangeSetID == nil || *stored.ChangeSetID != changeSetID {
				t.Fatalf("complete stored result = %#v, %v", stored, ok)
			}
			stored.Row["mutated"] = true
			again, _ := decoded.Payload()
			if _, leaked := again.Row["mutated"]; leaked {
				t.Fatal("stored result payload exposed mutable row state")
			}
		})
	}

	t.Run("malformed incomplete and wrong-kind payloads fail closed", func(t *testing.T) {
		invalid := []struct {
			name string
			kind artifacts.StoredMutationKind
			body map[string]any
		}{
			{name: "missing row", kind: artifacts.StoredMutationCreate, body: map[string]any{"view_schema_id": artifacts.NotesViewSchemaID, "change_set_id": changeSetID.String()}},
			{name: "missing row version", kind: artifacts.StoredMutationCreate, body: map[string]any{"view_schema_id": artifacts.NotesViewSchemaID, "change_set_id": changeSetID.String(), "row": map[string]any{"record_id": recordID.String()}}},
			{name: "zero row version", kind: artifacts.StoredMutationPatch, body: map[string]any{"view_schema_id": artifacts.NotesViewSchemaID, "change_set_id": changeSetID.String(), "row": map[string]any{"record_id": recordID.String(), "row_version": float64(0)}}},
			{name: "wrong create kind", kind: artifacts.StoredMutationCreate, body: func() map[string]any {
				value := cloneArtifactPayload(t, base)
				value["source_record_id"] = sourceRecordID.String()
				value["link_type"] = "references_artifact"
				return value
			}()},
			{name: "incomplete linked note", kind: artifacts.StoredMutationLinkedNote, body: cloneArtifactPayload(t, base)},
			{name: "unknown result field", kind: artifacts.StoredMutationPatch, body: func() map[string]any {
				value := cloneArtifactPayload(t, base)
				value["unexpected"] = true
				return value
			}()},
		}
		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(tc.body)
				if err != nil {
					t.Fatalf("marshal invalid fixture: %v", err)
				}
				if _, err := decodeArtifactStoredResult(tc.kind, data, incidentID); err == nil {
					t.Fatalf("invalid stored payload was accepted: %s", data)
				}
			})
		}
	})

	t.Run("fixed persisted bytes remain replayable", func(t *testing.T) {
		const stored = `{"change_set_id":"22222222-2222-4222-8222-222222222222","link_type":"references_artifact","row":{"record_id":"11111111-1111-4111-8111-111111111111","row_version":4},"source_record_id":"33333333-3333-4333-8333-333333333333","view_schema_id":"cartulary.view.notes.v1"}`
		decoded, err := decodeArtifactStoredResult(artifacts.StoredMutationLinkedNote, []byte(stored), uuid.MustParse("44444444-4444-4444-8444-444444444444"))
		if err != nil {
			t.Fatalf("decode fixed persisted payload: %v", err)
		}
		encoded, err := encodeArtifactStoredResult(decoded)
		if err != nil {
			t.Fatalf("encode fixed persisted payload: %v", err)
		}
		actual, err := json.Marshal(encoded)
		if err != nil {
			t.Fatalf("marshal fixed persisted payload: %v", err)
		}
		if string(actual) != stored {
			t.Fatalf("fixed persisted payload changed:\n got %s\nwant %s", actual, stored)
		}
	})
}

func TestArtifactMutationOutcomeCompatibility(t *testing.T) {
	t.Parallel()
	changeSetID := uuid.New()
	base := artifacts.MutationResult{
		IncidentID: uuid.New(), RecordID: uuid.New(), ChangeSetID: &changeSetID,
		ClientTxnID: "txn-outcome", RowVersion: 3, ViewSchemaID: artifacts.NotesViewSchemaID,
		Row: map[string]any{"row_version": float64(3)},
	}
	tests := []struct {
		name     string
		outcome  artifacts.MutationOutcome
		status   int
		replayed bool
	}{
		{name: "created", outcome: artifacts.MutationOutcomeCreated, status: http.StatusCreated},
		{name: "updated", outcome: artifacts.MutationOutcomeUpdated, status: http.StatusOK},
		{name: "kept saved", outcome: artifacts.MutationOutcomeKeptSaved, status: http.StatusOK},
		{name: "replayed", outcome: artifacts.MutationOutcomeReplayed, status: http.StatusOK, replayed: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			input.Outcome = tc.outcome
			result := artifactMutationResult(input)
			if result.StatusCode != tc.status || result.Replayed != tc.replayed || result.ChangeSetID != changeSetID {
				t.Fatalf("outcome %q translated to %#v", tc.outcome, result)
			}
		})
	}

	withoutChangeSet := base
	withoutChangeSet.Outcome = artifacts.MutationOutcomeKeptSaved
	withoutChangeSet.ChangeSetID = nil
	result := artifactMutationResult(withoutChangeSet)
	if result.ChangeSetID != uuid.Nil {
		t.Fatalf("absent change set translated to %s", result.ChangeSetID)
	}
	if _, present := result.Payload["change_set_id"]; present {
		t.Fatalf("absent change set leaked into payload: %#v", result.Payload)
	}

	if artifactStoredStatus(artifacts.StoredMutationCreate) != http.StatusCreated ||
		artifactStoredStatus(artifacts.StoredMutationLinkedNote) != http.StatusCreated ||
		artifactStoredStatus(artifacts.StoredMutationPatch) != http.StatusOK {
		t.Fatal("stored mutation status mapping changed")
	}
}

func cloneArtifactPayload(t testing.TB, input map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal payload clone: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal payload clone: %v", err)
	}
	return result
}
