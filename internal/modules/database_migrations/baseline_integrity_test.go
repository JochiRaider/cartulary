package database_migrations_test

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestConsolidatedBaselineIntegrity_Integration(t *testing.T) {
	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	ctx := t.Context()
	var usable bool
	if err := db.QueryRowContext(ctx, `SELECT
 current_setting('server_version_num')='180006' AND current_setting('data_checksums')='on'
 AND NOT has_sequence_privilege('cartulary_runtime','public.goose_db_version_id_seq','USAGE,SELECT,UPDATE')
 AND NOT has_sequence_privilege('cartulary_recovery','public.goose_db_version_id_seq','USAGE,SELECT,UPDATE')
 AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_constraint k JOIN pg_catalog.pg_namespace n ON n.oid=k.connamespace WHERE n.nspname='public' AND (NOT k.convalidated OR NOT k.conenforced))
 AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_index i JOIN pg_catalog.pg_class c ON c.oid=i.indrelid JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND (NOT i.indisvalid OR NOT i.indisready))`).Scan(&usable); err != nil || !usable {
		t.Fatal("baseline engine/integrity state", usable, err)
	}
	actor, incident, other, src, dst := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO users(id,email,display_name,password_hash) VALUES($1,'baseline@example.test','Baseline','unused')`, []any{actor}},
		{`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'BASE','base','Baseline','active',$3,$3),($2,'OTHER','other','Other','active',$3,$3)`, []any{incident, other, actor}},
		{`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$3,'evidence',$4,$4),($2,$3,'artifact',$4,$4)`, []any{src, dst, incident, actor}},
	} {
		if _, err := db.ExecContext(ctx, q.sql, q.args...); err != nil {
			t.Fatal(err)
		}
	}
	insertLink := `INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,provenance,owner_user_id,created_by_user_id,field_key) VALUES($1,$2,$3,'references_record','manual',$4,$4,$5)`
	requireCode := func(err error, code string) {
		t.Helper()
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) || pgerr.Code != code {
			t.Fatalf("got %v, want SQLSTATE %s", err, code)
		}
	}
	// Deployment-scoped imports have no incident key. Their payload still must
	// reference an existing job, and imported artifacts may bind that operation.
	jobID := uuid.New()
	insertImportPayload := `INSERT INTO public.incident_bundle_job_payloads(job_id,job_kind,actor_user_id,uploaded_sha256,request_json,created_at) VALUES($1,'import',$2,repeat('a',64),'{}',now())`
	_, err := db.ExecContext(ctx, insertImportPayload, jobID, actor)
	requireCode(err, "23503")
	if _, err := db.ExecContext(ctx, `INSERT INTO public.jobs(job_id,scope_kind,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,handler_name,job_kind,progress_unit_id) VALUES($1,'deployment','queued',true,$2,now(),now(),0,'baseline','baseline','cartulary.baseline.progress.v1')`, jobID, actor); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, insertImportPayload, jobID, actor); err != nil {
		t.Fatal("deployment import payload", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO public.reporting_imported_artifact_files(incident_id,bundle_path,operation_id,size_bytes,sha256,content) VALUES($1,'ext/snapshots/baseline.json',$2,2,encode(sha256('{}'::bytea),'hex'),'{}'::bytea)`, incident, jobID); err != nil {
		t.Fatal("deployment import artifact", err)
	}
	// The nullable exception must not admit an export bound to a deployment job.
	_, err = db.ExecContext(ctx, `UPDATE public.incident_bundle_job_payloads SET job_kind='export',incident_id=$1,uploaded_sha256=NULL WHERE job_id=$2`, incident, jobID)
	requireCode(err, "23503")

	for _, field := range []any{nil, "field"} {
		if _, err := db.ExecContext(ctx, insertLink, incident, src, dst, actor, field); err != nil {
			t.Fatal(err)
		}
		_, err := db.ExecContext(ctx, insertLink, incident, src, dst, actor, field)
		requireCode(err, "23505")
	}
	if _, err := db.ExecContext(ctx, `UPDATE record_links SET deleted_at=now(),deleted_by_user_id=$1`, actor); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, insertLink, incident, src, dst, actor, nil); err != nil {
		t.Fatal("soft deletion did not release identity", err)
	}
	_, err = db.ExecContext(ctx, insertLink, other, src, dst, actor, nil)
	requireCode(err, "23503")
	// A globally valid blob cannot be attached through a different incident.
	blob := uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,byte_size,target_expires_at,pending_expires_at) VALUES($1,$2,$3,'baseline',1,now()+interval '1 hour',now()+interval '1 day')`, blob, other, actor); err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO evidence(record_id,incident_id,object_blob_id) VALUES($1,$2,$3)`, src, incident, blob)
	requireCode(err, "23503")
	// A partially absent composition tuple used to yield UNKNOWN and pass CHECK.
	_, err = db.ExecContext(ctx, `INSERT INTO reporting_releases(incident_id,snapshot_id,created_by_user_id,client_txn_id,release_scope,release_state,snapshot_at,source_change_set_high_watermark,derivation_version,export_model_sha256,template_id,template_version,redaction_profile_id,redaction_profile_version,redaction_profile_sha256,output_kind,create_job_id,render_admitted_at,composition_id,composition_version,composition_sha256)
 VALUES($1,$2,$3,'baseline','internal','draft',now(),'boundary','profile',repeat('a',64),'template','v1','profile','v1',repeat('b',64),'pdf',$4,now(),$5,NULL,NULL)`, incident, uuid.New(), actor, uuid.New(), uuid.New())
	requireCode(err, "23514")
	var pgerr *pgconn.PgError
	if !errors.As(err, &pgerr) || pgerr.ConstraintName != "reporting_releases_composition_tuple_ck" {
		t.Fatal("wrong rejection", err)
	}
	// Sequence allocation is intentionally nontransactional: rollback leaves a gap.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var allocated, next int64
	if err := tx.QueryRowContext(ctx, `SELECT nextval('public.record_revisions_revision_id_seq')`).Scan(&allocated); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT nextval('public.record_revisions_revision_id_seq')`).Scan(&next); err != nil || next <= allocated {
		t.Fatal("sequence rollback semantics", allocated, next, err)
	}
}
