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

func (p rollbackPublicationService) captureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RecordSnapshot, error) {
	return p.store.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (p rollbackPublicationService) appendNonRowMutationTx(ctx context.Context, tx pgx.Tx, params AppendNonRowMutationParams) error {
	return p.store.appender.AppendNonRowMutationTx(ctx, tx, params)
}

func (p rollbackPublicationService) appendRecordMutationTx(ctx context.Context, tx pgx.Tx, params AppendRecordMutationParams) error {
	return p.store.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (p rollbackPublicationService) appendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	return p.store.appender.AppendRecordRevisionAndIntentTx(ctx, tx, params)
}

func (p rollbackPublicationService) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return p.store.rebuildProjectionsTx(ctx, tx, incidentID)
}
