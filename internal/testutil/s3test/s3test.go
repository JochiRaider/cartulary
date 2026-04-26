package s3test

import (
	"bytes"
	"context"
	"errors"
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
	minioPortMappingTimeout   = 30 * time.Second
	minioHealthPollInterval   = 500 * time.Millisecond
	minioClientReadyTimeout   = 60 * time.Second
	minioClientAttemptTimeout = 5 * time.Second
)

func ContainerImage() string {
	return minioImage
}

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

	packageBucketMu sync.Mutex
	packageBuckets  map[string]*packageBucket
}

type packageBucket struct {
	mu     sync.Mutex
	bucket string
}

type fixtureAttribution struct {
	TestName      string
	CallerFile    string
	CallerPackage string
}

var (
	sharedHarnessMu   sync.Mutex
	sharedHarness     *Harness
	startOwnedHarness = StartOwned
	verifyAttachedFn  = verifyAttachedHarness
	startContainerFn  = testcontainers.GenericContainer
	waitReadyFn       = func(ctx context.Context, harness *Harness) error {
		return harness.WaitReady(ctx)
	}
	startPreflightFn func(context.Context) (string, error)
	startSleepFn     func(context.Context, time.Duration) error
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
		WaitingFor: minioPortWaitStrategy(),
	}

	harness, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service:   "minio testcontainer",
		Image:     minioImage,
		Preflight: startPreflightFn,
		Retryable: isRetryableMinIOStartupFailure,
		Sleep:     startSleepFn,
	}, func(ctx context.Context) (*Harness, error) {
		return startHarnessAttempt(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return harness, nil
}

func startHarnessAttempt(ctx context.Context, req testcontainers.ContainerRequest) (*Harness, error) {
	container, err := startContainerFn(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	attemptSucceeded := false
	defer func() {
		if attemptSucceeded {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = container.Terminate(cleanupCtx)
	}()

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve minio host: %w", err)
	}

	port, err := container.MappedPort(ctx, minioAPIPort)
	if err != nil {
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

	if err := waitReadyFn(ctx, harness); err != nil {
		return nil, fmt.Errorf("wait for minio readiness: %w", err)
	}

	attemptSucceeded = true
	return harness, nil
}

func minioPortWaitStrategy() *wait.HostPortStrategy {
	return wait.ForMappedPort(minioAPIPort).WithStartupTimeout(minioPortMappingTimeout)
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
		if isNonRetryableMinIOReadinessError(err) {
			return &minioReadinessError{LastErr: err}
		}

		lastErr = err
		select {
		case <-readyCtx.Done():
			if lastErr == nil {
				lastErr = readyCtx.Err()
			}
			return &minioReadinessError{
				LastErr:         lastErr,
				DeadlineExpired: errors.Is(readyCtx.Err(), context.DeadlineExceeded),
			}
		case <-ticker.C:
		}
	}
}

type minioReadinessError struct {
	LastErr         error
	DeadlineExpired bool
}

func (e *minioReadinessError) Error() string {
	if e == nil {
		return ""
	}
	if e.LastErr == nil {
		return "minio did not become ready via authenticated api"
	}
	return fmt.Sprintf("minio did not become ready via authenticated api: %v", e.LastErr)
}

func (e *minioReadinessError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.LastErr
}

func isRetryableMinIOStartupFailure(err error) bool {
	var readinessErr *minioReadinessError
	if !errors.As(err, &readinessErr) {
		return false
	}
	if !readinessErr.DeadlineExpired {
		return false
	}
	return !isNonRetryableMinIOReadinessError(readinessErr.LastErr)
}

func isNonRetryableMinIOReadinessError(err error) bool {
	if err == nil {
		return false
	}

	var response minio.ErrorResponse
	if errors.As(err, &response) {
		switch strings.ToLower(response.Code) {
		case "accessdenied", "invalidaccesskeyid", "signaturedoesnotmatch":
			return true
		}
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "invalid endpoint") ||
		strings.Contains(lower, "access denied") ||
		strings.Contains(lower, "invalid access key") ||
		strings.Contains(lower, "signaturedoesnotmatch") ||
		strings.Contains(lower, "signature does not match")
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
	if err := h.createBucket(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		return "", err
	}
	return name, nil
}

func (h *Harness) CreateBucket(ctx context.Context, name string) error {
	return h.createBucket(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) createBucket(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	start := time.Now()
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
		Type:    suiteservices.EventS3BucketCreated,
		Name:    name,
		Details: s3FixtureDetails("bucket", reuseScope, attribution, time.Since(start)),
	})

	return nil
}

func (h *Harness) PreparePackageBucketT(t testing.TB, prefix string) string {
	t.Helper()

	attribution := fixtureAttributionFor(t, "s3test")
	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeBucket(prefix)
	}
	fixture := h.packageBucket(key)
	fixture.mu.Lock()

	if fixture.bucket == "" {
		name := h.nextBucketName(prefix)
		if err := h.createBucket(context.Background(), name, suiteservices.FixtureReusePackage, attribution); err != nil {
			fixture.mu.Unlock()
			t.Fatalf("prepare package s3 bucket: %v", err)
		}
		fixture.bucket = name
	} else if err := h.cleanupPrefixWithDetails(context.Background(), fixture.bucket, "", suiteservices.FixtureReusePrefix, attribution); err != nil {
		fixture.mu.Unlock()
		t.Fatalf("reset package s3 bucket: %v", err)
	}

	t.Cleanup(func() {
		fixture.mu.Unlock()
	})
	return fixture.bucket
}

