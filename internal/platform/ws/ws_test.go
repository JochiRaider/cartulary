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
	t.Run("frames over the read limit fail the read", func(t *testing.T) {
		readErrors := make(chan error, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connection, err := Accept(w, r, "")
			if err != nil {
				readErrors <- err
				return
			}
			defer connection.CloseNow()
			connection.SetReadLimit(4)
			_, _, err = connection.Read(r.Context())
			readErrors <- err
		}))
		defer server.Close()

		target := "ws" + strings.TrimPrefix(server.URL, "http")
		client, _, err := websocket.Dial(context.Background(), target, nil)
		if err != nil {
			t.Fatalf("dial transport server: %v", err)
		}
		defer client.CloseNow()
		// One byte over the limit; the 4-byte "ping" below is accepted.
		if err := client.Write(context.Background(), websocket.MessageText, []byte("pings")); err != nil {
			t.Fatalf("write client frame: %v", err)
		}
		select {
		case err := <-readErrors:
			if err == nil {
				t.Fatal("server read accepted a frame over the read limit")
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for the over-limit read")
		}
	})

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

	t.Run("only a missing or configured Origin passes the Origin check", func(t *testing.T) {
		const configured = "https://cartulary.example"
		for _, tc := range []struct {
			name    string
			origins []string
			cookie  bool
			allowed bool
		}{
			{name: "no Origin", allowed: true},
			{name: "no Origin with a session cookie", cookie: true, allowed: true},
			{name: "configured origin", origins: []string{configured}, allowed: true},
			{name: "configured origin with a session cookie", origins: []string{configured}, cookie: true, allowed: true},
			{name: "untrusted origin with a session cookie", origins: []string{"https://untrusted.example"}, cookie: true},
			{name: "untrusted origin without a session cookie", origins: []string{"https://untrusted.example"}},
			{name: "configured host on another scheme", origins: []string{"http://cartulary.example"}},
			{name: "opaque origin", origins: []string{"null"}},
			{name: "empty Origin header", origins: []string{""}},
			{name: "repeated Origin headers", origins: []string{configured, configured}},
		} {
			request := httptest.NewRequest(http.MethodGet, "http://cartulary.test/ws", nil)
			for _, origin := range tc.origins {
				request.Header.Add("Origin", origin)
			}
			if tc.cookie {
				request.AddCookie(&http.Cookie{Name: "cartulary_session", Value: "session"})
			}
			if got := OriginAllowed(request, configured); got != tc.allowed {
				t.Fatalf("%s: OriginAllowed = %t, want %t", tc.name, got, tc.allowed)
			}
		}
	})
}

type transportTestError struct {
	message string
}

func (e *transportTestError) Error() string {
	return e.message
}
