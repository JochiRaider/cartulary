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
	"sync"
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

// fatalRecorder is implemented only by the server-harness profile, which
// records its own lifecycle fatal as harness evidence (TH-HARNESS-REQ-863).
// The production profile does not implement it, so production builds are
// unchanged.
type fatalRecorder interface {
	recordFatal(lookup func(string) (string, bool), reasonCode string)
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
	publication *publicationOperation
	Handler     http.Handler
	// Close closes owned components under ctx and returns those still open
	// when ctx expired (REQ-04-146a).
	Close func(context.Context) []string
	// BeginShutdown runs when HTTP shutdown begins, with the application
	// shutdown deadline: it closes WebSockets and starts background closes.
	BeginShutdown        func(context.Context)
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
				Handler: runtime.HTTPHandler(), Close: runtime.shutdown, BeginShutdown: runtime.beginShutdown, ActivatePublication: runtime.ActivatePublication,
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
			if recorder, ok := runner.profile.(fatalRecorder); ok {
				recorder.recordFatal(runner.lookupEnv, signal.ReasonCode)
			}
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
	// REQ-04-146a: one application shutdown deadline, measured from the stop
	// request, bounds the HTTP drain, publication and every component close.
	clock := newShutdownClock(runtime.ShutdownDrainTimeout)
	stopLogged := make(chan struct{})
	stopWatch := context.AfterFunc(ctx, func() {
		defer close(stopLogged)
		clock.start()
		logger.Info("shutdown phase", slog.String("phase", "stop_requested"), slog.Int64("deadline_ms", clock.drain.Milliseconds()))
	})
	defer func() {
		// A started stop callback always finishes logging before the run returns.
		if !stopWatch() {
			<-stopLogged
		}
	}()
	beginCtx, cancelBegin := context.WithCancel(context.Background())
	defer cancelBegin()
	beginShutdown := func() {
		if runtime.BeginShutdown == nil {
			return
		}
		deadlineCtx, cancelDeadline := context.WithDeadline(beginCtx, clock.deadline())
		context.AfterFunc(beginCtx, cancelDeadline)
		runtime.BeginShutdown(deadlineCtx)
	}
	stopRequested := false
	var incomplete []string
	defer func() {
		if !stopRequested || exitCode != 0 {
			return
		}
		if len(incomplete) > 0 {
			runner.writeShutdownIncomplete(incomplete)
			logger.Warn("shutdown phase", slog.String("phase", "incomplete"), slog.Int64("elapsed_ms", clock.elapsed().Milliseconds()), slog.Any("open_components", incomplete))
			exitCode = 1
			return
		}
		logger.Info("shutdown phase", slog.String("phase", "complete"), slog.Int64("elapsed_ms", clock.elapsed().Milliseconds()))
	}()
	if runtime.Close != nil {
		defer func() {
			if forceTerminating {
				return
			}
			ordinary := stopRequested && exitCode == 0
			deadline := shutdownDeadline
			if ordinary {
				deadline = clock.deadline()
			}
			closeCtx, cancelClose := context.Background(), context.CancelFunc(func() {})
			var shutdownExpired <-chan time.Time
			if !deadline.IsZero() {
				closeCtx, cancelClose = context.WithDeadline(closeCtx, deadline)
				// shutdown returns at the deadline; the guard covers a close
				// that cannot observe its context.
				timer := time.NewTimer(time.Until(deadline) + shutdownCloseGuard)
				defer timer.Stop()
				shutdownExpired = timer.C
			}
			defer cancelClose()
			closed := make(chan []string, 1)
			go func() { closed <- runtime.Close(closeCtx) }()
			select {
			case open := <-closed:
				if ordinary {
					incomplete = append(incomplete, open...)
					logger.Info("shutdown phase", slog.String("phase", "components"), slog.Int64("elapsed_ms", clock.elapsed().Milliseconds()))
				} else if len(open) > 0 && !shutdownDeadline.IsZero() {
					exitCode = 70
				}
			case signal := <-forcedFatal:
				writeFatal(signal)
				exitCode = 70
			case <-shutdownExpired:
				if ordinary {
					incomplete = append(incomplete, "runtime")
				} else {
					exitCode = 70
				}
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
			TLS:              tlsIdentity,
			Address:          address,
			InheritedFD:      runner.profile.inheritedListenerFD(runner.lookupEnv),
			Logger:           logger,
			ShutdownTimeout:  runtime.ShutdownDrainTimeout,
			ShutdownDeadline: clock.httpDrainDeadline,
			OnReady:          runtime.ActivatePublication,
			OnShutdown:       beginShutdown,
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
			// One deadline: a stop request that already started it is never
			// restarted by the fatal lifecycle (REQ-04-146, REQ-04-146a).
			shutdownDeadline = clock.deadline()
			forceTerminating = runner.awaitDrain(serveDone, time.Until(shutdownDeadline), forcedFatal)
		}
		writeFatal(fatal)
		return 70
	case publicationErr := <-publicationFailures:
		cancelServe()
		var fatal interface{ FatalReasonCode() string }
		if errors.As(publicationErr, &fatal) {
			forceTerminating = fatal.FatalReasonCode() == "operation_cancellation_grace_exhausted"
			if !forceTerminating {
				shutdownDeadline = clock.deadline()
				forceTerminating = runner.awaitDrain(serveDone, time.Until(shutdownDeadline), forcedFatal)
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
			stopRequested = true
			return 0
		}
		runner.writeStartupError(publicationErr, logger, "publish runtime")
		return 2
	case serveErr := <-serveDone:
		stopRequested = ctx.Err() != nil
		var serveStartupErr *httpruntime.StartupError
		if stopRequested {
			var drainCut *httpruntime.DrainDeadlineError
			cut := errors.As(serveErr, &drainCut)
			logger.Info("shutdown phase", slog.String("phase", "http_drain"), slog.Bool("cancelled_in_flight", cut), slog.Int64("elapsed_ms", clock.elapsed().Milliseconds()))
			switch {
			case cut:
				// REQ-04-146a: work still running at the close reserve boundary
				// is cancelled; that completes the HTTP phase.
				serveErr = nil
			case serveErr != nil && !errors.As(serveErr, &serveStartupErr):
				// Any other stop-initiated drain failure is incomplete ordinary
				// shutdown, never published_component_lost.
				incomplete = append(incomplete, "http_server")
				serveErr = nil
			}
		}
		if serveErr == nil {
			if runtime.publication != nil {
				runtime.publication.cancel()
				var publicationExpired <-chan time.Time
				if stopRequested {
					timer := time.NewTimer(time.Until(clock.deadline()))
					defer timer.Stop()
					publicationExpired = timer.C
				}
				select {
				case signal := <-forcedFatal:
					forceTerminating = true
					writeFatal(signal)
					return 70
				case <-runtime.publication.done:
				case <-publicationExpired:
					incomplete = append(incomplete, "publication")
					return 0
				}
				if stopRequested {
					logger.Info("shutdown phase", slog.String("phase", "publication"), slog.Int64("elapsed_ms", clock.elapsed().Milliseconds()))
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

// shutdownCloseGuard bounds how long the runner waits past the deadline for a
// component close that cannot observe its context.
const shutdownCloseGuard = 250 * time.Millisecond

// shutdownClock measures the REQ-04-146a application shutdown deadline from
// the first stop request; later stop requests never restart it.
type shutdownClock struct {
	once    sync.Once
	drain   time.Duration
	started time.Time
	until   time.Time
}

func newShutdownClock(drain time.Duration) *shutdownClock {
	if drain <= 0 {
		drain = httpruntime.DefaultShutdownTimeout
	}
	return &shutdownClock{drain: drain}
}

// shutdownCloseReserve is the part of the deadline kept for component closes:
// min(25% of the deadline, 5 s) (REQ-04-146a).
func shutdownCloseReserve(drain time.Duration) time.Duration {
	return min(drain/4, 5*time.Second)
}

// httpDrainDeadline ends the HTTP drain early enough to leave the close
// reserve for component closes.
func (c *shutdownClock) httpDrainDeadline() time.Time {
	return c.deadline().Add(-shutdownCloseReserve(c.drain))
}

func (c *shutdownClock) start() {
	c.once.Do(func() {
		c.started = time.Now()
		c.until = c.started.Add(c.drain)
	})
}

func (c *shutdownClock) deadline() time.Time {
	c.start()
	return c.until
}

func (c *shutdownClock) elapsed() time.Duration {
	c.start()
	return time.Since(c.started)
}

// writeShutdownIncomplete emits the one secret-safe REQ-04-146a diagnostic
// naming the components still open at the shutdown deadline.
func (runner serverRunner) writeShutdownIncomplete(components []string) {
	diagnostic := struct {
		Code           string   `json:"code"`
		OpenComponents []string `json:"open_components"`
	}{Code: "shutdown_incomplete", OpenComponents: components}
	encoded, _ := json.Marshal(diagnostic)
	_, _ = runner.stderr.Write(append(encoded, '\n'))
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
	if errors.Is(err, cryptography.ErrExecutionPolicy) {
		// Emit only the owner's fixed diagnostic, never the wrapping cause.
		_, _ = io.WriteString(runner.stderr, cryptography.ErrExecutionPolicy.Error()+"\n")
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
