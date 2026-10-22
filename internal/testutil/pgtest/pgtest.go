package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

const postgresImage = "docker.io/library/postgres:18.6-alpine@sha256:d3e1620b530c944afa6e887d22eb899824da68e19c52024bf98f5220c88a65b2"

const (
	postgresPortMappingTimeout      = 15 * time.Second
	postgresClientReadyTimeout      = 15 * time.Second
	postgresHealthPollInterval      = 250 * time.Millisecond
	postgresClientAttemptTimeout    = 5 * time.Second
	postgresCatalogAdmissionTimeout = 15 * time.Second

	postgresFixturePolicyTemplateClone = "template_clone"
	postgresFixturePolicyTransaction   = "transaction"
)

const (
	postgresFixturePolicyTestsEnv    = "CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS"
	postgresFixturePolicyPackagesEnv = "CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES"
	postgresFixturePolicyDefaultEnv  = "CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT"
)

const (
	// PostgresFixturePolicyTemplateClone is the scheduler policy for tests that
	// need committed per-test database clones instead of rollback transactions.
	PostgresFixturePolicyTemplateClone = postgresFixturePolicyTemplateClone
)

func ContainerImage() string {
	return postgresImage
}

type Harness struct {
	Container          testcontainers.Container
	Host               string
	Port               string
	User               string
	tlsDirectory       string
	generation         string
	closeOnce          sync.Once
	closeErr           error
	readinessProbes    atomic.Int64
	readinessConnected func(context.Context)
	readinessFailed    func(error)

	adminDSN    string
	dsnTemplate string
	templateDB  string
	schemaHash  string
	suiteHash   string
	processHash string
	counter     uint64
	shared      bool
	attached    bool

	templateMu      sync.Mutex
	ownedDatabaseMu sync.Mutex
	ownedDatabases  map[string]struct{}
	quarantined     atomic.Bool

	transactionDBMu sync.Mutex
	transactionDBs  map[string]*transactionDatabase
}

type TestDatabase struct {
	Name string
	DSN  string
}

// MigrationDatabase is a harness-issued capability for constructing migration
// history in a disposable scratch database. Its unexported identity prevents
// callers from wrapping an arbitrary database handle.
type MigrationDatabase struct {
	db       *sql.DB
	identity *migrationDatabaseIdentity
}

type migrationDatabaseIdentity struct{}

var issuedMigrationDatabaseIdentity = &migrationDatabaseIdentity{}

var errGenerationQuarantined = errors.New("postgres cluster generation is quarantined after fixture settlement failure")

type StartOptions struct {
	Labels         map[string]string
	Observer       testcontainersx.StartObserver
	AttemptTimeout time.Duration
	BeforeAcquire  func(*Harness) error
	Settled        func(*Harness) error
	// Acquired publishes private cleanup authority before starting the container.
	// It must not publish a ready binding.
	Acquired  func(*Harness) error
	Readiness func(suiteservices.PostgresReadinessDiagnostic)
}

type RollbackDB struct {
	tx   pgx.Tx
	conn *pgx.Conn
}

type transactionDatabase struct {
	mu       sync.Mutex
	db       *TestDatabase
	unusable error
}

func (f *transactionDatabase) lockForBorrow() error {
	f.mu.Lock()
	if f.unusable != nil {
		err := f.unusable
		f.mu.Unlock()
		return err
	}
	return nil
}

// Called with the fixture mutex held. Even after cancellation, each client
// operation gets its own bounded settlement budget. Uncertain state is never
// handed to another borrower; ownership stays in the database cleanup ledger.
func (f *transactionDatabase) settle(ctx context.Context, rollback, closeConnection func(context.Context) error) error {
	base := context.WithoutCancel(ctx)
	var failures []error
	for _, operation := range []struct {
		name string
		run  func(context.Context) error
	}{{"rollback", rollback}, {"connection close", closeConnection}} {
		if operation.run == nil {
			continue
		}
		attempt, cancel := context.WithTimeout(base, 5*time.Second)
		err := operation.run(attempt)
		cancel()
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", operation.name, err))
		}
	}
	if err := errors.Join(failures...); err != nil {
		f.unusable = err
	}
	return f.unusable
}

type fixtureAttribution struct {
	TestName              string
	CallerFile            string
	CallerPackage         string
	HarnessPackage        string
	PostgresFixturePolicy string
	ReuseGroup            string
}

var (
	sharedHarnessMu            sync.Mutex
	sharedHarness              *Harness
	schemaHashOnce             sync.Once
	schemaHashValue            string
	schemaHashErr              error
	initializeClusterFn        = initializeCluster
	inspectClusterGenerationFn = inspectClusterGeneration
	finalizeCloneFn            = finalizeClone
	startContainerFn           = testcontainersx.GenericContainer
	waitReadyFn                = func(ctx context.Context, harness *Harness) error {
		return harness.WaitReady(ctx)
	}
	startPreflightFn          func(context.Context) (string, error)
	startSleepFn              func(context.Context, time.Duration) error
	pingAdminDSNFn            = pingAdminDSN
	migrateDatabaseFn         = database_migrations.Apply
	initializeDatabaseFn      = InitializeFreshDatabase
	applyMigrationsThrough    = applyCanonicalMigrationsThrough
	rollbackMigrationsThrough = rollbackCanonicalMigrationsThrough
	createDatabaseFn          = createDatabase
	dropOwnedDatabaseFn       = postgrescleanup.DropOwnedDatabaseWithObservation
	sealTemplateDBFn          = SealTemplateDatabase
	openMigrationDatabaseFn   = OpenPurposeDatabase
)

