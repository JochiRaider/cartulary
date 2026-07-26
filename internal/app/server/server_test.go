package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

func testRuntime(handler http.Handler, closeRuntime func()) serverRuntime {
	return serverRuntime{
		Handler:              handler,
		Close:                closeRuntime,
		ActivatePublication:  func() error { return nil },
		FatalEvents:          make(chan processlifecycle.FatalSignal),
		Fatal:                func(string) bool { return true },
		ShutdownDrainSeconds: 1,
	}
}

func TestServerRunnerWritesDiagnosticsAndReturnsFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	diagnostics := config.NewDiagnosticsError(config.Diagnostic{
		Path:       "application.public_origin",
		ReasonCode: "invalid_origin",
		Message:    "invalid origin",
	})
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, diagnostics }

	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if got, want := stderr.String(), diagnostics.JSON()+"\n"; got != want {
		t.Fatalf("stderr got %q want %q", got, want)
	}
	if stdout.Len() != 0 {
		t.Fatalf("diagnostics emitted stdout: %q", stdout.String())
	}
}

func TestServerRunnerClosesRuntimeAndMapsServeFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	closed := false
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return testRuntime(http.NotFoundHandler(), func() { closed = true }), nil
	}
	runner.profile = failingServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 70 {
		t.Fatalf("exit code got %d want 70", exitCode)
	}
	if !closed {
		t.Fatal("runtime was not closed")
	}
	if !strings.Contains(stderr.String(), `"reason_code":"published_component_lost"`) {
		t.Fatalf("missing component loss diagnostic: %q", stderr.String())
	}
}

func TestServerRunnerLogsPublicContractAdmissionWithoutRequestData(t *testing.T) {
	var stdout bytes.Buffer
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		runtime := testRuntime(http.NotFoundHandler(), nil)
		runtime.PublicHTTP = httpapi.RouteDiagnostics{
			CanonicalSHA256:         "sha256-fixture",
			DocumentVersion:         "2.0.0",
			SupportedOperationCount: 12,
			ActiveOperationCount:    9,
			ClaimedProfiles:         []string{"enterprise_authentication"},
		}
		return runtime, nil
	}
	runner.profile = readyServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 0 {
		t.Fatalf("exit code got %d want 0", exitCode)
	}
	logged := stdout.String()
	for _, expected := range []string{
		"public HTTP contract admitted",
		"openapi_version=2.0.0",
		"openapi_sha256=sha256-fixture",
		"supported_operation_count=12",
		"active_operation_count=9",
		"enterprise_authentication",
	} {
		if !strings.Contains(logged, expected) {
			t.Fatalf("missing %q in admission log: %q", expected, logged)
		}
	}
	for _, forbidden := range []string{"request_body", "authorization", "cookie", "secret"} {
		if strings.Contains(strings.ToLower(logged), forbidden) {
			t.Fatalf("admission log contains forbidden request data key %q: %q", forbidden, logged)
		}
	}
}

func TestServerRunnerMapsListenerStartupFailureToExitTwo(t *testing.T) {
	var stdout bytes.Buffer
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return testRuntime(http.NotFoundHandler(), nil), nil
	}
	runner.profile = startupFailingServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if !strings.Contains(stdout.String(), "start server listener") {
		t.Fatalf("missing listener startup diagnostic: %q", stdout.String())
	}
}

func TestServerRunnerFatalLossClosesAdmissionDrainsAndExitsSeventy(t *testing.T) {
	var stderr bytes.Buffer
	fatalEvents := make(chan processlifecycle.FatalSignal, 1)
	fatalEvents <- processlifecycle.FatalSignal{ReasonCode: "published_component_lost", ExitCode: 70}
	activated := false
	closed := false
	runner := newServerRunner(io.Discard, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		runtime := testRuntime(http.NotFoundHandler(), func() { closed = true })
		runtime.ActivatePublication = func() error { activated = true; return nil }
		runtime.FatalEvents = fatalEvents
		return runtime, nil
	}
	runner.profile = drainingServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 70 {
		t.Fatalf("exit code got %d want 70", exitCode)
	}
	if !activated || !closed {
		t.Fatalf("publication/close = %t/%t", activated, closed)
	}
	if got := stderr.String(); got != "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"published_component_lost\"}\n" {
		t.Fatalf("fatal diagnostic = %q", got)
	}
}

func TestServerRunnerPublicationActivationFailureIsStartupFailure(t *testing.T) {
	runner := newServerRunner(io.Discard, io.Discard)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		runtime := testRuntime(http.NotFoundHandler(), nil)
		runtime.ActivatePublication = func() error { return errors.New("publication rejected") }
		return runtime, nil
	}
	runner.profile = readyServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
}

func TestServerRunnerMapsRuntimeSetupFailure(t *testing.T) {
	var stdout bytes.Buffer
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, errors.New("runtime unavailable")
	}
	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if !strings.Contains(stdout.String(), "setup runtime") {
		t.Fatalf("missing runtime setup diagnostic: %q", stdout.String())
	}
}

