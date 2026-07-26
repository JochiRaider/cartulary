package adapters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

type IncidentImportRebuilder struct {
	projector *RowProjector
}

func NewIncidentImportRebuilder(pool *pgxpool.Pool, timelineSources ...projections.TimelineSource) IncidentImportRebuilder {
	return IncidentImportRebuilder{
		projector: NewRowProjector(pool, timelineSources...),
	}
}

func (r IncidentImportRebuilder) RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.projector.RebuildIncidentTx(ctx, tx, incidentID)
}
