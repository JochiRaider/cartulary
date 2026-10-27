package s3test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const (
	seaweedFSS3Image                   = "docker.io/chrislusf/seaweedfs:4.48@sha256:4e61d15fd35994cb1e43e1e553dff106794841fd9a99ade2fc8c8bfce4d7872d"
	seaweedFSS3Port                    = "8333/tcp"
	defaultSeaweedFSS3BrowserPortStart = 19000
	defaultSeaweedFSS3BrowserPortEnd   = 19199
	objectStorePortMappingTimeout      = 30 * time.Second
	objectStoreHealthPollInterval      = 500 * time.Millisecond
	objectStoreClientReadyTimeout      = 120 * time.Second
	objectStoreClientAttemptTimeout    = 5 * time.Second
	objectStoreProbeCleanupTimeout     = 5 * time.Second
	objectStoreFixtureCleanupTimeout   = 15 * time.Second
	objectStoreAccessKey               = "cartulary-local"
	objectStoreSecretKey               = "cartulary-disposable-local-secret-32"
	objectStoreProbePayload            = "cartulary-object-store-readiness"
)

func ContainerImage() string {
	return seaweedFSS3Image
}

// Harness borrows a service; it owns only its local client and namespace state.
type Harness struct {
	binding             fixtureBinding
	ops                 fixtureOperations
	Endpoint            string
	AccessKey           string
	SecretKey           string
	Secure              bool
	RootCertificateFile string
	clientMu            sync.Mutex
	connection          *s3transport.Connection

	suiteHash   string
	processHash string
	counter     uint64
	attached    bool
	probeBucket string

	packageBucketMu sync.Mutex
	packageBuckets  map[string]*packageBucket
}

