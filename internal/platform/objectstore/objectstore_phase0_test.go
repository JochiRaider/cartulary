package objectstore_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"

	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/platform/objectstore"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestPhase0_ObjectStoreInitialization_I_0_02(t *testing.T) {
	s3Harness := s3test.Start(t)

	bucket := fmt.Sprintf("phase0-i-0-02-%d", time.Now().UnixNano())
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	}()

	env := s3Harness.Env(bucket)
	client, err := objectstore.SetupWithEnv(context.Background(), config.Config{}, env)
	if err != nil {
		t.Fatalf("setup object store: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		t.Fatalf("check bucket existence: %v", err)
	}
	if !exists {
		t.Fatalf("expected setup to create configured bucket %q", bucket)
	}

	payload := []byte("phase0 object-store bootstrap")
	_, err = client.PutObject(context.Background(), bucket, "proof.txt", bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{})
	if err != nil {
		t.Fatalf("put object: %v", err)
	}

	object, err := client.GetObject(context.Background(), bucket, "proof.txt", minio.GetObjectOptions{})
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	defer object.Close()

	got, err := io.ReadAll(object)
	if err != nil {
		t.Fatalf("read object payload: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected object payload: got %q want %q", got, payload)
	}

	if err := client.RemoveObject(context.Background(), bucket, "proof.txt", minio.RemoveObjectOptions{}); err != nil {
		t.Fatalf("remove object: %v", err)
	}
}
