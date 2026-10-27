package s3test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// OwnedStore is the physical service owner. Fixture grants borrowed operations
// without giving borrowers container termination or access to private TLS keys.
type OwnedStore struct {
	Container    testcontainers.Container
	Fixture      *Harness
	tlsDirectory string
	closeOnce    sync.Once
	closeErr     error
}

type StartOptions struct {
	Labels         map[string]string
	Observer       testcontainersx.StartObserver
	AttemptTimeout time.Duration
	BeforeAcquire  func(*OwnedStore) error
	Settled        func(*OwnedStore) error
	Acquired       func(*OwnedStore) error
}

type startupOperations struct {
	container func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error)
	ready     func(context.Context, *Harness) error
	preflight func(context.Context) (string, error)
	sleep     func(context.Context, time.Duration) error
}

func defaultStartupOperations() startupOperations {
	return startupOperations{container: testcontainersx.GenericContainer, ready: func(ctx context.Context, h *Harness) error { return h.WaitReady(ctx) }}
}

func StartOwned(ctx context.Context, options StartOptions) (*OwnedStore, error) {
	return startOwned(ctx, options, defaultStartupOperations())
}

func (h *OwnedStore) Close(ctx context.Context) error {
	if h == nil {
		return nil
	}
	h.closeOnce.Do(func() {
		h.closeErr = h.Fixture.Close(ctx)
		if h.Container != nil {
			h.closeErr = errors.Join(h.closeErr, h.Container.Terminate(ctx))
		}
		if h.closeErr == nil && h.tlsDirectory != "" {
			h.closeErr = os.RemoveAll(h.tlsDirectory)
		}
	})
	return h.closeErr
}

func startOwned(ctx context.Context, options StartOptions, ops startupOperations) (*OwnedStore, error) {
	binding := bindFixture(processFixtureInputs())
	if err := suiteservices.CheckServiceDependencies(binding.environment(), "object_store"); err != nil {
		return nil, err
	}
	req := testcontainers.ContainerRequest{
		Image:        seaweedFSS3Image,
		ExposedPorts: []string{seaweedFSS3Port},
		Cmd: []string{
			"server",
			// This disposable fixture has one master; bootstrap it explicitly.
			"-master.peers=none",
			"-dir=/data",
			"-s3",
			"-s3.port=8333",
			"-s3.allowedOrigins=" + seaweedFSS3AllowedOrigins(nil),
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
		Preflight:      ops.preflight,
		Retryable:      isRetryableObjectStoreStartupFailure,
		Sleep:          ops.sleep,
		Observer:       options.Observer,
	}, func(ctx context.Context) (*OwnedStore, error) {
		return startOwnedAttempt(ctx, req, options, ops, binding)
	})
	if err != nil {
		return harness, err
	}
	if err := ops.ready(ctx, harness.Fixture); err != nil {
		attachOwnedContainerDiagnostic(err, harness)
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cleanupErr := harness.Close(cleanupCtx)
		cancel()
		if cleanupErr == nil && harness.Container != nil && options.Settled != nil {
			cleanupErr = options.Settled(harness)
		}
		setOwnedLaneCleanupOutcome(err, cleanupErr)
		if cleanupErr != nil {
			return harness, &unsettledStartup{errors.Join(err, cleanupErr)}
		}
		return nil, fmt.Errorf("wait for object-store readiness: %w", err)
	}
	return harness, nil
}

func seaweedFSS3AllowedOrigins(inputs map[string]string) string {
	if origins := strings.TrimSpace(suiteservices.LookupEnvValue(inputs, "OBJECT_STORE_CORS_ALLOWED_ORIGINS")); origins != "" {
		return origins
	}
	if origins := strings.TrimSpace(suiteservices.LookupEnvValue(inputs, "OBJECT_STORE_CORS_ORIGIN")); origins != "" {
		return origins
	}
	return defaultSeaweedFSS3AllowedOrigins()
}

func defaultSeaweedFSS3AllowedOrigins() string {
	origins := []string{"https://localhost:5173", "https://127.0.0.1:5173"}
	for port := defaultSeaweedFSS3BrowserPortStart; port <= defaultSeaweedFSS3BrowserPortEnd; port++ {
		origins = append(origins, fmt.Sprintf("https://localhost:%d", port), fmt.Sprintf("https://127.0.0.1:%d", port))
	}
	return strings.Join(origins, ",")
}