type packageBucket struct {
	mu       sync.Mutex
	bucket   string
	unusable error
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

type s3ReadinessClient struct {
	client *s3.Client
}

type objectStoreReadinessConfig struct {
	Phase           string
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

var errObjectStoreReadinessSemanticMismatch = errors.New("object-store readiness semantic mismatch")

func Start(t testing.TB) *Harness {
	t.Helper()

	harness, err := StartShared(context.Background())
	if err != nil {
		failObjectStoreSetup(t, "start", err)
	}

	return harness
}

func (h *Harness) WaitReady(ctx context.Context) error {
	config := objectStoreReadinessConfig{
		Phase:           "initial_lane",
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
	Phase           string
	Stage           string
	Outcome         string
	Attempts        int
	AttemptTimeouts int
	Elapsed         time.Duration
	CleanupOutcome  string
	CauseCounts     map[string]int
	ContainerState  string
	HealthState     string
	ExitCode        *int
	OOMKilled       *bool
	LastErr         error
	DeadlineExpired bool
}

func (e *objectStoreReadinessError) Error() string {
	if e == nil {
		return ""
	}
	message, present := ReadinessDiagnosticText(e)
	if !present {
		return "object-store readiness failed"
	}
	return message
}

func ReadinessDiagnosticFromError(err error) (suiteservices.ObjectStoreReadinessDiagnostic, bool) {
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) {
		return suiteservices.ObjectStoreReadinessDiagnostic{}, false
	}
	diagnostic := suiteservices.ObjectStoreReadinessDiagnostic{
		SchemaID:            "cartulary.object_store_readiness_diagnostic.v1",
		Phase:               normalizedReadinessPhase(readinessErr.Phase),
		Stage:               normalizedDiagnosticReadinessStage(readinessErr.Stage),
		Outcome:             normalizedReadinessOutcome(readinessErr),
		AttemptCount:        max(readinessErr.Attempts, 1),
		AttemptTimeoutCount: max(readinessErr.AttemptTimeouts, 0),
		ElapsedMS:           max(readinessErr.Elapsed.Milliseconds(), 0),
		CleanupOutcome:      normalizedDiagnosticCleanupOutcome(readinessErr.CleanupOutcome),
		CauseCounts:         cloneCauseCounts(readinessErr.CauseCounts),
		ContainerState:      readinessErr.ContainerState,
		HealthState:         readinessErr.HealthState,
		ExitCode:            readinessErr.ExitCode,
		OOMKilled:           readinessErr.OOMKilled,
	}
	if len(diagnostic.CauseCounts) == 0 {
		diagnostic.CauseCounts = map[string]int{classifyObjectStoreReadinessCause(readinessErr.LastErr): diagnostic.AttemptCount}
	}
	return diagnostic, true
}

// ReadinessDiagnosticText returns the complete bounded public diagnostic for a
// typed readiness failure. It deliberately reconstructs the message from
// normalized fields instead of formatting the wrapped provider or cleanup
// errors, which can contain Docker endpoints and container identities.
func ReadinessDiagnosticText(err error) (string, bool) {
	diagnostic, present := ReadinessDiagnosticFromError(err)
	if !present {
		return "", false
	}
	causes := make([]string, 0, len(diagnostic.CauseCounts))
	for cause, count := range diagnostic.CauseCounts {
		causes = append(causes, fmt.Sprintf("%s:%d", cause, count))
	}
	sort.Strings(causes)
	return fmt.Sprintf(
		"object-store readiness failed: phase=%s stage=%s outcome=%s attempts=%d cause_counts=%s cleanup=%s",
		diagnostic.Phase,
		diagnostic.Stage,
		diagnostic.Outcome,
		diagnostic.AttemptCount,
		strings.Join(causes, ","),
		diagnostic.CleanupOutcome,
	), true
}

func normalizedReadinessPhase(phase string) string {
	switch phase {
	case "initial_lane", "broker_namespace", "package_admission":
		return phase
	default:
		return "initial_lane"
	}
}

func normalizedDiagnosticReadinessStage(stage string) string {
	switch stage {
	case "list", "create_namespace", "put", "head", "delete", "delete_verify":
		return stage
	default:
		return "list"
	}
}

func normalizedReadinessOutcome(readinessErr *objectStoreReadinessError) string {
	if readinessErr == nil {
		return "unavailable"
	}
	switch readinessErr.Outcome {
	case "deadline_expired", "capability_rejected", "cancelled", "unavailable":
		return readinessErr.Outcome
	}
	if readinessErr.DeadlineExpired {
		return "deadline_expired"
	}
	if errors.Is(readinessErr.LastErr, context.Canceled) {
		return "cancelled"
	}
	if isNonRetryableObjectStoreReadinessError(readinessErr.LastErr) {
		return "capability_rejected"
	}
	return "unavailable"
}

func normalizedDiagnosticCleanupOutcome(outcome string) string {
	switch outcome {
	case "completed", "failed":
		return outcome
	default:
		return "not_needed"
	}
}

func cloneCauseCounts(values map[string]int) map[string]int {
	cloned := make(map[string]int, len(values))
	for key, value := range values {
		if value > 0 {
			cloned[key] = value
		}
	}
	return cloned
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
	startedAt := time.Now()
	readyCtx, cancel := context.WithTimeout(ctx, config.ReadyTimeout)
	defer cancel()

	waitPoll := h.ops.waitPoll
	if waitPoll == nil {
		ticker := time.NewTicker(config.PollInterval)
		defer ticker.Stop()
		waitPoll = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				return nil
			}
		}
	}

	attempts := 0
	attemptTimeouts := 0
	causeCounts := make(map[string]int)
	lastFailure := &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed"}
	for {
		attempts++
		attemptCtx, attemptCancel := context.WithTimeout(readyCtx, config.AttemptTimeout)
		client, err := h.operations().readinessClient(attemptCtx, h)
		if err != nil {
			lastFailure = &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed", Cause: err}
		} else {
			lastFailure = runObjectStoreMutationProbe(attemptCtx, client, config)
		}
		attemptCtxErr := attemptCtx.Err()
		attemptCancel()
		if lastFailure == nil {
			return nil
		}
		attemptTimedOut := errors.Is(attemptCtxErr, context.DeadlineExceeded) ||
			errors.Is(lastFailure.Cause, context.DeadlineExceeded)
		if attemptTimedOut {
			attemptTimeouts++
		}
		causeCounts[classifyObjectStoreReadinessCause(lastFailure.Cause)]++
		outcome := ""
		switch {
		case isNonRetryableObjectStoreReadinessError(lastFailure.Cause):
			outcome = "capability_rejected"
		case errors.Is(lastFailure.Cause, context.Canceled):
			outcome = "cancelled"
		case lastFailure.CleanupErr != nil:
			// Uncertain cleanup ends admission. A later successful probe cannot
			// erase this failure, and cleanup cannot replace the primary cause.
			outcome = "unavailable"
			if errors.Is(lastFailure.Cause, context.DeadlineExceeded) {
				outcome = "deadline_expired"
			}
		}
		if outcome != "" {
			return newObjectStoreReadinessError(config.Phase, lastFailure, attempts, attemptTimeouts, time.Since(startedAt), causeCounts, outcome)
		}

		if err := waitPoll(readyCtx); err != nil {
			outcome := "cancelled"
			if errors.Is(err, context.DeadlineExceeded) {
				outcome = "deadline_expired"
			}
			lastFailure.Cause = errors.Join(lastFailure.Cause, err)
			return newObjectStoreReadinessError(config.Phase, lastFailure, attempts, attemptTimeouts, time.Since(startedAt), causeCounts, outcome)
		}
	}
}

