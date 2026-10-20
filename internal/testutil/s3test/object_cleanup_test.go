package s3test

import (
	"context"
	"errors"
	"fmt"
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

type cleanupMemoryClient struct {
	objects    map[string]bool
	batches    []int
	lists      int
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
			if len(client.batches) != (count+999)/1000 {
				t.Fatalf("unexpected batch requests: %v", client.batches)
			}
			for _, size := range client.batches {
				if size < 1 || size > 1000 {
					t.Fatalf("invalid batch size: %d", size)
				}
			}
			if client.lists != max(1, len(client.batches))+1 {
				t.Fatalf("missing traversal or independent empty verification: %d lists", client.lists)
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
				client.listHook = func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
					p := page()
					p.IsTruncated = aws.Bool(true)
					if name == "duplicate_token" {
						p.NextContinuationToken = aws.String("same")
					}
					if name == "cyclic_token" {
						p.NextContinuationToken = aws.String(fmt.Sprint(client.lists % 2))
					}
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
			if client.lists > 3 {
				t.Fatalf("pagination did not terminate: %d requests", client.lists)
			}
		})
	}
}

func TestPackageBucketFailedGenerationCannotBeReadmitted(t *testing.T) {
	for _, stage := range []string{"create", "reset", "readiness", "cancelled_reset"} {
		t.Run(stage, func(t *testing.T) {
			stubPackageBucketAdmission(t)
			harness := &Harness{suiteHash: "suitehash", processHash: "process1"}
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
				preparePackageCreateBucketFn = fail
			}
			if stage == "reset" || stage == "cancelled_reset" {
				preparePackageCleanupFn = fail
			}
			if stage == "readiness" {
				preparePackageWaitReadyFn = func(context.Context, *Harness, string) error { calls++; return cause }
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

type countedCleanupClient struct {
	*s3.Client
	batches []int
}

func (c *countedCleanupClient) DeleteObjects(ctx context.Context, input *s3.DeleteObjectsInput, options ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	c.batches = append(c.batches, len(input.Delete.Objects))
	return c.Client.DeleteObjects(ctx, input, options...)
}

func TestPinnedObjectStoreBatchCleanupTraversesMultiplePages(t *testing.T) {
	harness := Start(t)
	ctx := context.Background()
	bucket, err := harness.BootstrapBucket(ctx, "batch-cleanup")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	client, err := harness.Client(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1007; i++ {
		_, err := client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(fmt.Sprintf("owned/%05d", i)), Body: strings.NewReader("x")})
		if err != nil {
			t.Fatalf("prepare object %d: %v", i, err)
		}
	}
	if _, err := client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String("sibling/keep"), Body: strings.NewReader("x")}); err != nil {
		t.Fatal(err)
	}
	counted := &countedCleanupClient{Client: client}
	cleanupCtx, cancel := context.WithTimeout(ctx, objectStoreFixtureCleanupTimeout)
	defer cancel()
	if err := cleanupObjectPrefix(cleanupCtx, counted, bucket, "owned/"); err != nil {
		t.Fatal(err)
	}
	if len(counted.batches) != 2 || counted.batches[0] != 1000 || counted.batches[1] != 7 {
		t.Fatalf("expected 2 bounded requests for 1007 objects: %v", counted.batches)
	}
	if _, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String("sibling/keep")}); err != nil {
		t.Fatalf("sibling was removed: %v", err)
	}
}

func TestOwnedBucketReservesCleanupBeforeClientFailure(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	t.Setenv(suiteservices.SuiteIDEnv, "suite-reservation")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "run-reservation")
	harness := &Harness{Endpoint: "invalid endpoint", suiteHash: "suitehash", processHash: "process1"}
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
