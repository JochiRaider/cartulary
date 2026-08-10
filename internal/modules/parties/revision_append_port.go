package parties

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type revisionAppendPort interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.CapturedRecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendCapturedRecordMutationTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordMutationParams) error
	AppendCapturedRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordRevisionParams) error
}

type revisionAppendAdapter struct{ appender *revisions.Appender }

func newRevisionAppendAdapter(appender *revisions.Appender) revisionAppendAdapter {
	return revisionAppendAdapter{appender: appender}
}

func (a revisionAppendAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.CapturedRecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a revisionAppendAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendCapturedRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendCapturedRecordMutationParams) error {
	return a.appender.AppendCapturedRecordMutationTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendCapturedRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendCapturedRecordRevisionParams) error {
	return a.appender.AppendCapturedRecordRevisionTx(ctx, tx, params)
}
