package reporting_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReportingIdentityCutoverDoesNotReconstructHistory_Integration(t *testing.T) {
	h := pgtest.Start(t)
	migration := h.MigrationDatabaseThroughT(t, 62)
	ctx := context.Background()
	db := migration.SQL()
	if err := reporting.PreflightCutover(ctx, db); err != nil {
		t.Fatal("empty cutover", err)
	}
	_, err := db.ExecContext(ctx, `
 INSERT INTO users(id,email,display_name,password_hash) VALUES ('63000000-0000-4000-8000-000000000001','report-cutover@example.test','Cutover','not-used');
 INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
 VALUES ('63000000-0000-4000-8000-000000000002','REPORT-CUTOVER','report-cutover','Cutover','active','63000000-0000-4000-8000-000000000001','63000000-0000-4000-8000-000000000001');
 INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id)
 VALUES('63000000-0000-4000-8000-000000000003','incident','63000000-0000-4000-8000-000000000002','queued',true,'63000000-0000-4000-8000-000000000001',now(),now(),0,'incident_membership','snapshot_reporting.job_worker_v2','snapshot_reporting.snapshot_create_v2','snapshot_reporting.snapshot_create.materialization.v1');
 INSERT INTO reporting_snapshots(snapshot_id,incident_id,created_by_user_id,client_txn_id,snapshot_at,source_change_set_high_watermark,source_boundary_json,derivation_version,export_model_sha256,export_model_json,create_job_id)
 VALUES('63000000-0000-4000-8000-000000000004','63000000-0000-4000-8000-000000000002','63000000-0000-4000-8000-000000000001','cutover',now(),'boundary','{}','cartulary.reporting_derivation_profile.v2',repeat('a',64),'{}','63000000-0000-4000-8000-000000000003');`)
	if err != nil {
		t.Fatal(err)
	}
	var rejected *reporting.CutoverRequiredError
	if err := reporting.PreflightCutover(ctx, db); !errors.As(err, &rejected) {
		t.Fatal("retained history admitted", err)
	}
	if !strings.Contains(rejected.RemediationReportJSON(), `"category":"historical_snapshots","count":1`) || strings.Contains(rejected.RemediationReportJSON(), "REPORT-CUTOVER") {
		t.Fatal("unsafe cutover report")
	}
	if err := migration.ApplyThrough(ctx, 63); err == nil {
		t.Fatal("migration reconstructed retained history")
	}
	var count, head int
	var unchanged bool
	if err := db.QueryRowContext(ctx, `SELECT count(*),bool_and(export_model_json='{}'::jsonb) FROM reporting_snapshots`).Scan(&count, &unchanged); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&head); err != nil {
		t.Fatal(err)
	}
	if count != 1 || !unchanged || head != 62 {
		t.Fatal("preflight mutated history", count, unchanged, head)
	}
}