func Start(t testing.TB) *Harness {
	t.Helper()
	suiteservices.RequireServiceDependencies(t, "pgtest", "postgres")

	harness, err := StartShared(context.Background())
	if err != nil {
		failPostgresSetup(t, "start", err)
	}

	return harness
}

func StartShared(ctx context.Context) (*Harness, error) {
	if err := suiteservices.CheckServiceDependencies(nil, "postgres"); err != nil {
		return nil, err
	}
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness != nil {
		if err := sharedHarness.admitPreparation(); err != nil {
			return nil, err
		}
		return sharedHarness, nil
	}

	harness, attached, err := startAttachedHarness(ctx)
	if err != nil {
		return nil, err
	}
	if !attached {
		return nil, errors.New("postgres requires an explicit suite attachment; dedicated lifecycle fixtures must use StartOwned")
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
	if err := suiteservices.CheckServiceDependencies(nil, "postgres"); err != nil {
		return nil, err
	}
	return startHarnessWithOptions(ctx, options)
}

func StopShared(ctx context.Context) error {
	sharedHarnessMu.Lock()
	defer sharedHarnessMu.Unlock()

	if sharedHarness == nil {
		return nil
	}
	if sharedHarness.Container == nil {
		// Detaching does not settle an attached physical generation. Retain
		// its quarantine so a later StartShared cannot readmit the same binding.
		if !sharedHarness.quarantined.Load() {
			sharedHarness = nil
		}
		return nil
	}

	if err := sharedHarness.closeOwned(ctx); err != nil {
		return err
	}
	sharedHarness = nil
	return nil
}

func startHarnessWithOptions(ctx context.Context, options StartOptions) (*Harness, error) {
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
		AttemptTimeout: options.AttemptTimeout,
		RetryBackoff:   500 * time.Millisecond,
		Preflight:      startPreflightFn,
		Retryable:      isRetryablePostgresStartupFailure,
		Sleep:          startSleepFn,
		Observer:       options.Observer,
	}, func(ctx context.Context) (*Harness, error) {
		return startHarnessAttempt(ctx, req, options)
	})
	return harness, err
}

func startHarnessAttempt(ctx context.Context, req testcontainers.ContainerRequest, options StartOptions) (partial *Harness, retErr error) {
	diagnostic := newStartupDiagnostic(ctx)
	stage := "container_start"
	req, tlsDirectory, err := preparePostgresTLS(req)
	if err != nil {
		return nil, err
	}
	harness := &Harness{User: "cartulary", tlsDirectory: tlsDirectory, schemaHash: pgschema.MustHash(), suiteHash: resolveSuiteHash(), processHash: suiteservices.ProcessHash()}
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
		cleanupErr := harness.closeOwned(cleanupCtx)
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
	container, err := startContainerFn(createCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})
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
	harness.adminDSN = harness.dsnFor("postgres")
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
	if err := waitReadyFn(context.WithValue(ctx, endpointTraceKey{}, diagnostic.endpoint), harness); err != nil {
		return nil, &postgresGenerationFailure{fmt.Errorf("wait for postgres readiness: %w", err)}
	}
	diagnostic.advance(stage)
	stage = "generation_initialization"
	harness.generation, err = initializeClusterFn(ctx, harness.adminDSN)
	if err != nil {
		return nil, &postgresGenerationFailure{fmt.Errorf("initialize postgres generation: %w", err)}
	}
	return harness, nil
}

func postgresPortWaitStrategy() *wait.HostPortStrategy {
	return wait.ForMappedPort("5432/tcp").WithStartupTimeout(postgresPortMappingTimeout)
}

func startAttachedHarness(ctx context.Context) (*Harness, bool, error) {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGAdminDSNEnv))
	dsnTemplate := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGDSNTemplateEnv))
	if adminDSN == "" && dsnTemplate == "" {
		return nil, false, nil
	}
	if adminDSN == "" || dsnTemplate == "" {
		return nil, false, fmt.Errorf("attach postgres harness: %s and %s must both be set", suiteservices.PGAdminDSNEnv, suiteservices.PGDSNTemplateEnv)
	}
	if !strings.Contains(dsnTemplate, "{database}") {
		return nil, false, fmt.Errorf("attach postgres harness: %s must contain {database}", suiteservices.PGDSNTemplateEnv)
	}

	harness := &Harness{
		adminDSN:    adminDSN,
		dsnTemplate: dsnTemplate,
		templateDB:  strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGTemplateDBEnv)),
		schemaHash:  pgschema.MustHash(),
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
		attached:    true,
	}
	if advertisedHash := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGSchemaHashEnv)); advertisedHash != harness.schemaHash {
		return nil, false, fmt.Errorf("attach postgres harness: %s mismatch: got %s want %s", suiteservices.PGSchemaHashEnv, advertisedHash, harness.schemaHash)
	}
	if suiteservices.LookupEnvValue(nil, suiteservices.PGProvisioningDigestEnv) != ProvisioningDigest() || suiteservices.LookupEnvValue(nil, suiteservices.PGTemplateDigestEnv) != TemplateDigest(harness.schemaHash) || harness.templateDB == "" {
		return nil, false, errors.New("attach postgres harness: incompatible provisioning or template identity")
	}
	if !slices.Contains([]string{"shared", "exclusive"}, suiteservices.LookupEnvValue(nil, suiteservices.PGIsolationEnv)) || suiteservices.LookupEnvValue(nil, suiteservices.PGAllocationRefEnv) == "" {
		return nil, false, errors.New("attach postgres harness: isolation ownership is missing")
	}
	generation := suiteservices.LookupEnvValue(nil, suiteservices.PGGenerationEnv)
	actual, err := inspectClusterGenerationFn(ctx, harness.adminDSN)
	if err != nil || generation == "" || actual != generation {
		return nil, false, errors.New("attach postgres harness: physical generation mismatch")
	}
	harness.generation = generation

	if err := pingAdminDSNFn(ctx, adminDSN, nil); err != nil {
		return nil, false, fmt.Errorf("attach postgres harness: ping admin dsn: %w", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type: suiteservices.EventPostgresAttach,
		Details: map[string]any{
			"schema_hash": harness.schemaHash,
		},
	})
	return harness, true, nil
}

