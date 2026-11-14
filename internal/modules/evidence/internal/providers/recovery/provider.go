package recovery

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	recoverycontract "github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type provider struct {
	db dbport.DB
}

func New(db dbport.DB) recoverycontract.EvidenceRecoveryProvider {
	return &provider{db: db}
}

func (provider *provider) ListRequiredRecoveryObjects(ctx context.Context) ([]recoverycontract.EvidenceObjectState, error) {
	if provider == nil || nilDatabase(provider.db) {
		return nil, fmt.Errorf("evidence recovery provider requires database")
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
 WHERE b.upload_state IN ('available', 'quarantined')
 ORDER BY b.storage_key ASC, b.object_blob_id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list required Evidence recovery objects: %w", err)
	}
	defer rows.Close()

	objects := make([]recoverycontract.EvidenceObjectState, 0)
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
			return nil, fmt.Errorf("scan required Evidence recovery object: %w", err)
		}
		if !policy.ValidRecoveryBlobMetadata(byteSize, int64Pointer(observedSize), stringPointer(observedSHA), stringPointer(expectedSHA)) {
			return nil, fmt.Errorf("%w: invalid finalized Evidence object metadata", recoverycontract.ErrInvalidBackupArtifact)
		}
		objects = append(objects, recoverycontract.EvidenceObjectState{
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
		return nil, fmt.Errorf("iterate required Evidence recovery objects: %w", err)
	}
	return objects, nil
}

func nilDatabase(database dbport.DB) bool {
	if database == nil {
		return true
	}
	value := reflect.ValueOf(database)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
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
