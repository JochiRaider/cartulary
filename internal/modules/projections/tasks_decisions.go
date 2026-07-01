package projections

import (
	"context"

	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) refreshTaskRequestTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return taskdecisionprojection.RefreshTaskRequestTx(ctx, tx, recordID)
}

func (s *Store) refreshDecisionTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return taskdecisionprojection.RefreshDecisionTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentTaskRequestsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, taskRequestsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentTaskRequestsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return taskdecisionprojection.RebuildIncidentTaskRequestsTx(ctx, tx, incidentID)
}

func (s *Store) RebuildIncidentDecisionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, decisionsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentDecisionsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return taskdecisionprojection.RebuildIncidentDecisionsTx(ctx, tx, incidentID)
}
