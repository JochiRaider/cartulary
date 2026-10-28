package protocoltest

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestObservedRecordEvent(t *testing.T) {
	incident, record, actor, changeSet := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	payload := map[string]any{
		"record_id": record.String(), "row_version": int64(9007199254740993),
		"change_set_id": changeSet.String(), "actor_user_id": actor.String(), "client_txn_id": "observed",
		"changed_field_keys": []string{"note.body", "note.title"},
		"affected_views": []map[string]any{{"view_schema_id": "cartulary.view.notes.v1", "change_kind": "patch", "patch_cells": map[string]any{
			"record_id": record.String(), "row_version": int64(9007199254740993),
			"cells": map[string]any{"note.body": map[string]any{"value": nil}},
		}}},
		"future_member": "retained",
	}
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	message := SequencedMessage(t, "record_changed", incident, "observed-event", 17, now, payload)
	got, err := RecordChangeFromSequencedMessage(message)
	if err != nil {
		t.Fatal(err)
	}
	if got.IncidentID != incident || got.RecordID != record || got.ActorUserID != actor || got.ChangeSetID != changeSet || got.RowVersion != 9007199254740993 || got.StreamSeq != 17 || got.EventID != "observed-event" || !got.EmittedAt.Equal(now) {
		t.Fatalf("lost observed identity or precision: %#v", got)
	}
	if !bytes.Equal(got.Payload, message.Payload) {
		t.Fatal("original payload changed")
	}
	cells := got.AffectedViews[0].PatchCells["cells"].(map[string]any)
	if value, present := cells["note.body"].(map[string]any)["value"]; !present || value != nil {
		t.Fatal("explicit null lost")
	}
	if _, present := cells["note.title"]; present {
		t.Fatal("missing cell reconstructed")
	}
	for _, tc := range []struct{ name, old, replacement string }{
		{"unsorted keys", `["note.body","note.title"]`, `["note.title","note.body"]`},
		{"duplicate keys", `["note.body","note.title"]`, `["note.body","note.body"]`},
		{"missing keys", `"changed_field_keys":["note.body","note.title"],`, ``},
		{"null keys", `["note.body","note.title"]`, `null`},
		{"invalid change kind", `"change_kind":"patch"`, `"change_kind":"invalid"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := message
			bad.Payload = json.RawMessage(strings.Replace(string(message.Payload), tc.old, tc.replacement, 1))
			if bytes.Equal(bad.Payload, message.Payload) {
				t.Fatal("invalid fixture did not change")
			}
			if _, err := RecordChangeFromSequencedMessage(bad); err == nil {
				t.Fatal("invalid observation was repaired or accepted")
			}
		})
	}
	message.IncidentID = ""
	if _, err := RecordChangeFromSequencedMessage(message); err == nil {
		t.Fatal("missing envelope identity accepted")
	}
	message.Type = "job_progress"
	if _, err := RecordChangeFromSequencedMessage(message); err == nil {
		t.Fatal("wrong event family accepted")
	}
}
