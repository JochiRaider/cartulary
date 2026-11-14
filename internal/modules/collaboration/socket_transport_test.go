package collaboration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestIncidentSocketFrameFailureContract_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	t.Run("all presence bearing client families reject invalid presence", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-unclaimed-presence")
		for _, invalidPresence := range []map[string]any{
			{"mode": "viewing", "sheet_ref": map[string]string{"kind": "extension_workspace", "extension_profile_id": "network_flow_activity", "workspace_key": "network_analysis"}},
			{"mode": "editing", "sheet_ref": map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"}, "field_key": "timeline.activity_synopsis_text"},
		} {
			for _, family := range []string{"hello", "resume", "presence_update"} {
				t.Run(family, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					var conn *wstest.Client
					if family == "presence_update" {
						conn = connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.Session.Value)
					} else {
						var err error
						conn, _, err = incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value})
						if err != nil {
							t.Fatal(err)
						}
					}
					defer conn.CloseNow()
					payload := map[string]any{"presence": invalidPresence}
					if family != "presence_update" {
						payload["client_instance_id"] = "unclaimed-workspace-client"
					}
					if family == "resume" {
						payload["resume_token"], payload["last_seen_stream_seq"] = "unknown-token", 0
					}
					if err := conn.WriteJSON(ctx, platformws.Message{Type: family, IncidentID: incidentID, EmittedAt: time.Now().UTC().Format(time.RFC3339Nano), Payload: collabtestprotocol.RawPayload(t, payload)}); err != nil {
						t.Fatal(err)
					}
					code, reason := "invalid_websocket_handshake", "invalid_first_message"
					if family == "presence_update" {
						code, reason = "invalid_websocket_message", "invalid_message"
					}
					requireErrorCode(t, readRawCollaborationMessage(t, ctx, conn), code)
					_, _, err := conn.Read(ctx)
					wstest.RequireClose(t, err, wstest.StatusPolicyViolation, reason)
				})
			}
		}
	})
	t.Run("additive client payloads retain canonical presence across all families", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-additive-payload")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		observer := connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.Session.Value)
		defer observer.CloseNow()
		connect := func() *wstest.Client {
			conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value})
			if err != nil {
				t.Fatal(err)
			}
			return conn
		}
		conn := connect()
		defer conn.CloseNow()
		presence := map[string]any{
			"mode": "viewing", "MODE": false, "future": []any{1, nil},
			"sheet_ref": map[string]any{"kind": "view_schema", "id": "cartulary.view.timeline.v2", "KIND": false, "future": map[string]any{"nested": true}},
		}
		payload := map[string]any{"client_instance_id": "additive-client", "CLIENT_INSTANCE_ID": false, "presence": presence, "future": true}
		write := func(target *wstest.Client, family string, value map[string]any) {
			raw, err := json.Marshal(platformws.Message{Type: family, IncidentID: incidentID, EmittedAt: time.Now().UTC().Format(time.RFC3339Nano), Payload: collabtestprotocol.RawPayload(t, value)})
			if err != nil {
				t.Fatal(err)
			}
			if err := target.Write(ctx, wstest.MessageText, raw); err != nil {
				t.Fatal(err)
			}
		}
		write(conn, "hello", payload)
		ack := readRawCollaborationMessage(t, ctx, conn)
		if ack.Type != "hello_ack" {
			t.Fatalf("hello response = %s", ack.Type)
		}
		var accepted struct {
			ResumeToken string `json:"resume_token"`
		}
		if err := json.Unmarshal(ack.Payload, &accepted); err != nil || accepted.ResumeToken == "" {
			t.Fatalf("hello token unavailable: %v", err)
		}
		if snapshot := readRawCollaborationMessage(t, ctx, conn); snapshot.Type != "presence_snapshot" {
			t.Fatalf("hello snapshot = %s", snapshot.Type)
		}
		assertDelta := func(wantMode string) {
			message := readRawCollaborationMessage(t, ctx, observer)
			var delta struct {
				Kind     string                     `json:"delta_kind"`
				Presence map[string]json.RawMessage `json:"presence"`
			}
			if err := json.Unmarshal(message.Payload, &delta); err != nil || message.Type != "presence_delta" || delta.Kind != "upsert" {
				t.Fatalf("presence observation = %#v / %v", message, err)
			}
			if string(delta.Presence["mode"]) != fmt.Sprintf("%q", wantMode) {
				t.Fatalf("presence mode = %s", delta.Presence["mode"])
			}
			if _, exists := delta.Presence["future"]; exists {
				t.Fatal("unknown presence member retained")
			}
			var sheet map[string]any
			if err := json.Unmarshal(delta.Presence["sheet_ref"], &sheet); err != nil || len(sheet) != 2 || sheet["kind"] != "view_schema" || sheet["id"] != "cartulary.view.timeline.v2" {
				t.Fatalf("canonical sheet = %#v / %v", sheet, err)
			}
		}
		assertDelta("viewing")
		presence["mode"] = "idle"
		write(conn, "presence_update", payload)
		assertDelta("idle")
		write(conn, "pong", map[string]any{"future": map[string]any{"nested": true}})
		assertDelta("idle")
		conn.CloseNow()
		resumed := connect()
		defer resumed.CloseNow()
		payload["resume_token"], payload["last_seen_stream_seq"] = accepted.ResumeToken, 0
		write(resumed, "resume", payload)
		ack = readRawCollaborationMessage(t, ctx, resumed)
		var result struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(ack.Payload, &result); err != nil || ack.Type != "resume_ack" || result.Status != "replayed" {
			t.Fatalf("resume = %#v / %v", ack, err)
		}
		if snapshot := readRawCollaborationMessage(t, ctx, resumed); snapshot.Type != "presence_snapshot" {
			t.Fatalf("resume snapshot = %s", snapshot.Type)
		}
	})
	t.Run("resume requires an explicit valid sequence", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-resume-sequence")
		for _, sequence := range []any{nil, -1, 1.5, "0", false, "missing"} {
			t.Run(fmt.Sprintf("%v", sequence), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value})
				if err != nil {
					t.Fatal(err)
				}
				defer conn.CloseNow()
				payload := map[string]any{"client_instance_id": "client", "resume_token": "unknown-token", "presence": timelinePresence(), "last_seen_stream_seq": sequence}
				if sequence == "missing" {
					delete(payload, "last_seen_stream_seq")
				}
				raw, err := json.Marshal(platformws.Message{Type: "resume", IncidentID: incidentID, EmittedAt: time.Now().UTC().Format(time.RFC3339Nano), Payload: collabtestprotocol.RawPayload(t, payload)})
				if err != nil {
					t.Fatal(err)
				}
				if err := conn.Write(ctx, wstest.MessageText, raw); err != nil {
					t.Fatal(err)
				}
				requireErrorCode(t, readRawCollaborationMessage(t, ctx, conn), "invalid_websocket_handshake")
				_, _, err = conn.Read(ctx)
				wstest.RequireClose(t, err, wstest.StatusPolicyViolation, "invalid_first_message")
			})
		}
	})
	t.Run("maximum sized additive hello is accepted", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-frame-maximum")
		conn, _, err := incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
			SessionToken: admin.Session.Value,
		})
		if err != nil {
			t.Fatalf("dial incident socket: %v", err)
		}
		defer conn.CloseNow()
		hello, err := json.Marshal(platformws.Message{
			Type: "hello", IncidentID: incidentID,
			EmittedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Payload: collabtestprotocol.RawPayload(t, map[string]any{
				"client_instance_id": "maximum-frame-client",
				"presence":           timelinePresence(),
			}),
		})
		if err != nil {
			t.Fatal(err)
		}
		const suffix = `"}`
		prefix := append(hello[:len(hello)-1], []byte(`,"future":"`)...)
		payload := append(prefix, bytes.Repeat([]byte("x"), 32768-len(prefix)-len(suffix))...)
		payload = append(payload, suffix...)
		if len(payload) != 32768 {
			t.Fatalf("boundary fixture length = %d", len(payload))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.Write(ctx, wstest.MessageText, payload); err != nil {
			t.Fatalf("write maximum frame: %v", err)
		}
		if message := readRawCollaborationMessage(t, ctx, conn); message.Type != "hello_ack" {
			t.Fatalf("maximum hello response = %q", message.Type)
		}
		if message := readRawCollaborationMessage(t, ctx, conn); message.Type != "presence_snapshot" {
			t.Fatalf("maximum hello snapshot = %q", message.Type)
		}
		// One whitespace byte keeps valid JSON while exceeding the wire byte bound.
		if err := conn.Write(ctx, wstest.MessageText, append(payload, ' ')); err != nil {
			t.Fatalf("write over-limit frame: %v", err)
		}
		_, _, err = conn.Read(ctx)
		wstest.RequireClose(t, err, wstest.StatusMessageTooBig, "message_too_large")
	})

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
			payload:     bytes.Repeat([]byte("x"), 32769),
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
				SessionToken: admin.Session.Value,
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
			conn := connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.Session.Value)
			defer conn.CloseNow()

			payload := append(append([]byte(nil), testCase.payload[:len(testCase.payload)-1]...), []byte(fmt.Sprintf(",\"incident_id\":%q,\"emitted_at\":%q}", incidentID, time.Now().UTC().Format(time.RFC3339Nano)))...)
			if err := conn.Write(ctx, wstest.MessageText, payload); err != nil {
				t.Fatalf("write invalid later message: %v", err)
			}
			message := readRawCollaborationMessage(t, ctx, conn)
			requireErrorCode(t, message, "invalid_websocket_message")
			_, _, err := conn.Read(ctx)
			wstest.RequireClose(t, err, wstest.StatusPolicyViolation, "invalid_message")
		})
	}
	t.Run("client envelopes reject invalid authority and metadata", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-envelope")
		for _, later := range []bool{false, true} {
			for _, tc := range []struct {
				name   string
				mutate func(map[string]any)
			}{
				{"numeric type", func(m map[string]any) { m["type"] = 42 }},
				{"numeric incident", func(m map[string]any) { m["incident_id"] = 42 }},
				{"boolean timestamp", func(m map[string]any) { m["emitted_at"] = false }},
				{"object client sequence", func(m map[string]any) { m["stream_seq"] = map[string]any{} }},
				{"missing incident", func(m map[string]any) { delete(m, "incident_id") }},
				{"wrong case incident", func(m map[string]any) { m["INCIDENT_ID"] = m["incident_id"]; delete(m, "incident_id") }},
				{"empty incident", func(m map[string]any) { m["incident_id"] = "" }},
				{"different incident", func(m map[string]any) { m["incident_id"] = "10000000-0000-4000-8000-000000000001" }},
				{"missing timestamp", func(m map[string]any) { delete(m, "emitted_at") }},
				{"empty timestamp", func(m map[string]any) { m["emitted_at"] = "" }},
				{"malformed timestamp", func(m map[string]any) { m["emitted_at"] = "yesterday" }},
				{"comma timestamp fraction", func(m map[string]any) { m["emitted_at"] = "2026-07-13T12:00:00,123Z" }},
				{"one digit timestamp hour", func(m map[string]any) { m["emitted_at"] = "2026-07-13T1:00:00Z" }},
				{"non UTC timestamp", func(m map[string]any) { m["emitted_at"] = "2026-07-13T12:00:00+01:00" }},
				{"client sequence", func(m map[string]any) { m["stream_seq"] = 1 }},
				{"null client sequence", func(m map[string]any) { m["stream_seq"] = nil }},
				{"missing payload", func(m map[string]any) { delete(m, "payload") }},
				{"null payload", func(m map[string]any) { m["payload"] = nil }},
				{"array payload", func(m map[string]any) { m["payload"] = []any{} }},
			} {
				t.Run(fmt.Sprintf("later=%t/%s", later, tc.name), func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					var conn *wstest.Client
					message := map[string]any{
						"type": "hello", "incident_id": incidentID, "emitted_at": "2026-07-13T12:00:00Z",
						"payload": map[string]any{"client_instance_id": "envelope-client", "presence": timelinePresence()},
						"future":  map[string]any{"ignored": true},
					}
					code, reason := "invalid_websocket_handshake", "invalid_first_message"
					if later {
						conn = connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.Session.Value)
						message["type"], message["payload"] = "pong", map[string]any{}
						code, reason = "invalid_websocket_message", "invalid_message"
					} else {
						var err error
						conn, _, err = incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value})
						if err != nil {
							t.Fatal(err)
						}
					}
					defer conn.CloseNow()
					tc.mutate(message)
					if err := conn.WriteJSON(ctx, message); err != nil {
						t.Fatal(err)
					}
					response := readRawCollaborationMessage(t, ctx, conn)
					if response.Type != "error" {
						t.Fatalf("invalid envelope response = %q", response.Type)
					}
					requireErrorCode(t, response, code)
					_, _, err := conn.Read(ctx)
					wstest.RequireClose(t, err, wstest.StatusPolicyViolation, reason)
				})
			}
		}
	})

	t.Run("valid non-object JSON is a semantic rejection", func(t *testing.T) {
		harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-json-shape")
		for _, later := range []bool{false, true} {
			for _, raw := range []string{`null`, `[]`, `42`, `true`, `"text"`} {
				t.Run(fmt.Sprintf("later=%t/%s", later, raw), func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					var conn *wstest.Client
					code, reason := "invalid_websocket_handshake", "invalid_first_message"
					if later {
						conn = connectRawHello(t, ctx, harness.Server.HTTP.URL, incidentID, admin.Session.Value)
						code, reason = "invalid_websocket_message", "invalid_message"
					} else {
						var err error
						conn, _, err = incidentwstest.TryConnect(harness.Server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value})
						if err != nil {
							t.Fatal(err)
						}
					}
					defer conn.CloseNow()
					if err := conn.Write(ctx, wstest.MessageText, []byte(raw)); err != nil {
						t.Fatal(err)
					}
					message := readRawCollaborationMessage(t, ctx, conn)
					if message.Type != "error" {
						t.Fatalf("shape rejection response = %q", message.Type)
					}
					requireErrorCode(t, message, code)
					_, _, err := conn.Read(ctx)
					wstest.RequireClose(t, err, wstest.StatusPolicyViolation, reason)
				})
			}
		}
	})

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
		Type: "hello", IncidentID: incidentID, EmittedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Payload: collabtestprotocol.RawPayload(t, map[string]any{
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
