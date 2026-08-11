package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestMigrateRunnerAcceptsOnlyExplicitUp(t *testing.T) {
	gotApply := false

	runner := newTestMigrateRunner(t)
	runner.apply = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		gotApply = true
		return nil
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
	runner.apply = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		migrateCalled = true
		return nil
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
	if output := stderr.String(); output != "migration_operation_failed\n" {
		t.Fatalf("unexpected safe config failure: %q", output)
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
	runner.apply = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		migrateCalled = true
		return nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if migrateCalled {
		t.Fatal("expected db-open failure to stop before running migrations")
	}
	if output := stderr.String(); output != "migration_operation_failed\n" {
		t.Fatalf("unexpected safe db-open failure: %q", output)
	}
}

func TestMigrateRunnerPrintsMigrationRemediationReport(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr
	want := `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"prod_ddl_rebaseline_v2","from_version":30,"to_version":29,"findings":[{"field":"schema_migration_lineage","raw_value":"cartulary.prod_ddl_rebaseline.v1","reason_code":"historical_migration_lineage","remediation_hint":"Destroy and recreate this database, then apply the Production DDL Rebaseline v2 catalog from version 1."}]}` + "\n"
	runner.apply = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		return fakeRemediationFailure{report: strings.TrimSuffix(want, "\n")}
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if output := stderr.String(); output != want {
		t.Fatalf("unexpected remediation stderr\n got: %q\nwant: %q", output, want)
	}
}

type fakeRemediationFailure struct {
	report string
}

func (failure fakeRemediationFailure) Error() string {
	return "private remediation transport"
}

func (failure fakeRemediationFailure) ReasonCode() string {
	return "historical_migration_lineage"
}

func (failure fakeRemediationFailure) RemediationReportJSON() string {
	return failure.report
}

type fakeMigrationFailure struct {
	reason string
}

func (failure fakeMigrationFailure) Error() string {
	return "vendor=postgres://secret@private-host query=SELECT_sensitive bind=password path=/private/migrations"
}

func (failure fakeMigrationFailure) ReasonCode() string {
	return failure.reason
}

func TestMigrateRunnerPrintsSafeMigrationReason(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr
	runner.apply = func(context.Context, *sql.DB, *database_migrations.Source) error {
		return fakeMigrationFailure{reason: "schema_migration_execution_failed"}
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if output := stderr.String(); output != "schema_migration_execution_failed\n" {
		t.Fatalf("unexpected safe migration stderr: %q", output)
	}
}

type migrateContextMarkerKey struct{}

func TestMigrateRunnerRunPassesContextToMigration(t *testing.T) {
	runner := newTestMigrateRunner(t)
	ctx := context.WithValue(context.Background(), migrateContextMarkerKey{}, "marker")

	var gotMarker any
	runner.apply = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		gotMarker = ctx.Value(migrateContextMarkerKey{})
		return nil
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
		postgres.PurposeMigration,
	)
	loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)

	db, err := sql.Open("pgx", "postgres://cartulary:cartulary@127.0.0.1:1/cartulary?sslmode=disable")
	if err != nil {
		t.Fatalf("open test database handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	source, err := database_migrations.BuildCanonicalEmbedded(
		fstest.MapFS{"00001_test.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")}},
		".",
	)
	if err != nil {
		t.Fatalf("construct test migration source: %v", err)
	}

	return migrateRunner{
		stderr: bytes.NewBuffer(nil),
		loadConfig: func() (configassembly.Loaded, error) {
			return loaded, nil
		},
		openSQL: func(settings postgres.Settings) (*sql.DB, error) {
			return db, nil
		},
		apply: func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
			return nil
		},
		source: func() (*database_migrations.Source, error) {
			return source, nil
		},
	}
}
