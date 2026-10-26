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
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

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
	postgresStartupTimeout          = 180 * time.Second
	postgresStartupAttemptTimeout   = 35 * time.Second
	postgresPortMappingTimeout      = 15 * time.Second
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

// Harness grants database fixture access, never physical cluster ownership.
type Harness struct {
	binding     fixtureBinding
	generation  string
	retired     atomic.Bool
	operations  *databaseOperations
	adminDSN    string
	dsnTemplate string
	templateDB  string
	schemaHash  string
	suiteHash   string
	processHash string
	counter     uint64

	templateMu      sync.Mutex
	ownedDatabaseMu sync.Mutex
	ownedDatabases  map[string]struct{}
	quarantined     atomic.Bool

	events func(suiteservices.Event) error

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
	db         *sql.DB
	identity   *migrationDatabaseIdentity
	operations databaseOperations
}

// Nonzero size ensures distinct issuer tokens cannot share an address.
type migrationDatabaseIdentity struct{ _ byte }

var issuedMigrationDatabaseIdentity = &migrationDatabaseIdentity{}

var errGenerationQuarantined = errors.New("postgres cluster generation is quarantined after fixture settlement failure")

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
	schemaHashOnce  sync.Once
	schemaHashValue string
	schemaHashErr   error
)

func Start(t testing.TB) *Harness {
	t.Helper()
	suiteservices.RequireServiceDependencies(t, "pgtest", "postgres")

	harness, err := StartShared(t.Context())
	if err != nil {
		failPostgresSetup(t, "start", err)
	}

	return harness
}

// StartShared borrows only an explicitly admitted suite binding.
func StartShared(ctx context.Context) (*Harness, error) {
	return sharedAttachments.attach(ctx, processFixtureInputs())
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
	if err := h.createDatabase(ctx, name, "", reuseScope, attribution); err != nil {
		return nil, err
	}

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
	h.templateMu.Lock()
	template := h.templateDB
	h.templateMu.Unlock()
	if template == "" {
		return nil, errors.New("postgres fixture requires an explicitly prepared sealed template")
	}
	name := h.nextDatabaseName(prefix)
	if err := h.createDatabase(ctx, name, template, reuseScope, attribution); err != nil {
		return nil, err
	}
	if err := h.recordEvent(suiteservices.Event{
		Type: suiteservices.EventPostgresTemplateUse, Name: name,
		Details: map[string]any{"template_database": template, "target": h.binding.inputs[suiteservices.TargetEnv]},
	}); err != nil {
		return nil, errors.Join(err, h.disposeDatabase(ctx, name, reuseScope, attribution))
	}
	return &TestDatabase{Name: name, DSN: h.dsnFor(name)}, nil
}

func (h *Harness) PrepareIsolatedDatabaseT(t testing.TB, prefix string) *TestDatabase {
	t.Helper()

	h.requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyTemplateClone)
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
	db, err := OpenPurposeDatabase(t.Context(), testDB.DSN, purpose)
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
	testDB, err := h.newDatabase(t.Context(), "migration_scratch", suiteservices.FixtureReuseMigrationScratch, attribution)
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
		return h.databaseOperations().migrate(ctx, db.SQL(), source)
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

func (h *Harness) admitMigration() error {
	if err := h.admitPreparation(); err != nil {
		return err
	}
	if h.binding.inputs["CARTULARY_FIXTURE_CAPABILITY"] != "postgres_migration" || h.binding.inputs[suiteservices.PGIsolationEnv] != "exclusive" {
		return errors.New("postgres migration requires an admitted postgres_migration capability with exclusive isolation")
	}
	return nil
}

func (h *Harness) migrationDatabaseT(t testing.TB, apply func(context.Context, *MigrationDatabase) error) *MigrationDatabase {
	t.Helper()

	if err := h.admitMigration(); err != nil {
		failDatabasePreparation(t, "migration fixture admission", err)
	}

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

	openedDB, err := h.databaseOperations().openMigration(t.Context(), testDB.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatalf("open migration scratch database: %v", err)
	}
	db = openedDB
	migrationDB := &MigrationDatabase{db: db, identity: issuedMigrationDatabaseIdentity, operations: h.databaseOperations()}

	migrationDuration, err := suiteservices.MeasureMigration(t.Context(), h.binding.environment(), func() error { return apply(t.Context(), migrationDB) })
	if err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}
	_ = h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBMigrated,
		Name:    testDB.Name,
		Kind:    "scratch",
		Details: h.postgresPreparationDetails(suiteservices.PostgresPreparationFreshMigration, "", suiteservices.FixtureReuseMigrationScratch, fixtureAttributionFor(t, "pgtest"), migrationDuration),
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
	source, err := db.operations.source()
	if err != nil {
		return err
	}
	return db.operations.applyThrough(ctx, db.db, source, version)
}

