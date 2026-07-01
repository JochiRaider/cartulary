package projections

import (
	"context"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) RefreshAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, assessmentsViewSchemaID, recordID)
}

func (s *Store) refreshAssessmentTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return assessmentprojection.RefreshAssessmentTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentAssessmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, assessmentsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentAssessmentsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return assessmentprojection.RebuildIncidentAssessmentsTx(ctx, tx, incidentID)
}
