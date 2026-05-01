package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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
	suiteHash   string
	processHash string
	counter     uint64
	shared      bool
	attached    bool

	packageDBMu sync.Mutex
	packageDBs  map[string]*packageDatabase
	groupDBMu   sync.Mutex
	groupDBs    map[string]*groupDatabase
}

type TestDatabase struct {
	Name string
	DSN  string
}

type StartOptions struct {
	Labels   map[string]string
	Observer testcontainersx.StartObserver
}

type RollbackDB struct {
	tx   pgx.Tx
	conn *pgx.Conn
}

type packageDatabase struct {
	mu        sync.Mutex
	db        *TestDatabase
	resetSQLs map[string]string
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
}

var (
	sharedHarnessMu   sync.Mutex
	sharedHarness     *Harness
	startOwnedHarness = StartOwned
	startContainerFn  = testcontainers.GenericContainer
	waitReadyFn       = func(ctx context.Context, harness *Harness) error {
		return harness.WaitReady(ctx)
	}
	startPreflightFn    func(context.Context) (string, error)
	startSleepFn        func(context.Context, time.Duration) error
	pingAdminDSNFn      = pingAdminDSN
	migrateDatabaseFn   = postgres.Migrate
	createDatabaseFn    = createDatabase
	dropDatabaseFn      = dropDatabase
	listMutableTablesFn = listMutableTables
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
		Service:      "postgres testcontainer",
		Image:        postgresImage,
		MaxAttempts:  3,
		RetryBackoff: 500 * time.Millisecond,
		Preflight:    startPreflightFn,
		Retryable:    isRetryablePostgresStartupFailure,
		Sleep:        startSleepFn,
		Observer:     options.Observer,
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
		suiteHash:   resolveSuiteHash(),
		processHash: suiteservices.ProcessHash(),
		attached:    true,
	}
	if err := pingAdminDSNFn(ctx, adminDSN); err != nil {
		return nil, false, fmt.Errorf("attach postgres harness: ping admin dsn: %w", err)
	}
	recordSuiteEvent(suiteservices.Event{Type: suiteservices.EventPostgresAttach})
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

func (h *Harness) PrepareDatabase(ctx context.Context, prefix string) (*TestDatabase, postgres.MigrationStatus, error) {
	return h.prepareDatabase(ctx, prefix, suiteservices.FixtureReusePerTest, fixtureAttribution{})
}

func (h *Harness) prepareDatabase(ctx context.Context, prefix string, reuseScope string, attribution fixtureAttribution) (*TestDatabase, postgres.MigrationStatus, error) {
	if h.templateDB != "" {
		name := h.nextDatabaseName(prefix)
		start := time.Now()
		if err := h.createDatabase(ctx, name, h.templateDB); err != nil {
			return nil, postgres.MigrationStatus{}, err
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
			}, postgres.MigrationStatus{
				Command:          "template-clone",
				Directory:        dbmigrations.RepositoryPath,
				TemplateClone:    true,
				TemplateDatabase: h.templateDB,
			}, nil
	}

	testDB, err := h.newDatabase(ctx, prefix, reuseScope, attribution)
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		return nil, postgres.MigrationStatus{}, fmt.Errorf("open test database handle: %w", err)
	}
	defer db.Close()

	migrateStart := time.Now()
	status, err := migrateDatabaseFn(ctx, db, dbmigrations.Source(), "up")
	if err != nil {
		return nil, postgres.MigrationStatus{}, err
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", reuseScope, attribution, time.Since(migrateStart)),
	})

	return testDB, status, nil
}

func (h *Harness) PrepareDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	// PrepareDatabaseT is the isolated-database path. Prefer
	// PreparePackageDatabaseT for ordinary mutable integration tests, and keep
	// this helper for tests that intentionally need per-test database identity
	// or are asserting pgtest clone/cleanup semantics.
	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyTemplateClone
	testDB, _, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePerTest, attribution)
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

func (h *Harness) MigrationDatabaseT(t testing.TB, prefix string, command string, args ...string) *sql.DB {
	t.Helper()

	// MigrationDatabaseT always creates a fresh scratch database and replays the
	// requested migration path. Use PrepareDatabaseT or PreparePackageDatabaseT
	// for current-head schema assertions; keep MigrationDatabaseT for tests that
	// prove migration runner behavior, boundary upgrades, or backfills.
	attribution := fixtureAttributionFor(t, "pgtest")
	testDB, err := h.newDatabase(context.Background(), prefix, suiteservices.FixtureReuseMigrationScratch, attribution)
	if err != nil {
		t.Fatalf("create migration scratch database: %v", err)
	}

	var db *sql.DB
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
		}
		if h.retainDatabaseOnCleanup() {
			h.recordRetainedDatabase(testDB.Name, suiteservices.FixtureReuseMigrationScratch, attribution)
			return
		}
		if err := h.dropDatabase(context.Background(), testDB.Name, suiteservices.FixtureReuseMigrationScratch, attribution); err != nil {
			t.Fatalf("drop migration scratch database: %v", err)
		}
	})

	openedDB, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration scratch database: %v", err)
	}
	db = openedDB

	migrateStart := time.Now()
	if _, err := migrateDatabaseFn(context.Background(), db, dbmigrations.Source(), command, args...); err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", suiteservices.FixtureReuseMigrationScratch, attribution, time.Since(migrateStart)),
	})

	return db
}