func newObjectStoreReadinessError(
	phase string,
	failure *objectStoreProbeFailure,
	attempts int,
	attemptTimeouts int,
	elapsed time.Duration,
	causeCounts map[string]int,
	outcome string,
) error {
	if failure == nil {
		failure = &objectStoreProbeFailure{Stage: "list", CleanupOutcome: "not_needed"}
	}
	return &objectStoreReadinessError{
		Phase:           normalizedReadinessPhase(phase),
		Stage:           failure.Stage,
		Outcome:         outcome,
		Attempts:        attempts,
		AttemptTimeouts: attemptTimeouts,
		Elapsed:         elapsed,
		CleanupOutcome:  failure.CleanupOutcome,
		CauseCounts:     cloneCauseCounts(causeCounts),
		LastErr:         failure.Cause,
		DeadlineExpired: outcome == "deadline_expired",
	}
}

func classifyObjectStoreReadinessCause(err error) string {
	if errors.Is(err, errObjectStoreReadinessSemanticMismatch) {
		return "semantic_mismatch"
	}
	if isNonRetryableObjectStoreReadinessError(err) {
		return "capability_rejected"
	}
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "operation_timeout"
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) {
		if networkErr.Timeout() {
			return "operation_timeout"
		}
		return "transport_unreachable"
	}
	var response smithy.APIError
	if errors.As(err, &response) {
		return "service_error"
	}
	return "unknown_transient"
}

func setOwnedLaneCleanupOutcome(err error, cleanupErr error) {
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) {
		return
	}
	readinessErr.CleanupOutcome = "completed"
	if cleanupErr != nil {
		readinessErr.CleanupOutcome = "failed"
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
		return &objectStoreProbeFailure{Stage: "head", Cause: errObjectStoreReadinessSemanticMismatch}
	}
	if err := client.Delete(ctx, config.Bucket, config.Key); err != nil {
		return &objectStoreProbeFailure{Stage: "delete", Cause: err}
	}
	if _, err := client.HeadSize(ctx, config.Bucket, config.Key); err == nil {
		return &objectStoreProbeFailure{Stage: "delete_verify", Cause: errObjectStoreReadinessSemanticMismatch}
	} else if !isNoSuchObjectError(err) {
		return &objectStoreProbeFailure{Stage: "delete_verify", Cause: err}
	}
	objectMayExist = false
	return nil
}

