package collaboration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func TestRecordChangeIntentBuildsSortedCompactPatch(t *testing.T) {
	testRecordChangeIntentBuildsSortedCompactPatch(t)
}

func testRecordChangeIntentBuildsSortedCompactPatch(t *testing.T) {
	changeSetID := uuid.New()
	recordID := uuid.New()
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "test.record-change", SourceOwnerID: "test", RecordTypes: []string{"test"},
		AffectedViews: []ViewPublicationContribution{
			{ViewSchemaID: "cartulary.view.notes.v1", PublicFieldKeys: []string{"note.body", "note.title"}, PatchFieldKeys: []string{"note.body", "note.title"}},
			{ViewSchemaID: "cartulary.view.hosts.v1", PublicFieldKeys: []string{"host.evidence_count"}, PatchFieldKeys: []string{"host.evidence_count"}},
			{ViewSchemaID: "cartulary.view.timeline.v2", PublicFieldKeys: []string{"timeline.evidence_count"}, PatchFieldKeys: []string{"timeline.evidence_count"}},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	appender := &publicationAppender{catalog: catalog}
	patch := collabprotocol.BuildViewRowPatch(map[string]any{
		"record_id": recordID.String(), "row_version": int64(3),
		"cells": map[string]any{"note.title": map[string]any{"value": "Title"}, "note.body": map[string]any{"value": "Body"}},
	}, []string{"note.body", "note.title"})
	intent, err := appender.recordChangedIntent(RecordChangeIntentInput{
		IncidentID:      uuid.New(),
		RecordID:        recordID,
		RowVersion:      3,
		ChangeSetID:     changeSetID,
		ClientTxnID:     "txn-publisher",
		ActorUserID:     uuid.New(),
		PublicFieldKeys: []string{"note.body", "note.title", "note.body"},
		AffectedViews:   []AffectedViewChange{{ViewSchemaID: "cartulary.view.notes.v1", RecordID: recordID, RowVersion: 3, ChangeKind: "patch", PatchCells: patch}},
		MutationOrdinal: 2, CreatedAt: time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC),
	})
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
	decodedPatch := affectedViews[0].(map[string]any)["patch_cells"].(map[string]any)
	cells := decodedPatch["cells"].(map[string]any)
	if cells["note.title"].(map[string]any)["value"] != "Title" {
		t.Fatalf("missing note.title patch: %#v", decodedPatch)
	}
	if cells["note.body"].(map[string]any)["value"] != "Body" {
		t.Fatalf("missing note.body patch: %#v", decodedPatch)
	}

	multiViewIntent, err := appender.recordChangedIntent(RecordChangeIntentInput{
		IncidentID:      uuid.New(),
		RecordID:        recordID,
		RowVersion:      4,
		ChangeSetID:     changeSetID,
		ClientTxnID:     "txn-multi-view",
		ActorUserID:     uuid.New(),
		PublicFieldKeys: []string{"host.evidence_count", "timeline.evidence_count"},
		AffectedViews: []AffectedViewChange{
			{ViewSchemaID: "cartulary.view.timeline.v2", RecordID: recordID, RowVersion: 4, ChangeKind: "invalidate"},
			{ViewSchemaID: "cartulary.view.hosts.v1", RecordID: recordID, RowVersion: 4, ChangeKind: "invalidate"},
		},
		MutationOrdinal: 3, CreatedAt: time.Date(2026, time.July, 26, 12, 1, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	var multiPayload map[string]any
	if err := json.Unmarshal(multiViewIntent.CanonicalPayload, &multiPayload); err != nil {
		t.Fatal(err)
	}
	multiViews := multiPayload["affected_views"].([]any)
	if len(multiViews) != 2 ||
		multiViews[0].(map[string]any)["view_schema_id"] != "cartulary.view.hosts.v1" ||
		multiViews[1].(map[string]any)["view_schema_id"] != "cartulary.view.timeline.v2" {
		t.Fatalf("multi-view intent order = %#v", multiViews)
	}
	streamSeq := int64(11)
	parsed, err := collabprotocol.RecordChangeFromSequencedMessage(collabprotocol.Message{
		Type:       "record_changed",
		IncidentID: "",
		EventID:    uuid.NewString(),
		EmittedAt:  time.Date(2026, time.July, 26, 12, 1, 1, 0, time.UTC).Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    multiViewIntent.CanonicalPayload,
	})
	if err == nil || parsed.RecordID != uuid.Nil {
		t.Fatalf("sequenced parsing without envelope incident identity must fail, got parsed=%#v err=%v", parsed, err)
	}

	incidentID := uuid.New()
	parsed, err = collabprotocol.RecordChangeFromSequencedMessage(collabprotocol.Message{
		Type:       "record_changed",
		IncidentID: incidentID.String(),
		EventID:    uuid.NewString(),
		EmittedAt:  time.Date(2026, time.July, 26, 12, 1, 1, 0, time.UTC).Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    collabprotocol.RawPayload(multiPayload),
	})
	if err != nil {
		t.Fatalf("parse canonical multi-view message: %v", err)
	}
	if len(parsed.AffectedViews) != 2 || parsed.AffectedViews[0].ViewSchemaID != "cartulary.view.hosts.v1" || parsed.AffectedViews[1].ViewSchemaID != "cartulary.view.timeline.v2" {
		t.Fatalf("parsed multi-view changes = %#v", parsed.AffectedViews)
	}

	_, err = appender.recordChangedIntent(RecordChangeIntentInput{
		IncidentID:  uuid.New(),
		RecordID:    recordID,
		RowVersion:  5,
		ChangeSetID: changeSetID,
		ActorUserID: uuid.New(),
		AffectedViews: []AffectedViewChange{
			{ViewSchemaID: "cartulary.view.hosts.v1", RecordID: recordID, RowVersion: 5, ChangeKind: "invalidate"},
			{ViewSchemaID: "cartulary.view.hosts.v1", RecordID: recordID, RowVersion: 5, ChangeKind: "invalidate"},
		},
		MutationOrdinal: 4, CreatedAt: time.Date(2026, time.July, 26, 12, 2, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("duplicate affected views must be rejected")
	}
}
