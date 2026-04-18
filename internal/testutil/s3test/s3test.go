package s3test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type Harness struct {
	Container testcontainers.Container
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool

	counter uint64
}

func Start(t testing.TB) *Harness {
	t.Helper()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Cmd:          []string{"server", "/data", "--console-address", ":9001"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start minio testcontainer: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("resolve minio host: %v", err)
	}

	port, err := container.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatalf("resolve minio mapped port: %v", err)
	}

	harness := &Harness{
		Container: container,
		Endpoint:  host + ":" + port.Port(),
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Secure:    false,
	}

	if err := harness.WaitReady(ctx); err != nil {
		_ = harness.Close(ctx)
		t.Fatalf("wait for minio readiness: %v", err)
	}

	t.Cleanup(func() {
		if err := harness.Close(context.Background()); err != nil {
			t.Fatalf("terminate minio testcontainer: %v", err)
		}
	})

	return harness
}

func (h *Harness) WaitReady(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		client, err := h.Client(ctx)
		if err == nil {
			_, err = client.ListBuckets(ctx)
		}
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("minio did not become ready: %w", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (h *Harness) Client(ctx context.Context) (*minio.Client, error) {
	_ = ctx
	return minio.New(h.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(h.AccessKey, h.SecretKey, ""),
		Secure: h.Secure,
	})
}

func (h *Harness) BootstrapBucket(ctx context.Context, prefix string) (string, error) {
	name := fmt.Sprintf("%s-%06d", sanitizeBucket(prefix), atomic.AddUint64(&h.counter, 1))
	if err := h.CreateBucket(ctx, name); err != nil {
		return "", err
	}
	return name, nil
}

func (h *Harness) CreateBucket(ctx context.Context, name string) error {
	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	if err := client.MakeBucket(ctx, name, minio.MakeBucketOptions{}); err != nil {
		exists, bucketErr := client.BucketExists(ctx, name)
		if bucketErr != nil || !exists {
			return fmt.Errorf("create bucket %s: %w", name, err)
		}
	}

	return nil
}

func (h *Harness) RoundTrip(ctx context.Context, bucket string, key string, payload []byte) ([]byte, error) {
	client, err := h.Client(ctx)
	if err != nil {
		return nil, err
	}

	_, err = client.PutObject(ctx, bucket, key, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("put object %s/%s: %w", bucket, key, err)
	}

	object, err := client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %s/%s: %w", bucket, key, err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("read object %s/%s: %w", bucket, key, err)
	}

	return data, nil
}

func (h *Harness) CleanupBucket(ctx context.Context, bucket string) error {
	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	for objectInfo := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if objectInfo.Err != nil {
			return fmt.Errorf("list bucket %s: %w", bucket, objectInfo.Err)
		}
		if err := client.RemoveObject(ctx, bucket, objectInfo.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove object %s/%s: %w", bucket, objectInfo.Key, err)
		}
	}

	if err := client.RemoveBucket(ctx, bucket); err != nil {
		return fmt.Errorf("remove bucket %s: %w", bucket, err)
	}

	return nil
}

func (h *Harness) Env(bucket string) map[string]string {
	return map[string]string{
		objectstore.EndpointEnv:  h.Endpoint,
		objectstore.AccessKeyEnv: h.AccessKey,
		objectstore.SecretKeyEnv: h.SecretKey,
		objectstore.SecureEnv:    "false",
		objectstore.BucketEnv:    bucket,
	}
}

func (h *Harness) EnvForServiceRef(serviceRef string, bucket string) map[string]string {
	keys, err := objectstore.EnvKeysForServiceRef(serviceRef)
	if err != nil {
		return map[string]string{}
	}

	return map[string]string{
		keys.Endpoint:  h.Endpoint,
		keys.AccessKey: h.AccessKey,
		keys.SecretKey: h.SecretKey,
		keys.Secure:    "false",
		keys.Bucket:    bucket,
	}
}

func (h *Harness) Close(ctx context.Context) error {
	if h == nil || h.Container == nil {
		return nil
	}

	return h.Container.Terminate(ctx)
}

func sanitizeBucket(prefix string) string {
	prefix = strings.ToLower(prefix)
	prefix = strings.ReplaceAll(prefix, "_", "-")
	prefix = strings.ReplaceAll(prefix, " ", "-")
	if prefix == "" {
		return "cartulary-test"
	}
	return prefix
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}
