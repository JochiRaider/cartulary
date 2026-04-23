package s3test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

const (
	minioImage                = "minio/minio:RELEASE.2025-09-07T16-13-09Z"
	minioAPIPort              = "9000/tcp"
	minioConsolePort          = "9001/tcp"
	minioStartupTimeout       = 2 * time.Minute
	minioHealthPollInterval   = 500 * time.Millisecond
	minioClientReadyTimeout   = 60 * time.Second
	minioClientAttemptTimeout = 5 * time.Second
)

type Harness struct {
	Container testcontainers.Container
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool

	suiteHash   string
	processHash string
	counter     uint64
	shared      bool
	attached    bool
}

var (
	sharedHarnessMu   sync.Mutex
	sharedHarness     *Harness
	startOwnedHarness = StartOwned
	verifyAttachedFn  = verifyAttachedHarness
)

func Start(t testing.TB) *Harness {
	t.Helper()

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("%v", err)
	}

	return harness
}

func StartShared(ctx context.Context) (*Harness, error) {
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness != nil {
		return sharedHarness, nil
	}

	harness, attached, err := startAttachedHarness(ctx)
	if err != nil {
		return nil, err
	}
	if !attached {
		harness, err = startOwnedHarness(ctx)
	}
	if err != nil {
		return nil, err
	}
	harness.shared = true
	sharedHarness = harness
	return harness, nil
}

func StartOwned(ctx context.Context) (*Harness, error) {
	return startHarness(ctx)
}

func StopShared(ctx context.Context) error {
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness == nil || sharedHarness.Container == nil {
		sharedHarness = nil
		return nil
	}

	container := sharedHarness.Container
	sharedHarness = nil
	return container.Terminate(ctx)
}

func startHarness(ctx context.Context) (*Harness, error) {
	req := testcontainers.ContainerRequest{
		Image:        minioImage,
		ExposedPorts: []string{minioAPIPort, minioConsolePort},
		Cmd:          []string{"server", "/data", "--console-address", ":9001"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		WaitingFor: wait.ForListeningPort(minioAPIPort).WithStartupTimeout(minioStartupTimeout),
	}

	container, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service: "minio testcontainer",
		Image:   minioImage,
	}, func(ctx context.Context) (testcontainers.Container, error) {
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve minio host: %w", err)
	}

	port, err := container.MappedPort(ctx, minioAPIPort)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve minio mapped port: %w", err)
	}

	harness := &Harness{
		Container:   container,
		Endpoint:    host + ":" + port.Port(),
		AccessKey:   "minioadmin",
		SecretKey:   "minioadmin",
		Secure:      false,
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
	}

	if err := harness.WaitReady(ctx); err != nil {
		_ = harness.Close(ctx)
		return nil, fmt.Errorf("wait for minio readiness: %w", err)
	}

	return harness, nil
}

func startAttachedHarness(ctx context.Context) (*Harness, bool, error) {
	endpoint := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3EndpointEnv))
	accessKey := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3AccessKeyEnv))
	secretKey := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3SecretKeyEnv))
	if endpoint == "" && accessKey == "" && secretKey == "" {
		return nil, false, nil
	}
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, false, fmt.Errorf("attach minio harness: %s, %s, and %s must all be set", suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv)
	}

	secure, err := suiteservices.ParseBool(suiteservices.LookupEnvValue(nil, suiteservices.S3SecureEnv))
	if err != nil {
		return nil, false, fmt.Errorf("attach minio harness: %w", err)
	}

	harness := &Harness{
		Endpoint:    endpoint,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Secure:      secure,
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
		attached:    true,
	}
	if err := verifyAttachedFn(ctx, harness); err != nil {
		return nil, false, fmt.Errorf("attach minio harness: authenticated readiness: %w", err)
	}
	recordSuiteEvent(suiteservices.Event{Type: suiteservices.EventS3Attach})
	return harness, true, nil
}

func (h *Harness) WaitReady(ctx context.Context) error {
	readyCtx, cancel := context.WithTimeout(ctx, minioClientReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(minioHealthPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		attemptCtx, attemptCancel := context.WithTimeout(readyCtx, minioClientAttemptTimeout)
		client, err := h.Client(attemptCtx)
		if err == nil {
			_, err = client.ListBuckets(attemptCtx)
		}
		attemptCancel()
		if err == nil {
			return nil
		}

		lastErr = err
		select {
		case <-readyCtx.Done():
			if lastErr == nil {
				lastErr = readyCtx.Err()
			}
			return fmt.Errorf("minio did not become ready via authenticated api: %w", lastErr)
		case <-ticker.C:
		}
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
	name := h.nextBucketName(prefix)
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
	recordSuiteEvent(suiteservices.Event{
		Type: suiteservices.EventS3BucketCreated,
		Name: name,
	})

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
	recordSuiteEvent(suiteservices.Event{
		Type: suiteservices.EventS3BucketCleaned,
		Name: bucket,
	})

	return nil
}

func (h *Harness) Env(bucket string) map[string]string {
	return map[string]string{
		objectstore.EndpointEnv:  h.Endpoint,
		objectstore.AccessKeyEnv: h.AccessKey,
		objectstore.SecretKeyEnv: h.SecretKey,
		objectstore.SecureEnv:    strconv.FormatBool(h.Secure),
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
		keys.Secure:    strconv.FormatBool(h.Secure),
		keys.Bucket:    bucket,
	}
}

func (h *Harness) Close(ctx context.Context) error {
	if h == nil || h.Container == nil {
		return nil
	}
	// Shared harnesses stay alive until StopShared or test-process teardown.
	if h.shared {
		return nil
	}

	return h.Container.Terminate(ctx)
}

func sanitizeBucket(prefix string) string {
	prefix = strings.ToLower(prefix)
	prefix = strings.ReplaceAll(prefix, "_", "-")
	prefix = strings.ReplaceAll(prefix, " ", "-")

	var builder strings.Builder
	for _, r := range prefix {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-':
			builder.WriteRune(r)
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "cartulary-test"
	}
	return result
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

func resolveSuiteHash() string {
	if hash := suiteservices.SuiteHash(nil); hash != "" {
		return hash
	}
	return suiteservices.ShortHash("local-suite", 8)
}

func (h *Harness) nextBucketName(prefix string) string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	processHash := h.processHash
	if processHash == "" {
		processHash = suiteservices.ProcessHash()
	}

	base := fmt.Sprintf("ct-%s-%s-%06d", suiteHash, processHash, atomic.AddUint64(&h.counter, 1))
	suffix := sanitizeBucket(prefix)
	if suffix == "" {
		return truncateBucketName(base, 63)
	}

	available := 63 - len(base) - 1
	if available <= 0 {
		return truncateBucketName(base, 63)
	}
	return truncateBucketName(base+"-"+truncateBucketName(suffix, available), 63)
}

func truncateBucketName(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return strings.Trim(value[:max], "-")
}

func verifyAttachedHarness(ctx context.Context, harness *Harness) error {
	client, err := harness.Client(ctx)
	if err != nil {
		return err
	}
	_, err = client.ListBuckets(ctx)
	return err
}

func recordSuiteEvent(event suiteservices.Event) {
	_ = suiteservices.RecordEvent(nil, event)
}
