package imports

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func VNextRecoveryObjectInventory() recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider(
		"module.imports",
		"imports.source_streams",
		"imports.snapshot_source_stream_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, sqlshape.Literal(`
SELECT source_stream_ref, source_media_type, source_byte_size,
       source_content_sha256, source_bytes
  FROM import_source_streams
 ORDER BY source_stream_ref ASC
`))
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

// VNextCommittedSourceInventory retains metadata only. Each object streams
// bounded chunks from the same caller-owned Recovery snapshot on demand.
func VNextCommittedSourceInventory() recovery.VNextObjectInventoryProvider {
	return recovery.NewVNextObjectInventoryProvider("module.imports", "imports.committed_sources", "imports.snapshot_committed_source_inventory.v1",
		func(ctx context.Context, snapshot recovery.VNextSnapshot) ([]recovery.VNextObjectMember, error) {
			rows, err := snapshot.QueryRows(ctx, sqlshape.Literal(`SELECT content_sha256,byte_size FROM import_committed_source_blobs ORDER BY content_sha256`))
			if err != nil {
				return nil, fmt.Errorf("inventory committed source blobs: %w", err)
			}
			defer rows.Close()
			members := []recovery.VNextObjectMember{}
			for rows.Next() {
				var digest string
				var size int64
				if err := rows.Scan(&digest, &size); err != nil {
					return nil, err
				}
				members = append(members, recovery.VNextObjectMember{LogicalObjectID: digest, StorageKey: "database-inline/import_committed_source_blobs/" + digest,
					ContentType: mediaTypeOctetStream, PlaintextBytes: size, PlaintextSHA256: digest,
					Open: func(readCtx context.Context) (io.ReadCloser, error) {
						reader := newSourceReader(readCtx, nil, ImportSourceCapability{SourceStreamRef: digest, SourceContentSHA256: digest, SourceByteSize: size}, false)
						reader.readChunk = func(offset, length int64) ([]byte, error) {
							chunks, err := snapshot.QueryRows(readCtx, sqlshape.Literal(committedBlobChunkQuery), digest, offset, length)
							if err != nil {
								return nil, err
							}
							defer chunks.Close()
							if !chunks.Next() {
								if err := chunks.Err(); err != nil {
									return nil, err
								}
								return nil, fmt.Errorf("committed source chunk unavailable")
							}
							var value []byte
							if err := chunks.Scan(&value); err != nil {
								return nil, err
							}
							if chunks.Next() {
								return nil, fmt.Errorf("committed source chunk is not unique")
							}
							return value, chunks.Err()
						}
						return reader, nil
					},
				})
			}
			return members, rows.Err()
		})
}
