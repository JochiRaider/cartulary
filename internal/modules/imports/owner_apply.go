package imports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
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
	assessments    *assessments.Store
	hostidentity   *hostidentity.Store
	evidence       *evidence.Store
	indicators     *indicators.Store
	parties        *parties.Store
	tasksDecisions *tasksdecisions.Store
}

type importOwnerApplyResult struct {
	Response  ownerfacade.ImportOwnerCreateResponse
	Operation string
}

func (s *Service) applyGenericOwnerUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData, target importTarget) error {
	now := s.now().UTC()
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin import apply unit transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, start.IncidentID); err != nil {
		return err
	}
	clientTxnID := fmt.Sprintf("import:%s:%s:%s", start.ImportSessionID, unit.UnitID, start.ClientTxnID)
	requestID := "req-" + clientTxnID
	changeSetID, err := newRevisionAppendAdapter().AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
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
		assessments:    assessments.NewStore(s.store.pool),
		hostidentity:   hostidentity.NewStore(s.store.pool),
		evidence:       evidence.NewStore(s.store.pool),
		indicators:     indicators.NewStore(s.store.pool),
		parties:        parties.NewStore(s.store.pool),
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
		result, err := applyOwnerCreateTx(ctx, tx, stores, request, changeSetID, index+1, now)
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

func importOwnerCreateRequest(start ApplyStartResult, unit ApplyUnitData, actorID uuid.UUID, sourceRow map[string]any, rowRef int, clientTxnID string) (ownerfacade.ImportOwnerCreateRequest, error) {
	cells := sourceRowCellsByOrdinal(sourceRow)
	request := ownerfacade.ImportOwnerCreateRequest{
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
		SourceRowProvenance: ownerfacade.ImportSourceRowProvenance{SourceRowRef: rowRef},
	}
	for _, column := range unit.ApprovedMapping.SourceColumns {
		cell := cells[column.SourceColumnOrdinal]
		rawValue, _ := cell["display_text"].(string)
		cellKind, _ := cell["cell_kind"].(string)
		if column.FieldKey == nil {
			request.UnknownValues = append(request.UnknownValues, ownerfacade.ImportUnknownValue{
				SourceColumnOrdinal: column.SourceColumnOrdinal,
				SourceHeaderText:    column.SourceHeaderText,
				RawValue:            rawValue,
				CellKind:            cellKind,
			})
			continue
		}
		transformed, err := transformImportValue(rawValue, column)
		if err != nil {
			return ownerfacade.ImportOwnerCreateRequest{}, err
		}
		value, include, err := ownerfacade.NormalizeImportScalar(unit.ApprovedMapping.TargetViewSchemaID, *column.FieldKey, transformed, column.EmptyValuePolicy)
		if err != nil {
			return ownerfacade.ImportOwnerCreateRequest{}, err
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
		request.FieldValues = append(request.FieldValues, ownerfacade.ImportFieldValue{
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
	request ownerfacade.ImportOwnerCreateRequest,
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
		response, err := stores.parties.CreateImportRowTx(ctx, tx, parties.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		operation := "create"
		if response.CreatedOrReused == "reused" {
			operation = "reuse"
		}
		return importOwnerApplyResult{Response: response, Operation: operation}, err
	}
	if request.TargetViewSchemaID == assessmentsImportViewSchemaID {
		response, err := stores.assessments.CreateImportRowTx(ctx, tx, assessments.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		return importOwnerApplyResult{Response: response, Operation: "create"}, err
	}
	if request.TargetViewSchemaID == hostidentity.HostsViewSchemaID || request.TargetViewSchemaID == hostidentity.IdentitiesViewSchemaID {
		response, err := stores.hostidentity.CreateImportRowTx(ctx, tx, hostidentity.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		operation := "create"
		if response.CreatedOrReused == "reused" {
			operation = "reuse"
		}
		return importOwnerApplyResult{Response: response, Operation: operation}, err
	}
	if request.TargetViewSchemaID == indicators.ViewSchemaID {
		response, err := stores.indicators.CreateImportRowTx(ctx, tx, indicators.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		return importOwnerApplyResult{Response: response, Operation: "create"}, err
	}
	if request.TargetViewSchemaID == evidenceImportViewSchemaID {
		response, err := stores.evidence.CreateImportRowTx(ctx, tx, evidence.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		return importOwnerApplyResult{Response: response, Operation: "create"}, err
	}
	if request.TargetViewSchemaID == taskRequestsImportViewSchemaID || request.TargetViewSchemaID == decisionsImportViewSchemaID {
		response, err := stores.tasksDecisions.CreateImportRowTx(ctx, tx, tasksdecisions.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		return importOwnerApplyResult{Response: response, Operation: "create"}, err
	}
	if artifacts.IsArtifactBackedView(request.TargetViewSchemaID) {
		response, err := stores.artifacts.CreateImportRowTx(ctx, tx, artifacts.ImportCreateCommand{
			Request:     request,
			ChangeSetID: changeSetID,
			SequenceNo:  sequenceNo,
			Now:         now,
		})
		return importOwnerApplyResult{Response: response, Operation: "create"}, err
	}

	return importOwnerApplyResult{}, importApplyBlockedError("owner_create_contract_unavailable")
}

func importValuesByField(fields []ownerfacade.ImportFieldValue) map[string]ownerfacade.ImportScalarValue {
	values := make(map[string]ownerfacade.ImportScalarValue, len(fields))
	for _, field := range fields {
		values[field.FieldKey] = field.NormalizedValue
	}
	return values
}

func validateImportOwnerCreate(viewSchemaID string, values map[string]ownerfacade.ImportScalarValue) error {
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
	case indicators.ViewSchemaID:
		for _, field := range []string{"indicator.indicator_type", "indicator.value_kind", "indicator.display_value"} {
			if !hasImportText(values, field) {
				return fmt.Errorf("import create %s: missing required field", field)
			}
		}
	case partiesImportViewSchemaID:
		if !hasImportText(values, "party.display_name") {
			return fmt.Errorf("import create party.display_name: missing required field")
		}
		if !validImportText(values, "party.party_kind", parties.ValidKind) {
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

func validateImportReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, values map[string]ownerfacade.ImportScalarValue) error {
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

func hasImportText(values map[string]ownerfacade.ImportScalarValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func hasImportUUID(values map[string]ownerfacade.ImportScalarValue, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil && *value.UUID != uuid.Nil
}

func validImportText(values map[string]ownerfacade.ImportScalarValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return false
	}
	return predicate(*value.Text)
}

func importText(value ownerfacade.ImportScalarValue) string {
	if value.Text == nil {
		return ""
	}
	return *value.Text
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
