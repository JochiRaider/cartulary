package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestMigrateRunnerAcceptsOnlyExplicitUp(t *testing.T) {
	gotApply := false

	runner := newTestMigrateRunner(t)
	runner.apply = func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
		gotApply = true
		return database_migrations.MigrationStatus{SourceName: source.Name}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if !gotApply {
		t.Fatal("migrate up did not invoke the typed apply operation")
	}
}

func TestMigrateRunnerRejectsUnsupportedGrammarBeforeConfiguration(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"up", "extra"},
		{"-command", "up"},
		{"up-to", "7"},
		{"down"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stderr := &bytes.Buffer{}
			runner := newTestMigrateRunner(t)
			runner.stderr = stderr
			runner.loadConfig = func() (configassembly.Loaded, error) {
				t.Fatal("invalid grammar loaded config")
				return configassembly.Loaded{}, nil
			}
			if exitCode := runner.runCLI(context.Background(), args); exitCode != 2 {
				t.Fatalf("exit code got %d want 2", exitCode)
			}
			if got := stderr.String(); got != "usage: migrate up\n" {
				t.Fatalf("usage got %q", got)
			}
		})
	}
}

func TestMigrateRunnerConfigLoadFailure(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr
	runner.loadConfig = func() (configassembly.Loaded, error) {
		return configassembly.Loaded{}, errors.New("config unavailable")
	}

	openCalled := false
	runner.openSQL = func(settings postgres.Settings) (*sql.DB, error) {
		openCalled = true
		return nil, nil
	}

	migrateCalled := false
	runner.apply = func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
		migrateCalled = true
		return database_migrations.MigrationStatus{}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if openCalled {
		t.Fatal("expected config failure to stop before opening postgres")
	}
	if migrateCalled {
		t.Fatal("expected config failure to stop before running migrations")
	}
	if output := stderr.String(); !strings.Contains(output, "load config: config unavailable") {
		t.Fatalf("expected config failure in stderr, got %q", output)
	}
}

func TestMigrateRunnerDBOpenFailure(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr

	migrateCalled := false
	runner.openSQL = func(settings postgres.Settings) (*sql.DB, error) {
		return nil, errors.New("dsn rejected")
	}
	runner.apply = func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
		migrateCalled = true
		return database_migrations.MigrationStatus{}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if migrateCalled {
		t.Fatal("expected db-open failure to stop before running migrations")
	}
	if output := stderr.String(); !strings.Contains(output, "open postgres: dsn rejected") {
		t.Fatalf("expected db-open failure in stderr, got %q", output)
	}
}

func TestMigrateRunnerPrintsMigrationRemediationReport(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr
	rawLineageID := "cartulary.prod_ddl_rebaseline.v1"
	runner.apply = func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
		return database_migrations.MigrationStatus{}, &database_migrations.MigrationRemediationError{
			Report: database_migrations.MigrationRemediationReport{
				SchemaID:    "cartulary.migration_remediation_report.v1",
				Boundary:    "prod_ddl_rebaseline_v1",
				FromVersion: 49,
				ToVersion:   23,
				Findings: []database_migrations.MigrationRemediationFinding{
					{
						Field:           "schema_migration_lineage",
						RawValue:        &rawLineageID,
						ReasonCode:      "historical_migration_lineage",
						RemediationHint: "reset or export/import",
					},
				},
			},
		}
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	output := stderr.String()
	required := []string{
		`"schema_id":"cartulary.migration_remediation_report.v1"`,
		`"boundary":"prod_ddl_rebaseline_v1"`,
		`"reason_code":"historical_migration_lineage"`,
		"migrate failed",
	}
	for _, needle := range required {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected stderr to contain %q, got %q", needle, output)
		}
	}
}

type migrateContextMarkerKey struct{}

func TestMigrateRunnerRunPassesContextToMigration(t *testing.T) {
	runner := newTestMigrateRunner(t)
	ctx := context.WithValue(context.Background(), migrateContextMarkerKey{}, "marker")

	var gotMarker any
	runner.apply = func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
		gotMarker = ctx.Value(migrateContextMarkerKey{})
		return database_migrations.MigrationStatus{SourceName: source.Name}, nil
	}

	if err := runner.run(ctx); err != nil {
		t.Fatalf("run migration: %v", err)
	}
	if gotMarker != "marker" {
		t.Fatalf("migration did not receive caller context marker: got %#v", gotMarker)
	}
}

func newTestMigrateRunner(t testing.TB) migrateRunner {
	t.Helper()
	roots := configtest.SetupTempRoots(t)
	configtest.BindPostgresDSNToDatabaseRoot(
		t,
		roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"],
		"postgres://unit-test",
	)
	loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)

	db, err := sql.Open("pgx", "postgres://cartulary:cartulary@127.0.0.1:1/cartulary?sslmode=disable")
	if err != nil {
		t.Fatalf("open test database handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return migrateRunner{
		stderr: bytes.NewBuffer(nil),
		loadConfig: func() (configassembly.Loaded, error) {
			return loaded, nil
		},
		openSQL: func(settings postgres.Settings) (*sql.DB, error) {
			return db, nil
		},
		apply: func(ctx context.Context, db *sql.DB, source database_migrations.MigrationSource) (database_migrations.MigrationStatus, error) {
			return database_migrations.MigrationStatus{SourceName: source.Name}, nil
		},
		source: database_migrations.MigrationSource{
			Path: "db/migrations",
			Name: "db/migrations",
		},
	}
}
