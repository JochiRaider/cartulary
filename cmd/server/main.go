package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/todo/cartulary/internal/app"
	"example.com/todo/cartulary/internal/platform/config"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		writeStartupError(err, logger, "load config")
		os.Exit(1)
	}

	runtime, err := app.NewRuntime(ctx, cfg, app.Options{})
	if err != nil {
		writeStartupError(err, logger, "setup runtime")
		os.Exit(1)
	}
	defer runtime.Close()

	addr := ":8080"
	if configuredAddr, ok := os.LookupEnv(httpAddrEnv); ok && configuredAddr != "" {
		addr = configuredAddr
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           runtime.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Info("starting cartulary bootstrap server", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func writeStartupError(err error, logger *slog.Logger, action string) {
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) {
		_, _ = os.Stderr.WriteString(diagnosticsErr.JSON())
		_, _ = os.Stderr.WriteString("\n")
		return
	}

	logger.Error(action, "error", err)
}
