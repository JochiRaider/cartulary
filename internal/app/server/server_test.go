package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	tlstest "github.com/JochiRaider/cartulary/internal/testutil/tlstest/transport"
)

func testRuntime(handler http.Handler, closeRuntime func()) serverRuntime {
	return serverRuntime{
		Handler:              handler,
		Close:                closeRuntime,
		ActivatePublication:  func() error { return nil },
		FatalEvents:          make(chan processlifecycle.FatalSignal),
		Fatal:                func(string) bool { return true },
		ShutdownDrainTimeout: time.Second,
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
	runner.loadConfig = serverTestConfiguration(t)
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
	runner.loadConfig = serverTestConfiguration(t)
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
	runner.loadConfig = serverTestConfiguration(t)
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
	t.Run("grace escalates an existing drain without replacing its reason", func(t *testing.T) {
		lifecycle := processlifecycle.New()
		if err := lifecycle.Publish(); err != nil {
			t.Fatal(err)
		}
		lifecycle.Fatal("published_component_lost")
		if lifecycle.Fatal("operation_cancellation_grace_exhausted") || lifecycle.State() != processlifecycle.StateTerminating || lifecycle.AdmissionOpen() || lifecycle.FatalReason() != "published_component_lost" {
			t.Fatal("grace did not force the selected lifecycle to termination")
		}
		<-lifecycle.FatalEvents()
		select {
		case <-lifecycle.FatalEvents():
			t.Fatal("duplicate fatal event")
		default:
		}
	})
	t.Run("grace exhaustion interrupts runtime construction", func(t *testing.T) {
		var stderr bytes.Buffer
		release, finished := make(chan struct{}), make(chan struct{})
		runner := newServerRunner(io.Discard, &stderr)
		runner.loadConfig = serverTestConfiguration(t)
		runner.buildRuntime = func(_ context.Context, _ configassembly.Loaded, options Options) (serverRuntime, error) {
			defer close(finished)
			runtime := &Runtime{lifecycle: processlifecycle.New(), immediateFatal: options.immediateFatal}
			runtime.fatalIntegrity(extensiondeadline.ErrGraceExhausted)
			<-release
			return serverRuntime{}, extensiondeadline.ErrGraceExhausted
		}
		done := make(chan int, 1)
		go func() { done <- runner.run(context.Background()) }()
		select {
		case code := <-done:
			if code != 70 {
				t.Errorf("forced startup exit = %d", code)
			}
		case <-time.After(time.Second):
			close(release)
			<-finished
			<-done
			t.Fatal("startup waited for the non-cooperative operation")
		}
		close(release)
		<-finished
		if got := stderr.String(); got != "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"operation_cancellation_grace_exhausted\"}\n" {
			t.Fatalf("forced startup diagnostic = %q", got)
		}
	})
	var stderr bytes.Buffer
	fatalEvents := make(chan processlifecycle.FatalSignal, 1)
	fatalEvents <- processlifecycle.FatalSignal{ReasonCode: "published_component_lost", ExitCode: 70}
	activated := false
	closed := false
	runner := newServerRunner(io.Discard, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
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
	runner.loadConfig = serverTestConfiguration(t)
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
	t.Run("execution policy uses only the owner diagnostic", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runner := newServerRunner(&stdout, &stderr)
		runner.loadConfig = serverTestConfiguration(t)
		runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
			return serverRuntime{}, errors.Join(errors.New("private SQL /private/root credential"), cryptography.ErrExecutionPolicy)
		}
		if code := runner.run(context.Background()); code != 2 {
			t.Fatalf("exit code got %d want 2", code)
		}
		if stdout.Len() != 0 || stderr.String() != cryptography.ErrExecutionPolicy.Error()+"\n" {
			t.Fatal("execution policy rejection omitted its safe diagnostic or disclosed an upstream cause")
		}
	})
	t.Run("incompatible crypto state uses only the owner diagnostic", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runner := newServerRunner(&stdout, &stderr)
		runner.loadConfig = serverTestConfiguration(t)
		runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
			return serverRuntime{}, errors.Join(errors.New("private SQL /private/root credential"), database_migrations.ErrIncompatibleCryptoState)
		}
		if code := runner.run(context.Background()); code != 2 {
			t.Fatalf("exit code got %d want 2", code)
		}
		if stdout.Len() != 0 || stderr.String() != database_migrations.ErrIncompatibleCryptoState.Error()+"\n" {
			t.Fatal("crypto state rejection omitted its safe diagnostic or disclosed an upstream cause")
		}
	})
	t.Run("typed PostgreSQL failure retains only its safe classification", func(t *testing.T) {
		_, admissionErr := postgres.ResolveSettings(postgres.Binding{}, postgres.PurposeRuntime, map[string]string{})
		if admissionErr == nil {
			t.Fatal("invalid binding was accepted")
		}
		var stdout, stderr bytes.Buffer
		runner := newServerRunner(&stdout, &stderr)
		runner.loadConfig = serverTestConfiguration(t)
		runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
			return serverRuntime{}, errors.Join(errors.New("private SQL /private/root credential"), admissionErr)
		}
		if code := runner.run(context.Background()); code != 2 {
			t.Fatalf("exit code got %d want 2", code)
		}
		var diagnostic struct {
			Error struct {
				Code    string `json:"code"`
				Details struct {
					Items []config.Diagnostic `json:"items"`
				} `json:"details"`
			} `json:"error"`
		}
		if err := json.Unmarshal(stderr.Bytes(), &diagnostic); err != nil || diagnostic.Error.Code != config.InvalidDeploymentConfigCode || len(diagnostic.Error.Details.Items) != 1 {
			t.Fatalf("missing safe deployment diagnostic: %s", stderr.String())
		}
		item := diagnostic.Error.Details.Items[0]
		if item.Path != "roots.database_storage" || item.ReasonCode != postgres.ReasonBindingInvalid || item.Message == "" {
			t.Fatalf("unexpected database diagnostic: %+v", item)
		}
		if stdout.Len() != 0 || strings.Contains(stderr.String(), "private") || strings.Contains(stderr.String(), "credential") || strings.Contains(stderr.String(), "SQL") {
			t.Fatal("database startup diagnostic disclosed an upstream cause")
		}
	})
	var stdout bytes.Buffer
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = serverTestConfiguration(t)
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, errors.New("private SQL /private/root credential")
	}
	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if strings.Contains(stdout.String(), "private") || strings.Contains(stdout.String(), "credential") || strings.Contains(stdout.String(), "SQL") {
		t.Fatal("unmapped startup exception escaped")
	}
	if !strings.Contains(stdout.String(), "setup runtime") {
		t.Fatalf("missing runtime setup diagnostic: %q", stdout.String())
	}
}

