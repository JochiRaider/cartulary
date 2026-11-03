package httpruntime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	tlstest "github.com/JochiRaider/cartulary/internal/testutil/tlstest/transport"
)

func TestServePublicationAcceptLoopReadiness_Integration(t *testing.T) {
	t.Run("acknowledgment waits for the actual server loop", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		entered, release, acknowledged := make(chan struct{}), make(chan struct{}), make(chan struct{})
		var releaseOnce sync.Once
		unblock := func() { releaseOnce.Do(func() { close(release) }) }
		server := &http.Server{Handler: http.NotFoundHandler(), BaseContext: func(net.Listener) context.Context {
			close(entered)
			<-release
			return ctx
		}}
		rejected := errors.New("publication rejected")
		done, finished := make(chan error, 1), make(chan struct{})
		go func() {
			defer close(finished)
			done <- serveBoundListener(ctx, server, listener, Options{ShutdownTimeout: time.Second, OnReady: func() error {
				close(acknowledged)
				return rejected
			}})
		}()
		t.Cleanup(func() {
			unblock()
			cancel()
			_ = server.Close()
			_ = listener.Close()
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Error("bound listener did not stop")
			}
		})
		select {
		case <-entered:
		case <-acknowledged:
			t.Fatal("publication acknowledged before the HTTP server loop started")
		case <-time.After(time.Second):
			t.Fatal("HTTP server loop did not start")
		}
		select {
		case <-acknowledged:
			t.Fatal("publication acknowledged while server initialization was blocked")
		case <-time.After(20 * time.Millisecond):
		}
		unblock()
		var startup *StartupError
		if err := waitServe(t, done); !errors.As(err, &startup) || !errors.Is(err, rejected) {
			t.Fatalf("rejected publication outcome=%v", err)
		}
	})

	for _, failure := range []struct {
		name string
		stop func()
	}{
		{name: "panic", stop: func() { panic("private listener failure") }},
		{name: "goroutine exit", stop: runtime.Goexit},
	} {
		t.Run("prepared loop "+failure.name+" is component loss", func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- serveBoundListener(ctx, &http.Server{Handler: http.NotFoundHandler()}, abruptPublicationListener{Listener: listener, stop: failure.stop}, Options{ShutdownTimeout: time.Second})
			}()
			var lost *ComponentLossError
			if err := waitServe(t, done); !errors.As(err, &lost) || lost.Err == nil || lost.Err.Error() != "HTTP event loop terminated unexpectedly" {
				t.Fatalf("unexpected loop termination=%v", err)
			}
		})
	}
	t.Run("cancellation after preparation releases the bound address", func(t *testing.T) {
		identity := tlstest.NewServer(t, "127.0.0.1")
		probe, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		address := probe.Addr().String()
		_ = probe.Close()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		err = Serve(ctx, http.NotFoundHandler(), Options{TLS: identity.TLS, Address: address, Logger: discardLogger(), OnReady: func() error { cancel(); return nil }})
		if err != nil {
			t.Fatalf("canceled preparation=%v", err)
		}
		rebound, err := net.Listen("tcp", address)
		if err != nil {
			t.Fatalf("canceled preparation retained listener: %v", err)
		}
		_ = rebound.Close()
	})

	t.Run("TLS requests remain quiescent until activation", func(t *testing.T) {
		identity := tlstest.NewServer(t, "127.0.0.1")
		probe, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		address := probe.Addr().String()
		_ = probe.Close()
		ctx, cancel := context.WithCancel(t.Context())
		entered, release, handled := make(chan struct{}), make(chan struct{}), make(chan struct{}, 1)
		var releaseOnce sync.Once
		unblock := func() { releaseOnce.Do(func() { close(release) }) }
		done := make(chan error, 1)
		go func() {
			done <- Serve(ctx, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handled <- struct{}{}
				w.WriteHeader(http.StatusNoContent)
			}), Options{TLS: identity.TLS, Address: address, Logger: discardLogger(), OnReady: func() error {
				close(entered)
				<-release
				return nil
			}})
		}()
		t.Cleanup(func() { unblock(); cancel() })
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("publication callback was not reached")
		}
		client := identity.Client(t, "127.0.0.1")
		client.Timeout = 2 * time.Second
		response := make(chan error, 1)
		go func() {
			res, err := client.Get("https://" + address)
			if res != nil {
				_ = res.Body.Close()
				if res.StatusCode != http.StatusNoContent {
					err = errors.New("unexpected response status")
				}
			}
			response <- err
		}()
		select {
		case <-handled:
			t.Fatal("request executed before publication")
		case err := <-response:
			t.Fatalf("request completed before publication: %v", err)
		case <-time.After(20 * time.Millisecond):
		}
		unblock()
		if err := <-response; err != nil {
			t.Fatal(err)
		}
		cancel()
		if err := waitServe(t, done); err != nil {
			t.Fatal(err)
		}
	})
}

type abruptPublicationListener struct {
	net.Listener
	stop func()
}

func (listener abruptPublicationListener) Accept() (net.Conn, error) {
	listener.stop()
	return nil, net.ErrClosed
}
