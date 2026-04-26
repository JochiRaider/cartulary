package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

func TestRunPassThroughModeStartsNoServices(t *testing.T) {
	receivedEnv := map[string]string{}
	recordedEvents := []suiteservices.Event{}
	deps := dependencies{
		startPostgres: func(context.Context, map[string]string) (postgresService, error) {
			t.Fatal("startPostgres must not run in pass-through mode")
			return postgresService{}, nil
		},
		startMinIO: func(context.Context, map[string]string) (minioService, error) {
			t.Fatal("startMinIO must not run in pass-through mode")
			return minioService{}, nil
		},
		startChild: func(argv []string, env map[string]string) (childProcess, error) {
			receivedEnv = cloneEnv(env)
			return fakeChild{}, nil
		},
		createTemplate: func(context.Context, string, string) error { return nil },
		recordEvent: func(_ map[string]string, event suiteservices.Event) {
			recordedEvents = append(recordedEvents, event)
		},
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
	if len(recordedEvents) != 1 || recordedEvents[0].Type != suiteservices.EventWrapperPassThrough {
		t.Fatalf("expected one pass-through wrapper event, got %#v", recordedEvents)
	}
}

func TestRunSchedulesServiceReaperOnChildFailureAndPropagatesStatus(t *testing.T) {
	postgresClosed := 0
	minioClosed := 0
	reaperLease := ""
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			containerID: "postgres-container",
			image:       "postgres:test",
			labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "suite-redaction",
				testServiceLabelRunID:   "wrapper-tests",
				testServiceLabelService: suiteservices.ServicePostgres,
			},
			close: func(context.Context) error {
				postgresClosed++
				return nil
			},
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
		return minioService{
			endpoint:    "127.0.0.1:9000",
			accessKey:   "minio-access",
			secretKey:   "minio-secret",
			containerID: "minio-container",
			image:       "minio:test",
			labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "suite-redaction",
				testServiceLabelRunID:   "wrapper-tests",
				testServiceLabelService: suiteservices.ServiceMinIO,
			},
			close: func(context.Context) error {
				minioClosed++
				return nil
			},
		}, nil
	}
	deps.startReaper = func(leasePath string, env map[string]string) error {
		reaperLease = leasePath
		return nil
	}
	deps.startChild = func(argv []string, env map[string]string) (childProcess, error) {
		return startChildProcess([]string{"bash", "-lc", "exit 7"}, env)
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 7 {
		t.Fatalf("unexpected exit status: got %d want 7", status)
	}
	if postgresClosed != 0 {
		t.Fatalf("postgres termination must be delegated to the reaper, got %d direct cleanup call(s)", postgresClosed)
	}
	if minioClosed != 0 {
		t.Fatalf("minio termination must be delegated to the reaper, got %d direct cleanup call(s)", minioClosed)
	}
	if reaperLease == "" {
		t.Fatal("expected service reaper to be scheduled")
	}
	lease, err := readServiceLease(reaperLease)
	if err != nil {
		t.Fatalf("read reaper lease: %v", err)
	}
	if lease.SuiteID != "suite-redaction" || len(lease.Services) != 2 {
		t.Fatalf("unexpected lease: %#v", lease)
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
	events := loadTestEvents(t, deps)
	requireTimingEvent(t, events, bucketSetup, "test-services wrapper setup")
	requireTimingEvent(t, events, bucketServiceWait, "test-services start postgres")
	requireTimingEvent(t, events, bucketMigration, "test-services prepare postgres template database")
	requireTimingEvent(t, events, bucketServiceWait, "test-services start minio")
	requireTimingEvent(t, events, bucketTeardown, "test-services schedule service reaper")
}

func TestRunReturnsBeforeSlowServiceTerminationAfterChildSuccess(t *testing.T) {
	deps := defaultTestDependencies(t)
	closeCalled := make(chan struct{}, 2)
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			containerID: "postgres-container",
			image:       "postgres:test",
			labels:      suiteServiceLabels(map[string]string{suiteservices.SuiteIDEnv: "suite-redaction", "CARTULARY_TEST_RUN_ID": "wrapper-tests"}, suiteservices.ServicePostgres),
			close:       func(context.Context) error { closeCalled <- struct{}{}; time.Sleep(time.Second); return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
		return minioService{
			endpoint:    "127.0.0.1:9000",
			accessKey:   "minio-access",
			secretKey:   "minio-secret",
			containerID: "minio-container",
			image:       "minio:test",
			labels:      suiteServiceLabels(map[string]string{suiteservices.SuiteIDEnv: "suite-redaction", "CARTULARY_TEST_RUN_ID": "wrapper-tests"}, suiteservices.ServiceMinIO),
			close:       func(context.Context) error { closeCalled <- struct{}{}; time.Sleep(time.Second); return nil },
		}, nil
	}
	reaperScheduled := false
	deps.startReaper = func(string, map[string]string) error {
		reaperScheduled = true
		return nil
	}

	start := time.Now()
	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected exit status: got %d want 0", status)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("successful run waited on service termination: %v", elapsed)
	}
	if !reaperScheduled {
		t.Fatal("expected detached service reaper to be scheduled")
	}
	select {
	case <-closeCalled:
		t.Fatal("service close must not run synchronously on successful child completion")
	default:
	}
}

