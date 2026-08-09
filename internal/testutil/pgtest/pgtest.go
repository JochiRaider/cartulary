package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

const postgresImage = "postgres:16-alpine"

const (
	postgresPortMappingTimeout   = 15 * time.Second
	postgresClientReadyTimeout   = 15 * time.Second
	postgresHealthPollInterval   = 250 * time.Millisecond
	postgresClientAttemptTimeout = 5 * time.Second

	postgresFixturePolicyTemplateClone = "template_clone"
	postgresFixturePolicyPackageReset  = "package_reset"
	postgresFixturePolicyTransaction   = "transaction"
	postgresFixturePolicyGroupClone    = "group_clone"
)

const (
	postgresFixturePolicyTestsEnv    = "CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS"
	postgresFixturePolicyPackagesEnv = "CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES"
	postgresFixturePolicyDefaultEnv  = "CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT"
	postgresResetTablesTestsEnv      = "CARTULARY_POSTGRES_RESET_TABLES_TESTS"
	postgresResetTablesPackagesEnv   = "CARTULARY_POSTGRES_RESET_TABLES_PACKAGES"
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
	Container testcontainers.Container
	Host      string
	Port      string
	User      string
	Password  string

	adminDSN    string
	dsnTemplate string
	templateDB  string
	schemaHash  string
	suiteHash   string
	processHash string
	counter     uint64
	shared      bool
	attached    bool

	templateMu sync.Mutex

	packageDBMu sync.Mutex
	packageDBs  map[string]*packageDatabase
	groupDBMu   sync.Mutex
	groupDBs    map[string]*groupDatabase
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

type StartOptions struct {
	Labels         map[string]string
	Observer       testcontainersx.StartObserver
	AttemptTimeout time.Duration
}

type RollbackDB struct {
	tx   pgx.Tx
	conn *pgx.Conn
}

type packageDatabase struct {
	mu         sync.Mutex
	db         *TestDatabase
	resetPlans map[string]resetPlan
	resetSQLs  map[string]string
}

type resetPlan struct {
	statement string
	tables    []string
}

type groupDatabase struct {
	mu                sync.Mutex
	db                *TestDatabase
	cleanupRegistered bool
}

type fixtureAttribution struct {
	TestName              string
	CallerFile            string
	CallerPackage         string
	HarnessPackage        string
	PostgresFixturePolicy string
	PostgresResetTables   []string
	ReuseGroup            string
}

var (
	sharedHarnessMu   sync.Mutex
	sharedHarness     *Harness
	schemaHashOnce    sync.Once
	schemaHashValue   string
	schemaHashErr     error
	startOwnedHarness = StartOwned
	startContainerFn  = testcontainers.GenericContainer
	waitReadyFn       = func(ctx context.Context, harness *Harness) error {
		return harness.WaitReady(ctx)
	}
	startPreflightFn          func(context.Context) (string, error)
	startSleepFn              func(context.Context, time.Duration) error
	pingAdminDSNFn            = pingAdminDSN
	migrateDatabaseFn         = database_migrations.Apply
	applyMigrationsThrough    = applyCanonicalMigrationsThrough
	rollbackMigrationsThrough = rollbackCanonicalMigrationsThrough
	createDatabaseFn          = createDatabase
	dropDatabaseFn            = dropDatabase
	markTemplateDBFn          = markTemplateDatabase
	listMutableTablesFn       = listMutableTables
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
	return StartOwnedWithLabels(ctx, nil)
}

func StartOwnedWithLabels(ctx context.Context, labels map[string]string) (*Harness, error) {
	return StartOwnedWithOptions(ctx, StartOptions{Labels: labels})
}

func StartOwnedWithOptions(ctx context.Context, options StartOptions) (*Harness, error) {
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

func startHarnessWithOptions(ctx context.Context, options StartOptions) (*Harness, error) {
	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "postgres",
			"POSTGRES_USER":     "cartulary",
			"POSTGRES_PASSWORD": "cartulary",
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
		return nil, fmt.Errorf("resolve postgres host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("resolve postgres mapped port: %w", err)
	}

	harness := &Harness{
		Container:   container,
		Host:        host,
		Port:        port.Port(),
		User:        "cartulary",
		Password:    "cartulary",
		schemaHash:  pgschema.MustHash(),
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
	}
	harness.adminDSN = harness.dsnFor("postgres")

	if err := waitReadyFn(ctx, harness); err != nil {
		return nil, fmt.Errorf("wait for postgres readiness: %w", err)
	}

	attemptSucceeded = true
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
	if advertisedHash := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.PGSchemaHashEnv)); advertisedHash != "" && advertisedHash != harness.schemaHash {
		return nil, false, fmt.Errorf("attach postgres harness: %s mismatch: got %s want %s", suiteservices.PGSchemaHashEnv, advertisedHash, harness.schemaHash)
	}
	if err := pingAdminDSNFn(ctx, adminDSN); err != nil {
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

	var lastErr error
	for {
		attemptCtx, attemptCancel := context.WithTimeout(readyCtx, postgresClientAttemptTimeout)
		err := pingAdminDSNFn(attemptCtx, h.adminDSN)
		attemptCancel()
		if err == nil {
			return nil
		}
		if isNonRetryablePostgresReadinessError(err) {
			return &postgresReadinessError{LastErr: err}
		}

		lastErr = err
		select {
		case <-readyCtx.Done():
			if lastErr == nil {
				lastErr = readyCtx.Err()
			}
			return &postgresReadinessError{
				LastErr:         lastErr,
				DeadlineExpired: errors.Is(readyCtx.Err(), context.DeadlineExceeded),
			}
		case <-ticker.C:
		}
	}
}

type postgresReadinessError struct {
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

func isRetryablePostgresStartupFailure(err error) bool {
	var readinessErr *postgresReadinessError
	if errors.As(err, &readinessErr) {
		if !readinessErr.DeadlineExpired {
			return false
		}
		return !isNonRetryablePostgresReadinessError(readinessErr.LastErr)
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
		case "28P01", "28000", "3D000":
			return true
		}
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "password authentication failed") ||
		strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "role") && strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "database") && strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "no pg_hba.conf entry") ||
		strings.Contains(lower, "invalid connection")
}

