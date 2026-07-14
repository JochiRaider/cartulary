package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestMigrateRunnerAcceptsOnlyExplicitUp(t *testing.T) {
	var gotCommand string

	runner := newTestMigrateRunner(t)
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotCommand = command
		if len(args) != 0 {
			t.Fatalf("migrate up received unexpected arguments: %#v", args)
		}
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if gotCommand != "up" {
		t.Fatalf("unexpected command: got %q want up", gotCommand)
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
			runner.loadConfig = func() (config.Config, error) {
				t.Fatal("invalid grammar loaded config")
				return config.Config{}, nil
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
	runner.loadConfig = func() (config.Config, error) {
		return config.Config{}, errors.New("config unavailable")
	}

	openCalled := false
	runner.openSQL = func(cfg config.Config) (*sql.DB, error) {
		openCalled = true
		return nil, nil
	}

	migrateCalled := false
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalled = true
		return postgres.MigrationStatus{}, nil
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
	runner.openSQL = func(cfg config.Config) (*sql.DB, error) {
		return nil, errors.New("dsn rejected")
	}
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalled = true
		return postgres.MigrationStatus{}, nil
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
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		return postgres.MigrationStatus{}, &postgres.MigrationRemediationError{
			Report: postgres.MigrationRemediationReport{
				SchemaID:    "cartulary.migration_remediation_report.v1",
				Boundary:    "prod_ddl_rebaseline_v1",
				FromVersion: 49,
				ToVersion:   23,
				Findings: []postgres.MigrationRemediationFinding{
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
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotMarker = ctx.Value(migrateContextMarkerKey{})
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
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

	db, err := sql.Open("pgx", "postgres://cartulary:cartulary@127.0.0.1:1/cartulary?sslmode=disable")
	if err != nil {
		t.Fatalf("open test database handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return migrateRunner{
		stderr: bytes.NewBuffer(nil),
		loadConfig: func() (config.Config, error) {
			return config.Config{}, nil
		},
		openSQL: func(cfg config.Config) (*sql.DB, error) {
			return db, nil
		},
		migrate: func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
			return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
		},
		source: postgres.MigrationSource{
			Path: "db/migrations",
			Name: "db/migrations",
		},
	}
}
