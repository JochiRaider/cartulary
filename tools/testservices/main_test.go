package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

func TestRunPassThroughModeStartsNoServices(t *testing.T) {
	receivedEnv := map[string]string{}
	deps := dependencies{
		startPostgres: func(context.Context) (postgresService, error) {
			t.Fatal("startPostgres must not run in pass-through mode")
			return postgresService{}, nil
		},
		startMinIO: func(context.Context) (minioService, error) {
			t.Fatal("startMinIO must not run in pass-through mode")
			return minioService{}, nil
		},
		startChild: func(argv []string, env map[string]string) (childProcess, error) {
			receivedEnv = cloneEnv(env)
			return fakeChild{}, nil
		},
		createTemplate: func(context.Context, string, string) error { return nil },
		recordEvent:    func(map[string]string, suiteservices.Event) {},
		refreshSummary: func(map[string]string) {},
		suiteID: func() (string, error) {
			return "unused", nil
		},
		notifySignals: func(chan<- os.Signal, ...os.Signal) {},
		stopSignals:   func(chan<- os.Signal) {},
	}

	env := map[string]string{
		suiteservices.ActiveEnv:  "1",
		"CARTULARY_SENTINEL_VAR": "preserved",
	}

	status := run([]string{"run", "--", "ignored"}, env, deps)
	if status != 0 {
		t.Fatalf("unexpected exit status: got %d want 0", status)
	}
	if receivedEnv["CARTULARY_SENTINEL_VAR"] != "preserved" {
		t.Fatalf("expected child to inherit existing environment, got %#v", receivedEnv)
	}
}

func TestRunCleansUpOwnedServicesOnChildFailureAndPropagatesStatus(t *testing.T) {
	postgresClosed := 0
	minioClosed := 0
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close: func(context.Context) error {
				postgresClosed++
				return nil
			},
		}, nil
	}
	deps.startMinIO = func(context.Context) (minioService, error) {
		return minioService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "minio-access",
			secretKey: "minio-secret",
			close: func(context.Context) error {
				minioClosed++
				return nil
			},
		}, nil
	}
	deps.startChild = func(argv []string, env map[string]string) (childProcess, error) {
		return startChildProcess([]string{"bash", "-lc", "exit 7"}, env)
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 7 {
		t.Fatalf("unexpected exit status: got %d want 7", status)
	}
	if postgresClosed != 1 {
		t.Fatalf("expected one postgres cleanup, got %d", postgresClosed)
	}
	if minioClosed != 1 {
		t.Fatalf("expected one minio cleanup, got %d", minioClosed)
	}
	if loadScope(t, deps).Failure != nil {
		t.Fatal("non-zero child exit after a successful start must not populate startup failure summary")
	}

	scope := loadScope(t, deps)
	if scope.Cleanup.Status != "failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
	if scope.Cleanup.ChildExitStatus == nil || *scope.Cleanup.ChildExitStatus != 7 {
		t.Fatalf("unexpected child exit status: got %#v", scope.Cleanup)
	}
}

func TestRunRedactsCredentialsInDiagnostics(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:supersecret@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:supersecret@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context) (minioService, error) {
		return minioService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "access-secret",
			secretKey: "minio-secret",
			secure:    false,
			close:     func(context.Context) error { return nil },
		}, nil
	}
	deps.startChild = func(argv []string, env map[string]string) (childProcess, error) {
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected exit status: got %d want 0", status)
	}

	summaryPath := filepath.Join(
		deps.resultsDir,
		deps.env["CARTULARY_TEST_RUN_ID"],
		"_shared",
		"test-services",
		"suite-redaction",
		"service-scope.json",
	)
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read diagnostics summary: %v", err)
	}

	scope := decodeScope(t, raw)
	if scope.Failure != nil {
		t.Fatalf("successful run must omit failure summary, got %#v", scope.Failure)
	}

	for _, secret := range []string{"supersecret", "access-secret", "minio-secret", "postgres://"} {
		if string(raw) == "" {
			t.Fatal("expected diagnostics summary output")
		}
		if contains := string(raw); contains != "" && stringContains(contains, secret) {
			t.Fatalf("diagnostics summary must redact %q, got %s", secret, raw)
		}
	}
}