func (h *Harness) packageBucket(key string) *packageBucket {
	h.packageBucketMu.Lock()
	defer h.packageBucketMu.Unlock()

	if h.packageBuckets == nil {
		h.packageBuckets = make(map[string]*packageBucket)
	}
	fixture := h.packageBuckets[key]
	if fixture == nil {
		fixture = &packageBucket{}
		h.packageBuckets[key] = fixture
	}
	return fixture
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
	start := time.Now()
	if err := h.cleanupPrefix(ctx, bucket, ""); err != nil {
		return err
	}

	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	if err := client.RemoveBucket(ctx, bucket); err != nil {
		return fmt.Errorf("remove bucket %s: %w", bucket, err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventS3BucketCleaned,
		Name:    bucket,
		Details: s3FixtureDetails("bucket", suiteservices.FixtureReusePerTest, fixtureAttribution{}, time.Since(start)),
	})

	return nil
}

func (h *Harness) CleanupPrefix(ctx context.Context, bucket string, prefix string) error {
	return h.cleanupPrefixWithDetails(ctx, bucket, prefix, suiteservices.FixtureReusePrefix, fixtureAttribution{})
}

func (h *Harness) cleanupPrefixWithDetails(ctx context.Context, bucket string, prefix string, reuseScope string, attribution fixtureAttribution) error {
	start := time.Now()
	if err := h.cleanupPrefix(ctx, bucket, prefix); err != nil {
		return err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventS3PrefixCleaned,
		Name:    bucket,
		Details: s3FixtureDetailsWithPrefix("prefix", prefix, reuseScope, attribution, time.Since(start)),
	})
	return nil
}

func (h *Harness) cleanupPrefix(ctx context.Context, bucket string, prefix string) error {
	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	for objectInfo := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if objectInfo.Err != nil {
			return fmt.Errorf("list bucket %s: %w", bucket, objectInfo.Err)
		}
		if err := client.RemoveObject(ctx, bucket, objectInfo.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove object %s/%s: %w", bucket, objectInfo.Key, err)
		}
	}

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

func s3FixtureDetails(strategy string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	return s3FixtureDetailsWithPrefix(strategy, "", reuseScope, attribution, duration)
}

func s3FixtureDetailsWithPrefix(strategy string, prefix string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	if reuseScope == "" {
		reuseScope = suiteservices.FixtureReusePerTest
	}
	details := map[string]any{
		"duration_ms": duration.Milliseconds(),
		"reuse_scope": reuseScope,
		"strategy":    strategy,
		"target":      suiteservices.LookupEnvValue(nil, suiteservices.TargetEnv),
	}
	if prefix != "" {
		details["object_prefix"] = prefix
	}
	if attribution.TestName != "" {
		details["test_name"] = attribution.TestName
	}
	if attribution.CallerFile != "" {
		details["caller_file"] = attribution.CallerFile
	}
	if attribution.CallerPackage != "" {
		details["caller_package"] = attribution.CallerPackage
	}
	return details
}

func fixtureAttributionFor(t testing.TB, harnessPackage string) fixtureAttribution {
	t.Helper()

	attribution := fixtureAttribution{TestName: t.Name()}
	file := callerFile(harnessPackage)
	attribution.CallerFile = repoRelativePath(file)
	attribution.CallerPackage = callerPackage(attribution.CallerFile)
	return attribution
}

func callerFile(harnessPackage string) string {
	pcs := make([]uintptr, 32)
	count := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:count])
	fallback := ""
	for {
		frame, more := frames.Next()
		file := filepath.ToSlash(frame.File)
		if file != "" && !strings.Contains(file, "/internal/testutil/"+harnessPackage+"/") && !strings.Contains(file, "/testing/") && !strings.Contains(file, "/src/runtime/") {
			if fallback == "" {
				fallback = file
			}
			if strings.HasSuffix(file, "_test.go") && !strings.Contains(file, "/internal/testutil/") {
				return file
			}
		}
		if !more {
			break
		}
	}
	return fallback
}

func repoRelativePath(path string) string {
	if path == "" {
		return ""
	}
	root, err := suiteservices.FindRepoRoot()
	if err != nil {
		return filepath.ToSlash(path)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func callerPackage(file string) string {
	if file == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Dir(file))
}