func (h *Harness) WaitReady(ctx context.Context) error {
	readyCtx, cancel := context.WithTimeout(ctx, postgresClientReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(postgresHealthPollInterval)
	defer ticker.Stop()

	attempts := 0
	var lastErr error
	for {
		if err := readyCtx.Err(); err != nil {
			return &postgresReadinessError{Attempts: attempts, LastErr: errors.Join(lastErr, err), DeadlineExpired: errors.Is(err, context.DeadlineExceeded)}
		}
		attempts++
		h.readinessProbes.Store(int64(attempts))
		attemptCtx, attemptCancel := context.WithTimeout(readyCtx, postgresClientAttemptTimeout)
		err := pingAdminDSNFn(attemptCtx, h.adminDSN, h.readinessConnected)
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
		case <-readyCtx.Done():
			if lastErr == nil {
				lastErr = readyCtx.Err()
			}
			return &postgresReadinessError{
				Attempts:        attempts,
				LastErr:         errors.Join(lastErr, readyCtx.Err()),
				DeadlineExpired: errors.Is(readyCtx.Err(), context.DeadlineExceeded),
			}
		case <-ticker.C:
		}
	}
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

func (h *Harness) AdminDSN() string {
	return h.adminDSN
}

func (h *Harness) NewDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()
	attribution := fixtureAttributionFor(t, "pgtest")
	testDB, err := h.newDatabase(context.Background(), prefix, suiteservices.FixtureReusePerTest, attribution)
	if err != nil {
		failDatabasePreparation(t, "create postgres database", err)
	}
	h.registerDatabaseCleanupT(t, testDB.Name, suiteservices.FixtureReusePerTest, attribution)
	return testDB
}

func (h *Harness) newDatabase(ctx context.Context, prefix string, reuseScope string, attribution fixtureAttribution) (*TestDatabase, error) {
	name := h.nextDatabaseName(prefix)
	start := time.Now()
	if err := h.createDatabase(ctx, name, ""); err != nil {
		return nil, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBCreated,
		Name:    name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", reuseScope, attribution, time.Since(start)),
	})

	return &TestDatabase{
		Name: name,
		DSN:  h.dsnFor(name),
	}, nil
}

func (h *Harness) PrepareDatabase(ctx context.Context, prefix string) (*TestDatabase, error) {
	return h.prepareDatabase(ctx, prefix, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) prepareDatabase(ctx context.Context, prefix string, reuseScope string, attribution fixtureAttribution) (*TestDatabase, error) {
	if err := h.admitPreparation(); err != nil {
		return nil, err
	}
	if h.templateDB == "" && !h.attached {
		if err := h.ensureLocalTemplateDatabase(ctx); err != nil {
			return nil, err
		}
	}
	if h.templateDB != "" {
		name := h.nextDatabaseName(prefix)
		start := time.Now()
		if err := h.createDatabase(ctx, name, h.templateDB); err != nil {
			return nil, err
		}
		recordSuiteEvent(suiteservices.Event{
			Type:    suiteservices.EventPostgresDBCreated,
			Name:    name,
			Kind:    "template-clone",
			Details: postgresPreparationDetails(suiteservices.PostgresPreparationTemplateClone, h.templateDB, reuseScope, attribution, time.Since(start)),
		})
		recordSuiteEvent(suiteservices.Event{
			Type: suiteservices.EventPostgresTemplateUse,
			Name: name,
			Details: map[string]any{
				"template_database": h.templateDB,
				"target":            suiteservices.LookupEnvValue(nil, suiteservices.TargetEnv),
			},
		})

		return &TestDatabase{
			Name: name,
			DSN:  h.dsnFor(name),
		}, nil
	}

	testDB, err := h.newDatabase(ctx, prefix, reuseScope, attribution)
	if err != nil {
		return nil, err
	}

	db, err := OpenPurposeDatabase(testDB.DSN, postgres.PurposeMigration)
	if err != nil {
		return nil, fmt.Errorf("open test database handle: %w", err)
	}
	defer db.Close()

	source, err := dbmigrations.Source()
	if err != nil {
		return nil, fmt.Errorf("load migration source: %w", err)
	}
	migrationDuration, err := suiteservices.MeasureMigration(ctx, func() error { return initializeDatabaseFn(ctx, db, testDB.DSN, source) })
	if err != nil {
		return nil, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", reuseScope, attribution, migrationDuration),
	})

	return testDB, nil
}

