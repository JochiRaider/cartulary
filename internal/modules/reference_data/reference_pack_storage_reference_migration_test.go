package reference_data_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReferencePackStorageReferenceHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "reference-pack-storage-reference-head", postgres.PurposeRecovery)
	ctx := context.Background()
	for _, name := range []string{"reference_packs", "reference_pack_job_payloads", "reference_pack_attestations", "reference_pack_activation_state"} {
		var present bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+name).Scan(&present); err != nil || present {
			t.Fatal("retired mutable store remains", name, present, err)
		}
	}
	var reference, path int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE column_name='storage_ref'), count(*) FILTER(WHERE column_name IN ('storage_path','bundle_storage_path')) FROM information_schema.columns WHERE table_schema='public' AND table_name='reference_pack_objects'`).Scan(&reference, &path); err != nil || reference != 1 || path != 0 {
		t.Fatal("immutable storage must retain opaque references", reference, path, err)
	}
}
