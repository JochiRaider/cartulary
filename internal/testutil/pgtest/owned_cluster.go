package pgtest

import (
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/jackc/pgx/v5/pgconn"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// OwnedCluster is the sole authority for a physical PostgreSQL lifetime.
type OwnedCluster struct {
	Container          testcontainers.Container
	Host               string
	Port               string
	User               string
	tlsDirectory       string
	closeOnce          sync.Once
	closeErr           error
	readinessProbes    atomic.Int64
	readinessConnected func(context.Context)
	readinessFailed    func(error)

	ready         atomic.Bool
	fixture       *Harness
	operations    startupOperations
	templateOps   *templateOperations
	preparationMu sync.Mutex
	preparation   *templateAttempt
}

type startupOperations struct {
	start      func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error)
	wait       func(context.Context, *OwnedCluster) error
	initialize func(context.Context, string) (string, error)
	ping       func(context.Context, string, func(context.Context)) error
	preflight  func(context.Context) (string, error)
	sleep      func(context.Context, time.Duration) error
	logs       func(context.Context, testcontainers.Container) ([]postgresLogChunk, bool, error)
}

func defaultStartupOperations() startupOperations {
	return startupOperations{start: testcontainersx.GenericContainerObserved, wait: func(ctx context.Context, owner *OwnedCluster) error { return owner.WaitReady(ctx) }, initialize: initializeCluster, ping: pingAdminDSN, logs: readPostgresStartupLogs}
}

// StartOwned acquires a physical lifetime. On unresolved startup settlement it
// returns the partial owner together with the initiating and cleanup failures.
func StartOwned(ctx context.Context, options StartOptions) (*OwnedCluster, error) {
	return startOwned(ctx, options, defaultStartupOperations())
}

// Fixture is available only after successful physical admission and until close.
func (h *OwnedCluster) Fixture() *Harness {
	if h == nil || !h.ready.Load() || h.fixture.retired.Load() {
		return nil
	}
	return h.fixture
}
func (h *OwnedCluster) AdminDSN() string {
	if h == nil || h.fixture == nil {
		return ""
	}
	return h.fixture.adminDSN
}
func (h *OwnedCluster) DSNTemplate() string {
	if h == nil || h.fixture == nil {
		return ""
	}
	return h.fixture.dsnTemplate
}
func (h *OwnedCluster) ProvisioningGeneration() string {
	if h == nil || h.fixture == nil {
		return ""
	}
	return h.fixture.generation
}

func startOwned(ctx context.Context, options StartOptions, ops startupOperations) (*OwnedCluster, error) {
	binding := bindFixture(processFixtureInputs())
	if err := suiteservices.CheckServiceDependencies(binding.environment(), "postgres"); err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, errors.New("postgres acquisition requires a context")
	}
	ctx, cancel := context.WithTimeout(ctx, postgresStartupTimeout)
	defer cancel()
	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"PGDATA":               "/var/lib/postgresql/18/docker",
			"POSTGRES_DB":          "postgres",
			"POSTGRES_USER":        "cartulary",
			"POSTGRES_PASSWORD":    "cartulary",
			"POSTGRES_INITDB_ARGS": postgresInitdbArgs,
		},
		WaitingFor: postgresPortWaitStrategy(),
	}
	if len(options.Labels) > 0 {
		req.Labels = cloneLabels(options.Labels)
	}

	harness, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service:        "postgres testcontainer",
		Image:          postgresImage,
		MaxAttempts:    3,
		AttemptTimeout: postgresStartupAttemptTimeout,
		RetryBackoff:   500 * time.Millisecond,
		Preflight:      ops.preflight,
		Retryable:      isRetryablePostgresStartupFailure,
		Sleep:          ops.sleep,
		Observer:       options.Observer,
	}, func(ctx context.Context) (*OwnedCluster, error) {
		return startOwnedAttempt(ctx, req, options, ops, binding)
	})
	return harness, err
}

