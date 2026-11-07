// Package referenceadmissionfixture constructs deliberately incompatible owner
// records in isolated harness databases. It never changes production admission.
package referenceadmissionfixture

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/jackc/pgx/v5"
)

var Cases = []string{"current manifest extra", "current manifest missing", "current set unordered", "current set noncanonical", "current set identity", "entra manifest", "entra set", "entra portable catalog", "entra portable preparation", "entra snapshot", "entra render", "entra composition", "manifest", "set", "portable catalog", "portable preparation", "pending", "prepared", "snapshot", "render", "composition"}

const identity = "00000000-0000-4000-8000-000000000001"

var digest = strings.Repeat("a", 64)

func Seed(t testing.TB, db *sql.DB, kind string) {
	t.Helper()
	if strings.HasPrefix(kind, "current ") {
		seedMalformedCurrent(t, db, kind)
		return
	}
	if strings.HasPrefix(kind, "entra ") {
		seedEntra(t, db, strings.TrimPrefix(kind, "entra "))
		return
	}
	seed(t, db, kind, nil)
}

func SeedPortable(t testing.TB, db *sql.DB, catalog []byte) {
	t.Helper()
	canonical, err := canonicaljson.Canonicalize(catalog)
	if err != nil {
		t.Fatal(err)
	}
	seed(t, db, "portable catalog", canonical)
}

// SeedBindingQueryFailure leaves the global inspection queries readable and
// supplies a structurally valid Reporting binding. Only resolving that binding
// needs catalog ordering by incident, so this fault reaches the owner callback.
func SeedBindingQueryFailure(t testing.TB, db *sql.DB) {
	t.Helper()
	Seed(t, db, "snapshot")
	if _, err := db.ExecContext(t.Context(), `ALTER TABLE reference_pack_portable_catalogs RENAME COLUMN incident_id TO unavailable_binding_id`); err != nil {
		t.Fatal(err)
	}
}

// SeedMissingTable makes a required current table unavailable without deleting
// its retained contents. Runtime admission must report an operational failure.
func SeedMissingTable(t testing.TB, db *sql.DB, kind string) {
	t.Helper()
	var query string
	switch kind {
	case "reference table missing":
		query = `ALTER TABLE reference_pack_portable_catalogs RENAME TO unavailable_reference_catalogs`
	case "reporting table missing":
		query = `ALTER TABLE reporting_composition_preview_outputs RENAME TO unavailable_reporting_previews`
	default:
		t.Fatalf("unknown missing-table fixture %q", kind)
	}
	if _, err := db.ExecContext(t.Context(), query); err != nil {
		t.Fatal(err)
	}
}

func seed(t testing.TB, db *sql.DB, kind string, catalog []byte) {
	seedWith(t, db, kind, catalog, nil)
}

