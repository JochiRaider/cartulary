package database_migrations_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReportingFailureRecordMigration_Integration(t *testing.T) {
	for _, state := range []string{"empty", "pending", "retained"} {
		t.Run(state, func(t *testing.T) {
			database := pgtest.Start(t).MigrationDatabaseThroughT(t, 36)
			db := database.SQL()
			if state != "empty" {
				_, err := db.ExecContext(t.Context(), `INSERT INTO application_crypto_format(singleton,format_id) VALUES(true,'cartulary.application_crypto_format.v1') ON CONFLICT DO NOTHING;
INSERT INTO users(id,email,display_name,password_hash) VALUES('00000000-0000-0000-0000-000000000001','failure-cutover@example.test','Cutover fixture','unused');
INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES('00000000-0000-0000-0000-000000000002','cutover','cutover','Cutover fixture','active','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001');
INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id) VALUES('00000000-0000-0000-0000-000000000003','incident','00000000-0000-0000-0000-000000000002','queued',true,'00000000-0000-0000-0000-000000000001',now(),now(),0,'incident_membership','cutover_fixture','cutover_fixture','test.cutover_fixture.operation.v1');`)
				if err != nil {
					t.Fatal(err)
				}
				digest := strings.Repeat("a", 64)
				// Storage-valid private input; the migration must never interpret
				// or convert its retired diagnostic bytes, even if malformed.
				recipe, err := json.Marshal(map[string]any{
					"schema_id": "cartulary.reporting_render_admission.v1", "preparation_algorithm": "materialize_retained_reporting_render_v1",
					"source_capture_sha256": digest, "authored_redaction_profile_sha256": digest, "release_scope": "internal_draft",
					"authored_redaction_profile": map[string]any{"schema_id": "cartulary.redaction_profile.v3", "version": "5", "allowed_release_scopes": []string{"internal_draft"}},
					"template_contract":          map[string]any{"schema_id": "cartulary.reporting_template_pack_manifest.v3"}, "template_sha256": digest,
					"graph_sources_sha256": digest, "graph_sources": map[string]any{"schema_id": "cartulary.reporting_graph_input.v1", "results": []any{}},
					"identity": map[string]any{"release_id": "00000000-0000-0000-0000-000000000004", "preview_attempt_id": nil},
				})
				if err != nil {
					t.Fatal(err)
				}
				var diagnostic []byte
				var diagnosticSHA *string
				if state == "retained" {
					diagnostic = []byte("migration-diagnostic-private-sentinel")
					diagnosticSHA = &digest
				}
				_, err = db.ExecContext(t.Context(), `INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,request_json,admitted_profile_sha256,admitted_graph_sources_sha256,admitted_template_manifest_sha256,render_admission_sha256,render_failure_summary,render_failure_sha256) VALUES('00000000-0000-0000-0000-000000000003','release_create','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001',$1,$2,$2,$2,$2,$3,$4)`, recipe, digest, diagnostic, diagnosticSHA)
				if err != nil {
					t.Fatal(err)
				}
			}
			var beforeLedger, beforeRows string
			if err := db.QueryRowContext(t.Context(), `SELECT (SELECT jsonb_agg(to_jsonb(v) ORDER BY id)::text FROM goose_db_version v),(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY job_id),'[]')::text FROM reporting_job_payloads p)`).Scan(&beforeLedger, &beforeRows); err != nil {
				t.Fatal(err)
			}
			err := database.ApplyThrough(t.Context(), 37)
			var present bool
			var retiredColumns int
			if scanErr := db.QueryRowContext(t.Context(), `SELECT to_regclass('public.reporting_render_failure_records') IS NOT NULL,(SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='reporting_job_payloads' AND column_name IN ('render_failure_summary','render_failure_sha256'))`).Scan(&present, &retiredColumns); scanErr != nil {
				t.Fatal(scanErr)
			}
			if state == "retained" {
				if err == nil || !strings.Contains(err.Error(), "reporting failure journal cutover requires empty retired diagnostics") || strings.Contains(err.Error(), "migration-diagnostic-private-sentinel") || present || retiredColumns != 2 {
					t.Fatal("retained history did not reject safely", present, retiredColumns, err)
				}
				var afterLedger, afterRows string
				if err := db.QueryRowContext(t.Context(), `SELECT (SELECT jsonb_agg(to_jsonb(v) ORDER BY id)::text FROM goose_db_version v),(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY job_id),'[]')::text FROM reporting_job_payloads p)`).Scan(&afterLedger, &afterRows); err != nil {
					t.Fatal(err)
				}
				if beforeLedger != afterLedger || beforeRows != afterRows {
					t.Fatal("rejected cutover changed ledger or retained bytes")
				}
				return
			}
			if err != nil || !present || retiredColumns != 0 {
				t.Fatal("current-only cutover failed", present, retiredColumns, err)
			}
			var records, jobs, inputs int
			if err := db.QueryRowContext(t.Context(), `SELECT (SELECT count(*) FROM reporting_render_failure_records),(SELECT count(*) FROM jobs),(SELECT count(*) FROM reporting_job_payloads)`).Scan(&records, &jobs, &inputs); err != nil {
				t.Fatal(err)
			}
			want := 0
			if state == "pending" {
				want = 1
			}
			if records != 0 || jobs != want || inputs != want {
				t.Fatal("cutover fabricated evidence or removed current input", records, jobs, inputs)
			}
			if state == "pending" {
				var before []map[string]any
				if err := json.Unmarshal([]byte(beforeRows), &before); err != nil {
					t.Fatal(err)
				}
				delete(before[0], "render_failure_summary")
				delete(before[0], "render_failure_sha256")
				expected, err := json.Marshal(before)
				if err != nil {
					t.Fatal(err)
				}
				var equal bool
				if err := db.QueryRowContext(t.Context(), `SELECT jsonb_agg(to_jsonb(p) ORDER BY job_id)=$1::jsonb FROM reporting_job_payloads p`, expected).Scan(&equal); err != nil || !equal {
					t.Fatal("pending input changed during cutover", err)
				}
			}
		})
	}
}
