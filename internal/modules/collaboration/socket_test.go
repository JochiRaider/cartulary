package collaboration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestIncidentSocketHandshakeResume_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	t.Run("first application message rejects every closed token except hello or resume", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-u-6-07-first-message")

		for _, messageType := range []string{
			"hello_ack",
			"resume_ack",
			"presence_snapshot",
			"presence_delta",
			"presence_update",
			"record_changed",
			"job_progress",
			"ping",
			"pong",
			"error",
			"session_revoked",
		} {
			t.Run(messageType, func(t *testing.T) {
				conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
					SessionToken: admin.SessionCookie.Value,
				})
				if err != nil {
					t.Fatalf("dial incident socket: %v", err)
				}
				defer conn.CloseNow()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := conn.WriteJSON(ctx, platformws.Message{
					Type:    messageType,
					Payload: invalidFirstMessagePayload(messageType),
				}); err != nil {
					t.Fatalf("send invalid first message %q: %v", messageType, err)
				}

				var message platformws.Message
				if err := conn.ReadJSON(ctx, &message); err != nil {
					t.Fatalf("read invalid handshake error for %q: %v", messageType, err)
				}
				if message.Type != "error" {
					t.Fatalf("invalid first message %q response type = %q want error", messageType, message.Type)
				}
				requireErrorCode(t, message, "invalid_websocket_handshake")

				var closed platformws.Message
				err = conn.ReadJSON(ctx, &closed)
				wstest.RequireClose(t, err, wstest.StatusPolicyViolation, "invalid_first_message")
			})
		}
	})

	t.Run("later hello or resume closes with route-owned invalid message behavior", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-u-6-07-later-handshake")

		for _, tc := range []struct {
			name         string
			buildMessage func(*incidentwstest.Client) platformws.Message
		}{
			{
				name: "hello",
				buildMessage: func(*incidentwstest.Client) platformws.Message {
					return platformws.Message{
						Type: "hello",
						Payload: platformws.RawPayload(map[string]any{
							"client_instance_id": "collaboration-u-6-07-later",
							"presence":           timelinePresence(),
						}),
					}
				},
			},
			{
				name: "resume",
				buildMessage: func(client *incidentwstest.Client) platformws.Message {
					return platformws.Message{
						Type: "resume",
						Payload: platformws.RawPayload(map[string]any{
							"client_instance_id":   "collaboration-u-6-07-later",
							"resume_token":         client.HelloAck.ResumeToken,
							"last_seen_stream_seq": 0,
							"presence":             timelinePresence(),
						}),
					}
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
					SessionToken:     admin.SessionCookie.Value,
					ClientInstanceID: "collaboration-u-6-07-later-" + tc.name,
					Presence:         timelinePresence(),
				})
				defer client.Close(wstest.StatusNormalClosure, "test_complete")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := client.Send(ctx, tc.buildMessage(client)); err != nil {
					t.Fatalf("send repeated %s: %v", tc.name, err)
				}

				message, err := client.AwaitNextMessage(5 * time.Second)
				if err != nil {
					t.Fatalf("receive invalid message error: %v", err)
				}
				if message.Type != "error" {
					t.Fatalf("repeated %s response type = %q want error", tc.name, message.Type)
				}
				requireErrorCode(t, message, "invalid_websocket_message")

				closeErr := client.AwaitClose(5 * time.Second)
				wstest.RequireClose(t, closeErr, wstest.StatusPolicyViolation, "invalid_message")
			})
		}
	})

	t.Run("invalid stale or mismatched resume resets without partial replay", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-u-6-07-resume-reset")

		initial := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-resume-source",
			Presence:         timelinePresence(),
		})
		token := initial.HelloAck.ResumeToken
		initial.Close(wstest.StatusNormalClosure, "test_complete")

		mismatched := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-other-client",
			Presence:         timelinePresence(),
		}, token, 0)
		defer mismatched.Close(wstest.StatusNormalClosure, "test_complete")
		if mismatched.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("mismatched client resume status = %q want %q", mismatched.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(mismatched.ReplayedMessages) != 0 {
			t.Fatalf("mismatched client resume replayed messages: %#v", mismatched.ReplayedMessages)
		}
		mismatched.Close(wstest.StatusNormalClosure, "test_complete")

		fresh := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-future-source",
			Presence:         timelinePresence(),
		})
		futureToken := fresh.HelloAck.ResumeToken
		fresh.Close(wstest.StatusNormalClosure, "test_complete")

		future := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-future-source",
			Presence:         timelinePresence(),
		}, futureToken, 999)
		defer future.Close(wstest.StatusNormalClosure, "test_complete")
		if future.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("future last_seen resume status = %q want %q", future.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(future.ReplayedMessages) != 0 {
			t.Fatalf("future last_seen resume replayed messages: %#v", future.ReplayedMessages)
		}
		future.Close(wstest.StatusNormalClosure, "test_complete")

		expiring := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-expired-source",
			Presence:         timelinePresence(),
		})
		expiredToken := expiring.HelloAck.ResumeToken
		expiring.Close(wstest.StatusNormalClosure, "test_complete")
		harness.Server.Clock.Advance(platformws.ResumeWindow + time.Second)

		expired := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken:     admin.SessionCookie.Value,
			ClientInstanceID: "collaboration-u-6-07-expired-source",
			Presence:         timelinePresence(),
		}, expiredToken, 0)
		defer expired.Close(wstest.StatusNormalClosure, "test_complete")
		if expired.ResumeAck.Status != platformws.ResumeStatusResetNeeded {
			t.Fatalf("expired resume status = %q want %q", expired.ResumeAck.Status, platformws.ResumeStatusResetNeeded)
		}
		if len(expired.ReplayedMessages) != 0 {
			t.Fatalf("expired resume replayed messages: %#v", expired.ReplayedMessages)
		}
	})
}

