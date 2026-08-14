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
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

const (
	seaweedFSS3Image                   = "docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9"
	seaweedFSS3Port                    = "8333/tcp"
	defaultSeaweedFSS3BrowserPortStart = 19000
	defaultSeaweedFSS3BrowserPortEnd   = 19199
	objectStorePortMappingTimeout      = 30 * time.Second
	objectStoreHealthPollInterval      = 500 * time.Millisecond
	objectStoreClientReadyTimeout      = 120 * time.Second
	objectStoreClientAttemptTimeout    = 5 * time.Second
	objectStoreProbeCleanupTimeout     = 5 * time.Second
	objectStoreAccessKey               = "cartulary-local"
	objectStoreSecretKey               = "cartulary-local-secret"
	objectStoreProbePayload            = "cartulary-object-store-readiness"
)

func ContainerImage() string {
	return seaweedFSS3Image
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
	probeBucket string

	packageBucketMu sync.Mutex
	packageBuckets  map[string]*packageBucket
}

type StartOptions struct {
	Labels         map[string]string
	Observer       testcontainersx.StartObserver
	AttemptTimeout time.Duration
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

type objectStoreReadinessClient interface {
	ListBuckets(context.Context) error
	CreateNamespace(context.Context, string) error
	Put(context.Context, string, string, []byte) error
	HeadSize(context.Context, string, string) (int64, error)
	Delete(context.Context, string, string) error
	DeleteNamespace(context.Context, string) error
}

type minioReadinessClient struct {
	client *minio.Client
}

type objectStoreReadinessConfig struct {
	ReadyTimeout    time.Duration
	PollInterval    time.Duration
	AttemptTimeout  time.Duration
	CreateNamespace bool
	ListFirst       bool
	Bucket          string
	Key             string
}

type objectStoreProbeFailure struct {
	Stage          string
	CleanupOutcome string
	Cause          error
	CleanupErr     error
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
	newReadinessClientFn = func(ctx context.Context, harness *Harness) (objectStoreReadinessClient, error) {
		client, err := harness.Client(ctx)
		if err != nil {
			return nil, err
		}
		return minioReadinessClient{client: client}, nil
	}
	startPreflightFn             func(context.Context) (string, error)
	startSleepFn                 func(context.Context, time.Duration) error
	preparePackageCreateBucketFn = func(ctx context.Context, h *Harness, name string, attribution fixtureAttribution) error {
		return h.createBucket(ctx, name, suiteservices.FixtureReusePackage, attribution)
	}
	preparePackageCleanupFn = func(ctx context.Context, h *Harness, bucket string, attribution fixtureAttribution) error {
		return h.cleanupPrefixWithDetails(ctx, bucket, "", suiteservices.FixtureReusePrefix, attribution)
	}
	preparePackageWaitReadyFn = func(ctx context.Context, h *Harness, bucket string) error {
		return h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
			ReadyTimeout:    objectStoreClientReadyTimeout,
			PollInterval:    objectStoreHealthPollInterval,
			AttemptTimeout:  objectStoreClientAttemptTimeout,
			CreateNamespace: false,
			ListFirst:       false,
			Bucket:          bucket,
			Key:             h.nextProbeKey(),
		})
	}
)

func Start(t testing.TB) *Harness {
	t.Helper()
	suiteservices.RequireServiceDependencies(t, "s3test", "object_store")

	harness, err := StartShared(context.Background())
	if err != nil {
		failObjectStoreSetup(t, "start", err)
	}

	return harness
}

func StartShared(ctx context.Context) (*Harness, error) {
	if err := suiteservices.CheckServiceDependencies(nil, "object_store"); err != nil {
		return nil, err
	}
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
	return StartOwnedWithLabels(ctx, nil)
}

func StartOwnedWithLabels(ctx context.Context, labels map[string]string) (*Harness, error) {
	return StartOwnedWithOptions(ctx, StartOptions{Labels: labels})
}