func TestServerRunnerMapsFatalRuntimeSetupFailureToExitSeventy(t *testing.T) {
	var stderr bytes.Buffer
	runner := newServerRunner(io.Discard, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
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

func TestServerRunnerMapsTypedApplicationLeaseLossesToExitSeventy(t *testing.T) {
	for name, tc := range map[string]struct {
		err        error
		reasonCode string
	}{
		"application process": {
			err:        processlease.ErrApplicationProcessLeaseLost,
			reasonCode: "application_process_lease_lost",
		},
		"recovery serving": {
			err:        processlease.ErrRecoveryServingLeaseLost,
			reasonCode: "recovery_serving_lease_lost",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var stderr bytes.Buffer
			runner := newServerRunner(io.Discard, &stderr)
			runner.loadConfig = serverTestConfiguration(t)
			runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
				return serverRuntime{}, tc.err
			}
			if exitCode := runner.run(context.Background()); exitCode != 70 {
				t.Fatalf("exit code got %d want 70", exitCode)
			}
			if !strings.Contains(stderr.String(), `"reason_code":"`+tc.reasonCode+`"`) {
				t.Fatalf("missing lease-loss fatal diagnostic: %q", stderr.String())
			}
		})
	}
}

func TestServerRunnerMapsTypedLeaseAdmissionFailuresToExitTwo(t *testing.T) {
	for name, admissionErr := range map[string]error{
		"application process": processlease.ErrApplicationProcessActive,
		"recovery serving":    processlease.ErrRecoveryServingLeaseActive,
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			runner := newServerRunner(&stdout, &stderr)
			runner.loadConfig = serverTestConfiguration(t)
			runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
				coordinator, err := extensions.NewGeneratedCoordinator()
				if err != nil {
					t.Fatal(err)
				}
				if errors.Is(admissionErr, processlease.ErrApplicationProcessActive) {
					return serverRuntime{}, coordinator.ApplicationLeaseAdmissionFailure(30, admissionErr)
				}
				return serverRuntime{}, coordinator.RecoveryLeaseAdmissionFailure(admissionErr)
			}
			if exitCode := runner.run(context.Background()); exitCode != 2 {
				t.Fatalf("exit code got %d want 2", exitCode)
			}
			if name == "recovery serving" && strings.Contains(stderr.String(), "invalid_deployment_config") {
				t.Fatal("Recovery lease misclassified as configuration")
			}
			if name == "application process" && !strings.Contains(stderr.String(), `"timeout_seconds":30`) {
				t.Fatal("lease timeout missing")
			}
			if !strings.Contains(stderr.String(), admissionErr.Error()) {
				t.Fatalf("missing lease-admission reason %q: %q", admissionErr, stderr.String())
			}
		})
	}
}

func TestServerRunnerWritesMigrationRemediationToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
	remediation := fakeServerRemediationFailure{
		report: `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"prod_ddl_rebaseline_v4","from_version":30,"to_version":29,"findings":[{"field":"schema_migration_lineage","reason_code":"historical_migration_lineage","remediation_hint":"Destroy and recreate this database, then apply the Production DDL Rebaseline v4 catalog from version 1."}]}`,
	}
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, remediation
	}
	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("migration remediation emitted stdout: %q", stdout.String())
	}
	if got, want := stderr.String(), remediation.RemediationReportJSON()+"\n"; got != want {
		t.Fatalf("unexpected remediation JSON: got %q want %q", got, want)
	}
}

type fakeServerRemediationFailure struct {
	report string
}

func (failure fakeServerRemediationFailure) Error() string {
	return "private remediation transport"
}

func (failure fakeServerRemediationFailure) ReasonCode() string {
	return "historical_migration_lineage"
}

func (failure fakeServerRemediationFailure) RemediationReportJSON() string {
	return failure.report
}

type fakeServerMigrationFailure struct {
	reason string
}

func (failure fakeServerMigrationFailure) Error() string {
	return "postgres://secret@private-host SELECT sensitive bind=password /private/migrations"
}

func (failure fakeServerMigrationFailure) ReasonCode() string {
	return failure.reason
}

func TestServerRunnerWritesMigrationFailureDiagnostic(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		return serverRuntime{}, fakeServerMigrationFailure{reason: "schema_migration_execution_failed"}
	}

	if exitCode := runner.run(context.Background()); exitCode != 2 {
		t.Fatalf("exit code got %d want 2", exitCode)
	}
	want := config.NewDiagnosticsError(config.Diagnostic{
		Path:       "database.schema_version",
		ReasonCode: "schema_migration_execution_failed",
		Message:    "Database migration validation failed.",
	}).JSON() + "\n"
	if got := stderr.String(); got != want {
		t.Fatalf("unexpected migration diagnostic: got %q want %q", got, want)
	}
	if stdout.Len() != 0 {
		t.Fatalf("migration failure emitted stdout: %q", stdout.String())
	}
}

func TestServerRunnerCancellationDuringRuntimeSetupReturnsSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := newServerRunner(nil, nil)
	runner.loadConfig = serverTestConfiguration(t)
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

