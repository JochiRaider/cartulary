package imports

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func VNextRecoveryObjectInventory() recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.imports",
		"imports.source_streams",
		"imports.snapshot_source_stream_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, `
SELECT source_stream_ref, source_media_type, source_byte_size,
       source_content_sha256, source_bytes
  FROM import_source_streams
 ORDER BY source_stream_ref ASC
`)
			if err != nil {
				return nil, fmt.Errorf("inventory Import source streams: %w", err)
			}
			defer rows.Close()
			var members []recovery.VNextObjectMember
			for rows.Next() {
				var logicalID, contentType, digest string
				var size int64
				var body []byte
				if err := rows.Scan(&logicalID, &contentType, &size, &digest, &body); err != nil {
					return nil, fmt.Errorf("scan Import source stream: %w", err)
				}
				immutable := append([]byte(nil), body...)
				members = append(members, recovery.VNextObjectMember{
					LogicalObjectID: logicalID,
					StorageKey:      "database-inline/import_source_streams/" + logicalID,
					ContentType:     contentType,
					PlaintextBytes:  size,
					PlaintextSHA256: digest,
					Open: func(context.Context) (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader(immutable)), nil
					},
				})
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate Import source streams: %w", err)
			}
			return members, nil
		},
	)
}