func (h *Harness) AdminDSN() string {
	return h.adminDSN
}

func (h *Harness) NewDatabase(ctx context.Context, prefix string) (*TestDatabase, error) {
	return h.newDatabase(ctx, prefix, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) NewMigrationDatabase(ctx context.Context, prefix string) (*TestDatabase, error) {
	return h.newDatabase(ctx, prefix, suiteservices.FixtureReuseMigrationScratch, fixtureAttribution{})
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

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		return nil, fmt.Errorf("open test database handle: %w", err)
	}
	defer db.Close()

	migrateStart := time.Now()
	source, err := dbmigrations.Source()
	if err != nil {
		return nil, fmt.Errorf("load migration source: %w", err)
	}
	err = migrateDatabaseFn(ctx, db, source)
	if err != nil {
		return nil, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", reuseScope, attribution, time.Since(migrateStart)),
	})

	return testDB, nil
}

func (h *Harness) ensureLocalTemplateDatabase(ctx context.Context) error {
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
	db, err := sql.Open("pgx", templateDSN)
	if err != nil {
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("open local postgres template database: %w", err)
	}
	migrateStart := time.Now()
	source, err := dbmigrations.Source()
	if err != nil {
		_ = db.Close()
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("load migration source: %w", err)
	}
	if err := migrateDatabaseFn(ctx, db, source); err != nil {
		_ = db.Close()
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("migrate local postgres template database: %w", err)
	}
	if err := db.Close(); err != nil {
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return fmt.Errorf("close local postgres template database: %w", err)
	}
	if err := markTemplateDBFn(ctx, h.adminDSN, name); err != nil {
		_ = h.dropDatabase(context.Background(), name, suiteservices.FixtureReuseSuiteTemplate, fixtureAttribution{})
		return err
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
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationTemplate, name, suiteservices.FixtureReuseSuiteTemplate, attribution, time.Since(migrateStart)),
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
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if h.retainPreparedDatabaseOnCleanup() {
			h.recordRetainedDatabase(testDB.Name, suiteservices.FixtureReusePerTest, attribution)
			return
		}
		if err := h.dropDatabase(context.Background(), testDB.Name, suiteservices.FixtureReusePerTest, attribution); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})
	return testDB
}

