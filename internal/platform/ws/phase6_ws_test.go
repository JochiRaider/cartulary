package ws

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase6_PresenceHeartbeatRevocation_U_6_08(t *testing.T) {
	t.Run("resume replay filters to replayable incident messages", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		sessionID := uuid.New()
		clientInstanceID := "phase6-u-6-08-client"
		now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

		token, _, err := hub.IssueResumeToken(sessionID, incidentID, clientInstanceID, now.Add(time.Hour), now)
		if err != nil {
			t.Fatalf("issue resume token: %v", err)
		}

		hub.mu.Lock()
		hub.highWater[incidentID] = 4
		hub.replay[incidentID] = []replayEntry{
			{Message: replayMessage("record_changed", incidentID, 1), StoredAt: now},
			{Message: replayMessage("presence_delta", incidentID, 2), StoredAt: now},
			{Message: replayMessage("job_progress", incidentID, 3), StoredAt: now},
			{Message: replayMessage("session_revoked", incidentID, 4), StoredAt: now},
			{Message: EphemeralMessage(incidentID, "presence_snapshot", nil, now), StoredAt: now},
		}
		hub.mu.Unlock()

		status, missed, highWater := hub.ReplayMessages(sessionID, incidentID, clientInstanceID, token, 0, now)
		if status != ResumeStatusReplayed {
			t.Fatalf("resume status = %q want %q", status, ResumeStatusReplayed)
		}
		if highWater != 4 {
			t.Fatalf("highWater = %d want 4", highWater)
		}
		gotTypes := messageTypes(missed)
		if want := []string{"record_changed", "job_progress"}; !reflect.DeepEqual(gotTypes, want) {
			t.Fatalf("replayed message types = %#v want %#v", gotTypes, want)
		}

		status, missed, _ = hub.ReplayMessages(sessionID, incidentID, clientInstanceID, token, 5, now)
		if status != ResumeStatusResetNeeded || len(missed) != 0 {
			t.Fatalf("future last_seen must reset without replay, got status=%q missed=%#v", status, missed)
		}
	})

	t.Run("resume reset covers expired mismatched and too old tokens", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		sessionID := uuid.New()
		clientInstanceID := "phase6-u-6-08-reset"
		now := time.Date(2026, 5, 5, 12, 30, 0, 0, time.UTC)

		expired, _, err := hub.IssueResumeToken(sessionID, incidentID, clientInstanceID, now.Add(time.Hour), now)
		if err != nil {
			t.Fatalf("issue expired resume token: %v", err)
		}
		valid, _, err := hub.IssueResumeToken(sessionID, incidentID, clientInstanceID, now.Add(time.Hour), now)
		if err != nil {
			t.Fatalf("issue valid resume token: %v", err)
		}

		hub.mu.Lock()
		record := hub.resumeTokens[expired]
		record.ExpiresAt = now.Add(-time.Second)
		hub.resumeTokens[expired] = record
		hub.highWater[incidentID] = 11
		hub.replay[incidentID] = []replayEntry{
			{Message: replayMessage("record_changed", incidentID, 10), StoredAt: now},
			{Message: replayMessage("job_progress", incidentID, 11), StoredAt: now},
		}
		hub.mu.Unlock()

		cases := []struct {
			name             string
			token            string
			incidentID       uuid.UUID
			clientInstanceID string
			lastSeen         int64
		}{
			{name: "expired", token: expired, incidentID: incidentID, clientInstanceID: clientInstanceID, lastSeen: 9},
			{name: "unknown", token: "unknown-token", incidentID: incidentID, clientInstanceID: clientInstanceID, lastSeen: 9},
			{name: "mismatched incident", token: valid, incidentID: uuid.New(), clientInstanceID: clientInstanceID, lastSeen: 9},
			{name: "mismatched client", token: valid, incidentID: incidentID, clientInstanceID: "other-client", lastSeen: 9},
			{name: "too old", token: valid, incidentID: incidentID, clientInstanceID: clientInstanceID, lastSeen: 8},
		}
		for _, tc := range cases {
			status, missed, _ := hub.ReplayMessages(sessionID, tc.incidentID, tc.clientInstanceID, tc.token, tc.lastSeen, now)
			if status != ResumeStatusResetNeeded || len(missed) != 0 {
				t.Fatalf("%s resume must reset without replay, got status=%q missed=%#v", tc.name, status, missed)
			}
		}
	})

	t.Run("presence snapshots are incident scoped sorted and expire", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		otherIncidentID := uuid.New()
		userID := uuid.New()
		now := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)
		firstConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		secondConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

		hub.UpsertPresence(incidentID, firstConnectionID, userID, "Analyst", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v1"},
			Mode:     "viewing",
		}, now)
		hub.UpsertPresence(incidentID, secondConnectionID, userID, "Analyst", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v1"},
			Mode:     "editing",
			RecordID: stringPointer(uuid.NewString()),
			FieldKey: stringPointer("timeline.summary"),
		}, now)
		hub.UpsertPresence(otherIncidentID, uuid.New(), userID, "Other", PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v1"},
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
		requirePhase6Reason(t, sessionRevocations, "concurrency_limit")

		hub.RevokeIncidentSession(incidentID, sessionID, "incident_access_revoked")
		requirePhase6Reason(t, incidentRevocations, "incident_access_revoked")
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
				"timeline.summary":       map[string]any{"value": "Patched"},
				"timeline.capture_state": map[string]any{"value": "enriched"},
				"timeline.details":       map[string]any{"value": "Omitted"},
			},
			"group_values": map[string]any{
				"timeline.capture_state": "enriched",
				"timeline.details":       "not-a-group-key",
			},
		}
		patch := BuildViewRowPatch(row, []string{
			"timeline.summary",
			"timeline.capture_state",
		})
		payload := RecordChangePayload(RecordChange{
			IncidentID:       incidentID,
			RecordID:         recordID,
			RowVersion:       7,
			ChangeSetID:      changeSetID,
			ClientTxnID:      "txn-phase6-patch",
			ActorUserID:      actorUserID,
			ChangedFieldKeys: []string{"timeline.summary", "timeline.capture_state"},
			ViewSchemaID:     "cartulary.view.timeline.v1",
			PatchCells:       patch,
		})

		changedKeys, _ := payload["changed_field_keys"].([]string)
		if want := []string{"timeline.capture_state", "timeline.summary"}; !reflect.DeepEqual(changedKeys, want) {
			t.Fatalf("changed_field_keys = %#v want %#v", changedKeys, want)
		}
		affectedViews, _ := payload["affected_views"].([]map[string]any)
		if len(affectedViews) != 1 || affectedViews[0]["change_kind"] != "patch" {
			t.Fatalf("expected one patch affected view, got %#v", payload["affected_views"])
		}
		patchCells, _ := affectedViews[0]["patch_cells"].(map[string]any)
		cells, _ := patchCells["cells"].(map[string]any)
		if _, ok := cells["timeline.details"]; ok {
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
			ClientTxnID:      "txn-phase6-invalidate",
			ActorUserID:      actorUserID,
			ChangedFieldKeys: []string{"timeline.summary"},
			ViewSchemaID:     "cartulary.view.timeline.v1",
		})
		fallbackViews, _ := fallback["affected_views"].([]map[string]any)
		if len(fallbackViews) != 1 || fallbackViews[0]["change_kind"] != "invalidate" {
			t.Fatalf("expected invalidate fallback, got %#v", fallback["affected_views"])
		}
	})
}

func replayMessage(messageType string, incidentID uuid.UUID, streamSeq int64) Message {
	return Message{
		Type:       messageType,
		IncidentID: incidentID.String(),
		StreamSeq:  &streamSeq,
		Payload:    RawPayload(map[string]any{}),
	}
}

func messageTypes(messages []Message) []string {
	types := make([]string, 0, len(messages))
	for _, message := range messages {
		types = append(types, message.Type)
	}
	return types
}

func stringPointer(value string) *string {
	return &value
}

func requirePhase6Reason(t testing.TB, ch <-chan string, want string) {
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
