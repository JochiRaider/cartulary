package assessments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type ImportCreateCommand struct {
	Request     tabularingest.ImportOwnerCreateRequest
	ChangeSetID uuid.UUID
	SequenceNo  int
	Now         time.Time
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (tabularingest.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != AssessmentsViewSchemaID {
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("assessment import surface %q not mapped", request.TargetViewSchemaID)
	}
	values := assessmentImportValuesByField(request.FieldValues)
	createRequest, err := assessmentCreateRequestFromImport(request, values)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := validateCreateRequestShape(createRequest); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := validateSubjectTx(ctx, tx, request.IncidentID, *createRequest.SubjectRef, createRequest.SubjectType); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if requestAssessor := createRequest.Assessor; requestAssessor != nil {
		if err := validateAssessorTx(ctx, tx, *requestAssessor); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
	}

	now := command.Now.UTC()
	assessedAt := now
	if createRequest.AssessedAt != nil {
		assessedAt = createRequest.AssessedAt.UTC()
	}
	assessor := request.ActorUserID
	if createRequest.Assessor != nil {
		assessor = *createRequest.Assessor
	}

	recordID, err := s.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      request.IncidentID,
		RecordType:      "assessment",
		CreatedByUserID: request.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: request.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := insertAssessmentSourceTx(ctx, tx, recordID, request.IncidentID, createRequest, assessor, assessedAt, now); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	for _, supportRef := range uniqueUUIDs(createRequest.SupportRefs) {
		if _, _, err := s.linkStore.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
			IncidentID:  request.IncidentID,
			SrcRecordID: recordID,
			DstRecordID: supportRef,
			LinkType:    links.LinkType(links.LinkTypeSupportedBy),
			Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
			OwnerUserID: request.ActorUserID,
			Now:         now,
		}); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
	}
	if err := s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.AssessmentsViewSchemaID, recordID); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	projected, err := loadProjectionRecordTx(ctx, tx, recordID)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	row := BuildAssessmentRow(projected)
	rowVersion := projected.RowVersion
	afterVersionID := fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    command.ChangeSetID,
		SequenceNo:     command.SequenceNo,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: command.ChangeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		AfterValue:  row,
	}); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	return tabularingest.ImportOwnerCreateResponse{
		RecordID:             recordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      "created",
		OwnerResultCode:      "created",
		RowRefresh:           row,
	}, nil
}

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if viewSchemaID != AssessmentsViewSchemaID {
		return nil, fmt.Errorf("assessment import projection surface %q not mapped", viewSchemaID)
	}
	if err := s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.AssessmentsViewSchemaID, recordID); err != nil {
		return nil, err
	}
	return s.rowProjector.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

func assessmentCreateRequestFromImport(request tabularingest.ImportOwnerCreateRequest, values map[string]tabularingest.ImportScalarValue) (CreateRequest, error) {
	result := CreateRequest{ClientTxnID: request.ClientTxnID}
	if value, ok := values["assessment.subject_ref"]; ok && value.UUID != nil {
		result.SubjectRef = value.UUID
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

func assessmentImportValuesByField(fields []tabularingest.ImportFieldValue) map[string]tabularingest.ImportScalarValue {
	values := make(map[string]tabularingest.ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}