func (h *Harness) NewMigrationDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()
	attribution := fixtureAttributionFor(t, "pgtest")
	testDB, err := h.newDatabase(context.Background(), prefix, suiteservices.FixtureReuseMigrationScratch, attribution)
	if err != nil {
		t.Fatalf("create migration scratch database: %v", err)
	}
	t.Cleanup(func() {
		if h.retainDatabaseOnCleanup() {
			h.recordRetainedDatabase(testDB.Name, suiteservices.FixtureReuseMigrationScratch, attribution)
			return
		}
		if err := h.dropDatabase(context.Background(), testDB.Name, suiteservices.FixtureReuseMigrationScratch, attribution); err != nil {
			t.Fatalf("drop migration scratch database: %v", err)
		}
	})
	return testDB
}

func (h *Harness) MigrationDatabaseT(t testing.TB, prefix string) *MigrationDatabase {
	return h.migrationDatabaseT(t, prefix, func(ctx context.Context, db *MigrationDatabase) error {
		source, err := dbmigrations.Source()
		if err != nil {
			return err
		}
		return migrateDatabaseFn(ctx, db.SQL(), source)
	})
}

func (h *Harness) MigrationDatabaseThroughT(t testing.TB, prefix string, version int64) *MigrationDatabase {
	return h.migrationDatabaseT(t, prefix, func(ctx context.Context, db *MigrationDatabase) error {
		return db.ApplyThrough(ctx, version)
	})
}

func (h *Harness) migrationDatabaseT(t testing.TB, prefix string, apply func(context.Context, *MigrationDatabase) error) *MigrationDatabase {
	t.Helper()

	// MigrationDatabaseT always creates a fresh scratch database and replays the
	// requested migration path. Use PrepareIsolatedDatabaseT or an explicit
	// group/reset/transaction helper
	// for current-head schema assertions; keep MigrationDatabaseT for tests that
	// prove migration runner behavior, boundary upgrades, or backfills.
	testDB := h.NewMigrationDatabaseT(t, prefix)

	var db *sql.DB
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
		}
	})

	openedDB, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration scratch database: %v", err)
	}
	db = openedDB
	migrationDB := &MigrationDatabase{db: db, identity: issuedMigrationDatabaseIdentity}

	migrateStart := time.Now()
	if err := apply(context.Background(), migrationDB); err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", suiteservices.FixtureReuseMigrationScratch, fixtureAttributionFor(t, "pgtest"), time.Since(migrateStart)),
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
	if version <= 0 {
		return errors.New("migration apply-through version must be positive")
	}
	if !db.valid() {
		return errors.New("migration database capability is not harness-issued")
	}
	return applyMigrationsThrough(ctx, db.db, version)
}

// RollbackThrough rolls this scratch database back through a non-negative target.
func (db *MigrationDatabase) RollbackThrough(ctx context.Context, version int64) error {
	if version < 0 {
		return errors.New("migration rollback-through version must be non-negative")
	}
	if !db.valid() {
		return errors.New("migration database capability is not harness-issued")
	}
	return rollbackMigrationsThrough(ctx, db.db, version)
}

func (db *MigrationDatabase) valid() bool {
	return db != nil && db.identity == issuedMigrationDatabaseIdentity && db.db != nil
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
	sourceFS, err := fs.Sub(dbmigrations.Files, dbmigrations.EmbeddedPath)
	if err != nil {
		return nil, fmt.Errorf("resolve canonical migration source: %w", err)
	}
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		sourceFS,
		goose.WithDisableGlobalRegistry(true),
		goose.WithLogger(log.New(io.Discard, "", 0)),
	)
}

