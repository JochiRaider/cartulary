package reporting

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func VNextRecoveryObjectInventory(
	source recovery.VNextObjectSource,
) recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.reporting",
		"reporting.render_preview_members",
		"reporting.snapshot_output_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT logical_id, object_ref, media_type, size_bytes, file_sha256
  FROM (
        SELECT 'render:' || release_id::text || ':' || bundle_path AS logical_id,
               object_ref, media_type, size_bytes, file_sha256
          FROM reporting_render_bundle_files
         WHERE storage_kind = 'object_store'
        UNION ALL
        SELECT 'preview:' || preview_attempt_id::text || ':' || bundle_path AS logical_id,
               object_ref, media_type, size_bytes, file_sha256
          FROM reporting_composition_preview_output_files
         WHERE storage_kind = 'object_store'
       ) members
 ORDER BY logical_id ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Reporting output members: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var rawLogicalID, storageKey, contentType, digest string
				var size int64
				if err := rows.Scan(&rawLogicalID, &storageKey, &contentType, &size, &digest); err != nil {
					return nil, fmt.Errorf("scan Reporting output member: %w", err)
				}
				logicalID := recovery.VNextLogicalObjectID("reporting", rawLogicalID)
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, contentType, size, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Reporting output members: %w", err)
			}
			return members, nil
		},
	)
}
