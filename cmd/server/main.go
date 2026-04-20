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

	"github.com/JochiRaider/cartulary/internal/app"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"
const enableTestRoutesEnv = "CARTULARY_ENABLE_TEST_ROUTES"

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

	options := app.Options{}
	if os.Getenv(enableTestRoutesEnv) == "1" {
		testClock := httpapi.NewTestClock()
		options.Now = testClock.Now
		options.HTTP.AdditionalRoutes = []httpapi.RouteRegistrar{
			httpapi.RegisterTestClockRoutes(testClock),
			auth.RegisterTestRoutes(),
			timeline.RegisterTestRoutes(),
		}
	}

	runtime, err := app.NewRuntime(ctx, cfg, options)
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
