package flowtest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestSessionSocketClientCapturesRevocationThenClose(t *testing.T) {
	mux := http.NewServeMux()
	incidentID := "10000000-0000-0000-0000-000000000001"
	mux.HandleFunc("/ws/v1/incidents/"+incidentID, func(w http.ResponseWriter, r *http.Request) {
		conn, err := platformws.Accept(w, r, "")
		if err != nil {
			return
		}
		defer conn.CloseNow()

		var hello platformws.Message
		if err := platformws.ReadJSON(context.Background(), conn, &hello); err != nil {
			return
		}
		if hello.Type != "hello" {
			_ = conn.Close(websocket.StatusPolicyViolation, "unexpected_message")
			return
		}
		now := time.Now().UTC()
		if err := platformws.WriteJSON(context.Background(), conn, platformws.Message{
			Type: "hello_ack",
			Payload: platformws.RawPayload(map[string]any{
				"connection_id":         "20000000-0000-0000-0000-000000000001",
				"resume_token":          "resume-token",
				"server_time":           now.Format(time.RFC3339Nano),
				"heartbeat_interval_ms": int(platformws.HeartbeatInterval / time.Millisecond),
				"presence_ttl_ms":       int(platformws.PresenceTTL / time.Millisecond),
				"resume_window_ms":      int(platformws.ResumeWindow / time.Millisecond),
			}),
		}); err != nil {
			return
		}
		if err := platformws.WriteJSON(context.Background(), conn, platformws.Message{
			Type:    "presence_snapshot",
			Payload: platformws.RawPayload(map[string]any{"presences": []any{}}),
		}); err != nil {
			return
		}

		var trigger platformws.Message
		if err := platformws.ReadJSON(context.Background(), conn, &trigger); err != nil {
			return
		}
		if trigger.Type != "trigger_session_revoked" {
			_ = conn.Close(websocket.StatusPolicyViolation, "unexpected_message")
			return
		}

		if err := platformws.WriteJSON(context.Background(), conn, platformws.Message{
			Type:    "session_revoked",
			Payload: platformws.RawPayload(map[string]any{"reason_code": "test_reason"}),
		}); err != nil {
			return
		}
		_ = conn.Close(websocket.StatusPolicyViolation, "session_revoked")
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := ConnectSessionSocket(t, server.URL, incidentID, "session-token")

	if err := client.Send(context.Background(), platformws.Message{Type: "trigger_session_revoked"}); err != nil {
		t.Fatalf("send trigger_session_revoked: %v", err)
	}

	message, err := client.AwaitNextMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("await session_revoked message: %v", err)
	}
	wstest.RequireSessionRevoked(t, message)

	closeErr := client.AwaitClose(5 * time.Second)
	if closeErr == nil {
		t.Fatal("expected websocket close after session_revoked")
	}
	wstest.RequireClose(t, closeErr, websocket.StatusPolicyViolation, "session_revoked")
}
