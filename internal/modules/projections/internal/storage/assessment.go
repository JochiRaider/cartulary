package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

func (store *Store) UpsertAssessmentTx(
	ctx context.Context,
	tx pgx.Tx,
	input assessmentprojection.ProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO assessment_grid_projection (
    record_id, incident_id, row_version, subject_ref, subject_type,
    assessment_state, confidence_score, confidence_band, rationale,
    assessor, assessed_at, supporting_link_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    subject_ref = EXCLUDED.subject_ref,
    subject_type = EXCLUDED.subject_type,
    assessment_state = EXCLUDED.assessment_state,
    confidence_score = EXCLUDED.confidence_score,
    confidence_band = EXCLUDED.confidence_band,
    rationale = EXCLUDED.rationale,
    assessor = EXCLUDED.assessor,
    assessed_at = EXCLUDED.assessed_at,
    supporting_link_count = EXCLUDED.supporting_link_count
`, input.RecordID, input.IncidentID, input.RowVersion, input.SubjectRef,
		input.SubjectType, input.AssessmentState, input.ConfidenceScore,
		input.ConfidenceBand, input.Rationale, input.Assessor,
		input.AssessedAt.UTC(), input.SupportingLinkCount); err != nil {
		return fmt.Errorf("upsert assessment projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteAssessmentRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM assessment_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete assessment projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteAssessmentIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM assessment_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear assessment projection rows: %w", err)
	}
	return nil
}
