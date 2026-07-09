package app

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestMigrateRunnerParsesPositionalCommandAndArgs(t *testing.T) {
	var gotCommand string
	var gotArgs []string

	runner := newTestMigrateRunner(t)
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotCommand = command
		gotArgs = append([]string(nil), args...)
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"up-to", "5"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if gotCommand != "up-to" {
		t.Fatalf("unexpected command: got %q want %q", gotCommand, "up-to")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "5" {
		t.Fatalf("unexpected command args: got %#v want %#v", gotArgs, []string{"5"})
	}
}

func TestMigrateRunnerParsesCommandFlagAndArgs(t *testing.T) {
	var gotCommand string
	var gotArgs []string

	runner := newTestMigrateRunner(t)
	runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotCommand = command
		gotArgs = append([]string(nil), args...)
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	if exitCode := runner.runCLI(context.Background(), []string{"-command", "up-to", "7"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if gotCommand != "up-to" {
		t.Fatalf("unexpected command: got %q want %q", gotCommand, "up-to")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "7" {
		t.Fatalf("unexpected command args: got %#v want %#v", gotArgs, []string{"7"})
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

	if exitCode := runner.runCLI(context.Background(), nil); exitCode != 1 {
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

	if exitCode := runner.runCLI(context.Background(), nil); exitCode != 1 {
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

	if exitCode := runner.runCLI(context.Background(), nil); exitCode != 1 {
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

	if err := runner.run(ctx, "up", nil); err != nil {
		t.Fatalf("run migration: %v", err)
	}
	if gotMarker != "marker" {
		t.Fatalf("migration did not receive caller context marker: got %#v", gotMarker)
	}
}

func TestRunMigrateCLICompatibilityWrapperInvokesMigration(t *testing.T) {
	oldFactory := newMigrateRunnerForCLI
	t.Cleanup(func() {
		newMigrateRunnerForCLI = oldFactory
	})

	var gotCommand string
	var gotArgs []string
	migrateCalls := 0
	newMigrateRunnerForCLI = func(stderr io.Writer) migrateRunner {
		runner := newTestMigrateRunner(t)
		runner.stderr = stderr
		runner.migrate = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
			migrateCalls++
			gotCommand = command
			gotArgs = append([]string(nil), args...)
			return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
		}
		return runner
	}

	stderr := &bytes.Buffer{}
	if exitCode := RunMigrateCLI([]string{"-command", "up-to", "7"}, stderr); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0; stderr=%q", exitCode, stderr.String())
	}
	if migrateCalls != 1 {
		t.Fatalf("expected one migration call, got %d", migrateCalls)
	}
	if gotCommand != "up-to" {
		t.Fatalf("unexpected command: got %q want %q", gotCommand, "up-to")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "7" {
		t.Fatalf("unexpected command args: got %#v want %#v", gotArgs, []string{"7"})
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
