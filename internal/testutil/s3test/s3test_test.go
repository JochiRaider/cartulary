package s3test

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
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

func TestMinIOContainerWaitStrategyOnlyWaitsForPortMapping(t *testing.T) {
	strategy := minioPortWaitStrategy()
	if got := strategy.String(); !strings.Contains(got, "to be mapped") {
		t.Fatalf("expected mapped-port-only wait strategy, got %q", got)
	}
	if timeout := strategy.Timeout(); timeout == nil || *timeout != minioPortMappingTimeout {
		t.Fatalf("unexpected port mapping timeout: got %v want %v", timeout, minioPortMappingTimeout)
	}
}

func TestHarnessStartsMinIOAndRoundTripsObjects(t *testing.T) {
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
		t.Fatalf("create minio client: %v", err)
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
