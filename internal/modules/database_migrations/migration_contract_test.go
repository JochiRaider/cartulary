package database_migrations

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

const validMigrationBody = "-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n"

func TestSourceConstructionRejectsInvalidCatalogs(t *testing.T) {
	valid := fstest.MapFS{"migrations/00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}}
	tests := []struct {
		name      string
		fsys      fs.FS
		root      string
		lineageID string
		boundary  string
	}{
		{name: "nil filesystem", root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "empty root", fsys: valid, lineageID: "lineage", boundary: "boundary"},
		{name: "absolute root", fsys: valid, root: "/migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "escaping root", fsys: valid, root: "../migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "missing lineage", fsys: valid, root: "migrations", boundary: "boundary"},
		{name: "missing boundary", fsys: valid, root: "migrations", lineageID: "lineage"},
		{name: "missing root", fsys: fstest.MapFS{}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "empty catalog", fsys: fstest.MapFS{"migrations": &fstest.MapFile{Mode: fs.ModeDir}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "unexpected entry", fsys: fstest.MapFS{"migrations/.keep": &fstest.MapFile{}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "unexpected directory", fsys: fstest.MapFS{"migrations/nested/00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "invalid filename", fsys: fstest.MapFS{"migrations/1_bad.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "duplicate version", fsys: fstest.MapFS{
			"migrations/00001_first.sql":  &fstest.MapFile{Data: []byte(validMigrationBody)},
			"migrations/00001_second.sql": &fstest.MapFile{Data: []byte(validMigrationBody)},
		}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "noncontiguous version", fsys: fstest.MapFS{"migrations/00002_second.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "missing down", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "missing up", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Down\nSELECT 1;\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "repeated up", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\n-- +goose Up\n-- +goose Down\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "repeated down", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\n-- +goose Down\n-- +goose Down\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "unsupported directive", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\n-- +goose Future\n-- +goose Down\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
		{name: "unbalanced block", fsys: fstest.MapFS{"migrations/00001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\n-- +goose StatementBegin\nSELECT 1;\n-- +goose Down\n")}}, root: "migrations", lineageID: "lineage", boundary: "boundary"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildSource(test.fsys, test.root, test.lineageID, test.boundary); err == nil {
				t.Fatal("expected source construction failure")
			}
			if test.lineageID != "" && test.boundary != "" {
				if _, err := BuildCanonicalEmbedded(test.fsys, test.root); err == nil {
					t.Fatal("expected canonical source construction failure")
				}
			}
		})
	}
}

func TestSourceSnapshotIsImmutable(t *testing.T) {
	input := fstest.MapFS{"migrations/00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}}
	source, err := buildSource(input, "migrations", "lineage", "boundary")
	if err != nil {
		t.Fatalf("construct source: %v", err)
	}
	input["migrations/00001_valid.sql"].Data = []byte("mutated")
	delete(input, "migrations/00001_valid.sql")

	inspection, err := InspectSource(source)
	if err != nil {
		t.Fatalf("inspect immutable source snapshot: %v", err)
	}
	wantDigest := "0ff0c7582e520452d3d905c2c5233fd9800f3cca64e60210e66d395314a637d7"
	if len(inspection.Entries) != 1 || inspection.Entries[0].Filename != "00001_valid.sql" || inspection.Entries[0].SHA256 != wantDigest {
		t.Fatalf("source snapshot changed after input mutation: %#v", inspection)
	}
	inspection.Entries[0].Filename = "mutated.sql"
	inspection.Entries = nil
	again, err := InspectSource(source)
	if err != nil {
		t.Fatalf("inspect source after projection mutation: %v", err)
	}
	if len(again.Entries) != 1 || again.Entries[0].Filename != "00001_valid.sql" {
		t.Fatalf("inspection mutation reached source state: %#v", again)
	}
}

func TestApplyRejectsInvalidInputsBeforeDatabaseAccess(t *testing.T) {
	requireMigrationFailureReason(t, Apply(context.Background(), nil, nil), reasonMigrationSourceInvalid)
	requireMigrationFailureReason(t, Apply(context.Background(), nil, &Source{}), reasonMigrationSourceInvalid)
	source, err := buildSource(
		fstest.MapFS{"00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}},
		".",
		"lineage",
		"boundary",
	)
	if err != nil {
		t.Fatalf("construct valid source: %v", err)
	}
	requireMigrationFailureReason(t, Apply(context.Background(), nil, source), reasonMigrationDatabaseUnavailable)
	//lint:ignore SA1012 This contract test intentionally exercises the public nil-context guard.
	requireMigrationFailureReason(t, Apply(nil, nil, source), reasonMigrationContextInvalid)
}

func requireMigrationFailureReason(t testing.TB, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected migration failure %q", want)
	}
	var failure MigrationFailure
	if !errors.As(err, &failure) {
		t.Fatalf("expected MigrationFailure, got %T: %v", err, err)
	}
	if failure.ReasonCode() != want || failure.Error() != want {
		t.Fatalf("migration failure = %q / %q, want %q", failure.ReasonCode(), failure.Error(), want)
	}
}

func TestSchemaReadinessRejectsInvalidSourceAndNilPool(t *testing.T) {
	requireMigrationFailureReason(t, EnsureSchemaReady(context.Background(), nil, nil), reasonMigrationSourceInvalid)
	requireMigrationFailureReason(t, EnsureSchemaReady(context.Background(), nil, &Source{}), reasonMigrationSourceInvalid)
	source, err := buildSource(
		fstest.MapFS{"00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}},
		".",
		"lineage",
		"boundary",
	)
	if err != nil {
		t.Fatalf("construct valid source: %v", err)
	}
	requireMigrationFailureReason(t, EnsureSchemaReady(context.Background(), (*pgxpool.Pool)(nil), source), reasonMigrationDatabaseUnavailable)
}

func TestMigrationRemediationReportContracts(t *testing.T) {
	source, err := buildSource(
		fstest.MapFS{"00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}},
		".",
		"cartulary.production.v1",
		"migration_lineage",
	)
	if err != nil {
		t.Fatalf("construct remediation source: %v", err)
	}
	tests := []struct {
		name  string
		state migrationLineageState
		want  string
	}{
		{
			name:  "missing",
			state: migrationLineageState{},
			want:  `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"migration_lineage","from_version":1,"to_version":1,"findings":[{"field":"schema_migration_lineage","raw_value_pair":{"current_version":1,"expected_lineage_id":"cartulary.production.v1","lineage_table_present":false,"observed_lineage_ids":[],"repository_head_version":1,"target_version":1},"reason_code":"historical_migration_lineage","remediation_hint":"Destroy and recreate this database, then apply the Production DDL Rebaseline v2 catalog from version 1."}]}`,
		},
		{
			name:  "wrong",
			state: migrationLineageState{TablePresent: true, ObservedIDs: []string{"cartulary.legacy.v1"}},
			want:  `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"migration_lineage","from_version":1,"to_version":1,"findings":[{"field":"schema_migration_lineage","raw_value":"cartulary.legacy.v1","raw_value_pair":{"current_version":1,"expected_lineage_id":"cartulary.production.v1","lineage_table_present":true,"observed_lineage_ids":["cartulary.legacy.v1"],"repository_head_version":1,"target_version":1},"reason_code":"historical_migration_lineage","remediation_hint":"Destroy and recreate this database, then apply the Production DDL Rebaseline v2 catalog from version 1."}]}`,
		},
		{
			name: "mixed",
			state: migrationLineageState{
				TablePresent: true,
				ObservedIDs:  []string{"cartulary.legacy.v1", "cartulary.production.v1"},
			},
			want: `{"schema_id":"cartulary.migration_remediation_report.v1","boundary":"migration_lineage","from_version":1,"to_version":1,"findings":[{"field":"schema_migration_lineage","raw_value":"cartulary.legacy.v1","raw_value_pair":{"current_version":1,"expected_lineage_id":"cartulary.production.v1","lineage_table_present":true,"observed_lineage_ids":["cartulary.legacy.v1","cartulary.production.v1"],"repository_head_version":1,"target_version":1},"reason_code":"historical_migration_lineage","remediation_hint":"Destroy and recreate this database, then apply the Production DDL Rebaseline v2 catalog from version 1."}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := newMigrationLineageRemediationError(source, test.state, 1, 1, 1)
			var reporter RemediationReporter
			if !errors.As(report, &reporter) {
				t.Fatalf("expected remediation reporter, got %T", report)
			}
			if got := reporter.RemediationReportJSON(); got != test.want {
				t.Fatalf("remediation JSON mismatch\n got: %s\nwant: %s", got, test.want)
			}
			if reporter.Error() != test.want || reporter.ReasonCode() != reasonHistoricalMigrationLineage {
				t.Fatalf("remediation error mismatch: %s", report.Error())
			}
		})
	}
}
