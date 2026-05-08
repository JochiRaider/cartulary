package collaboration_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase3test"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestPhase6_IncidentSocketHandshakeResume_U_6_07(t *testing.T) {
	runtime := phase3test.StartRuntime(t)

	t.Run("first application message must be hello or resume", func(t *testing.T) {
		harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-u-6-07-first-message")

		conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken: admin.SessionCookie.Value,
		})
		if err != nil {
			t.Fatalf("dial incident socket: %v", err)
		}
		defer conn.CloseNow()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := platformws.WriteJSON(ctx, conn, platformws.Message{
			Type:    "presence_update",
			Payload: platformws.RawPayload(map[string]any{"presence": timelinePresence()}),
		}); err != nil {
			t.Fatalf("send invalid first message: %v", err)
		}

		var message platformws.Message
		if err := platformws.ReadJSON(ctx, conn, &message); err != nil {
			t.Fatalf("read invalid handshake error: %v", err)
		}
		if message.Type != "error" {
			t.Fatalf("invalid first message response type = %q want error", message.Type)
		}
		requireErrorCode(t, message, "invalid_websocket_handshake")

		var closed platformws.Message
		err = platformws.ReadJSON(ctx, conn, &closed)
		wstest.RequireClose(t, err, websocket.StatusPolicyViolation, "invalid_first_message")
	})

	t.Run("later hello or resume closes with route-owned invalid message behavior", func(t *testing.T) {
		harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-u-6-07-later-handshake")

		client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-later",
			Presence:         timelinePresence(),
		})
		defer client.Close(websocket.StatusNormalClosure, "test_complete")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Send(ctx, platformws.Message{
			Type: "hello",
			Payload: platformws.RawPayload(map[string]any{
				"client_instance_id": "phase6-u-6-07-later",
				"presence":           timelinePresence(),
			}),
		}); err != nil {
			t.Fatalf("send repeated hello: %v", err)
		}

		message, err := client.AwaitNextMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("receive invalid message error: %v", err)
		}
		if message.Type != "error" {
			t.Fatalf("repeated hello response type = %q want error", message.Type)
		}
		requireErrorCode(t, message, "invalid_websocket_message")

		closeErr := client.AwaitClose(5 * time.Second)
		wstest.RequireClose(t, closeErr, websocket.StatusPolicyViolation, "invalid_message")
	})

	t.Run("invalid stale or mismatched resume resets without partial replay", func(t *testing.T) {
		harness, admin, incidentID := setupPhase6SocketIncident(t, runtime, "phase6-u-6-07-resume-reset")

		initial := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-resume-source",
			Presence:         timelinePresence(),
		})
		token := initial.HelloAck.ResumeToken
		initial.Close(websocket.StatusNormalClosure, "test_complete")

		mismatched := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-other-client",
			Presence:         timelinePresence(),
		}, token, 0)
		defer mismatched.Close(websocket.StatusNormalClosure, "test_complete")
		if mismatched.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("mismatched client resume status = %q want %q", mismatched.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(mismatched.ReplayedMessages) != 0 {
			t.Fatalf("mismatched client resume replayed messages: %#v", mismatched.ReplayedMessages)
		}
		mismatched.Close(websocket.StatusNormalClosure, "test_complete")

		fresh := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-future-source",
			Presence:         timelinePresence(),
		})
		futureToken := fresh.HelloAck.ResumeToken
		fresh.Close(websocket.StatusNormalClosure, "test_complete")

		future := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-future-source",
			Presence:         timelinePresence(),
		}, futureToken, 999)
		defer future.Close(websocket.StatusNormalClosure, "test_complete")
		if future.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("future last_seen resume status = %q want %q", future.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(future.ReplayedMessages) != 0 {
			t.Fatalf("future last_seen resume replayed messages: %#v", future.ReplayedMessages)
		}
		future.Close(websocket.StatusNormalClosure, "test_complete")

		expiring := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-expired-source",
			Presence:         timelinePresence(),
		})
		expiredToken := expiring.HelloAck.ResumeToken
		expiring.Close(websocket.StatusNormalClosure, "test_complete")
		harness.Server.Clock.Advance(platformws.ResumeWindow + time.Second)

		expired := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "phase6-u-6-07-expired-source",
			Presence:         timelinePresence(),
		}, expiredToken, 0)
		defer expired.Close(websocket.StatusNormalClosure, "test_complete")
		if expired.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("expired resume status = %q want %q", expired.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(expired.ReplayedMessages) != 0 {
			t.Fatalf("expired resume replayed messages: %#v", expired.ReplayedMessages)
		}
	})
}

