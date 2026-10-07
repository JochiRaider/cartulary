package runtime

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) LocateRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID, query viewschema.QueryMeta) (querypage.Location, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return querypage.Location{}, err
	}
	defer tx.Rollback(ctx)
	read := func(q viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
		return s.queryRowsPage(ctx, tx, incidentID, viewSchemaID, q, window)
	}
	location, err := querypage.Locate(recordID.String(), query, read)
	if err != nil {
		return querypage.Location{}, err
	}
	return location, tx.Commit(ctx)
}
