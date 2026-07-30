package projections

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
)

const (
	deleteAssessmentProjectionRowSQL = `
DELETE FROM assessment_grid_projection WHERE record_id = $1
`
	upsertAssessmentProjectionRowSQL = `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_score,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
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
`
)

type AssessmentSource interface {
	BuildProjectionMutationTx(
		context.Context,
		pgx.Tx,
		uuid.UUID,
	) (assessmentprojection.ProjectionMutation, error)
	ListProjectionInputsTx(
		context.Context,
		pgx.Tx,
		uuid.UUID,
		*uuid.UUID,
		int,
	) (assessmentprojection.ProjectionInputPage, error)
}

func (s *Store) ApplyAssessmentMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	if err := mutation.Validate(); err != nil {
		return err
	}
	switch mutation.Kind {
	case assessmentprojection.ProjectionMutationUpsert:
		return s.upsertAssessmentRowTx(ctx, tx, mutation.Input)
	case assessmentprojection.ProjectionMutationDelete:
		if _, err := tx.Exec(
			ctx,
			deleteAssessmentProjectionRowSQL,
			mutation.RecordID,
		); err != nil {
			return fmt.Errorf("delete assessment projection row: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported assessment projection mutation kind %q", mutation.Kind)
	}
}

func (s *Store) upsertAssessmentRowTx(
	ctx context.Context,
	tx pgx.Tx,
	input assessmentprojection.ProjectionInput,
) error {
	if _, err := tx.Exec(
		ctx,
		upsertAssessmentProjectionRowSQL,
		assessmentProjectionArgs(input)...,
	); err != nil {
		return fmt.Errorf("upsert assessment projection row: %w", err)
	}
	return nil
}

func ApplyAssessmentMutationSQLTx(
	ctx context.Context,
	tx *sql.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	if tx == nil {
		return errors.New("assessment projection SQL transaction is required")
	}
	if err := mutation.Validate(); err != nil {
		return err
	}
	switch mutation.Kind {
	case assessmentprojection.ProjectionMutationUpsert:
		if _, err := tx.ExecContext(
			ctx,
			upsertAssessmentProjectionRowSQL,
			assessmentProjectionArgs(mutation.Input)...,
		); err != nil {
			return fmt.Errorf("upsert assessment projection row: %w", err)
		}
		return nil
	case assessmentprojection.ProjectionMutationDelete:
		if _, err := tx.ExecContext(
			ctx,
			deleteAssessmentProjectionRowSQL,
			mutation.RecordID,
		); err != nil {
			return fmt.Errorf("delete assessment projection row: %w", err)
		}
		return nil
	default:
		return fmt.Errorf(
			"unsupported assessment projection mutation kind %q",
			mutation.Kind,
		)
	}
}

func assessmentProjectionArgs(input assessmentprojection.ProjectionInput) []any {
	return []any{
		input.RecordID,
		input.IncidentID,
		input.RowVersion,
		input.SubjectRef,
		input.SubjectType,
		input.AssessmentState,
		input.ConfidenceScore,
		input.ConfidenceBand,
		input.Rationale,
		input.Assessor,
		input.AssessedAt.UTC(),
		input.SupportingLinkCount,
	}
}

func (s *Store) refreshAssessmentTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source AssessmentSource,
) error {
	if s == nil || source == nil {
		return errors.New("assessment projection source is required")
	}
	mutation, err := source.BuildProjectionMutationTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	return s.ApplyAssessmentMutationTx(ctx, tx, mutation)
}

func (s *Store) RebuildIncidentAssessmentsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, assessmentsViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentAssessmentsTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source AssessmentSource,
) error {
	if s == nil || source == nil {
		return errors.New("assessment projection source is required")
	}
	if _, err := tx.Exec(
		ctx,
		`DELETE FROM assessment_grid_projection WHERE incident_id = $1`,
		incidentID,
	); err != nil {
		return fmt.Errorf("clear assessment projection rows: %w", err)
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(
			ctx,
			tx,
			incidentID,
			afterRecordID,
			500,
		)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.ApplyAssessmentMutationTx(
				ctx,
				tx,
				assessmentprojection.ProjectionMutation{
					Kind:     assessmentprojection.ProjectionMutationUpsert,
					RecordID: input.RecordID,
					Input:    input,
				},
			); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			break
		}
		afterRecordID = page.NextRecordID
	}
	return nil
}
