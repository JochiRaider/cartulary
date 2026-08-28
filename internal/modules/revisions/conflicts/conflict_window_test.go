package conflicts

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCanonicalConflictWindowUsesRetainedFactsWithoutTimestampInference_Unit(t *testing.T) {
	fields := []string{
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.tags",
		"timeline.attached_evidence_ids",
	}
	descriptors := make([]FieldDescriptor, 0, len(fields)+1)
	descriptors = append(descriptors, FieldDescriptor{
		FieldKey: "timeline.activity_synopsis_text", ValueKind: "text",
		Writable: true, ConflictResolutionClass: "text_compare_merge",
	})
	for _, fieldKey := range fields {
		descriptors = append(descriptors, FieldDescriptor{
			FieldKey: fieldKey, ValueKind: "collection",
			Writable: true, ConflictResolutionClass: "collection_review",
		})
	}
	set, err := newValidatedFieldDescriptorSet("cartulary.view.timeline.v2", descriptors)
	if err != nil {
		t.Fatalf("build descriptors: %v", err)
	}
	projector, err := NewRevisionSnapshotProjector("fixture.timeline.v1", map[string]string{
		"timeline.activity_synopsis_text": "summary",
	})
	if err != nil {
		t.Fatalf("build projector: %v", err)
	}
	actor := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	collisionTime := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	baseSnapshot := conflictSnapshotJSON(t, "base")
	serverSnapshot := conflictSnapshotJSON(t, "server")
	facts := make([]RevisionConflictFact, 0, len(fields))
	for _, fieldKey := range fields {
		facts = append(facts, RevisionConflictFact{
			FieldKey: fieldKey, BeforePresent: true, BeforeValue: []byte(`{"value":[]}`),
			AfterPresent: true, AfterValue: []byte(`{"value":[{"item_ref":"server"}]}`),
		})
	}
	rows := []RevisionWindowRow{
		{RowVersion: 1, AfterJSON: baseSnapshot, ActorUserID: actor, CreatedAt: collisionTime},
		{RowVersion: 2, BeforeJSON: baseSnapshot, AfterJSON: serverSnapshot, ActorUserID: actor, CreatedAt: collisionTime, ConflictFacts: facts},
		// A same-timestamp row with neither a scalar change nor retained facts
		// must not invent another changed field.
		{RowVersion: 3, BeforeJSON: serverSnapshot, AfterJSON: serverSnapshot, ActorUserID: uuid.New(), CreatedAt: collisionTime},
	}
	window, err := BuildCanonicalPatchConflictWindow(uuid.New(), 1, 3, rows, set, projector)
	if err != nil {
		t.Fatalf("build canonical window: %v", err)
	}
	if len(window.ChangedFields) != len(fields)+1 {
		t.Fatalf("changed fields = %v, want scalar plus four retained collection facts", window.ChangedFields)
	}
	baseCells := window.BaseRow["cells"].(map[string]any)
	for _, fieldKey := range fields {
		changed, ok := window.ChangedFields[fieldKey]
		if !ok || changed.ServerUpdatedBy != actor || !changed.ServerUpdatedAt.Equal(collisionTime) {
			t.Fatalf("changed metadata for %q = %#v", fieldKey, changed)
		}
		if want := map[string]any{"value": []any{}}; !reflect.DeepEqual(baseCells[fieldKey], want) {
			t.Fatalf("base cell %q = %#v, want %#v", fieldKey, baseCells[fieldKey], want)
		}
	}
}

func TestCanonicalConflictWindowRejectsMissingBaseBeforeReadingFacts_Unit(t *testing.T) {
	set, err := newValidatedFieldDescriptorSet("fixture", []FieldDescriptor{{
		FieldKey: "timeline.tags", ValueKind: "collection", Writable: true,
		ConflictResolutionClass: "collection_review",
	}})
	if err != nil {
		t.Fatalf("build descriptors: %v", err)
	}
	projector, err := NewRevisionSnapshotProjector("fixture.timeline.v1", map[string]string{"timeline.tags": "tags"})
	if err != nil {
		t.Fatalf("build projector: %v", err)
	}
	recordID := uuid.New()
	_, err = BuildCanonicalPatchConflictWindow(recordID, 1, 2, []RevisionWindowRow{{
		RowVersion: 2,
		BeforeJSON: conflictSnapshotJSON(t, "base"), AfterJSON: conflictSnapshotJSON(t, "server"),
		ConflictFacts: []RevisionConflictFact{{FieldKey: "timeline.tags", BeforePresent: true, BeforeValue: []byte(`{"value":[]}`)}},
	}}, set, projector)
	var windowErr *RevisionWindowError
	if !errors.As(err, &windowErr) || windowErr.RecordID != recordID {
		t.Fatalf("missing base error = %T %[1]v, want RevisionWindowError", err)
	}
}

func conflictSnapshotJSON(t testing.TB, summary string) []byte {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{
		"snapshot_schema_id": "fixture.timeline.v1",
		"source":             map[string]any{"summary": summary},
		"record":             map[string]any{},
	})
	if err != nil {
		t.Fatalf("encode conflict snapshot: %v", err)
	}
	return encoded
}
