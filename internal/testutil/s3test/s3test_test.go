package s3test

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	resetSharedHarness(t)

	t.Setenv(suiteservices.S3EndpointEnv, "127.0.0.1:9000")
	t.Setenv(suiteservices.S3AccessKeyEnv, "suite-access")
	t.Setenv(suiteservices.S3SecretKeyEnv, "suite-secret")
	t.Setenv(suiteservices.S3SecureEnv, "true")

	oldStart := startOwnedHarness
	oldVerify := verifyAttachedFn
	t.Cleanup(func() {
		startOwnedHarness = oldStart
		verifyAttachedFn = oldVerify
	})

	startCalls := 0
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return nil, errors.New("owned harness should not start when attach env is present")
	}

	verifyCalls := 0
	verifyAttachedFn = func(ctx context.Context, harness *Harness) error {
		verifyCalls++
		if harness.Endpoint != "127.0.0.1:9000" {
			t.Fatalf("unexpected attach endpoint: %q", harness.Endpoint)
		}
		return nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start attached harness: %v", err)
	}
	if startCalls != 0 {
		t.Fatalf("expected attach mode to skip owned startup, got %d calls", startCalls)
	}
	if verifyCalls != 1 {
		t.Fatalf("expected one attach verification call, got %d", verifyCalls)
	}
	if !harness.attached {
		t.Fatal("expected attached harness")
	}
	if harness.Container != nil {
		t.Fatal("expected attach mode to avoid creating a testcontainer")
	}
	if !harness.Secure {
		t.Fatal("expected attach secure flag to honor the environment")
	}
}

func TestStartSharedFallsBackToOwnedHarnessWhenAttachEnvIsAbsent(t *testing.T) {
	resetSharedHarness(t)
	t.Setenv(suiteservices.S3EndpointEnv, "")
	t.Setenv(suiteservices.S3AccessKeyEnv, "")
	t.Setenv(suiteservices.S3SecretKeyEnv, "")
	t.Setenv(suiteservices.S3SecureEnv, "")

	oldStart := startOwnedHarness
	t.Cleanup(func() {
		startOwnedHarness = oldStart
	})

	startCalls := 0
	want := &Harness{
		Endpoint:    "127.0.0.1:9000",
		AccessKey:   "suite-access",
		SecretKey:   "suite-secret",
		suiteHash:   "suitehash",
		processHash: "processid",
	}
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return want, nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start shared harness: %v", err)
	}
	if startCalls != 1 {
		t.Fatalf("expected one owned startup, got %d", startCalls)
	}
	if harness != want {
		t.Fatal("expected StartShared to return the owned harness")
	}
	if !harness.shared {
		t.Fatal("expected owned harness returned by StartShared to be marked shared")
	}
}

func TestBucketNamesAreUniqueAcrossSimulatedProcesses(t *testing.T) {
	first := &Harness{suiteHash: "suitehash", processHash: "procaaaa"}
	second := &Harness{suiteHash: "suitehash", processHash: "procbbbb"}

	seen := make(map[string]struct{})
	validBucket := regexp.MustCompile(`^[a-z0-9-]+$`)
	for _, harness := range []*Harness{first, second} {
		for i := 0; i < 8; i++ {
			name := harness.nextBucketName("Prefix With Spaces_and_symbols!")
			if len(name) > 63 {
				t.Fatalf("bucket name exceeds DNS length limit: %q", name)
			}
			if !validBucket.MatchString(name) {
				t.Fatalf("bucket name must be dns compatible, got %q", name)
			}
			if name[0] == '-' || name[len(name)-1] == '-' {
				t.Fatalf("bucket name must start and end with an alphanumeric character, got %q", name)
			}
			if _, exists := seen[name]; exists {
				t.Fatalf("bucket name collision detected: %q", name)
			}
			seen[name] = struct{}{}
		}
	}
}

func TestObjectStoreContainerWaitStrategyOnlyWaitsForPortMapping(t *testing.T) {
	strategy := objectStorePortWaitStrategy()
	if got := strategy.String(); !strings.Contains(got, "to be mapped") {
		t.Fatalf("expected mapped-port-only wait strategy, got %q", got)
	}
	if timeout := strategy.Timeout(); timeout == nil || *timeout != objectStorePortMappingTimeout {
		t.Fatalf("unexpected port mapping timeout: got %v want %v", timeout, objectStorePortMappingTimeout)
	}
}

func TestOwnedObjectStoreAppliesContainerLabels(t *testing.T) {
	stubOwnedObjectStoreStartup(t)

	var gotLabels map[string]string
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		gotLabels = req.Labels
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("8333/tcp"),
		}, nil
	}
	waitReadyFn = func(context.Context, *Harness) error { return nil }

	labels := map[string]string{"cartulary.test": "suite"}
	if _, err := StartOwnedWithLabels(context.Background(), labels); err != nil {
		t.Fatalf("start labeled object store: %v", err)
	}
	if gotLabels["cartulary.test"] != "suite" {
		t.Fatalf("container labels not applied: %#v", gotLabels)
	}
}