func TestIncidentSocketFrameFailureContract_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	for _, testCase := range []struct {
		name        string
		messageKind wstest.MessageType
		payload     []byte
		wantStatus  wstest.StatusCode
		wantReason  string
		wantError   string
	}{
		{
			name:        "binary application frame",
			messageKind: wstest.MessageBinary,
			payload:     []byte(`{"type":"hello","payload":{}}`),
			wantStatus:  wstest.StatusUnsupportedData,
			wantReason:  "binary_message_unsupported",
		},
		{
			name:        "malformed JSON",
			messageKind: wstest.MessageText,
			payload:     []byte(`{"type":`),
			wantStatus:  wstest.StatusInvalidFramePayloadData,
			wantReason:  "invalid_json",
		},
		{
			name:        "oversized message",
			messageKind: wstest.MessageText,
			payload:     bytes.Repeat([]byte("x"), platformws.MaximumMessageBytes+1),
			wantStatus:  wstest.StatusMessageTooBig,
			wantReason:  "message_too_large",
		},
		{
			name:        "duplicate member in first message",
			messageKind: wstest.MessageText,
			payload:     []byte(`{"type":"hello","type":"hello","payload":{}}`),
			wantStatus:  wstest.StatusPolicyViolation,
			wantReason:  "invalid_first_message",
			wantError:   "invalid_websocket_handshake",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-frame-"+strings.ReplaceAll(testCase.name, " ", "-"))
			conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
				SessionToken: admin.SessionCookie.Value,
			})
			if err != nil {
				t.Fatalf("dial incident socket: %v", err)
			}
			defer conn.CloseNow()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := conn.Write(ctx, testCase.messageKind, testCase.payload); err != nil {
				t.Fatalf("write invalid frame: %v", err)
			}
			if testCase.wantError != "" {
				message := readRawCollaborationMessage(t, ctx, conn)
				requireErrorCode(t, message, testCase.wantError)
			}
			_, _, err = conn.Read(ctx)
			wstest.RequireClose(t, err, testCase.wantStatus, testCase.wantReason)
		})
	}

	for _, testCase := range []struct {
		name    string
		payload []byte
	}{
		{
			name:    "unknown later type",
			payload: []byte(`{"type":"future_message","payload":{}}`),
		},
		{
			name:    "invalid later payload",
			payload: []byte(`{"type":"presence_update","payload":[]}`),
		},
		{
			name:    "duplicate later payload member",
			payload: []byte(`{"type":"presence_update","payload":{"presence":{},"presence":{}}}`),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-later-"+strings.ReplaceAll(testCase.name, " ", "-"))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn := connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.SessionCookie.Value)
			defer conn.CloseNow()

			if err := conn.Write(ctx, wstest.MessageText, testCase.payload); err != nil {
				t.Fatalf("write invalid later message: %v", err)
			}
			message := readRawCollaborationMessage(t, ctx, conn)
			requireErrorCode(t, message, "invalid_websocket_message")
			_, _, err := conn.Read(ctx)
			wstest.RequireClose(t, err, wstest.StatusPolicyViolation, "invalid_message")
		})
	}
}

