package httpruntime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAddress           = ":8080"
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultShutdownTimeout   = 5 * time.Second
)

type Options struct {
	Address           string
	InheritedFD       string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
	Logger            *slog.Logger
}

func Serve(ctx context.Context, handler http.Handler, options Options) error {
	if err := ctx.Err(); err != nil {
		return nil
	}

	options = normalizeOptions(options)
	listener, inherited, err := openListener(options)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           handler,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
	}
	if inherited {
		options.Logger.Info("serving cartulary bootstrap server on inherited listener", "addr", server.Addr)
	} else {
		options.Logger.Info("starting cartulary bootstrap server", "addr", server.Addr)
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()

	select {
	case serveErr := <-serveDone:
		if errors.Is(serveErr, http.ErrServerClosed) {
			return nil
		}
		return serveErr
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
		shutdownErr := server.Shutdown(shutdownCtx)
		cancel()
		if shutdownErr != nil {
			_ = server.Close()
		}

		serveErr := <-serveDone
		if shutdownErr != nil {
			return fmt.Errorf("shutdown http server: %w", shutdownErr)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
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
