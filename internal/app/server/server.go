package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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
	publication          *publicationOperation
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
	var browser *webassets.Build
	return serverRunner{
		stdout: normalizeServerWriter(stdout),
		stderr: normalizeServerWriter(stderr),
		loadConfig: func() (configassembly.Loaded, error) {
			var err error
			browser, err = webassets.LoadEmbedded()
			if err != nil {
				return configassembly.Loaded{}, err
			}
			coordinator, err := extensionassembly.BrowserCoordinator(browser)
			if err != nil {
				return configassembly.Loaded{}, err
			}
			policy, err := extensionassembly.ConfigurationPolicyFromCoordinator(coordinator)
			if err != nil {
				return configassembly.Loaded{}, err
			}
			loaded, err := configassembly.LoadWithExtensionPolicy(configassembly.LoadOptions{}, policy)
			if err != nil {
				return configassembly.Loaded{}, err
			}
			return loaded, nil
		},
		buildRuntime: func(ctx context.Context, loaded configassembly.Loaded, options Options) (serverRuntime, error) {
			options.BrowserBuild = browser
			runtime, err := newRuntime(ctx, loaded, options)
			if err != nil {
				return serverRuntime{}, err
			}
			return serverRuntime{
				Handler: runtime.HTTPHandler(), Close: runtime.Close, ActivatePublication: runtime.ActivatePublication,
				publication: runtime.publicationOperation,
				FatalEvents: runtime.fatalEvents(), Fatal: runtime.publishedComponentLost,
				ShutdownDrainTimeout: runtime.drainTimeout(),
				PublicHTTP:           runtime.publicHTTPDiagnostics(),
			}, nil
		},
		lookupEnv: os.LookupEnv,
		profile:   newServerProfile(),
	}
}