func (h *Harness) ensureLocalTemplateDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, TemplatePreparationTimeout)
	defer cancel()
	h.templateMu.Lock()
	defer h.templateMu.Unlock()

	if h.templateDB != "" {
		return nil
	}
	if h.schemaHash == "" {
		hash, err := pgschema.Hash()
		if err != nil {
			return err
		}
		h.schemaHash = hash
	}

	name := templateDatabaseName(h.suiteHash, h.schemaHash)
	if err := h.createDatabase(ctx, name, ""); err != nil {
		return fmt.Errorf("create local postgres template database: %w", err)
	}
	templateDSN := h.dsnFor(name)
	db, err := OpenPurposeDatabase(templateDSN, postgres.PurposeMigration)
	if err != nil {
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("open local postgres template database: %w", err)
	}
	source, err := dbmigrations.Source()
	if err != nil {
		_ = db.Close()
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("load migration source: %w", err)
	}
	migrationDuration, err := suiteservices.MeasureMigration(ctx, func() error { return initializeDatabaseFn(ctx, db, templateDSN, source) })
	if err != nil {
		var appliedVersion int64
		_ = db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_id), 0) FROM public.goose_db_version WHERE is_applied`).Scan(&appliedVersion)
		_ = db.Close()
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("migrate local postgres template database after version %d: %w", appliedVersion, err)
	}
	if err := db.Close(); err != nil {
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("close local postgres template database: %w", err)
	}
	if err := sealTemplateDBFn(ctx, h.adminDSN, name); err != nil {
		return errors.Join(err, h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{}))
	}

	attribution := fixtureAttribution{PostgresFixturePolicy: "suite_template"}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBCreated,
		Name:    name,
		Kind:    "template",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationTemplate, name, suiteservices.FixtureReuseSuiteTemplate, attribution, 0),
	})
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    name,
		Kind:    "template",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationTemplate, name, suiteservices.FixtureReuseSuiteTemplate, attribution, migrationDuration),
	})
	h.templateDB = name
	return nil
}

func (h *Harness) PrepareIsolatedDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyTemplateClone)
	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyTemplateClone
	testDB, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePerTest, attribution)
	if err != nil {
		failDatabasePreparation(t, "prepare postgres database", err)
	}
	h.registerDatabaseCleanupT(t, testDB.Name, suiteservices.FixtureReusePerTest, attribution)
	return testDB
}

// OpenIsolatedDatabaseT opens a fresh current-head database through the
// production purpose-role initializer. Owner contract tests use this instead
// of replaying historical migration boundaries.
func (h *Harness) OpenIsolatedDatabaseT(t testing.TB, prefix string, purpose postgres.Purpose) *sql.DB {
	t.Helper()
	testDB := h.PrepareIsolatedDatabaseT(t, prefix)
	db, err := OpenPurposeDatabase(testDB.DSN, purpose)
	if err != nil {
		t.Fatalf("open isolated postgres database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func (h *Harness) newMigrationDatabaseT(t testing.TB) *TestDatabase {
	t.Helper()
	attribution := fixtureAttributionFor(t, "pgtest")
	testDB, err := h.newDatabase(context.Background(), "migration_scratch", suiteservices.FixtureReuseMigrationScratch, attribution)
	if err != nil {
		failDatabasePreparation(t, "create migration scratch database", err)
	}
	h.registerDatabaseCleanupT(t, testDB.Name, suiteservices.FixtureReuseMigrationScratch, attribution)
	return testDB
}

func (h *Harness) registerDatabaseCleanupT(t testing.TB, name string, reuseScope string, attribution fixtureAttribution) {
	t.Helper()
	t.Cleanup(func() {
		if err := h.dropDatabase(context.Background(), name, reuseScope, attribution); err != nil {
			testfailure.CleanupFailure(t, "pgtest", "postgres", "delete", err)
		}
	})
}

func (h *Harness) MigrationDatabaseT(t testing.TB) *MigrationDatabase {
	return h.migrationDatabaseT(t, func(ctx context.Context, db *MigrationDatabase) error {
		source, err := dbmigrations.Source()
		if err != nil {
			return err
		}
		return migrateDatabaseFn(ctx, db.SQL(), source)
	})
}

func (h *Harness) MigrationDatabaseThroughT(t testing.TB, version int64) *MigrationDatabase {
	t.Helper()
	if err := validateApplyThroughVersion(version); err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}
	return h.migrationDatabaseT(t, func(ctx context.Context, db *MigrationDatabase) error {
		return db.ApplyThrough(ctx, version)
	})
}

func (h *Harness) migrationDatabaseT(t testing.TB, apply func(context.Context, *MigrationDatabase) error) *MigrationDatabase {
	t.Helper()

	// MigrationDatabaseT always creates a fresh scratch database and replays the
	// requested migration path. Use PrepareIsolatedDatabaseT or an explicit
	// group/reset/transaction helper
	// for current-head schema assertions; keep MigrationDatabaseT for tests that
	// prove migration runner behavior, boundary upgrades, or backfills.
	testDB := h.newMigrationDatabaseT(t)

	var db *sql.DB
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
		}
	})

	openedDB, err := openMigrationDatabaseFn(testDB.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatalf("open migration scratch database: %v", err)
	}
	db = openedDB
	migrationDB := &MigrationDatabase{db: db, identity: issuedMigrationDatabaseIdentity}

	migrationDuration, err := suiteservices.MeasureMigration(context.Background(), func() error { return apply(context.Background(), migrationDB) })
	if err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", suiteservices.FixtureReuseMigrationScratch, fixtureAttributionFor(t, "pgtest"), migrationDuration),
	})

	return migrationDB
}

// SQL returns the database handle owned by this disposable migration-scratch
// capability. The harness closes it during test cleanup.
func (db *MigrationDatabase) SQL() *sql.DB {
	if !db.valid() {
		return nil
	}
	return db.db
}

// ApplyThrough advances this scratch database through a positive target.
func (db *MigrationDatabase) ApplyThrough(ctx context.Context, version int64) error {
	if err := validateApplyThroughVersion(version); err != nil {
		return err
	}
	if !db.valid() {
		return errors.New("migration database capability is not harness-issued")
	}
	return applyMigrationsThrough(ctx, db.db, version)
}

// RollbackThrough rolls this scratch database back through a non-negative target.
func (db *MigrationDatabase) RollbackThrough(ctx context.Context, version int64) error {
	if err := validateRollbackThroughVersion(version); err != nil {
		return err
	}
	if !db.valid() {
		return errors.New("migration database capability is not harness-issued")
	}
	return rollbackMigrationsThrough(ctx, db.db, version)
}

func (db *MigrationDatabase) valid() bool {
	return db != nil && db.identity == issuedMigrationDatabaseIdentity && db.db != nil
}

func validateApplyThroughVersion(version int64) error {
	if version <= 0 {
		return errors.New("migration apply-through version must be positive")
	}
	return nil
}

func validateRollbackThroughVersion(version int64) error {
	if version < 0 {
		return errors.New("migration rollback-through version must be non-negative")
	}
	return nil
}

func applyCanonicalMigrationsThrough(ctx context.Context, db *sql.DB, version int64) error {
	provider, err := newCanonicalMigrationProvider(db)
	if err != nil {
		return err
	}
	_, err = provider.UpTo(ctx, version)
	return err
}

func rollbackCanonicalMigrationsThrough(ctx context.Context, db *sql.DB, version int64) error {
	provider, err := newCanonicalMigrationProvider(db)
	if err != nil {
		return err
	}
	_, err = provider.DownTo(ctx, version)
	return err
}

func newCanonicalMigrationProvider(db *sql.DB) (*goose.Provider, error) {
	source, err := dbmigrations.Source()
	if err != nil {
		return nil, fmt.Errorf("load canonical migration source: %w", err)
	}
	locker, err := sourcecatalog.NewSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("construct canonical migration locker: %w", err)
	}
	return sourcecatalog.NewProvider(db, source, locker)
}

func (h *Harness) BeginRollbackDBT(t testing.TB, prefix string) *RollbackDB {
	t.Helper()
	requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyTransaction)

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyTransaction
	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeIdentifier(prefix)
	}
	attribution.ReuseGroup = key
	fixture := h.transactionDatabase(key)
	if err := fixture.lockForBorrow(); err != nil {
		t.Fatalf("transaction postgres fixture is unusable after cleanup failure: %v", err)
	}
	if err := h.admitPreparation(); err != nil {
		fixture.mu.Unlock()
		failPostgresSetup(t, "capability", err)
	}

	if fixture.db == nil {
		testDB, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReuseTransaction, attribution)
		if err != nil {
			fixture.mu.Unlock()
			failDatabasePreparation(t, "prepare transaction postgres database", err)
		}
		fixture.db = testDB
	}

	conn, err := pgx.Connect(context.Background(), fixture.db.DSN)
	if err != nil {
		fixture.mu.Unlock()
		t.Fatalf("open transaction postgres connection: %v", err)
	}

	start := time.Now()
	tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		cleanupErr := fixture.settle(t.Context(), nil, conn.Close)
		fixture.mu.Unlock()
		if cleanupErr != nil {
			testfailure.CleanupFailure(t, "pgtest", "postgres", "capability", cleanupErr)
		}
		t.Fatalf("begin postgres rollback transaction: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresTransaction,
		Name:    fixture.db.Name,
		Details: postgresFixtureDetails("", suiteservices.FixtureReuseTransaction, attribution, time.Since(start)),
	})

	t.Cleanup(func() {
		defer fixture.mu.Unlock()
		if err := fixture.settle(t.Context(), tx.Rollback, conn.Close); err != nil {
			testfailure.CleanupFailure(t, "pgtest", "postgres", "capability", err)
		}
	})
	return &RollbackDB{tx: tx, conn: conn}
}

func (h *Harness) transactionDatabase(key string) *transactionDatabase {
	h.transactionDBMu.Lock()
	defer h.transactionDBMu.Unlock()

	if h.transactionDBs == nil {
		h.transactionDBs = make(map[string]*transactionDatabase)
	}
	fixture := h.transactionDBs[key]
	if fixture == nil {
		fixture = &transactionDatabase{}
		h.transactionDBs[key] = fixture
	}
	return fixture
}

func (db *RollbackDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.tx.Exec(ctx, sql, args...)
}

func (db *RollbackDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.tx.Query(ctx, sql, args...)
}

func (db *RollbackDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.tx.QueryRow(ctx, sql, args...)
}

func (db *RollbackDB) BeginTx(ctx context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	return db.tx.Begin(ctx)
}

func (h *Harness) DropDatabase(ctx context.Context, name string) error {
	return h.dropDatabase(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) dropDatabase(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	if !h.ownsDatabase(name) {
		return fmt.Errorf("refuse to drop postgres database %q without exact harness ownership", name)
	}
	start := time.Now()
	if err := dropOwnedDatabaseFn(ctx, h.adminDSN, name, func(attempt postgrescleanup.DropAttempt) {
		if h.generation != "" {
			generation := h.generation
			attempt.ServiceGeneration = &generation
		}
		if attempt.Validate() == nil {
			recordSuiteEvent(suiteservices.Event{Type: suiteservices.EventPostgresDropAttempt, Details: map[string]any{"attempt": attempt}})
		}
	}); err != nil {
		h.quarantined.Store(true)
		return err
	}
	details := postgresFixtureDetails("", reuseScope, attribution, time.Since(start))
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBDropped,
		Name:    name,
		Details: details,
	})

	return nil
}

func (h *Harness) ownsDatabase(name string) bool {
	h.ownedDatabaseMu.Lock()
	defer h.ownedDatabaseMu.Unlock()
	_, ok := h.ownedDatabases[name]
	return ok
}

func (h *Harness) Close(ctx context.Context) error {
	if h == nil {
		return nil
	}
	// Shared harnesses stay alive until StopShared or test-process teardown.
	if h.shared {
		return nil
	}

	return h.closeOwned(ctx)
}

func (h *Harness) closeOwned(ctx context.Context) error {
	h.closeOnce.Do(func() {
		if h.Container != nil {
			h.closeErr = h.Container.Terminate(ctx)
		}
		if h.closeErr == nil && h.tlsDirectory != "" {
			h.closeErr = os.RemoveAll(h.tlsDirectory)
		}
	})
	return h.closeErr
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

func (db *TestDatabase) Env() map[string]string {
	return map[string]string{
		suiteservices.PostgresDSNEnv: db.DSN,
	}
}

func (db *TestDatabase) EnvForServiceRef(serviceRef string, purpose postgres.Purpose) map[string]string {
	key, err := postgres.EnvKeyForServiceRef(serviceRef, purpose)
	if err != nil {
		return map[string]string{}
	}
	purposeDSN, err := db.DSNForPurpose(purpose)
	if err != nil {
		return map[string]string{}
	}
	return map[string]string{key: purposeDSN}
}

// DSNForPurpose returns a test-only credential for the exact deployment login
// associated with purpose. The physical PostgreSQL initializer owns these disposable-harness
// login credentials and their fixed, non-inheriting role memberships.
func (db *TestDatabase) DSNForPurpose(purpose postgres.Purpose) (string, error) {
	if db == nil {
		return "", errors.New("test database is nil")
	}
	return purposeDSN(db.DSN, purpose)
}

func (h *Harness) dsnFor(database string) string {
	if h.dsnTemplate != "" {
		return strings.ReplaceAll(h.dsnTemplate, "{database}", database)
	}
	return certificateDSN(h.Host, h.Port, database, h.User, h.tlsDirectory)
}

func sanitizeIdentifier(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")

	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_':
			builder.WriteRune(r)
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "testdb"
	}
	return result
}

func resolveSuiteHash() string {
	if hash := suiteservices.SuiteHash(nil); hash != "" {
		return hash
	}
	return suiteservices.ShortHash("local-suite", 8)
}

func (h *Harness) nextDatabaseName(prefix string) string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	processHash := h.processHash
	if processHash == "" {
		processHash = suiteservices.ProcessHash()
	}

	base := fmt.Sprintf("ct_%s_%s_%06d", suiteHash, processHash, atomic.AddUint64(&h.counter, 1))
	suffix := sanitizeIdentifier(prefix)
	if suffix == "" {
		return truncateIdentifier(base, 63)
	}

	available := 63 - len(base) - 1
	if available <= 0 {
		return truncateIdentifier(base, 63)
	}
	return truncateIdentifier(base+"_"+truncateIdentifier(suffix, available), 63)
}

func truncateIdentifier(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return strings.TrimRight(value[:max], "_")
}

func (h *Harness) createDatabase(ctx context.Context, name string, templateDB string) error {
	if err := suiteservices.CheckServiceDependencies(nil, "postgres"); err != nil {
		return err
	}
	if err := h.admitPreparation(); err != nil {
		return err
	}
	if err := createDatabaseFn(ctx, h.adminDSN, name, templateDB); err != nil {
		// Provisioning can fail after CREATE commits or while disposing that
		// partial resource. Do not infer a reusable generation from an error.
		h.quarantined.Store(true)
		return err
	}
	h.ownedDatabaseMu.Lock()
	if h.ownedDatabases == nil {
		h.ownedDatabases = make(map[string]struct{})
	}
	h.ownedDatabases[name] = struct{}{}
	h.ownedDatabaseMu.Unlock()
	return nil
}

func (h *Harness) admitPreparation() error {
	if h.quarantined.Load() {
		return errGenerationQuarantined
	}
	return nil
}

func failDatabasePreparation(t testing.TB, operation string, err error) {
	t.Helper()
	if errors.Is(err, errGenerationQuarantined) {
		failPostgresSetup(t, "capability", err)
	}
	t.Fatalf("%s: %v", operation, err)
}

func failPostgresSetup(t testing.TB, defaultStage string, err error) {
	t.Helper()
	testfailure.Fail(t, postgresSetupEnvelope(defaultStage, err))
}

func postgresSetupEnvelope(defaultStage string, err error) testfailure.Envelope {
	failureClass := "harness"
	failureReason := "fixture_error"
	stage := defaultStage
	attempts := 0
	if errors.Is(err, errGenerationQuarantined) {
		stage = "capability"
	}

	var readinessErr *postgresReadinessError
	if errors.As(err, &readinessErr) {
		attempts = readinessErr.Attempts
		if readinessErr.DeadlineExpired {
			failureClass = "infra"
			failureReason = "service_readiness_timeout"
		} else if isNonRetryablePostgresReadinessError(readinessErr.LastErr) {
			stage = "capability"
		}
	}
	var startErr *testcontainersx.StartFailure
	if errors.As(err, &startErr) && attempts == 0 {
		attempts = startErr.AttemptsStarted
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
		"pgtest",
		"postgres",
		stage,
		attempts,
		"not_required",
	)
}

func pingAdminDSN(ctx context.Context, dsn string, connected func(context.Context)) error {
	var db *sql.DB
	var err error
	if trace, ok := ctx.Value(endpointTraceKey{}).(*endpointTrace); ok {
		config, parseErr := pgx.ParseConfig(dsn)
		if parseErr != nil {
			return parseErr
		}
		trace.observeConfig(config)
		db = stdlib.OpenDB(*config)
	} else {
		db, err = sql.Open("pgx", dsn)
	}
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	if connected != nil {
		connected(ctx)
	}
	var version int
	var checksums string
	if err := db.QueryRowContext(ctx, `SELECT current_setting('server_version_num')::integer, current_setting('data_checksums')::text`).Scan(&version, &checksums); err != nil {
		return errors.New(postgres.ReasonAdmissionFailed)
	}
	if version != postgres.RequiredServerVersionNum {
		return errors.New(postgres.ReasonServerVersionMismatch)
	}
	if checksums != "on" {
		return errors.New(postgres.ReasonDataChecksumsDisabled)
	}
	return nil
}

func createDatabase(ctx context.Context, adminDSN string, name string, templateDB string) error {
	var err error
	created := false
	if templateDB != "" {
		created, err = CloneDatabase(ctx, adminDSN, name, templateDB)
	} else {
		err = postgrescatalog.WithMutation(
			ctx,
			adminDSN,
			name,
			postgresCatalogAdmissionTimeout,
			func(admin *sql.Conn) error {
				statement := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
				if _, err := admin.ExecContext(ctx, statement); err != nil {
					return fmt.Errorf("create test database %s: %w", name, err)
				}
				created = true
				return nil
			},
		)
	}
	if err == nil && templateDB == "" {
		err = PrepareFreshDatabase(ctx, adminDSN, name)
	}
	if err != nil && created {
		// Setup cancellation must not suppress disposal of our committed CREATE.
		// Each cleanup stage keeps the common owner's separate existing bound.
		cleanupErr := postgrescleanup.DropOwnedDatabase(context.Background(), adminDSN, name)
		return errors.Join(err, cleanupErr)
	}
	return err
}

// OpenPurposeDatabase opens a test-only database handle through the same
// purpose-role initializer used by production composition.
func OpenPurposeDatabase(dsn string, purpose postgres.Purpose) (*sql.DB, error) {
	contract, ok := testPurposeContracts[purpose]
	if !ok {
		return nil, errors.New("unsupported test postgres purpose")
	}
	selectedDSN, err := purposeDSN(dsn, purpose)
	if err != nil {
		return nil, err
	}
	return postgres.OpenSQL(context.Background(), postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          selectedDSN,
		Purpose:      purpose,
		ExpectedRole: contract.role,
	})
}

type testPurposeContract struct {
	login string
	role  string
}

var testPurposeContracts = map[postgres.Purpose]testPurposeContract{
	postgres.PurposeRuntime: {
		login: "cartulary_runtime_login",
		role:  "cartulary_runtime",
	},
	postgres.PurposeMigration: {
		login: "cartulary_migration_login",
		role:  "cartulary_schema_owner",
	},
	postgres.PurposeRecovery: {
		login: "cartulary_recovery_login",
		role:  "cartulary_recovery",
	},
}

func purposeDSN(dsn string, purpose postgres.Purpose) (string, error) {
	contract, ok := testPurposeContracts[purpose]
	if !ok {
		return "", errors.New("unsupported test postgres purpose")
	}
	return purposeCertificateDSN(dsn, contract.login)
}

func templateDatabaseName(suiteHash string, schemaHash string) string {
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	return truncateIdentifier(fmt.Sprintf("ct_tpl_%s_%s", suiteHash, pgschema.ShortHash(TemplateDigest(schemaHash))), 63)
}

func recordSuiteEvent(event suiteservices.Event) {
	_ = suiteservices.RecordEvent(nil, event)
}

func postgresPreparationDetails(strategy string, templateDB string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	details := postgresFixtureDetails(strategy, reuseScope, attribution, duration)
	details["preparation_strategy"] = strategy
	if templateDB != "" {
		details["template_database"] = templateDB
	}
	return details
}

func postgresFixtureDetails(strategy string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	if reuseScope == "" {
		reuseScope = suiteservices.FixtureReusePerTest
	}
	details := map[string]any{
		"duration_ms":          duration.Milliseconds(),
		"reuse_scope":          reuseScope,
		"strategy":             strategy,
		"preparation_strategy": strategy,
		"target":               suiteservices.LookupEnvValue(nil, suiteservices.TargetEnv),
	}
	if attribution.TestName != "" {
		details["test_name"] = attribution.TestName
	}
	if attribution.CallerFile != "" {
		details["caller_file"] = attribution.CallerFile
	}
	callerPackage := attribution.CallerPackage
	if callerPackage == "" && attribution.HarnessPackage != "" {
		callerPackage = "internal/testutil/" + attribution.HarnessPackage
	}
	if callerPackage != "" {
		details["caller_package"] = callerPackage
	}
	if attribution.PostgresFixturePolicy != "" {
		details["fixture_policy"] = attribution.PostgresFixturePolicy
	}
	if hash := fixtureSchemaHash(); hash != "" {
		details["schema_hash"] = hash
	}
	if fixtureClass := postgresFixtureClass(attribution.PostgresFixturePolicy, reuseScope); fixtureClass != "" {
		details["fixture_class"] = fixtureClass
	}
	if attribution.ReuseGroup != "" {
		details["reuse_group"] = attribution.ReuseGroup
	}
	return details
}

func fixtureSchemaHash() string {
	if hash := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGSchemaHashEnv)); hash != "" {
		return hash
	}
	schemaHashOnce.Do(func() {
		schemaHashValue, schemaHashErr = pgschema.Hash()
	})
	if schemaHashErr != nil {
		return ""
	}
	return schemaHashValue
}

func postgresFixtureClass(policy string, reuseScope string) string {
	switch policy {
	case postgresFixturePolicyTransaction:
		return "transaction"
	case postgresFixturePolicyTemplateClone:
		return "isolated_clone"
	case "suite_template":
		return "suite_template"
	}
	switch reuseScope {
	case suiteservices.FixtureReuseMigrationScratch:
		return "migration_scratch"
	case suiteservices.FixtureReuseTransaction:
		return "transaction"
	}
	return ""
}

func resolvePostgresFixturePolicy(attribution fixtureAttribution) string {
	topLevelTest := topLevelTestName(attribution.TestName)
	if policy := lookupFixturePolicy(postgresFixturePolicyTestsEnv, topLevelTest); policy != "" {
		return policy
	}
	if policy := lookupFixturePolicy(postgresFixturePolicyPackagesEnv, attribution.CallerPackage); policy != "" {
		return policy
	}
	if policy := normalizePostgresFixturePolicy(suiteservices.LookupEnvValue(nil, postgresFixturePolicyDefaultEnv)); policy != "" {
		return policy
	}
	return ""
}

// ExplicitPostgresFixturePolicyT returns the scheduler-assigned fixture policy
// for the current test, or an empty string when no explicit assignment exists.
func ExplicitPostgresFixturePolicyT(t testing.TB) string {
	t.Helper()

	attribution := fixtureAttributionFor(t, "pgtest")
	topLevelTest := topLevelTestName(attribution.TestName)
	if policy := lookupFixturePolicy(postgresFixturePolicyTestsEnv, topLevelTest); policy != "" {
		return policy
	}
	if policy := lookupFixturePolicy(postgresFixturePolicyPackagesEnv, attribution.CallerPackage); policy != "" {
		return policy
	}
	return normalizePostgresFixturePolicy(suiteservices.LookupEnvValue(nil, postgresFixturePolicyDefaultEnv))
}

func requireSelectedPostgresFixturePolicyT(t testing.TB, want string) {
	t.Helper()

	got := ExplicitPostgresFixturePolicyT(t)
	if err := validateSelectedPostgresFixturePolicy(got, want); err != nil {
		t.Fatal(err)
	}
}

func validateSelectedPostgresFixturePolicy(got string, want string) error {
	if got == "" {
		return fmt.Errorf("postgres fixture policy %q must be explicitly assigned", want)
	}
	if got != want {
		return fmt.Errorf("postgres fixture policy mismatch: call site selected %q, target selected %q", want, got)
	}
	return nil
}

func lookupFixturePolicy(envName string, key string) string {
	key = normalizeFixturePolicyKey(key)
	if key == "" {
		return ""
	}
	for _, assignment := range splitFixturePolicyAssignments(suiteservices.LookupEnvValue(nil, envName)) {
		name, value, ok := strings.Cut(assignment, "=")
		if !ok {
			continue
		}
		if normalizeFixturePolicyKey(name) == key {
			return normalizePostgresFixturePolicy(value)
		}
	}
	return ""
}

func splitFixturePolicyAssignments(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})
}

func topLevelTestName(testName string) string {
	if before, _, ok := strings.Cut(testName, "/"); ok {
		return before
	}
	return testName
}

func normalizeFixturePolicyKey(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "./")
	return value
}

func normalizePostgresFixturePolicy(value string) string {
	switch strings.TrimSpace(value) {
	case postgresFixturePolicyTemplateClone:
		return postgresFixturePolicyTemplateClone
	case postgresFixturePolicyTransaction:
		return postgresFixturePolicyTransaction
	default:
		return ""
	}
}

func fixtureAttributionFor(t testing.TB, harnessPackage string) fixtureAttribution {
	t.Helper()

	attribution := fixtureAttribution{TestName: t.Name(), HarnessPackage: harnessPackage}
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