func StartOwnedWithOptions(ctx context.Context, options StartOptions) (*Harness, error) {
	if err := suiteservices.CheckServiceDependencies(nil, "object_store"); err != nil {
		return nil, err
	}
	return startHarnessWithOptions(ctx, options)
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

func startHarness(ctx context.Context, labels map[string]string) (*Harness, error) {
	return startHarnessWithOptions(ctx, StartOptions{Labels: labels})
}

func startHarnessWithOptions(ctx context.Context, options StartOptions) (*Harness, error) {
	req := testcontainers.ContainerRequest{
		Image:        seaweedFSS3Image,
		ExposedPorts: []string{seaweedFSS3Port},
		Cmd: []string{
			"server",
			"-dir=/data",
			"-s3",
			"-s3.port=8333",
			"-s3.allowedOrigins=" + seaweedFSS3AllowedOrigins(),
			"-s3.port.iceberg=0",
			"-webdav=false",
		},
		Env: map[string]string{
			"AWS_ACCESS_KEY_ID":     objectStoreAccessKey,
			"AWS_SECRET_ACCESS_KEY": objectStoreSecretKey,
		},
		WaitingFor: objectStorePortWaitStrategy(),
	}
	if len(options.Labels) > 0 {
		req.Labels = cloneLabels(options.Labels)
	}

	harness, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service:        "object-store testcontainer",
		Image:          seaweedFSS3Image,
		MaxAttempts:    testcontainersx.DefaultMaxAttempts,
		AttemptTimeout: options.AttemptTimeout,
		RetryBackoff:   testcontainersx.DefaultRetryBackoff,
		Preflight:      startPreflightFn,
		Retryable:      isRetryableObjectStoreStartupFailure,
		Sleep:          startSleepFn,
		Observer:       options.Observer,
	}, func(ctx context.Context) (*Harness, error) {
		return startHarnessAttempt(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return harness, nil
}

func seaweedFSS3AllowedOrigins() string {
	if origins := strings.TrimSpace(suiteservices.LookupEnvValue(nil, "OBJECT_STORE_CORS_ALLOWED_ORIGINS")); origins != "" {
		return origins
	}
	if origins := strings.TrimSpace(suiteservices.LookupEnvValue(nil, "OBJECT_STORE_CORS_ORIGIN")); origins != "" {
		return origins
	}
	return defaultSeaweedFSS3AllowedOrigins()
}

func defaultSeaweedFSS3AllowedOrigins() string {
	origins := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	for port := defaultSeaweedFSS3BrowserPortStart; port <= defaultSeaweedFSS3BrowserPortEnd; port++ {
		origins = append(origins, fmt.Sprintf("http://localhost:%d", port), fmt.Sprintf("http://127.0.0.1:%d", port))
	}
	return strings.Join(origins, ",")
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
		return nil, fmt.Errorf("resolve object-store host: %w", err)
	}

	port, err := container.MappedPort(ctx, seaweedFSS3Port)
	if err != nil {
		return nil, fmt.Errorf("resolve object-store mapped port: %w", err)
	}

	harness := &Harness{
		Container:   container,
		Endpoint:    host + ":" + port.Port(),
		AccessKey:   objectStoreAccessKey,
		SecretKey:   objectStoreSecretKey,
		Secure:      false,
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
	}

	if err := waitReadyFn(ctx, harness); err != nil {
		return nil, fmt.Errorf("wait for object-store readiness: %w", err)
	}

	attemptSucceeded = true
	return harness, nil
}

func objectStorePortWaitStrategy() *wait.HostPortStrategy {
	return wait.ForMappedPort(seaweedFSS3Port).WithStartupTimeout(objectStorePortMappingTimeout)
}

func startAttachedHarness(ctx context.Context) (*Harness, bool, error) {
	endpoint := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3EndpointEnv))
	accessKey := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3AccessKeyEnv))
	secretKey := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3SecretKeyEnv))
	if endpoint == "" && accessKey == "" && secretKey == "" {
		return nil, false, nil
	}
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, false, fmt.Errorf("attach object-store harness: %s, %s, and %s must all be set", suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv)
	}

	secure, err := suiteservices.ParseBool(suiteservices.LookupEnvValue(nil, suiteservices.S3SecureEnv))
	if err != nil {
		return nil, false, fmt.Errorf("attach object-store harness: %w", err)
	}

	harness := &Harness{
		Endpoint:    endpoint,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Secure:      secure,
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
		attached:    true,
		probeBucket: strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.S3ProbeBucketEnv)),
	}
	if err := verifyAttachedFn(ctx, harness); err != nil {
		return nil, false, fmt.Errorf("attach object-store harness: authenticated readiness: %w", err)
	}
	recordSuiteEvent(suiteservices.Event{Type: suiteservices.EventS3Attach})
	return harness, true, nil
}

func (h *Harness) WaitReady(ctx context.Context) error {
	config := objectStoreReadinessConfig{
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          h.nextBucketName("readiness"),
		Key:             h.nextProbeKey(),
	}
	if h.attached && h.probeBucket != "" {
		config.CreateNamespace = false
		config.Bucket = h.probeBucket
	}
	return h.waitForObjectStoreReadiness(ctx, config)
}

