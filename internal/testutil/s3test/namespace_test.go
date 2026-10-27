package s3test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

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

func TestPreparePackageBucketCoreClassifiesPutDeadlineAndCleanupFailure(t *testing.T) {
	ops := stubPackageBucketAdmission()
	ops.readyPackage = func(context.Context, *Harness, string) error {
		return &objectStoreReadinessError{
			Stage:           "put",
			Attempts:        25,
			CleanupOutcome:  "failed",
			LastErr:         context.DeadlineExceeded,
			DeadlineExpired: true,
		}
	}
	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	_, release, err := harness.preparePackageBucket(
		context.Background(),
		"package-timeout",
		fixtureAttribution{CallerPackage: "example/timeout"},
	)
	if err == nil || release != nil {
		t.Fatalf("expected terminal package admission error, got release=%v err=%v", release != nil, err)
	}
	envelope := objectStoreSetupEnvelope("create_bucket", err)
	if envelope.FailureClass != "infra" || envelope.FailureReason != "service_readiness_timeout" ||
		envelope.ReadinessStage != "put" || envelope.AttemptCount != 25 || envelope.CleanupOutcome != "failed" {
		t.Fatalf("unexpected timeout envelope: %#v", envelope)
	}
}

func TestPreparePackageBucketCoreRejectsCapabilityAndCancellation(t *testing.T) {
	for name, readinessErr := range map[string]error{
		"capability": &objectStoreReadinessError{
			Stage:          "create_namespace",
			Attempts:       1,
			CleanupOutcome: "completed",
			LastErr:        &smithy.GenericAPIError{Code: "AccessDenied"},
		},
		"cancelled": context.Canceled,
	} {
		t.Run(name, func(t *testing.T) {
			ops := stubPackageBucketAdmission()
			ops.readyPackage = func(context.Context, *Harness, string) error { return readinessErr }
			harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
			_, _, err := harness.preparePackageBucket(
				context.Background(),
				"package-failure",
				fixtureAttribution{CallerPackage: "example/" + name},
			)
			if err == nil {
				t.Fatal("expected package admission failure")
			}
			envelope := objectStoreSetupEnvelope("create_bucket", err)
			if name == "capability" && (envelope.FailureClass != "harness" || envelope.ReadinessStage != "create_bucket") {
				t.Fatalf("unexpected capability envelope: %#v", envelope)
			}
			if name == "cancelled" && (envelope.FailureClass != "interrupted" || envelope.FailureReason != "cancelled_or_interrupted") {
				t.Fatalf("unexpected cancellation envelope: %#v", envelope)
			}
		})
	}
}

func TestPreparePackageBucketCoreReusesPackageBucket(t *testing.T) {
	ops := stubPackageBucketAdmission()
	creates := 0
	resets := 0
	ops.createPackage = func(context.Context, *Harness, string, fixtureAttribution) error {
		creates++
		return nil
	}
	ops.cleanupPackage = func(context.Context, *Harness, string, fixtureAttribution) error {
		resets++
		return nil
	}
	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	attribution := fixtureAttribution{CallerPackage: "example/reuse"}
	first, releaseFirst, err := harness.preparePackageBucket(context.Background(), "reuse", attribution)
	if err != nil {
		t.Fatalf("first package admission: %v", err)
	}
	releaseFirst()
	second, releaseSecond, err := harness.preparePackageBucket(context.Background(), "reuse", attribution)
	if err != nil {
		t.Fatalf("second package admission: %v", err)
	}
	releaseSecond()
	if first != second || creates != 1 || resets != 1 {
		t.Fatalf("package bucket was not reused: first=%q second=%q creates=%d resets=%d", first, second, creates, resets)
	}
}

func TestPreparePackageBucketCoreAdmitsDistinctPackagesConcurrently(t *testing.T) {
	ops := stubPackageBucketAdmission()
	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	type result struct {
		bucket  string
		release func()
		err     error
	}
	results := make(chan result, 2)
	for _, callerPackage := range []string{"example/first", "example/second"} {
		go func(packageName string) {
			bucket, release, err := harness.preparePackageBucket(
				context.Background(),
				"concurrent",
				fixtureAttribution{CallerPackage: packageName},
			)
			results <- result{bucket: bucket, release: release, err: err}
		}(callerPackage)
	}
	first := <-results
	second := <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent admission failed: first=%v second=%v", first.err, second.err)
	}
	first.release()
	second.release()
	if first.bucket == second.bucket {
		t.Fatalf("distinct packages shared bucket %q", first.bucket)
	}
}

func stubPackageBucketAdmission() fixtureOperations {
	return fixtureOperations{
		createPackage:  func(context.Context, *Harness, string, fixtureAttribution) error { return nil },
		cleanupPackage: func(context.Context, *Harness, string, fixtureAttribution) error { return nil },
		readyPackage:   func(context.Context, *Harness, string) error { return nil },
	}
}

