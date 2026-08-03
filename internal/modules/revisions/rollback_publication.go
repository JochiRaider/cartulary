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

func (p rollbackPublicationService) appendMutationTx(ctx context.Context, tx pgx.Tx, params AppendMutationParams) error {
	return p.store.appender.AppendMutationTx(ctx, tx, params)
}

func (p rollbackPublicationService) appendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	return p.store.appender.AppendRecordRevisionTx(ctx, tx, params)
}

func (p rollbackPublicationService) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return p.store.rebuildProjectionsTx(ctx, tx, incidentID)
}
