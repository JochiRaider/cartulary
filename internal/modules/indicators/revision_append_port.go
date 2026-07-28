package indicators

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type revisionAppendPort interface {
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
}

type revisionAppendAdapter struct{ appender *revisions.Appender }

func newRevisionAppendAdapter(appender *revisions.Appender) revisionAppendAdapter {
	return revisionAppendAdapter{appender: appender}
}

func (a revisionAppendAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendMutationParams) error {
	return a.appender.AppendMutationTx(ctx, tx, params)
}

func (a revisionAppendAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, params)
}
