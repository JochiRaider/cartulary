package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
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

	for _, secret := range []string{"supersecret", "access-secret", "minio-secret", "postgres://"} {
		if string(raw) == "" {
			t.Fatal("expected diagnostics summary output")
		}
		if contains := string(raw); contains != "" && stringContains(contains, secret) {
			t.Fatalf("diagnostics summary must redact %q, got %s", secret, raw)
		}
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
