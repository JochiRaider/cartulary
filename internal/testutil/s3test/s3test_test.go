package s3test

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
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
	t.Setenv(suiteservices.S3ProbeBucketEnv, "ct-suite-readiness-probe")

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
		if harness.probeBucket != "ct-suite-readiness-probe" {
			t.Fatalf("unexpected broker probe bucket: %q", harness.probeBucket)
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

func TestSeaweedFSS3AllowedOriginsDefaultSupportsLoopbackBrowserPortRange(t *testing.T) {
	t.Setenv("OBJECT_STORE_CORS_ALLOWED_ORIGINS", "")
	t.Setenv("OBJECT_STORE_CORS_ORIGIN", "")

	got := seaweedFSS3AllowedOrigins()
	for _, want := range []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:19000",
		"http://127.0.0.1:19000",
		"http://localhost:19199",
		"http://127.0.0.1:19199",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("default SeaweedFS S3 allowed origins missing %q in %q", want, got)
		}
	}
	for _, forbidden := range []string{"*", "http://0.0.0.0", "http://[::]"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("default SeaweedFS S3 allowed origins must stay exact and loopback-scoped, got %q", got)
		}
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

func TestOwnedObjectStoreReadinessTimeoutIsTerminalForTheStartedLane(t *testing.T) {
	stubOwnedObjectStoreStartup(t)

	starts := 0
	terminations := 0
	readinessChecks := 0
	var events []testcontainersx.StartEvent
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
		readinessChecks++
		return &objectStoreReadinessError{
			Stage:           "put",
			Attempts:        3,
			CleanupOutcome:  "completed",
			LastErr:         context.DeadlineExceeded,
			DeadlineExpired: true,
		}
	}

	_, err := StartOwnedWithOptions(context.Background(), StartOptions{
		Observer: func(event testcontainersx.StartEvent) {
			events = append(events, event)
		},
	})
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

func TestObjectStoreMutationProbeRetriesTransientStagesOnSameLane(t *testing.T) {
	for _, stage := range []string{"put", "head"} {
		t.Run(stage, func(t *testing.T) {
			client := newFakeReadinessClient()
			client.failures[stage] = 1
			configureReadinessClient(t, client)

			harness := &Harness{suiteHash: "suitehash", processHash: "process1"}
			err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
				ReadyTimeout:    100 * time.Millisecond,
				PollInterval:    time.Millisecond,
				AttemptTimeout:  20 * time.Millisecond,
				CreateNamespace: true,
				ListFirst:       true,
				Bucket:          "ct-suitehash-process1-readiness",
				Key:             ".cartulary-readiness/probe",
			})
			if err != nil {
				t.Fatalf("expected transient %s failure to recover on the same lane: %v", stage, err)
			}
			if got := client.callCount(stage); got < 2 {
				t.Fatalf("expected %s to be retried within readiness, got %d calls", stage, got)
			}
			client.assertNoProbeResidue(t)
		})
	}
}

func TestObjectStoreMutationProbeRejectsCapabilitiesImmediately(t *testing.T) {
	client := newFakeReadinessClient()
	client.failureErrors["create_namespace"] = minio.ErrorResponse{Code: "AccessDenied", StatusCode: 403}
	client.failures["create_namespace"] = -1
	configureReadinessClient(t, client)

	harness := &Harness{suiteHash: "suitehash", processHash: "process1"}
	err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    100 * time.Millisecond,
		PollInterval:    time.Millisecond,
		AttemptTimeout:  20 * time.Millisecond,
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          "ct-suitehash-process1-readiness",
		Key:             ".cartulary-readiness/probe",
	})
	if err == nil {
		t.Fatal("expected capability rejection")
	}
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) {
		t.Fatalf("expected typed readiness error, got %T", err)
	}
	if readinessErr.Attempts != 1 || readinessErr.Stage != "create_namespace" {
		t.Fatalf("expected immediate create_namespace rejection, got %#v", readinessErr)
	}
	if client.callCount("delete_namespace") != 1 {
		t.Fatalf("capability rejection did not clean its proven probe namespace: %#v", client.calls)
	}
	if strings.Contains(err.Error(), "AccessDenied") {
		t.Fatalf("readiness diagnostic exposed provider error text: %q", err)
	}
	client.assertNoProbeResidue(t)
}

