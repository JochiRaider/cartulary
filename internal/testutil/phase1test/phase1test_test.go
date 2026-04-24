package phase1test

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
	mux.HandleFunc("/ws/v1/test/session-lifecycle", func(w http.ResponseWriter, r *http.Request) {
		conn, err := platformws.Accept(w, r, "")
		if err != nil {
			return
		}
		defer conn.CloseNow()

		if err := platformws.WriteJSON(context.Background(), conn, platformws.Message{
			Type:    "connected",
			Payload: platformws.RawPayload(map[string]any{"session_id": "test-session"}),
		}); err != nil {
			return
		}

		var message platformws.Message
		if err := platformws.ReadJSON(context.Background(), conn, &message); err != nil {
			return
		}
		if message.Type != "trigger_session_revoked" {
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

	headers := http.Header{}
	headers.Set("Authorization", "Bearer session-token")
	rawClient := wstest.ConnectWithHeaders(t, server.URL, "/ws/v1/test/session-lifecycle", headers)
	client := newSessionSocketClient(t, rawClient)

	if err := rawClient.Send(context.Background(), platformws.Message{Type: "trigger_session_revoked"}); err != nil {
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
