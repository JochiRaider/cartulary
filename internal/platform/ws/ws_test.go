package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestConnectionFrameIOAndOrigin(t *testing.T) {
	t.Run("text frame I/O stays transport generic", func(t *testing.T) {
		serverErrors := make(chan error, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connection, err := Accept(w, r, "")
			if err != nil {
				serverErrors <- err
				return
			}
			defer connection.CloseNow()
			connection.SetReadLimit(4)
			kind, payload, err := connection.Read(r.Context())
			if err != nil {
				serverErrors <- err
				return
			}
			if kind != MessageText || string(payload) != "ping" {
				serverErrors <- &transportTestError{message: "unexpected inbound frame"}
				return
			}
			serverErrors <- connection.Write(r.Context(), MessageText, []byte("pong"))
		}))
		defer server.Close()

		target := "ws" + strings.TrimPrefix(server.URL, "http")
		client, _, err := websocket.Dial(context.Background(), target, nil)
		if err != nil {
			t.Fatalf("dial transport server: %v", err)
		}
		defer client.CloseNow()
		if err := client.Write(context.Background(), websocket.MessageText, []byte("ping")); err != nil {
			t.Fatalf("write client frame: %v", err)
		}
		kind, payload, err := client.Read(context.Background())
		if err != nil {
			t.Fatalf("read server frame: %v", err)
		}
		if kind != websocket.MessageText || string(payload) != "pong" {
			t.Fatalf("outbound frame = %v %q", kind, payload)
		}
		select {
		case err := <-serverErrors:
			if err != nil {
				t.Fatalf("server transport: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for transport server")
		}
	})

	t.Run("cookie browser Origin check fails before upgrade", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "http://cartulary.test/ws", nil)
		request.AddCookie(&http.Cookie{Name: "cartulary_session", Value: "session"})
		request.Header.Set("Origin", "https://untrusted.example")
		recorder := httptest.NewRecorder()
		if !RejectUntrustedBrowserOrigin(recorder, request, "https://cartulary.example") {
			t.Fatal("expected untrusted cookie-browser Origin rejection")
		}
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("Origin response = %d want %d", recorder.Code, http.StatusForbidden)
		}
	})
}

type transportTestError struct {
	message string
}

func (e *transportTestError) Error() string {
	return e.message
}
