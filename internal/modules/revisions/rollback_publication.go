package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// rollbackPublicationService is the single history/projection publication seam
// used by the transactional applier. Source owners still own the concrete
// projection and record-revision semantics supplied to commandStore.
type rollbackPublicationService struct {
	store *commandStore
}

func (p rollbackPublicationService) appendChangeSetTx(ctx context.Context, tx pgx.Tx, params AppendChangeSetParams) (uuid.UUID, error) {
	return p.store.appender.AppendChangeSetTx(ctx, tx, params)
}

func (p rollbackPublicationService) captureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (CapturedRecordSnapshot, error) {
	return p.store.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (p rollbackPublicationService) appendNonRowMutationTx(ctx context.Context, tx pgx.Tx, params AppendNonRowMutationParams) error {
	return p.store.appender.AppendNonRowMutationTx(ctx, tx, params)
}

func (p rollbackPublicationService) appendCapturedRecordMutationTx(ctx context.Context, tx pgx.Tx, params AppendCapturedRecordMutationParams) error {
	return p.store.appender.AppendCapturedRecordMutationTx(ctx, tx, params)
}

func (p rollbackPublicationService) appendCapturedRecordRevisionTx(ctx context.Context, tx pgx.Tx, params AppendCapturedRecordRevisionParams) error {
	return p.store.appender.AppendCapturedRecordRevisionTx(ctx, tx, params)
}

func (p rollbackPublicationService) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return p.store.rebuildProjectionsTx(ctx, tx, incidentID)
}