func (h *Harness) PreparePackageDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	// PreparePackageDatabaseT selects the fixture policy assigned by the target
	// plan. Package reset is an explicit compatibility path for tests that prove
	// reset behavior; ordinary integration tests should use template clones, and
	// store-domain tests should use BeginRollbackDBT.
	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = resolvePostgresFixturePolicy(attribution)
	attribution.PostgresResetTables = resolvePostgresResetTables(attribution)
	if attribution.PostgresFixturePolicy == postgresFixturePolicyTransaction {
		t.Fatalf("postgres fixture policy %q requires BeginRollbackTxT, not PreparePackageDatabaseT", postgresFixturePolicyTransaction)
	}
	if attribution.PostgresFixturePolicy == postgresFixturePolicyGroupClone {
		t.Fatalf("postgres fixture policy %q requires PrepareGroupDatabaseT, not PreparePackageDatabaseT", postgresFixturePolicyGroupClone)
	}
	if attribution.PostgresFixturePolicy == postgresFixturePolicyTemplateClone {
		testDB, _, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePerTest, attribution)
		if err != nil {
			t.Fatalf("prepare isolated postgres database: %v", err)
		}
		t.Cleanup(func() {
			if h.retainPreparedDatabaseOnCleanup() {
				h.recordRetainedDatabase(testDB.Name, suiteservices.FixtureReusePerTest, attribution)
				return
			}
			if err := h.dropDatabase(context.Background(), testDB.Name, suiteservices.FixtureReusePerTest, attribution); err != nil {
				t.Fatalf("drop isolated postgres database: %v", err)
			}
		})
		return testDB
	}

	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeIdentifier(prefix)
	}
	fixture := h.packageDatabase(key)
	fixture.mu.Lock()

	if fixture.db == nil {
		testDB, _, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePackage, attribution)
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

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyTransaction
	key := attribution.CallerPackage
	if key == "" {
		key = sanitizeIdentifier(prefix)
	}
	fixture := h.packageDatabase(key)
	fixture.mu.Lock()

	if fixture.db == nil {
		testDB, _, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReusePackage, attribution)
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

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = postgresFixturePolicyGroupClone
	key := attribution.CallerPackage + ":" + topLevelTestName(attribution.TestName) + ":" + sanitizeIdentifier(groupKey)
	if key == "::" {
		key = sanitizeIdentifier(prefix)
	}
	fixture := h.groupDatabase(key)
	fixture.mu.Lock()
	defer fixture.mu.Unlock()

	if fixture.db == nil {
		testDB, _, err := h.prepareDatabase(context.Background(), prefix, suiteservices.FixtureReuseGroup, attribution)
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

	statement, err := h.buildResetStatement(ctx, db, nil)
	if err != nil {
		return err
	}
	if statement != "" {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("truncate mutable postgres tables: %w", err)
		}
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBReset,
		Name:    name,
		Details: postgresFixtureDetails("", reuseScope, attribution, time.Since(start)),
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
	if fixture.resetSQLs == nil {
		fixture.resetSQLs = make(map[string]string)
	}
	statement, ok := fixture.resetSQLs[resetKey]
	if !ok {
		var err error
		statement, err = h.buildResetStatement(ctx, db, attribution.PostgresResetTables)
		if err != nil {
			return err
		}
		fixture.resetSQLs[resetKey] = statement
	}
	if statement != "" {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("truncate mutable postgres tables: %w", err)
		}
	}
	recordSuiteEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBReset,
		Name:    fixture.db.Name,
		Details: postgresFixtureDetails("", suiteservices.FixtureReusePackage, attribution, time.Since(start)),
	})
	return nil
}

func (h *Harness) buildResetStatement(ctx context.Context, db *sql.DB, resetTables []string) (string, error) {
	if len(resetTables) > 0 {
		if err := validateTargetedResetTables(ctx, db, resetTables); err != nil {
			return "", err
		}
		tables := append([]string(nil), resetTables...)
		for index, table := range tables {
			tables[index] = quoteIdentifier(table)
		}
		return "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE", nil
	}
	tables, err := listMutableTablesFn(ctx, db)
	if err != nil {
		return "", err
	}
	if len(tables) == 0 {
		return "", nil
	}
	for index, table := range tables {
		tables[index] = quoteIdentifier(table)
	}
	return "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE", nil
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
		postgres.PostgresDSNEnv: db.DSN,
	}
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
	return details
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
	return postgresFixturePolicyTemplateClone
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
