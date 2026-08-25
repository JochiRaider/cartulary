package assessments

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

type ImportCreateDependencies struct {
	Subjects      SubjectValidator
	Assessors     AssessorValidator
	Records       RecordEnvelopeCreator
	Revisions     ownerfacade.LiveRecordRevisionAppender
	Projections   AssessmentProjectionPort
	Collaboration collaboration.RecordChangedAppender
}

type importCreateFacade struct {
	source       assessmentSourceRepository
	subjects     SubjectValidator
	assessors    AssessorValidator
	records      RecordEnvelopeCreator
	revisions    ownerfacade.LiveRecordRevisionAppender
	projections  AssessmentProjectionPort
	publications collaboration.RecordChangedAppender
}

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	dependencies ImportCreateDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != AssessmentsViewSchemaID {
		return nil, fmt.Errorf("assessment import surface %q not mapped", targetViewSchemaID)
	}
	switch {
	case dependencies.Subjects == nil:
		return nil, errors.New("construct assessment import facade: subject validator is required")
	case dependencies.Assessors == nil:
		return nil, errors.New("construct assessment import facade: assessor validator is required")
	case dependencies.Records == nil:
		return nil, errors.New("construct assessment import facade: record-envelope creator is required")
	case dependencies.Revisions == nil:
		return nil, errors.New("construct assessment import facade: revision appender is required")
	case dependencies.Projections == nil:
		return nil, errors.New("construct assessment import facade: projection port is required")
	case dependencies.Collaboration == nil:
		return nil, errors.New("construct assessment import facade: Collaboration publication appender is required")
	}
	owner := &importCreateFacade{
		source:       assessmentSourceRepository{},
		subjects:     dependencies.Subjects,
		assessors:    dependencies.Assessors,
		records:      dependencies.Records,
		revisions:    dependencies.Revisions,
		projections:  dependencies.Projections,
		publications: dependencies.Collaboration,
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		owner.CreateImportRowTx,
	)
}

func (f *importCreateFacade) CreateImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ownerfacade.ImportOwnerCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != AssessmentsViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"assessment import surface %q not mapped",
			request.TargetViewSchemaID,
		)
	}
	input, err := assessmentCreateInputFromImport(
		request,
		ownerfacade.ValuesByField(request.FieldValues),
	)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := validateCreateInputShape(input); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	valid, err := f.subjects.ValidateAssessmentSubjectTx(
		ctx,
		tx,
		request.IncidentID,
		input.SubjectType,
		input.SubjectRef,
	)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"validate assessment import subject: %w",
			err,
		)
	}
	if !valid {
		return ownerfacade.ImportOwnerCreateResponse{}, &CreateValidationError{
			Field:      "assessment.subject_ref",
			ReasonCode: "invalid_value",
		}
	}

	assessorID := request.ActorUserID
	if input.Assessor != nil {
		assessorID = *input.Assessor
	}
	valid, err = f.assessors.ValidateAssessmentAssessorTx(
		ctx,
		tx,
		request.IncidentID,
		assessorID,
	)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"validate assessment import assessor: %w",
			err,
		)
	}
	if !valid {
		return ownerfacade.ImportOwnerCreateResponse{}, &CreateValidationError{
			Field:      "assessment.assessor",
			ReasonCode: "invalid_value",
		}
	}

	now := command.Now.UTC()
	assessedAt := now
	if input.AssessedAt != nil {
		assessedAt = input.AssessedAt.UTC()
	}
	recordID, err := f.records.CreateAssessmentEnvelopeTx(ctx, tx, RecordEnvelopeCreate{
		IncidentID: request.IncidentID,
		RecordType: "assessment",
		ActorID:    request.ActorUserID,
		Now:        now,
		RowVersion: 1,
	})
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"create imported assessment envelope: %w",
			err,
		)
	}
	if err := f.source.InsertTx(ctx, tx, assessmentSourceCreate{
		RecordID:        recordID,
		IncidentID:      request.IncidentID,
		SubjectRef:      input.SubjectRef,
		SubjectType:     input.SubjectType,
		AssessmentState: input.AssessmentState,
		ConfidenceScore: input.ConfidenceScore,
		Rationale:       input.Rationale,
		Assessor:        assessorID,
		AssessedAt:      assessedAt,
		Now:             now,
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := f.projections.RefreshAndLoadAssessmentRowTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"refresh imported assessment projection: %w",
			err,
		)
	}
	response, err := ownerfacade.FinalizeLiveRecordTx(ctx, tx, f.revisions, f.publications, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       "create",
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		Row:             row,
		CreatedAt:       command.Now,
	})
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("finalize imported assessment: %w", err)
	}
	return response, nil
}

func assessmentCreateInputFromImport(
	request ownerfacade.ImportOwnerCreateRequest,
	values map[string]ownerfacade.ImportScalarValue,
) (CreateInput, error) {
	result := CreateInput{ClientTxnID: request.ClientTxnID}
	if value, ok := values["assessment.subject_ref"]; ok && value.UUID != nil {
		result.SubjectRef = *value.UUID
	}
	if value, ok := values["assessment.subject_type"]; ok && value.Text != nil {
		result.SubjectType = *value.Text
	}
	if value, ok := values["assessment.assessment_state"]; ok && value.Text != nil {
		result.AssessmentState = *value.Text
	}
	if value, ok := values["assessment.confidence_score"]; ok && value.Number != nil {
		score := int(*value.Number)
		result.ConfidenceScore = &score
	}
	if value, ok := values["assessment.rationale"]; ok && value.Text != nil {
		result.Rationale = *value.Text
	}
	if value, ok := values["assessment.assessor"]; ok && value.UUID != nil {
		result.Assessor = value.UUID
	}
	if value, ok := values["assessment.assessed_at"]; ok && value.Timestamp != nil {
		result.AssessedAt = value.Timestamp
	}
	return result, nil
}
