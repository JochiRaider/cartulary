package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/app"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"
const httpListenFDEnv = "CARTULARY_HTTP_LISTEN_FD"
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
		harnessControls := harnessruntime.NewControls()
		networkFlowControls := networkflowharnesscontrol.NewControls()
		options.Now = testClock.Now
		options.HTTP.Dependencies.PublicErrorFaults = harnessControls.PublicErrorFaults
		options.HTTP.AdditionalRoutes = append(
			harnessruntime.RegisterRoutes(harnessControls, testClock, networkFlowControls.Contribution()),
			auth.RegisterTestRoutes(),
			savedviews.RegisterTestRoutes(),
			timeline.RegisterTestRoutes(),
		)
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
	if err := serveHTTP(server, logger); err != nil && err != http.ErrServerClosed {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func serveHTTP(server *http.Server, logger *slog.Logger) error {
	if configuredFD := os.Getenv(httpListenFDEnv); configuredFD != "" {
		fd, err := strconv.Atoi(configuredFD)
		if err != nil {
			return err
		}
		listenerFile := os.NewFile(uintptr(fd), "cartulary-http-listener")
		if listenerFile == nil {
			return errors.New("create inherited http listener file")
		}
		defer listenerFile.Close()

		listener, err := net.FileListener(listenerFile)
		if err != nil {
			return err
		}
		defer listener.Close()

		server.Addr = listener.Addr().String()
		logger.Info("serving cartulary bootstrap server on inherited listener", "addr", server.Addr)
		return server.Serve(listener)
	}

	return server.ListenAndServe()
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