func TestObjectStoreMutationProbeDeadlineRetainsCleanupEvidence(t *testing.T) {
	client := newFakeReadinessClient()
	client.failures["put"] = -1
	configureReadinessClient(t, client)

	harness := &Harness{suiteHash: "suitehash", processHash: "process1"}
	err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    15 * time.Millisecond,
		PollInterval:    time.Millisecond,
		AttemptTimeout:  5 * time.Millisecond,
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          "ct-suitehash-process1-readiness",
		Key:             ".cartulary-readiness/probe",
	})
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) || !readinessErr.DeadlineExpired {
		t.Fatalf("expected readiness deadline expiry, got %v", err)
	}
	if readinessErr.Attempts < 2 || readinessErr.CleanupOutcome != "completed" {
		t.Fatalf("expected same-lane polling with cleanup evidence, got %#v", readinessErr)
	}
	client.assertNoProbeResidue(t)
}

func TestObjectStoreMutationProbePreservesPrimaryFailureDuringCleanupFailure(t *testing.T) {
	client := newFakeReadinessClient()
	primaryErr := errors.New("temporary put failure")
	client.failures["put"] = 1
	client.failureErrors["put"] = primaryErr
	client.failures["delete_namespace"] = 1
	failure := runObjectStoreMutationProbe(context.Background(), client, objectStoreReadinessConfig{
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          "ct-suitehash-process1-readiness",
		Key:             ".cartulary-readiness/probe",
	})
	if failure == nil {
		t.Fatal("expected probe failure")
	}
	if failure.Cause != primaryErr || failure.Stage != "put" || failure.CleanupOutcome != "failed" || failure.CleanupErr == nil {
		t.Fatalf("cleanup must remain secondary to the put failure, got %#v", failure)
	}
	client.assertNoProbeResidue(t)
}

func TestAttachedObjectStoreUsesMutationProbeContract(t *testing.T) {
	client := newFakeReadinessClient()
	client.namespaces["ct-suite-readiness-probe"] = true
	configureReadinessClient(t, client)

	harness := &Harness{
		suiteHash:   "suitehash",
		processHash: "process1",
		attached:    true,
		probeBucket: "ct-suite-readiness-probe",
	}
	if err := verifyAttachedHarness(context.Background(), harness); err != nil {
		t.Fatalf("verify attached object store: %v", err)
	}
	for _, stage := range []string{"list", "put", "head", "delete", "delete_verify"} {
		if client.callCount(stage) == 0 {
			t.Fatalf("attached readiness did not execute %s", stage)
		}
	}
	if client.callCount("create_namespace") != 0 || client.callCount("delete_namespace") != 0 {
		t.Fatalf("attached readiness changed its broker-owned namespace: %#v", client.calls)
	}
	if !client.namespaces["ct-suite-readiness-probe"] || len(client.objects) != 0 {
		t.Fatalf("attached readiness did not preserve a clean broker namespace: namespaces=%#v objects=%#v", client.namespaces, client.objects)
	}
}

func TestPackageBucketAdmissionProbesWithoutRemovingTheBucket(t *testing.T) {
	client := newFakeReadinessClient()
	client.namespaces["package-bucket"] = true
	configureReadinessClient(t, client)

	harness := &Harness{suiteHash: "suitehash", processHash: "process1"}
	if err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    100 * time.Millisecond,
		PollInterval:    time.Millisecond,
		AttemptTimeout:  20 * time.Millisecond,
		CreateNamespace: false,
		ListFirst:       false,
		Bucket:          "package-bucket",
		Key:             ".cartulary-readiness/package-probe",
	}); err != nil {
		t.Fatalf("admit package bucket: %v", err)
	}
	if !client.namespaces["package-bucket"] {
		t.Fatal("package admission removed its borrowed package bucket")
	}
	if client.callCount("create_namespace") != 0 || client.callCount("delete_namespace") != 0 {
		t.Fatalf("package admission changed namespace ownership: %#v", client.calls)
	}
	if len(client.objects) != 0 {
		t.Fatalf("package admission left probe objects: %#v", client.objects)
	}
}