func TestRunRecordsMinIOStartupFailureWithStructuredSummary(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context) (minioService, error) {
		return minioService{}, &testcontainersx.StartFailure{
			Operation:             "start",
			Service:               "minio testcontainer",
			Image:                 "minio/minio:RELEASE.2025-09-07T16-13-09Z",
			DockerEndpoint:        "unix:///var/run/docker.sock",
			AttemptsStarted:       1,
			MaxAttempts:           2,
			Retryable:             true,
			RetryBlockedByContext: true,
			Cause:                 errors.New("docker.sock connection refused password=minio-secret access_key=access-secret"),
		}
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil {
		t.Fatal("expected startup failure summary")
	}
	if scope.Failure.Service != suiteservices.ServiceMinIO {
		t.Fatalf("unexpected failure service: got %q", scope.Failure.Service)
	}
	if scope.Failure.Stage != stageMinIOStart {
		t.Fatalf("unexpected failure stage: got %q", scope.Failure.Stage)
	}
	if scope.Failure.Operation != "start" {
		t.Fatalf("unexpected failure operation: got %q", scope.Failure.Operation)
	}
	if !scope.Failure.Retryable {
		t.Fatal("expected retryable failure classification")
	}
	if !scope.Failure.RetryBlockedByContext {
		t.Fatal("expected retry-blocked-by-context classification")
	}
	if scope.Failure.AttemptsStarted != 1 || scope.Failure.MaxAttempts != 2 {
		t.Fatalf("unexpected attempts: got %#v", scope.Failure)
	}
	if scope.Failure.DockerEndpoint != "unix:///var/run/docker.sock" {
		t.Fatalf("unexpected docker endpoint: got %q", scope.Failure.DockerEndpoint)
	}
	if !strings.Contains(scope.Failure.Message, "connection refused") {
		t.Fatalf("expected connection failure detail, got %q", scope.Failure.Message)
	}
	for _, secret := range []string{"minio-secret", "access-secret"} {
		if strings.Contains(scope.Failure.Message, secret) {
			t.Fatalf("failure message must redact %q, got %q", secret, scope.Failure.Message)
		}
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
}

func TestRunRecordsPostgresTemplateFailureWithStructuredSummary(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.createTemplate = func(context.Context, string, string) error {
		return errors.New("open postgres admin handle: postgres://cartulary:supersecret@127.0.0.1:5432/postgres?sslmode=disable")
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil {
		t.Fatal("expected template failure summary")
	}
	if scope.Failure.Service != suiteservices.ServicePostgres {
		t.Fatalf("unexpected failure service: got %q", scope.Failure.Service)
	}
	if scope.Failure.Stage != stagePostgresTemplate {
		t.Fatalf("unexpected failure stage: got %q", scope.Failure.Stage)
	}
	if scope.Failure.Operation != "prepare postgres template database" {
		t.Fatalf("unexpected failure operation: got %q", scope.Failure.Operation)
	}
	if strings.Contains(scope.Failure.Message, "supersecret") || strings.Contains(scope.Failure.Message, "postgres://") {
		t.Fatalf("failure message must be redacted, got %q", scope.Failure.Message)
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
}

func TestRunRecordsChildStartFailureWithStructuredSummary(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context) (minioService, error) {
		return minioService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "minio-access",
			secretKey: "minio-secret",
			secure:    false,
			close:     func(context.Context) error { return nil },
		}, nil
	}
	deps.startChild = func([]string, map[string]string) (childProcess, error) {
		return nil, errors.New("exec child failed password=supersecret")
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil {
		t.Fatal("expected child-start failure summary")
	}
	if scope.Failure.Stage != stageChildStart {
		t.Fatalf("unexpected failure stage: got %q", scope.Failure.Stage)
	}
	if scope.Failure.Service != "" {
		t.Fatalf("child-start failure must not attribute a service, got %q", scope.Failure.Service)
	}
	if scope.Failure.Operation != "start child command" {
		t.Fatalf("unexpected failure operation: got %q", scope.Failure.Operation)
	}
	if strings.Contains(scope.Failure.Message, "supersecret") {
		t.Fatalf("failure message must be redacted, got %q", scope.Failure.Message)
	}
	if scope.Cleanup.Status != "child_start_failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
}

type testDeps struct {
	dependencies
	env        map[string]string
	resultsDir string
}

func defaultTestDependencies(t testing.TB) testDeps {
	t.Helper()

	resultsDir := t.TempDir()
	env := map[string]string{
		"CARTULARY_TEST_RESULTS_DIR": resultsDir,
		"CARTULARY_TEST_RUN_ID":      "wrapper-tests",
	}

	return testDeps{
		dependencies: dependencies{
			startPostgres: func(context.Context) (postgresService, error) {
				return postgresService{}, nil
			},
			startMinIO: func(context.Context) (minioService, error) {
				return minioService{}, nil
			},
			startChild: func(argv []string, env map[string]string) (childProcess, error) {
				return fakeChild{}, nil
			},
			createTemplate: func(context.Context, string, string) error { return nil },
			recordEvent: func(env map[string]string, event suiteservices.Event) {
				_ = suiteservices.RecordEvent(env, event)
			},
			refreshSummary: func(env map[string]string) {
				_ = suiteservices.RefreshSummary(env)
			},
			suiteID: func() (string, error) {
				return "suite-redaction", nil
			},
			notifySignals: func(chan<- os.Signal, ...os.Signal) {},
			stopSignals:   func(chan<- os.Signal) {},
		},
		env:        env,
		resultsDir: resultsDir,
	}
}

type fakeChild struct{}

func (fakeChild) Wait() error            { return nil }
func (fakeChild) Signal(os.Signal) error { return nil }
func (fakeChild) Kill() error            { return nil }
func (fakeChild) PID() int               { return 4242 }

func stringContains(haystack string, needle string) bool {
	return strings.Contains(haystack, needle)
}

func loadScope(t testing.TB, deps testDeps) suiteservices.ServiceScope {
	t.Helper()

	raw, err := os.ReadFile(summaryPath(deps))
	if err != nil {
		t.Fatalf("read diagnostics summary: %v", err)
	}
	return decodeScope(t, raw)
}

func decodeScope(t testing.TB, raw []byte) suiteservices.ServiceScope {
	t.Helper()

	var scope suiteservices.ServiceScope
	if err := json.Unmarshal(raw, &scope); err != nil {
		t.Fatalf("decode diagnostics summary: %v", err)
	}
	return scope
}

func summaryPath(deps testDeps) string {
	return filepath.Join(
		deps.resultsDir,
		deps.env["CARTULARY_TEST_RUN_ID"],
		"_shared",
		"test-services",
		"suite-redaction",
		"service-scope.json",
	)
}
