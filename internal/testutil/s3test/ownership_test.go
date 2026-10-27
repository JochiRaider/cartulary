package s3test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/aws/smithy-go"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	t.Run("bound_admission", checkBoundAttachment)
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	admission := &sharedAdmission{}

	t.Setenv(suiteservices.S3EndpointEnv, "127.0.0.1:9000")
	t.Setenv(suiteservices.S3AccessKeyEnv, "suite-access")
	t.Setenv(suiteservices.S3SecretKeyEnv, "suite-secret")
	t.Setenv(suiteservices.S3SecureEnv, "true")
	t.Setenv(suiteservices.S3ProbeBucketEnv, "ct-suite-readiness-probe")
	t.Setenv("SSL_CERT_FILE", filepath.Join(t.TempDir(), "fixture-root.pem"))

	verifyCalls := 0
	admission.verify = func(ctx context.Context, harness *Harness) error {
		verifyCalls++
		if harness.Endpoint != "127.0.0.1:9000" {
			t.Fatalf("unexpected attach endpoint: %q", harness.Endpoint)
		}
		if harness.probeBucket != "ct-suite-readiness-probe" {
			t.Fatalf("unexpected broker probe bucket: %q", harness.probeBucket)
		}
		return nil
	}

	harness, err := admission.start(context.Background())
	if err != nil {
		t.Fatalf("start attached harness: %v", err)
	}
	if verifyCalls != 1 {
		t.Fatalf("expected one attach verification call, got %d", verifyCalls)
	}
	if !harness.attached {
		t.Fatal("expected attached harness")
	}
	if !harness.Secure {
		t.Fatal("expected attach secure flag to honor the environment")
	}
}

func TestStartSharedRejectsMissingAttachment(t *testing.T) {
	t.Run("incomplete_and_independent_bindings", checkIncompleteAttachment)
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	admission := &sharedAdmission{}
	for _, name := range []string{suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv, suiteservices.S3SecureEnv} {
		t.Setenv(name, "")
	}
	harness, err := admission.start(t.Context())
	if err == nil || harness != nil || admission.harness != nil {
		t.Fatal("missing attachment must fail without publishing a harness", harness, err)
	}
}

func TestObjectStoreStartupMappingDeadline(t *testing.T) {
	strategy := objectStorePortWaitStrategy()
	if timeout := strategy.Timeout(); timeout == nil || *timeout != objectStorePortMappingTimeout {
		t.Fatalf("unexpected port mapping timeout: got %v want %v", timeout, objectStorePortMappingTimeout)
	}
}

func TestSeaweedFSS3AllowedOriginsDefaultSupportsLoopbackBrowserPortRange(t *testing.T) {
	t.Setenv("OBJECT_STORE_CORS_ALLOWED_ORIGINS", "")
	t.Setenv("OBJECT_STORE_CORS_ORIGIN", "")

	got := seaweedFSS3AllowedOrigins(nil)
	for _, want := range []string{
		"https://localhost:5173",
		"https://127.0.0.1:5173",
		"https://localhost:19000",
		"https://127.0.0.1:19000",
		"https://localhost:19199",
		"https://127.0.0.1:19199",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("default SeaweedFS S3 allowed origins missing %q in %q", want, got)
		}
	}
	for _, forbidden := range []string{"*", "http://", "https://0.0.0.0", "https://[::]"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("default SeaweedFS S3 allowed origins must stay exact and loopback-scoped, got %q", got)
		}
	}
}