func connectRawHello(
	t testing.TB,
	ctx context.Context,
	serverURL string,
	incidentID string,
	sessionToken string,
) *wstest.Client {
	t.Helper()
	conn, _, err := incidentwstest.TryConnect(serverURL, incidentID, incidentwstest.ConnectOptions{
		SessionToken: sessionToken,
	})
	if err != nil {
		t.Fatalf("dial incident socket: %v", err)
	}
	hello, err := json.Marshal(platformws.Message{
		Type: "hello",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id": "strict-frame-client",
			"presence":           timelinePresence(),
		}),
	})
	if err != nil {
		t.Fatalf("marshal hello: %v", err)
	}
	if err := conn.Write(ctx, wstest.MessageText, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}
	helloAck := readRawCollaborationMessage(t, ctx, conn)
	if helloAck.Type != "hello_ack" {
		t.Fatalf("first response = %q want hello_ack", helloAck.Type)
	}
	snapshot := readRawCollaborationMessage(t, ctx, conn)
	if snapshot.Type != "presence_snapshot" {
		t.Fatalf("second response = %q want presence_snapshot", snapshot.Type)
	}
	return conn
}

func readRawCollaborationMessage(
	t testing.TB,
	ctx context.Context,
	conn *wstest.Client,
) platformws.Message {
	t.Helper()
	kind, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read Collaboration message: %v", err)
	}
	if kind != wstest.MessageText {
		t.Fatalf("outbound message kind = %v want text", kind)
	}
	if bytes.HasSuffix(payload, []byte("\n")) {
		t.Fatalf("outbound message has trailing LF: %q", payload)
	}
	var message platformws.Message
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode Collaboration message: %v", err)
	}
	return message
}

func invalidFirstMessagePayload(messageType string) json.RawMessage {
	switch messageType {
	case "presence_update":
		return platformws.RawPayload(map[string]any{"presence": timelinePresence()})
	case "hello_ack":
		return platformws.RawPayload(map[string]any{"resume_token": "unexpected"})
	case "resume_ack":
		return platformws.RawPayload(map[string]any{"status": platformws.ResumeStatusReplayed})
	case "presence_snapshot":
		return platformws.RawPayload(map[string]any{"presences": []any{}})
	case "presence_delta":
		return platformws.RawPayload(map[string]any{"delta_kind": "upsert", "presence": map[string]any{}})
	case "record_changed":
		return platformws.RawPayload(map[string]any{"record_id": "record-1"})
	case "job_progress":
		return platformws.RawPayload(map[string]any{"job_id": "job-1"})
	case "session_revoked":
		return platformws.RawPayload(map[string]any{"reason_code": "session_revoked"})
	default:
		return platformws.RawPayload(map[string]any{})
	}
}

func TestIncidentSocketHeartbeatIdleExpiry_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness, admin, adminID, incidentID := setupSocketIncidentWithAdminID(t, runtime, "collaboration-u-6-08-heartbeat-idle")
	sessionID := sessionIDForCookie(t, harness, adminID.String())
	before := querySessionTiming(t, harness, sessionID)

	client := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-u-6-08-heartbeat",
		Presence:         timelinePresence(),
	})
	defer client.Close(wstest.StatusNormalClosure, "test_complete")

	harness.Server.Clock.Advance(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Send(ctx, platformws.Message{
		Type:    "pong",
		Payload: platformws.RawPayload(map[string]any{}),
	}); err != nil {
		t.Fatalf("send heartbeat pong: %v", err)
	}

	afterHeartbeat := querySessionTiming(t, harness, sessionID)
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
	afterExpiry := querySessionTiming(t, harness, sessionID)
	if afterExpiry.RevokeReasonCode != "session_expired" {
		t.Fatalf("socket idle expiry revoke_reason_code = %q want session_expired", afterExpiry.RevokeReasonCode)
	}
}

