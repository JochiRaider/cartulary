package recoverycontribution

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func VNextRecoveryObjectInventory(
	source recovery.VNextObjectSource,
) recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.reference_data",
		"reference_packs.members",
		"reference_data.snapshot_member_inventory.v2",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT o.object_id::text, o.storage_ref, o.sha256, o.size_bytes
  FROM reference_pack_objects o
 WHERE EXISTS (SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id)
 OR EXISTS (SELECT 1 FROM reference_pack_envelopes e WHERE e.container_ref=o.storage_ref)
 ORDER BY o.object_id
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Reference Pack members: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var objectID, storageKey, digest string
				var size int64
				if err := rows.Scan(&objectID, &storageKey, &digest, &size); err != nil {
					return nil, fmt.Errorf("scan Reference Pack member: %w", err)
				}
				info, err := source.StatRecoveryObject(ctx, storageKey)
				if err != nil {
					return nil, fmt.Errorf("stat retained Reference Pack object: %w", err)
				}
				if info.PlaintextBytes != size {
					return nil, fmt.Errorf("retained Reference Pack object size mismatch")
				}
				logicalID := recovery.VNextLogicalObjectID("reference-pack", objectID)
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, "application/octet-stream", size, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Reference Pack members: %w", err)
			}
			return members, nil
		},
	)
}
