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

	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
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
		startObjectStore: func(context.Context, map[string]string) (objectStoreService, error) {
			t.Fatal("startObjectStore must not run in pass-through mode")
			return objectStoreService{}, nil
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
	objectStoreClosed := 0
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:    "127.0.0.1:9000",
			accessKey:   "object-store-access",
			secretKey:   "object-store-secret",
			containerID: "object-store-container",
			image:       "object-store:test",
			labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "suite-redaction",
				testServiceLabelRunID:   "wrapper-tests",
				testServiceLabelService: suiteservices.ServiceObjectStore,
			},
			close: func(context.Context) error {
				objectStoreClosed++
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
	if objectStoreClosed != 0 {
		t.Fatalf("object-store termination must be delegated to the reaper, got %d direct cleanup call(s)", objectStoreClosed)
	}
	if reaperLease == "" {
		t.Fatal("expected service reaper to be scheduled")
	}
	lease, err := readServiceLease(reaperLease)
	if err != nil {
		t.Fatalf("read reaper lease: %v", err)
	}
	if lease.SuiteID != "suite-redaction" || len(lease.Resources) != 2 {
		t.Fatalf("unexpected lease: %#v", lease)
	}
	if lease.LeaseID == "" || lease.RunRoot == "" || lease.OwnerPID == 0 || lease.CleanupState != "not_started" {
		t.Fatalf("lease missing cleanup-proof fields: %#v", lease)
	}
	if loadScope(t, deps).Failure != nil {
		t.Fatal("non-zero child exit after a successful start must not populate startup failure summary")
	}

	lifecycleEnv := cloneEnv(deps.env)
	lifecycleEnv[suiteservices.SuiteIDEnv] = "suite-redaction"
	state, ok, err := suiteservices.CurrentLifecycleState(lifecycleEnv)
	if err != nil || !ok {
		t.Fatalf("read delegated cleanup lifecycle state: state=%q ok=%t err=%v", state, ok, err)
	}
	if state != "cleaning" {
		t.Fatalf("delegated cleanup must remain cleaning until terminate-suite completes, got %q", state)
	}
	lifecycleRecords, err := suiteservices.ReadLifecycleEvents(lifecycleEnv)
	if err != nil {
		t.Fatalf("read delegated cleanup lifecycle events: %v", err)
	}
	cleanupStarts := 0
	cleanupTerminals := 0
	for _, record := range lifecycleRecords {
		if record.Event == suiteservices.LifecycleEventCleanupStarted {
			cleanupStarts++
		}
		if record.Event == suiteservices.LifecycleEventCleanupSucceeded || record.Event == suiteservices.LifecycleEventCleanupFailed {
			cleanupTerminals++
		}
	}
	if cleanupStarts != 1 || cleanupTerminals != 0 {
		t.Fatalf("delegated cleanup must emit one start and defer its terminal event, starts=%d terminals=%d records=%#v", cleanupStarts, cleanupTerminals, lifecycleRecords)
	}
	events := loadTestEvents(t, deps)
	requireTimingEvent(t, events, bucketSetup, "test-services wrapper setup")
	requireTimingEvent(t, events, bucketServiceWait, "test-services start postgres")
	requireTimingEvent(t, events, bucketMigration, "test-services prepare postgres template database")
	requireTimingEvent(t, events, bucketServiceWait, "test-services start object-store")
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:    "127.0.0.1:9000",
			accessKey:   "object-store-access",
			secretKey:   "object-store-secret",
			containerID: "object-store-container",
			image:       "object-store:test",
			labels:      suiteServiceLabels(map[string]string{suiteservices.SuiteIDEnv: "suite-redaction", "CARTULARY_TEST_RUN_ID": "wrapper-tests"}, suiteservices.ServiceObjectStore),
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

func TestRunStartsObjectStoreWhilePostgresTemplateIsPreparing(t *testing.T) {
	deps := defaultTestDependencies(t)
	objectStoreStarted := make(chan struct{})
	releaseObjectStore := make(chan struct{})
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
	deps.startObjectStore = func(ctx context.Context, _ map[string]string) (objectStoreService, error) {
		close(objectStoreStarted)
		select {
		case <-releaseObjectStore:
			return objectStoreService{
				endpoint:  "127.0.0.1:9000",
				accessKey: "object-store-access",
				secretKey: "object-store-secret",
				secure:    false,
				close:     func(context.Context) error { return nil },
			}, nil
		case <-ctx.Done():
			return objectStoreService{}, ctx.Err()
		}
	}
	deps.createTemplate = func(context.Context, string, string) error {
		select {
		case <-objectStoreStarted:
			close(releaseObjectStore)
			return nil
		case <-time.After(time.Second):
			return errors.New("object-store did not start before postgres template preparation")
		}
	}
	deps.startChild = func(argv []string, env map[string]string) (childProcess, error) {
		childStarted = true
		if env[suiteservices.PGTemplateDBEnv] == "" {
			t.Fatal("child must receive the migrated template database name")
		}
		if env[suiteservices.S3EndpointEnv] != "127.0.0.1:9000" {
			t.Fatalf("child missing object-store endpoint: %#v", env)
		}
		if env[suiteservices.S3AccessKeyEnv] != "object-store-access" || env[suiteservices.S3SecretKeyEnv] != "object-store-secret" {
			t.Fatalf("child missing object-store credentials: %#v", env)
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
	requireTimingEvent(t, events, bucketServiceWait, "test-services start object-store")
	requireTimingEvent(t, events, bucketMigration, "test-services prepare postgres template database")
}

func TestSuiteServiceStartupUsesServiceSpecificAttemptTimeouts(t *testing.T) {
	previousPostgresStarter := startPostgresHarnessWithOptions
	previousObjectStoreStarter := startObjectStoreHarnessWithOptions
	defer func() {
		startPostgresHarnessWithOptions = previousPostgresStarter
		startObjectStoreHarnessWithOptions = previousObjectStoreStarter
	}()

	var postgresAttemptTimeout time.Duration
	var objectStoreAttemptTimeout time.Duration

	startPostgresHarnessWithOptions = func(ctx context.Context, options pgtest.StartOptions) (*pgtest.Harness, error) {
		postgresAttemptTimeout = options.AttemptTimeout
		return &pgtest.Harness{
			Host:     "127.0.0.1",
			Port:     "5432",
			User:     "cartulary",
			Password: "cartulary",
		}, nil
	}
	startObjectStoreHarnessWithOptions = func(ctx context.Context, options s3test.StartOptions) (*s3test.Harness, error) {
		objectStoreAttemptTimeout = options.AttemptTimeout
		return &s3test.Harness{
			Endpoint:  "127.0.0.1:9000",
			AccessKey: "object-store-access",
			SecretKey: "object-store-secret",
		}, nil
	}

	env := map[string]string{suiteservices.SuiteIDEnv: "suite-attempt-timeouts"}
	if _, err := startPostgresService(context.Background(), env); err != nil {
		t.Fatalf("start postgres service: %v", err)
	}
	if _, err := startObjectStoreService(context.Background(), env); err != nil {
		t.Fatalf("start object-store service: %v", err)
	}

	if postgresAttemptTimeout != 35*time.Second {
		t.Fatalf("unexpected postgres attempt timeout: got %v want %v", postgresAttemptTimeout, 35*time.Second)
	}
	if objectStoreAttemptTimeout != 2*time.Minute {
		t.Fatalf("unexpected object-store attempt timeout: got %v want %v", objectStoreAttemptTimeout, 2*time.Minute)
	}
	if postgresAttemptTimeout == objectStoreAttemptTimeout {
		t.Fatal("postgres and object-store suite startup attempts must not share one timeout budget")
	}
}

func TestRunDisablesRyukOnlyForSuiteStartup(t *testing.T) {
	previous, hadPrevious := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED")
	_ = os.Unsetenv("TESTCONTAINERS_RYUK_DISABLED")
	defer func() {
		if hadPrevious {
			_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", previous)
			return
		}
		_ = os.Unsetenv("TESTCONTAINERS_RYUK_DISABLED")
	}()

	deps := defaultTestDependencies(t)
	postgresSawRyukDisabled := false
	objectStoreSawRyukDisabled := false
	childSawRyukSetting := false
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		postgresSawRyukDisabled = os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "true"
		return postgresService{
			adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
			dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
			containerID: "postgres-container",
			image:       "postgres:test",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		objectStoreSawRyukDisabled = os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "true"
		return objectStoreService{
			endpoint:    "127.0.0.1:9000",
			accessKey:   "object-store-access",
			secretKey:   "object-store-secret",
			containerID: "object-store-container",
			image:       "object-store:test",
			close:       func(context.Context) error { return nil },
		}, nil
	}
	deps.startChild = func(_ []string, env map[string]string) (childProcess, error) {
		_, childSawRyukSetting = env["TESTCONTAINERS_RYUK_DISABLED"]
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 0 {
		t.Fatalf("unexpected exit status: got %d want 0", status)
	}
	if !postgresSawRyukDisabled || !objectStoreSawRyukDisabled {
		t.Fatalf("suite service startup must disable Ryuk, postgres=%t object_store=%t", postgresSawRyukDisabled, objectStoreSawRyukDisabled)
	}
	if childSawRyukSetting {
		t.Fatal("child scheduler env must not receive a synthetic Ryuk override")
	}
	if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); ok {
		t.Fatal("suite-owned Ryuk override must be restored after run")
	}

	scope := loadScope(t, deps)
	if !scope.Preflight.RyukDisabledForSuiteStartup || scope.Preflight.Status != "pass" {
		t.Fatalf("unexpected preflight summary: %#v", scope.Preflight)
	}
}

func TestRunFailsFastWhenSuitePreflightFails(t *testing.T) {
	deps := defaultTestDependencies(t)
	startedPostgres := false
	startedObjectStore := false
	startedChild := false
	deps.preflightSuite = func(context.Context, map[string]string) (suitePreflightResult, error) {
		return suitePreflightResult{
			DockerEndpoint: "unix:///var/run/docker.sock",
			DockerOK:       false,
		}, errors.New("ping docker endpoint unix:///var/run/docker.sock: connection refused")
	}
	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		startedPostgres = true
		return postgresService{}, nil
	}
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		startedObjectStore = true
		return objectStoreService{}, nil
	}
	deps.startChild = func([]string, map[string]string) (childProcess, error) {
		startedChild = true
		return fakeChild{}, nil
	}

	status := run([]string{"run", "--", "ignored"}, deps.env, deps.dependencies)
	if status != 1 {
		t.Fatalf("unexpected exit status: got %d want 1", status)
	}
	if startedPostgres || startedObjectStore || startedChild {
		t.Fatalf("preflight failure must stop before services or child, postgres=%t object_store=%t child=%t", startedPostgres, startedObjectStore, startedChild)
	}
	scope := loadScope(t, deps)
	if scope.Preflight.Status != "fail" || scope.Preflight.DockerEndpoint != "unix:///var/run/docker.sock" {
		t.Fatalf("unexpected preflight summary: %#v", scope.Preflight)
	}
	if scope.Preflight.FailureClass != suiteservices.FailureClassInfra || scope.Preflight.FailureReason != "preflight_error" {
		t.Fatalf("unexpected preflight failure fields: %#v", scope.Preflight)
	}
	if scope.Failure == nil || scope.Failure.Stage != stageStartupPreflight || scope.Failure.FailureClass != suiteservices.FailureClassInfra {
		t.Fatalf("unexpected preflight failure summary: %#v", scope.Failure)
	}
	if scope.Failure.FailureReason != "preflight_error" {
		t.Fatalf("unexpected preflight failure reason: %#v", scope.Failure)
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: %#v", scope.Cleanup)
	}
	requireLifecycleFailure(t, deps, suiteservices.LifecycleEventStartupFailed, suiteservices.FailureClassInfra, "preflight_error")
	requireTimingEventStatus(t, loadTestEvents(t, deps), bucketSetup, "test-services suite startup preflight", "fail")
}

func TestRunCleansUpObjectStoreWhenPostgresStartupFails(t *testing.T) {
	deps := defaultTestDependencies(t)
	objectStoreClosed := 0
	childStarted := false

	deps.startPostgres = func(context.Context, map[string]string) (postgresService, error) {
		return postgresService{}, errors.New("postgres refused startup")
	}
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "object-store-access",
			secretKey: "object-store-secret",
			close: func(context.Context) error {
				objectStoreClosed++
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
	if objectStoreClosed != 1 {
		t.Fatalf("expected object-store cleanup after postgres startup failure, got %d", objectStoreClosed)
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil || scope.Failure.Service != suiteservices.ServicePostgres || scope.Failure.Stage != stagePostgresStart {
		t.Fatalf("unexpected startup failure summary: %#v", scope.Failure)
	}
	if scope.Failure.FailureClass != suiteservices.FailureClassInfra {
		t.Fatalf("unexpected startup failure class: got %q", scope.Failure.FailureClass)
	}
	if scope.Failure.FailureReason != "service_start_error" {
		t.Fatalf("unexpected startup failure reason: got %q", scope.Failure.FailureReason)
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: %#v", scope.Cleanup)
	}
	requireLifecycleFailure(t, deps, suiteservices.LifecycleEventStartupFailed, suiteservices.FailureClassInfra, "service_start_error")
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "access-secret",
			secretKey: "object-store-secret",
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
	if info, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("stat diagnostics summary: %v", err)
	} else if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("diagnostics summary must be owner-only, got mode %v", info.Mode().Perm())
	}
	if info, err := os.Stat(filepath.Dir(summaryPath)); err != nil {
		t.Fatalf("stat diagnostics dir: %v", err)
	} else if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("diagnostics dir must be owner-only, got mode %v", info.Mode().Perm())
	}

	scope := decodeScope(t, raw)
	if scope.Failure != nil {
		t.Fatalf("successful run must omit failure summary, got %#v", scope.Failure)
	}

	for _, secret := range []string{"supersecret", "access-secret", "object-store-secret"} {
		if string(raw) == "" {
			t.Fatal("expected diagnostics summary output")
		}
		if contains := string(raw); contains != "" && stringContains(contains, secret) {
			t.Fatalf("diagnostics summary must redact %q, got %s", secret, raw)
		}
	}
}

