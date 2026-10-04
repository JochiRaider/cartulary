package reference_data

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testRemovedObjectCollection(t *testing.T, c *Coordinator, storage *coordinatorMemoryStorage, key, version string) {
	t.Helper()
	ctx := context.Background()
	var before int
	if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_object_refs WHERE owner_kind='version' AND owner_id=$1`, versionObjectID(key, version)).Scan(&before); err != nil || before == 0 {
		t.Fatal("missing extraction fixture", before, err)
	}
	backup, err := AcquireBackupRetention(ctx, c.pool)
	if err != nil {
		t.Fatal(err)
	}
	// The backup guard must stop retirement itself, not just the final unlink.
	deadline, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	err = CollectUnreferencedObjects(deadline, c.pool, storage)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("collection bypassed backup retention", err)
	}
	var after int
	if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_object_refs WHERE owner_kind='version' AND owner_id=$1`, versionObjectID(key, version)).Scan(&after); err != nil || after != before {
		t.Fatal("backup-selected references retired", after, err)
	}
	if err := backup.Close(); err != nil {
		t.Fatal(err)
	}
	if err := backup.Close(); err != nil {
		t.Fatal("retention close not idempotent", err)
	}
	if err := CollectUnreferencedObjects(ctx, c.pool, storage); err != nil {
		t.Fatal(err)
	}
	if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_object_refs WHERE owner_kind='version' AND owner_id=$1`, versionObjectID(key, version)).Scan(&after); err != nil || after != 0 {
		t.Fatal("removed extraction retained", after, err)
	}
	var retained, deleted int
	if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_objects WHERE NOT available`).Scan(&deleted); err != nil || deleted == 0 {
		t.Fatal("nothing collected", deleted, err)
	}
	if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_envelopes e JOIN reference_pack_objects o ON o.storage_ref=e.container_ref WHERE e.pack_key=$1 AND e.pack_version=$2 AND o.available`, key, version).Scan(&retained); err != nil || retained == 0 {
		t.Fatal("historical containers lost", retained, err)
	}
	rows, err := c.pool.Query(ctx, `SELECT storage_ref,available FROM reference_pack_objects`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var ref string
		var available bool
		if err := rows.Scan(&ref, &available); err != nil {
			t.Fatal(err)
		}
		_, exists := storage.objects[ref]
		if exists != available {
			t.Fatal("physical collection differs from authoritative retirement", ref, exists, available)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := RestoreHistoricalState(ctx, c.pool, storage, DefaultLimits()); err != nil {
		t.Fatal("removed history failed full authentication", err)
	}
	var index *string
	if err := c.pool.QueryRow(ctx, `SELECT current_index_id::text FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, key, version).Scan(&index); err != nil || index != nil {
		t.Fatal("restore recreated removed index", index, err)
	}
	if err := CollectUnreferencedObjects(ctx, c.pool, storage); err != nil {
		t.Fatal("repeat collection", err)
	}
}
