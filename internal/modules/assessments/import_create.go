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
	revisions    ownerfacade.LiveRecordRevisionAppender
	publications collaboration.RecordChangedAppender
	creator      assessmentCreateService
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
	case isNilDependency(dependencies.Subjects):
		return nil, errors.New("construct assessment import facade: subject validator is required")
	case isNilDependency(dependencies.Assessors):
		return nil, errors.New("construct assessment import facade: assessor validator is required")
	case isNilDependency(dependencies.Records):
		return nil, errors.New("construct assessment import facade: record-envelope creator is required")
	case isNilDependency(dependencies.Revisions):
		return nil, errors.New("construct assessment import facade: revision appender is required")
	case isNilDependency(dependencies.Projections):
		return nil, errors.New("construct assessment import facade: projection port is required")
	case isNilDependency(dependencies.Collaboration):
		return nil, errors.New("construct assessment import facade: collaboration publication appender is required")
	}
	owner := &importCreateFacade{
		revisions:    dependencies.Revisions,
		publications: dependencies.Collaboration,
		creator: newAssessmentCreateService(
			dependencies.Subjects,
			dependencies.Assessors,
			dependencies.Records,
			dependencies.Projections,
		),
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		owner.createImportRowTx,
	)
}

func (f *importCreateFacade) createImportRowTx(
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
	input, err := assessmentCreateInputFromImport(request)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := f.creator.validateInput(input); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	create := assessmentCreateContext{
		IncidentID:  request.IncidentID,
		ActorUserID: request.ActorUserID,
		Input:       input,
		Now:         command.Now.UTC(),
	}
	if err := f.creator.validateSubjectTx(ctx, tx, create); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}

	assessorID, err := f.creator.resolveAssessorTx(ctx, tx, create)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}

	recordID, err := f.creator.insertTx(ctx, tx, create, assessorID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := f.creator.refreshProjectionTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
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
) (CreateInput, error) {
	if _, err := ownerfacade.IndexImportFieldValues(request.FieldValues); err != nil {
		return CreateInput{}, err
	}
	result := CreateInput{ClientTxnID: request.ClientTxnID}
	for _, field := range request.FieldValues {
		value := field.NormalizedValue
		expected, ok := assessmentImportFieldKinds[field.FieldKey]
		if !ok {
			if field.FieldKey == "assessment.support_refs" {
				return CreateInput{}, ownerfacade.NewImportOwnerCreateValidationError(
					"collection_owner_support_required",
					field.FieldKey,
					"collection_review",
					nil,
				)
			}
			return CreateInput{}, ownerfacade.NewImportOwnerCreateValidationError(
				"field_not_import_writable",
				field.FieldKey,
				"create_writable",
				nil,
			)
		}
		if value.Kind() == ownerfacade.ImportScalarNull {
			return CreateInput{}, ownerfacade.NewImportOwnerCreateValidationError(
				"field_not_nullable",
				field.FieldKey,
				"clearable",
				nil,
			)
		}
		if value.Kind() != expected.kind {
			return CreateInput{}, ownerfacade.NewImportOwnerCreateValidationError(
				expected.reasonCode,
				field.FieldKey,
				expected.guard,
				nil,
			)
		}
		switch field.FieldKey {
		case "assessment.subject_ref":
			result.SubjectRef, _ = value.UUID()
		case "assessment.subject_type":
			result.SubjectType, _ = value.Text()
		case "assessment.assessment_state":
			result.AssessmentState, _ = value.Text()
		case "assessment.confidence_score":
			if number, present := value.Number(); present {
				score := int(number)
				result.ConfidenceScore = &score
			}
		case "assessment.rationale":
			result.Rationale, _ = value.Text()
		case "assessment.assessor":
			if id, present := value.UUID(); present {
				result.Assessor = &id
			}
		case "assessment.assessed_at":
			if timestamp, present := value.Timestamp(); present {
				result.AssessedAt = &timestamp
			}
		}
	}
	return result, nil
}

type assessmentImportFieldKind struct {
	kind       ownerfacade.ImportScalarKind
	reasonCode string
	guard      string
}

var assessmentImportFieldKinds = map[string]assessmentImportFieldKind{
	"assessment.subject_ref": {
		kind: ownerfacade.ImportScalarUUID, reasonCode: "invalid_uuid", guard: "uuid",
	},
	"assessment.subject_type": {
		kind: ownerfacade.ImportScalarText, reasonCode: "invalid_text", guard: "line_v1",
	},
	"assessment.assessment_state": {
		kind: ownerfacade.ImportScalarText, reasonCode: "invalid_text", guard: "line_v1",
	},
	"assessment.confidence_score": {
		kind: ownerfacade.ImportScalarNumber, reasonCode: "invalid_integer", guard: "number",
	},
	"assessment.rationale": {
		kind: ownerfacade.ImportScalarText, reasonCode: "invalid_text", guard: "multiline_body_v1",
	},
	"assessment.assessor": {
		kind: ownerfacade.ImportScalarUUID, reasonCode: "invalid_uuid", guard: "uuid",
	},
	"assessment.assessed_at": {
		kind: ownerfacade.ImportScalarTimestamp, reasonCode: "invalid_timestamp", guard: "timestamp_instant_v1",
	},
}
