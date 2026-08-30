package assessments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type assessmentImportRevisionAppender interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
}

type ImportCreateDependencies struct {
	Subjects      SubjectValidator
	Assessors     AssessorValidator
	Records       RecordEnvelopeCreator
	Revisions     assessmentImportRevisionAppender
	Projections   AssessmentProjectionPort
	Collaboration collaboration.RecordChangedAppender
}

type importCreateFacade struct {
	revisions    assessmentImportRevisionAppender
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
	response, err := f.finalizeImportCreateTx(ctx, tx, command, recordID, row)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("finalize imported assessment: %w", err)
	}
	return response, nil
}

func (f *importCreateFacade) finalizeImportCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	command ownerfacade.ImportOwnerCreateCommand,
	recordID uuid.UUID,
	row map[string]any,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	rowVersion, err := canonicalRowVersion(row)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	afterVersionID := fmt.Sprintf("assessment:%s:%d", recordID, rowVersion)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    command.ChangeSetID,
		SequenceNo:     command.SequenceNo,
		TargetKind:     "assessment",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	privateFieldKeys := assessmentCreateRevisionFieldKeys()
	if err := f.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:   command.ChangeSetID,
		RecordID:      recordID,
		RowVersion:    rowVersion,
		AfterSnapshot: &afterSnapshot,
		ConflictFacts: assessmentCreateRevisionFacts(row, privateFieldKeys),
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	request := command.Request
	if err := appendAssessmentCreatePublicationTx(
		ctx, tx, f.publications, request.IncidentID, request.ActorUserID,
		request.ClientTxnID, command.ChangeSetID, recordID, rowVersion,
		max(command.SequenceNo-1, 0), command.Now, row,
	); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:             recordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      "created",
		OwnerResultCode:      "created",
		RowRefresh:           row,
	}, nil
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

var assessmentCreateFields = [...]string{
	"assessment.assessed_at",
	"assessment.assessment_state",
	"assessment.assessor",
	"assessment.confidence_score",
	"assessment.rationale",
	"assessment.subject_ref",
	"assessment.subject_type",
	"assessment.support_refs",
}

func assessmentCreateRevisionFieldKeys() []string {
	return append([]string(nil), assessmentCreateFields[:]...)
}

func assessmentCreatePublicationFieldKeys() []string {
	return append([]string(nil), assessmentCreateFields[:]...)
}

func assessmentCreateRevisionFacts(row map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	cells, _ := row["cells"].(map[string]any)
	facts := make([]revisions.RevisionConflictFact, 0, len(fieldKeys))
	for _, key := range fieldKeys {
		value, present := cells[key]
		facts = append(facts, revisions.RevisionConflictFact{
			FieldKey: key, AfterPresent: present, AfterValue: value,
		})
	}
	return facts
}

func appendAssessmentCreatePublicationTx(
	ctx context.Context,
	tx pgx.Tx,
	publications collaboration.RecordChangedAppender,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	mutationOrdinal int,
	createdAt time.Time,
	row map[string]any,
) error {
	fieldKeys := assessmentCreatePublicationFieldKeys()
	patch := collabprotocol.BuildViewRowPatch(row, fieldKeys)
	changeKind := "invalidate"
	if patch != nil {
		changeKind = "patch"
	}
	return publications.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID: incidentID, RecordID: recordID, ChangeSetID: changeSetID,
		ActorUserID: actorUserID, RowVersion: rowVersion, ClientTxnID: clientTxnID,
		MutationOrdinal: mutationOrdinal, CreatedAt: createdAt.UTC(), PublicFieldKeys: fieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{
			ViewSchemaID: AssessmentsViewSchemaID, RecordID: recordID,
			RowVersion: rowVersion, ChangeKind: changeKind, PatchCells: patch,
		}},
	})
}