func TestIncidentSocketPresenceScopeEphemeral_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness, admin, _, incidentA := setupSocketIncidentWithAdminID(t, runtime, "collaboration-u-6-08-presence-scope-a")
	incidentBResource := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-collaboration-u-6-08-presence-scope-b",
		"incident_key":  "IR-COLLABORATIONU608PRESENCEB",
		"title":         "collaboration-u-6-08-presence-scope-b",
	})
	incidentB := incidentBResource["incident_id"].(string)

	firstA := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-u-6-08-presence-a-first",
		Presence:         timelinePresence(),
	})
	defer firstA.Close(wstest.StatusNormalClosure, "test_complete")
	if got := len(firstA.PresenceSnapshot); got != 1 {
		t.Fatalf("incident A initial presence_snapshot length = %d want 1: %#v", got, firstA.PresenceSnapshot)
	}
	firstAToken := firstA.HelloAck.ResumeToken

	firstB := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentB, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-u-6-08-presence-b-first",
		Presence:         timelinePresence(),
	})
	defer firstB.Close(wstest.StatusNormalClosure, "test_complete")
	if got := len(firstB.PresenceSnapshot); got != 1 {
		t.Fatalf("incident B initial presence_snapshot length = %d want 1: %#v", got, firstB.PresenceSnapshot)
	}

	secondA := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-u-6-08-presence-a-second",
		Presence:         timelinePresence(),
	})
	defer secondA.Close(wstest.StatusNormalClosure, "test_complete")
	if got := len(secondA.PresenceSnapshot); got != 2 {
		t.Fatalf("incident A second presence_snapshot length = %d want 2: %#v", got, secondA.PresenceSnapshot)
	}
	upsert := requirePresenceDelta(t, firstA, "upsert")
	if upsert.ConnectionID != secondA.HelloAck.ConnectionID {
		t.Fatalf("incident A upsert connection_id = %s want %s", upsert.ConnectionID, secondA.HelloAck.ConnectionID)
	}
	requireNoSocketMessage(t, firstB, 200*time.Millisecond, "incident B must not receive incident A presence_delta")

	secondA.Close(wstest.StatusNormalClosure, "test_complete")
	remove := requirePresenceDelta(t, firstA, "remove")
	if remove.ConnectionID != secondA.HelloAck.ConnectionID {
		t.Fatalf("incident A remove connection_id = %s want %s", remove.ConnectionID, secondA.HelloAck.ConnectionID)
	}

	firstA.Close(wstest.StatusNormalClosure, "test_complete")
	resumedA := incidentwstest.ConnectAndResume(t, harness.Server.HTTP.URL, incidentA, incidentwstest.ConnectOptions{
		SessionToken:     admin.SessionCookie.Value,
		ClientInstanceID: "collaboration-u-6-08-presence-a-first",
		Presence:         timelinePresence(),
	}, firstAToken, 0)
	defer resumedA.Close(wstest.StatusNormalClosure, "test_complete")
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

func setupSocketIncident(t testing.TB, runtime *appsupport.Runtime, prefix string) (*appsupport.ServerHarness, flowtest.LoginResult, string) {
	t.Helper()

	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        prefix,
		Dependencies:  httpapi.DependencySet{},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	admin, _ := flowtest.ProvisionBootstrapAdminUUID(t, harness.Server.HTTP.URL)
	incident := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
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

func sessionIDForCookie(t testing.TB, harness *appsupport.ServerHarness, userID string) string {
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

type SessionTiming struct {
	LastQualifyingActivityAt time.Time
	IdleExpiresAt            time.Time
	SessionExpiresAt         time.Time
	RevokeReasonCode         string
}

func querySessionTiming(t testing.TB, harness *appsupport.ServerHarness, sessionID string) SessionTiming {
	t.Helper()

	var timing SessionTiming
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
