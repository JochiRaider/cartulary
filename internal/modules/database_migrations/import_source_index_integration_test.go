package database_migrations_test

import (
	"database/sql"
	"errors"
	"testing"

	platformpostgres "github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestImportSourceIndexPreservesAssociationsAndPrivileges_Integration(t *testing.T) {
	database := pgtest.Start(t).MigrationDatabaseThroughT(t, 48)
	db := database.SQL()
	// Synthetic structural rows exercise the populated increment, not document
	// acquisition or an alternative Imports application path.
	for _, statement := range []string{
		`INSERT INTO public.users(id,email,display_name,password_hash) VALUES ('00000000-0000-4000-8000-000000000001','index@example.test','Index fixture','unused')`,
		`INSERT INTO public.incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ('00000000-0000-4000-8000-000000000002','INDEX','index','Index fixture','active','00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000001')`,
		`INSERT INTO public.records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id)
VALUES ('00000000-0000-4000-8000-000000000003','00000000-0000-4000-8000-000000000002','timeline_event','00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000001')`,
		`INSERT INTO public.import_committed_source_blobs(content_sha256,byte_size) VALUES(encode(sha256(''::bytea),'hex'),0)`,
		`INSERT INTO public.import_committed_sources(committed_source_id,incident_id,content_sha256,source_file_kind,source_media_type,original_filename,parser_profile_id,parser_version,locator_kind,locator,source_rect_a1,header_row_ref,data_start_row_ref,approved_mapping_json,mapping_fingerprint,target_kind,target_view_schema_id,extension_profile_id,origin_import_session_id,origin_import_unit_id,created_by_user_id,committed_at)
VALUES('00000000-0000-4000-8000-000000000004','00000000-0000-4000-8000-000000000002',encode(sha256(''::bytea),'hex'),'csv','text/csv','fixture.csv','fixture','1','csv_file','fixture','A1:A2',1,2,'{}',repeat('a',64),'view_schema','cartulary.view.timeline.v2','','00000000-0000-4000-8000-000000000005','00000000-0000-4000-8000-000000000006','00000000-0000-4000-8000-000000000001','2026-01-01T00:00:00Z')`,
		`INSERT INTO public.import_committed_source_records(incident_id,committed_source_id,source_row_ref,record_id)
VALUES('00000000-0000-4000-8000-000000000002','00000000-0000-4000-8000-000000000004',2,'00000000-0000-4000-8000-000000000003')`,
	} {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	const fk = "public.import_committed_source_records.import_committed_source_records_source_fk"
	beforeCoverage := queryForeignKeyCoverage(t, db)
	if beforeCoverage[fk] {
		t.Fatal("predecessor unexpectedly covers source FK")
	}
	before := sourceAssociationState(t, db)
	if err := database.ApplyThrough(t.Context(), 49); err != nil {
		t.Fatal(err)
	}
	assertMigrationHead(t, db, 49)
	afterCoverage := queryForeignKeyCoverage(t, db)
	if len(beforeCoverage) != len(afterCoverage) || !afterCoverage[fk] || before != sourceAssociationState(t, db) {
		t.Fatal("index increment changed source data or failed to cover the existing FK")
	}
	// Down is tested only on this owned scratch database.
	if err := database.RollbackThrough(t.Context(), 48); err != nil {
		t.Fatal(err)
	}
	if queryForeignKeyCoverage(t, db)[fk] || before != sourceAssociationState(t, db) {
		t.Fatal("index Down changed data or retained index coverage")
	}
	if err := database.ApplyThrough(t.Context(), 49); err != nil {
		t.Fatal(err)
	}
	fixture, err := database.RuntimeDatabase()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := pgtest.OpenPurposeDatabase(t.Context(), fixture.DSN, platformpostgres.PurposeRuntime)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	recovery, err := pgtest.OpenPurposeDatabase(t.Context(), fixture.DSN, platformpostgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	defer recovery.Close()
	for _, statement := range []string{
		`UPDATE public.import_committed_source_records SET source_row_ref=3`,
		`DELETE FROM public.import_committed_source_records`,
		`TRUNCATE public.import_committed_source_records`,
	} {
		_, err := runtime.ExecContext(t.Context(), statement)
		var denied *pgconn.PgError
		if !errors.As(err, &denied) || denied.Code != "42501" {
			t.Fatalf("append-only operation was not denied: %v", err)
		}
	}
	if before != sourceAssociationState(t, runtime) {
		t.Fatal("denied operations changed provenance")
	}
	// Recovery may restore exact rows even though runtime cannot delete them.
	tx, err := recovery.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(t.Context(), `DELETE FROM public.import_committed_source_records`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), `INSERT INTO public.import_committed_source_records VALUES('00000000-0000-4000-8000-000000000002','00000000-0000-4000-8000-000000000004',2,'00000000-0000-4000-8000-000000000003')`); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if before != sourceAssociationState(t, recovery) {
		t.Fatal("Recovery changed exact association data")
	}
	if _, err := runtime.ExecContext(t.Context(), `INSERT INTO public.import_committed_source_records VALUES('00000000-0000-4000-8000-000000000002','00000000-0000-4000-8000-000000000004',3,'00000000-0000-4000-8000-000000000003')`); err != nil {
		t.Fatal("runtime append", err)
	}
	if _, err := runtime.ExecContext(t.Context(), `UPDATE public.records SET deleted_at=now(),deleted_by_user_id=created_by_user_id`); err != nil {
		t.Fatal(err)
	}
	assertAssociationCount(t, runtime, 2)
	if _, err := runtime.ExecContext(t.Context(), `DELETE FROM public.import_committed_sources`); err != nil {
		t.Fatal("parent cascade", err)
	}
	assertAssociationCount(t, runtime, 0)
}

func sourceAssociationState(t testing.TB, db *sql.DB) string {
	t.Helper()
	var value string
	if err := db.QueryRowContext(t.Context(), `SELECT jsonb_build_object(
'sources',(SELECT jsonb_agg(to_jsonb(s) ORDER BY committed_source_id) FROM public.import_committed_sources s),
'bindings',(SELECT jsonb_agg(to_jsonb(r) ORDER BY committed_source_id,source_row_ref) FROM public.import_committed_source_records r))::text`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertAssociationCount(t testing.TB, db *sql.DB, want int) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM public.import_committed_source_records`).Scan(&count); err != nil || count != want {
		t.Fatalf("association count=%d, want %d: %v", count, want, err)
	}
}

func TestForeignKeyCoverageUsesInstalledIndexKeys_Integration(t *testing.T) {
	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	for _, statement := range []string{
		`CREATE TABLE public.coverage_parent (a int, b int, PRIMARY KEY(a,b))`,
		`CREATE TABLE public.coverage_child (a int,b int,c int, CONSTRAINT coverage_fk FOREIGN KEY(a,b) REFERENCES public.coverage_parent(a,b))`,
	} {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		definition string
		covered    bool
	}{
		{"(a,b)", true}, {"(a,b,c)", true}, {"(b,a)", false}, {"(a,c,b)", false},
		{"(a) INCLUDE (b)", false}, {"(a,b) WHERE a IS NOT NULL", false}, {"((a+0),b)", false},
		{"(a,b) INCLUDE (c)", true}, {"(a,b,(c+1))", true},
	} {
		t.Run(tc.definition, func(t *testing.T) {
			if _, err := db.ExecContext(t.Context(), "CREATE INDEX coverage_candidate ON public.coverage_child "+tc.definition); err != nil {
				t.Fatal(err)
			}
			if got := queryForeignKeyCoverage(t, db)["public.coverage_child.coverage_fk"]; got != tc.covered {
				t.Errorf("coverage=%t, want %t", got, tc.covered)
			}
			if _, err := db.ExecContext(t.Context(), `DROP INDEX public.coverage_candidate`); err != nil {
				t.Fatal(err)
			}
			if queryForeignKeyCoverage(t, db)["public.coverage_child.coverage_fk"] {
				t.Fatal("dropped index still covers FK")
			}
		})
	}
	for _, definition := range []string{"(a,b) INCLUDE (c)", "(a,b,(c+1))"} {
		t.Run("drop_dependency_"+definition, func(t *testing.T) {
			if _, err := db.ExecContext(t.Context(), "CREATE INDEX coverage_candidate ON public.coverage_child "+definition); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(t.Context(), `ALTER TABLE public.coverage_child DROP COLUMN c`); err != nil {
				t.Fatal(err)
			}
			if queryForeignKeyCoverage(t, db)["public.coverage_child.coverage_fk"] {
				t.Fatal("index with dropped dependency still covers FK")
			}
			if _, err := db.ExecContext(t.Context(), `ALTER TABLE public.coverage_child ADD COLUMN c int`); err != nil {
				t.Fatal(err)
			}
		})
	}
}
