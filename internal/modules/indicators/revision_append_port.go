package indicators

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type revisionAppendPort interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.CapturedRecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
	AppendCapturedRecordMutationTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordMutationParams) error
	AppendCapturedRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordRevisionParams) error
}

func (a revisionAppendAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.CapturedRecordSnapshot, error) {
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

func (a revisionAppendAdapter) AppendCapturedRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendCapturedRecordMutationParams) error {
	return a.appender.AppendCapturedRecordMutationTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendCapturedRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendCapturedRecordRevisionParams) error {
	return a.appender.AppendCapturedRecordRevisionTx(ctx, tx, params)
}