func startOwnedAttempt(ctx context.Context, req testcontainers.ContainerRequest, options StartOptions, ops startupOperations, binding fixtureBinding) (partial *OwnedCluster, retErr error) {
	diagnostic := newStartupDiagnostic(ctx)
	diagnostic.logs = ops.logs
	stage := "container_start"
	req, tlsDirectory, err := preparePostgresTLS(req)
	if err != nil {
		return nil, err
	}
	harness := &OwnedCluster{User: "cartulary", tlsDirectory: tlsDirectory, operations: ops, fixture: &Harness{binding: binding, schemaHash: pgschema.MustHash(), suiteHash: resolveSuiteHash(binding.environment()), processHash: suiteservices.ProcessHash(), events: options.Events}}
	intentRecorded, providerCalled := false, false
	defer func() {
		if retErr == nil {
			diagnostic.advance(stage)
			diagnostic.Stage, diagnostic.Outcome, diagnostic.CleanupOutcome = stage, "ready", "not_needed"
			diagnostic.ElapsedMS = time.Since(diagnostic.start).Milliseconds()
			diagnostic.ProbeCount = int(harness.readinessProbes.Load())
			if options.Readiness != nil {
				options.Readiness(diagnostic.PostgresReadinessDiagnostic)
			}
			return
		}
		diagnostic.finish(stage, retErr)
		diagnostic.observeContainer(harness.Container)
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// No ready binding or borrower exists. Do not spend the removal budget
		// waiting for graceful shutdown of an unpublished failed allocation.
		cleanupErr := harness.closeOwned(cleanupCtx, testcontainers.StopTimeout(0))
		cancel()
		if cleanupErr == nil && (harness.Container != nil || (intentRecorded && !providerCalled)) && options.Settled != nil {
			cleanupErr = options.Settled(harness)
		}
		diagnostic.CleanupOutcome = "completed"
		partial = nil
		if cleanupErr != nil {
			diagnostic.CleanupOutcome = "unresolved"
			partial = harness
		}
		retErr = &postgresStartupFailure{primary: retErr, cleanup: cleanupErr, diagnostic: diagnostic}
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
	container, err := ops.start(createCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	}, diagnostic.publicationObserver(req.ExposedPorts))
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
		return nil, errors.New("postgres provider returned no container")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := container.Start(ctx); err != nil {
		return nil, err
	}
	diagnostic.advance(stage)
	stage = "endpoint_resolution"
	endpointCtx, cancelEndpoint := context.WithTimeout(ctx, postgresPortMappingTimeout)
	defer cancelEndpoint()
	host, err := container.Host(endpointCtx)
	if err != nil {
		return nil, fmt.Errorf("resolve postgres host: %w", err)
	}

	port, err := container.MappedPort(endpointCtx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("resolve postgres mapped port: %w", err)
	}
	diagnostic.endpoint.selectEndpoint(endpointCtx, container, host, port.Port())
	cancelEndpoint()

	harness.Host, harness.Port = host, port.Port()
	harness.fixture.adminDSN = certificateDSN(host, port.Port(), "postgres", harness.User, tlsDirectory)
	harness.fixture.dsnTemplate = strings.ReplaceAll(certificateDSN(host, port.Port(), "{database}", harness.User, tlsDirectory), "%7Bdatabase%7D", "{database}")
	diagnostic.advance(stage)
	stage = "sql_connectivity"
	harness.readinessConnected = func(probeCtx context.Context) {
		if stage == "sql_connectivity" {
			ms := time.Since(diagnostic.start).Milliseconds()
			diagnostic.FirstSQLConnectionMS = &ms
			diagnostic.advance(stage)
			stage = "security_capability"
			diagnostic.phaseContext = probeCtx
		}
	}
	harness.readinessFailed = diagnostic.recordProbeFailure
	defer func() { harness.readinessConnected = nil; harness.readinessFailed = nil }()
	if err := ops.wait(context.WithValue(ctx, endpointTraceKey{}, diagnostic.endpoint), harness); err != nil {
		return nil, &postgresGenerationFailure{fmt.Errorf("wait for postgres readiness: %w", err)}
	}
	diagnostic.advance(stage)
	stage = "generation_initialization"
	harness.fixture.generation, err = ops.initialize(ctx, harness.AdminDSN())
	if err != nil {
		return nil, &postgresGenerationFailure{fmt.Errorf("initialize postgres generation: %w", err)}
	}
	harness.ready.Store(true)
	return harness, nil
}

func postgresPortWaitStrategy() *wait.HostPortStrategy {
	return wait.ForMappedPort("5432/tcp").WithStartupTimeout(postgresPortMappingTimeout)
}

