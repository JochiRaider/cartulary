package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	postgresFixturePolicyTemplateClone = "template_clone"
	postgresFixturePolicyPackageReset  = "package_reset"
)

const (
	postgresFixturePolicyTestsEnv    = "CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS"
	postgresFixturePolicyPackagesEnv = "CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES"
	postgresFixturePolicyDefaultEnv  = "CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT"
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
}

type TestDatabase struct {
	Name string
	DSN  string
}

type packageDatabase struct {
	mu       sync.Mutex
	db       *TestDatabase
	resetSQL string
}

type fixtureAttribution struct {
	TestName              string
	CallerFile            string
	CallerPackage         string
	PostgresFixturePolicy string
}

var (
	sharedHarnessMu     sync.Mutex
	sharedHarness       *Harness
	startOwnedHarness   = StartOwned
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
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "postgres",
			"POSTGRES_USER":     "cartulary",
			"POSTGRES_PASSWORD": "cartulary",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainersx.StartWithRetry(ctx, testcontainersx.StartConfig{
		Service: "postgres testcontainer",
		Image:   postgresImage,
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
		return nil, fmt.Errorf("resolve postgres host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
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

	if err := harness.WaitReady(ctx); err != nil {
		_ = harness.Close(ctx)
		return nil, fmt.Errorf("wait for postgres readiness: %w", err)
	}

	return harness, nil
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
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for {
		err := pingAdminDSNFn(ctx, h.adminDSN)
		if err == nil {
			return nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("postgres did not become ready: %w", lastErr)
		}
		time.Sleep(250 * time.Millisecond)
	}
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
	status, err := migrateDatabaseFn(db, dbmigrations.Source(), "up")
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

	attribution := fixtureAttributionFor(t, "pgtest")
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
	if _, err := migrateDatabaseFn(db, dbmigrations.Source(), command, args...); err != nil {
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

	attribution := fixtureAttributionFor(t, "pgtest")
	attribution.PostgresFixturePolicy = resolvePostgresFixturePolicy(attribution)
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

	statement, err := h.buildResetStatement(ctx, db)
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

	if fixture.resetSQL == "" {
		statement, err := h.buildResetStatement(ctx, db)
		if err != nil {
			return err
		}
		fixture.resetSQL = statement
	}
	if fixture.resetSQL != "" {
		if _, err := db.ExecContext(ctx, fixture.resetSQL); err != nil {
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

func (h *Harness) buildResetStatement(ctx context.Context, db *sql.DB) (string, error) {
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

func BeginRollbackTxT(t testing.TB, db *sql.DB) *sql.Tx {
	t.Helper()

	attribution := fixtureAttributionFor(t, "pgtest")
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
	if attribution.CallerPackage != "" {
		details["caller_package"] = attribution.CallerPackage
	}
	if attribution.PostgresFixturePolicy != "" {
		details["fixture_policy"] = attribution.PostgresFixturePolicy
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
	return postgresFixturePolicyPackageReset
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
	case postgresFixturePolicyPackageReset:
		return postgresFixturePolicyPackageReset
	default:
		return ""
	}
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

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