func isNonRetryableObjectStoreReadinessError(err error) bool {
	if errors.Is(err, cryptography.ErrTLSPeer) || errors.Is(err, cryptography.ErrTLSConfiguration) || errors.Is(err, s3transport.ErrConfiguration) {
		return true
	}
	if err == nil {
		return false
	}

	var response smithy.APIError
	if errors.As(err, &response) {
		switch strings.ToLower(response.ErrorCode()) {
		case "accessdenied", "allaccessdisabled", "authorizationheadermalformed",
			"invalidaccesskeyid", "invalidrequest", "methodnotallowed",
			"notimplemented", "signaturedoesnotmatch", "xnotimplemented":
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
	code := s3transport.ErrorCode(err)
	return code == "NoSuchKey" || code == "NoSuchObject" || code == "NotFound"
}

func (h *Harness) Client(ctx context.Context) (*s3.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	h.clientMu.Lock()
	defer h.clientMu.Unlock()
	if h.connection == nil {
		if !h.Secure {
			return nil, s3transport.ErrConfiguration
		}
		connection, err := s3transport.New(s3transport.Options{Endpoint: h.Endpoint, AccessKey: h.AccessKey, SecretKey: h.SecretKey, RootCertificatePath: h.RootCertificatePath()})
		if err != nil {
			return nil, err
		}
		h.connection = connection
	}
	return h.connection.Client, nil
}
func (c s3ReadinessClient) ListBuckets(ctx context.Context) error {
	_, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err
}
func (c s3ReadinessClient) CreateNamespace(ctx context.Context, bucket string) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	return err
}
func (c s3ReadinessClient) Put(ctx context.Context, bucket, key string, payload []byte) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(payload), ChecksumAlgorithm: types.ChecksumAlgorithmSha256})
	return err
}
func (c s3ReadinessClient) HeadSize(ctx context.Context, bucket, key string) (int64, error) {
	info, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return 0, err
	}
	return aws.ToInt64(info.ContentLength), nil
}
func (c s3ReadinessClient) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	return err
}
func (c s3ReadinessClient) DeleteNamespace(ctx context.Context, bucket string) error {
	_, err := c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	if isNoSuchBucketError(err) {
		return nil
	}
	return err
}

func (h *Harness) BootstrapBucket(ctx context.Context, prefix string) (string, error) {
	return h.bootstrapBucket(ctx, prefix, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) BootstrapBucketT(t testing.TB, prefix string) string {
	t.Helper()
	attribution := fixtureAttributionFor(t, "s3test")
	bucket, err := h.bootstrapBucket(context.Background(), prefix, suiteservices.FixtureReusePerTest, attribution)
	if err != nil {
		failObjectStoreSetup(t, "create_bucket", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), objectStoreFixtureCleanupTimeout)
		defer cancel()
		if err := h.cleanupBucketWithDetails(ctx, bucket, suiteservices.FixtureReusePerTest, attribution); err != nil {
			t.Fatalf("cleanup object-store bucket: %v", err)
		}
	})
	return bucket
}

func (h *Harness) bootstrapBucket(ctx context.Context, prefix string, reuseScope string, attribution fixtureAttribution) (string, error) {
	name := h.nextBucketName(prefix)
	if err := h.createOwnedBucket(ctx, name, reuseScope, attribution); err != nil {
		return "", err
	}
	return name, nil
}

func (h *Harness) CreateBucket(ctx context.Context, name string) error {
	return h.createBucket(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

// createOwnedBucket is only called with a freshly generated, invocation-owned
// name. Persist its cleanup authority before the provider can create it, including
// when a transport error leaves the outcome unknown.
func (h *Harness) createOwnedBucket(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	if err := suiteservices.CheckServiceDependencies(h.binding.environment(), "object_store"); err != nil {
		return err
	}
	if err := suiteservices.RecordEvent(h.binding.environment(), suiteservices.Event{
		Type:    suiteservices.EventS3BucketReserved,
		Name:    name,
		Details: h.fixtureDetails("bucket", reuseScope, attribution, 0),
	}); err != nil {
		return fmt.Errorf("reserve object-store bucket cleanup ownership: %w", err)
	}
	return h.createBucket(ctx, name, reuseScope, attribution)
}

func (h *Harness) createBucket(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	if err := suiteservices.CheckServiceDependencies(h.binding.environment(), "object_store"); err != nil {
		return err
	}
	start := time.Now()
	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(name)}); err != nil {
		exists, bucketErr := s3transport.BucketExists(ctx, client, name)
		if bucketErr != nil || !exists {
			return fmt.Errorf("create bucket %s: %w", name, err)
		}
	}
	return h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventS3BucketCreated,
		Name:    name,
		Details: h.fixtureDetails("bucket", reuseScope, attribution, time.Since(start)),
	})
}

