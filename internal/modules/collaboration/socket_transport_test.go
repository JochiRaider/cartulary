package collaboration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

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
