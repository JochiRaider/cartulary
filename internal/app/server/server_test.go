package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestServerRunnerWritesDiagnosticsAndReturnsFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	diagnostics := config.NewDiagnosticsError(config.Diagnostic{
		Path:       "application.public_origin",
		ReasonCode: "invalid_origin",
		Message:    "invalid origin",
	})
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = func() (config.Config, error) { return config.Config{}, diagnostics }

	if exitCode := runner.run(context.Background()); exitCode != 1 {
		t.Fatalf("exit code got %d want 1", exitCode)
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
	closed := false
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = func() (config.Config, error) { return config.Config{}, nil }
	runner.buildRuntime = func(context.Context, config.Config, Options) (http.Handler, func(), error) {
		return http.NotFoundHandler(), func() { closed = true }, nil
	}
	runner.profile = failingServerProfile{}

	if exitCode := runner.run(context.Background()); exitCode != 1 {
		t.Fatalf("exit code got %d want 1", exitCode)
	}
	if !closed {
		t.Fatal("runtime was not closed")
	}
	if !strings.Contains(stdout.String(), "server exited") {
		t.Fatalf("missing server exit diagnostic: %q", stdout.String())
	}
}

func TestServerRunnerMapsRuntimeSetupFailure(t *testing.T) {
	var stdout bytes.Buffer
	runner := newServerRunner(&stdout, io.Discard)
	runner.loadConfig = func() (config.Config, error) { return config.Config{}, nil }
	runner.buildRuntime = func(context.Context, config.Config, Options) (http.Handler, func(), error) {
		return nil, nil, errors.New("runtime unavailable")
	}
	if exitCode := runner.run(context.Background()); exitCode != 1 {
		t.Fatalf("exit code got %d want 1", exitCode)
	}
	if !strings.Contains(stdout.String(), "setup runtime") {
		t.Fatalf("missing runtime setup diagnostic: %q", stdout.String())
	}
}

func TestServerRunnerWritesMigrationRemediationToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = func() (config.Config, error) { return config.Config{}, nil }
	runner.buildRuntime = func(context.Context, config.Config, Options) (http.Handler, func(), error) {
		return nil, nil, &postgres.MigrationRemediationError{
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
	if exitCode := runner.run(context.Background()); exitCode != 1 {
		t.Fatalf("exit code got %d want 1", exitCode)
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
	runner.loadConfig = func() (config.Config, error) { return config.Config{}, nil }
	runner.buildRuntime = func(context.Context, config.Config, Options) (http.Handler, func(), error) {
		cancel()
		return nil, nil, context.Canceled
	}
	if exitCode := runner.run(ctx); exitCode != 0 {
		t.Fatalf("exit code got %d want 0", exitCode)
	}
}

func TestServerRunnerCancellationBeforeStartupReturnsSuccess(t *testing.T) {
	runner := newServerRunner(nil, nil)
	runner.loadConfig = func() (config.Config, error) {
		t.Fatal("cancelled runner loaded configuration")
		return config.Config{}, nil
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

func TestServerRunnerRejectsHarnessEnvironmentBeforeConfigLoad(t *testing.T) {
	for _, key := range harnessOnlyServerEnv {
		t.Run(key, func(t *testing.T) {
			var stderr bytes.Buffer
			runner := newServerRunner(io.Discard, &stderr)
			runner.lookupEnv = func(name string) (string, bool) {
				return "", name == key
			}
			runner.loadConfig = func() (config.Config, error) {
				t.Fatal("production runner loaded config with a harness-only key")
				return config.Config{}, nil
			}
			if exitCode := runner.run(context.Background()); exitCode != 1 {
				t.Fatalf("exit code got %d want 1", exitCode)
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
	runner.loadConfig = func() (config.Config, error) {
		return config.Config{}, config.NewDiagnosticsError(config.Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "unsupported_config_schema_id",
			Message:    "unsupported schema",
		})
	}
	if exitCode := runner.run(context.Background()); exitCode != 1 {
		t.Fatalf("exit code got %d want 1", exitCode)
	}
}
