package projections

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskDecisionSource interface {
	RefreshTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshDecisionTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildIncidentTaskRequestsTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildIncidentDecisionsTx(context.Context, pgx.Tx, uuid.UUID) error
}

func (s *Store) refreshTaskRequestTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, source TaskDecisionSource) error {
	if source == nil {
		return errors.New("task/decision projection source is required")
	}
	return source.RefreshTaskRequestTx(ctx, tx, recordID)
}

func (s *Store) refreshDecisionTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, source TaskDecisionSource) error {
	if source == nil {
		return errors.New("task/decision projection source is required")
	}
	return source.RefreshDecisionTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentTaskRequestsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, taskRequestsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentTaskRequestsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source TaskDecisionSource) error {
	if source == nil {
		return errors.New("task/decision projection source is required")
	}
	return source.RebuildIncidentTaskRequestsTx(ctx, tx, incidentID)
}

func (s *Store) RebuildIncidentDecisionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, decisionsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentDecisionsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source TaskDecisionSource) error {
	if source == nil {
		return errors.New("task/decision projection source is required")
	}
	return source.RebuildIncidentDecisionsTx(ctx, tx, incidentID)
}
