package recovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// EvidenceObjectState is the Recovery-owned projection of one available
// Evidence blob. Evidence remains responsible for realizing this projection
// from its private persistence model.
type EvidenceObjectState struct {
	ObjectBlobID      uuid.UUID
	StorageKey        string
	ByteSize          int64
	ObservedSize      *int64
	ExpectedSHA256Hex *string
	ObservedSHA256Hex *string
	BlobHash          *string
}

type EvidenceRecoveryProvider interface {
	ListAvailableRecoveryObjects(context.Context) ([]EvidenceObjectState, error)
	CountRecoveryRows(context.Context) (int64, error)
}

func AvailableBlobObjectIDsByStorageRef(
	ctx context.Context,
	provider EvidenceRecoveryProvider,
) (map[string]uuid.UUID, error) {
	if provider == nil {
		return nil, fmt.Errorf("%w: Evidence recovery provider is required for object-store backup manifest blob index", ErrInvalidBackupArtifact)
	}
	objects, err := provider.ListAvailableRecoveryObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list available Evidence recovery objects: %w", err)
	}
	ids := make(map[string]uuid.UUID, len(objects))
	for _, object := range objects {
		storageKey := strings.TrimSpace(object.StorageKey)
		if storageKey == "" || object.ObjectBlobID == uuid.Nil {
			return nil, fmt.Errorf("%w: Evidence recovery object identity is incomplete", ErrInvalidBackupArtifact)
		}
		if _, exists := ids[storageKey]; exists {
			return nil, fmt.Errorf("%w: duplicate durable object blob storage_ref %q", ErrInvalidBackupArtifact, storageKey)
		}
		ids[storageKey] = object.ObjectBlobID
	}
	return ids, nil
}
