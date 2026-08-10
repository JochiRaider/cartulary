package workbookassembly

import (
	"encoding/json"
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
			decoded, err := decodeArtifactStoredResult(kind, data)
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
		})
	}

	t.Run("fixed persisted bytes remain replayable", func(t *testing.T) {
		const stored = `{"change_set_id":"22222222-2222-4222-8222-222222222222","link_type":"references_artifact","row":{"record_id":"11111111-1111-4111-8111-111111111111","row_version":4},"source_record_id":"33333333-3333-4333-8333-333333333333","view_schema_id":"cartulary.view.notes.v1"}`
		decoded, err := decodeArtifactStoredResult(artifacts.StoredMutationLinkedNote, []byte(stored))
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