func (h *Harness) PreparePackageResetDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	// Package reset is an explicit compatibility path for tests that prove a
	// closed reset surface. Ordinary integration tests use isolated clones,
	// grouped committed-state tests use PrepareGroupDatabaseT, and store tests
	// use BeginRollbackDBT.
	requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyPackageReset)
	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyPackageReset
	attribution.PostgresResetTables = resolvePostgresResetTables(attribution)

	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeIdentifier(prefix)
	}
	attribution.ReuseGroup = key
	fixture := h.packageDatabase(key)
	fixture.mu.Lock()

	if fixture.db == nil {
		testDB, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePackage, attribution)
		if err != nil {
			fixture.mu.Unlock()
			t.Fatalf("prepare package postgres database: %v", err)
		}
		fixture.db = testDB
	} else if err := h.resetPackageDatabase(context.Background(), fixture, attribution); err != nil {
		fixture.mu.Unlock()
		t.Fatalf("reset package postgres database: %v", err)
	}

	t.Cleanup(func() {
		fixture.mu.Unlock()
	})
	return fixture.db
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
	fixture := h.packageDatabase(key)
	fixture.mu.Lock()

	if fixture.db == nil {
		testDB, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePackage, attribution)
		if err != nil {
			fixture.mu.Unlock()
			t.Fatalf("prepare transaction postgres database: %v", err)
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
		_ = conn.Close(context.Background())
		fixture.mu.Unlock()
		t.Fatalf("begin postgres rollback transaction: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresTransaction,
		Name:    fixture.db.Name,
		Details: postgresFixtureDetails("", suiteservices.FixtureReuseTransaction, attribution, time.Since(start)),
	})

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
		_ = conn.Close(context.Background())
		fixture.mu.Unlock()
	})
	return &RollbackDB{tx: tx, conn: conn}
}

func (h *Harness) PrepareGroupDatabaseT(t testing.TB, prefix string, groupKey string) *TestDatabase {
	t.Helper()
	requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyGroupClone)

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyGroupClone
	return h.prepareGroupDatabaseT(t, prefix, groupKey, attribution)
}

func (h *Harness) prepareGroupDatabaseT(t testing.TB, prefix string, groupKey string, attribution fixtureAttribution) *TestDatabase {
	t.Helper()

	attribution.PostgresFixturePolicy = postgresFixturePolicyGroupClone
	key := attribution.CallerPackage + ":" + topLevelTestName(attribution.TestName) + ":" + sanitizeIdentifier(groupKey)
	if key == "::" {
		key = sanitizeIdentifier(prefix)
	}
	attribution.ReuseGroup = key
	fixture := h.groupDatabase(key)
	fixture.mu.Lock()
	defer fixture.mu.Unlock()

	if fixture.db == nil {
		testDB, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReuseGroup, attribution)
		if err != nil {
			t.Fatalf("prepare grouped postgres database: %v", err)
		}
		fixture.db = testDB
	}
	if !fixture.cleanupRegistered {
		fixture.cleanupRegistered = true
		t.Cleanup(func() {
			if h.retainPreparedDatabaseOnCleanup() {
				h.recordRetainedDatabase(fixture.db.Name, suiteservices.FixtureReuseGroup, attribution)
				return
			}
			if err := h.dropDatabase(context.Background(), fixture.db.Name, suiteservices.FixtureReuseGroup, attribution); err != nil {
				t.Fatalf("drop grouped postgres database: %v", err)
			}
		})
	}

	return fixture.db
}

func (h *Harness) packageDatabase(key string) *packageDatabase {
	h.packageDBMu.Lock()
	defer h.packageDBMu.Unlock()

	if h.packageDBs == nil {
		h.packageDBs = make(map[string]*packageDatabase)
	}
	fixture := h.packageDBs[key]
	if fixture == nil {
		fixture = &packageDatabase{}
		h.packageDBs[key] = fixture
	}
	return fixture
}

