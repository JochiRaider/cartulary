package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectionRebuilder interface {
	RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

func (s *commandStore) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.projectionRebuilder.RebuildIncidentTx(ctx, tx, incidentID)
}
