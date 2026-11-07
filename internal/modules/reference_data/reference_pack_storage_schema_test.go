package reference_data_test

import (
	"context"
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
	var reference, path int
	if err := db.QueryRow(ctx, `SELECT count(*) FILTER(WHERE column_name='storage_ref'), count(*) FILTER(WHERE column_name IN ('storage_path','bundle_storage_path')) FROM information_schema.columns WHERE table_schema='public' AND table_name='reference_pack_objects'`).Scan(&reference, &path); err != nil || reference != 1 || path != 0 {
		t.Fatal("immutable storage must retain opaque references", reference, path, err)
	}
	if err := contractadmission.Validate(ctx, db); err != nil {
		t.Fatal("fresh admission", err)
	}
}
