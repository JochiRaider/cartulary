package workbook

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
)

type revisionHistoryPort interface {
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]historyquery.RevisionWindowRow, error)
}

type revisionHistoryAdapter struct{ reader historyquery.Reader }

func newRevisionHistoryAdapter() revisionHistoryAdapter {
	return revisionHistoryAdapter{reader: historyquery.NewReader()}
}

func (a revisionHistoryAdapter) LoadRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) ([]historyquery.RevisionWindowRow, error) {
	return a.reader.LoadRevisionWindowTx(ctx, tx, recordID, baseRowVersion, currentRowVersion)
}
