package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
)

type ProjectionRebuilder interface {
	RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

type defaultProjectionRebuilder struct{}

func (defaultProjectionRebuilder) RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return projectionadapters.NewRowProjector(nil).RebuildIncidentTx(ctx, tx, incidentID)
}

func (s *Store) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.projectionRebuilder.RebuildIncidentTx(ctx, tx, incidentID)
}