// Only an unresolved attempt carries a partial handle. Its ordinary cleanup
// receipt is memoized, and cannot authorize a replacement physical generation.
type unsettledStartup struct{ error }

func (e *unsettledStartup) Unwrap() error        { return e.error }
func (e *unsettledStartup) RetryForbidden() bool { return true }

func startOwnedAttempt(ctx context.Context, req testcontainers.ContainerRequest, options StartOptions, ops startupOperations, binding fixtureBinding) (partial *OwnedStore, retErr error) {
	req, directory, err := prepareObjectStoreTLS(req, binding.environment())
	if err != nil {
		return nil, err
	}
	harness := &OwnedStore{tlsDirectory: directory, Fixture: &Harness{binding: binding, AccessKey: objectStoreAccessKey, SecretKey: objectStoreSecretKey, Secure: true, RootCertificateFile: filepath.Join(directory, "root.pem"), suiteHash: suiteservices.SuiteHash(binding.environment()), processHash: suiteservices.ProcessHash()}}
	intentRecorded, providerCalled := false, false
	defer func() {
		if retErr == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cleanupErr := harness.Close(cleanupCtx)
		cancel()
		if cleanupErr == nil && (harness.Container != nil || (intentRecorded && !providerCalled)) && options.Settled != nil {
			cleanupErr = options.Settled(harness)
		}
		partial = nil
		if cleanupErr != nil {
			partial = harness
			retErr = &unsettledStartup{errors.Join(retErr, cleanupErr)}
		}
	}()
	if options.BeforeAcquire != nil {
		if err := options.BeforeAcquire(harness); err != nil {
			return nil, err
		}
	}
	intentRecorded = options.BeforeAcquire != nil
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	providerCalled = true
	createCtx, finishCreate := context.WithCancel(ctx)
	if options.BeforeAcquire != nil {
		if deadline, ok := ctx.Deadline(); ok {
			finishCreate()
			createCtx, finishCreate = context.WithDeadline(context.WithoutCancel(ctx), deadline)
		}
	}
	container, err := ops.container(createCtx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: false})
	finishCreate()
	harness.Container = container
	if container != nil && options.Acquired != nil {
		if publicationErr := options.Acquired(harness); publicationErr != nil {
			return nil, errors.Join(err, publicationErr)
		}
	}
	if err != nil {
		return nil, err
	}
	if container == nil {
		return nil, errors.New("object-store provider returned no container")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := container.Start(ctx); err != nil {
		return nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve object-store host: %w", err)
	}
	port, err := container.MappedPort(ctx, seaweedFSS3Port)
	if err != nil {
		return nil, fmt.Errorf("resolve object-store mapped port: %w", err)
	}
	harness.Fixture.Endpoint = host + ":" + port.Port()
	return harness, nil
}

func objectStorePortWaitStrategy() *wait.HostPortStrategy {
	return wait.ForMappedPort(seaweedFSS3Port).WithStartupTimeout(objectStorePortMappingTimeout)
}

func attachOwnedContainerDiagnostic(err error, harness *OwnedStore) {
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) || harness == nil || harness.Container == nil {
		return
	}
	inspectCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	state, inspectErr := harness.Container.State(inspectCtx)
	if inspectErr != nil || state == nil {
		readinessErr.ContainerState = "unavailable"
		readinessErr.HealthState = "unavailable"
		return
	}
	readinessErr.ContainerState = string(state.Status)
	if readinessErr.ContainerState == "" {
		readinessErr.ContainerState = "unavailable"
	}
	readinessErr.HealthState = "none"
	if state.Health != nil {
		readinessErr.HealthState = string(state.Health.Status)
	}
	exitCode := state.ExitCode
	oomKilled := state.OOMKilled
	readinessErr.ExitCode = &exitCode
	readinessErr.OOMKilled = &oomKilled
}

// BootstrapProbeBucket creates a broker-owned namespace and proves the exact
// mutation path that attached child processes will use within unique prefixes.
// The broker retains ownership of the bucket for the suite lifetime.
func (o *OwnedStore) BootstrapProbeBucket(ctx context.Context) (string, error) {
	h := o.Fixture
	bucket, err := h.BootstrapBucket(ctx, "readiness-shared")
	if err != nil {
		return "", fmt.Errorf("create broker object-store probe namespace: %w", err)
	}
	if err := h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
		Phase:           "broker_namespace",
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

func (h *OwnedStore) ContainerID() string {
	if h == nil || h.Container == nil {
		return ""
	}
	return h.Container.GetContainerID()
}
