package collaboration

import (
	"testing"

	"github.com/google/uuid"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

func TestRecordChangePublisherPublishesSortedCompactPatch(t *testing.T) {
	hub := platformws.NewHub()
	publisher := NewRecordChangePublisher(hub)
	changes, unsubscribe := hub.SubscribeRecordChanges(1)
	defer unsubscribe()

	recordID := uuid.New()
	publisher.Publish(RecordChange{
		IncidentID:       uuid.New(),
		RecordID:         recordID,
		RowVersion:       3,
		ChangeSetID:      uuid.New(),
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
	})

	change := <-changes
	if change.RecordID != recordID {
		t.Fatalf("record_id = %s want %s", change.RecordID, recordID)
	}
	if got, want := change.ChangedFieldKeys, []string{"note.body", "note.title"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("changed keys = %#v want %#v", got, want)
	}
	cells := change.PatchCells["cells"].(map[string]any)
	if cells["note.title"].(map[string]any)["value"] != "Title" {
		t.Fatalf("missing note.title patch: %#v", change.PatchCells)
	}
	if cells["note.body"].(map[string]any)["value"] != "Body" {
		t.Fatalf("missing note.body patch: %#v", change.PatchCells)
	}
}
