package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
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
	buildRuntime func(context.Context, config.Config, Options) (serverRuntime, error)
	lookupEnv    func(string) (string, bool)
	profile      serverProfile
}

type serverRuntime struct {
	Handler              http.Handler
	Close                func()
	ActivatePublication  func() error
	FatalEvents          <-chan processlifecycle.FatalSignal
	Fatal                func(string) bool
	ShutdownDrainSeconds int64
}

func RunServerContext(ctx context.Context, stdout io.Writer, stderr io.Writer) int {
	return newServerRunner(stdout, stderr).run(ctx)
}

func newServerRunner(stdout io.Writer, stderr io.Writer) serverRunner {
	return serverRunner{
		stdout:     normalizeServerWriter(stdout),
		stderr:     normalizeServerWriter(stderr),
		loadConfig: config.Load,
		buildRuntime: func(ctx context.Context, cfg config.Config, options Options) (serverRuntime, error) {
			runtime, err := NewRuntime(ctx, cfg, options)
			if err != nil {
				return serverRuntime{}, err
			}
			return serverRuntime{
				Handler: runtime.Handler, Close: runtime.Close, ActivatePublication: runtime.ActivatePublication,
				FatalEvents: runtime.FatalEvents(), Fatal: runtime.Lifecycle.Fatal,
				ShutdownDrainSeconds: runtime.Config.Timeouts.Extensions.ShutdownDrainSeconds,
			}, nil
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
		return 2
	}

	cfg, err := runner.loadConfig()
	if err != nil {
		runner.writeStartupError(err, logger, "load config")
		return 2
	}

	runtime, err := runner.buildRuntime(ctx, cfg, runner.profile.runtimeOptions(runner.lookupEnv))
	if err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			return 0
		}
		if errors.Is(err, processlease.ErrLeaseLost) {
			runner.writeFatalDiagnostic(processlifecycle.FatalSignal{ReasonCode: "application_process_lease_lost", ExitCode: 70})
			return 70
		}
		runner.writeStartupError(err, logger, "setup runtime")
		return 2
	}
	if runtime.Close != nil {
		defer runtime.Close()
	}
	if runtime.ActivatePublication == nil {
		runner.writeStartupError(errors.New("extension_publication_failed"), logger, "activate publication")
		return 2
	}

	address := httpruntime.DefaultAddress
	if configuredAddress, ok := runner.lookupEnv(httpAddrEnv); ok && configuredAddress != "" {
		address = configuredAddress
	}
	serveCtx, cancelServe := context.WithCancel(ctx)
	defer cancelServe()
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- runner.profile.serve(serveCtx, runtime.Handler, httpruntime.Options{
			Address:         address,
			InheritedFD:     runner.profile.inheritedListenerFD(runner.lookupEnv),
			Logger:          logger,
			ShutdownTimeout: time.Duration(runtime.ShutdownDrainSeconds) * time.Second,
			OnReady:         runtime.ActivatePublication,
		})
	}()
	select {
	case fatal := <-runtime.FatalEvents:
		cancelServe()
		runner.awaitDrain(serveDone, runtime.ShutdownDrainSeconds)
		runner.writeFatalDiagnostic(fatal)
		return 70
	case serveErr := <-serveDone:
		if serveErr == nil {
			return 0
		}
		var startupErr *httpruntime.StartupError
		if errors.As(serveErr, &startupErr) {
			runner.writeStartupError(startupErr, logger, "start server listener")
			return 2
		}
		if runtime.Fatal != nil {
			runtime.Fatal("published_component_lost")
		}
		runner.writeFatalDiagnostic(processlifecycle.FatalSignal{ReasonCode: "published_component_lost", ExitCode: 70})
		return 70
	}
}

func (runner serverRunner) awaitDrain(serveDone <-chan error, seconds int64) {
	if seconds <= 0 {
		seconds = 1
	}
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	defer timer.Stop()
	select {
	case <-serveDone:
	case <-timer.C:
	}
}

func (runner serverRunner) writeFatalDiagnostic(signal processlifecycle.FatalSignal) {
	diagnostic := struct {
		Code       string `json:"code"`
		ReasonCode string `json:"reason_code"`
	}{Code: "extension_integrity_failure", ReasonCode: signal.ReasonCode}
	encoded, _ := json.Marshal(diagnostic)
	_, _ = runner.stderr.Write(append(encoded, '\n'))
}

func (runner serverRunner) writeStartupError(err error, logger *slog.Logger, action string) {
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) {
		_, _ = io.WriteString(runner.stderr, diagnosticsErr.JSON())
		_, _ = io.WriteString(runner.stderr, "\n")
		return
	}
	var remediationErr *postgres.MigrationRemediationError
	if errors.As(err, &remediationErr) {
		_, _ = io.WriteString(runner.stderr, remediationErr.ReportJSON())
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