func TestOwnedObjectStoreRetriesReadinessTimeoutAndTerminatesFailedAttempt(t *testing.T) {
	stubOwnedObjectStoreStartup(t)

	starts := 0
	terminations := 0
	readinessChecks := 0
	var events []testcontainersx.StartEvent
	ports := []network.Port{
		network.MustParsePort("8333/tcp"),
		network.MustParsePort("8334/tcp"),
	}
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		return fakeObjectStoreContainer{
			host: "127.0.0.1",
			port: ports[starts-1],
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	waitReadyFn = func(ctx context.Context, harness *Harness) error {
		readinessChecks++
		if readinessChecks == 1 {
			return &objectStoreReadinessError{
				LastErr:         context.DeadlineExceeded,
				DeadlineExpired: true,
			}
		}
		return nil
	}

	harness, err := StartOwnedWithOptions(context.Background(), StartOptions{
		Observer: func(event testcontainersx.StartEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if starts != 2 {
		t.Fatalf("expected two container attempts, got %d", starts)
	}
	if readinessChecks != 2 {
		t.Fatalf("expected two readiness checks, got %d", readinessChecks)
	}
	if terminations != 1 {
		t.Fatalf("expected failed readiness attempt to terminate its container once, got %d", terminations)
	}
	if harness.Endpoint != "127.0.0.1:8334" {
		t.Fatalf("expected second attempt endpoint, got %q", harness.Endpoint)
	}
	if !observedObjectStoreRetry(events) {
		t.Fatalf("expected observer to record retryable attempt and retry decision, got %#v", events)
	}
}

func TestOwnedObjectStoreDoesNotRetryAuthenticationReadinessFailure(t *testing.T) {
	stubOwnedObjectStoreStartup(t)

	starts := 0
	terminations := 0
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
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
	waitReadyFn = func(ctx context.Context, harness *Harness) error {
		return &objectStoreReadinessError{
			LastErr: minio.ErrorResponse{
				Code:    "AccessDenied",
				Message: "Access Denied",
			},
			DeadlineExpired: true,
		}
	}

	_, err := startHarness(context.Background(), nil)
	if err == nil {
		t.Fatal("expected authentication readiness failure")
	}
	if starts != 1 {
		t.Fatalf("expected no retry for auth failure, got %d starts", starts)
	}
	if terminations != 1 {
		t.Fatalf("expected failed auth attempt to terminate its container once, got %d", terminations)
	}

	var startFailure *testcontainersx.StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if startFailure.Retryable {
		t.Fatal("auth readiness failure must not be retryable")
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != testcontainersx.DefaultMaxAttempts {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
}

func TestHarnessStartsSeaweedFSS3AndRoundTripsObjects(t *testing.T) {
	harness := Start(t)

	bucket, err := harness.BootstrapBucket(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	payload := []byte("cartulary bootstrap object")
	got, err := harness.RoundTrip(context.Background(), bucket, "proof.txt", payload)
	if err != nil {
		t.Fatalf("round trip object: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected object payload: %q", got)
	}
}

func TestCleanupPrefixPreservesSiblingObjects(t *testing.T) {
	harness := Start(t)

	bucket, err := harness.BootstrapBucket(context.Background(), "prefix-cleanup")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	if _, err := harness.RoundTrip(context.Background(), bucket, "current/proof.txt", []byte("current")); err != nil {
		t.Fatalf("write current prefix object: %v", err)
	}
	if _, err := harness.RoundTrip(context.Background(), bucket, "sibling/proof.txt", []byte("sibling")); err != nil {
		t.Fatalf("write sibling prefix object: %v", err)
	}
	if err := harness.CleanupPrefix(context.Background(), bucket, "current/"); err != nil {
		t.Fatalf("cleanup prefix: %v", err)
	}

	client, err := harness.Client(context.Background())
	if err != nil {
		t.Fatalf("create s3 client: %v", err)
	}
	if _, err := client.StatObject(context.Background(), bucket, "current/proof.txt", minio.StatObjectOptions{}); err == nil {
		t.Fatal("expected current prefix object to be removed")
	}
	if _, err := client.StatObject(context.Background(), bucket, "sibling/proof.txt", minio.StatObjectOptions{}); err != nil {
		t.Fatalf("expected sibling prefix object to remain: %v", err)
	}
}

func resetSharedHarness(t testing.TB) {
	t.Helper()

	sharedHarnessMu.Lock()
	sharedHarness = nil
	sharedHarnessMu.Unlock()

	t.Cleanup(func() {
		sharedHarnessMu.Lock()
		sharedHarness = nil
		sharedHarnessMu.Unlock()
	})
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

func stubOwnedObjectStoreStartup(t testing.TB) {
	t.Helper()

	oldStartContainer := startContainerFn
	oldWaitReady := waitReadyFn
	oldPreflight := startPreflightFn
	oldSleep := startSleepFn
	t.Cleanup(func() {
		startContainerFn = oldStartContainer
		waitReadyFn = oldWaitReady
		startPreflightFn = oldPreflight
		startSleepFn = oldSleep
	})

	startPreflightFn = func(context.Context) (string, error) {
		return "unix:///var/run/docker.sock", nil
	}
	startSleepFn = func(context.Context, time.Duration) error {
		return nil
	}
}

type fakeObjectStoreContainer struct {
	testcontainers.Container
	host      string
	port      network.Port
	terminate func(context.Context) error
}

func (c fakeObjectStoreContainer) Host(context.Context) (string, error) {
	return c.host, nil
}

func (c fakeObjectStoreContainer) MappedPort(context.Context, string) (network.Port, error) {
	return c.port, nil
}

func (c fakeObjectStoreContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	if c.terminate == nil {
		return nil
	}
	return c.terminate(ctx)
}