func seedWith(t testing.TB, db *sql.DB, kind string, catalog []byte, overrides map[string][]byte) {
	t.Helper()
	raw := func(key string, fallback []byte) []byte {
		if v, ok := overrides[key]; ok {
			return v
		}
		return fallback
	}

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
			t.Fatalf("seed %s: %v", kind, err)
		}
	}
	if kind == "pending" {
		exec(`INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,frozen_input) VALUES($1,'reconcile','application_release',now(),$2)`, identity, []byte(`{"schema_id":"cartulary.reference_pack_operation_input.v2"}`))
		return
	}
	exec(`INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES($1,'reconcile','application_release',now(),now(),'{}','{}')`, identity)
	switch kind {
	case "manifest":
		exec(`INSERT INTO reference_pack_key_state(pack_key) VALUES('framework.attack')`)
		exec(`INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at) VALUES('framework.attack','retained','operator_imported','staged',now())`)
		exec(`INSERT INTO reference_pack_versions(pack_key,pack_version,pack_release_sequence,manifest_sha256,payload_sha256,manifest_bytes) VALUES('framework.attack','retained',1,$1,$1,$2)`, digest, raw("manifest", []byte(`{"pack_contract_version":"cartulary.reference_pack_contract.v1"}`)))
	case "set":
		exec(`INSERT INTO reference_pack_sets(pack_set_id,pack_set_sha256,canonical_set,canonical_provenance,first_operation_id) VALUES($1,$2,$3,'[]',$4)`, "rpset_"+digest, digest, raw("set", []byte(`{"schema_id":"cartulary.reference_pack_set.v1"}`)), identity)
	case "portable preparation":
		exec(`INSERT INTO reference_pack_portable_preparations(operation_id,canonical_references,canonical_context) VALUES($1,$2,'{}')`, identity, raw("references", []byte(`{"schema_id":"reference_pack_refs.v1"}`)))
	case "prepared":
		exec(`INSERT INTO reference_pack_attempts(attempt_id,operation_id,started_at) VALUES($1,$1,now())`, identity)
		exec(`INSERT INTO reference_pack_attempt_members(attempt_id,ordinal,pack_key,pack_version,verdict,canonical_prepared) VALUES($1,1,'framework.attack','retained','succeeded',$2)`, identity, []byte(`{"schema_id":"retired"}`))
	default:
		exec(`INSERT INTO users(id,email,display_name,password_hash) VALUES($1,'admission@example.test','Admission fixture','unused')`, identity)
		exec(`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'ADMISSION','admission','Admission fixture','active',$1,$1)`, identity)
		if kind == "portable catalog" {
			raw := []byte(`{"schema_id":"reference_pack_refs.v1"}`)
			if catalog != nil {
				raw = catalog
			}
			exec(`INSERT INTO reference_pack_portable_catalogs(incident_id,operation_id,catalog_sha256,canonical_references,canonical_resolution) VALUES($1,$1,encode(sha256($2),'hex'),$2,'{}')`, identity, raw)
			return
		}
		exec(`INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,handler_name,job_kind,progress_unit_id,auth_policy) VALUES($1,'incident',$1,'queued',true,$1,now(),now(),0,'reporting.snapshot_create.v1','reporting.snapshot_create.v1','cartulary.reporting.snapshot.v1','incident_membership')`, identity)
		binding := fmt.Sprintf(`{"pack_set_id":"rpset_%s","pack_set_sha256":"%s","provenance":[]}`, digest, digest)
		binding = string(raw("binding", []byte(binding)))
		if kind == "snapshot" {
			admission := `{"schema_id":"cartulary.reporting_snapshot_admission.v1","reference_packs":` + binding + `}`
			exec(`INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,snapshot_admission_json,snapshot_admission_sha256) VALUES($1,'snapshot_create',$1,$1,$2,$3)`, identity, admission, digest)
			return
		}
		manifest := fmt.Sprintf(`{"schema_id":"cartulary.render_bundle_manifest.v3","manifest_version":"3","renderer_version":"21","admitted_export_model_sha256":"%s","export_model_sha256":"%s","reference_packs":%s}`, digest, digest, binding)
		if kind == "render" {
			exec(`INSERT INTO reporting_snapshots(snapshot_id,incident_id,created_by_user_id,client_txn_id,snapshot_at,source_change_set_high_watermark,source_boundary_json,derivation_version,source_capture_sha256,create_job_id) VALUES($1,$1,$1,'fixture',now(),'0','{}','1',$2,$1)`, identity, digest)
			exec(`INSERT INTO reporting_releases(release_id,incident_id,snapshot_id,created_by_user_id,client_txn_id,release_scope,release_state,snapshot_at,source_change_set_high_watermark,derivation_version,export_model_sha256,template_id,template_version,redaction_profile_id,redaction_profile_version,redaction_profile_sha256,output_kind,create_job_id,render_admitted_at) VALUES($1,$1,$1,$1,'fixture','internal_draft','render_failed',now(),'0','1',$2,'fixture','1','fixture','5',$2,'slidev',$1,now())`, identity, digest)
			exec(`INSERT INTO reporting_render_bundles(release_id,bundle_manifest_sha256,bundle_manifest_json,primary_bundle_path,primary_media_type) VALUES($1,$2,$3,'index.html','text/html')`, identity, digest, manifest)
			return
		}
		if kind != "composition" {
			t.Fatalf("unknown admission fixture %q", kind)
		}
		exec(`INSERT INTO report_compositions(composition_id,incident_id,created_by_user_id,client_txn_id,template_id,template_version) VALUES($1,$1,$1,'fixture','fixture','1')`, identity)
		exec(`INSERT INTO report_composition_preview_attempts(preview_attempt_id,incident_id,composition_id,source_kind,draft_version,preview_source_sha256,preview_source_json,snapshot_id,derivation_version,template_id,template_version,redaction_profile_id,redaction_profile_version,redaction_profile_sha256,render_environment_profile_id,output_kind,render_attempt_id,created_by_user_id,preview_source_bytes) VALUES($1,$1,$1,'draft',1,$2,'{}','fixture','1','fixture','1','fixture','5',$2,'fixture','slidev',$1,$1,'{}')`, identity, digest)
		exec(`INSERT INTO reporting_composition_preview_outputs(preview_attempt_id,render_attempt_id,output_kind,output_media_type,output_sha256,redaction_profile_id,redaction_profile_version,redaction_profile_sha256,redaction_manifest_sha256,redaction_manifest_json,bundle_manifest_sha256,bundle_manifest_json,primary_bundle_path,primary_media_type) VALUES($1,$1,'slidev','text/html',$2,'fixture','5',$2,$2,'{"schema_id":"cartulary.redaction_manifest.v5","container_decisions":[]}',$2,$3,'index.html','text/html')`, identity, digest, manifest)
	}
}

// State captures all persisted public-table rows and sequences, not only the
// deliberately incompatible row. Names originate exclusively in pg_catalog.
func State(t testing.TB, db *sql.DB) string {
	t.Helper()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	hash := sha256.New()
	for _, name := range names {
		var data string
		query := `SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM ` + pgx.Identifier{"public", name}.Sanitize() + ` t`
		if err := db.QueryRowContext(ctx, query).Scan(&data); err != nil {
			t.Fatal(name, err)
		}
		fmt.Fprintf(hash, "%s\x00%s\x00", name, data)
	}
	var sequences string
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(s) ORDER BY sequencename),'[]'::jsonb)::text FROM pg_sequences s WHERE schemaname='public'`).Scan(&sequences); err != nil {
		t.Fatal(err)
	}
	hash.Write([]byte(sequences))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Sentinels(t testing.TB, temporary, packs string) {
	t.Helper()
	for path, data := range map[string]string{filepath.Join(temporary, "reference-packs/verification", identity, strings.Repeat("a", 64)): "abandoned scratch", filepath.Join(packs, "incoming/retained.pack"): "incoming evidence", filepath.Join(packs, "reference-packs/objects", strings.Repeat("a", 64)+"-"+identity): "collectible object"} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func Files(t testing.TB, root string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%s\x00", relative, info.Mode())
		if entry.Type().IsRegular() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			hash.Write(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
