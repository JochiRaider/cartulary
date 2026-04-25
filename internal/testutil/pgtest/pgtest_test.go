package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	resetSharedHarness(t)

	t.Setenv(suiteservices.PGAdminDSNEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable")
	t.Setenv(suiteservices.PGTemplateDBEnv, "suite_template")

	oldStart := startOwnedHarness
	oldPing := pingAdminDSNFn
	t.Cleanup(func() {
		startOwnedHarness = oldStart
		pingAdminDSNFn = oldPing
	})

	startCalls := 0
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return nil, errors.New("owned harness should not start when attach env is present")
	}

	var pingDSN string
	pingAdminDSNFn = func(ctx context.Context, dsn string) error {
		pingDSN = dsn
		return nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start attached harness: %v", err)
	}
	if startCalls != 0 {
		t.Fatalf("expected attach mode to skip owned startup, got %d calls", startCalls)
	}
	if !harness.attached {
		t.Fatal("expected attached harness")
	}
	if harness.Container != nil {
		t.Fatal("expected attach mode to avoid creating a testcontainer")
	}
	if harness.templateDB != "suite_template" {
		t.Fatalf("unexpected template database: got %q", harness.templateDB)
	}
	if pingDSN != harness.AdminDSN() {
		t.Fatalf("unexpected admin dsn ping: got %q want %q", pingDSN, harness.AdminDSN())
	}
	if got := harness.dsnFor("tenant_one"); !strings.Contains(got, "/tenant_one?") {
		t.Fatalf("expected attach dsn template replacement, got %q", got)
	}
}

func TestStartSharedFallsBackToOwnedHarnessWhenAttachEnvIsAbsent(t *testing.T) {
	resetSharedHarness(t)
	t.Setenv(suiteservices.PGAdminDSNEnv, "")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "")
	t.Setenv(suiteservices.PGTemplateDBEnv, "")

	oldStart := startOwnedHarness
	t.Cleanup(func() {
		startOwnedHarness = oldStart
	})

	startCalls := 0
	want := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		suiteHash:   "suitehash",
		processHash: "processid",
	}
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return want, nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start shared harness: %v", err)
	}
	if startCalls != 1 {
		t.Fatalf("expected one owned startup, got %d", startCalls)
	}
	if harness != want {
		t.Fatal("expected StartShared to return the owned harness")
	}
	if !harness.shared {
		t.Fatal("expected owned harness returned by StartShared to be marked shared")
	}
}

func TestDatabaseNamesAreUniqueAcrossSimulatedProcesses(t *testing.T) {
	first := &Harness{suiteHash: "suitehash", processHash: "procaaaa"}
	second := &Harness{suiteHash: "suitehash", processHash: "procbbbb"}

	seen := make(map[string]struct{})
	validName := regexp.MustCompile(`^[a-z0-9_]+$`)
	for _, harness := range []*Harness{first, second} {
		for i := 0; i < 8; i++ {
			name := harness.nextDatabaseName("Prefix With Spaces-and-symbols!")
			if len(name) > 63 {
				t.Fatalf("database name exceeds postgres identifier limit: %q", name)
			}
			if !validName.MatchString(name) {
				t.Fatalf("database name must be sanitized, got %q", name)
			}
			if _, exists := seen[name]; exists {
				t.Fatalf("database name collision detected: %q", name)
			}
			seen[name] = struct{}{}
		}
	}
}

func TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "suite-template-mode")
	t.Setenv(suiteservices.TargetEnv, "backend-store")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "template-mode")

	oldCreate := createDatabaseFn
	oldMigrate := migrateDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		migrateDatabaseFn = oldMigrate
	})

	var createCalls []struct {
		Name     string
		Template string
	}
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		createCalls = append(createCalls, struct {
			Name     string
			Template string
		}{
			Name:     name,
			Template: templateDB,
		})
		return nil
	}

	migrateCalls := 0
	migrateDatabaseFn = func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalls++
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	testDB, status, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database from template: %v", err)
	}
	if len(createCalls) != 1 {
		t.Fatalf("expected one database create call, got %d", len(createCalls))
	}
	if createCalls[0].Template != "suite_template" {
		t.Fatalf("expected template clone create call, got template %q", createCalls[0].Template)
	}
	if migrateCalls != 0 {
		t.Fatalf("expected template mode to skip per-database migrations, got %d calls", migrateCalls)
	}
	if !status.TemplateClone {
		t.Fatal("expected migration status to mark template clone mode")
	}
	if status.Directory != dbmigrations.RepositoryPath {
		t.Fatalf("unexpected migration directory in status: got %q want %q", status.Directory, dbmigrations.RepositoryPath)
	}
	if status.TemplateDatabase != "suite_template" {
		t.Fatalf("unexpected template database in status: got %q", status.TemplateDatabase)
	}
	if testDB.Name == "" || !strings.Contains(testDB.DSN, testDB.Name) {
		t.Fatalf("expected prepared database dsn to reference the cloned database, got %#v", testDB)
	}

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	if len(scope.Postgres.DatabasePreparations) != 1 {
		t.Fatalf("expected one database preparation, got %#v", scope.Postgres.DatabasePreparations)
	}
	preparation := scope.Postgres.DatabasePreparations[0]
	if preparation.Name != testDB.Name {
		t.Fatalf("unexpected prepared database name: got %q want %q", preparation.Name, testDB.Name)
	}
	if preparation.Strategy != suiteservices.PostgresPreparationTemplateClone {
		t.Fatalf("expected template-clone preparation, got %#v", preparation)
	}
	if preparation.TemplateDatabase != "suite_template" {
		t.Fatalf("unexpected template database: got %#v", preparation)
	}
	if preparation.Target != "backend-store" {
		t.Fatalf("unexpected preparation target: got %#v", preparation)
	}
}

func TestHarnessStartsPostgresAndRunsCurrentMigrationPath(t *testing.T) {
	harness := Start(t)

	testDB, status, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	defer func() {
		if err := harness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop database: %v", err)
		}
	}()

	if testDB.DSN == "" {
		t.Fatal("expected database dsn")
	}
	if status.Empty {
		t.Fatal("expected the current bootstrap migration set to include numbered migrations")
	}
}

func TestPrepareDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareDatabaseT(t, "bootstrap_t")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if !exists {
		t.Fatal("expected PrepareDatabaseT to return a migrated database")
	}
}

func TestPreparePackageDatabaseTReusesAndResetsMutableTables(t *testing.T) {
	harness := Start(t)

	var firstName string
	t.Run("first use seeds rows", func(t *testing.T) {
		first := harness.PreparePackageDatabaseT(t, "package-reset")
		firstName = first.Name

		db, err := sql.Open("pgx", first.DSN)
		if err != nil {
			t.Fatalf("open package database: %v", err)
		}
		defer db.Close()

		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "package-reset@example.test", "Package Reset", "hash"); err != nil {
			t.Fatalf("seed package database: %v", err)
		}
	})
	t.Run("second use resets rows", func(t *testing.T) {
		second := harness.PreparePackageDatabaseT(t, "package-reset")
		if second.Name != firstName {
			t.Fatalf("expected package database reuse, got %q want %q", second.Name, firstName)
		}
		db, err := sql.Open("pgx", second.DSN)
		if err != nil {
			t.Fatalf("open reused package database: %v", err)
		}
		defer db.Close()

		var userCount int
		if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
			t.Fatalf("count reset users: %v", err)
		}
		if userCount != 0 {
			t.Fatalf("expected package reset to clear users, got %d", userCount)
		}

		var gooseCount int
		if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM goose_db_version`).Scan(&gooseCount); err != nil {
			t.Fatalf("count goose versions: %v", err)
		}
		if gooseCount == 0 {
			t.Fatal("expected package reset to preserve migration metadata")
		}
	})
}

func resetSharedHarness(t testing.TB) {
	t.Helper()

	sharedHarnessMu.Lock()
	sharedHarness = nil
	sharedHarnessMu.Unlock()

	t.Cleanup(func() {
		sharedHarnessMu.Lock()
		sharedHarness = nil
		sharedHarnessMu.Unlock()
	})
}
