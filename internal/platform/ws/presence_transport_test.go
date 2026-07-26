package ws

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPresenceReplayRevocationTransport(t *testing.T) {
	t.Run("presence snapshots are incident scoped sorted and expire", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		otherIncidentID := uuid.New()
		userID := uuid.New()
		now := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)
		firstConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		secondConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

		hub.UpsertPresence(incidentID, firstConnectionID, userID, "Analyst", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "viewing",
		}, now)
		hub.UpsertPresence(incidentID, secondConnectionID, userID, "Analyst", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "editing",
			RecordID: stringPointer(uuid.NewString()),
			FieldKey: stringPointer("timeline.activity_synopsis_text"),
		}, now)
		hub.UpsertPresence(otherIncidentID, uuid.New(), userID, "Other", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "viewing",
		}, now)

		snapshot := hub.PresenceSnapshot(incidentID, now)
		if got := len(snapshot); got != 2 {
			t.Fatalf("incident snapshot length = %d want 2: %#v", got, snapshot)
		}
		if snapshot[0].ConnectionID != secondConnectionID.String() || snapshot[1].ConnectionID != firstConnectionID.String() {
			t.Fatalf("presence snapshot not sorted by connection_id: %#v", snapshot)
		}
		for _, presence := range snapshot {
			if presence.DisplayName == "Other" {
				t.Fatalf("presence snapshot leaked another incident presence: %#v", snapshot)
			}
		}

		expired := hub.PresenceSnapshot(incidentID, now.Add(PresenceTTL+time.Nanosecond))
		if len(expired) != 0 {
			t.Fatalf("expired presence must be pruned, got %#v", expired)
		}
	})

	t.Run("revocation subscribers preserve public reason codes", func(t *testing.T) {
		hub := NewHub()
		sessionID := uuid.New()
		incidentID := uuid.New()

		sessionRevocations, unregisterSession := hub.RegisterSession(sessionID)
		defer unregisterSession()
		incidentRevocations, unregisterIncident := hub.RegisterIncidentSession(incidentID, sessionID)
		defer unregisterIncident()

		hub.RevokeSession(sessionID, "concurrency_limit")
		requireReason(t, sessionRevocations, "concurrency_limit")

		hub.RevokeIncidentSession(incidentID, sessionID, "incident_access_revoked")
		requireReason(t, incidentRevocations, "incident_access_revoked")
	})

	t.Run("subscription teardown is idempotent and safe during publish", func(t *testing.T) {
		const iterations = 256
		for iteration := 0; iteration < iterations; iteration++ {
			hub := NewHub()
			incidentID := uuid.New()
			messages, unsubscribeIncident := hub.SubscribeIncident(incidentID, 1)
			start := make(chan struct{})
			var workers sync.WaitGroup
			workers.Add(2)

			go func() {
				defer workers.Done()
				<-start
				for publish := 0; publish < 8; publish++ {
					if err := hub.DeliverReplayable(sequencedJobMessage(incidentID, int64(publish*2+1))); err != nil {
						t.Errorf("deliver job progress: %v", err)
						return
					}
					if err := hub.DeliverReplayable(sequencedRecordMessage(incidentID, int64(publish*2+2))); err != nil {
						t.Errorf("deliver record change: %v", err)
						return
					}
				}
			}()
			go func() {
				defer workers.Done()
				<-start
				unsubscribeIncident()
				unsubscribeIncident()
			}()

			close(start)
			workers.Wait()

			for len(messages) > 0 {
				<-messages
			}
			if err := hub.DeliverReplayable(sequencedJobMessage(incidentID, 100)); err != nil {
				t.Fatalf("deliver after unsubscribe: %v", err)
			}
			if err := hub.DeliverReplayable(sequencedRecordMessage(incidentID, 101)); err != nil {
				t.Fatalf("deliver record after unsubscribe: %v", err)
			}
			requireNoIncidentMessage(t, messages)
		}
	})

	t.Run("record changes emit canonical patch cells with invalidate fallback", func(t *testing.T) {
		incidentID := uuid.New()
		recordID := uuid.New()
		changeSetID := uuid.New()
		actorUserID := uuid.New()
		row := map[string]any{
			"record_id":   recordID.String(),
			"row_version": int64(7),
			"cells": map[string]any{
				"timeline.activity_synopsis_text": map[string]any{"value": "Patched"},
				"timeline.capture_state":          map[string]any{"value": "enriched"},
				"timeline.raw_activity_text":      map[string]any{"value": "Omitted"},
			},
			"group_values": map[string]any{
				"timeline.capture_state":     "enriched",
				"timeline.raw_activity_text": "not-a-group-key",
			},
		}
		patch := BuildViewRowPatch(row, []string{
			"timeline.activity_synopsis_text",
			"timeline.capture_state",
		})
		payload := RecordChangePayload(RecordChange{
			IncidentID:       incidentID,
			RecordID:         recordID,
			RowVersion:       7,
			ChangeSetID:      changeSetID,
			ClientTxnID:      "txn-collaboration-patch",
			ActorUserID:      actorUserID,
			ChangedFieldKeys: []string{"timeline.activity_synopsis_text", "timeline.capture_state"},
			ViewSchemaID:     "cartulary.view.timeline.v2",
			PatchCells:       patch,
		})

		changedKeys, _ := payload["changed_field_keys"].([]string)
		if want := []string{"timeline.activity_synopsis_text", "timeline.capture_state"}; !reflect.DeepEqual(changedKeys, want) {
			t.Fatalf("changed_field_keys = %#v want %#v", changedKeys, want)
		}
		affectedViews, _ := payload["affected_views"].([]map[string]any)
		if len(affectedViews) != 1 || affectedViews[0]["change_kind"] != "patch" {
			t.Fatalf("expected one patch affected view, got %#v", payload["affected_views"])
		}
		patchCells, _ := affectedViews[0]["patch_cells"].(map[string]any)
		cells, _ := patchCells["cells"].(map[string]any)
		if _, ok := cells["timeline.raw_activity_text"]; ok {
			t.Fatalf("patch_cells must omit unchanged cells, got %#v", cells)
		}
		if len(cells) != 2 {
			t.Fatalf("patch_cells cells length = %d want 2: %#v", len(cells), cells)
		}
		groupValues, _ := patchCells["group_values"].(map[string]any)
		if !reflect.DeepEqual(groupValues, map[string]any{"timeline.capture_state": "enriched"}) {
			t.Fatalf("patch group_values = %#v", groupValues)
		}

		fallback := RecordChangePayload(RecordChange{
			IncidentID:       incidentID,
			RecordID:         recordID,
			RowVersion:       8,
			ChangeSetID:      changeSetID,
			ClientTxnID:      "txn-collaboration-invalidate",
			ActorUserID:      actorUserID,
			ChangedFieldKeys: []string{"timeline.activity_synopsis_text"},
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})
		fallbackViews, _ := fallback["affected_views"].([]map[string]any)
		if len(fallbackViews) != 1 || fallbackViews[0]["change_kind"] != "invalidate" {
			t.Fatalf("expected invalidate fallback, got %#v", fallback["affected_views"])
		}
	})
}

func sequencedJobMessage(incidentID uuid.UUID, streamSeq int64) Message {
	now := time.Now().UTC()
	return Message{
		Type:       "job_progress",
		IncidentID: incidentID.String(),
		EventID:    uuid.NewString(),
		EmittedAt:  now.Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload: RawPayload(NewIncidentJobProgressPayload(
			"job-subscription-teardown",
			incidentID,
			JobStatusRunning,
			JobProgress{},
			now,
		)),
	}
}

func sequencedRecordMessage(incidentID uuid.UUID, streamSeq int64) Message {
	now := time.Now().UTC()
	return Message{
		Type:       "record_changed",
		IncidentID: incidentID.String(),
		EventID:    uuid.NewString(),
		EmittedAt:  now.Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload: RawPayload(RecordChangePayload(RecordChange{
			IncidentID:       incidentID,
			RecordID:         uuid.New(),
			RowVersion:       1,
			ChangeSetID:      uuid.New(),
			ActorUserID:      uuid.New(),
			ChangedFieldKeys: []string{},
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})),
	}
}

func stringPointer(value string) *string {
	return &value
}

func requireReason(t testing.TB, ch <-chan string, want string) {
	t.Helper()

	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("revocation reason = %q want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for revocation reason %q", want)
	}
}