func TestRunRecordsObjectStoreStartupFailureWithStructuredSummary(t *testing.T) {
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{}, &testcontainersx.StartFailure{
			Operation:             "start",
			Service:               "object-store testcontainer",
			Image:                 "docker.io/chrislusf/seaweedfs:4.17:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9",
			DockerEndpoint:        "unix:///var/run/docker.sock",
			AttemptsStarted:       1,
			MaxAttempts:           2,
			Retryable:             true,
			RetryBlockedByContext: true,
			Cause:                 errors.New("docker.sock connection refused password=object-store-secret access_key=access-secret"),
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
		t.Fatal("child must not start when object-store startup fails")
	}

	scope := loadScope(t, deps)
	if scope.Failure == nil {
		t.Fatal("expected startup failure summary")
	}
	if scope.Failure.FailureClass != suiteservices.FailureClassInfra {
		t.Fatalf("unexpected object-store failure class: got %q", scope.Failure.FailureClass)
	}
	if scope.Failure.FailureReason != "service_start_error" {
		t.Fatalf("unexpected object-store failure reason: got %q", scope.Failure.FailureReason)
	}
	if scope.Failure.Service != suiteservices.ServiceObjectStore {
		t.Fatalf("unexpected failure service: got %q", scope.Failure.Service)
	}
	if scope.Failure.Stage != stageObjectStoreStart {
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
	for _, secret := range []string{"object-store-secret", "access-secret"} {
		if strings.Contains(scope.Failure.Message, secret) {
			t.Fatalf("failure message must redact %q, got %q", secret, scope.Failure.Message)
		}
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
	requireLifecycleFailure(t, deps, suiteservices.LifecycleEventStartupFailed, suiteservices.FailureClassInfra, "service_start_error")
}

func TestServiceStartupAttemptTimingSummarizesRetries(t *testing.T) {
	deps := defaultTestDependencies(t)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-startup-attempts"
	activeEnv[suiteservices.TargetEnv] = "browser-e2e"

	start := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	recordServiceStartupAttempt(activeEnv, suiteservices.ServicePostgres, testcontainersx.StartEvent{
		Attempt:     1,
		MaxAttempts: 3,
		StartTime:   start,
		EndTime:     start.Add(20 * time.Millisecond),
		Duration:    20 * time.Millisecond,
		Status:      "fail",
		Retryable:   true,
		Err:         errors.New("wait for postgres readiness: postgres://cartulary:secret@127.0.0.1:5432/postgres password=secret"),
	})
	recordServiceStartupAttempt(activeEnv, suiteservices.ServicePostgres, testcontainersx.StartEvent{
		Attempt:     2,
		MaxAttempts: 3,
		StartTime:   start.Add(25 * time.Millisecond),
		EndTime:     start.Add(30 * time.Millisecond),
		Duration:    5 * time.Millisecond,
		Status:      "pass",
	})
	recordServiceStartupAttempt(activeEnv, suiteservices.ServiceObjectStore, testcontainersx.StartEvent{
		Attempt:     1,
		MaxAttempts: 2,
		StartTime:   start,
		EndTime:     start.Add(10 * time.Millisecond),
		Duration:    10 * time.Millisecond,
		Status:      "pass",
	})
	deps.refreshSummary(activeEnv)

	scope, ok, err := suiteservices.Summarize(activeEnv)
	if err != nil {
		t.Fatalf("summarize startup attempts: %v", err)
	}
	if !ok {
		t.Fatal("expected startup attempt summary")
	}
	if scope.Postgres.Startup.AttemptCount != 2 || scope.Postgres.Startup.RetryCount != 1 {
		t.Fatalf("unexpected postgres startup summary: %#v", scope.Postgres.Startup)
	}
	if scope.Postgres.Startup.SlowestAttemptDurationMS != 20 {
		t.Fatalf("unexpected slowest postgres attempt: %#v", scope.Postgres.Startup)
	}
	if scope.Postgres.Startup.FinalAttempt != 2 || scope.Postgres.Startup.FinalStatus != "pass" {
		t.Fatalf("unexpected final postgres attempt: %#v", scope.Postgres.Startup)
	}
	if len(scope.Postgres.Startup.Attempts) != 2 || !scope.Postgres.Startup.Attempts[0].RetryScheduled {
		t.Fatalf("expected retry-scheduled first postgres attempt, got %#v", scope.Postgres.Startup.Attempts)
	}
	if strings.Contains(scope.Postgres.Startup.Attempts[0].Message, "secret@") || strings.Contains(scope.Postgres.Startup.Attempts[0].Message, "password=secret") {
		t.Fatalf("startup attempt message must be redacted, got %q", scope.Postgres.Startup.Attempts[0].Message)
	}
	if scope.ObjectStore.Startup.AttemptCount != 1 || scope.ObjectStore.Startup.FinalStatus != "pass" {
		t.Fatalf("unexpected object-store startup summary: %#v", scope.ObjectStore.Startup)
	}
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketServiceWait, "test-services start postgres attempt 1", "fail")
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketServiceWait, "test-services start postgres attempt 2", "pass")
}