type cleanupMemoryClient struct {
	objects    map[string]bool
	batches    []int
	lists      int
	emptyProof bool
	listHook   func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	deleteHook func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}

func (c *cleanupMemoryClient) ListObjectsV2(_ context.Context, input *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	c.lists++
	if c.listHook != nil {
		return c.listHook(input)
	}
	keys := []string{}
	for key := range c.objects {
		if strings.HasPrefix(key, aws.ToString(input.Prefix)) && key > aws.ToString(input.ContinuationToken) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 && aws.ToString(input.ContinuationToken) == "" {
		c.emptyProof = true
	}
	count := min(len(keys), int(aws.ToInt32(input.MaxKeys)))
	page := &s3.ListObjectsV2Output{IsTruncated: aws.Bool(count < len(keys))}
	for _, key := range keys[:count] {
		page.Contents = append(page.Contents, types.Object{Key: aws.String(key)})
	}
	if count < len(keys) {
		page.NextContinuationToken = aws.String(keys[count-1])
	}
	return page, nil
}

func (c *cleanupMemoryClient) DeleteObjects(_ context.Context, input *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	c.batches = append(c.batches, len(input.Delete.Objects))
	if c.deleteHook != nil {
		return c.deleteHook(input)
	}
	for _, object := range input.Delete.Objects {
		delete(c.objects, aws.ToString(object.Key))
	}
	return &s3.DeleteObjectsOutput{}, nil
}

func TestObjectCleanupBatchesAndProvesExactEmptyNamespace(t *testing.T) {
	for _, count := range []int{0, 1, 1000, 2507} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			client := &cleanupMemoryClient{objects: map[string]bool{"sibling/keep": true}}
			for i := 0; i < count; i++ {
				client.objects[fmt.Sprintf("owned/%05d", i)] = true
			}
			if err := cleanupObjectPrefix(context.Background(), client, "bucket", "owned/"); err != nil {
				t.Fatal(err)
			}
			if len(client.objects) != 1 || !client.objects["sibling/keep"] {
				t.Fatalf("cleanup crossed ownership or left residue: %d objects", len(client.objects))
			}
			for _, size := range client.batches {
				if size < 1 || size > 1000 {
					t.Fatalf("invalid batch size: %d", size)
				}
			}
			if !client.emptyProof {
				t.Fatal("missing fresh empty-namespace observation")
			}
		})
	}
}

func TestObjectCleanupRejectsUncertainCompletion(t *testing.T) {
	failure := errors.New("provider transport failure")
	page := func() *s3.ListObjectsV2Output { return &s3.ListObjectsV2Output{IsTruncated: aws.Bool(false)} }
	for _, name := range []string{"list", "delete", "partial", "residue", "verify", "nil_page", "nil_result", "missing_truncation", "missing_token", "duplicate_token", "cyclic_token", "outside_prefix", "empty_key", "common_prefix", "cancel_list", "cancel_delete", "missing_bucket", "missing_at_verification"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := &cleanupMemoryClient{objects: map[string]bool{"owned/key": true}}
			escapedCycle := errors.New("requested after repeated continuation token")
			switch name {
			case "list":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) { return nil, failure }
			case "delete":
				client.deleteHook = func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) { return nil, failure }
			case "partial":
				client.deleteHook = func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
					return &s3.DeleteObjectsOutput{Errors: []types.Error{{Code: aws.String("AccessDenied"), Message: aws.String("untrusted-provider-detail")}}}, nil
				}
			case "residue":
				client.deleteHook = func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) { return &s3.DeleteObjectsOutput{}, nil }
			case "verify":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
					if client.lists > 1 {
						return nil, failure
					}
					return page(), nil
				}
			case "nil_page":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) { return nil, nil }
			case "nil_result":
				client.deleteHook = func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) { return nil, nil }
			case "missing_truncation":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) { return &s3.ListObjectsV2Output{}, nil }
			case "missing_token", "duplicate_token", "cyclic_token":
				repeated := false
				seen := map[string]bool{}
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
					if repeated {
						return nil, escapedCycle
					}
					p := page()
					p.IsTruncated = aws.Bool(true)
					if name == "duplicate_token" {
						p.NextContinuationToken = aws.String("same")
					}
					if name == "cyclic_token" {
						p.NextContinuationToken = aws.String(fmt.Sprint(client.lists % 2))
					}
					token := aws.ToString(p.NextContinuationToken)
					repeated = token == "" || seen[token]
					seen[token] = true
					return p, nil
				}
			case "outside_prefix", "empty_key", "common_prefix":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
					p := page()
					p.Contents = []types.Object{{Key: aws.String("owned/key")}}
					if name == "outside_prefix" {
						p.Contents = append(p.Contents, types.Object{Key: aws.String("sibling/keep")})
					}
					if name == "empty_key" {
						p.Contents = append(p.Contents, types.Object{})
					}
					if name == "common_prefix" {
						p.CommonPrefixes = []types.CommonPrefix{{Prefix: aws.String("owned/")}}
					}
					return p, nil
				}
			case "cancel_list":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) { cancel(); return page(), nil }
			case "cancel_delete":
				client.deleteHook = func(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
					cancel()
					return &s3.DeleteObjectsOutput{}, nil
				}
			case "missing_bucket", "missing_at_verification":
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
					if name == "missing_at_verification" && client.lists == 1 {
						return page(), nil
					}
					return nil, &smithy.GenericAPIError{Code: "NoSuchBucket"}
				}
			}
			err := cleanupObjectPrefix(ctx, client, "bucket", "owned/")
			if name == "missing_bucket" || name == "missing_at_verification" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("uncertain cleanup reported success")
			}
			if strings.Contains(err.Error(), "untrusted-provider-detail") {
				t.Fatal("per-object error leaked provider details")
			}
			if (name == "list" || name == "delete" || name == "verify") && !errors.Is(err, failure) {
				t.Fatalf("lost transport cause: %v", err)
			}
			if strings.HasPrefix(name, "cancel_") && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
			if (name == "outside_prefix" || name == "empty_key" || name == "common_prefix") && len(client.batches) != 0 {
				t.Fatal("invalid listing mutated objects")
			}
			if errors.Is(err, escapedCycle) {
				t.Fatal("pagination continued after invalid or repeated token")
			}
		})
	}
}