func TestRunStartsMinIOWhilePostgresTemplateIsPreparing(t *testing.T) {
	deps := defaultTestDependencies(t)
	minioStarted := make(chan struct{})
	releaseMinIO := make(chan struct{})
	childStarted := false

	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(ctx context.Context, _ map[string]string) (minioService, error) {
		close(minioStarted)
		select {
		case <-releaseMinIO:
			return minioService{
				endpoint:  "127.0.0.1:9000",
				accessKey: "minio-access",
				secretKey: "minio-secret",
				secure:    false,
				close:     func(context.Context) error { return nil },
			}, nil
		case <-ctx.Done():
			return minioService{}, ctx.Err()
		}
	}
	deps.createTemplate = func(context.Context, string, string) error {
		select {
		case <-minioStarted:
			close(releaseMinIO)
			return nil
		case <-time.After(time.Second):
			return errors.New("minio did not start before postgres template preparation")
		}
	}
	deps.startChild = func(argv []string, env map[string]string) (childProcess, error) {
		childStarted = true
		if env[suiteservices.PGTemplateDBEnv] == "" {
			t.Fatal("child must receive the migrated template database name")
		}
		if env[suiteservices.S3EndpointEnv] != "127.0.0.1:9000" {
			t.Fatalf("child missing minio endpoint: %#v", env)
		}
		if env[suiteservices.S3AccessKeyEnv] != "minio-access" || env[suiteservices.S3SecretKeyEnv] != "minio-secret" {
			t.Fatalf("child missing minio credentials: %#v", env)
		}
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected exit status: got %d want 0", status)
	}
	if !childStarted {
		t.Fatal("expected child command to start after both services are ready")
	}

	events := loadTestEvents(t, deps)
	requireTimingEvent(t, events, bucketServiceWait, "test-services start postgres")
	requireTimingEvent(t, events, bucketServiceWait, "test-services start minio")
	requireTimingEvent(t, events, bucketMigration, "test-services prepare postgres template database")
}

