package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"

var harnessOnlyServerEnv = []string{
	"CARTULARY_ENABLE_TEST_ROUTES",
	"CARTULARY_HTTP_LISTEN_FD",
}

type serverProfile interface {
	validateEnvironment(func(string) (string, bool)) error
	runtimeOptions(func(string) (string, bool)) Options
	inheritedListenerFD(func(string) (string, bool)) string
	serve(context.Context, http.Handler, httpruntime.Options) error
}

type serverRunner struct {
	stdout       io.Writer
	stderr       io.Writer
	loadConfig   func() (config.Config, error)
	buildRuntime func(context.Context, config.Config, Options) (http.Handler, func(), error)
	lookupEnv    func(string) (string, bool)
	profile      serverProfile
}

func RunServerContext(ctx context.Context, stdout io.Writer, stderr io.Writer) int {
	return newServerRunner(stdout, stderr).run(ctx)
}

func newServerRunner(stdout io.Writer, stderr io.Writer) serverRunner {
	return serverRunner{
		stdout:     normalizeServerWriter(stdout),
		stderr:     normalizeServerWriter(stderr),
		loadConfig: config.Load,
		buildRuntime: func(ctx context.Context, cfg config.Config, options Options) (http.Handler, func(), error) {
			runtime, err := NewRuntime(ctx, cfg, options)
			if err != nil {
				return nil, nil, err
			}
			return runtime.Handler, runtime.Close, nil
		},
		lookupEnv: os.LookupEnv,
		profile:   newServerProfile(),
	}
}

func (runner serverRunner) run(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	logger := slog.New(slog.NewTextHandler(runner.stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := runner.profile.validateEnvironment(runner.lookupEnv); err != nil {
		runner.writeStartupError(err, logger, "validate server profile")
		return 1
	}

	cfg, err := runner.loadConfig()
	if err != nil {
		runner.writeStartupError(err, logger, "load config")
		return 1
	}

	handler, closeRuntime, err := runner.buildRuntime(ctx, cfg, runner.profile.runtimeOptions(runner.lookupEnv))
	if err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			return 0
		}
		runner.writeStartupError(err, logger, "setup runtime")
		return 1
	}
	if closeRuntime != nil {
		defer closeRuntime()
	}

	address := httpruntime.DefaultAddress
	if configuredAddress, ok := runner.lookupEnv(httpAddrEnv); ok && configuredAddress != "" {
		address = configuredAddress
	}
	if err := runner.profile.serve(ctx, handler, httpruntime.Options{
		Address:     address,
		InheritedFD: runner.profile.inheritedListenerFD(runner.lookupEnv),
		Logger:      logger,
	}); err != nil {
		logger.Error("server exited", "error", err)
		return 1
	}
	return 0
}

func (runner serverRunner) writeStartupError(err error, logger *slog.Logger, action string) {
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) {
		_, _ = io.WriteString(runner.stderr, diagnosticsErr.JSON())
		_, _ = io.WriteString(runner.stderr, "\n")
		return
	}
	logger.Error(action, "error", err)
}

func normalizeServerWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
