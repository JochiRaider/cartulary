package httptestx

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func RegisterBootstrapRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		_ = deps

		mux.HandleFunc("/api/v1/test/success", func(w http.ResponseWriter, r *http.Request) {
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
				"service": "bootstrap",
				"status":  "ok",
			})
		})

		mux.HandleFunc("/api/v1/test/error", func(w http.ResponseWriter, r *http.Request) {
			_ = httpapi.WriteError(w, r, http.StatusServiceUnavailable, "bootstrap_error", "bootstrap harness error", map[string]any{
				"reason_code": "bootstrap_unavailable",
				"scope":       "harness",
			})
		})

		mux.HandleFunc("/ws/v1/bootstrap-harness", func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer conn.CloseNow()

			ctx := context.Background()
			var first platformws.Message
			if err := wsjson.Read(ctx, conn, &first); err != nil {
				_ = conn.Close(websocket.StatusPolicyViolation, "invalid_first_message")
				return
			}
			if first.Type != "handshake" {
				_ = wsjson.Write(ctx, conn, platformws.Message{
					Type:    "protocol_error",
					Payload: collabtestprotocol.RawPayload(map[string]any{"reason_code": "first_message_must_be_handshake"}),
				})
				_ = conn.Close(websocket.StatusPolicyViolation, "invalid_first_message")
				return
			}

			if err := wsjson.Write(ctx, conn, platformws.Message{
				Type:    "handshake_ack",
				Payload: collabtestprotocol.RawPayload(map[string]any{"status": "connected", "boundary": "/ws/v1/bootstrap-harness"}),
			}); err != nil {
				return
			}

			for {
				var message platformws.Message
				if err := wsjson.Read(ctx, conn, &message); err != nil {
					return
				}

				switch message.Type {
				case "echo":
					if err := wsjson.Write(ctx, conn, message); err != nil {
						return
					}
				case "trigger_session_revoked":
					_ = wsjson.Write(ctx, conn, platformws.Message{
						Type:    "session_revoked",
						Payload: collabtestprotocol.RawPayload(map[string]any{"reason_code": "bootstrap_revoked"}),
					})
					_ = conn.Close(websocket.StatusPolicyViolation, "session_revoked")
					return
				case "close_me":
					_ = conn.Close(websocket.StatusNormalClosure, "bootstrap_complete")
					return
				default:
					if err := wsjson.Write(ctx, conn, platformws.Message{
						Type:    "protocol_error",
						Payload: collabtestprotocol.RawPayload(map[string]any{"reason_code": "unknown_message"}),
					}); err != nil {
						return
					}
				}
			}
		})
		return nil
	}
}
