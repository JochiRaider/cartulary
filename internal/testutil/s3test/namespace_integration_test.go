package s3test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestResetBucketPreservesNamespaceAndProvesMutation(t *testing.T) {
	harness := Start(t)

	bucket, err := harness.BootstrapBucket(context.Background(), "stable-reset")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	if _, err := harness.RoundTrip(context.Background(), bucket, "prior/object.txt", []byte("prior")); err != nil {
		t.Fatalf("write prior generation: %v", err)
	}
	if err := harness.ResetBucket(context.Background(), bucket); err != nil {
		t.Fatalf("reset bucket: %v", err)
	}

	client, err := harness.Client(context.Background())
	if err != nil {
		t.Fatalf("create s3 client: %v", err)
	}
	exists, err := s3transport.BucketExists(context.Background(), client, bucket)
	if err != nil || !exists {
		t.Fatalf("stable bucket missing after reset: exists=%t err=%v", exists, err)
	}
	if _, err := client.HeadObject(context.Background(), &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String("prior/object.txt")}); !isNoSuchObjectError(err) {
		t.Fatalf("prior object absence was not proven: %v", err)
	}
	if got, err := harness.RoundTrip(context.Background(), bucket, "next/object.txt", []byte("next")); err != nil || string(got) != "next" {
		t.Fatalf("next generation mutation: %v", err)
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

func TestObjectStoreNamespaceLifecycle(t *testing.T) {
	h := Start(t)
	ctx := t.Context()
	bucket, err := h.BootstrapBucket(ctx, "namespace")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := h.CleanupBucket(context.Background(), bucket); err != nil {
			t.Error("final cleanup", err)
		}
	})
	payload := []byte("exact roundtrip payload")
	got, err := h.RoundTrip(ctx, bucket, "roundtrip/proof", payload)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatal("roundtrip payload changed", err)
	}
	sibling := []byte("sibling bytes must survive cleanup")
	if got, err := h.RoundTrip(ctx, bucket, "sibling/keep", sibling); err != nil || !bytes.Equal(got, sibling) {
		t.Fatal("sibling setup", err)
	}
	client, err := h.Client(ctx)
	if err != nil {
		t.Fatal(err)
	}
	const objectCount = 1007
	for i := 0; i < objectCount; i++ {
		if _, err := client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(fmt.Sprintf("owned/%05d", i)), Body: strings.NewReader("x")}); err != nil {
			t.Fatal("pagination setup", err)
		}
	}
	counted := &countedCleanupClient{Client: client}
	cleanupCtx, cancel := context.WithTimeout(ctx, objectStoreFixtureCleanupTimeout)
	defer cancel()
	if err := cleanupObjectPrefix(cleanupCtx, counted, bucket, "owned/"); err != nil {
		t.Fatal(err)
	}
	removed := 0
	for _, size := range counted.batches {
		if size < 1 || size > objectCleanupBatchLimit {
			t.Fatal("invalid real delete batch", size)
		}
		removed += size
	}
	if removed != objectCount {
		t.Fatal("incomplete real pagination", removed)
	}
	empty, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String("owned/")})
	if err != nil || empty == nil || len(empty.Contents) != 0 || aws.ToBool(empty.IsTruncated) {
		t.Fatal("real cleanup did not prove empty prefix", err)
	}
	if err := h.CleanupPrefix(ctx, bucket, "roundtrip/"); err != nil {
		t.Fatal("fixture prefix cleanup", err)
	}
	t.Run("cleanup_publication_failure", func(t *testing.T) {
		if _, err := h.RoundTrip(t.Context(), bucket, "publication/proof", payload); err != nil {
			t.Fatal(err)
		}
		// The service operation can succeed while its required journal write
		// fails. That failure must remain visible to the cleanup caller.
		inputs := maps.Clone(h.binding.environment())
		blocked := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(blocked, []byte("blocked"), 0o600); err != nil {
			t.Fatal(err)
		}
		inputs["CARTULARY_TEST_RESULTS_DIR"] = blocked
		borrower, err := harnessFromBinding(bindFixture(inputs))
		if err != nil {
			t.Fatal(err)
		}
		defer borrower.Close(context.Background())
		if err := borrower.CleanupPrefix(t.Context(), bucket, "publication/"); err == nil {
			t.Fatal("failed cleanup publication reported success")
		}
		_, err = client.HeadObject(t.Context(), &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String("publication/proof")})
		if !isNoSuchObjectError(err) {
			t.Fatal("publication failure case did not reach successful deletion", err)
		}
	})
	for _, key := range []string{"owned/00000", "owned/01006", "roundtrip/proof"} {
		_, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
		if !isNoSuchObjectError(err) {
			t.Fatalf("typed absence required for %s: %v", key, err)
		}
	}
	response, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String("sibling/keep")})
	if err != nil {
		t.Fatal("read sibling", err)
	}
	retained, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(retained, sibling) {
		t.Fatal("sibling payload changed", readErr, closeErr)
	}
	// Recovery uses saved inputs despite conflicting ambient credentials/assignment.
	saved := h.binding.environment()
	t.Setenv(suiteservices.S3EndpointEnv, "invalid ambient endpoint")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	if err := CleanupBucketFromBinding(ctx, saved, bucket); err != nil {
		t.Fatal("saved-binding recovery cleanup", err)
	}
	if exists, err := s3transport.BucketExists(ctx, client, bucket); err != nil || exists {
		t.Fatal("final bucket deletion unproven", exists, err)
	}
	if err := h.CleanupBucket(ctx, bucket); err != nil {
		t.Fatal("idempotent cleanup", err)
	}
}