// RollbackThrough rolls this scratch database back through a non-negative target.
func (db *MigrationDatabase) RollbackThrough(ctx context.Context, version int64) error {
	if err := validateRollbackThroughVersion(version); err != nil {
		return err
	}
	if !db.valid() {
		return errors.New("migration database capability is not harness-issued")
	}
	source, err := db.operations.source()
	if err != nil {
		return err
	}
	return db.operations.rollbackThrough(ctx, db.db, source, version)
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

func applyCanonicalMigrationsThrough(ctx context.Context, db *sql.DB, source *database_migrations.Source, version int64) error {
	provider, err := newMigrationProvider(db, source)
	if err != nil {
		return err
	}
	_, err = provider.UpTo(ctx, version)
	return err
}

func rollbackCanonicalMigrationsThrough(ctx context.Context, db *sql.DB, source *database_migrations.Source, version int64) error {
	provider, err := newMigrationProvider(db, source)
	if err != nil {
		return err
	}
	_, err = provider.DownTo(ctx, version)
	return err
}

func newMigrationProvider(db *sql.DB, source *database_migrations.Source) (*goose.Provider, error) {
	locker, err := sourcecatalog.NewSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("construct canonical migration locker: %w", err)
	}
	return sourcecatalog.NewProvider(db, source, locker)
}

func (h *Harness) BeginRollbackDBT(t testing.TB, prefix string) *RollbackDB {
	t.Helper()
	h.requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyTransaction)

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
	_ = h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresTransaction,
		Name:    fixture.db.Name,
		Details: h.postgresFixtureDetails("", suiteservices.FixtureReuseTransaction, attribution, time.Since(start)),
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
	var observationErr error
	dropErr := h.databaseOperations().drop(ctx, h.adminDSN, name, func(attempt postgrescleanup.DropAttempt) {
		if h.generation != "" {
			generation := h.generation
			attempt.ServiceGeneration = &generation
		}
		if attempt.Validate() == nil {
			observationErr = errors.Join(observationErr, h.recordEvent(suiteservices.Event{Type: suiteservices.EventPostgresDropAttempt, Details: map[string]any{"attempt": attempt}}))
		}
	})
	if err := errors.Join(dropErr, observationErr); err != nil {
		h.quarantined.Store(true)
		return err
	}
	details := h.postgresFixtureDetails("", reuseScope, attribution, time.Since(start))
	if err := h.recordEvent(suiteservices.Event{
		Type:    suiteservices.EventPostgresDBDropped,
		Name:    name,
		Details: details,
	}); err != nil {
		h.quarantined.Store(true)
		return err
	}

	return nil
}

