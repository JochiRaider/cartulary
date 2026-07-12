package app

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
	runner.serve = func(context.Context, http.Handler, httpruntime.Options) error {
		return errors.New("listener unavailable")
	}

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

func TestServerRunnerEnablesTestRoutesOnlyForExactOne(t *testing.T) {
	for _, tc := range []struct {
		value       string
		wantEnabled bool
	}{
		{value: "", wantEnabled: false},
		{value: "true", wantEnabled: false},
		{value: "1", wantEnabled: true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			runner := newServerRunner(nil, nil)
			runner.lookupEnv = func(key string) (string, bool) {
				if key == enableTestRoutesEnv {
					return tc.value, tc.value != ""
				}
				return "", false
			}
			options := runner.runtimeOptions()
			if got := len(options.HTTP.AdditionalRoutes) > 0; got != tc.wantEnabled {
				t.Fatalf("routes enabled got %v want %v", got, tc.wantEnabled)
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
