package incidentbundles

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func VNextRecoveryObjectInventory(
	source recovery.VNextObjectSource,
) recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.incidentbundles",
		"incident_bundles.files",
		"incidentbundles.snapshot_file_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT bundle_id::text, bundle_storage_ref, bundle_byte_size, bundle_sha256
  FROM incident_bundle_exports
 ORDER BY bundle_id ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Incident Bundle files: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var logicalID, storageKey, digest string
				var size int64
				if err := rows.Scan(&logicalID, &storageKey, &size, &digest); err != nil {
					return nil, fmt.Errorf("scan Incident Bundle file: %w", err)
				}
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, "application/zip", size, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Incident Bundle files: %w", err)
			}
			return members, nil
		},
	)
}