func TestServerRunnerMapsFatalRuntimeSetupFailureToExitSeventy(t *testing.T) {
	var stderr bytes.Buffer
	runner := newServerRunner(io.Discard, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, &stagedobjects.FatalIntegrityError{Cause: errors.New("private contradiction detail")}
	}
	if exitCode := runner.run(context.Background()); exitCode != 70 {
		t.Fatalf("exit code got %d want 70", exitCode)
	}
	if got := stderr.String(); got != "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"staged_object_publication_mismatch\"}\n" {
		t.Fatalf("fatal startup diagnostic = %q", got)
	}
}

func TestServerRunnerMapsConfirmedLeaseLossToExitSeventy(t *testing.T) {
	var stderr bytes.Buffer
	runner := newServerRunner(io.Discard, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, processlease.ErrLeaseLost
	}
	if exitCode := runner.run(context.Background()); exitCode != 70 {
		t.Fatalf("exit code got %d want 70", exitCode)
	}
	if !strings.Contains(stderr.String(), `"reason_code":"application_process_lease_lost"`) {
		t.Fatalf("missing lease-loss fatal diagnostic: %q", stderr.String())
	}
}

func TestServerRunnerWritesMigrationRemediationToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, &postgres.MigrationRemediationError{
			Report: postgres.MigrationRemediationReport{
				SchemaID:    "cartulary.migration_remediation_report.v1",
				Boundary:    "prod_ddl_rebaseline_v1",
				FromVersion: 40,
				ToVersion:   33,
				Findings: []postgres.MigrationRemediationFinding{
					{
						Field:           "schema_migration_lineage",
						ReasonCode:      "historical_migration_lineage",
						RemediationHint: "reset",
					},
				},
			},
		}
	}
	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("migration remediation emitted stdout: %q", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, `"schema_id":"cartulary.migration_remediation_report.v1"`) ||
		!strings.Contains(got, `"reason_code":"historical_migration_lineage"`) {
		t.Fatalf("missing remediation JSON in stderr: %q", got)
	}
}

func TestServerRunnerCancellationDuringRuntimeSetupReturnsSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := newServerRunner(nil, nil)
	runner.loadConfig = func() (configassembly.Loaded, error) { return configassembly.Loaded{}, nil }
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		cancel()
		return serverRuntime{}, context.Canceled
	}
	if exitCode := runner.run(ctx); exitCode != 0 {
		t.Fatalf("exit code got %d want 0", exitCode)
	}
}

func TestServerRunnerCancellationBeforeStartupReturnsSuccess(t *testing.T) {
	runner := newServerRunner(nil, nil)
	runner.loadConfig = func() (configassembly.Loaded, error) {
		t.Fatal("cancelled runner loaded configuration")
		return configassembly.Loaded{}, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if exitCode := runner.run(ctx); exitCode != 0 {
		t.Fatalf("exit code got %d want 0", exitCode)
	}
}

type failingServerProfile struct{}

func (failingServerProfile) validateEnvironment(func(string) (string, bool)) error {
	return nil
}

func (failingServerProfile) runtimeOptions(func(string) (string, bool)) Options {
	return Options{}
}

func (failingServerProfile) inheritedListenerFD(func(string) (string, bool)) string {
	return ""
}

func (failingServerProfile) serve(context.Context, http.Handler, httpruntime.Options) error {
	return errors.New("listener unavailable")
}

type startupFailingServerProfile struct{ failingServerProfile }

func (startupFailingServerProfile) serve(context.Context, http.Handler, httpruntime.Options) error {
	return &httpruntime.StartupError{Err: errors.New("address unavailable")}
}

type drainingServerProfile struct{ failingServerProfile }

func (drainingServerProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	<-ctx.Done()
	return nil
}

type readyServerProfile struct{ failingServerProfile }

func (readyServerProfile) serve(_ context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	return nil
}

func TestServerRunnerRejectsHarnessEnvironmentBeforeConfigLoad(t *testing.T) {
	for _, key := range harnessOnlyServerEnv {
		t.Run(key, func(t *testing.T) {
			var stderr bytes.Buffer
			runner := newServerRunner(io.Discard, &stderr)
			runner.lookupEnv = func(name string) (string, bool) {
				return "", name == key
			}
			runner.loadConfig = func() (configassembly.Loaded, error) {
				t.Fatal("production runner loaded config with a harness-only key")
				return configassembly.Loaded{}, nil
			}
			if exitCode := runner.run(context.Background()); exitCode != 2 {
				t.Fatalf("exit code got %d want 2", exitCode)
			}
			if !strings.Contains(stderr.String(), "harness_profile_required") {
				t.Fatalf("missing profile diagnostic: %q", stderr.String())
			}
		})
	}
}

type failingServerWriter struct{}

func (failingServerWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestServerRunnerFailingDiagnosticsWriterDoesNotPanicOrSucceed(t *testing.T) {
	runner := newServerRunner(nil, failingServerWriter{})
	runner.loadConfig = func() (configassembly.Loaded, error) {
		return configassembly.Loaded{}, config.NewDiagnosticsError(config.Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "unsupported_config_schema_id",
			Message:    "unsupported schema",
		})
	}
	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
}
