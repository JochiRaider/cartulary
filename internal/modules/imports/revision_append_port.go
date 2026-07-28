package imports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type revisionAppendPort interface {
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
}

type revisionAppendAdapter struct{ appender *revisions.Appender }

func newRevisionAppendAdapter(appender *revisions.Appender) revisionAppendPort {
	return revisionAppendAdapter{appender: appender}
}

func (a revisionAppendAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}