type objectStoreReadinessError struct {
	Stage           string
	Attempts        int
	CleanupOutcome  string
	LastErr         error
	DeadlineExpired bool
}

func (e *objectStoreReadinessError) Error() string {
	if e == nil {
		return ""
	}
	reason := "unavailable"
	if e.DeadlineExpired {
		reason = "deadline_expired"
	} else if isNonRetryableObjectStoreReadinessError(e.LastErr) {
		reason = "capability_rejected"
	}
	return fmt.Sprintf("object-store readiness failed: stage=%s attempts=%d cleanup=%s reason=%s", e.Stage, e.Attempts, e.CleanupOutcome, reason)
}

func (e *objectStoreReadinessError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.LastErr
}

func isRetryableObjectStoreStartupFailure(_ error) bool {
	// Readiness polling is the terminal phase of a started lane. Only failures
	// before this phase may cause StartWithRetry to create a replacement lane.
	return false
}

func (h *Harness) waitForObjectStoreReadiness(ctx context.Context, config objectStoreReadinessConfig) error {
	readyCtx, cancel := context.WithTimeout(ctx, config.ReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	attempts := 0
	lastFailure := &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed"}
	for {
		attempts++
		attemptCtx, attemptCancel := context.WithTimeout(readyCtx, config.AttemptTimeout)
		client, err := newReadinessClientFn(attemptCtx, h)
		if err != nil {
			lastFailure = &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed", Cause: err}
		} else {
			lastFailure = runObjectStoreMutationProbe(attemptCtx, client, config)
		}
		attemptCancel()
		if lastFailure == nil {
			return nil
		}
		if isNonRetryableObjectStoreReadinessError(lastFailure.Cause) ||
			isNonRetryableObjectStoreReadinessError(lastFailure.CleanupErr) {
			return newObjectStoreReadinessError(lastFailure, attempts, false)
		}

		select {
		case <-readyCtx.Done():
			return newObjectStoreReadinessError(lastFailure, attempts, errors.Is(readyCtx.Err(), context.DeadlineExceeded))
		case <-ticker.C:
		}
	}
}

func newObjectStoreReadinessError(failure *objectStoreProbeFailure, attempts int, deadlineExpired bool) error {
	if failure == nil {
		failure = &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed"}
	}
	return &objectStoreReadinessError{
		Stage:           failure.Stage,
		Attempts:        attempts,
		CleanupOutcome:  failure.CleanupOutcome,
		LastErr:         failure.Cause,
		DeadlineExpired: deadlineExpired,
	}
}

func runObjectStoreMutationProbe(ctx context.Context, client objectStoreReadinessClient, config objectStoreReadinessConfig) (failure *objectStoreProbeFailure) {
	namespaceMayExist := false
	objectMayExist := false
	cleanupOutcome := "not_needed"

	defer func() {
		if !objectMayExist && !namespaceMayExist {
			if failure != nil && failure.CleanupOutcome == "" {
				failure.CleanupOutcome = cleanupOutcome
			}
			return
		}

		cleanupCtx, cancel := context.WithTimeout(context.Background(), objectStoreProbeCleanupTimeout)
		defer cancel()
		cleanupErr := client.Delete(cleanupCtx, config.Bucket, config.Key)
		if namespaceMayExist {
			cleanupErr = errors.Join(cleanupErr, client.DeleteNamespace(cleanupCtx, config.Bucket))
		}
		if cleanupErr != nil {
			cleanupOutcome = "failed"
		} else {
			cleanupOutcome = "completed"
		}
		if failure == nil && cleanupErr != nil {
			failure = &objectStoreProbeFailure{Stage: "delete_verify", Cause: cleanupErr, CleanupErr: cleanupErr}
		}
		if failure != nil {
			failure.CleanupOutcome = cleanupOutcome
			failure.CleanupErr = cleanupErr
		}
	}()

	if config.ListFirst {
		if err := client.ListBuckets(ctx); err != nil {
			return &objectStoreProbeFailure{Stage: "list", Cause: err}
		}
	}
	if config.CreateNamespace {
		// The generated namespace is run-owned before the create call. A lost
		// response may report failure after creation, so cleanup must still prove
		// that the namespace is absent.
		namespaceMayExist = true
		if err := client.CreateNamespace(ctx, config.Bucket); err != nil {
			return &objectStoreProbeFailure{Stage: "create_namespace", Cause: err}
		}
	}

	payload := []byte(objectStoreProbePayload)
	objectMayExist = true
	if err := client.Put(ctx, config.Bucket, config.Key, payload); err != nil {
		return &objectStoreProbeFailure{Stage: "put", Cause: err}
	}
	size, err := client.HeadSize(ctx, config.Bucket, config.Key)
	if err != nil {
		return &objectStoreProbeFailure{Stage: "head", Cause: err}
	}
	if size != int64(len(payload)) {
		return &objectStoreProbeFailure{Stage: "head", Cause: fmt.Errorf("probe size mismatch")}
	}
	if err := client.Delete(ctx, config.Bucket, config.Key); err != nil {
		return &objectStoreProbeFailure{Stage: "delete", Cause: err}
	}
	if _, err := client.HeadSize(ctx, config.Bucket, config.Key); err == nil {
		return &objectStoreProbeFailure{Stage: "delete_verify", Cause: fmt.Errorf("probe object remains visible")}
	} else if !isNoSuchObjectError(err) {
		return &objectStoreProbeFailure{Stage: "delete_verify", Cause: err}
	}
	objectMayExist = false
	return nil
}

func isNonRetryableObjectStoreReadinessError(err error) bool {
	if err == nil {
		return false
	}

	var response minio.ErrorResponse
	if errors.As(err, &response) {
		switch strings.ToLower(response.Code) {
		case "accessdenied", "allaccessdisabled", "authorizationheadermalformed",
			"invalidaccesskeyid", "invalidrequest", "methodnotallowed",
			"notimplemented", "signaturedoesnotmatch", "xnotimplemented":
			return true
		}
		if response.StatusCode == 401 || response.StatusCode == 403 || response.StatusCode == 405 || response.StatusCode == 501 {
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

func isNoSuchObjectError(err error) bool {
	if err == nil {
		return false
	}
	response := minio.ToErrorResponse(err)
	return strings.EqualFold(response.Code, "NoSuchKey") ||
		strings.EqualFold(response.Code, "NoSuchObject") || strings.EqualFold(response.Code, "NotFound")
}

func (h *Harness) Client(ctx context.Context) (*minio.Client, error) {
	_ = ctx
	return minio.New(h.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(h.AccessKey, h.SecretKey, ""),
		Secure: h.Secure,
	})
}

func (c minioReadinessClient) ListBuckets(ctx context.Context) error {
	_, err := c.client.ListBuckets(ctx)
	return err
}

func (c minioReadinessClient) CreateNamespace(ctx context.Context, bucket string) error {
	return c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func (c minioReadinessClient) Put(ctx context.Context, bucket string, key string, payload []byte) error {
	_, err := c.client.PutObject(ctx, bucket, key, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{})
	return err
}

func (c minioReadinessClient) HeadSize(ctx context.Context, bucket string, key string) (int64, error) {
	info, err := c.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	return info.Size, err
}

func (c minioReadinessClient) Delete(ctx context.Context, bucket string, key string) error {
	return c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (c minioReadinessClient) DeleteNamespace(ctx context.Context, bucket string) error {
	err := c.client.RemoveBucket(ctx, bucket)
	if isNoSuchBucketError(err) {
		return nil
	}
	return err
}

func (h *Harness) BootstrapBucket(ctx context.Context, prefix string) (string, error) {
	name := h.nextBucketName(prefix)
	if err := h.createBucket(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		return "", err
	}
	return name, nil
}

// BootstrapProbeBucket creates a broker-owned namespace and proves the exact
// mutation path that attached child processes will use within unique prefixes.
// The broker retains ownership of the bucket for the suite lifetime.
func (h *Harness) BootstrapProbeBucket(ctx context.Context) (string, error) {
	bucket, err := h.BootstrapBucket(ctx, "readiness-shared")
	if err != nil {
		return "", fmt.Errorf("create broker object-store probe namespace: %w", err)
	}
	if err := h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: false,
		ListFirst:       false,
		Bucket:          bucket,
		Key:             h.nextProbeKey(),
	}); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), objectStoreProbeCleanupTimeout)
		cleanupErr := h.CleanupBucket(cleanupCtx, bucket)
		cancel()
		return "", errors.Join(fmt.Errorf("admit broker object-store probe namespace: %w", err), cleanupErr)
	}
	return bucket, nil
}

