package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

func (s *Store) ApplyAssessmentMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if err := mutation.Validate(); err != nil {
		return err
	}
	switch mutation.Kind {
	case assessmentprojection.ProjectionMutationUpsert:
		return s.physical.UpsertAssessmentTx(ctx, tx, mutation.Input)
	case assessmentprojection.ProjectionMutationDelete:
		return s.physical.DeleteAssessmentRowTx(ctx, tx, mutation.RecordID)
	default:
		return fmt.Errorf("unsupported assessment projection mutation kind %q", mutation.Kind)
	}
}

func (s *Store) refreshAssessmentTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source assessmentprojection.SourceReader,
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

func (s *Store) RebuildAssessments(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, assessmentsViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentAssessmentsTx(ctx, tx, incidentID)
	})
}

func (s *Store) rebuildIncidentAssessmentsTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source assessmentprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("assessment projection source is required")
	}
	if err := s.physical.DeleteAssessmentIncidentTx(ctx, tx, incidentID); err != nil {
		return err
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
