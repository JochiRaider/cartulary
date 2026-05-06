package collaboration_test

import (
	"context"
	"encoding/json"
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