func TestPackageBucketFailedGenerationCannotBeReadmitted(t *testing.T) {
	for _, stage := range []string{"create", "reset", "readiness", "cancelled_reset"} {
		t.Run(stage, func(t *testing.T) {
			ops := stubPackageBucketAdmission()
			harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
			attribution := fixtureAttribution{CallerPackage: "example/failed"}
			if stage == "reset" || stage == "cancelled_reset" {
				_, release, err := harness.preparePackageBucket(context.Background(), "failed", attribution)
				if err != nil {
					t.Fatal(err)
				}
				release()
			}
			cause := errors.New("uncertain fixture settlement")
			if stage == "cancelled_reset" {
				cause = context.Canceled
			}
			calls := 0
			fail := func(context.Context, *Harness, string, fixtureAttribution) error { calls++; return cause }
			if stage == "create" {
				harness.ops.createPackage = fail
			}
			if stage == "reset" || stage == "cancelled_reset" {
				harness.ops.cleanupPackage = fail
			}
			if stage == "readiness" {
				harness.ops.readyPackage = func(context.Context, *Harness, string) error { calls++; return cause }
			}
			_, release, err := harness.preparePackageBucket(context.Background(), "failed", attribution)
			if !errors.Is(err, cause) || release != nil {
				t.Fatalf("failed admission lost cause: %v", err)
			}
			if harness.packageBucket(attribution.CallerPackage).bucket == "" {
				t.Fatal("failed generation lost cleanup identity")
			}
			done := make(chan error, 1)
			go func() {
				_, release, err := harness.preparePackageBucket(context.Background(), "failed", attribution)
				if release != nil {
					release()
				}
				done <- err
			}()
			select {
			case err := <-done:
				if !errors.Is(err, cause) || calls != 1 {
					t.Fatalf("generation was readmitted or retried: err=%v calls=%d", err, calls)
				}
			case <-time.After(time.Second):
				t.Fatal("failed admission retained the package lock")
			}
		})
	}
}

func TestOwnedBucketReservesCleanupBeforeClientFailure(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	t.Setenv(suiteservices.SuiteIDEnv, "suite-reservation")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "run-reservation")
	harness := &Harness{binding: bindFixture(processFixtureInputs()), Endpoint: "invalid endpoint", suiteHash: "suitehash", processHash: "process1"}
	if _, err := harness.BootstrapBucket(context.Background(), "reserved"); err == nil {
		t.Fatal("invalid client unexpectedly admitted")
	}
	ledger, ok, err := suiteservices.CurrentResourceLedger(nil)
	if err != nil || !ok || len(ledger.Buckets) != 1 {
		t.Fatalf("uncertain create lost cleanup ownership: ok=%t err=%v ledger=%#v", ok, err, ledger)
	}
	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil || !ok || scope.ObjectStore.BucketCreateCount != 0 {
		t.Fatalf("failed create counted as success: ok=%t err=%v", ok, err)
	}
}
