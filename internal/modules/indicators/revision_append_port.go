package indicators

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type revisionAppendPort interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
}

func (a revisionAppendAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

type revisionAppendAdapter struct{ appender *revisions.Appender }

func newRevisionAppendAdapter(appender *revisions.Appender) revisionAppendAdapter {
	return revisionAppendAdapter{appender: appender}
}

func (a revisionAppendAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendNonRowMutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionAndIntentTx(ctx, tx, params)
}