func (h *Harness) groupDatabase(key string) *groupDatabase {
	h.groupDBMu.Lock()
	defer h.groupDBMu.Unlock()

	if h.groupDBs == nil {
		h.groupDBs = make(map[string]*groupDatabase)
	}
	fixture := h.groupDBs[key]
	if fixture == nil {
		fixture = &groupDatabase{}
		h.groupDBs[key] = fixture
	}
	return fixture
}

func (h *Harness) ResetDatabase(ctx context.Context, name string) error {
	return h.resetDatabase(ctx, name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) resetDatabase(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	start := time.Now()
	db, err := sql.Open("pgx", h.dsnFor(name))
	if err != nil {
		return fmt.Errorf("open postgres reset handle: %w", err)
	}
	defer db.Close()

	statement, tables, err := h.buildResetStatement(ctx, db, nil)
	if err != nil {
		return err
	}
	beforeCounts, err := countRowsByTable(ctx, db, tables)
	if err != nil {
		return err
	}
	gooseBefore, gooseExistsBefore, err := countRowsIfTableExists(ctx, db, "goose_db_version")
	if err != nil {
		return err
	}
	routeIDBefore, routeIDExistsBefore, err := countRowsIfTableExists(ctx, db, "route_idempotency")
	if err != nil {
		return err
	}
	if statement != "" {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("truncate mutable postgres tables: %w", err)
		}
	}
	afterCounts, err := countRowsByTable(ctx, db, tables)
	if err != nil {
		return err
	}
	gooseAfter, gooseExistsAfter, err := countRowsIfTableExists(ctx, db, "goose_db_version")
	if err != nil {
		return err
	}
	routeIDAfter, routeIDExistsAfter, err := countRowsIfTableExists(ctx, db, "route_idempotency")
	if err != nil {
		return err
	}
	details := postgresFixtureDetails("", reuseScope, attribution, time.Since(start))
	addPostgresResetProofDetails(details, tables, beforeCounts, afterCounts, gooseBefore, gooseExistsBefore, gooseAfter, gooseExistsAfter, routeIDBefore, routeIDExistsBefore, routeIDAfter, routeIDExistsAfter)
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBReset,
		Name:    name,
		Details: details,
	})
	return nil
}

func (h *Harness) resetPackageDatabase(ctx context.Context, fixture *packageDatabase, attribution fixtureAttribution) error {
	start := time.Now()
	db, err := sql.Open("pgx", h.dsnFor(fixture.db.Name))
	if err != nil {
		return fmt.Errorf("open postgres reset handle: %w", err)
	}
	defer db.Close()

	resetKey := strings.Join(attribution.PostgresResetTables, ",")
	if fixture.resetPlans == nil {
		fixture.resetPlans = make(map[string]resetPlan)
	}
	plan, ok := fixture.resetPlans[resetKey]
	if !ok {
		var err error
		plan.statement, plan.tables, err = h.buildResetStatement(ctx, db, attribution.PostgresResetTables)
		if err != nil {
			return err
		}
		fixture.resetPlans[resetKey] = plan
		if fixture.resetSQLs == nil {
			fixture.resetSQLs = make(map[string]string)
		}
		fixture.resetSQLs[resetKey] = plan.statement
	}
	beforeCounts, err := countRowsByTable(ctx, db, plan.tables)
	if err != nil {
		return err
	}
	gooseBefore, gooseExistsBefore, err := countRowsIfTableExists(ctx, db, "goose_db_version")
	if err != nil {
		return err
	}
	routeIDBefore, routeIDExistsBefore, err := countRowsIfTableExists(ctx, db, "route_idempotency")
	if err != nil {
		return err
	}
	if plan.statement != "" {
		if _, err := db.ExecContext(ctx, plan.statement); err != nil {
			return fmt.Errorf("truncate mutable postgres tables: %w", err)
		}
	}
	afterCounts, err := countRowsByTable(ctx, db, plan.tables)
	if err != nil {
		return err
	}
	gooseAfter, gooseExistsAfter, err := countRowsIfTableExists(ctx, db, "goose_db_version")
	if err != nil {
		return err
	}
	routeIDAfter, routeIDExistsAfter, err := countRowsIfTableExists(ctx, db, "route_idempotency")
	if err != nil {
		return err
	}
	details := postgresFixtureDetails("", suiteservices.FixtureReusePackage, attribution, time.Since(start))
	addPostgresResetProofDetails(details, plan.tables, beforeCounts, afterCounts, gooseBefore, gooseExistsBefore, gooseAfter, gooseExistsAfter, routeIDBefore, routeIDExistsBefore, routeIDAfter, routeIDExistsAfter)
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBReset,
		Name:    fixture.db.Name,
		Details: details,
	})
	return nil
}

