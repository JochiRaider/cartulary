package httpruntime

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAddress           = ":8080"
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultShutdownTimeout   = 5 * time.Second
)

type Options struct {
	// TLS is the immutable identity admitted before application resource acquisition.
	TLS               *tls.Config
	Address           string
	InheritedFD       string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
	// ShutdownDeadline, when set, bounds the drain by an absolute deadline
	// evaluated when shutdown begins, instead of ShutdownTimeout measured from
	// that moment. The application owns one shutdown deadline and reserves
	// part of it for component closes (Core 04 REQ-04-146a).
	ShutdownDeadline func() time.Time
	Logger           *slog.Logger
	OnReady          func() error
	// OnShutdown runs when ordinary shutdown begins (http.Server.Shutdown), so
	// owners of hijacked connections such as WebSockets can close them; the
	// server does not wait on hijacked connections.
	OnShutdown func()
}

type StartupError struct{ Err error }

func (e *StartupError) Error() string { return "http listener startup failed" }
func (e *StartupError) Unwrap() error { return e.Err }

// DrainDeadlineError reports that the drain reached its deadline: the server
// force-closed the remaining connections, cancelling their in-flight work.
type DrainDeadlineError struct{ Err error }

func (e *DrainDeadlineError) Error() string { return "http drain reached its deadline" }
func (e *DrainDeadlineError) Unwrap() error { return e.Err }

type ComponentLossError struct{ Err error }

func (e *ComponentLossError) Error() string { return "published http listener terminated" }
func (e *ComponentLossError) Unwrap() error { return e.Err }

func Serve(ctx context.Context, handler http.Handler, options Options) error {
	if err := ctx.Err(); err != nil {
		return nil
	}

	options = normalizeOptions(options)
	if options.TLS == nil {
		return &StartupError{Err: errors.New("admitted application TLS identity is required")}
	}
	listener, inherited, err := openListener(options)
	if err != nil {
		return &StartupError{Err: err}
	}
	defer listener.Close()
	listener = tls.NewListener(listener, options.TLS.Clone())

	server := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           handler,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		// The server's own connection errors, such as a client that drops a
		// TLS handshake, are structured log records. The standard logger would
		// write them as raw lines on stderr, which carries only the process's
		// safe diagnostic (AC-604).
		ErrorLog: slog.NewLogLogger(options.Logger.Handler(), slog.LevelWarn),
	}
	if inherited {
		options.Logger.Info("serving cartulary bootstrap server on inherited listener", "addr", server.Addr)
	} else {
		options.Logger.Info("starting cartulary bootstrap server", "addr", server.Addr)
	}
	return serveBoundListener(ctx, server, listener, options)
}

// serveBoundListener owns activation and shutdown after listener admission.
func serveBoundListener(ctx context.Context, server *http.Server, listener net.Listener, options Options) error {
	if options.OnShutdown != nil {
		server.RegisterOnShutdown(options.OnShutdown)
	}
	gate := newPublicationListener(listener)
	defer gate.Close()
	serveDone := make(chan error, 1)
	go func() {
		// A lost accept loop must notify its supervisor even after panic or Goexit.
		// Never project the panic payload into startup/fatal diagnostics.
		serveErr := errors.New("HTTP event loop terminated unexpectedly")
		defer func() { _ = recover(); serveDone <- serveErr }()
		serveErr = server.Serve(gate)
	}()
	stopPrepared := func() {
		_ = server.Close()
		_ = gate.Close()
		<-serveDone
	}
	select {
	case <-gate.ready:
	case serveErr := <-serveDone:
		return &StartupError{Err: serveErr}
	case <-ctx.Done():
		stopPrepared()
		return nil
	}
	if ctx.Err() != nil {
		stopPrepared()
		return nil
	}
	if options.OnReady != nil {
		if err := options.OnReady(); err != nil {
			stopPrepared()
			return &StartupError{Err: fmt.Errorf("activate publication: %w", err)}
		}
	}
	if ctx.Err() != nil {
		stopPrepared()
		return nil
	}
	close(gate.admitted)

	select {
	case serveErr := <-serveDone:
		if ctx.Err() != nil && errors.Is(serveErr, http.ErrServerClosed) {
			return nil
		}
		return &ComponentLossError{Err: serveErr}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
		if options.ShutdownDeadline != nil {
			cancel()
			shutdownCtx, cancel = context.WithDeadline(context.Background(), options.ShutdownDeadline())
		}
		shutdownErr := server.Shutdown(shutdownCtx)
		cancel()
		if shutdownErr != nil {
			_ = server.Close()
		}

		serveErr := <-serveDone
		if errors.Is(shutdownErr, context.DeadlineExceeded) {
			return &DrainDeadlineError{Err: shutdownErr}
		}
		if shutdownErr != nil {
			return fmt.Errorf("shutdown http server: %w", shutdownErr)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}

// publicationListener acknowledges only after http.Server enters its accept
// loop. It accepts no connection until the publication callback succeeds.
// Closing a prepared listener releases that loop without admitting work.
type publicationListener struct {
	net.Listener
	ready     chan struct{}
	admitted  chan struct{}
	stopped   chan struct{}
	readyOnce sync.Once
	closeOnce sync.Once
	closeErr  error
}

func newPublicationListener(listener net.Listener) *publicationListener {
	return &publicationListener{Listener: listener, ready: make(chan struct{}), admitted: make(chan struct{}), stopped: make(chan struct{})}
}

func (listener *publicationListener) Accept() (net.Conn, error) {
	listener.readyOnce.Do(func() { close(listener.ready) })
	select {
	case <-listener.stopped:
		return nil, net.ErrClosed
	case <-listener.admitted:
	}
	select {
	case <-listener.stopped:
		return nil, net.ErrClosed
	default:
	}
	return listener.Listener.Accept()
}

func (listener *publicationListener) Close() error {
	listener.closeOnce.Do(func() {
		close(listener.stopped)
		listener.closeErr = listener.Listener.Close()
	})
	return listener.closeErr
}

func normalizeOptions(options Options) Options {
	if strings.TrimSpace(options.Address) == "" {
		options.Address = DefaultAddress
	}
	if options.ReadHeaderTimeout <= 0 {
		options.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}
	if options.ShutdownTimeout <= 0 {
		options.ShutdownTimeout = DefaultShutdownTimeout
	}
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return options
}

func openListener(options Options) (net.Listener, bool, error) {
	configuredFD := strings.TrimSpace(options.InheritedFD)
	if configuredFD == "" {
		listener, err := net.Listen("tcp", options.Address)
		if err != nil {
			return nil, false, fmt.Errorf("listen on %s: %w", options.Address, err)
		}
		return listener, false, nil
	}

	fd, err := strconv.Atoi(configuredFD)
	if err != nil || fd < 0 {
		if err == nil {
			err = fmt.Errorf("file descriptor must be non-negative")
		}
		return nil, true, fmt.Errorf("parse inherited http listener fd %q: %w", configuredFD, err)
	}

	listenerFile := os.NewFile(uintptr(fd), "cartulary-http-listener")
	if listenerFile == nil {
		return nil, true, errors.New("create inherited http listener file")
	}
	listener, listenerErr := net.FileListener(listenerFile)
	closeErr := listenerFile.Close()
	if listenerErr != nil {
		return nil, true, fmt.Errorf("convert inherited http listener fd %d: %w", fd, listenerErr)
	}
	if closeErr != nil {
		_ = listener.Close()
		return nil, true, fmt.Errorf("close inherited http listener file: %w", closeErr)
	}
	return listener, true, nil
}
