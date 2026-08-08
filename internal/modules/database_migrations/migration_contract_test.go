package database_migrations

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrationSourceDefaultsAndEmptyBehavior(t *testing.T) {
	status, err := Apply(context.Background(), nil, MigrationSource{
		BaseFS: fstest.MapFS{
			".gitkeep": &fstest.MapFile{Data: nil, Mode: 0o600},
		},
	})
	if err != nil {
		t.Fatalf("migrate empty default source: %v", err)
	}
	if status.SourceName != "." || !status.Empty {
		t.Fatalf("unexpected default empty-source status: %#v", status)
	}

	named := NewEmbeddedMigrationSource(fstest.MapFS{}, ".", "embedded migrations")
	status, err = Apply(context.Background(), nil, named)
	if err != nil {
		t.Fatalf("migrate named empty source: %v", err)
	}
	if status.SourceName != "embedded migrations" || !status.Empty {
		t.Fatalf("unexpected named empty-source status: %#v", status)
	}
}

func TestMigrationOperationRejectsNilContextBeforeInspection(t *testing.T) {
	//lint:ignore SA1012 This contract test intentionally exercises the public nil-context guard.
	status, err := Apply(nil, nil, NewMigrationSource("/path/that/must/not/be-inspected"))
	if !errors.Is(err, errNilMigrateContext) {
		t.Fatalf("expected nil-context error, got status=%#v err=%v", status, err)
	}
	if status.SourceName != "/path/that/must/not/be-inspected" || status.Empty {
		t.Fatalf("unexpected nil-context status: %#v", status)
	}
}

func TestGooseLoggerConfigurationLifecycle(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "private", "goose.log")
	logger, closer, err := newGooseLogger(logPath)
	if err != nil {
		t.Fatalf("configure goose logger: %v", err)
	}
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat goose log: %v", err)
	}
	if got := info.Mode().Perm(); got != fs.FileMode(0o600) {
		t.Fatalf("goose log mode = %#o, want 0600", got)
	}
	logger.Printf("retained")
	if err := closer.Close(); err != nil {
		t.Fatalf("close goose logger: %v", err)
	}

	directoryInfo, err := os.Stat(filepath.Dir(logPath))
	if err != nil {
		t.Fatalf("stat goose log directory: %v", err)
	}
	if got := directoryInfo.Mode().Perm(); got != fs.FileMode(0o700) {
		t.Fatalf("goose log directory mode = %#o, want 0700", got)
	}
	_, closer, err = newGooseLogger(logPath)
	if err != nil {
		t.Fatalf("reopen goose logger: %v", err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("close reopened goose logger: %v", err)
	}
	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read closed goose log: %v", err)
	}
	if !strings.Contains(string(payload), "retained") {
		t.Fatalf("goose logger did not append retained content: %q", payload)
	}
}

func TestSchemaReadinessHandlesNilPoolAndEmptySource(t *testing.T) {
	if err := EnsureSchemaReady(context.Background(), nil, NewMigrationSource("/missing")); err != nil {
		t.Fatalf("nil pool should be readiness-neutral: %v", err)
	}

	emptySource := NewEmbeddedMigrationSource(fstest.MapFS{}, ".", "empty")
	if err := EnsureSchemaReady(context.Background(), &pgxpool.Pool{}, emptySource); err != nil {
		t.Fatalf("empty source should be readiness-neutral: %v", err)
	}
}

func TestMigrationRemediationReportJSONContract(t *testing.T) {
	raw := "cartulary.legacy_line.v1"
	err := &MigrationRemediationError{Report: MigrationRemediationReport{
		SchemaID:    "cartulary.migration_remediation_report.v1",
		Boundary:    "migration_lineage",
		FromVersion: 1,
		ToVersion:   2,
		Findings: []MigrationRemediationFinding{{
			Field:        "schema_migration_lineage",
			RawValue:     &raw,
			RawValuePair: map[string]any{"expected_lineage_id": "cartulary.production.v1", "lineage_table_present": true},
			ReasonCode:   "historical_migration_lineage",
			RemediationHint: "Reset this database, or move data through an explicit owner-approved export/import path before applying " +
				"the production DDL rebaseline.",
		}},
	}}

	want := `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"migration_lineage","from_version":1,"to_version":2,"findings":[{"field":"schema_migration_lineage","raw_value":"cartulary.legacy_line.v1","raw_value_pair":{"expected_lineage_id":"cartulary.production.v1","lineage_table_present":true},"reason_code":"historical_migration_lineage","remediation_hint":"Reset this database, or move data through an explicit owner-approved export/import path before applying the production DDL rebaseline."}]}`
	if got := err.ReportJSON(); got != want {
		t.Fatalf("remediation JSON mismatch\n got: %s\nwant: %s", got, want)
	}
	if err.Error() != want {
		t.Fatalf("remediation error mismatch: %s", err.Error())
	}
}
