package recoveryprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Provider struct {
	db postgres.DB
}

func New(db postgres.DB) *Provider {
	return &Provider{db: db}
}

func (provider *Provider) ListAvailableRecoveryObjects(ctx context.Context) ([]recovery.EvidenceObjectState, error) {
	if provider == nil || provider.db == nil {
		return nil, fmt.Errorf("Evidence recovery provider requires database")
	}
	rows, err := provider.db.Query(ctx, `
SELECT b.object_blob_id,
       b.storage_key,
       b.byte_size,
       b.observed_size,
       b.expected_sha256_hex,
       b.observed_sha256_hex,
       e.blob_hash
  FROM object_blobs b
  LEFT JOIN evidence e
    ON e.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'available'
 ORDER BY b.storage_key ASC, b.object_blob_id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list available Evidence recovery objects: %w", err)
	}
	defer rows.Close()

	objects := make([]recovery.EvidenceObjectState, 0)
	for rows.Next() {
		var objectBlobID uuid.UUID
		var storageKey string
		var byteSize int64
		var observedSize pgtype.Int8
		var expectedSHA pgtype.Text
		var observedSHA pgtype.Text
		var blobHash pgtype.Text
		if err := rows.Scan(
			&objectBlobID,
			&storageKey,
			&byteSize,
			&observedSize,
			&expectedSHA,
			&observedSHA,
			&blobHash,
		); err != nil {
			return nil, fmt.Errorf("scan available Evidence recovery object: %w", err)
		}
		objects = append(objects, recovery.EvidenceObjectState{
			ObjectBlobID:      objectBlobID,
			StorageKey:        storageKey,
			ByteSize:          byteSize,
			ObservedSize:      int64Pointer(observedSize),
			ExpectedSHA256Hex: stringPointer(expectedSHA),
			ObservedSHA256Hex: stringPointer(observedSHA),
			BlobHash:          stringPointer(blobHash),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available Evidence recovery objects: %w", err)
	}
	return objects, nil
}

func (provider *Provider) CountRecoveryRows(ctx context.Context) (int64, error) {
	if provider == nil || provider.db == nil {
		return 0, fmt.Errorf("Evidence recovery provider requires database")
	}
	var count int64
	if err := provider.db.QueryRow(ctx, `SELECT COUNT(*) FROM object_blobs`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count Evidence recovery rows: %w", err)
	}
	return count, nil
}

func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	out := value.Int64
	return &out
}

func stringPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	out := value.String
	return &out
}
