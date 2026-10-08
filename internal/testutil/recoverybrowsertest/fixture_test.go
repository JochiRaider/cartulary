package recoverybrowsertest

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest/transport"
)

func TestRuntimeServerCloseCancelsHTTPAndWebSocketRequests(t *testing.T) {
	identity := transport.NewServer(t, "127.0.0.1")
	listener, err := tls.Listen("tcp", "127.0.0.1:0", identity.TLS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	httpStarted, httpStopped := make(chan struct{}), make(chan struct{})
	websocketStopped := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/socket" {
			defer close(websocketStopped)
			connection, acceptErr := websocket.Accept(w, r, nil)
			if acceptErr != nil {
				t.Errorf("accept WebSocket: %v", acceptErr)
				return
			}
			defer connection.CloseNow()
			_, _, _ = connection.Read(r.Context())
			return
		}
		defer close(httpStopped)
		close(httpStarted)
		<-r.Context().Done()
		w.WriteHeader(http.StatusNoContent)
	})
	server, served, err := startRuntimeServer(ctx, listener, handler, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client := identity.Client(t, "127.0.0.1")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+listener.Addr().String()+"/waiting", nil)
	if err != nil {
		t.Fatal(err)
	}
	responseDone := make(chan error, 1)
	go func() {
		response, requestErr := client.Do(request)
		if requestErr == nil {
			_, requestErr = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusNoContent {
				t.Errorf("HTTP status = %d", response.StatusCode)
			}
		}
		responseDone <- requestErr
	}()
	select {
	case <-httpStarted:
	case <-ctx.Done():
		t.Fatal("HTTP handler did not start")
	}
	connection, _, err := websocket.Dial(ctx, "wss://"+listener.Addr().String()+"/socket", &websocket.DialOptions{HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	// Neither client closes first: EOF retirement must end both request lifetimes.
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	for name, stopped := range map[string]<-chan struct{}{"HTTP": httpStopped, "WebSocket": websocketStopped} {
		select {
		case <-stopped:
		default:
			t.Errorf("%s handler remains active after Close", name)
		}
	}
	if err := server.Close(); err != nil {
		t.Fatalf("repeated Close: %v", err)
	}
	for name, result := range map[string]<-chan error{"Serve": served, "HTTP response": responseDone} {
		select {
		case err := <-result:
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
		case <-ctx.Done():
			t.Errorf("%s did not finish", name)
		}
	}
}

func TestRuntimeServerCloseRetiresTLSConnectionsWithoutRequests(t *testing.T) {
	identity := transport.NewServer(t, "127.0.0.1")
	listener, err := tls.Listen("tcp", "127.0.0.1:0", identity.TLS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	server, served, err := startRuntimeServer(ctx, listener, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("preconnected client submitted no request")
	}), func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client := identity.Client(t, "127.0.0.1")
	dialer := tls.Dialer{Config: client.Transport.(*http.Transport).TLSClientConfig}
	connection, err := dialer.DialContext(ctx, "tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	// A completed TLS handshake is still StateNew until HTTP headers arrive.
	// There is no request work to drain and no client closure to wait for.
	// Observe the server's completion: Dial can return before the server has
	// consumed the client's Finished message, which cancellation can still abort.
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		complete := false
		server.connectionsMu.Lock()
		for accepted := range server.connections {
			if connection, ok := accepted.(*tls.Conn); ok && connection.ConnectionState().HandshakeComplete {
				complete = true
			}
		}
		server.connectionsMu.Unlock()
		if complete {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("server did not complete TLS handshake")
		case <-tick.C:
		}
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("Serve did not retire")
	}
}