func TestOwnedObjectStoreAppliesContainerConfiguration(t *testing.T) {
	t.Run("settlement_and_borrower_lifetime", checkOwnedStoreSettlement)
	ops := stubOwnedObjectStoreStartup(t)

	var gotLabels map[string]string
	var gotCommand []string
	ops.container = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		gotLabels = req.Labels
		gotCommand = append([]string(nil), req.Cmd...)
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
		}, nil
	}
	ops.ready = func(context.Context, *Harness) error { return nil }

	labels := map[string]string{"cartulary.test": "suite"}
	owner, err := startOwned(context.Background(), StartOptions{Labels: labels}, ops)
	if err != nil {
		t.Fatalf("start labeled object store: %v", err)
	}
	t.Cleanup(func() {
		if err := owner.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	if gotLabels["cartulary.test"] != "suite" {
		t.Fatalf("container labels not applied: %#v", gotLabels)
	}
	var peers []string
	for _, argument := range gotCommand {
		if strings.HasPrefix(argument, "-master.peers=") {
			peers = append(peers, argument)
		}
	}
	if !reflect.DeepEqual(peers, []string{"-master.peers=none"}) {
		t.Fatalf("disposable object store must declare exactly one standalone master: %v", gotCommand)
	}
}

func TestOwnedObjectStoreReadinessTimeoutIsTerminalForTheStartedLane(t *testing.T) {
	ops := stubOwnedObjectStoreStartup(t)

	starts := 0
	terminations := 0
	readinessChecks := 0
	var events []testcontainersx.StartEvent
	ops.container = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	ops.ready = func(ctx context.Context, harness *Harness) error {
		readinessChecks++
		return &objectStoreReadinessError{
			Stage:           "put",
			Attempts:        3,
			CleanupOutcome:  "completed",
			LastErr:         context.DeadlineExceeded,
			DeadlineExpired: true,
		}
	}

	_, err := startOwned(context.Background(), StartOptions{
		Observer: func(event testcontainersx.StartEvent) {
			events = append(events, event)
		},
	}, ops)
	if err == nil {
		t.Fatal("expected terminal readiness failure")
	}
	if starts != 1 {
		t.Fatalf("expected one container lane, got %d", starts)
	}
	if readinessChecks != 1 {
		t.Fatalf("expected one readiness controller, got %d", readinessChecks)
	}
	if terminations != 1 {
		t.Fatalf("expected terminal lane to be cleaned once, got %d", terminations)
	}
	if observedObjectStoreRetry(events) {
		t.Fatalf("readiness expiry must not schedule a replacement lane, got %#v", events)
	}
	diagnostic, ok := ReadinessDiagnosticFromError(err)
	if !ok || diagnostic.CleanupOutcome != "completed" {
		t.Fatalf("owned readiness cleanup diagnostic = %#v present=%t", diagnostic, ok)
	}
}

func TestOwnedObjectStoreCleanupFailureRemainsSecondaryTypedEvidence(t *testing.T) {
	ops := stubOwnedObjectStoreStartup(t)

	const rawCleanupIdentity = "container b56c88f3a1b0 via unix:///var/run/docker.sock"
	cleanupErr := errors.New(rawCleanupIdentity)
	terminations := 0
	ops.container = func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
			terminate: func(context.Context) error {
				terminations++
				return cleanupErr
			},
		}, nil
	}
	ops.ready = func(context.Context, *Harness) error {
		return &objectStoreReadinessError{
			Phase:           "initial_lane",
			Stage:           "list",
			Outcome:         "deadline_expired",
			Attempts:        33,
			AttemptTimeouts: 3,
			Elapsed:         120 * time.Second,
			CleanupOutcome:  "not_needed",
			CauseCounts: map[string]int{
				"operation_timeout":     3,
				"transport_unreachable": 30,
			},
			LastErr:         context.DeadlineExceeded,
			DeadlineExpired: true,
		}
	}

	partial, err := startOwned(context.Background(), StartOptions{}, ops)
	if err == nil || partial == nil {
		t.Fatal("expected terminal readiness failure")
	}
	t.Cleanup(func() { _ = os.RemoveAll(partial.tlsDirectory) })
	for range 2 {
		if !errors.Is(partial.Close(t.Context()), cleanupErr) {
			t.Fatal("partial cleanup receipt lost")
		}
	}
	if terminations != 1 {
		t.Fatal("partial startup cleanup retried", terminations)
	}
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("secondary cleanup error was not preserved internally: %v", err)
	}
	diagnostic, present := ReadinessDiagnosticFromError(err)
	if !present || diagnostic.Outcome != "deadline_expired" || diagnostic.CleanupOutcome != "failed" {
		t.Fatalf("cleanup failure replaced or failed to update readiness evidence: present=%t diagnostic=%#v", present, diagnostic)
	}
	message, present := ReadinessDiagnosticText(err)
	if !present {
		t.Fatal("expected bounded readiness diagnostic text")
	}
	for _, forbidden := range []string{rawCleanupIdentity, "docker.sock", "b56c88f3a1b0"} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("bounded readiness diagnostic retained %q: %q", forbidden, message)
		}
	}
	for _, required := range []string{
		"phase=initial_lane",
		"stage=list",
		"outcome=deadline_expired",
		"attempts=33",
		"cause_counts=operation_timeout:3,transport_unreachable:30",
		"cleanup=failed",
	} {
		if !strings.Contains(message, required) {
			t.Fatalf("bounded readiness diagnostic omitted %q: %q", required, message)
		}
	}
}

