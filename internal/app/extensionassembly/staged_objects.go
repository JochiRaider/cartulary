package extensionassembly

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

// StagedObjectRepository converts the physical extensionstore model into the
// stagedobjects owner's logical lifecycle port.
type StagedObjectRepository struct {
	store *extensionstore.Store
}

func NewStagedObjectRepository(store *extensionstore.Store) (*StagedObjectRepository, error) {
	if store == nil {
		return nil, errors.New("staged-object repository requires extensionstore")
	}
	return &StagedObjectRepository{store: store}, nil
}

func (r *StagedObjectRepository) Allocate(ctx context.Context, allocation stagedobjects.Allocation) error {
	if r == nil || r.store == nil ||
		allocation.OperationID == "" ||
		!allocation.ExpiresAt.Equal(allocation.StagedAt.Add(stagedobjects.StagingLifetime)) {
		return stagedobjects.NewFailure(stagedobjects.FailureIntegrity, "invalid_allocation", nil)
	}
	object := extensionstore.NewStagedObject(
		allocation.StagingID,
		allocation.OperationID,
		allocation.OwnerProfileID,
		allocation.StorageIdentity,
		allocation.StagedAt,
	)
	if allocation.Resource != nil {
		object.ReservedResourceKind = &allocation.Resource.Kind
		object.ReservedResourceID = &allocation.Resource.ID
	}
	err := r.store.AllocateStagedObject(ctx, object)
	return classifyStagedRepositoryError(err)
}

func (r *StagedObjectRepository) Seal(ctx context.Context, operationID, profileID string, sealed stagedobjects.SealedReference, now time.Time) error {
	return classifyStagedRepositoryError(r.store.SealStagedObject(ctx, operationID, profileID, sealed.StagingID, sealed.ByteSize, sealed.SHA256, now))
}
func (r *StagedObjectRepository) ReadSealed(ctx context.Context, operationID, profileID string, reference stagedobjects.Reference, now time.Time) (stagedobjects.SealedObject, error) {
	object, err := r.store.StagedObject(ctx, reference.StagingID)
	if err != nil {
		return stagedobjects.SealedObject{}, classifyStagedRepositoryError(err)
	}
	if object.OperationID != operationID || object.OwnerProfileID != profileID || object.State != extensionstore.StagedReady || !now.Before(object.StagingExpiresAt) {
		return stagedobjects.SealedObject{}, stagedobjects.ErrScopeDenied
	}
	if object.ExpectedByteSize == nil || object.ExpectedSHA256 == nil {
		return stagedobjects.SealedObject{}, stagedobjects.ErrIntegrity
	}
	return stagedobjects.SealedObject{StorageIdentity: object.StorageIdentity, ContentBinding: stagedobjects.ContentBinding{ByteSize: *object.ExpectedByteSize, SHA256: *object.ExpectedSHA256}}, nil
}

func (r *StagedObjectRepository) Abandon(ctx context.Context, stagingID string, now time.Time) error {
	return classifyStagedRepositoryError(r.store.AbandonStagedObject(ctx, stagingID, now))
}

func (r *StagedObjectRepository) PrepareCleanupBatch(ctx context.Context, cutoff, now time.Time, limit int) ([]stagedobjects.CleanupCandidate, error) {
	objects, err := r.store.PrepareCleanupBatch(ctx, cutoff, now, limit)
	if err != nil {
		return nil, classifyStagedRepositoryError(err)
	}
	candidates := make([]stagedobjects.CleanupCandidate, len(objects))
	for index, object := range objects {
		candidates[index] = stagedobjects.CleanupCandidate{
			StagingID:          object.StagingID,
			StorageIdentity:    object.StorageIdentity,
			DeleteAttemptCount: object.DeleteAttemptCount,
		}
	}
	return candidates, nil
}

func (r *StagedObjectRepository) RecordDeletionSuccess(ctx context.Context, stagingID string) error {
	return classifyStagedRepositoryError(r.store.RecordDeletionSuccess(ctx, stagingID))
}

func (r *StagedObjectRepository) RecordDeletionFailure(ctx context.Context, failure stagedobjects.DeletionFailure) error {
	return classifyStagedRepositoryError(r.store.RecordDeletionFailure(
		ctx,
		failure.StagingID,
		failure.AttemptCount,
		failure.SafeErrorCode,
		failure.NextAttemptAt,
	))
}

func classifyStagedRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, extensionstore.ErrIntegrity):
		return stagedobjects.NewFailure(stagedobjects.FailureIntegrity, "extensionstore_integrity", err)
	case errors.Is(err, extensionstore.ErrInvalidTransition),
		errors.Is(err, extensionstore.ErrNotFound):
		return stagedobjects.NewFailure(stagedobjects.FailureIntegrity, "extensionstore_transition", err)
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return stagedobjects.NewFailure(stagedobjects.FailureDependency, "postgres_deadline", err)
	default:
		return stagedobjects.NewFailure(stagedobjects.FailureDependency, "postgres_unavailable", err)
	}
}

// StagedObjectBytes is the platform object-store adapter. It returns only
// semantic outcomes, never platform DTOs or backend error details.
type StagedObjectBytes struct {
	store objectstore.Store
}

func NewStagedObjectBytes(store objectstore.Store) (*StagedObjectBytes, error) {
	if store == nil {
		return nil, errors.New("staged-object bytes require object store")
	}
	return &StagedObjectBytes{store: store}, nil
}

func (s *StagedObjectBytes) Put(ctx context.Context, storageIdentity string, payload io.Reader) (stagedobjects.ByteWriteOutcome, error) {
	if s == nil || s.store == nil {
		return stagedobjects.ByteWriteDependency, errors.New("object store unavailable")
	}
	err := s.store.PutObject(ctx, storageIdentity, payload, -1, "application/octet-stream")
	if err == nil {
		return stagedobjects.ByteWriteSuccess, nil
	}
	if adapterErr, ok := objectstore.AsAdapterError(err); ok {
		switch {
		case objectstore.IsDependencyError(err):
			return stagedobjects.ByteWriteDependency, adapterErr
		case adapterErr.Code == objectstore.ErrorCodeIntegrityMismatch,
			adapterErr.Code == objectstore.ErrorCodeInvalidRequest:
			return stagedobjects.ByteWriteIntegrity, adapterErr
		}
	}
	return stagedobjects.ByteWriteIndeterminate, err
}

func (s *StagedObjectBytes) Open(ctx context.Context, identity string) (io.ReadCloser, error) {
	if s == nil || s.store == nil {
		return nil, stagedobjects.ErrDependency
	}
	stream, _, err := s.store.ReadObject(ctx, identity, objectstore.ReadOptions{})
	return stream, err
}

func (s *StagedObjectBytes) Delete(ctx context.Context, storageIdentity string) (stagedobjects.DeleteOutcome, error) {
	if s == nil || s.store == nil {
		return stagedobjects.DeleteDependency, errors.New("object store unavailable")
	}
	var err error
	if typed, ok := s.store.(objectstore.TypedStore); ok {
		err = typed.Delete(ctx, objectstore.DeleteObjectRequest{
			Key:     storageIdentity,
			Purpose: objectstore.PurposeStagedCleanup,
		})
	} else {
		err = s.store.DeleteObject(ctx, storageIdentity)
	}
	if err == nil {
		return stagedobjects.DeleteSuccess, nil
	}
	if objectstore.IsObjectNotFound(err) {
		return stagedobjects.DeleteAbsent, nil
	}
	if adapterErr, ok := objectstore.AsAdapterError(err); ok {
		switch {
		case objectstore.IsDependencyError(adapterErr):
			return stagedobjects.DeleteDependency, adapterErr
		case adapterErr.Code == objectstore.ErrorCodeIntegrityMismatch,
			adapterErr.Code == objectstore.ErrorCodeInvalidRequest:
			return stagedobjects.DeleteIntegrity, adapterErr
		}
	}
	return stagedobjects.DeleteRetryableUnknown, err
}

type StagedPublicationCapability struct {
	operationID string
	tx          pgx.Tx
	now         func() time.Time
}

func NewStagedPublicationCapability(operationID string, tx pgx.Tx, now func() time.Time) (*StagedPublicationCapability, error) {
	if operationID == "" || tx == nil {
		return nil, stagedobjects.ErrScopeDenied
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &StagedPublicationCapability{operationID: operationID, tx: tx, now: now}, nil
}

func (c *StagedPublicationCapability) OperationID() string {
	if c == nil {
		return ""
	}
	return c.operationID
}

func (c *StagedPublicationCapability) PublishStagedObject(ctx context.Context, profileID string, publication stagedobjects.Publication) error {
	if c == nil || c.tx == nil {
		return stagedobjects.ErrScopeDenied
	}
	err := extensionstore.MarkStagedPublished(
		ctx,
		c.tx,
		c.operationID,
		profileID,
		publication.StagingID,
		publication.ResourceKind,
		publication.ResourceID,
		publication.ByteSize,
		publication.SHA256,
		c.now().UTC().Truncate(time.Microsecond),
	)
	return classifyStagedRepositoryError(err)
}