func TestPhase6_IncidentSocketHeartbeatIdleExpiry_U_6_08(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	harness, admin, adminID, incidentID := setupPhase6SocketIncidentWithAdminID(t, runtime, "phase6-u-6-08-heartbeat-idle")
	sessionID := sessionIDForCookie(t, harness, adminID.String())
	before := queryPhase6SessionTiming(t, harness, sessionID)

	client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-u-6-08-heartbeat",
		Presence:         timelinePresence(),
	})
	defer client.Close(websocket.StatusNormalClosure, "test_complete")

	harness.Server.Clock.Advance(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Send(ctx, platformws.Message{
		Type:    "pong",
		Payload: platformws.RawPayload(map[string]any{}),
	}); err != nil {
		t.Fatalf("send heartbeat pong: %v", err)
	}

	afterHeartbeat := queryPhase6SessionTiming(t, harness, sessionID)
	if !afterHeartbeat.LastQualifyingActivityAt.Equal(before.LastQualifyingActivityAt) {
		t.Fatalf("heartbeat must not slide last_qualifying_activity_at: before=%s after=%s", before.LastQualifyingActivityAt, afterHeartbeat.LastQualifyingActivityAt)
	}
	if !afterHeartbeat.IdleExpiresAt.Equal(before.IdleExpiresAt) {
		t.Fatalf("heartbeat must not slide idle_expires_at: before=%s after=%s", before.IdleExpiresAt, afterHeartbeat.IdleExpiresAt)
	}
	if !afterHeartbeat.SessionExpiresAt.Equal(before.SessionExpiresAt) {
		t.Fatalf("heartbeat must not slide session_expires_at: before=%s after=%s", before.SessionExpiresAt, afterHeartbeat.SessionExpiresAt)
	}

	harness.Server.Clock.Advance(26*time.Minute + time.Second)
	incidentwstest.ExpectSessionRevoked(t, client, "session_expired")
	afterExpiry := queryPhase6SessionTiming(t, harness, sessionID)
	if afterExpiry.RevokeReasonCode != "session_expired" {
		t.Fatalf("socket idle expiry revoke_reason_code = %q want session_expired", afterExpiry.RevokeReasonCode)
	}
}

func TestPhase6_IncidentSocketPresenceScopeEphemeral_U_6_08(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	harness, admin, _, incidentA := setupPhase6SocketIncidentWithAdminID(t, runtime, "phase6-u-6-08-presence-scope-a")
	incidentBResource := phase3test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-phase6-u-6-08-presence-scope-b",
		"incident_key":  "IR-PHASE6U608PRESENCEB",
		"title":         "phase6-u-6-08-presence-scope-b",
	})
	incidentB := incidentBResource["incident_id"].(string)

	firstA := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-u-6-08-presence-a-first",
		Presence:         timelinePresence(),
	})
	defer firstA.Close(websocket.StatusNormalClosure, "test_complete")
	if got := len(firstA.PresenceSnapshot); got != 1 {
		t.Fatalf("incident A initial presence_snapshot length = %d want 1: %#v", got, firstA.PresenceSnapshot)
	}
	firstAToken := firstA.HelloAck.ResumeToken

	firstB := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentB, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-u-6-08-presence-b-first",
		Presence:         timelinePresence(),
	})
	defer firstB.Close(websocket.StatusNormalClosure, "test_complete")
	if got := len(firstB.PresenceSnapshot); got != 1 {
		t.Fatalf("incident B initial presence_snapshot length = %d want 1: %#v", got, firstB.PresenceSnapshot)
	}

	secondA := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-u-6-08-presence-a-second",
		Presence:         timelinePresence(),
	})
	defer secondA.Close(websocket.StatusNormalClosure, "test_complete")
	if got := len(secondA.PresenceSnapshot); got != 2 {
		t.Fatalf("incident A second presence_snapshot length = %d want 2: %#v", got, secondA.PresenceSnapshot)
	}
	upsert := requirePresenceDelta(t, firstA, "upsert")
	if upsert.ConnectionID != secondA.HelloAck.ConnectionID {
		t.Fatalf("incident A upsert connection_id = %s want %s", upsert.ConnectionID, secondA.HelloAck.ConnectionID)
	}
	requireNoSocketMessage(t, firstB, 200*time.Millisecond, "incident B must not receive incident A presence_delta")

	secondA.Close(websocket.StatusNormalClosure, "test_complete")
	remove := requirePresenceDelta(t, firstA, "remove")
	if remove.ConnectionID != secondA.HelloAck.ConnectionID {
		t.Fatalf("incident A remove connection_id = %s want %s", remove.ConnectionID, secondA.HelloAck.ConnectionID)
	}

	firstA.Close(websocket.StatusNormalClosure, "test_complete")
	resumedA := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "phase6-u-6-08-presence-a-first",
		Presence:         timelinePresence(),
	}, firstAToken, 0)
	defer resumedA.Close(websocket.StatusNormalClosure, "test_complete")
	if len(resumedA.ReplayedMessages) != 0 {
		t.Fatalf("presence messages must remain ephemeral and absent from resume replay: %#v", resumedA.ReplayedMessages)
	}
	if got := len(resumedA.PresenceSnapshot); got != 1 {
		t.Fatalf("resumed incident A presence_snapshot length = %d want 1: %#v", got, resumedA.PresenceSnapshot)
	}
	if resumedA.PresenceSnapshot[0].ConnectionID == firstB.HelloAck.ConnectionID {
		t.Fatalf("resumed incident A snapshot leaked incident B presence: %#v", resumedA.PresenceSnapshot)
	}
}