func TestOwnedObjectStoreRetriesOnlyPreReadinessStartupAndUsesFreshReadinessContext(t *testing.T) {
	ops := stubOwnedObjectStoreStartup(t)

	starts := 0
	readinessChecks := 0
	var events []testcontainersx.StartEvent
	ops.container = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		if starts == 1 {
			return nil, errors.New("docker.sock connection reset by peer")
		}
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
		}, nil
	}
	ops.ready = func(ctx context.Context, harness *Harness) error {
		readinessChecks++
		if _, hasDeadline := ctx.Deadline(); hasDeadline {
			t.Fatal("readiness reused the bounded startup-attempt context")
		}
		return nil
	}

	owner, err := startOwned(context.Background(), StartOptions{
		AttemptTimeout: time.Minute,
		Observer: func(event testcontainersx.StartEvent) {
			events = append(events, event)
		},
	}, ops)
	if err != nil {
		t.Fatalf("start object store after transient pre-readiness failure: %v", err)
	}
	t.Cleanup(func() {
		if err := owner.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	if starts != 2 || readinessChecks != 1 {
		t.Fatalf("startup/readiness phase counts = %d/%d, want 2/1", starts, readinessChecks)
	}
	if !observedObjectStoreRetry(events) {
		t.Fatalf("pre-readiness Docker failure did not use declared retry: %#v", events)
	}
}

func TestOwnedObjectStoreDoesNotRetryAuthenticationReadinessFailure(t *testing.T) {
	ops := stubOwnedObjectStoreStartup(t)

	starts := 0
	terminations := 0
	ops.container = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	ops.ready = func(ctx context.Context, harness *Harness) error {
		return &objectStoreReadinessError{
			Phase:          "initial_lane",
			Stage:          "create_namespace",
			Outcome:        "capability_rejected",
			Attempts:       1,
			CleanupOutcome: "completed",
			LastErr: &smithy.GenericAPIError{
				Code:    "AccessDenied",
				Message: "Access Denied",
			},
		}
	}

	_, err := startOwned(context.Background(), StartOptions{}, ops)
	if err == nil {
		t.Fatal("expected authentication readiness failure")
	}
	if starts != 1 {
		t.Fatalf("expected no retry for auth failure, got %d starts", starts)
	}
	if terminations != 1 {
		t.Fatalf("expected failed auth attempt to terminate its container once, got %d", terminations)
	}

	diagnostic, ok := ReadinessDiagnosticFromError(err)
	if !ok || diagnostic.Outcome != "capability_rejected" || diagnostic.AttemptCount != 1 {
		t.Fatalf("unexpected capability readiness diagnostic: %#v present=%t", diagnostic, ok)
	}
}

func observedObjectStoreRetry(events []testcontainersx.StartEvent) bool {
	sawRetryableAttempt := false
	sawRetryScheduled := false
	for _, event := range events {
		if event.Type == testcontainersx.StartEventAttemptEnd && event.Attempt == 1 && event.Retryable && event.Status == "fail" {
			sawRetryableAttempt = true
		}
		if event.Type == testcontainersx.StartEventRetryScheduled && event.Attempt == 1 {
			sawRetryScheduled = true
		}
	}
	return sawRetryableAttempt && sawRetryScheduled
}

func stubOwnedObjectStoreStartup(t testing.TB) startupOperations {
	t.Helper()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	ops := defaultStartupOperations()
	ops.preflight = func(context.Context) (string, error) { return "unix:///var/run/docker.sock", nil }
	ops.sleep = func(context.Context, time.Duration) error { return nil }
	return ops
}

type fakeObjectStoreContainer struct {
	testcontainers.Container
	host      string
	port      network.Port
	state     *dockercontainer.State
	stateErr  error
	terminate func(context.Context) error
	start     func(context.Context) error
}

func (c fakeObjectStoreContainer) Start(ctx context.Context) error {
	if c.start != nil {
		return c.start(ctx)
	}
	return nil
}

func (c fakeObjectStoreContainer) Host(context.Context) (string, error) {
	return c.host, nil
}

func (c fakeObjectStoreContainer) MappedPort(context.Context, string) (network.Port, error) {
	return c.port, nil
}

func (c fakeObjectStoreContainer) State(context.Context) (*dockercontainer.State, error) {
	if c.stateErr != nil {
		return nil, c.stateErr
	}
	if c.state != nil {
		return c.state, nil
	}
	return &dockercontainer.State{Status: dockercontainer.StateRunning, Running: true}, nil
}

func (c fakeObjectStoreContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	if c.terminate == nil {
		return nil
	}
	return c.terminate(ctx)
}

func checkOwnedStoreSettlement(t *testing.T) {
	for _, stage := range []string{"before", "publication", "cancelled_acquisition", "success"} {
		t.Run(stage, func(t *testing.T) {
			ops := stubOwnedObjectStoreStartup(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			cause := errors.New("publication failure")
			var order []string
			var owner *OwnedStore
			ops.container = func(_ context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
				order = append(order, "acquire")
				if req.Started {
					t.Fatal("physical identity must be published before start")
				}
				for _, file := range req.Files {
					if file.FileMode != 0o600 {
						t.Fatal("private identity permissions", file.FileMode)
					}
				}
				if stage == "cancelled_acquisition" {
					cancel()
				}
				return fakeObjectStoreContainer{host: "127.0.0.1", port: network.MustParsePort("8333/tcp"),
					start:     func(context.Context) error { order = append(order, "start"); return nil },
					terminate: func(context.Context) error { order = append(order, "terminate"); return nil },
				}, nil
			}
			ops.ready = func(_ context.Context, h *Harness) error {
				if h.Endpoint != "127.0.0.1:8333" || !h.Secure || h.RootCertificatePath() == "" {
					t.Fatal("readiness lacks verified mapped endpoint")
				}
				order = append(order, "ready")
				return nil
			}
			result, err := startOwned(ctx, StartOptions{
				BeforeAcquire: func(o *OwnedStore) error {
					owner = o
					order = append(order, "before")
					if stage == "before" {
						return cause
					}
					return nil
				},
				Acquired: func(o *OwnedStore) error {
					order = append(order, "publish")
					if o.Container == nil {
						t.Fatal("missing partial owner")
					}
					if stage == "publication" {
						return cause
					}
					return nil
				},
				Settled: func(*OwnedStore) error { order = append(order, "settle"); return nil },
			}, ops)
			if stage == "success" {
				if err != nil || result == nil {
					t.Fatal("owned start", err)
				}
				poisoned := result.Fixture.packageBucket("poisoned")
				poisoned.unusable = cause
				if err := result.Fixture.Close(t.Context()); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(result.tlsDirectory); err != nil || poisoned.unusable != cause {
					t.Fatal("borrower close altered owner or quarantine", err)
				}
				if !reflect.DeepEqual(order, []string{"before", "acquire", "publish", "start", "ready"}) {
					t.Fatal("borrower terminated owner or startup order changed", order)
				}
				for range 2 {
					if err := result.Close(t.Context()); err != nil {
						t.Fatal(err)
					}
				}
				if !reflect.DeepEqual(order, []string{"before", "acquire", "publish", "start", "ready", "terminate"}) {
					t.Fatal("physical close repeated", order)
				}
			} else {
				if err == nil || result != nil {
					t.Fatal("failed acquisition published owner", result, err)
				}
				expected := []string{"before"}
				if stage != "before" {
					expected = append(expected, "acquire", "publish", "terminate", "settle")
				}
				if !reflect.DeepEqual(order, expected) {
					t.Fatal("partial ownership ordering", order, expected)
				}
				if stage == "cancelled_acquisition" && !errors.Is(err, context.Canceled) {
					t.Fatal("lost caller cancellation", err)
				}
				if stage != "cancelled_acquisition" && !errors.Is(err, cause) {
					t.Fatal("lost publication cause", err)
				}
			}
			if owner == nil {
				t.Fatal("missing owner intent")
			}
			if _, err := os.Stat(owner.tlsDirectory); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("settled TLS identity remains", err)
			}
		})
	}
}

func attachmentTestEnvironment(t *testing.T) map[string]string {
	t.Helper()
	env := map[string]string{
		suiteservices.HarnessServiceDependenciesEnv: "object_store",
		suiteservices.S3EndpointEnv:                 "127.0.0.1:9000",
		suiteservices.S3AccessKeyEnv:                "fixture-access",
		suiteservices.S3SecretKeyEnv:                "fixture-secret",
		suiteservices.S3SecureEnv:                   "true",
		suiteservices.S3ProbeBucketEnv:              "ct-suite-probe",
		suiteservices.SuiteIDEnv:                    "attachment-tests",
		suiteservices.TargetEnv:                     "test-slice",
		"SSL_CERT_FILE":                             filepath.Join(t.TempDir(), "fixture-root.pem"),
		"CARTULARY_TEST_RESULTS_DIR":                t.TempDir(),
		"CARTULARY_TEST_RUN_ID":                     "attachment-tests",
	}
	for key, value := range env {
		t.Setenv(key, value)
	}
	return env
}

func checkBoundAttachment(t *testing.T) {
	t.Run("changed_binding", func(t *testing.T) {
		for _, key := range []string{
			suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv,
			suiteservices.S3SecureEnv, suiteservices.S3ProbeBucketEnv, "SSL_CERT_FILE",
			suiteservices.HarnessServiceDependenciesEnv, suiteservices.SuiteIDEnv, suiteservices.TargetEnv,
			suiteservices.SuiteRuntimeRootEnv, suiteservices.SuiteRuntimeLeaseIDEnv, suiteservices.SuiteRuntimeRunIDEnv,
			suiteservices.CallModeEnv, suiteservices.LifecycleModeEnv, suiteservices.PersistentBorrowerEnv,
			"CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID",
		} {
			t.Run(key, func(t *testing.T) {
				attachmentTestEnvironment(t)
				calls := 0
				admission := &sharedAdmission{verify: func(context.Context, *Harness) error { calls++; return nil }}
				first, err := admission.start(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				again, err := admission.start(t.Context())
				if err != nil || again != first {
					t.Fatal("same binding not reused", err)
				}
				t.Setenv(key, "changed")
				if got, err := admission.start(t.Context()); err == nil || got != nil {
					t.Fatal("changed binding reused cached admission")
				}
				if calls != 1 {
					t.Fatal("changed binding caused new admission", calls)
				}
			})
		}
	})
	t.Run("terminal_failure", func(t *testing.T) {
		attachmentTestEnvironment(t)
		failure := errors.New("terminal readiness failure")
		calls := 0
		admission := &sharedAdmission{verify: func(context.Context, *Harness) error { calls++; return failure }}
		for range 2 {
			h, err := admission.start(t.Context())
			if h != nil || !errors.Is(err, failure) {
				t.Fatal("lost terminal failure", err)
			}
		}
		if calls != 1 {
			t.Fatal("terminal admission retried", calls)
		}
	})
	t.Run("cancelled_waiter", func(t *testing.T) {
		attachmentTestEnvironment(t)
		entered, release := make(chan struct{}), make(chan struct{})
		admission := &sharedAdmission{verify: func(ctx context.Context, _ *Harness) error { close(entered); <-release; return ctx.Err() }}
		first := make(chan error, 1)
		go func() { _, err := admission.start(t.Context()); first <- err }()
		<-entered
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		waiting := &observedWaitContext{Context: ctx, observed: make(chan struct{})}
		waiter := make(chan error, 1)
		go func() { _, err := admission.start(waiting); waiter <- err }()
		select {
		case <-waiting.observed:
		case <-time.After(time.Second):
			t.Error("waiter did not reach cancellable admission waiting")
		}
		cancel()
		select {
		case err := <-waiter:
			if !errors.Is(err, context.Canceled) {
				t.Error("waiter cancellation lost", err)
			}
		case <-time.After(time.Second):
			t.Error("cancelled waiter blocked behind another admission")
		}
		close(release)
		if err := <-first; err != nil {
			t.Fatal("waiter cancelled another caller's admission", err)
		}
	})
	t.Run("publication_failure", func(t *testing.T) {
		attachmentTestEnvironment(t)
		path := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CARTULARY_TEST_RESULTS_DIR", path)
		calls := 0
		admission := &sharedAdmission{verify: func(context.Context, *Harness) error { calls++; return nil }}
		for range 2 {
			if h, err := admission.start(t.Context()); h != nil || err == nil {
				t.Fatal("unpublished admission reported success")
			}
		}
		if calls != 1 {
			t.Fatal("publication failure was retried", calls)
		}
	})
	t.Run("saved_dependencies_and_destination", func(t *testing.T) {
		original := attachmentTestEnvironment(t)
		original[suiteservices.S3EndpointEnv] = "invalid endpoint"
		t.Setenv(suiteservices.S3EndpointEnv, original[suiteservices.S3EndpointEnv])
		admission := &sharedAdmission{verify: func(context.Context, *Harness) error { return nil }}
		h, err := admission.start(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
		t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
		_, err = h.BootstrapBucket(t.Context(), "saved")
		var dependencyErr *suiteservices.ServiceDependencyError
		if err == nil || errors.As(err, &dependencyErr) {
			t.Fatal("operation reread ambient dependency inputs", err)
		}
		ledger, ok, err := suiteservices.CurrentResourceLedger(original)
		if err != nil || !ok || len(ledger.Buckets) != 1 {
			t.Fatal("reservation redirected from saved destination", ok, err)
		}
	})
}

// Done is observed only once the admission path reaches its wait select, after
// the early cancellation check. This keeps the cancellation witness deterministic.
type observedWaitContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
}

func (c *observedWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.observed) })
	return c.Context.Done()
}

