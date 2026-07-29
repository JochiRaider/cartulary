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
		"module.extensions",
		"extensions.staged_objects",
		"extensions.snapshot_staged_object_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT staging_id, storage_identity, expected_byte_size, expected_sha256
  FROM extension_staged_objects
 WHERE state IN ('ready', 'published')
   AND delete_state <> 'deleted'
 ORDER BY staging_id ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Extension staged objects: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var logicalID, storageKey, digest string
				var size int64
				if err := rows.Scan(&logicalID, &storageKey, &size, &digest); err != nil {
					return nil, fmt.Errorf("scan Extension staged object: %w", err)
				}
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, "application/octet-stream", size, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Extension staged objects: %w", err)
			}
			return members, nil
		},
	)
}