func (h *Harness) buildResetStatement(ctx context.Context, db *sql.DB, resetTables []string) (string, []string, error) {
	if len(resetTables) > 0 {
		if err := validateTargetedResetTables(ctx, db, resetTables); err != nil {
			return "", nil, err
		}
		tables := append([]string(nil), resetTables...)
		quotedTables := make([]string, len(tables))
		for index, table := range tables {
			quotedTables[index] = quoteIdentifier(table)
		}
		return "TRUNCATE TABLE " + strings.Join(quotedTables, ", ") + " RESTART IDENTITY CASCADE", tables, nil
	}
	tables, err := listMutableTablesFn(ctx, db)
	if err != nil {
		return "", nil, err
	}
	if len(tables) == 0 {
		return "", nil, nil
	}
	quotedTables := make([]string, len(tables))
	for index, table := range tables {
		quotedTables[index] = quoteIdentifier(table)
	}
	return "TRUNCATE TABLE " + strings.Join(quotedTables, ", ") + " RESTART IDENTITY CASCADE", tables, nil
}

func countRowsByTable(ctx context.Context, db *sql.DB, tables []string) (map[string]int64, error) {
	counts := make(map[string]int64, len(tables))
	for _, table := range tables {
		count, err := countRows(ctx, db, table)
		if err != nil {
			return nil, err
		}
		counts[table] = count
	}
	return counts, nil
}

func countRowsIfTableExists(ctx context.Context, db *sql.DB, table string) (int64, bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1::text) IS NOT NULL`, "public."+table).Scan(&exists); err != nil {
		return 0, false, fmt.Errorf("check postgres table %s: %w", table, err)
	}
	if !exists {
		return 0, false, nil
	}
	count, err := countRows(ctx, db, table)
	if err != nil {
		return 0, true, err
	}
	return count, true, nil
}

func countRows(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var count int64
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifier(table))).Scan(&count); err != nil {
		return 0, fmt.Errorf("count postgres table %s: %w", table, err)
	}
	return count, nil
}

func addPostgresResetProofDetails(details map[string]any, tables []string, beforeCounts map[string]int64, afterCounts map[string]int64, gooseBefore int64, gooseExistsBefore bool, gooseAfter int64, gooseExistsAfter bool, routeIDBefore int64, routeIDExistsBefore bool, routeIDAfter int64, routeIDExistsAfter bool) {
	if len(tables) > 0 {
		details["actual_reset_tables"] = strings.Join(tables, ",")
	}
	details["actual_reset_table_count"] = len(tables)
	details["reset_row_counts_before"] = beforeCounts
	details["reset_row_counts_after"] = afterCounts
	if gooseExistsBefore || gooseExistsAfter {
		details["goose_db_version_rows_before"] = gooseBefore
		details["goose_db_version_rows_after"] = gooseAfter
		details["goose_db_version_preserved"] = gooseExistsBefore && gooseExistsAfter && gooseBefore == gooseAfter
	}
	if routeIDExistsBefore || routeIDExistsAfter {
		details["route_idempotency_rows_before"] = routeIDBefore
		details["route_idempotency_rows_after"] = routeIDAfter
	}
}

func listMutableTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT tablename
		  FROM pg_tables
		 WHERE schemaname = 'public'
   AND tablename <> 'goose_db_version'
 ORDER BY tablename`)
	if err != nil {
		return nil, fmt.Errorf("list mutable postgres tables: %w", err)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan mutable postgres table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mutable postgres tables: %w", err)
	}
	return tables, nil
}