func (h *Harness) ownsDatabase(name string) bool {
	h.ownedDatabaseMu.Lock()
	defer h.ownedDatabaseMu.Unlock()
	_, ok := h.ownedDatabases[name]
	return ok
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
	return strings.ReplaceAll(h.dsnTemplate, "{database}", database)
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

func resolveSuiteHash(inputs map[string]string) string {
	if hash := suiteservices.SuiteHash(inputs); hash != "" {
		return hash
	}
	return suiteservices.ShortHash("local-suite", 8)
}

func (h *Harness) nextDatabaseName(prefix string) string {
	suiteHash := h.suiteHash
	if suiteHash == "" {
		suiteHash = resolveSuiteHash(h.binding.environment())
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

func (h *Harness) createDatabase(ctx context.Context, name, template string, reuseScope string, attribution fixtureAttribution) error {
	if err := suiteservices.CheckServiceDependencies(h.binding.environment(), "postgres"); err != nil {
		return err
	}
	if err := h.admitPreparation(); err != nil {
		return err
	}
	start := time.Now()
	created, err := h.databaseOperations().create(ctx, h.adminDSN, name, template, func() error {
		h.ownedDatabaseMu.Lock()
		if h.ownedDatabases == nil {
			h.ownedDatabases = make(map[string]struct{})
		}
		h.ownedDatabases[name] = struct{}{}
		h.ownedDatabaseMu.Unlock()
		kind, strategy := "scratch", suiteservices.PostgresPreparationFreshMigration
		if template != "" {
			kind, strategy = "template-clone", suiteservices.PostgresPreparationTemplateClone
		}
		if reuseScope == suiteservices.FixtureReuseSuiteTemplate {
			kind, strategy = "template", suiteservices.PostgresPreparationTemplate
		}
		return h.recordEvent(suiteservices.Event{Type: suiteservices.EventPostgresDBCreated, Name: name, Kind: kind,
			Details: h.postgresPreparationDetails(strategy, template, reuseScope, attribution, time.Since(start))})
	})
	if !created && err == nil {
		err = errors.New("postgres creation returned no ownership proof")
	}
	if err != nil {
		h.quarantined.Store(true)
		if created {
			err = errors.Join(err, h.disposeDatabase(ctx, name, reuseScope, attribution))
		}
		return err
	}
	return nil
}

// Failed setup receives a fresh bounded cleanup lifetime; cancellation never
// discards confirmed ownership. The cleanup owner retains its tighter bound.
func (h *Harness) disposeDatabase(ctx context.Context, name, reuseScope string, attribution fixtureAttribution) error {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 35*time.Second)
	defer cancel()
	return h.dropDatabase(cleanup, name, reuseScope, attribution)
}

func (h *Harness) admitPreparation() error {
	if h.retired.Load() {
		return errors.New("postgres owner is closed")
	}
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

func createDatabase(ctx context.Context, adminDSN string, name string, templateDB string, confirmed func() error) (bool, error) {
	var err error
	created := false
	if templateDB != "" {
		created, err = cloneDatabase(ctx, adminDSN, name, templateDB, func(ctx context.Context, admin *sql.Conn, name string) error {
			if err := confirmed(); err != nil {
				return err
			}
			return finalizeClone(ctx, admin, name)
		})
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
				return confirmed()
			},
		)
	}
	if err == nil && templateDB == "" {
		err = PrepareFreshDatabase(ctx, adminDSN, name)
	}
	return created, err
}

// OpenPurposeDatabase opens a test-only database handle through the same
// purpose-role initializer used by production composition.
func OpenPurposeDatabase(ctx context.Context, dsn string, purpose postgres.Purpose) (*sql.DB, error) {
	if ctx == nil {
		return nil, errors.New("postgres purpose admission requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	contract, ok := testPurposeContracts[purpose]
	if !ok {
		return nil, errors.New("unsupported test postgres purpose")
	}
	selectedDSN, err := purposeDSN(dsn, purpose)
	if err != nil {
		return nil, err
	}
	db, err := postgres.OpenSQL(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          selectedDSN,
		Purpose:      purpose,
		ExpectedRole: contract.role,
	})
	if err != nil {
		return nil, errors.Join(err, ctx.Err())
	}
	return db, nil
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
		suiteHash = suiteservices.ShortHash("local-suite", 8)
	}
	return truncateIdentifier(fmt.Sprintf("ct_tpl_%s_%s", suiteHash, pgschema.ShortHash(TemplateDigest(schemaHash))), 63)
}

func (h *Harness) recordEvent(event suiteservices.Event) error {
	if h.events != nil {
		return h.events(event)
	}
	return suiteservices.RecordEvent(h.binding.environment(), event)
}

func (h *Harness) postgresPreparationDetails(strategy string, templateDB string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	details := h.postgresFixtureDetails(strategy, reuseScope, attribution, duration)
	details["preparation_strategy"] = strategy
	if templateDB != "" {
		details["template_database"] = templateDB
	}
	return details
}

func (h *Harness) postgresFixtureDetails(strategy string, reuseScope string, attribution fixtureAttribution, duration time.Duration) map[string]any {
	if reuseScope == "" {
		reuseScope = suiteservices.FixtureReusePerTest
	}
	details := map[string]any{
		"duration_ms":          duration.Milliseconds(),
		"reuse_scope":          reuseScope,
		"strategy":             strategy,
		"preparation_strategy": strategy,
		"target":               h.binding.inputs[suiteservices.TargetEnv],
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
	hash := h.schemaHash
	if hash == "" {
		hash = fixtureSchemaHash()
	}
	if hash != "" {
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

func resolvePostgresFixturePolicy(inputs map[string]string, attribution fixtureAttribution) string {
	topLevelTest := topLevelTestName(attribution.TestName)
	if policy := lookupFixturePolicy(inputs, postgresFixturePolicyTestsEnv, topLevelTest); policy != "" {
		return policy
	}
	if policy := lookupFixturePolicy(inputs, postgresFixturePolicyPackagesEnv, attribution.CallerPackage); policy != "" {
		return policy
	}
	if policy := normalizePostgresFixturePolicy(inputs[postgresFixturePolicyDefaultEnv]); policy != "" {
		return policy
	}
	return ""
}

// ExplicitPostgresFixturePolicyT returns the scheduler-assigned fixture policy
// for the current test, or an empty string when no explicit assignment exists.
func ExplicitPostgresFixturePolicyT(t testing.TB) string {
	t.Helper()

	return resolvePostgresFixturePolicy(processFixtureInputs(), fixtureAttributionFor(t, "pgtest"))
}

func (h *Harness) requireSelectedPostgresFixturePolicyT(t testing.TB, want string) {
	t.Helper()
	if err := h.admitPreparation(); err != nil {
		failDatabasePreparation(t, "postgres fixture admission", err)
	}

	got := resolvePostgresFixturePolicy(h.binding.environment(), fixtureAttributionFor(t, "pgtest"))
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

func lookupFixturePolicy(inputs map[string]string, envName string, key string) string {
	key = normalizeFixturePolicyKey(key)
	if key == "" {
		return ""
	}
	for _, assignment := range splitFixturePolicyAssignments(inputs[envName]) {
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
		if file != "" && (!strings.Contains(file, "/internal/testutil/"+harnessPackage+"/") || strings.HasSuffix(file, "_test.go")) && !strings.Contains(file, "/testing/") && !strings.Contains(file, "/src/runtime/") {
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
