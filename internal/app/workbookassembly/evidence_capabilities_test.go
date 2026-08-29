package workbookassembly

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
)

func TestEvidenceStoredIdempotencyPayloadCompatibility(t *testing.T) {
	recordID := uuid.MustParse("00000000-0000-4000-8000-000000210101")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000210102")
	changeSetID := uuid.MustParse("00000000-0000-4000-8000-000000210103")
	row := map[string]any{
		"record_id":   recordID.String(),
		"row_version": float64(4),
		"cells":       map[string]any{"evidence.title": map[string]any{"value": "Disk image"}},
	}
	for _, test := range []struct {
		name      string
		operation evidence.OperationID
		kind      evidence.StoredMutationKind
		result    evidence.StoredMutationResult
		status    int
	}{
		{
			name: "create", operation: evidence.OperationCreate, kind: evidence.StoredMutationCreate,
			result: evidence.NewStoredCreateResult(evidence.StoredMutationPayload{
				ViewSchemaID: evidence.ViewSchemaID, IncidentID: incidentID, RecordID: recordID,
				RowVersion: 4, ChangeSetID: &changeSetID, Row: row,
			}),
			status: 201,
		},
		{
			name: "patch", operation: evidence.OperationPatch, kind: evidence.StoredMutationPatch,
			result: evidence.NewStoredPatchResult(evidence.StoredMutationPayload{
				ViewSchemaID: evidence.ViewSchemaID, IncidentID: incidentID, RecordID: recordID,
				RowVersion: 4, ChangeSetID: &changeSetID, Row: row,
			}),
			status: 200,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			kind, ok := storedEvidenceKindForOperation(test.operation)
			if !ok || kind != test.kind || storedEvidenceStatus(kind) != test.status {
				t.Fatalf("operation mapping = (%q,%v,%d), want (%q,true,%d)", kind, ok, storedEvidenceStatus(kind), test.kind, test.status)
			}
			payload, err := encodeStoredEvidenceResult(test.result)
			if err != nil {
				t.Fatalf("encode stored result: %v", err)
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal stored result: %v", err)
			}
			const want = `{"change_set_id":"00000000-0000-4000-8000-000000210103","row":{"cells":{"evidence.title":{"value":"Disk image"}},"record_id":"00000000-0000-4000-8000-000000210101","row_version":4},"view_schema_id":"cartulary.view.evidence.v1"}`
			if string(encoded) != want {
				t.Fatalf("stored JSON = %s, want %s", encoded, want)
			}
			decoded, err := decodeStoredEvidenceResult(test.kind, encoded, incidentID)
			if err != nil {
				t.Fatalf("decode stored result: %v", err)
			}
			stored, ok := decoded.Payload()
			if !ok || stored.RecordID != recordID || stored.RowVersion != 4 || stored.ChangeSetID == nil || *stored.ChangeSetID != changeSetID {
				t.Fatalf("decoded stored result = %#v", stored)
			}
		})
	}

	if kind, ok := storedEvidenceKindForOperation(evidence.OperationConflictResolve); !ok || kind != evidence.StoredMutationPatch {
		t.Fatalf("conflict operation mapping = (%q,%v)", kind, ok)
	}

	for _, test := range []struct {
		operation  evidence.LifecycleOperationID
		wantRoute  string
		wantStatus int
	}{
		{evidence.LifecycleOperationBlobCreate, "object_blobs.create", 201},
		{evidence.LifecycleOperationBlobAttach, "evidence.attach_blob", 200},
	} {
		status, ok := evidenceLifecycleStatus(test.operation)
		if !ok || string(test.operation) != test.wantRoute || status != test.wantStatus {
			t.Fatalf("lifecycle mapping = (%q,%v,%d), want (%q,true,%d)", test.operation, ok, status, test.wantRoute, test.wantStatus)
		}
	}
}
