package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

// CollectUnreferencedObjects reclaims byte copies, never immutable logical
// identities, attempts, attestations, envelopes or removal history. A removal
// tombstone is durable collection work; no in-memory scheduling is required.
// Successful envelopes retain their exact signed containers for historical
// authentication. Extracted copies of removed versions have no consumer use.
func CollectUnreferencedObjects(ctx context.Context, db postgres.DB, storage CollectionStorage) error {
	if db == nil || storage == nil {
		return errors.New("reference pack: incomplete collection dependencies")
	}
	return storage.WithPublishedCollection(ctx, func(objects PublishedObjects) (resultErr error) {
		guard, err := acquireByteRetention(ctx, db, true)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, guard.Close()) }()
		// The guard precedes this transaction and outlives its commit. A failed or
		// ambiguous commit never authorizes deletion in this sweep; a later sweep
		// reads the authoritative result before retrying physical cleanup.
		if err := retireUnneededObjects(ctx, db); err != nil {
			return err
		}
		return objects.Visit(ctx, func(ref StorageRef) error {
			var retained bool
			if err := db.QueryRow(ctx, `SELECT
    EXISTS(SELECT 1 FROM reference_pack_objects o WHERE o.storage_ref=$1 AND (o.available OR EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id)))
    OR EXISTS(SELECT 1 FROM reference_pack_envelopes e WHERE e.container_ref=$1)
    OR EXISTS(SELECT 1 FROM reference_pack_operations p WHERE p.terminal_at IS NULL AND convert_from(p.frozen_input,'UTF8')::jsonb->>'container_reference'=$1)`, ref.String()).Scan(&retained); err != nil {
				return err
			}
			if retained {
				return nil
			}
			// No publisher can create a reference while this exclusive namespace and
			// database guard are held. Each publication uses a unique physical name.
			return objects.Remove(ref)
		})
	})
}

func retireUnneededObjects(ctx context.Context, db postgres.DB) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Same ordered key guards used by pin, remove and exact reimport. An import
	// that won the guard clears removal or installs pending work before this
	// recheck; an import that lost it writes entirely new immutable objects.
	rows, err := tx.Query(ctx, `SELECT k.pack_key FROM reference_pack_key_state k
 WHERE EXISTS(SELECT 1 FROM reference_pack_candidates c WHERE c.pack_key=k.pack_key AND c.removed)
 ORDER BY k.pack_key COLLATE "C" FOR UPDATE OF k`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM reference_pack_object_refs r USING reference_pack_candidates c
 WHERE r.owner_kind='version' AND r.owner_id='rpver_'||encode(sha256(convert_to(c.pack_key,'UTF8')||decode('00','hex')||convert_to(c.pack_version,'UTF8')),'hex')
 AND c.distribution_kind='operator_imported' AND c.removed AND c.health='missing' AND c.missing_reason='administrative_removal'
 AND NOT EXISTS(SELECT 1 FROM reference_pack_set_members m WHERE m.pack_key=c.pack_key AND m.pack_version=c.pack_version AND
  (EXISTS(SELECT 1 FROM reference_pack_current_set s WHERE s.pack_set_id=m.pack_set_id) OR EXISTS(SELECT 1 FROM reference_pack_pins p WHERE p.pack_set_id=m.pack_set_id)))
 AND NOT EXISTS(SELECT 1 FROM reference_pack_version_pins p WHERE p.pack_key=c.pack_key AND p.pack_version=c.pack_version)
 AND NOT EXISTS(SELECT 1 FROM reference_pack_operation_keys k JOIN reference_pack_operations p USING(operation_id) WHERE k.pack_key=c.pack_key AND p.terminal_at IS NULL)`); err != nil {
		return err
	}
	// Preparation owns bytes only while the exact operation can resume. Its
	// immutable input descriptors and attempt evidence remain after completion.
	// Envelope, version, release and backup references independently retain bytes.
	if _, err := tx.Exec(ctx, `DELETE FROM reference_pack_object_refs r USING reference_pack_operations p
 WHERE r.owner_kind='operation' AND r.owner_id=p.operation_id::text AND p.terminal_at IS NOT NULL`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_objects o SET available=false,generation=generation+1 WHERE o.available
 AND NOT EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id)
 AND NOT EXISTS(SELECT 1 FROM reference_pack_envelopes e WHERE e.container_ref=o.storage_ref)
 AND NOT EXISTS(SELECT 1 FROM reference_pack_operations p WHERE p.terminal_at IS NULL AND convert_from(p.frozen_input,'UTF8')::jsonb->>'container_reference'=o.storage_ref)
 AND NOT EXISTS(SELECT 1 FROM reference_pack_portable_selections s JOIN reference_pack_operations p USING(operation_id) WHERE s.input_object_id=o.object_id AND p.terminal_at IS NULL)`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