func checkIncompleteAttachment(t *testing.T) {
	for _, key := range []string{suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv, suiteservices.S3SecureEnv, suiteservices.S3ProbeBucketEnv, "SSL_CERT_FILE"} {
		t.Run(key, func(t *testing.T) {
			env := attachmentTestEnvironment(t)
			delete(env, key)
			admission := &sharedAdmission{verify: func(context.Context, *Harness) error { t.Fatal("incomplete binding reached readiness"); return nil }}
			if h, err := admission.startWithInputs(t.Context(), env); h != nil || err == nil {
				t.Fatal("incomplete binding admitted")
			}
		})
	}
	for _, declaration := range []string{"", "postgres", "unknown", "object_store,object_store", "postgres,object_store"} {
		t.Run("dependency/"+declaration, func(t *testing.T) {
			env := attachmentTestEnvironment(t)
			env[suiteservices.HarnessServiceDependenciesEnv] = declaration
			admission := &sharedAdmission{verify: func(context.Context, *Harness) error { t.Fatal("invalid declaration reached readiness"); return nil }}
			_, err := admission.startWithInputs(t.Context(), env)
			var dependencyErr *suiteservices.ServiceDependencyError
			if !errors.As(err, &dependencyErr) {
				t.Fatal("invalid dependency binding accepted", err)
			}
		})
	}
	t.Run("independent_copied_bindings", func(t *testing.T) {
		first := attachmentTestEnvironment(t)
		second := make(map[string]string)
		for k, v := range first {
			second[k] = v
		}
		second[suiteservices.S3EndpointEnv] = "127.0.0.1:9001"
		a, b := &sharedAdmission{verify: func(context.Context, *Harness) error { return nil }}, &sharedAdmission{verify: func(context.Context, *Harness) error { return nil }}
		h1, err := a.startWithInputs(t.Context(), first)
		if err != nil {
			t.Fatal(err)
		}
		h2, err := b.startWithInputs(t.Context(), second)
		if err != nil {
			t.Fatal(err)
		}
		first[suiteservices.S3EndpointEnv] = "mutated"
		second[suiteservices.S3EndpointEnv] = "mutated"
		t.Setenv(suiteservices.S3EndpointEnv, "ambient-conflict")
		if h1.Endpoint != "127.0.0.1:9000" || h2.Endpoint != "127.0.0.1:9001" || h1.binding.inputs[suiteservices.S3EndpointEnv] != h1.Endpoint {
			t.Fatal("binding retained mutable or ambient inputs")
		}
		recovered, err := harnessFromBinding(h1.binding)
		if err != nil || recovered.Endpoint != h1.Endpoint || recovered.binding.inputs["CARTULARY_TEST_RESULTS_DIR"] != h1.binding.inputs["CARTULARY_TEST_RESULTS_DIR"] {
			t.Fatal("recovery binding changed", err)
		}
	})
}

func TestCreateBucketRejectsOmittedServiceBeforeClientAcquisition(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")

	err := (&Harness{binding: bindFixture(processFixtureInputs()), Endpoint: "invalid endpoint"}).CreateBucket(context.Background(), "guarded")
	var dependencyErr *suiteservices.ServiceDependencyError
	if !errors.As(err, &dependencyErr) || dependencyErr.Service != "object_store" || dependencyErr.Reason != "omitted" {
		t.Fatalf("unexpected dependency error: %#v", err)
	}
}
