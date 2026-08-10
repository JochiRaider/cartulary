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
		{name: "create", operation: artifacts.OperationWorkbookCreate, kind: artifacts.StoredMutationCreate},
		{name: "patch", operation: artifacts.OperationWorkbookPatch, kind: artifacts.StoredMutationPatch},
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
