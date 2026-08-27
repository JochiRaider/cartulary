package assessments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type assessmentCreateService struct {
	source      assessmentSourceRepository
	subjects    SubjectValidator
	assessors   AssessorValidator
	records     RecordEnvelopeCreator
	projections AssessmentProjectionPort
}

type assessmentCreateContext struct {
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Input       CreateInput
	Now         time.Time
}

func newAssessmentCreateService(
	subjects SubjectValidator,
	assessors AssessorValidator,
	records RecordEnvelopeCreator,
	projections AssessmentProjectionPort,
) assessmentCreateService {
	return assessmentCreateService{
		source:      assessmentSourceRepository{},
		subjects:    subjects,
		assessors:   assessors,
		records:     records,
		projections: projections,
	}
}

func (assessmentCreateService) validateInput(input CreateInput) error {
	return validateCreateInputShape(input)
}

func (service assessmentCreateService) validateSubjectTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessmentCreateContext,
) error {
	valid, err := service.subjects.ValidateAssessmentSubjectTx(
		ctx,
		tx,
		create.IncidentID,
		create.Input.SubjectType,
		create.Input.SubjectRef,
	)
	if err != nil {
		return fmt.Errorf("validate assessment subject: %w", err)
	}
	if !valid {
		return &CreateValidationError{
			Field:      "assessment.subject_ref",
			ReasonCode: "invalid_value",
		}
	}
	return nil
}

func (service assessmentCreateService) resolveAssessorTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessmentCreateContext,
) (uuid.UUID, error) {
	assessorID := create.ActorUserID
	if create.Input.Assessor != nil {
		assessorID = *create.Input.Assessor
	}
	valid, err := service.assessors.ValidateAssessmentAssessorTx(
		ctx,
		tx,
		create.IncidentID,
		assessorID,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate assessment assessor: %w", err)
	}
	if !valid {
		return uuid.Nil, &CreateValidationError{
			Field:      "assessment.assessor",
			ReasonCode: "invalid_value",
		}
	}
	return assessorID, nil
}

func (service assessmentCreateService) insertTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessmentCreateContext,
	assessorID uuid.UUID,
) (uuid.UUID, error) {
	now := create.Now.UTC()
	assessedAt := now
	if create.Input.AssessedAt != nil {
		assessedAt = create.Input.AssessedAt.UTC()
	}
	recordID, err := service.records.CreateAssessmentEnvelopeTx(ctx, tx, RecordEnvelopeCreate{
		IncidentID: create.IncidentID,
		ActorID:    create.ActorUserID,
		Now:        now,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create assessment record envelope: %w", err)
	}
	if err := service.source.insertTx(ctx, tx, assessmentSourceCreate{
		RecordID:        recordID,
		IncidentID:      create.IncidentID,
		SubjectRef:      create.Input.SubjectRef,
		SubjectType:     create.Input.SubjectType,
		AssessmentState: create.Input.AssessmentState,
		ConfidenceScore: create.Input.ConfidenceScore,
		Rationale:       create.Input.Rationale,
		Assessor:        assessorID,
		AssessedAt:      assessedAt,
		Now:             now,
	}); err != nil {
		return uuid.Nil, err
	}
	return recordID, nil
}

func (service assessmentCreateService) refreshProjectionTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	row, err := service.projections.RefreshAndLoadAssessmentRowTx(ctx, tx, recordID)
	if err != nil {
		return nil, fmt.Errorf("refresh assessment projection: %w", err)
	}
	return row, nil
}
