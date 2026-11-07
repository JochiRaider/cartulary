package reference_data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReferencePackStorageReferenceHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "reference-pack-storage-reference-head")
	ctx := context.Background()
	if _, err := db.Exec(ctx, `SET LOCAL ROLE cartulary_recovery`); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"reference_packs", "reference_pack_job_payloads", "reference_pack_attestations", "reference_pack_activation_state"} {
		var present bool
		if err := db.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+name).Scan(&present); err != nil || present {
			t.Fatal("retired mutable store remains", name, present, err)
		}
	}
	var reference, path int
	if err := db.QueryRow(ctx, `SELECT count(*) FILTER(WHERE column_name='storage_ref'), count(*) FILTER(WHERE column_name IN ('storage_path','bundle_storage_path')) FROM information_schema.columns WHERE table_schema='public' AND table_name='reference_pack_objects'`).Scan(&reference, &path); err != nil || reference != 1 || path != 0 {
		t.Fatal("immutable storage must retain opaque references", reference, path, err)
	}
	if err := contractadmission.Validate(ctx, db); err != nil {
		t.Fatal("fresh admission", err)
	}
	// Retired bytes remain opaque evidence. Admission must fail without touching
	// either the historical bytes or the migration ledger.
	old, err := os.ReadFile("../../../fixtures/reference-packs/historical-v1/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var refs struct {
		Sets []json.RawMessage `json:"sets"`
	}
	if err = json.Unmarshal(old, &refs); err != nil || len(refs.Sets) == 0 {
		t.Fatal("historical set fixture", err)
	}
	old = refs.Sets[0]
	var set struct {
		ID  string `json:"pack_set_id"`
		SHA string `json:"pack_set_sha256"`
	}
	if err = json.Unmarshal(old, &set); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES('00000000-0000-4000-8000-000000000001','reconcile','application_release',now(),now(),'{}','{}')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO reference_pack_sets(pack_set_id,pack_set_sha256,canonical_set,canonical_provenance,first_operation_id) VALUES($1,$2,$3,$4,'00000000-0000-4000-8000-000000000001')`, set.ID, set.SHA, old, []byte("[]")); err != nil {
		t.Fatal(err)
	}
	if err = contractadmission.Validate(ctx, db); err == nil {
		t.Fatal("retired set admitted")
	}
	var retained []byte
	if err = db.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets WHERE pack_set_id=$1`, set.ID).Scan(&retained); err != nil || !bytes.Equal(retained, old) {
		t.Fatal("admission changed historical bytes", err)
	}
}
