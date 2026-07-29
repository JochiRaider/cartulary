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
		"reference_data.snapshot_member_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT pack_key, version, bundle_storage_ref, bundle_sha256
  FROM reference_packs
 WHERE status IN ('staged', 'available', 'disabled')
 ORDER BY pack_key ASC, version ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Reference Pack members: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var packKey, version, storageKey, digest string
				if err := rows.Scan(&packKey, &version, &storageKey, &digest); err != nil {
					return nil, fmt.Errorf("scan Reference Pack member: %w", err)
				}
				info, err := source.StatRecoveryObject(ctx, storageKey)
				if err != nil {
					return nil, fmt.Errorf("stat Reference Pack member %s/%s: %w", packKey, version, err)
				}
				logicalID := recovery.VNextLogicalObjectID("reference-pack", packKey, version)
				members = append(members, recovery.VNextStoredObjectMember(
					source, logicalID, storageKey, "application/zip", info.PlaintextBytes, digest,
				))
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Reference Pack members: %w", err)
			}
			return members, nil
		},
	)
}
