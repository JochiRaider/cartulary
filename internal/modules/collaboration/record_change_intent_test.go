package collaboration_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func testRecordChangeIntentBuildsSortedCompactPatch(t *testing.T) {
	changeSetID := uuid.New()
	recordID := uuid.New()
	intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
		IncidentID:       uuid.New(),
		RecordID:         recordID,
		RowVersion:       3,
		ChangeSetID:      changeSetID,
		ClientTxnID:      "txn-publisher",
		ActorUserID:      uuid.New(),
		ChangedFieldKeys: []string{"note.body", "note.title", "note.body"},
		ViewSchemaID:     "cartulary.view.notes.v1",
		Row: map[string]any{
			"record_id":   recordID.String(),
			"row_version": int64(3),
			"cells": map[string]any{
				"note.title": map[string]any{"value": "Title"},
				"note.body":  map[string]any{"value": "Body"},
			},
		},
	}, 2, time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := intent.IntentKey, "record_changed:"+changeSetID.String()+":"+recordID.String()+":3"; got != want {
		t.Fatalf("intent key = %q want %q", got, want)
	}
	if intent.MutationOrdinal != 2 {
		t.Fatalf("mutation ordinal = %d want 2", intent.MutationOrdinal)
	}

	var payload map[string]any
	if err := json.Unmarshal(intent.CanonicalPayload, &payload); err != nil {
		t.Fatal(err)
	}
	changedKeys := payload["changed_field_keys"].([]any)
	if got, want := changedKeys, []any{"note.body", "note.title"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("changed keys = %#v want %#v", got, want)
	}
	affectedViews := payload["affected_views"].([]any)
	patch := affectedViews[0].(map[string]any)["patch_cells"].(map[string]any)
	cells := patch["cells"].(map[string]any)
	if cells["note.title"].(map[string]any)["value"] != "Title" {
		t.Fatalf("missing note.title patch: %#v", patch)
	}
	if cells["note.body"].(map[string]any)["value"] != "Body" {
		t.Fatalf("missing note.body patch: %#v", patch)
	}
}
