package reference_data_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func testRetiredReferenceTableCutover(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	for _, test := range []struct{ name, table, category, sql string }{
		{"activation history", "reference_pack_activation_state", "legacy_pack_activations", `INSERT INTO reference_pack_activation_state(pack_key) VALUES('retired.activation')`},
		{"pack history", "reference_packs", "legacy_packs", `INSERT INTO reference_packs(pack_key,version,pack_kind,manifest_sha256,payload_sha256,pack_contract_version,verification_method,imported_at,bundle_sha256,bundle_storage_ref) VALUES('retired.pack','1','registry',repeat('a',64),repeat('b',64),'reference_pack.v1','manifest_sha256_v1',now(),repeat('c',64),'reference-packs/retained.bundle')`},
		{"audit history", "reference_pack_attestations", "legacy_pack_attestations", `INSERT INTO reference_pack_attestations(pack_key,pack_version,pack_kind,event_kind,manifest_sha256,payload_sha256,verification_method,verification_result,occurred_at) VALUES('retired.pack','1','registry','import',repeat('a',64),repeat('b',64),'manifest_sha256_v1','failed',now())`},
		{"job input", "reference_pack_job_payloads", "legacy_pack_jobs", `INSERT INTO reference_pack_job_payloads(job_id,job_kind,actor_user_id,created_at) VALUES('58000000-0000-4000-8000-000000000001','import','58000000-0000-4000-8000-000000000002',now())`},
	} {
		t.Run(test.name, func(t *testing.T) {
			scratch := harness.MigrationDatabaseThroughT(t, 57)
			db := scratch.SQL()
			if err := reference_data.PreflightCutover(ctx, db); err != nil {
				t.Fatal("empty retired stores rejected", err)
			}
			if test.table == "reference_pack_attestations" {
				if _, err := db.ExecContext(ctx, `INSERT INTO reference_packs(pack_key,version,pack_kind,manifest_sha256,payload_sha256,pack_contract_version,verification_method,imported_at,bundle_sha256,bundle_storage_ref) VALUES('retired.pack','1','registry',repeat('a',64),repeat('b',64),'reference_pack.v1','manifest_sha256_v1',now(),repeat('c',64),'reference-packs/retained.bundle')`); err != nil {
					t.Fatal(err)
				}
			}
			if test.table == "reference_pack_job_payloads" {
				if _, err := db.ExecContext(ctx, `INSERT INTO users(id,email,display_name,password_hash) VALUES('58000000-0000-4000-8000-000000000002','retired-job@example.test','Retained actor','not-used');
INSERT INTO jobs(job_id,scope_kind,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id) VALUES('58000000-0000-4000-8000-000000000001','deployment','queued',true,'58000000-0000-4000-8000-000000000002',now(),now(),0,'deployment_admin','retired_worker','reference_pack.import_v1','reference_pack.import.request.v1')`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := db.ExecContext(ctx, test.sql); err != nil {
				t.Fatal(err)
			}
			var required *reference_data.CutoverRequiredError
			if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
				t.Fatal("retired rows passed preflight", err)
			}
			report := required.RemediationReportJSON()
			if !strings.Contains(report, `"category":"`+test.category+`","count":1`) || strings.Contains(report, "retired.pack") || strings.Contains(report, "retained.bundle") {
				t.Fatal("unsafe retirement report", report)
			}
			if err := scratch.ApplyThrough(ctx, 58); err == nil {
				t.Fatal("retained legacy state silently discarded")
			}
			var count, head int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.`+test.table).Scan(&count); err != nil || count != 1 {
				t.Fatal("retained row changed", count, err)
			}
			if err := db.QueryRowContext(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&head); err != nil || head != 57 {
				t.Fatal("failed migration changed schema head", head, err)
			}
			for _, table := range []string{"reference_packs", "reference_pack_activation_state", "reference_pack_attestations", "reference_pack_job_payloads"} {
				var present bool
				if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&present); err != nil || !present {
					t.Fatal("failed migration partially dropped stores", table, err)
				}
			}
		})
	}
}