func TestRunRecordsPostgresTemplateFailureWithStructuredSummary(t *testing.T) {
	deps := defaultTestDependencies(t)
	objectStoreClosed := 0
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "access-secret",
			secretKey: "object-store-secret",
			close: func(context.Context) error {
				objectStoreClosed++
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
	if scope.Failure.FailureClass != suiteservices.FailureClassHelper {
		t.Fatalf("unexpected template failure class: got %q", scope.Failure.FailureClass)
	}
	if scope.Failure.FailureReason != "fixture_error" {
		t.Fatalf("unexpected template failure reason: got %q", scope.Failure.FailureReason)
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
	if strings.Contains(scope.Failure.Message, "supersecret") {
		t.Fatalf("failure message must be redacted, got %q", scope.Failure.Message)
	}
	if scope.Cleanup.Status != "startup_failed" {
		t.Fatalf("unexpected cleanup status: got %#v", scope.Cleanup)
	}
	if objectStoreClosed != 1 {
		t.Fatalf("expected object-store cleanup after template failure, got %d", objectStoreClosed)
	}
	requireLifecycleFailure(t, deps, suiteservices.LifecycleEventStartupFailed, suiteservices.FailureClassHelper, "fixture_error")
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
	deps.startObjectStore = func(context.Context, map[string]string) (objectStoreService, error) {
		return objectStoreService{
			endpoint:  "127.0.0.1:9000",
			accessKey: "object-store-access",
			secretKey: "object-store-secret",
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
	if scope.Failure.FailureClass != suiteservices.FailureClassHelper {
		t.Fatalf("unexpected child-start failure class: got %q", scope.Failure.FailureClass)
	}
	if scope.Failure.FailureReason != "child_target_failure" {
		t.Fatalf("unexpected child-start failure reason: got %q", scope.Failure.FailureReason)
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
		`export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN='postgres://cartulary:pa'"'"'ss@127.0.0.1:5432/ct_web?sslmode=disable'`,
		`export CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID='access'"'"'key'`,
		`export CARTULARY_S3_OBJECT_PRIMARY_BUCKET='ct-web'`,
		`export CARTULARY_S3_OBJECT_PRIMARY_SECURE='true'`,
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
	if scope.ObjectStore.BucketCreateCount != 1 || len(scope.ObjectStore.CreatedBuckets) != 1 || scope.ObjectStore.CreatedBuckets[0] != "ct-web" {
		t.Fatalf("expected browser fixture to create one isolated bucket, got %#v", scope.ObjectStore)
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
			t.Fatalf("leak check missing object-store endpoint: %#v", env)
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
		objectStoreService{
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
		t.Fatalf("expected suite services to close after browser fixture reclamation, postgres=%t object_store=%t", closedPG, closedS3)
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
	requireTimingEventStatus(t, loadTestEventsForEnv(t, activeEnv), bucketTeardown, "test-services terminate object-store", "pass")
}

func TestCleanupOwnedServicesTerminatesServicesConcurrently(t *testing.T) {
	deps := defaultTestDependencies(t)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "suite-concurrent-service-cleanup"
	activeEnv[suiteservices.TargetEnv] = "check"

	postgresStarted := make(chan struct{})
	objectStoreStarted := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	go func() {
		cleanupOwnedServices(
			deps.dependencies,
			activeEnv,
			postgresService{
				close: func(context.Context) error {
					close(postgresStarted)
					<-objectStoreStarted
					<-release
					return nil
				},
			},
			objectStoreService{
				close: func(context.Context) error {
					close(objectStoreStarted)
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
	case <-objectStoreStarted:
	case <-time.After(time.Second):
		t.Fatal("object-store cleanup did not start concurrently with postgres")
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
	activeEnv[suiteservices.TargetEnv] = "check"

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

func TestPreviousSuiteContainerCleanupEligibilityUsesCompletedSummaryOrAge(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"

	completedEnv := cloneEnv(deps.env)
	completedEnv[suiteservices.ActiveEnv] = "1"
	completedEnv[suiteservices.SuiteIDEnv] = "completed-suite"
	completedEnv["CARTULARY_TEST_RUN_ID"] = "completed-run"
	recordCleanupAndRefresh(deps.dependencies, completedEnv, "succeeded", 0)

	completedContainer := dockercontainer.Summary{
		ID:      "completed-container",
		Created: now.Add(-time.Minute).Unix(),
		Labels: map[string]string{
			testServiceLabelManaged: testServiceManagedValue,
			testServiceLabelSuiteID: "completed-suite",
			testServiceLabelRunID:   "completed-run",
			testServiceLabelService: suiteservices.ServicePostgres,
		},
	}
	if !previousSuiteContainerCleanupEligible(activeEnv, completedContainer, now) {
		t.Fatal("container with completed suite cleanup should be eligible")
	}

	activeContainer := dockercontainer.Summary{
		ID:      "active-container",
		Created: now.Add(-time.Hour).Unix(),
		Labels: map[string]string{
			testServiceLabelManaged: testServiceManagedValue,
			testServiceLabelSuiteID: "active-suite",
			testServiceLabelRunID:   "wrapper-tests",
			testServiceLabelService: suiteservices.ServicePostgres,
		},
	}
	if previousSuiteContainerCleanupEligible(activeEnv, activeContainer, now) {
		t.Fatal("current active suite container must not be eligible")
	}

	staleContainer := dockercontainer.Summary{
		ID:      "stale-container",
		Created: now.Add(-staleSuiteContainerAge - time.Second).Unix(),
		Labels: map[string]string{
			testServiceLabelManaged: testServiceManagedValue,
			testServiceLabelSuiteID: "stale-suite",
			testServiceLabelRunID:   "stale-run",
			testServiceLabelService: suiteservices.ServiceObjectStore,
		},
	}
	if !previousSuiteContainerCleanupEligible(activeEnv, staleContainer, now) {
		t.Fatal("aged previous suite container should be eligible")
	}

	freshContainer := staleContainer
	freshContainer.ID = "fresh-container"
	freshContainer.Created = now.Add(-time.Minute).Unix()
	if previousSuiteContainerCleanupEligible(activeEnv, freshContainer, now) {
		t.Fatal("fresh previous suite without completed cleanup must not be eligible")
	}
}

func staleSuiteContainer(id string, now time.Time) dockercontainer.Summary {
	return dockercontainer.Summary{
		ID:      id,
		Created: now.Add(-staleSuiteContainerAge - time.Second).Unix(),
		Labels: map[string]string{
			testServiceLabelManaged: testServiceManagedValue,
			testServiceLabelSuiteID: "stale-suite",
			testServiceLabelRunID:   "stale-run",
			testServiceLabelService: suiteservices.ServicePostgres,
		},
	}
}

func TestCleanupPreviousSuiteServiceContainersRemovesEligibleContainer(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{staleSuiteContainer("removed-container", now)},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("remove eligible stale container: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 1 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
	if len(cli.removeIDs) != 1 || cli.removeIDs[0] != "removed-container" {
		t.Fatalf("expected one remove attempt, got %#v", cli.removeIDs)
	}
}

func TestCleanupPreviousSuiteServiceContainersAcceptsNotFound(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{staleSuiteContainer("gone-container", now)},
		removeErrs: map[string]error{
			"gone-container": errors.New("Error response from daemon: No such container: gone-container"),
		},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("not-found stale-container removal must not fail preflight: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 1 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
}

func TestCleanupPreviousSuiteServiceContainersAcceptsConcurrentRemoval(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"

	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{{
			ID:      "fc621a862739f3b8c8a8fe32207ce9e7c9e9543f677377d0eb10dcf4e806ffc3",
			Created: now.Add(-staleSuiteContainerAge - time.Second).Unix(),
			Labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "stale-suite",
				testServiceLabelRunID:   "stale-run",
				testServiceLabelService: suiteservices.ServicePostgres,
			},
		}},
		removeErrs: map[string]error{
			"fc621a862739f3b8c8a8fe32207ce9e7c9e9543f677377d0eb10dcf4e806ffc3": errors.New("Error response from daemon: removal of container fc621a862739f3b8c8a8fe32207ce9e7c9e9543f677377d0eb10dcf4e806ffc3 is already in progress"),
		},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("concurrent stale-container removal must not fail preflight: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 1 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
	if len(cli.removeIDs) != 1 {
		t.Fatalf("expected one remove attempt, got %#v", cli.removeIDs)
	}
}

func TestCleanupPreviousSuiteServiceContainersTimeoutThenGone(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{staleSuiteContainer("timeout-gone", now)},
		removeErrs: map[string]error{
			"timeout-gone": context.DeadlineExceeded,
		},
		inspectErrs: map[string]error{
			"timeout-gone": errors.New("Error response from daemon: No such container: timeout-gone"),
		},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("timeout followed by not-found recheck must not fail preflight: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 1 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
	if len(cli.inspectIDs) != 1 || cli.inspectIDs[0] != "timeout-gone" {
		t.Fatalf("expected one inspect recheck, got %#v", cli.inspectIDs)
	}
}

func TestCleanupPreviousSuiteServiceContainersTimeoutThenRemovingOrDead(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"

	cases := []struct {
		name    string
		state   dockercontainer.State
		stateID string
	}{
		{
			name:    "removing",
			state:   dockercontainer.State{Status: dockercontainer.ContainerState("removing")},
			stateID: "timeout-removing",
		},
		{
			name:    "dead flag",
			state:   dockercontainer.State{Status: dockercontainer.ContainerState("exited"), Dead: true},
			stateID: "timeout-dead",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli := &fakeSuiteContainerClient{
				items: []dockercontainer.Summary{staleSuiteContainer(tc.stateID, now)},
				removeErrs: map[string]error{
					tc.stateID: context.DeadlineExceeded,
				},
				inspectResults: map[string]dockerclient.ContainerInspectResult{
					tc.stateID: {Container: dockercontainer.InspectResponse{State: &tc.state}},
				},
			}

			summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
			if err != nil {
				t.Fatalf("timeout followed by %s recheck must not fail preflight: %v", tc.name, err)
			}
			if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 1 {
				t.Fatalf("unexpected cleanup summary: %#v", summary)
			}
		})
	}
}

func TestCleanupPreviousSuiteServiceContainersTimeoutStillRunningFails(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	runningState := dockercontainer.State{Status: dockercontainer.ContainerState("running"), Running: true}
	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{staleSuiteContainer("timeout-running", now)},
		removeErrs: map[string]error{
			"timeout-running": context.DeadlineExceeded,
		},
		inspectResults: map[string]dockerclient.ContainerInspectResult{
			"timeout-running": {Container: dockercontainer.InspectResponse{State: &runningState}},
		},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err == nil {
		t.Fatal("expected timeout with still-running container to fail preflight")
	}
	if !strings.Contains(err.Error(), "state=running") {
		t.Fatalf("unexpected timeout recheck error: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary on fatal error: %#v", summary)
	}
}

func TestCleanupPreviousSuiteServiceContainersRequiresOwnershipProof(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	unproven := staleSuiteContainer("unproven-container", now)
	delete(unproven.Labels, testServiceLabelSuiteID)
	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{unproven},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("unproven stale-container candidate must be skipped: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
	if len(cli.removeIDs) != 0 || len(cli.inspectIDs) != 0 {
		t.Fatalf("unproven container must not be removed or inspected, remove=%#v inspect=%#v", cli.removeIDs, cli.inspectIDs)
	}
}

func TestCleanupPreviousSuiteServiceContainersReportsListFailure(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"
	cli := &fakeSuiteContainerClient{listErr: errors.New("docker unavailable")}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err == nil {
		t.Fatal("expected Docker list failure to fail preflight")
	}
	if !strings.Contains(err.Error(), "list managed suite service containers") {
		t.Fatalf("unexpected list error: %v", err)
	}
	if summary.Scanned != 0 || summary.Removed != 0 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary on list failure: %#v", summary)
	}
}

func TestCleanupPreviousSuiteServiceContainersSkipsCurrentSuite(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"

	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{{
			ID:      "active-container",
			Created: now.Add(-staleSuiteContainerAge - time.Hour).Unix(),
			Labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "active-suite",
				testServiceLabelRunID:   "wrapper-tests",
				testServiceLabelService: suiteservices.ServicePostgres,
			},
		}},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err != nil {
		t.Fatalf("cleanup current-suite container: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary: %#v", summary)
	}
	if len(cli.removeIDs) != 0 {
		t.Fatalf("current-suite container must not be removed, got %#v", cli.removeIDs)
	}
}

func TestCleanupPreviousSuiteServiceContainersReportsFatalRemove(t *testing.T) {
	deps := defaultTestDependencies(t)
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	activeEnv := cloneEnv(deps.env)
	activeEnv[suiteservices.ActiveEnv] = "1"
	activeEnv[suiteservices.SuiteIDEnv] = "active-suite"

	cli := &fakeSuiteContainerClient{
		items: []dockercontainer.Summary{{
			ID:      "fatal-container",
			Created: now.Add(-staleSuiteContainerAge - time.Second).Unix(),
			Labels: map[string]string{
				testServiceLabelManaged: testServiceManagedValue,
				testServiceLabelSuiteID: "stale-suite",
				testServiceLabelRunID:   "stale-run",
				testServiceLabelService: suiteservices.ServicePostgres,
			},
		}},
		removeErrs: map[string]error{"fatal-container": errors.New("permission denied")},
	}

	summary, err := cleanupPreviousSuiteServiceContainers(context.Background(), cli, activeEnv, now)
	if err == nil {
		t.Fatal("expected fatal stale-container remove error")
	}
	if !strings.Contains(err.Error(), "remove stale suite service container fatal-contai") {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	if summary.Scanned != 1 || summary.Removed != 0 || summary.Deferred != 0 {
		t.Fatalf("unexpected cleanup summary on fatal error: %#v", summary)
	}
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

	cleanupOwnedServices(deps.dependencies, activeEnv, postgresService{}, objectStoreService{}, "", "succeeded", 0)

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
			startObjectStore: func(context.Context, map[string]string) (objectStoreService, error) {
				return objectStoreService{}, nil
			},
			startChild: func(argv []string, env map[string]string) (childProcess, error) {
				return fakeChild{}, nil
			},
			startReaper: func(string, map[string]string) error { return nil },
			preflightSuite: func(context.Context, map[string]string) (suitePreflightResult, error) {
				return suitePreflightResult{DockerEndpoint: "unix:///var/run/docker.sock", DockerOK: true, ReaperReady: true}, nil
			},
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

type fakeSuiteContainerClient struct {
	items          []dockercontainer.Summary
	listErr        error
	removeErrs     map[string]error
	removeIDs      []string
	inspectErrs    map[string]error
	inspectResults map[string]dockerclient.ContainerInspectResult
	inspectIDs     []string
}

func (c *fakeSuiteContainerClient) ContainerList(context.Context, dockerclient.ContainerListOptions) (dockerclient.ContainerListResult, error) {
	if c.listErr != nil {
		return dockerclient.ContainerListResult{}, c.listErr
	}
	return dockerclient.ContainerListResult{Items: c.items}, nil
}

func (c *fakeSuiteContainerClient) ContainerRemove(_ context.Context, containerID string, _ dockerclient.ContainerRemoveOptions) (dockerclient.ContainerRemoveResult, error) {
	c.removeIDs = append(c.removeIDs, containerID)
	if c.removeErrs != nil {
		if err := c.removeErrs[containerID]; err != nil {
			return dockerclient.ContainerRemoveResult{}, err
		}
	}
	return dockerclient.ContainerRemoveResult{}, nil
}

func (c *fakeSuiteContainerClient) ContainerInspect(_ context.Context, containerID string, _ dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error) {
	c.inspectIDs = append(c.inspectIDs, containerID)
	if c.inspectErrs != nil {
		if err := c.inspectErrs[containerID]; err != nil {
			return dockerclient.ContainerInspectResult{}, err
		}
	}
	if c.inspectResults != nil {
		if result, ok := c.inspectResults[containerID]; ok {
			return result, nil
		}
	}
	return dockerclient.ContainerInspectResult{}, errors.New("container inspect result not configured")
}

func stringContains(haystack string, needle string) bool {
	return strings.Contains(haystack, needle)
}

func loadTestEvents(t testing.TB, deps testDeps) []suiteservices.Event {
	t.Helper()
	env := cloneEnv(deps.env)
	env[suiteservices.SuiteIDEnv] = "suite-redaction"
	return loadTestEventsForEnv(t, env)
}

func requireLifecycleFailure(t testing.TB, deps testDeps, event string, failureClass string, failureReason string) {
	t.Helper()
	env := cloneEnv(deps.env)
	env[suiteservices.SuiteIDEnv] = "suite-redaction"
	records, err := suiteservices.ReadLifecycleEvents(env)
	if err != nil {
		t.Fatalf("read lifecycle events: %v", err)
	}
	for _, record := range records {
		if record.Event != event {
			continue
		}
		if record.FailureClass == nil || *record.FailureClass != failureClass {
			t.Fatalf("unexpected lifecycle failure class for %s: %#v", event, record)
		}
		if record.FailureReason == nil || *record.FailureReason != failureReason {
			t.Fatalf("unexpected lifecycle failure reason for %s: %#v", event, record)
		}
		return
	}
	t.Fatalf("missing lifecycle failure event %s in %#v", event, records)
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
