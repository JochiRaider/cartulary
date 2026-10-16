package evidence_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEvidenceVerificationCutoverRejectsLegacyHistoryWithoutMutation_Integration(t *testing.T) {
	for _, terminal := range []bool{false, true} {
		name := "pending_attempts"
		if terminal {
			name = "exhausted_attempts"
		}
		t.Run(name, func(t *testing.T) {
			migration := pgtest.Start(t).MigrationDatabaseThroughT(t, 94)
			db, ctx := migration.SQL(), context.Background()
			_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,display_name,password_hash) VALUES ('95000000-0000-4000-8000-000000000001','verification-cutover@example.test','Cutover','not-used');
INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ('95000000-0000-4000-8000-000000000002','VERIFY-CUTOVER','verify-cutover','Cutover','active','95000000-0000-4000-8000-000000000001','95000000-0000-4000-8000-000000000001');
INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,byte_size,target_expires_at,pending_expires_at,finalize_attempt_count)
VALUES ('95000000-0000-4000-8000-000000000003','95000000-0000-4000-8000-000000000002','95000000-0000-4000-8000-000000000001','cutover-observation',3,now()+interval '1 hour',now()+interval '1 day',3);`)
			if err != nil {
				t.Fatal(err)
			}
			if terminal {
				_, err = db.ExecContext(ctx, `UPDATE object_blobs SET upload_state='failed',terminal_reason='finalize_retry_exhausted',failed_at=now(),finalize_attempt_count=4`)
				if err != nil {
					t.Fatal(err)
				}
			}
			var before, after string
			snapshot := `SELECT jsonb_build_object('blobs',(SELECT jsonb_agg(to_jsonb(b)) FROM object_blobs b),'migrations',(SELECT jsonb_agg(to_jsonb(g) ORDER BY id) FROM goose_db_version g))::text`
			if err = db.QueryRowContext(ctx, snapshot).Scan(&before); err != nil {
				t.Fatal(err)
			}
			err = migration.ApplyThrough(ctx, 95)
			if err == nil || !strings.Contains(err.Error(), "Evidence verification requires fresh initialization") {
				t.Fatalf("expected legacy history rejection, got %v", err)
			}
			if err = db.QueryRowContext(ctx, snapshot).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatal("rejected cutover changed data or migration state")
			}
			var absent bool
			if err = db.QueryRowContext(ctx, `SELECT to_regclass('public.evidence_blob_verifications') IS NULL`).Scan(&absent); err != nil {
				t.Fatal(err)
			}
			if !absent {
				t.Fatal("rejected cutover published verification relation")
			}
		})
	}
}