func validateTargetedResetTables(ctx context.Context, db *sql.DB, resetTables []string) error {
	tableSet := make(map[string]struct{}, len(resetTables))
	for _, table := range resetTables {
		table = strings.TrimSpace(table)
		if table == "" {
			return fmt.Errorf("targeted postgres reset includes an empty table name")
		}
		tableSet[table] = struct{}{}
	}

	mutableTables, err := listMutableTablesFn(ctx, db)
	if err != nil {
		return err
	}
	mutableSet := make(map[string]struct{}, len(mutableTables))
	for _, table := range mutableTables {
		mutableSet[table] = struct{}{}
	}
	for table := range tableSet {
		if _, ok := mutableSet[table]; !ok {
			return fmt.Errorf("targeted postgres reset table %s is not a mutable public table", table)
		}
	}

	rows, err := db.QueryContext(ctx, `
SELECT source.relname AS source_table, target.relname AS target_table
  FROM pg_constraint constraint_row
  JOIN pg_class source ON source.oid = constraint_row.conrelid
  JOIN pg_namespace source_namespace ON source_namespace.oid = source.relnamespace
  JOIN pg_class target ON target.oid = constraint_row.confrelid
  JOIN pg_namespace target_namespace ON target_namespace.oid = target.relnamespace
 WHERE constraint_row.contype = 'f'
   AND source_namespace.nspname = 'public'
   AND target_namespace.nspname = 'public'
 ORDER BY source.relname, target.relname`)
	if err != nil {
		return fmt.Errorf("list postgres reset table dependencies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sourceTable string
		var targetTable string
		if err := rows.Scan(&sourceTable, &targetTable); err != nil {
			return fmt.Errorf("scan postgres reset table dependency: %w", err)
		}
		_, sourceSelected := tableSet[sourceTable]
		_, targetSelected := tableSet[targetTable]
		if sourceSelected != targetSelected {
			return fmt.Errorf("targeted postgres reset table set must include both sides of FK %s -> %s", sourceTable, targetTable)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate postgres reset table dependencies: %w", err)
	}
	return nil
}

func BeginRollbackTxT(t testing.TB, db *sql.DB) *sql.Tx {
	t.Helper()

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = resolvePostgresFixturePolicy(attribution)
	if attribution.PostgresFixturePolicy == "" || attribution.PostgresFixturePolicy == postgresFixturePolicyPackageReset {
		attribution.PostgresFixturePolicy = postgresFixturePolicyTransaction
	}
	start := time.Now()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin postgres rollback transaction: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresTransaction,
		Details: postgresFixtureDetails("", suiteservices.FixtureReuseTransaction, attribution, time.Since(start)),
	})
	t.Cleanup(func() {
		_ = tx.Rollback()
	})
	return tx
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

func (h *Harness) DropMigrationDatabase(ctx context.Context, name string) error {
	return h.dropDatabase(ctx, name, suiteservices.FixtureReuseMigrationScratch, fixtureAttribution{})
}

func (h *Harness) dropDatabase(ctx context.Context, name string, reuseScope string, attribution fixtureAttribution) error {
	start := time.Now()
	if err := dropDatabaseFn(ctx, h.adminDSN, name); err != nil {
		return err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBDropped,
		Name:    name,
		Details: postgresFixtureDetails("", reuseScope, attribution, time.Since(start)),
	})

	return nil
}

func dropDatabase(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return fmt.Errorf("drop test database %s: %w", name, err)
	}
	return nil
}

func (h *Harness) retainDatabaseOnCleanup() bool {
	return h.attached && h.templateDB != "" && suiteservices.SuiteActive(nil)
}

func (h *Harness) retainPreparedDatabaseOnCleanup() bool {
	return h.retainDatabaseOnCleanup()
}

func (h *Harness) recordRetainedDatabase(name string, reuseScope string, attribution fixtureAttribution) {
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBRetained,
		Name:    name,
		Details: postgresFixtureDetails(suiteservices.PostgresPreparationTemplateClone, reuseScope, attribution, 0),
	})
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