func (runner serverRunner) run(ctx context.Context) (exitCode int) {
	fatalReported := false
	writeFatal := func(signal processlifecycle.FatalSignal) {
		if !fatalReported {
			runner.writeFatalDiagnostic(signal)
			fatalReported = true
		}
	}
	if ctx.Err() != nil {
		return 0
	}
	logger := slog.New(slog.NewTextHandler(runner.stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := cryptography.AdmitExecution(); err != nil {
		runner.writeStartupError(err, logger, "admit cryptographic execution")
		return 2
	}
	if err := runner.profile.validateEnvironment(runner.lookupEnv); err != nil {
		runner.writeStartupError(err, logger, "validate server profile")
		return 2
	}

	cfg, err := runner.loadConfig()
	if err != nil {
		runner.writeStartupError(err, logger, "load config")
		return 2
	}

	application := cfg.Deployment().Application
	origin, err := url.Parse(application.PublicOrigin)
	if err != nil || origin.Scheme != "https" || origin.Hostname() == "" {
		runner.writeStartupError(applicationTLSDiagnostic(), logger, "admit application TLS")
		return 2
	}
	tlsIdentity, err := cryptography.TLSServer(origin.Hostname(), application.TLSCertificatePath, application.TLSPrivateKeyPath)
	if err != nil {
		runner.writeStartupError(applicationTLSDiagnostic(), logger, "admit application TLS")
		return 2
	}

	forcedFatal := make(chan processlifecycle.FatalSignal, 1)
	defer func() {
		select {
		case signal := <-forcedFatal:
			writeFatal(signal)
			exitCode = 70
		default:
		}
	}()
	options := runner.profile.runtimeOptions(runner.lookupEnv)
	options.immediateFatal = func(signal processlifecycle.FatalSignal) {
		select {
		case forcedFatal <- signal:
		default:
		}
	}
	type buildResult struct {
		runtime serverRuntime
		err     error
	}
	built := make(chan buildResult, 1)
	go func() {
		var result buildResult
		defer func() { built <- result }()
		result.runtime, result.err = runner.buildRuntime(ctx, cfg, options)
	}()
	var runtime serverRuntime
	select {
	case signal := <-forcedFatal:
		writeFatal(signal)
		return 70
	case result := <-built:
		runtime, err = result.runtime, result.err
	}
	if err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			return 0
		}
		var fatalStartup interface{ FatalReasonCode() string }
		if errors.As(err, &fatalStartup) && fatalStartup.FatalReasonCode() != "" {
			writeFatal(processlifecycle.FatalSignal{ReasonCode: fatalStartup.FatalReasonCode(), ExitCode: 70})
			return 70
		}
		runner.writeStartupError(err, logger, "setup runtime")
		return 2
	}
	defer func() {
		select {
		case signal := <-runtime.FatalEvents:
			writeFatal(signal)
			exitCode = 70
		default:
		}
	}()
	forceTerminating := false
	var shutdownDeadline time.Time
	if runtime.Close != nil {
		defer func() {
			if forceTerminating {
				return
			}
			closed := make(chan struct{})
			go func() { defer close(closed); runtime.Close() }()
			var shutdownExpired <-chan time.Time
			if !shutdownDeadline.IsZero() {
				timer := time.NewTimer(time.Until(shutdownDeadline))
				defer timer.Stop()
				shutdownExpired = timer.C
			}
			select {
			case <-closed:
			case signal := <-forcedFatal:
				writeFatal(signal)
				exitCode = 70
			case <-shutdownExpired:
				exitCode = 70
			}
		}()
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
	var publicationFailures <-chan error
	if runtime.publication != nil {
		publicationFailures = runtime.publication.failures
	}
	go func() {
		serveDone <- runner.profile.serve(serveCtx, runtime.Handler, httpruntime.Options{
			TLS:             tlsIdentity,
			Address:         address,
			InheritedFD:     runner.profile.inheritedListenerFD(runner.lookupEnv),
			Logger:          logger,
			ShutdownTimeout: runtime.ShutdownDrainTimeout,
			OnReady:         runtime.ActivatePublication,
		})
	}()
	select {
	case fatal := <-forcedFatal:
		forceTerminating = true
		cancelServe()
		writeFatal(fatal)
		return 70
	case fatal := <-runtime.FatalEvents:
		cancelServe()
		if fatal.ReasonCode == "operation_cancellation_grace_exhausted" {
			forceTerminating = true
		} else {
			timeout := runtime.ShutdownDrainTimeout
			if timeout <= 0 {
				timeout = time.Second
			}
			shutdownDeadline = time.Now().Add(timeout)
			forceTerminating = runner.awaitDrain(serveDone, timeout, forcedFatal)
		}
		writeFatal(fatal)
		return 70
	case publicationErr := <-publicationFailures:
		cancelServe()
		var fatal interface{ FatalReasonCode() string }
		if errors.As(publicationErr, &fatal) {
			forceTerminating = fatal.FatalReasonCode() == "operation_cancellation_grace_exhausted"
			if !forceTerminating {
				shutdownDeadline = time.Now().Add(runtime.ShutdownDrainTimeout)
				forceTerminating = runner.awaitDrain(serveDone, runtime.ShutdownDrainTimeout, forcedFatal)
			}
			writeFatal(processlifecycle.FatalSignal{ReasonCode: fatal.FatalReasonCode(), ExitCode: 70})
			return 70
		}
		joined := make(chan error, 1)
		go func() { joined <- runtime.publication.joinTransport(serveDone) }()
		select {
		case signal := <-forcedFatal:
			forceTerminating = true
			writeFatal(signal)
			return 70
		case joinErr := <-joined:
			if joinErr != nil {
				forceTerminating = true
				writeFatal(processlifecycle.FatalSignal{ReasonCode: "operation_cancellation_grace_exhausted", ExitCode: 70})
				return 70
			}
		}
		if errors.Is(publicationErr, context.Canceled) && ctx.Err() != nil {
			return 0
		}
		runner.writeStartupError(publicationErr, logger, "publish runtime")
		return 2
	case serveErr := <-serveDone:
		if serveErr == nil {
			if runtime.publication != nil {
				runtime.publication.cancel()
				select {
				case signal := <-forcedFatal:
					forceTerminating = true
					writeFatal(signal)
					return 70
				case <-runtime.publication.done:
				}
				if publicationErr := runtime.publication.err; publicationErr != nil {
					var fatal interface{ FatalReasonCode() string }
					if errors.As(publicationErr, &fatal) {
						forceTerminating = fatal.FatalReasonCode() == "operation_cancellation_grace_exhausted"
						writeFatal(processlifecycle.FatalSignal{ReasonCode: fatal.FatalReasonCode(), ExitCode: 70})
						return 70
					}
					if errors.Is(publicationErr, context.Canceled) && ctx.Err() != nil {
						return 0
					}
					runner.writeStartupError(publicationErr, logger, "publish runtime")
					return 2
				}
			}
			return 0
		}
		var startupErr *httpruntime.StartupError
		if errors.As(serveErr, &startupErr) {
			if runtime.publication != nil {
				runtime.publication.cancel()
				select {
				case signal := <-forcedFatal:
					forceTerminating = true
					writeFatal(signal)
					return 70
				case <-runtime.publication.done:
				}
				var fatal interface{ FatalReasonCode() string }
				if errors.As(runtime.publication.err, &fatal) {
					forceTerminating = fatal.FatalReasonCode() == "operation_cancellation_grace_exhausted"
					shutdownDeadline = time.Now().Add(runtime.ShutdownDrainTimeout)
					writeFatal(processlifecycle.FatalSignal{ReasonCode: fatal.FatalReasonCode(), ExitCode: 70})
					return 70
				}
				if errors.Is(runtime.publication.err, context.Canceled) && ctx.Err() != nil {
					return 0
				}
				serveErr = runtime.publication.coordinator.PublicationFailure(runtime.publication.seconds, serveErr)
			}
			runner.writeStartupError(serveErr, logger, "start server listener")
			return 2
		}
		if runtime.Fatal != nil {
			runtime.Fatal("published_component_lost")
		}
		writeFatal(processlifecycle.FatalSignal{ReasonCode: "published_component_lost", ExitCode: 70})
		return 70
	}
}

func applicationTLSDiagnostic() error {
	return config.NewDiagnosticsError(config.Diagnostic{
		Path: "application", ReasonCode: "application_tls_invalid", Message: "Application TLS identity admission failed.",
	})
}

func (runner serverRunner) awaitDrain(serveDone <-chan error, timeout time.Duration, forcedFatal <-chan processlifecycle.FatalSignal) bool {
	if timeout <= 0 {
		timeout = time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-forcedFatal:
		return true
	case <-serveDone:
	case <-timer.C:
	}
	return false
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
	var admissionErr *extensions.StartupValidationError
	if errors.As(err, &admissionErr) {
		diagnostic := struct {
			Error struct {
				Code    string `json:"code"`
				Details struct {
					Items []extensions.StartupFinding `json:"items"`
				} `json:"details"`
			} `json:"error"`
		}{}
		diagnostic.Error.Code = admissionErr.StartupCode()
		diagnostic.Error.Details.Items = admissionErr.Findings
		encoded, _ := json.Marshal(diagnostic)
		_, _ = runner.stderr.Write(append(encoded, '\n'))
		return
	}
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) {
		_, _ = io.WriteString(runner.stderr, diagnosticsErr.JSON())
		_, _ = io.WriteString(runner.stderr, "\n")
		return
	}
	var postgresFailure *postgres.ConfigurationError
	if errors.As(err, &postgresFailure) {
		diagnostics := config.NewDiagnosticsError(config.Diagnostic{
			Path:       "roots.database_storage",
			ReasonCode: postgresFailure.Reason(),
			Message:    "Database connection admission failed.",
		})
		_, _ = io.WriteString(runner.stderr, diagnostics.JSON()+"\n")
		return
	}
	if errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
		// Emit only the owner's fixed diagnostic, never the wrapping cause.
		_, _ = io.WriteString(runner.stderr, database_migrations.ErrIncompatibleCryptoState.Error()+"\n")
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
	// Unmapped library errors are not an admitted diagnostic vocabulary.
	logger.Error(action)
}

func normalizeServerWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
