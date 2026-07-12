package httpruntime

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServeInheritedListenerAndGracefulCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	tcpListener := listener.(*net.TCPListener)
	listenerFile, err := tcpListener.File()
	if err != nil {
		t.Fatalf("listener file: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	go func() {
		done <- Serve(ctx, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(requestStarted)
			<-releaseRequest
			w.WriteHeader(http.StatusNoContent)
		}), Options{InheritedFD: fileDescriptorString(listenerFile), Logger: discardLogger()})
	}()

	responseDone := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + address)
		if resp != nil {
			_ = resp.Body.Close()
		}
		responseDone <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("inherited listener request did not start")
	}
	cancel()
	select {
	case err := <-done:
		t.Fatalf("graceful shutdown returned before request completed: %v", err)
	default:
	}
	close(releaseRequest)
	if err := waitServe(t, done); err != nil {
		t.Fatalf("graceful serve: %v", err)
	}
	if err := <-responseDone; err != nil {
		t.Fatalf("in-flight request: %v", err)
	}
}

func TestServeOrdinaryListenerReportsEffectiveAddress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var output synchronizedBuffer
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, http.NotFoundHandler(), Options{
			Address: "127.0.0.1:0",
			Logger:  slog.New(slog.NewTextHandler(&output, nil)),
		})
	}()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(output.String(), "addr=127.0.0.1:") {
		time.Sleep(10 * time.Millisecond)
	}
	if got := output.String(); !strings.Contains(got, "starting cartulary bootstrap server") || !strings.Contains(got, "addr=127.0.0.1:") {
		t.Fatalf("effective listening address was not logged: %q", got)
	}
	cancel()
	if err := waitServe(t, done); err != nil {
		t.Fatalf("ordinary listener shutdown: %v", err)
	}
}

func TestServeRejectsMalformedAndClosedInheritedFD(t *testing.T) {
	for _, tc := range []struct {
		name string
		fd   string
		want string
	}{
		{name: "malformed", fd: "not-a-fd", want: "parse inherited http listener fd"},
		{name: "negative", fd: "-1", want: "must be non-negative"},
		{name: "closed", fd: "999999", want: "convert inherited http listener fd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Serve(context.Background(), http.NotFoundHandler(), Options{InheritedFD: tc.fd, Logger: discardLogger()})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Serve() error got %v want containing %q", err, tc.want)
			}
		})
	}
}

func TestServeRejectsInheritedNonListenerFD(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "not-a-listener")
	if err != nil {
		t.Fatalf("create non-listener file: %v", err)
	}
	defer file.Close()
	err = Serve(context.Background(), http.NotFoundHandler(), Options{
		InheritedFD: fileDescriptorString(file),
		Logger:      discardLogger(),
	})
	if err == nil || !strings.Contains(err.Error(), "convert inherited http listener fd") {
		t.Fatalf("non-listener fd error got %v", err)
	}
}

func TestServeCancelledBeforeListenDoesNotBind(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve address: %v", err)
	}
	address := probe.Addr().String()
	_ = probe.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Serve(ctx, http.NotFoundHandler(), Options{Address: address, Logger: discardLogger()}); err != nil {
		t.Fatalf("cancelled Serve() error: %v", err)
	}

	rebound, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("cancelled Serve() retained listener: %v", err)
	}
	_ = rebound.Close()
}

func TestServeForcesCloseAfterShutdownTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	requestStarted := make(chan struct{})
	handlerExited := make(chan struct{})
	var once sync.Once
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(requestStarted) })
		<-r.Context().Done()
		close(handlerExited)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	listenerFile, err := listener.(*net.TCPListener).File()
	if err != nil {
		t.Fatalf("listener file: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, handler, Options{
			InheritedFD:     fileDescriptorString(listenerFile),
			ShutdownTimeout: 10 * time.Millisecond,
			Logger:          discardLogger(),
		})
	}()

	responseDone := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + address)
		if resp != nil {
			_ = resp.Body.Close()
		}
		responseDone <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	serveErr := waitServe(t, done)
	if serveErr == nil || !errors.Is(serveErr, context.DeadlineExceeded) {
		t.Fatalf("forced shutdown error got %v", serveErr)
	}
	select {
	case <-handlerExited:
	case <-time.After(5 * time.Second):
		t.Fatal("forced close did not cancel handler")
	}
	select {
	case <-responseDone:
	case <-time.After(5 * time.Second):
		t.Fatal("client did not return after forced close")
	}
}

func fileDescriptorString(file *os.File) string {
	return strconv.Itoa(int(file.Fd()))
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type synchronizedBuffer struct {
	mu     sync.Mutex
	buffer strings.Builder
}

func (buffer *synchronizedBuffer) Write(body []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.Write(body)
}

func (buffer *synchronizedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.String()
}

func waitServe(t testing.TB, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Serve")
		return nil
	}
}
