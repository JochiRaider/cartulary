package evidence

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func VNextRecoveryObjectInventory(
	source recovery.VNextObjectSource,
) recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.evidence",
		"evidence.blobs",
		"evidence.snapshot_blob_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT object_blob_id::text, storage_key,
       COALESCE(observed_content_type, content_type_hint, 'application/octet-stream'),
       observed_size, observed_sha256_hex
  FROM object_blobs
 WHERE upload_state = 'available'
   AND observed_size IS NOT NULL
   AND observed_sha256_hex IS NOT NULL
 ORDER BY object_blob_id ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Evidence recovery objects: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var logicalID, storageKey, contentType, digest string
				var size int64
				if err := rows.Scan(&logicalID, &storageKey, &contentType, &size, &digest); err != nil {
					return nil, fmt.Errorf("scan Evidence recovery object: %w", err)
				}
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, contentType, size, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Evidence recovery objects: %w", err)
			}
			return members, nil
		},
	)
}