func TestHarnessStartsSeaweedFSS3AndRoundTripsObjects(t *testing.T) {
	harness := Start(t)

	bucket, err := harness.BootstrapBucket(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	cleaned := false
	defer func() {
		if !cleaned {
			if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Fatalf("cleanup bucket: %v", err)
			}
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
	if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
		t.Fatalf("cleanup bucket: %v", err)
	}
	cleaned = true
	if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
		t.Fatalf("repeat cleanup bucket: %v", err)
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
	oldReadinessClient := newReadinessClientFn
	oldPreflight := startPreflightFn
	oldSleep := startSleepFn
	t.Cleanup(func() {
		startContainerFn = oldStartContainer
		waitReadyFn = oldWaitReady
		newReadinessClientFn = oldReadinessClient
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

func configureReadinessClient(t testing.TB, client objectStoreReadinessClient) {
	t.Helper()
	old := newReadinessClientFn
	newReadinessClientFn = func(context.Context, *Harness) (objectStoreReadinessClient, error) {
		return client, nil
	}
	t.Cleanup(func() {
		newReadinessClientFn = old
	})
}

type fakeReadinessClient struct {
	mu            sync.Mutex
	namespaces    map[string]bool
	objects       map[string][]byte
	failures      map[string]int
	failureErrors map[string]error
	calls         []string
}

func newFakeReadinessClient() *fakeReadinessClient {
	return &fakeReadinessClient{
		namespaces:    make(map[string]bool),
		objects:       make(map[string][]byte),
		failures:      make(map[string]int),
		failureErrors: make(map[string]error),
	}
}

func (c *fakeReadinessClient) ListBuckets(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.recordAndFail("list")
}

func (c *fakeReadinessClient) CreateNamespace(_ context.Context, bucket string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("create_namespace"); err != nil {
		return err
	}
	c.namespaces[bucket] = true
	return nil
}

func (c *fakeReadinessClient) Put(_ context.Context, bucket string, key string, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("put"); err != nil {
		return err
	}
	if !c.namespaces[bucket] {
		return minio.ErrorResponse{Code: "NoSuchBucket", StatusCode: 404}
	}
	c.objects[bucket+"/"+key] = append([]byte(nil), payload...)
	return nil
}

func (c *fakeReadinessClient) HeadSize(_ context.Context, bucket string, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	objectKey := bucket + "/" + key
	payload, exists := c.objects[objectKey]
	stage := "head"
	if !exists {
		stage = "delete_verify"
	}
	if err := c.recordAndFail(stage); err != nil {
		return 0, err
	}
	if !exists {
		return 0, minio.ErrorResponse{Code: "NoSuchKey", StatusCode: 404}
	}
	return int64(len(payload)), nil
}

func (c *fakeReadinessClient) Delete(_ context.Context, bucket string, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("delete"); err != nil {
		return err
	}
	delete(c.objects, bucket+"/"+key)
	return nil
}

func (c *fakeReadinessClient) DeleteNamespace(_ context.Context, bucket string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.recordAndFail("delete_namespace")
	delete(c.namespaces, bucket)
	for key := range c.objects {
		if strings.HasPrefix(key, bucket+"/") {
			delete(c.objects, key)
		}
	}
	return err
}

func (c *fakeReadinessClient) recordAndFail(stage string) error {
	c.calls = append(c.calls, stage)
	remaining := c.failures[stage]
	if remaining == 0 {
		return nil
	}
	if remaining > 0 {
		c.failures[stage] = remaining - 1
	}
	if err := c.failureErrors[stage]; err != nil {
		return err
	}
	return errors.New("temporary object-store unavailability")
}

func (c *fakeReadinessClient) callCount(stage string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for _, call := range c.calls {
		if call == stage {
			count++
		}
	}
	return count
}

func (c *fakeReadinessClient) assertNoProbeResidue(t testing.TB) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.namespaces) != 0 || len(c.objects) != 0 {
		t.Fatalf("object-store probe residue: namespaces=%#v objects=%#v", c.namespaces, c.objects)
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