func (h *Harness) PreparePackageBucketT(t testing.TB, prefix string) string {
	t.Helper()

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
	if fixture.unusable != nil {
		fixture.mu.Unlock()
		return "", nil, fmt.Errorf("package s3 bucket generation is unusable: %w", fixture.unusable)
	}
	fail := func(stage string, cause error) (string, func(), error) {
		fixture.unusable = fmt.Errorf("%s package s3 bucket: %w", stage, cause)
		fixture.mu.Unlock()
		return "", nil, fixture.unusable
	}
	if err := ctx.Err(); err != nil {
		fixture.mu.Unlock()
		return "", nil, err
	}

	if fixture.bucket == "" {
		fixture.bucket = h.nextBucketName(prefix)
		if err := h.operations().createPackage(ctx, h, fixture.bucket, attribution); err != nil {
			return fail("prepare", err)
		}
	} else if err := h.operations().cleanupPackage(ctx, h, fixture.bucket, attribution); err != nil {
		return fail("reset", err)
	}
	if err := h.operations().readyPackage(ctx, h, fixture.bucket); err != nil {
		return fail("admit", err)
	}

	return fixture.bucket, func() { fixture.mu.Unlock() }, nil
}

func failObjectStoreSetup(t testing.TB, defaultStage string, err error) {
	t.Helper()
	if diagnostic, ok := ReadinessDiagnosticText(err); ok {
		t.Log(diagnostic)
	}
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

	_, err = client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(payload), ChecksumAlgorithm: types.ChecksumAlgorithmSha256})
	if err != nil {
		return nil, fmt.Errorf("put object %s/%s: %w", bucket, key, err)
	}

	object, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("get object %s/%s: %w", bucket, key, err)
	}
	defer object.Body.Close()

	data, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, fmt.Errorf("read object %s/%s: %w", bucket, key, err)
	}

	return data, nil
}

func (h *Harness) CleanupBucket(ctx context.Context, bucket string) error {
	return h.cleanupBucketWithDetails(ctx, bucket, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) cleanupBucketWithDetails(ctx context.Context, bucket string, reuseScope string, attribution fixtureAttribution) error {
	ctx, cancel := context.WithTimeout(ctx, objectStoreFixtureCleanupTimeout)
	defer cancel()
	start := time.Now()
	if err := h.cleanupPrefix(ctx, bucket, ""); err != nil {
		return err
	}

	client, err := h.Client(ctx)
	if err != nil {
		return err
	}

	if _, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)}); err != nil && !isNoSuchBucketError(err) {
		return fmt.Errorf("remove bucket %s: %w", bucket, err)
	}
	return h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventS3BucketCleaned,
		Name:    bucket,
		Details: h.fixtureDetails("bucket", reuseScope, attribution, time.Since(start)),
	})
}

func (h *Harness) CleanupPrefix(ctx context.Context, bucket string, prefix string) error {
	return h.cleanupPrefixWithDetails(ctx, bucket, prefix, suiteservices.FixtureReusePrefix, fixtureAttribution{})
}