func setupPhase6SocketIncident(t testing.TB, runtime *phase3test.RuntimeHarness, prefix string) (*phase3test.ServerHarness, phase3test.LoginResult, string) {
	t.Helper()

	harness := runtime.StartServer(t, prefix)
	admin, _ := phase3test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase3test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-" + prefix,
		"incident_key":  "IR-" + strings.ToUpper(strings.ReplaceAll(prefix, "-", "")),
		"title":         prefix,
	})
	return harness, admin, incident["incident_id"].(string)
}

func requireNoSocketMessage(t testing.TB, client *incidentwstest.Client, duration time.Duration, reason string) {
	t.Helper()

	message, err := client.AwaitNextMessage(duration)
	if errors.Is(err, context.DeadlineExceeded) {
		return
	}
	if err != nil {
		t.Fatalf("%s: unexpected socket error: %v", reason, err)
	}
	t.Fatalf("%s: unexpected message %#v", reason, message)
}

func timelinePresence() platformws.PresenceInput {
	return platformws.PresenceInput{
		SheetRef: map[string]string{
			"kind": "view_schema",
			"id":   timeline.TimelineViewSchemaID,
		},
		Mode: "viewing",
	}
}

func requireErrorCode(t testing.TB, message platformws.Message, want string) {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if payload["code"] != want {
		t.Fatalf("error code = %v want %q payload=%#v", payload["code"], want, payload)
	}
}

func requireReplayTypes(t testing.TB, messages []platformws.Message, want ...string) {
	t.Helper()
	if len(messages) != len(want) {
		t.Fatalf("replayed message count = %d want %d: %#v", len(messages), len(want), messages)
	}
	var previous int64
	for i, message := range messages {
		if message.Type != want[i] {
			t.Fatalf("replayed message %d type = %q want %q", i, message.Type, want[i])
		}
		if message.StreamSeq == nil {
			t.Fatalf("replayed message %d missing stream_seq: %#v", i, message)
		}
		if *message.StreamSeq <= previous {
			t.Fatalf("replayed stream_seq not strictly ascending: previous=%d current=%d", previous, *message.StreamSeq)
		}
		previous = *message.StreamSeq
	}
}

func requirePresenceDelta(t testing.TB, client *incidentwstest.Client, kind string) platformws.PresenceRecord {
	t.Helper()

	message, err := client.AwaitNextMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("receive presence_delta: %v", err)
	}
	if message.Type != "presence_delta" {
		t.Fatalf("message type = %q want presence_delta", message.Type)
	}
	var payload struct {
		DeltaKind string                    `json:"delta_kind"`
		Presence  platformws.PresenceRecord `json:"presence"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode presence_delta payload: %v", err)
	}
	if payload.DeltaKind != kind {
		t.Fatalf("presence_delta kind = %q want %q", payload.DeltaKind, kind)
	}
	return payload.Presence
}

func sessionIDForCookie(t testing.TB, harness *phase3test.ServerHarness, userID string) string {
	t.Helper()

	var sessionID string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT id::text
  FROM user_sessions
 WHERE user_id = $1::uuid
   AND revoked_at IS NULL
 ORDER BY authenticated_at DESC, created_at DESC, id DESC
 LIMIT 1
`, userID).Scan(&sessionID); err != nil {
		t.Fatalf("query active session id: %v", err)
	}
	return sessionID
}

type phase6SessionTiming struct {
	LastQualifyingActivityAt time.Time
	IdleExpiresAt            time.Time
	SessionExpiresAt         time.Time
	RevokeReasonCode         string
}

func queryPhase6SessionTiming(t testing.TB, harness *phase3test.ServerHarness, sessionID string) phase6SessionTiming {
	t.Helper()

	var timing phase6SessionTiming
	var revokeReasonCode sql.NullString
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT last_qualifying_activity_at,
       idle_expires_at,
       session_expires_at,
       revoke_reason_code
  FROM user_sessions
 WHERE id = $1::uuid
`, sessionID).Scan(
		&timing.LastQualifyingActivityAt,
		&timing.IdleExpiresAt,
		&timing.SessionExpiresAt,
		&revokeReasonCode,
	); err != nil {
		t.Fatalf("query session timing: %v", err)
	}
	if revokeReasonCode.Valid {
		timing.RevokeReasonCode = revokeReasonCode.String
	}
	return timing
}
