package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

const (
	httpAddrEnv         = "CARTULARY_HTTP_ADDR"
	httpListenFDEnv     = "CARTULARY_HTTP_LISTEN_FD"
	enableTestRoutesEnv = "CARTULARY_ENABLE_TEST_ROUTES"
)

type serverRunner struct {
	stdout       io.Writer
	stderr       io.Writer
	loadConfig   func() (config.Config, error)
	buildRuntime func(context.Context, config.Config, Options) (http.Handler, func(), error)
	serve        func(context.Context, http.Handler, httpruntime.Options) error
	lookupEnv    func(string) (string, bool)
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
		serve:     httpruntime.Serve,
		lookupEnv: os.LookupEnv,
	}
}

func (runner serverRunner) run(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	logger := slog.New(slog.NewTextHandler(runner.stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := runner.loadConfig()
	if err != nil {
		runner.writeStartupError(err, logger, "load config")
		return 1
	}

	handler, closeRuntime, err := runner.buildRuntime(ctx, cfg, runner.runtimeOptions())
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
	inheritedFD, _ := runner.lookupEnv(httpListenFDEnv)
	if err := runner.serve(ctx, handler, httpruntime.Options{
		Address:     address,
		InheritedFD: inheritedFD,
		Logger:      logger,
	}); err != nil {
		logger.Error("server exited", "error", err)
		return 1
	}
	return 0
}

func (runner serverRunner) runtimeOptions() Options {
	options := Options{}
	enabled, _ := runner.lookupEnv(enableTestRoutesEnv)
	if enabled != "1" {
		return options
	}

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
	return options
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