// ResetBucket removes the complete contents of an owned, reusable namespace
// without changing its identity, then proves that the exact mutation path is
// ready before the next borrower is admitted.
func (h *Harness) ResetBucket(ctx context.Context, bucket string) error {
	if err := suiteservices.CheckServiceDependencies(h.binding.environment(), "object_store"); err != nil {
		return err
	}
	if strings.TrimSpace(bucket) == "" {
		return errors.New("object-store bucket reset requires an exact bucket")
	}
	if err := h.cleanupPrefixWithDetails(
		ctx,
		bucket,
		"",
		suiteservices.FixtureReusePrefix,
		fixtureAttribution{},
	); err != nil {
		return fmt.Errorf("empty object-store bucket %s: %w", bucket, err)
	}
	if err := h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
		Phase:           "package_admission",
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: false,
		ListFirst:       false,
		Bucket:          bucket,
		Key:             h.nextProbeKey(),
	}); err != nil {
		return fmt.Errorf("admit reset object-store bucket %s: %w", bucket, err)
	}
	return nil
}

func (h *Harness) cleanupPrefixWithDetails(ctx context.Context, bucket string, prefix string, reuseScope string, attribution fixtureAttribution) error {
	start := time.Now()
	if err := h.cleanupPrefix(ctx, bucket, prefix); err != nil {
		return err
	}
	return h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventS3PrefixCleaned,
		Name:    bucket,
		Details: h.fixtureDetailsWithPrefix("prefix", prefix, reuseScope, attribution, time.Since(start)),
	})
}

func (h *Harness) cleanupPrefix(ctx context.Context, bucket string, prefix string) error {
	ctx, cancel := context.WithTimeout(ctx, objectStoreFixtureCleanupTimeout)
	defer cancel()
	client, err := h.Client(ctx)
	if err != nil {
		return err
	}
	return cleanupObjectPrefix(ctx, client, bucket, prefix)
}

func isNoSuchBucketError(err error) bool {
	if err == nil {
		return false
	}
	return s3transport.ErrorCode(err) == "NoSuchBucket"
}

func (h *Harness) Env(bucket string) map[string]string {
	return map[string]string{
		objectstore.EndpointEnv:        h.Endpoint,
		objectstore.AccessKeyEnv:       h.AccessKey,
		objectstore.SecretKeyEnv:       h.SecretKey,
		objectstore.SecureEnv:          strconv.FormatBool(h.Secure),
		objectstore.BucketEnv:          bucket,
		objectstore.RootCertificateEnv: h.RootCertificatePath(),
	}
}

func (h *Harness) EnvForServiceRef(serviceRef string, bucket string) map[string]string {
	keys, err := objectstore.EnvKeysForServiceRef(serviceRef)
	if err != nil {
		return map[string]string{}
	}

	return map[string]string{
		keys.Endpoint:            h.Endpoint,
		keys.AccessKey:           h.AccessKey,
		keys.SecretKey:           h.SecretKey,
		keys.Secure:              strconv.FormatBool(h.Secure),
		keys.Bucket:              bucket,
		keys.RootCertificatePath: h.RootCertificatePath(),
	}
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

func (h *Harness) nextBucketName(prefix string) string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = suiteservices.SuiteHash(h.binding.environment())
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
		suiteHash = suiteservices.SuiteHash(h.binding.environment())
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

func (h *Harness) recordEvent(event suiteservices.Event) error {
	return suiteservices.RecordEvent(h.binding.environment(), event)
}

func (h *Harness) fixtureDetails(strategy string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	return h.fixtureDetailsWithPrefix(strategy, "", reuseScope, attribution, duration)
}

func (h *Harness) fixtureDetailsWithPrefix(strategy string, prefix string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	if reuseScope == "" {
		reuseScope = suiteservices.FixtureReusePerTest
	}
	details := map[string]any{
		"duration_ms": duration.Milliseconds(),
		"reuse_scope": reuseScope,
		"strategy":    strategy,
		"target":      suiteservices.LookupEnvValue(h.binding.environment(), suiteservices.TargetEnv),
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

// Close releases only this borrower's transport. Namespace/quarantine state and
// all physical owner resources remain intact; a later operation may reopen it.
func (h *Harness) Close(_ context.Context) error {
	if h == nil {
		return nil
	}
	h.clientMu.Lock()
	defer h.clientMu.Unlock()
	if h.connection == nil {
		return nil
	}
	err := h.connection.Close()
	h.connection = nil
	return err
}