func (db *TestDatabase) Env() map[string]string {
	return map[string]string{
		suiteservices.PostgresDSNEnv: db.DSN,
	}
}

func (db *TestDatabase) EnvForServiceRef(serviceRef string) map[string]string {
	key, err := postgres.EnvKeyForServiceRef(serviceRef)
	if err != nil {
		return map[string]string{}
	}
	return map[string]string{key: db.DSN}
}

func (h *Harness) dsnFor(database string) string {
	if h.dsnTemplate != "" {
		return strings.ReplaceAll(h.dsnTemplate, "{database}", database)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", h.User, h.Password, h.Host, h.Port, database)
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
	return createDatabaseFn(ctx, h.adminDSN, name, templateDB)
}

func pingAdminDSN(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

func createDatabase(ctx context.Context, adminDSN string, name string, templateDB string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	statement := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	if templateDB != "" {
		statement += fmt.Sprintf(` TEMPLATE "%s"`, templateDB)
	}
	if _, err := admin.ExecContext(ctx, statement); err != nil {
		if templateDB != "" {
			return fmt.Errorf("create test database %s from template %s: %w", name, templateDB, err)
		}
		return fmt.Errorf("create test database %s: %w", name, err)
	}
	return nil
}

func markTemplateDatabase(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin handle: %w", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		return fmt.Errorf("terminate template database connections: %w", err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH ALLOW_CONNECTIONS false`, name)); err != nil {
		return fmt.Errorf("disable template database connections: %w", err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE true`, name)); err != nil {
		return fmt.Errorf("mark template database as template: %w", err)
	}
	return nil
}

func templateDatabaseName(suiteHash string, schemaHash string) string {
	if suiteHash == "" {
		suiteHash = resolveSuiteHash()
	}
	return truncateIdentifier(fmt.Sprintf("ct_tpl_%s_%s", suiteHash, pgschema.ShortHash(schemaHash)), 63)
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
	if len(attribution.PostgresResetTables) > 0 {
		details["reset_tables"] = strings.Join(attribution.PostgresResetTables, ",")
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
	case postgresFixturePolicyPackageReset, postgresFixturePolicyGroupClone:
		return "reusable_database"
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
	case suiteservices.FixtureReusePackage, suiteservices.FixtureReuseGroup:
		return "reusable_database"
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

func resolvePostgresResetTables(attribution fixtureAttribution) []string {
	topLevelTest := topLevelTestName(attribution.TestName)
	if tables := lookupResetTables(postgresResetTablesTestsEnv, topLevelTest); len(tables) > 0 {
		return tables
	}
	return lookupResetTables(postgresResetTablesPackagesEnv, attribution.CallerPackage)
}

func lookupResetTables(envName string, key string) []string {
	key = normalizeFixturePolicyKey(key)
	if key == "" {
		return nil
	}
	for _, assignment := range splitFixturePolicyAssignments(suiteservices.LookupEnvValue(nil, envName)) {
		name, value, ok := strings.Cut(assignment, "=")
		if !ok {
			continue
		}
		if normalizeFixturePolicyKey(name) == key {
			return normalizeResetTables(value)
		}
	}
	return nil
}

func normalizeResetTables(value string) []string {
	seen := make(map[string]struct{})
	tables := make([]string, 0)
	for _, table := range strings.FieldsFunc(value, func(r rune) bool {
		return r == '|' || r == ':' || r == '+'
	}) {
		table = strings.TrimSpace(table)
		if table == "" {
			continue
		}
		if _, ok := seen[table]; ok {
			continue
		}
		seen[table] = struct{}{}
		tables = append(tables, table)
	}
	return tables
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
	case postgresFixturePolicyPackageReset:
		return postgresFixturePolicyPackageReset
	case postgresFixturePolicyTransaction:
		return postgresFixturePolicyTransaction
	case postgresFixturePolicyGroupClone:
		return postgresFixturePolicyGroupClone
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

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