func TestRunCleansUpMinIOWhenPostgresStartupFails(t *testing.T) {
	deps := defaultTestDependencies(t)
	minioClosed := 0
	childStarted := false

	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{}, errors.New("postgres refused startup")
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
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
	deps.startChild = func([]string, map[string]string) (childProcess, error) {
		childStarted = true
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}
	if childStarted {
		t.Fatal("child must not start when postgres startup fails")
	}
	if minioClosed != 1 {
		t.Fatalf("expected minio cleanup after postgres startup failure, got %d", minioClosed)
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil || scope.Failure.Service != suiteservices.ServicePostgres || scope.Failure.Stage != stagePostgresStart {
		t.Fatalf("unexpected startup failure summary: %#v", scope.Failure)
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: %#v", scope.Cleanup)
	}
}

func TestRunRedactsCredentialsInDiagnostics(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:supersecret@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:supersecret@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
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
	childStarted := false
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
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
	deps.startChild = func([]string, map[string]string) (childProcess, error) {
		childStarted = true
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}
	if childStarted {
		t.Fatal("child must not start when minio startup fails")
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
	minioClosed := 0
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
		return minioService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "access-secret",
			secretKey: "minio-secret",
			close: func(context.Context) error {
				minioClosed++
				return nil
			},
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
	if minioClosed != 1 {
		t.Fatalf("expected minio cleanup after template failure, got %d", minioClosed)
	}
}

func TestRunRecordsChildStartFailureWithStructuredSummary(t *testing.T) {
	deps := defaultTestDependencies(t)
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			host:        "127.0.0.1",
			port:        "5432",
			user:        "cartulary",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startMinIO = func(context.Context, map[string]string) (minioService, error) {
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

func TestPrepareWebE2ERequiresActiveSuiteAndTemplate(t *testing.T) {
	deps := defaultTestDependencies(t)
	called := false
	deps.prepareWebE2E = func(context.Context, map[string]string) (webE2EFixture, error) {
		called = true
		return webE2EFixture{}, nil
	}

	status := run([]string{"prepare-web-e2e", "--env-file", filepath.Join(t.TempDir(), "env"), "--metadata-file", filepath.Join(t.TempDir(), "metadata.json")}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("missing active suite must fail with status 1, got %d", status)
	}
	if called {
		t.Fatal("prepare must not run without an active suite")
	}

	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-web-e2e-cleanup"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"
	status = run([]string{"prepare-web-e2e", "--env-file", filepath.Join(t.TempDir(), "env"), "--metadata-file", filepath.Join(t.TempDir(), "metadata.json")}, activeEnv, deps.dependencies)
	if status != 1 {
		t.Fatalf("missing postgres template must fail with status 1, got %d", status)
	}
	if called {
		t.Fatal("prepare must not run without a migrated template database")
	}
}

func TestWarmImagesUsesPinnedServiceImages(t *testing.T) {
	deps := defaultTestDependencies(t)
	var got []string
	deps.warmImages = func(_ context.Context, images []string) error {
		got = append(got, images...)
		return nil
	}

	status := run([]string{"warm-images"}, deps.env, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected warm-images status: got %d want 0", status)
	}

	want := serviceImages()
	if len(got) != len(want) {
		t.Fatalf("unexpected image count: got %v want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("unexpected image at %d: got %q want %q", index, got[index], want[index])
		}
	}
}

func TestPrepareWebE2EWritesShellEnvAndMetadata(t *testing.T) {
	deps := defaultTestDependencies(t)
	envFile := filepath.Join(t.TempDir(), "browser.env")
	metadataFile := filepath.Join(t.TempDir(), "browser.json")
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-web-e2e"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"
	activeEnv[suiteservices.PGTemplateDBEnv] = "suite_template"

	deps.prepareWebE2E = func(_ context.Context, env map[string]string) (webE2EFixture, error) {
		if err := suiteservices.RecordEvent(env, suiteservices.Event{
			Type: suiteservices.EventPostgresDBCreated,
			Name: "ct_web",
			Kind: suiteservices.PostgresPreparationTemplateClone,
			Details: map[string]any{
				"preparation_strategy": suiteservices.PostgresPreparationTemplateClone,
				"template_database":    "suite_template",
				"target":               suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
			},
		}); err != nil {
			return webE2EFixture{}, err
		}
		if err := suiteservices.RecordEvent(env, suiteservices.Event{
			Type: suiteservices.EventPostgresTemplateUse,
			Name: "ct_web",
			Details: map[string]any{
				"template_database": "suite_template",
				"target":            suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
			},
		}); err != nil {
			return webE2EFixture{}, err
		}
		if err := suiteservices.RecordEvent(env, suiteservices.Event{
			Type: suiteservices.EventS3BucketCreated,
			Name: "ct-web",
			Details: map[string]any{
				"reuse_scope": suiteservices.FixtureReusePerTest,
				"target":      suiteservices.LookupEnvValue(env, suiteservices.TargetEnv),
			},
		}); err != nil {
			return webE2EFixture{}, err
		}
		return webE2EFixture{
			DatabaseName: "ct_web",
			DSN:          "postgres://cartulary:pa'ss@127.0.0.1:5432/ct_web?sslmode=disable",
			Bucket:       "ct-web",
			S3Endpoint:   "127.0.0.1:9000",
			S3AccessKey:  "access'key",
			S3SecretKey:  "secret",
			S3Secure:     true,
		}, nil
	}

	status := run([]string{"prepare-web-e2e", "--env-file", envFile, "--metadata-file", metadataFile}, activeEnv, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected prepare status: got %d want 0", status)
	}

	envRaw, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	envText := string(envRaw)
	for _, expected := range []string{
		`export CARTULARY_POSTGRES_DSN='postgres://cartulary:pa'"'"'ss@127.0.0.1:5432/ct_web?sslmode=disable'`,
		`export CARTULARY_S3_ACCESS_KEY_ID='access'"'"'key'`,
		`export CARTULARY_S3_BUCKET='ct-web'`,
		`export CARTULARY_S3_SECURE='true'`,
	} {
		if !strings.Contains(envText, expected) {
			t.Fatalf("env file missing %q in:\n%s", expected, envText)
		}
	}

	metadata, err := readWebE2EMetadata(metadataFile)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if metadata.DatabaseName != "ct_web" || metadata.Bucket != "ct-web" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}

	scope, ok, err := suiteservices.Summarize(activeEnv)
	if err != nil {
		t.Fatalf("summarize browser fixture events: %v", err)
	}
	if !ok {
		t.Fatal("expected browser fixture suite summary")
	}
	if len(scope.Postgres.DatabasePreparations) != 1 {
		t.Fatalf("expected one browser database preparation, got %#v", scope.Postgres.DatabasePreparations)
	}
	preparation := scope.Postgres.DatabasePreparations[0]
	if preparation.Name != "ct_web" ||
		preparation.Strategy != suiteservices.PostgresPreparationTemplateClone ||
		preparation.TemplateDatabase != "suite_template" ||
		preparation.Target != "browser-e2e-webserver-backed" {
		t.Fatalf("unexpected browser database preparation: %#v", preparation)
	}
	if scope.MinIO.BucketCreateCount != 1 || len(scope.MinIO.CreatedBuckets) != 1 || scope.MinIO.CreatedBuckets[0] != "ct-web" {
		t.Fatalf("expected browser fixture to create one isolated bucket, got %#v", scope.MinIO)
	}
	requireTimingEvent(t, loadTestEventsForEnv(t, activeEnv), bucketMigration, "test-services prepare browser e2e fixture")
}

func TestCleanupWebE2ERetiresFixtureWithoutImmediateCleanup(t *testing.T) {
	deps := defaultTestDependencies(t)
	metadataFile := filepath.Join(t.TempDir(), "browser.json")
	if err := writeWebE2EMetadata(metadataFile, webE2EMetadata{DatabaseName: "ct_web", Bucket: "ct-web"}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-web-e2e-cleanup"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"

	status := run([]string{"cleanup-web-e2e", "--metadata-file", metadataFile}, activeEnv, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected cleanup status: got %d want 0", status)
	}
	scope, ok, err := suiteservices.Summarize(activeEnv)
	if err != nil {
		t.Fatalf("summarize retired browser fixture: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}
	if scope.BrowserE2E.RetiredFixtureCount != 1 || len(scope.BrowserE2E.RetiredFixtures) != 1 {
		t.Fatalf("expected one retired browser fixture, got %#v", scope.BrowserE2E)
	}
	retired := scope.BrowserE2E.RetiredFixtures[0]
	if retired.DatabaseName != "ct_web" || retired.Bucket != "ct-web" || retired.Target != "browser-e2e-webserver-backed" {
		t.Fatalf("unexpected retired browser fixture: %#v", retired)
	}
	requireTimingEvent(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services retire browser e2e fixture")
}

func TestCleanupOwnedServicesReclaimsRetiredWebE2EFixturesOnce(t *testing.T) {
	deps := defaultTestDependencies(t)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-web-e2e-cleanup"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"
	for range 2 {
		recordWebE2EFixtureEvent(deps.dependencies, activeEnv, suiteservices.EventWebE2EFixtureRetired, webE2EMetadata{DatabaseName: "ct_web", Bucket: "ct-web"})
	}

	var (
		closedPG bool
		closedS3 bool
	)
	deps.cleanupWebE2EDB = func(context.Context, webE2EMetadata, map[string]string) error {
		t.Fatal("owned-stack teardown must not destructively drop browser fixture databases")
		return nil
	}
	deps.cleanupWebE2EBucket = func(context.Context, webE2EMetadata, map[string]string) error {
		t.Fatal("owned-stack teardown must not destructively clean browser fixture buckets")
		return nil
	}
	deps.detectWebE2ELeaks = func(ctx context.Context, fixtures []webE2EMetadata, env map[string]string) error {
		if len(fixtures) != 1 || fixtures[0].DatabaseName != "ct_web" || fixtures[0].Bucket != "ct-web" {
			t.Fatalf("expected one deduplicated leak-check fixture, got %#v", fixtures)
		}
		if env[suiteservices.PGAdminDSNEnv] != "postgres://suite/postgres" {
			t.Fatalf("leak check missing postgres admin dsn: %#v", env)
		}
		if env[suiteservices.S3EndpointEnv] != "127.0.0.1:9000" {
			t.Fatalf("leak check missing minio endpoint: %#v", env)
		}
		return nil
	}

	cleanupOwnedServices(
		deps.dependencies,
		activeEnv,
		postgresService{
			adminDSN: "postgres://suite/postgres",
			close: func(context.Context) error {
				closedPG = true
				return nil
			},
		},
		minioService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "access",
			secretKey: "secret",
			close: func(context.Context) error {
				closedS3 = true
				return nil
			},
		},
		"",
		"succeeded",
		0,
	)

	if !closedPG || !closedS3 {
		t.Fatalf("expected suite services to close after browser fixture reclamation, postgres=%t minio=%t", closedPG, closedS3)
	}
	scope, ok, err := suiteservices.Summarize(activeEnv)
	if err != nil {
		t.Fatalf("summarize reclaimed browser fixture: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}
	if scope.BrowserE2E.CleanedFixtureCount != 0 || len(scope.BrowserE2E.CleanedFixtures) != 0 {
		t.Fatalf("owned-stack teardown must not record destructive cleanup, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.ReclaimedFixtureCount != 1 || len(scope.BrowserE2E.ReclaimedFixtures) != 1 {
		t.Fatalf("expected one reclaimed browser fixture, got %#v", scope.BrowserE2E)
	}
	reclaimed := scope.BrowserE2E.ReclaimedFixtures[0]
	if reclaimed.ReclaimStrategy != webE2EReclaimStrategyOwnedStack {
		t.Fatalf("expected owned-stack reclaim strategy, got %#v", reclaimed)
	}
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services check browser e2e fixture leaks", "pass")
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services reclaim pooled browser fixtures by owned stack", "pass")
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services janitor stale browser fixtures", "pass")
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services terminate postgres", "pass")
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services terminate minio", "pass")
}

func TestCleanupOwnedServicesTerminatesServicesConcurrently(t *testing.T) {
	deps := defaultTestDependencies(t)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-concurrent-service-cleanup"
	activeEnv[suiteservices.TargetEnv] = "check-service-backed"

	postgresStarted := make(chan struct{})
	minioStarted := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	go func() {
		cleanupOwnedServices(
			deps.dependencies,
			activeEnv,
			postgresService{
				close: func(context.Context) error {
					close(postgresStarted)
					<-minioStarted
					<-release
					return nil
				},
			},
			minioService{
				close: func(context.Context) error {
					close(minioStarted)
					<-postgresStarted
					<-release
					return nil
				},
			},
			"",
			"succeeded",
			0,
		)
		close(done)
	}()

	select {
	case <-postgresStarted:
	case <-time.After(time.Second):
		t.Fatal("postgres cleanup did not start")
	}
	select {
	case <-minioStarted:
	case <-time.After(time.Second):
		t.Fatal("minio cleanup did not start concurrently with postgres")
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not finish after releasing concurrent service termination")
	}
}

func TestStaleWebE2EJanitorBoundsAndFiltersGeneratedFixtures(t *testing.T) {
	deps := defaultTestDependencies(t)
	staleEnv := cloneEnv(deps.env)
	staleEnv[suiteservices.ActiveEnv] = "1"
	staleEnv[suiteservices.SuiteIDEnv] = "suite-stale-browser-fixtures"
	staleEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"

	for index := range staleFixtureMaxCandidates + 3 {
		metadata := webE2EMetadata{
			DatabaseName: fmt.Sprintf("ct_abcd1234_ef567890_%06d_web_e2e", index+1),
			Bucket:       fmt.Sprintf("ct-abcd1234-ef567890-%06d-web-e2e", index+1),
			Target:       "browser-e2e-webserver-backed",
		}
		recordWebE2EFixtureEvent(deps.dependencies, staleEnv, suiteservices.EventWebE2EFixtureRetired, metadata)
	}
	recordWebE2EFixtureEvent(deps.dependencies, staleEnv, suiteservices.EventWebE2EFixtureRetired, webE2EMetadata{
		DatabaseName: "not_cartulary",
		Bucket:       "not-cartulary",
		Target:       "browser-e2e-webserver-backed",
	})
	deps.refreshSummary(staleEnv)

	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-active-browser-fixtures"
	activeEnv[suiteservices.TargetEnv] = "check-service-backed"

	var (
		cleanedMu sync.Mutex
		cleaned   []webE2EMetadata
	)
	deps.cleanupWebE2EDB = func(_ context.Context, metadata webE2EMetadata, _ map[string]string) error {
		cleanedMu.Lock()
		defer cleanedMu.Unlock()
		cleaned = append(cleaned, metadata)
		return nil
	}
	deps.cleanupWebE2EBucket = func(context.Context, webE2EMetadata, map[string]string) error { return nil }

	if err := cleanupStaleWebE2EFixtures(context.Background(), deps.dependencies, activeEnv); err != nil {
		t.Fatalf("cleanup stale browser fixtures: %v", err)
	}
	if len(cleaned) != staleFixtureMaxCandidates {
		t.Fatalf("expected janitor to clean bounded generated fixtures, got %d want %d", len(cleaned), staleFixtureMaxCandidates)
	}
	for _, fixture := range cleaned {
		if !generatedWebE2EFixture(fixture) {
			t.Fatalf("janitor cleaned non-generated fixture: %#v", fixture)
		}
	}
	events := loadTestEventsForEnv(t, activeEnv)
	requireTimingEventStatus(t, events, bucketTeardown, "test-services cleanup browser e2e fixture database", "pass")
	requireTimingEventStatus(t, events, bucketTeardown, "test-services cleanup browser e2e fixture bucket", "pass")
}

func TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak(t *testing.T) {
	deps := defaultTestDependencies(t)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-web-e2e-cleanup-fail"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e-webserver-backed"
	recordWebE2EFixtureEvent(deps.dependencies, activeEnv, suiteservices.EventWebE2EFixtureRetired, webE2EMetadata{DatabaseName: "ct_web", Bucket: "ct-web"})

	deps.detectWebE2ELeaks = func(context.Context, []webE2EMetadata, map[string]string) error {
		return errors.New("browser e2e fixture database \"ct_web\" has 1 active postgres connection")
	}
	deps.cleanupWebE2EDB = func(context.Context, webE2EMetadata, map[string]string) error {
		t.Fatal("leaked owned-stack fixtures must not fall through to destructive db cleanup")
		return nil
	}
	deps.cleanupWebE2EBucket = func(context.Context, webE2EMetadata, map[string]string) error {
		t.Fatal("leaked owned-stack fixtures must not fall through to destructive bucket cleanup")
		return nil
	}

	cleanupOwnedServices(deps.dependencies, activeEnv, postgresService{}, minioService{}, "", "succeeded", 0)

	scope, ok, err := suiteservices.Summarize(activeEnv)
	if err != nil {
		t.Fatalf("summarize failed browser fixture cleanup: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}
	if scope.Cleanup.Status != "cleanup_failed" {
		t.Fatalf("expected cleanup_failed status, got %#v", scope.Cleanup)
	}
	if scope.BrowserE2E.ReclaimedFixtureCount != 0 {
		t.Fatalf("leaked fixture must not be reclaimed, got %#v", scope.BrowserE2E)
	}
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services check browser e2e fixture leaks", "fail")
}

func TestPrepareWebE2ECleansFixtureWhenEnvWriteFails(t *testing.T) {
	deps := defaultTestDependencies(t)
	envFile := t.TempDir()
	metadataFile := filepath.Join(t.TempDir(), "browser.json")
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.PGTemplateDBEnv] = "suite_template"

	deps.prepareWebE2E = func(context.Context, map[string]string) (webE2EFixture, error) {
		return webE2EFixture{
			DatabaseName: "ct_web",
			DSN:          "postgres://cartulary:cartulary@127.0.0.1:5432/ct_web?sslmode=disable",
			Bucket:       "ct-web",
		}, nil
	}
	cleanupDBCalls := 0
	cleanupBucketCalls := 0
	deps.cleanupWebE2EDB = func(ctx context.Context, metadata webE2EMetadata, env map[string]string) error {
		cleanupDBCalls++
		if metadata.DatabaseName != "ct_web" || metadata.Bucket != "ct-web" {
			t.Fatalf("database cleanup received unexpected metadata: %#v", metadata)
		}
		return nil
	}
	deps.cleanupWebE2EBucket = func(ctx context.Context, metadata webE2EMetadata, env map[string]string) error {
		cleanupBucketCalls++
		if metadata.DatabaseName != "ct_web" || metadata.Bucket != "ct-web" {
			t.Fatalf("bucket cleanup received unexpected metadata: %#v", metadata)
		}
		return nil
	}

	status := run([]string{"prepare-web-e2e", "--env-file", envFile, "--metadata-file", metadataFile}, activeEnv, deps.dependencies)
	if status != 1 {
		t.Fatalf("env write failure must return status 1, got %d", status)
	}
	if cleanupDBCalls != 1 || cleanupBucketCalls != 1 {
		t.Fatalf("expected database and bucket cleanup after env write failure, got db=%d bucket=%d", cleanupDBCalls, cleanupBucketCalls)
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
			startPostgres: func(context.Context, map[string]string) (postgresService, error) {
				return postgresService{}, nil
			},
			startMinIO: func(context.Context, map[string]string) (minioService, error) {
				return minioService{}, nil
			},
			startChild: func(argv []string, env map[string]string) (childProcess, error) {
				return fakeChild{}, nil
			},
			startReaper:         func(string, map[string]string) error { return nil },
			createTemplate:      func(context.Context, string, string) error { return nil },
			prepareWebE2E:       func(context.Context, map[string]string) (webE2EFixture, error) { return webE2EFixture{}, nil },
			cleanupWebE2EDB:     func(context.Context, webE2EMetadata, map[string]string) error { return nil },
			cleanupWebE2EBucket: func(context.Context, webE2EMetadata, map[string]string) error { return nil },
			detectWebE2ELeaks:   func(context.Context, []webE2EMetadata, map[string]string) error { return nil },
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

func loadTestEvents(t testing.TB, deps testDeps) []suiteservices.Event {
	t.Helper()
	env := cloneEnv(deps.env)
	env[suiteservices.SuiteIDEnv] = "suite-redaction"
	return loadTestEventsForEnv(t, env)
}

func loadTestEventsForEnv(t testing.TB, env map[string]string) []suiteservices.Event {
	t.Helper()
	suiteDir, ok, err := suiteservices.ResolveSuiteArtifactDir(env)
	if err != nil {
		t.Fatalf("resolve suite artifact dir: %v", err)
	}
	if !ok {
		t.Fatal("expected suite artifact dir")
	}
	eventFiles, err := filepath.Glob(filepath.Join(suiteDir, "events", "*.json"))
	if err != nil {
		t.Fatalf("list suite service events: %v", err)
	}
	events := make([]suiteservices.Event, 0, len(eventFiles))
	for _, eventPath := range eventFiles {
		raw, err := os.ReadFile(eventPath)
		if err != nil {
			t.Fatalf("read event %s: %v", eventPath, err)
		}
		var event suiteservices.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatalf("decode event %s: %v", eventPath, err)
		}
		events = append(events, event)
	}
	return events
}

func requireTimingEvent(t testing.TB, events []suiteservices.Event, bucket string, label string) {
	t.Helper()
	requireTimingEventStatus(t, events, bucket, label, "")
}

func requireTimingEventStatus(t testing.TB, events []suiteservices.Event, bucket string, label string, status string) {
	t.Helper()
	for _, event := range events {
		if event.Type != suiteservices.EventTimingSpan {
			continue
		}
		if event.Details["bucket"] != bucket || event.Details["label"] != label {
			continue
		}
		if status != "" && event.Details["status"] != status {
			t.Fatalf("timing event %q status: got %#v want %q", label, event.Details["status"], status)
		}
		if event.Details["duration_ms"] == nil {
			t.Fatalf("timing event %q missing duration_ms: %#v", label, event)
		}
		return
	}
	t.Fatalf("missing timing event bucket=%s label=%q in %#v", bucket, label, events)
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