func TestSavedGraphAdmissionStartupDiagnostic(t *testing.T) {
	for _, finding := range []extensions.StartupFinding{
		{Path: "$", ReasonCode: "extension_admission_validation_failed", Message: "Extension admission validation failed.", Details: map[string]any{"profile_id": "network_flow_activity", "phase": "profile_preflight", "algorithm_id": "network_flow_activity.incident_portability_cutover_v1", "timed_out": false, "timeout_seconds": int64(60)}},
		{Path: "$", ReasonCode: "extension_state_initialization_failed", Message: "Extension state initialization failed.", Details: map[string]any{"profile_id": "network_flow_activity", "algorithm_id": "network_flow_activity.initialize_v1"}},
		{Path: "$", ReasonCode: "extension_state_initialization_timeout", Message: "Extension state initialization exceeded its timeout.", Details: map[string]any{"profile_id": "network_flow_activity", "timeout_seconds": int64(600)}},
	} {
		t.Run(finding.ReasonCode, func(t *testing.T) {
			var stderr bytes.Buffer
			runner := newServerRunner(io.Discard, &stderr)
			runner.loadConfig = serverTestConfiguration(t)
			runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
				return serverRuntime{}, &extensions.StartupValidationError{Findings: []extensions.StartupFinding{finding}}
			}
			if code := runner.run(context.Background()); code != 2 {
				t.Fatalf("exit=%d", code)
			}
			var diagnostic struct {
				Error struct {
					Code    string `json:"code"`
					Details struct {
						Items []extensions.StartupFinding `json:"items"`
					} `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(stderr.Bytes(), &diagnostic); err != nil || diagnostic.Error.Code != "invalid_deployment_config" || len(diagnostic.Error.Details.Items) != 1 {
				t.Fatalf("startup diagnostic=%s err=%v", stderr.String(), err)
			}
			got, _ := json.Marshal(diagnostic.Error.Details.Items[0])
			want, _ := json.Marshal(finding)
			if !bytes.Equal(got, want) {
				t.Fatalf("startup finding=%s want=%s", got, want)
			}
		})
	}
}

func serverTestConfiguration(t testing.TB) func() (configassembly.Loaded, error) {
	t.Helper()
	identity := tlstest.NewServer(t, "localhost", "127.0.0.1")
	loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, map[string]string{
		"CARTULARY__APPLICATION__TLS_CERTIFICATE_PATH": identity.CertificatePath,
		"CARTULARY__APPLICATION__TLS_PRIVATE_KEY_PATH": identity.PrivateKeyPath,
	})
	return func() (configassembly.Loaded, error) { return loaded, nil }
}

func TestServerRunnerRejectsTLSBeforeRuntime(t *testing.T) {
	for _, name := range []string{"missing", "wrong host", "missing key"} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			identity := tlstest.NewServer(t, "localhost")
			overlays := map[string]string{
				"CARTULARY__APPLICATION__TLS_CERTIFICATE_PATH": identity.CertificatePath,
				"CARTULARY__APPLICATION__TLS_PRIVATE_KEY_PATH": identity.PrivateKeyPath,
			}
			switch name {
			case "missing":
				overlays["CARTULARY__APPLICATION__TLS_CERTIFICATE_PATH"] = "/private/missing.crt"
			case "wrong host":
				overlays["CARTULARY__APPLICATION__PUBLIC_ORIGIN"] = "https://different.example"
			case "missing key":
				overlays["CARTULARY__APPLICATION__TLS_PRIVATE_KEY_PATH"] = "/private/missing.key"
			}
			loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, overlays)
			runner := newServerRunner(io.Discard, &output)
			runner.loadConfig = func() (configassembly.Loaded, error) { return loaded, nil }
			runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
				t.Fatal("TLS rejection acquired application resources")
				return serverRuntime{}, nil
			}
			if code := runner.run(context.Background()); code != 2 {
				t.Fatalf("exit = %d", code)
			}
			if !strings.Contains(output.String(), "application_tls_invalid") || strings.Contains(output.String(), "/private/") {
				t.Fatalf("unsafe or absent TLS diagnostic: %s", output.String())
			}
		})
	}
}