func (h *Harness) CreateBucket(ctx context.Context, name string) error {
	return h.createBucket(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) createBucket(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	if err := suiteservices.CheckServiceDependencies(nil, "object_store"); err != nil {
		return err
	}
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
	suiteservices.RequireServiceDependencies(t, "s3test", "object_store")

	attribution := fixtureAttributionFor(t, "s3test")
	bucket, release, err := h.preparePackageBucket(context.Background(), prefix, attribution)
	if err != nil {
		failObjectStoreSetup(t, "create_bucket", err)
	}
	t.Cleanup(release)
	return bucket
}

func (h *Harness) preparePackageBucket(
	ctx context.Context,
	prefix string,
	attribution fixtureAttribution,
) (string, func(), error) {
	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeBucket(prefix)
	}
	fixture := h.packageBucket(key)
	fixture.mu.Lock()

	if fixture.bucket == "" {
		name := h.nextBucketName(prefix)
		if err := preparePackageCreateBucketFn(ctx, h, name, attribution); err != nil {
			fixture.mu.Unlock()
			return "", nil, fmt.Errorf("prepare package s3 bucket: %w", err)
		}
		fixture.bucket = name
	} else if err := preparePackageCleanupFn(ctx, h, fixture.bucket, attribution); err != nil {
		fixture.mu.Unlock()
		return "", nil, fmt.Errorf("reset package s3 bucket: %w", err)
	}
	if err := preparePackageWaitReadyFn(ctx, h, fixture.bucket); err != nil {
		fixture.mu.Unlock()
		return "", nil, fmt.Errorf("admit package s3 bucket: %w", err)
	}

	return fixture.bucket, func() { fixture.mu.Unlock() }, nil
}

