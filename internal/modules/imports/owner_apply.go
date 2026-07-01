package imports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const importApplyChangeSetSource = "imports.apply"

const (
	evidenceImportViewSchemaID     = "cartulary.view.evidence.v1"
	partiesImportViewSchemaID      = "cartulary.view.parties.v1"
	taskRequestsImportViewSchemaID = "cartulary.view.task_requests.v1"
	decisionsImportViewSchemaID    = "cartulary.view.decisions.v1"
	assessmentsImportViewSchemaID  = "cartulary.view.assessments.v1"
)

type importOwnerStores struct {
	artifacts      *artifacts.Store
	entities       *entities.Store
	evidence       *evidence.Store
	links          *links.Store
	records        *records.Store
	revisions      *revisions.Store
	tasksDecisions *tasksdecisions.Store
}

type importOwnerApplyResult struct {
	Response  tabularingest.ImportOwnerCreateResponse
	Operation string
}

func (s *Service) applyGenericOwnerUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData, target importTarget) error {
	now := s.now().UTC()
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin import apply unit transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := incidents.EnsureIncidentOpenTx(ctx, tx, start.IncidentID); err != nil {
		return err
	}
	clientTxnID := fmt.Sprintf("import:%s:%s:%s", start.ImportSessionID, unit.UnitID, start.ClientTxnID)
	requestID := "req-" + clientTxnID
	changeSetID, err := revisions.NewStore().InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  start.IncidentID,
		ActorUserID: actor.ID,
		Source:      importApplyChangeSetSource,
		ClientTxnID: &clientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return err
	}

	stores := importOwnerStores{
		artifacts:      artifacts.NewStore(),
		entities:       s.entityStore,
		evidence:       evidence.NewStore(s.store.pool),
		links:          links.NewStore(),
		records:        records.NewStore(),
		revisions:      revisions.NewStore(s.store.pool),
		tasksDecisions: tasksdecisions.NewStore(),
	}

	for index, sourceRow := range unit.SourceRows {
		rowRef, _ := intFromAny(sourceRow["source_row_ref"])
		if rowRef <= 0 {
			return fmt.Errorf("import source row missing source_row_ref")
		}
		rowClientTxnID := fmt.Sprintf("%s:%d", clientTxnID, rowRef)
		request, err := importOwnerCreateRequest(start, unit, actor.ID, sourceRow, rowRef, rowClientTxnID)
		if err != nil {
			return err
		}
		result, err := applyOwnerCreateTx(ctx, tx, stores, actor.ID, target, request, changeSetID, index+1, now)
		if err != nil {
			return err
		}
		ownerResponse := map[string]any{
			"record_id":               result.Response.RecordID.String(),
			"row_version":             result.Response.RowVersion,
			"change_set_mutation_ref": result.Response.ChangeSetMutationRef,
			"created_or_reused":       result.Response.CreatedOrReused,
			"owner_result_code":       result.Response.OwnerResultCode,
			"target_view_schema_id":   request.TargetViewSchemaID,
			"import_session_id":       request.ImportSessionID.String(),
			"import_unit_id":          request.ImportUnitID.String(),
			"mapping_fingerprint":     request.MappingFingerprint,
			"source_row_ref":          request.SourceRowRef,
		}
		if err := s.store.InsertApplyJournalTx(ctx, tx, ApplyJournalParams{
			ImportSessionID:      request.ImportSessionID,
			ImportUnitID:         request.ImportUnitID,
			MappingFingerprint:   request.MappingFingerprint,
			SourceRowRef:         request.SourceRowRef,
			TargetViewSchemaID:   request.TargetViewSchemaID,
			OwnerCreateFacade:    target.CreateFacade,
			RecordID:             result.Response.RecordID,
			RowVersion:           result.Response.RowVersion,
			ChangeSetID:          changeSetID,
			ChangeSetMutationRef: result.Response.ChangeSetMutationRef,
			OwnerResultCode:      result.Response.OwnerResultCode,
			CreatedOrReused:      result.Response.CreatedOrReused,
			OwnerResponse:        ownerResponse,
			RowRefresh:           result.Response.RowRefresh,
			CreatedAt:            now,
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit import apply unit transaction: %w", err)
	}
	return nil
}

func importOwnerCreateRequest(start ApplyStartResult, unit ApplyUnitData, actorID uuid.UUID, sourceRow map[string]any, rowRef int, clientTxnID string) (tabularingest.ImportOwnerCreateRequest, error) {
	cells := sourceRowCellsByOrdinal(sourceRow)
	request := tabularingest.ImportOwnerCreateRequest{
		IncidentID:          start.IncidentID,
		ActorUserID:         actorID,
		TargetViewSchemaID:  unit.ApprovedMapping.TargetViewSchemaID,
		ImportSessionID:     start.ImportSessionID,
		ImportUnitID:        unit.UnitID,
		MappingFingerprint:  unit.MappingFingerprint,
		SourceFileKind:      unit.SourceFileKind,
		SourceContentSHA256: unit.SourceContentSHA256,
		ParserProfileID:     unit.ParserProfileID,
		ParserVersion:       unit.ParserVersion,
		LocatorKind:         unit.LocatorKind,
		Locator:             unit.Locator,
		SourceRectA1:        unit.SourceRectA1,
		SourceRowRef:        rowRef,
		ClientTxnID:         clientTxnID,
		SourceRowProvenance: tabularingest.ImportSourceRowProvenance{SourceRowRef: rowRef},
	}
	for _, column := range unit.ApprovedMapping.SourceColumns {
		cell := cells[column.SourceColumnOrdinal]
		rawValue, _ := cell["display_text"].(string)
		cellKind, _ := cell["cell_kind"].(string)
		if column.FieldKey == nil {
			request.UnknownValues = append(request.UnknownValues, tabularingest.ImportUnknownValue{
				SourceColumnOrdinal: column.SourceColumnOrdinal,
				SourceHeaderText:    column.SourceHeaderText,
				RawValue:            rawValue,
				CellKind:            cellKind,
			})
			continue
		}
		transformed, err := transformImportValue(rawValue, column)
		if err != nil {
			return tabularingest.ImportOwnerCreateRequest{}, err
		}
		value, include, err := tabularingest.NormalizeImportScalar(unit.ApprovedMapping.TargetViewSchemaID, *column.FieldKey, transformed, column.EmptyValuePolicy)
		if err != nil {
			return tabularingest.ImportOwnerCreateRequest{}, err
		}
		if !include {
			continue
		}
		var transformID *string
		if column.TransformID != nil {
			transformID = column.TransformID
		}
		var entityBinding *string
		if column.EntityBindingMode != nil {
			entityBinding = column.EntityBindingMode
		}
		request.FieldValues = append(request.FieldValues, tabularingest.ImportFieldValue{
			FieldKey:            *column.FieldKey,
			NormalizedValue:     value,
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			SourceHeaderText:    column.SourceHeaderText,
			RawValue:            rawValue,
			CellKind:            cellKind,
			TransformID:         transformID,
			EmptyValuePolicy:    column.EmptyValuePolicy,
			EntityBindingMode:   entityBinding,
		})
	}
	return request, nil
}

func applyOwnerCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	stores importOwnerStores,
	actorID uuid.UUID,
	target importTarget,
	request tabularingest.ImportOwnerCreateRequest,
	changeSetID uuid.UUID,
	sequenceNo int,
	now time.Time,
) (importOwnerApplyResult, error) {
	values := importValuesByField(request.FieldValues)
	if err := validateImportOwnerCreate(request.TargetViewSchemaID, values); err != nil {
		return importOwnerApplyResult{}, err
	}
	if err := validateImportReferencesTx(ctx, tx, request.IncidentID, values); err != nil {
		return importOwnerApplyResult{}, err
	}
	if request.TargetViewSchemaID == partiesImportViewSchemaID {
		result, found, err := reuseImportedPartyTx(ctx, tx, stores, request, changeSetID, sequenceNo, now)
		if err != nil || found {
			return result, err
		}
	}

	recordType := importRecordTypeForView(request.TargetViewSchemaID)
	if recordType == "" {
		return importOwnerApplyResult{}, importApplyBlockedError("owner_create_contract_unavailable")
	}
	recordID, err := stores.records.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      request.IncidentID,
		RecordType:      recordType,
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return importOwnerApplyResult{}, err
	}
	switch request.TargetViewSchemaID {
	case evidenceImportViewSchemaID:
		if err := stores.evidence.InsertWorkbookRowTx(ctx, tx, recordID, request.IncidentID, evidence.WorkbookCreateParams{Values: evidenceValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	case partiesImportViewSchemaID:
		if err := stores.entities.InsertPartyTx(ctx, tx, recordID, request.IncidentID, entities.PartyCreateParams{Values: partyValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	case taskRequestsImportViewSchemaID:
		if err := stores.tasksDecisions.InsertTaskRequestTx(ctx, tx, recordID, request.IncidentID, actorID, tasksdecisions.TaskCreateParams{Values: taskDecisionValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
		decisionID := importUUIDValue(values, "task.decision_record_id")
		if _, err := stores.links.SyncTaskDecisionReferenceTx(ctx, tx, request.IncidentID, recordID, decisionID, actorID, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	case decisionsImportViewSchemaID:
		if err := stores.tasksDecisions.InsertDecisionTx(ctx, tx, recordID, request.IncidentID, actorID, tasksdecisions.DecisionCreateParams{Values: taskDecisionValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	case assessmentsImportViewSchemaID:
		if err := stores.entities.InsertAssessmentTx(ctx, tx, recordID, request.IncidentID, actorID, entities.AssessmentCreateParams{Values: assessmentValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	default:
		if !artifacts.IsArtifactBackedView(request.TargetViewSchemaID) {
			return importOwnerApplyResult{}, importApplyBlockedError("owner_create_contract_unavailable")
		}
		if err := stores.artifacts.InsertRowTx(ctx, tx, recordID, request.IncidentID, actorID, artifacts.CreateParams{ViewSchemaID: request.TargetViewSchemaID, Values: artifactValuesFromImport(values)}, now); err != nil {
			return importOwnerApplyResult{}, err
		}
	}
	return finalizeImportOwnerCreateTx(ctx, tx, stores, request, target, changeSetID, sequenceNo, recordID, "created", "created", "create", now)
}

func reuseImportedPartyTx(ctx context.Context, tx pgx.Tx, stores importOwnerStores, request tabularingest.ImportOwnerCreateRequest, changeSetID uuid.UUID, sequenceNo int, now time.Time) (importOwnerApplyResult, bool, error) {
	values := importValuesByField(request.FieldValues)
	recordID, found, err := stores.entities.FindReusablePartyTx(ctx, tx, request.IncidentID, entities.PartyCreateParams{Values: partyValuesFromImport(values)})
	if err != nil || !found {
		return importOwnerApplyResult{}, false, err
	}
	result, err := finalizeImportOwnerCreateTx(ctx, tx, stores, request, importTarget{CreateFacade: createFacadeParty}, changeSetID, sequenceNo, recordID, "reused", "reused", "reuse", now)
	return result, true, err
}

func finalizeImportOwnerCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	stores importOwnerStores,
	request tabularingest.ImportOwnerCreateRequest,
	target importTarget,
	changeSetID uuid.UUID,
	sequenceNo int,
	recordID uuid.UUID,
	createdOrReused string,
	resultCode string,
	operation string,
	now time.Time,
) (importOwnerApplyResult, error) {
	row, err := refreshImportOwnerRowTx(ctx, tx, stores, request.TargetViewSchemaID, recordID)
	if err != nil {
		return importOwnerApplyResult{}, err
	}
	rowVersion, err := rowVersionFromImportRow(row)
	if err != nil {
		return importOwnerApplyResult{}, err
	}
	afterVersionID := importVersionID(recordID, rowVersion)
	if err := stores.revisions.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     sequenceNo,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  operation,
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return importOwnerApplyResult{}, err
	}
	if operation == "create" {
		if err := stores.revisions.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    recordID,
			RowVersion:  rowVersion,
			AfterValue:  row,
		}); err != nil {
			return importOwnerApplyResult{}, err
		}
	}
	mutationRef := fmt.Sprintf("change_set_mutation:%s:%d", changeSetID, sequenceNo)
	response := tabularingest.ImportOwnerCreateResponse{
		RecordID:             recordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: mutationRef,
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           row,
	}
	_ = target
	return importOwnerApplyResult{Response: response, Operation: operation}, nil
}

func refreshImportOwnerRowTx(ctx context.Context, tx pgx.Tx, stores importOwnerStores, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	switch viewSchemaID {
	case evidenceImportViewSchemaID:
		return stores.evidence.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
	case partiesImportViewSchemaID:
		return stores.entities.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
	case taskRequestsImportViewSchemaID:
		return stores.tasksDecisions.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
	case decisionsImportViewSchemaID:
		return stores.tasksDecisions.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
	case assessmentsImportViewSchemaID:
		return stores.entities.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
	default:
		if artifacts.IsArtifactBackedView(viewSchemaID) {
			return stores.artifacts.RefreshImportRowTx(ctx, tx, viewSchemaID, recordID)
		}
		return nil, fmt.Errorf("import projection surface %q not mapped", viewSchemaID)
	}
}

func importRecordTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case evidenceImportViewSchemaID:
		return "evidence"
	case partiesImportViewSchemaID:
		return "party"
	case taskRequestsImportViewSchemaID:
		return "task_request"
	case decisionsImportViewSchemaID:
		return "decision"
	case assessmentsImportViewSchemaID:
		return "assessment"
	default:
		if artifacts.IsArtifactBackedView(viewSchemaID) {
			return "artifact"
		}
		return ""
	}
}

func importValuesByField(fields []tabularingest.ImportFieldValue) map[string]tabularingest.ImportScalarValue {
	values := make(map[string]tabularingest.ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}

func artifactValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]artifacts.FieldValue {
	result := make(map[string]artifacts.FieldValue, len(values))
	for field, value := range values {
		result[field] = artifacts.FieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}

func evidenceValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]evidence.WorkbookFieldValue {
	result := make(map[string]evidence.WorkbookFieldValue, len(values))
	for field, value := range values {
		result[field] = evidence.WorkbookFieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}

func partyValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]entities.PartyFieldValue {
	result := make(map[string]entities.PartyFieldValue, len(values))
	for field, value := range values {
		result[field] = entities.PartyFieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}

func taskDecisionValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]tasksdecisions.FieldValue {
	result := make(map[string]tasksdecisions.FieldValue, len(values))
	for field, value := range values {
		result[field] = tasksdecisions.FieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}

func assessmentValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]entities.AssessmentFieldValue {
	result := make(map[string]entities.AssessmentFieldValue, len(values))
	for field, value := range values {
		result[field] = entities.AssessmentFieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}

func validateImportOwnerCreate(viewSchemaID string, values map[string]tabularingest.ImportScalarValue) error {
	switch viewSchemaID {
	case artifacts.NotesViewSchemaID:
		if !hasImportText(values, "note.title") && !hasImportText(values, "note.body") {
			return fmt.Errorf("import create %s: missing minimum create signal", viewSchemaID)
		}
	case evidenceImportViewSchemaID:
		if !hasImportText(values, "evidence.title") &&
			!hasImportText(values, "evidence.storage_ref") &&
			!hasImportText(values, "evidence.collector_party_text") &&
			!hasImportText(values, "evidence.source_party_text") {
			return fmt.Errorf("import create %s: missing minimum create signal", viewSchemaID)
		}
		if value, ok := values["evidence.lifecycle_state"]; ok && !evidence.ValidLifecycleState(importText(value)) {
			return fmt.Errorf("import create evidence.lifecycle_state: invalid value")
		}
	case partiesImportViewSchemaID:
		if !hasImportText(values, "party.display_name") {
			return fmt.Errorf("import create party.display_name: missing required field")
		}
		if !validImportText(values, "party.party_kind", validPartyKind) {
			return fmt.Errorf("import create party.party_kind: missing required field")
		}
	case artifacts.CommLogViewSchemaID:
		for _, field := range []string{"comm_log.comm_type", "comm_log.audience", "comm_log.channel_or_meeting", "comm_log.summary"} {
			if !hasImportText(values, field) {
				return fmt.Errorf("import create %s: missing required field", field)
			}
		}
		if !validImportText(values, "comm_log.comm_type", validCommType) {
			return fmt.Errorf("import create comm_log.comm_type: invalid value")
		}
	case artifacts.HandoffViewSchemaID:
		if !hasImportUUID(values, "handoff.incoming_owner_user_id") {
			return fmt.Errorf("import create handoff.incoming_owner_user_id: missing required field")
		}
		if !hasImportText(values, "handoff.current_state_summary") {
			return fmt.Errorf("import create handoff.current_state_summary: missing required field")
		}
	case artifacts.StatusReviewViewSchemaID:
		if !hasImportText(values, "status_review.current_state_summary") {
			return fmt.Errorf("import create status_review.current_state_summary: missing required field")
		}
	case artifacts.LessonViewSchemaID:
		if !hasImportText(values, "lesson.summary") {
			return fmt.Errorf("import create lesson.summary: missing required field")
		}
		if value, ok := values["lesson.closure_state"]; ok && !validClosureState(importText(value)) {
			return fmt.Errorf("import create lesson.closure_state: invalid value")
		}
	case taskRequestsImportViewSchemaID:
		if !hasImportText(values, "task.title") {
			return fmt.Errorf("import create task.title: missing required field")
		}
		if !validImportText(values, "task.task_kind", validTaskKind) {
			return fmt.Errorf("import create task.task_kind: missing required field")
		}
		if value, ok := values["task.status"]; ok && !validTaskStatus(importText(value)) {
			return fmt.Errorf("import create task.status: invalid value")
		}
		if value, ok := values["task.priority"]; ok && !validTaskPriority(importText(value)) {
			return fmt.Errorf("import create task.priority: invalid value")
		}
	case decisionsImportViewSchemaID:
		if !hasImportText(values, "decision.summary") {
			return fmt.Errorf("import create decision.summary: missing required field")
		}
		if !validImportText(values, "decision.decision_type", validDecisionType) {
			return fmt.Errorf("import create decision.decision_type: missing required field")
		}
		if !hasImportText(values, "decision.rationale") {
			return fmt.Errorf("import create decision.rationale: missing required field")
		}
		if value, ok := values["decision.status"]; ok {
			status := importText(value)
			if !validDecisionStatus(status) {
				return fmt.Errorf("import create decision.status: invalid value")
			}
			if status == "superseded" {
				return fmt.Errorf("import create decision.status: superseded direct write")
			}
		}
	case assessmentsImportViewSchemaID:
		for _, field := range []string{"assessment.subject_ref", "assessment.subject_type", "assessment.assessment_state", "assessment.rationale"} {
			if field == "assessment.subject_ref" {
				if !hasImportUUID(values, field) {
					return fmt.Errorf("import create %s: missing required field", field)
				}
				continue
			}
			if !hasImportText(values, field) {
				return fmt.Errorf("import create %s: missing required field", field)
			}
		}
		if value, ok := values["assessment.confidence_score"]; ok && value.Number != nil && !validConfidenceScore(*value.Number) {
			return fmt.Errorf("import create assessment.confidence_score: invalid value")
		}
	default:
		schema, ok := viewschema.Lookup(viewSchemaID)
		if ok && !schema.PermitsZeroFieldCreate && len(values) == 0 {
			return fmt.Errorf("import create %s: missing minimum create signal", viewSchemaID)
		}
	}
	return nil
}

func validateImportReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, values map[string]tabularingest.ImportScalarValue) error {
	for fieldKey, value := range values {
		if value.UUID == nil {
			continue
		}
		if strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateImportActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
		switch fieldKey {
		case "task.requester_party_id", "evidence.collector_party_id", "evidence.source_party_id":
			if err := validateImportTargetRecordTx(ctx, tx, incidentID, *value.UUID, "party", fieldKey); err != nil {
				return err
			}
		case "task.decision_record_id":
			if err := validateImportTargetRecordTx(ctx, tx, incidentID, *value.UUID, "decision", fieldKey); err != nil {
				return err
			}
		case "assessment.subject_ref":
			if err := validateImportTargetRecordTx(ctx, tx, incidentID, *value.UUID, "", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateImportActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("validate import user: %w", err)
	}
	if !exists {
		return fmt.Errorf("import create %s: invalid value", field)
	}
	return nil
}

func validateImportTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if expectedType == "" {
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return fmt.Errorf("validate import reference: %w", err)
		}
	} else if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate import reference: %w", err)
	}
	if !exists {
		return fmt.Errorf("import create %s: invalid value", field)
	}
	return nil
}

func rowVersionFromImportRow(row map[string]any) (int64, error) {
	switch value := row["row_version"].(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("import row has unexpected row_version type %T", value)
	}
}

func importVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func importUUIDValue(values map[string]tabularingest.ImportScalarValue, field string) *uuid.UUID {
	if value, ok := values[field]; ok {
		return value.UUID
	}
	return nil
}

func hasImportText(values map[string]tabularingest.ImportScalarValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func hasImportUUID(values map[string]tabularingest.ImportScalarValue, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil && *value.UUID != uuid.Nil
}

func validImportText(values map[string]tabularingest.ImportScalarValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return false
	}
	return predicate(*value.Text)
}

func importText(value tabularingest.ImportScalarValue) string {
	if value.Text == nil {
		return ""
	}
	return *value.Text
}

func validPartyKind(value string) bool {
	switch value {
	case "person", "team", "vendor", "system", "other":
		return true
	default:
		return false
	}
}

func validCommType(value string) bool {
	switch value {
	case "status_update", "stakeholder_update", "handoff", "executive_brief", "customer_notice", "regulator_notice", "other":
		return true
	default:
		return false
	}
}

func validClosureState(value string) bool {
	switch value {
	case "open", "closed", "":
		return true
	default:
		return false
	}
}

func validTaskKind(value string) bool {
	return tasksdecisions.ValidTaskKind(value)
}

func validTaskStatus(value string) bool {
	return tasksdecisions.ValidTaskStatus(value)
}

func validTaskPriority(value string) bool {
	return tasksdecisions.ValidTaskPriority(value)
}

func validDecisionType(value string) bool {
	return tasksdecisions.ValidDecisionType(value)
}

func validDecisionStatus(value string) bool {
	return tasksdecisions.ValidDecisionStatus(value)
}

func validConfidenceScore(value int64) bool {
	return value >= 0 && value <= 100
}
