package adapters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentImportRebuilder struct {
	pool *pgxpool.Pool
}

func NewIncidentImportRebuilder(pool *pgxpool.Pool) IncidentImportRebuilder {
	return IncidentImportRebuilder{pool: pool}
}

func (r IncidentImportRebuilder) RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return NewRowProjector(r.pool).RebuildIncidentTx(ctx, tx, incidentID)
}