func failObjectStoreSetup(t testing.TB, defaultStage string, err error) {
	t.Helper()
	testfailure.Fail(t, objectStoreSetupEnvelope(defaultStage, err))
}

func objectStoreSetupEnvelope(defaultStage string, err error) testfailure.Envelope {
	failureClass := "harness"
	failureReason := "fixture_error"
	stage := defaultStage
	attempts := 0
	cleanup := "not_required"

	var readinessErr *objectStoreReadinessError
	if errors.As(err, &readinessErr) {
		stage = normalizedReadinessStage(readinessErr.Stage)
		attempts = readinessErr.Attempts
		cleanup = normalizedCleanupOutcome(readinessErr.CleanupOutcome)
		if readinessErr.DeadlineExpired {
			failureClass = "infra"
			failureReason = "service_readiness_timeout"
		}
	}
	if errors.Is(err, context.Canceled) {
		failureClass = "interrupted"
		failureReason = "cancelled_or_interrupted"
	}
	var dependencyErr *suiteservices.ServiceDependencyError
	if errors.As(err, &dependencyErr) {
		stage = "dependency_guard"
	}
	return testfailure.NewEnvelope(
		failureClass,
		failureReason,
		"s3test",
		"object_store",
		stage,
		attempts,
		cleanup,
	)
}

func normalizedReadinessStage(stage string) string {
	switch stage {
	case "create_namespace":
		return "create_bucket"
	case "delete_verify":
		return "not_found"
	case "list":
		return "capability"
	case "put", "head", "delete":
		return stage
	default:
		return "capability"
	}
}

func normalizedCleanupOutcome(outcome string) string {
	switch outcome {
	case "completed", "failed":
		return outcome
	default:
		return "not_required"
	}
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

	if err := client.RemoveBucket(ctx, bucket); err != nil && !isNoSuchBucketError(err) {
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
			if isNoSuchBucketError(objectInfo.Err) {
				return nil
			}
			return fmt.Errorf("list bucket %s: %w", bucket, objectInfo.Err)
		}
		if err := client.RemoveObject(ctx, bucket, objectInfo.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove object %s/%s: %w", bucket, objectInfo.Key, err)
		}
	}

	return nil
}

func isNoSuchBucketError(err error) bool {
	if err == nil {
		return false
	}
	response := minio.ToErrorResponse(err)
	return strings.EqualFold(response.Code, "NoSuchBucket")
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

func (h *Harness) ContainerID() string {
	if h == nil || h.Container == nil {
		return ""
	}
	return h.Container.GetContainerID()
}

func cloneLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
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

func (h *Harness) nextProbeKey() string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	processHash := h.processHash
	if processHash == "" {
		processHash = suiteservices.ProcessHash()
	}
	return fmt.Sprintf(".cartulary-readiness/%s-%s-%06d", suiteHash, processHash, atomic.AddUint64(&h.counter, 1))
}

func truncateBucketName(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return strings.Trim(value[:max], "-")
}

func verifyAttachedHarness(ctx context.Context, harness *Harness) error {
	return harness.WaitReady(ctx)
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
