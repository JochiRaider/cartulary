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

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

const httpAddrEnv = "CARTULARY_HTTP_ADDR"

var harnessOnlyServerEnv = []string{
	"CARTULARY_ENABLE_TEST_ROUTES",
	"CARTULARY_HTTP_LISTEN_FD",
	conflicttokens.ConflictTokenFixtureRuntimeEnvName,
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
	loadConfig   func() (configassembly.Loaded, error)
	buildRuntime func(context.Context, configassembly.Loaded, Options) (serverRuntime, error)
	lookupEnv    func(string) (string, bool)
	profile      serverProfile
}

type serverRuntime struct {
	Handler              http.Handler
	Close                func()
	ActivatePublication  func() error
	FatalEvents          <-chan processlifecycle.FatalSignal
	Fatal                func(string) bool
	ShutdownDrainTimeout time.Duration
	PublicHTTP           httpapi.RouteDiagnostics
}

func RunServerContext(ctx context.Context, stdout io.Writer, stderr io.Writer) int {
	return newServerRunner(stdout, stderr).run(ctx)
}

func newServerRunner(stdout io.Writer, stderr io.Writer) serverRunner {
	return serverRunner{
		stdout: normalizeServerWriter(stdout),
		stderr: normalizeServerWriter(stderr),
		loadConfig: func() (configassembly.Loaded, error) {
			loaded, err := configassembly.Load(configassembly.LoadOptions{})
			if err != nil {
				return configassembly.Loaded{}, err
			}
			return loaded, nil
		},
		buildRuntime: func(ctx context.Context, loaded configassembly.Loaded, options Options) (serverRuntime, error) {
			runtime, err := newRuntime(ctx, loaded, options)
			if err != nil {
				return serverRuntime{}, err
			}
			return serverRuntime{
				Handler: runtime.HTTPHandler(), Close: runtime.Close, ActivatePublication: runtime.ActivatePublication,
				FatalEvents: runtime.fatalEvents(), Fatal: runtime.publishedComponentLost,
				ShutdownDrainTimeout: runtime.drainTimeout(),
				PublicHTTP:           runtime.publicHTTPDiagnostics(),
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
		var fatalStartup interface{ FatalReasonCode() string }
		if errors.As(err, &fatalStartup) && fatalStartup.FatalReasonCode() != "" {
			runner.writeFatalDiagnostic(processlifecycle.FatalSignal{ReasonCode: fatalStartup.FatalReasonCode(), ExitCode: 70})
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
	logger.Info(
		"public HTTP contract admitted",
		slog.String("openapi_version", runtime.PublicHTTP.DocumentVersion),
		slog.String("openapi_sha256", runtime.PublicHTTP.CanonicalSHA256),
		slog.Int("supported_operation_count", runtime.PublicHTTP.SupportedOperationCount),
		slog.Int("active_operation_count", runtime.PublicHTTP.ActiveOperationCount),
		slog.Any("claimed_profiles", runtime.PublicHTTP.ClaimedProfiles),
	)

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
			ShutdownTimeout: runtime.ShutdownDrainTimeout,
			OnReady:         runtime.ActivatePublication,
		})
	}()
	select {
	case fatal := <-runtime.FatalEvents:
		cancelServe()
		runner.awaitDrain(serveDone, runtime.ShutdownDrainTimeout)
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

func (runner serverRunner) awaitDrain(serveDone <-chan error, timeout time.Duration) {
	if timeout <= 0 {
		timeout = time.Second
	}
	timer := time.NewTimer(timeout)
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
	var remediationErr database_migrations.RemediationReporter
	if errors.As(err, &remediationErr) {
		_, _ = io.WriteString(runner.stderr, remediationErr.RemediationReportJSON())
		_, _ = io.WriteString(runner.stderr, "\n")
		return
	}
	var migrationFailure database_migrations.MigrationFailure
	if errors.As(err, &migrationFailure) {
		diagnostics := config.NewDiagnosticsError(config.Diagnostic{
			Path:       "database.schema_version",
			ReasonCode: migrationFailure.ReasonCode(),
			Message:    "Database migration validation failed.",
		})
		_, _ = io.WriteString(runner.stderr, diagnostics.JSON())
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
