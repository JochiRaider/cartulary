package collaboration

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func TestPresenceReplayRevocationTransport(t *testing.T) {
	t.Run("renewal preserves scopes and removal and expiry are final", func(t *testing.T) {
		now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		h := newHub("0.0.0+unknown")
		incidentID, connectionID, userID := uuid.New(), uuid.New(), uuid.New()
		input := collabprotocol.PresenceInput{SheetRef: map[string]string{"kind": "saved_view", "id": "saved-one"}, Mode: "editing", RecordID: stringPointer("row"), FieldKey: stringPointer("summary")}
		initial := h.UpsertPresence(incidentID, connectionID, userID, "Analyst", input, now)
		renewed, ok := h.RenewPresence(incidentID, connectionID, now.Add(15*time.Second))
		if !ok {
			t.Fatal("valid connection not renewed")
		}
		renewed.ExpiresAt = initial.ExpiresAt
		if !reflect.DeepEqual(initial, renewed) {
			t.Fatalf("renewal changed activity: %#v", renewed)
		}
		h.RemovePresence(incidentID, connectionID, now.Add(16*time.Second))
		if _, ok := h.RenewPresence(incidentID, connectionID, now.Add(17*time.Second)); ok {
			t.Fatal("removed presence restored")
		}
		h.UpsertPresence(incidentID, connectionID, userID, "Analyst", input, now)
		if _, ok := h.RenewPresence(incidentID, connectionID, now.Add(collabprotocol.PresenceTTL)); ok {
			t.Fatal("expired presence restored")
		}
	})
	t.Run("outgoing broadcasts cannot starve heartbeat requests", func(t *testing.T) {
		now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		if heartbeatPingDue(now.Add(14*time.Second), now, now) {
			t.Fatal("early heartbeat")
		}
		if !heartbeatPingDue(now.Add(15*time.Second), now, now) {
			t.Fatal("missing heartbeat at boundary")
		}
		if heartbeatPingDue(now.Add(16*time.Second), now, now.Add(15*time.Second)) {
			t.Fatal("repeated ping before interval")
		}
		if heartbeatPingDue(now.Add(16*time.Second), now.Add(15*time.Second), now) {
			t.Fatal("inbound frame ignored")
		}
		if !heartbeatPingDue(now.Add(30*time.Second), now.Add(15*time.Second), now.Add(15*time.Second)) {
			t.Fatal("renewal traffic suppressed next heartbeat")
		}
	})
	t.Run("pong renews expiry without changing the published observation", func(t *testing.T) {
		now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		incidentID, connectionID := uuid.New(), uuid.New()
		h := newHub("0.0.0+unknown")
		initial := h.UpsertPresence(incidentID, connectionID, uuid.New(), "Quiet Analyst", collabprotocol.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"}, Mode: "viewing",
		}, now)
		service := &routeService{hub: h, now: func() time.Time { return now.Add(15 * time.Second) }}
		if !service.handleClientMessage(context.Background(), nil, incidentID, connectionID, nil, httpauth.Principal{}, collabprotocol.Message{
			Type: "pong", Payload: collabtestprotocol.RawPayload(t, map[string]any{}),
		}) {
			t.Fatal("valid pong rejected")
		}
		snapshot := h.PresenceSnapshot(incidentID, now.Add(collabprotocol.PresenceTTL))
		if len(snapshot) != 1 {
			t.Fatalf("quiet connected presence expired: %#v", snapshot)
		}
		if snapshot[0].ObservedAt != initial.ObservedAt {
			t.Fatal("heartbeat changed activity observation")
		}
		if snapshot[0].ExpiresAt != now.Add(15*time.Second+collabprotocol.PresenceTTL).Format(time.RFC3339Nano) {
			t.Fatalf("heartbeat did not transmit renewed expiry: %#v", snapshot[0])
		}
	})
	t.Run("presence snapshots are incident scoped sorted and expire", func(t *testing.T) {
		hub := newHub("0.0.0+unknown")
		incidentID := uuid.New()
		otherIncidentID := uuid.New()
		userID := uuid.New()
		now := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)
		firstConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		secondConnectionID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

		hub.UpsertPresence(incidentID, firstConnectionID, userID, "Analyst", collabprotocol.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "viewing",
		}, now)
		hub.UpsertPresence(incidentID, secondConnectionID, userID, "Analyst", collabprotocol.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "editing",
			RecordID: stringPointer(uuid.NewString()),
			FieldKey: stringPointer("timeline.activity_synopsis_text"),
		}, now)
		hub.UpsertPresence(otherIncidentID, uuid.New(), userID, "Other", collabprotocol.PresenceInput{
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

		expired := hub.PresenceSnapshot(incidentID, now.Add(collabprotocol.PresenceTTL+time.Nanosecond))
		if len(expired) != 0 {
			t.Fatalf("expired presence must be pruned, got %#v", expired)
		}
	})

	t.Run("revocation subscribers preserve public reason codes and incident user isolation", func(t *testing.T) {
		hub := newHub("0.0.0+unknown")
		sessionID := uuid.New()
		incidentID := uuid.New()
		userID := uuid.New()
		otherUserID := uuid.New()
		otherIncidentID := uuid.New()

		sessionRevocations, unregisterSession := hub.RegisterSession(sessionID)
		defer unregisterSession()
		incidentRevocations, unregisterIncident := hub.RegisterIncidentUser(incidentID, userID)
		defer unregisterIncident()
		otherUserRevocations, unregisterOtherUser := hub.RegisterIncidentUser(incidentID, otherUserID)
		defer unregisterOtherUser()
		otherIncidentRevocations, unregisterOtherIncident := hub.RegisterIncidentUser(otherIncidentID, userID)
		defer unregisterOtherIncident()

		hub.RevokeSession(sessionID, "concurrency_limit")
		requireReason(t, sessionRevocations, "concurrency_limit")

		hub.RevokeIncidentUser(incidentID, userID, "incident_access_revoked")
		requireReason(t, incidentRevocations, "incident_access_revoked")
		requireNoRevocationReason(t, otherUserRevocations)
		requireNoRevocationReason(t, otherIncidentRevocations)
	})

	t.Run("subscription teardown is idempotent and safe during publish", func(t *testing.T) {
		const iterations = 256
		for iteration := 0; iteration < iterations; iteration++ {
			hub := newHub("0.0.0+unknown")
			incidentID := uuid.New()
			messages, unsubscribeIncident := hub.SubscribeIncident(incidentID, 1)
			start := make(chan struct{})
			var workers sync.WaitGroup
			workers.Add(2)

			go func() {
				defer workers.Done()
				<-start
				for publish := 0; publish < 8; publish++ {
					if err := hub.DeliverReplayable(sequencedJobMessage(t, incidentID, int64(publish*2+1))); err != nil {
						t.Errorf("deliver job progress: %v", err)
						return
					}
					if err := hub.DeliverReplayable(sequencedRecordMessage(t, incidentID, int64(publish*2+2))); err != nil {
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
			if err := hub.DeliverReplayable(sequencedJobMessage(t, incidentID, 100)); err != nil {
				t.Fatalf("deliver after unsubscribe: %v", err)
			}
			if err := hub.DeliverReplayable(sequencedRecordMessage(t, incidentID, 101)); err != nil {
				t.Fatalf("deliver record after unsubscribe: %v", err)
			}
			requireNoIncidentMessage(t, messages)
		}
	})

	t.Run("sparse patch preserves changed cells and grouping scalars", func(t *testing.T) {
		recordID := uuid.New()
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
		patch := collabprotocol.BuildViewRowPatch(row, []string{
			"timeline.activity_synopsis_text",
			"timeline.capture_state",
		})

		cells, _ := patch["cells"].(map[string]any)
		if _, ok := cells["timeline.raw_activity_text"]; ok {
			t.Fatalf("patch_cells must omit unchanged cells, got %#v", cells)
		}
		if len(cells) != 2 {
			t.Fatalf("patch_cells cells length = %d want 2: %#v", len(cells), cells)
		}
		groupValues, _ := patch["group_values"].(map[string]any)
		if !reflect.DeepEqual(groupValues, map[string]any{"timeline.capture_state": "enriched"}) {
			t.Fatalf("patch group_values = %#v", groupValues)
		}

	})
}

func sequencedJobMessage(t testing.TB, incidentID uuid.UUID, streamSeq int64) collabprotocol.Message {
	now := time.Now().UTC()
	return collabprotocol.Message{
		Type:       "job_progress",
		IncidentID: incidentID.String(),
		EventID:    uuid.NewString(),
		EmittedAt:  now.Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload: collabtestprotocol.RawPayload(t, collabtestprotocol.NewIncidentJobProgressPayload(
			"job-subscription-teardown",
			incidentID,
			collabprotocol.JobStatusRunning,
			collabprotocol.JobProgress{},
			now,
		)),
	}
}

func sequencedRecordMessage(t testing.TB, incidentID uuid.UUID, streamSeq int64) collabprotocol.Message {
	now := time.Now().UTC()
	return collabprotocol.Message{
		Type:       "record_changed",
		IncidentID: incidentID.String(),
		EventID:    uuid.NewString(),
		EmittedAt:  now.Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload: collabtestprotocol.RawPayload(t, map[string]any{
			"record_id": uuid.NewString(), "row_version": int64(1),
			"change_set_id": uuid.NewString(), "client_txn_id": "txn-subscription-teardown",
			"actor_user_id": uuid.NewString(), "changed_field_keys": []string{},
			"affected_views": []map[string]any{{"view_schema_id": "cartulary.view.timeline.v2", "change_kind": "invalidate"}},
		}),
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