// WaitReady consumes the caller's bounded acquisition budget. Port publication
// does not mean bootstrap has finished, so it must not start another deadline.
func (h *OwnedCluster) WaitReady(ctx context.Context) error {
	if ctx == nil {
		return &postgresReadinessError{LastErr: errors.New("postgres readiness requires a bounded context")}
	}
	if err := ctx.Err(); err != nil {
		return &postgresReadinessError{LastErr: err, DeadlineExpired: errors.Is(err, context.DeadlineExceeded)}
	}
	if _, bounded := ctx.Deadline(); !bounded {
		return &postgresReadinessError{LastErr: errors.New("postgres readiness requires a bounded context")}
	}

	ticker := time.NewTicker(postgresHealthPollInterval)
	defer ticker.Stop()

	attempts := 0
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return &postgresReadinessError{Attempts: attempts, LastErr: errors.Join(lastErr, err), DeadlineExpired: errors.Is(err, context.DeadlineExceeded)}
		}
		attempts++
		h.readinessProbes.Store(int64(attempts))
		attemptCtx, attemptCancel := context.WithTimeout(ctx, postgresClientAttemptTimeout)
		err := h.operations.ping(attemptCtx, h.AdminDSN(), h.readinessConnected)
		attemptCancel()
		if err == nil {
			return nil
		}
		if h.readinessFailed != nil {
			h.readinessFailed(err)
		}
		if isNonRetryablePostgresReadinessError(err) {
			return &postgresReadinessError{Attempts: attempts, LastErr: err}
		}

		lastErr = err
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = ctx.Err()
			}
			return &postgresReadinessError{
				Attempts:        attempts,
				LastErr:         errors.Join(lastErr, ctx.Err()),
				DeadlineExpired: errors.Is(ctx.Err(), context.DeadlineExceeded),
			}
		case <-ticker.C:
		}
	}
}

func (h *OwnedCluster) Close(ctx context.Context) error {
	if h == nil {
		return nil
	}

	return h.closeOwned(ctx)
}

func (h *OwnedCluster) closeOwned(ctx context.Context, options ...testcontainers.TerminateOption) error {
	h.closeOnce.Do(func() {
		if h.fixture != nil {
			h.fixture.retired.Store(true)
		}
		if h.Container != nil {
			h.closeErr = h.Container.Terminate(ctx, options...)
		}
		if h.closeErr == nil && h.tlsDirectory != "" {
			h.closeErr = os.RemoveAll(h.tlsDirectory)
		}
	})
	return h.closeErr
}

func (h *OwnedCluster) ContainerID() string {
	if h == nil || h.Container == nil {
		return ""
	}
	return h.Container.GetContainerID()
}

type StartOptions struct {
	Labels        map[string]string
	Observer      testcontainersx.StartObserver
	BeforeAcquire func(*OwnedCluster) error
	Settled       func(*OwnedCluster) error
	// Acquired publishes private cleanup authority before starting the container.
	// It must not publish a ready binding.
	Acquired  func(*OwnedCluster) error
	Readiness func(suiteservices.PostgresReadinessDiagnostic)
	// Events publishes fixture lifecycle facts in the physical owner's suite.
	Events func(suiteservices.Event) error
}

type postgresReadinessError struct {
	Attempts        int
	LastErr         error
	DeadlineExpired bool
}

func (e *postgresReadinessError) Error() string {
	if e == nil {
		return ""
	}
	if e.LastErr == nil {
		return "postgres did not become ready"
	}
	return fmt.Sprintf("postgres did not become ready: %v", e.LastErr)
}

func (e *postgresReadinessError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.LastErr
}

type postgresGenerationFailure struct{ error }

func (e *postgresGenerationFailure) Unwrap() error { return e.error }

func isRetryablePostgresStartupFailure(err error) bool {
	var startup *postgresStartupFailure
	if errors.As(err, &startup) {
		if startup.RetryForbidden() {
			return false
		}
		err = startup.primary
	}
	var generation *postgresGenerationFailure
	if errors.As(err, &generation) {
		return false
	}
	var readinessErr *postgresReadinessError
	if errors.As(err, &readinessErr) {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "context deadline exceeded") &&
		(strings.Contains(lower, "5432") ||
			strings.Contains(lower, "mapped port") ||
			strings.Contains(lower, "listening port") ||
			strings.Contains(lower, "port wait") ||
			strings.Contains(lower, "wait until ready"))
}

func isNonRetryablePostgresReadinessError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "28P01", "28000", "3D000", "42501":
			return true
		}
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "password authentication failed") ||
		strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "role") && strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "database") && strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "no pg_hba.conf entry") ||
		strings.Contains(lower, "invalid connection") ||
		strings.Contains(lower, postgres.ReasonServerVersionMismatch) ||
		strings.Contains(lower, postgres.ReasonDataChecksumsDisabled) ||
		strings.Contains(lower, postgres.ReasonAdmissionFailed)
}
