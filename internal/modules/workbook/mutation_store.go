package workbook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *Store) CreateWorkbookRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + request.ViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: incidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook create idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := incidents.EnsureIncidentOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	if err := validateCreateReferencesTx(ctx, tx, incidentID, request); err != nil {
		return MutationResult{}, err
	}

	if request.ViewSchemaID == PartiesViewSchemaID {
		result, reused, err := s.reusePartyCreateTx(ctx, tx, actor, incidentID, request, idempotencyKey, requestHash, requestID, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		if reused {
			if err := tx.Commit(ctx); err != nil {
				return MutationResult{}, fmt.Errorf("commit workbook party reuse transaction: %w", err)
			}
			return result, nil
		}
	}

	recordType := recordTypeForView(request.ViewSchemaID)
	if recordType == "" {
		return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
	}
	recordID, err := s.recordStore.InsertTx(ctx, tx, recordsInsertParams(incidentID, recordType, actor.ID, now.UTC()))
	if err != nil {
		return MutationResult{}, err
	}
	switch request.ViewSchemaID {
	case EvidenceViewSchemaID:
		if err := s.evidenceStore.InsertWorkbookRowTx(ctx, tx, recordID, incidentID, evidence.WorkbookCreateParams{Values: evidenceValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	case PartiesViewSchemaID:
		if err := s.entityStore.InsertPartyTx(ctx, tx, recordID, incidentID, entities.PartyCreateParams{Values: partyValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	case TaskRequestsViewSchemaID:
		if err := s.taskStore.InsertTaskRequestTx(ctx, tx, recordID, incidentID, actor.ID, tasksdecisions.TaskCreateParams{Values: taskDecisionValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
			return MutationResult{}, adaptOwnerMutationError(err)
		}
	case DecisionsViewSchemaID:
		if err := s.taskStore.InsertDecisionTx(ctx, tx, recordID, incidentID, actor.ID, tasksdecisions.DecisionCreateParams{Values: taskDecisionValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
			return MutationResult{}, adaptOwnerMutationError(err)
		}
	default:
		if err := s.artifactStore.InsertRowTx(ctx, tx, recordID, incidentID, actor.ID, artifacts.CreateParams{ViewSchemaID: request.ViewSchemaID, Values: artifactValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if err := s.applyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actor.ID, request.Collections, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID {
		decisionID := nullableUUIDValue(request.Values, "task.decision_record_id")
		if _, err := s.linkStore.SyncTaskDecisionReferenceTx(ctx, tx, incidentID, recordID, nullableUUIDPointer(decisionID), actor.ID, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, request.ViewSchemaID, recordID); err != nil {
		return MutationResult{}, err
	}

	row, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook create transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (s *Store) reusePartyCreateTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time) (MutationResult, bool, error) {
	recordID, found, err := s.entityStore.FindReusablePartyTx(ctx, tx, incidentID, entities.PartyCreateParams{Values: partyValuesFromWorkbook(request.Values)})
	if err != nil || !found {
		return MutationResult{}, false, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, PartiesViewSchemaID, recordID); err != nil {
		return MutationResult{}, false, err
	}
	row, err := s.loadGenericRowTx(ctx, tx, PartiesViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, false, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, false, err
	}
	rowVersion, err := rowVersionFromGenericRow(row)
	if err != nil {
		return MutationResult{}, false, err
	}
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "reuse",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, false, err
	}
	payload := BuildMutationPayload(PartiesViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, false, authn.ErrClientTxnConflict
		}
		return MutationResult{}, false, err
	}
	return MutationResult{
		Payload:      payload,
		StatusCode:   http.StatusOK,
		IncidentID:   incidentID,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		ClientTxnID:  request.ClientTxnID,
		RowVersion:   rowVersion,
		ViewSchemaID: PartiesViewSchemaID,
	}, true, nil
}

func rowVersionFromGenericRow(row map[string]any) (int64, error) {
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
		return 0, fmt.Errorf("generic row has unexpected row_version type %T", value)
	}
}

func adaptOwnerMutationError(err error) error {
	if err == nil {
		return nil
	}
	var evidenceValidation *evidence.ValidationError
	if errors.As(err, &evidenceValidation) {
		return mutationValidationError(evidenceValidation.Field, evidenceValidation.ReasonCode)
	}
	var evidenceLifecycle *evidence.LifecycleValidationError
	if errors.As(err, &evidenceLifecycle) {
		return &LifecycleValidationError{
			FromStatus:     evidenceLifecycle.FromStatus,
			ToStatus:       evidenceLifecycle.ToStatus,
			ReasonCode:     evidenceLifecycle.ReasonCode,
			ViolatedGuards: append([]string(nil), evidenceLifecycle.ViolatedGuards...),
		}
	}
	var taskValidation *tasksdecisions.ValidationError
	if errors.As(err, &taskValidation) {
		return mutationValidationError(taskValidation.Field, taskValidation.ReasonCode)
	}
	var taskLifecycle *tasksdecisions.LifecycleValidationError
	if errors.As(err, &taskLifecycle) {
		return &LifecycleValidationError{
			FromStatus:     taskLifecycle.FromStatus,
			ToStatus:       taskLifecycle.ToStatus,
			ReasonCode:     taskLifecycle.ReasonCode,
			ViolatedGuards: append([]string(nil), taskLifecycle.ViolatedGuards...),
		}
	}
	return err
}

func adaptRevisionConflictError(err error) error {
	if err == nil {
		return nil
	}
	var rowConflict *revisions.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	return err
}

func workbookPatchChangesFromWorkbook(changes []PatchChange) []revisions.WorkbookPatchChange {
	result := make([]revisions.WorkbookPatchChange, 0, len(changes))
	for _, change := range changes {
		result = append(result, workbookPatchChangeFromWorkbook(change))
	}
	return result
}

func workbookPatchChangeFromWorkbook(change PatchChange) revisions.WorkbookPatchChange {
	converted := revisions.WorkbookPatchChange{
		FieldKey: change.FieldKey,
	}
	if change.Value != nil {
		converted.Value = canonicalValue(*change.Value)
	}
	if change.Collection != nil {
		converted.Collection = &revisions.WorkbookCollectionActionPayload{Actions: workbookCollectionActionsFromWorkbook(change.Collection.Actions)}
	}
	return converted
}

func workbookCollectionActionsFromWorkbook(actions []CollectionAction) []revisions.WorkbookCollectionAction {
	result := make([]revisions.WorkbookCollectionAction, 0, len(actions))
	for _, action := range actions {
		result = append(result, workbookCollectionActionFromWorkbook(action))
	}
	return result
}

func workbookCollectionActionFromWorkbook(action CollectionAction) revisions.WorkbookCollectionAction {
	return revisions.WorkbookCollectionAction{
		Op:             action.Op,
		RawText:        action.RawText,
		LinkedRecordID: action.LinkedRecordID,
		PartyID:        action.PartyID,
		ItemRef:        action.ItemRef,
		RiskRefText:    action.RiskRefText,
		NormalizedText: action.NormalizedText,
	}
}

func artifactValuesFromWorkbook(values map[string]ValueChange) map[string]artifacts.FieldValue {
	result := make(map[string]artifacts.FieldValue, len(values))
	for field, value := range values {
		result[field] = artifactValueFromWorkbook(value)
	}
	return result
}

func artifactValueFromWorkbook(value ValueChange) artifacts.FieldValue {
	return artifacts.FieldValue{
		Text:      value.Text,
		Timestamp: value.Timestamp,
		UUID:      value.UUID,
		Number:    value.Number,
		Bool:      value.Bool,
	}
}

func evidenceValuesFromWorkbook(values map[string]ValueChange) map[string]evidence.WorkbookFieldValue {
	result := make(map[string]evidence.WorkbookFieldValue, len(values))
	for field, value := range values {
		result[field] = evidenceValueFromWorkbook(value)
	}
	return result
}

func evidenceValueFromWorkbook(value ValueChange) evidence.WorkbookFieldValue {
	return evidence.WorkbookFieldValue{
		Text:      value.Text,
		Timestamp: value.Timestamp,
		UUID:      value.UUID,
		Number:    value.Number,
		Bool:      value.Bool,
	}
}

func partyValuesFromWorkbook(values map[string]ValueChange) map[string]entities.PartyFieldValue {
	result := make(map[string]entities.PartyFieldValue, len(values))
	for field, value := range values {
		result[field] = partyValueFromWorkbook(value)
	}
	return result
}

func partyValueFromWorkbook(value ValueChange) entities.PartyFieldValue {
	return entities.PartyFieldValue{
		Text:      value.Text,
		Timestamp: value.Timestamp,
		UUID:      value.UUID,
		Number:    value.Number,
		Bool:      value.Bool,
	}
}

func taskDecisionValuesFromWorkbook(values map[string]ValueChange) map[string]tasksdecisions.FieldValue {
	result := make(map[string]tasksdecisions.FieldValue, len(values))
	for field, value := range values {
		result[field] = taskDecisionValueFromWorkbook(value)
	}
	return result
}

func taskDecisionValueFromWorkbook(value ValueChange) tasksdecisions.FieldValue {
	return tasksdecisions.FieldValue{
		Text:      value.Text,
		Timestamp: value.Timestamp,
		UUID:      value.UUID,
		Number:    value.Number,
		Bool:      value.Bool,
	}
}

func (s *Store) CreateLinkedNote(ctx context.Context, actor authn.UserRecord, sourceRecordID uuid.UUID, request LinkedNoteCreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookLinkedNoteRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    sourceRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed linked note payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: NotesViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query linked note idempotency: %w", err)
	}
	create := CreateRequest{
		ViewSchemaID: NotesViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       request.Values,
		Collections:  request.Collections,
	}
	if err := validateCreateRequest(create); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin linked note transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	incidentID, err := loadLinkedNoteSourceIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := incidents.EnsureIncidentOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	recordID, err := s.recordStore.InsertTx(ctx, tx, recordsInsertParams(incidentID, "artifact", actor.ID, now.UTC()))
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.artifactStore.InsertRowTx(ctx, tx, recordID, incidentID, actor.ID, artifacts.CreateParams{ViewSchemaID: create.ViewSchemaID, Values: artifactValuesFromWorkbook(create.Values)}, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.applyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actor.ID, create.Collections, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.linkStore.InsertLinkedNoteReferenceTx(ctx, tx, incidentID, sourceRecordID, recordID, actor.ID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, NotesViewSchemaID, recordID); err != nil {
		return MutationResult{}, err
	}
	row, err := s.loadGenericRowTx(ctx, tx, NotesViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookLinkedNoteRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	linkAfter := map[string]any{
		"src_record_id": sourceRecordID.String(),
		"dst_record_id": recordID.String(),
		"link_type":     "references_artifact",
	}
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    2,
		TargetKind:    "record_link",
		TargetID:      sourceRecordID.String() + ":references_artifact:" + recordID.String(),
		OperationKind: "create",
		AfterValue:    linkAfter,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(NotesViewSchemaID, changeSetID, row)
	payload["source_record_id"] = sourceRecordID.String()
	payload["link_type"] = "references_artifact"
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit linked note transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     NotesViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (s *Store) LinkedNoteSourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	incidentID, err := loadLinkedNoteSourceIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return incidentID, nil
}

func (s *Store) PatchWorkbookRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if len(request.Changes) == 0 {
		return MutationResult{}, mutationValidationError("changes", "empty_changes")
	}
	if request.ViewSchemaID == entities.HostsViewSchemaID || request.ViewSchemaID == entities.IdentitiesViewSchemaID {
		entityRequest, err := entityPatchRequestFromWorkbook(request)
		if err != nil {
			return MutationResult{}, err
		}
		result, err := s.entityStore.PatchEntityRow(ctx, actor, recordID, entityRequest, requestHash, requestID, now, workbookPatchRouteKey)
		var entityConflict *entities.RowVersionConflictError
		switch {
		case errors.As(err, &entityConflict):
			return MutationResult{}, &RowVersionConflictError{
				RecordID:          entityConflict.RecordID,
				BaseRowVersion:    entityConflict.BaseRowVersion,
				CurrentRowVersion: entityConflict.CurrentRowVersion,
			}
		case errors.Is(err, entities.ErrNoEffectivePatchChange):
			return MutationResult{}, mutationValidationError("changes", "no_effective_change")
		case err != nil:
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:          result.Payload,
			StatusCode:       result.StatusCode,
			Replayed:         result.Replayed,
			IncidentID:       result.IncidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     request.ViewSchemaID,
			ChangedFieldKeys: result.ChangedFieldKeys,
		}, nil
	}
	return s.applyWorkbookPatch(ctx, actor, recordID, request, requestHash, requestID, now, workbookPatchRouteKey)
}

func entityPatchRequestFromWorkbook(request PatchRequest) (entities.PatchRequest, error) {
	changes := make([]entities.PatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		if !isEntityDirectPatchField(request.ViewSchemaID, change.FieldKey) {
			return entities.PatchRequest{}, mutationValidationError("field_key", "unsupported_field_key")
		}
		if change.Collection != nil || change.Value == nil {
			return entities.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
		}
		var value *string
		switch change.Value.Kind {
		case "text":
			if change.Value.Text == nil {
				return entities.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
			}
			value = change.Value.Text
		case "null":
			value = nil
		default:
			return entities.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
		}
		changes = append(changes, entities.PatchChange{
			FieldKey: change.FieldKey,
			Value:    value,
		})
	}
	return entities.PatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}, nil
}

func isEntityDirectPatchField(viewSchemaID string, fieldKey string) bool {
	switch viewSchemaID {
	case entities.HostsViewSchemaID:
		switch fieldKey {
		case "host.display_name", "host.hostname", "host.aad_device_id", "host.fqdn",
			"host.location", "host.os_platform", "host.business_owner", "host.criticality", "host.containment_status":
			return true
		default:
			return false
		}
	case entities.IdentitiesViewSchemaID:
		switch fieldKey {
		case "identity.display_name", "identity.aad_object_id", "identity.sid", "identity.upn", "identity.email", "identity.sam_account_name",
			"identity.privilege_level", "identity.mfa_state", "identity.reset_status":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (s *Store) applyWorkbookPatch(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook patch payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	if err := incidents.EnsureIncidentOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return MutationResult{}, err
	}
	effectiveBeforeVersion := request.BaseRowVersion
	if meta.RowVersion != request.BaseRowVersion {
		if meta.RowVersion < request.BaseRowVersion {
			return MutationResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		window, err := s.revisionStore.LoadWorkbookPatchConflictWindowTx(ctx, tx, recordID, request.ViewSchemaID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return MutationResult{}, adaptRevisionConflictError(err)
		}
		if change, changed, ok := revisions.OverlappingWorkbookPatchChange(workbookPatchChangesFromWorkbook(request.Changes), window.ChangedFields); ok {
			current, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflictPayload, err := revisions.BuildWorkbookSameFieldConflict(revisions.SameFieldConflictParams{
				RouteKey:          workbookConflictResolveRouteKey,
				RecordID:          recordID,
				ViewSchemaID:      request.ViewSchemaID,
				BaseRowVersion:    request.BaseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       requestHash,
				Window:            window,
				Change:            change,
				Changed:           changed,
				CurrentRow:        current,
				Codec:             s.conflictTokens,
			})
			if err != nil {
				return MutationResult{}, adaptRevisionConflictError(err)
			}
			return MutationResult{}, &SameFieldConflictError{Conflict: conflictPayload}
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	if err := s.validatePatchLifecycleTx(ctx, tx, recordID, request); err != nil {
		return MutationResult{}, adaptOwnerMutationError(err)
	}
	changed, err := s.applyPatchTx(ctx, tx, meta.IncidentID, recordID, actor.ID, request, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, mutationValidationError("changes", "no_effective_change")
	}
	rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.touchSourceRowTx(ctx, tx, request.ViewSchemaID, recordID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, request.ViewSchemaID, recordID); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := workbookVersionID(recordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = workbookVersionID(recordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        recordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (s *Store) ResolveWorkbookConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims workbookConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ResolutionKind == "keep_saved" {
		return s.clearWorkbookConflict(ctx, actor, recordID, claims, request, requestHash)
	}
	if request.ResolvedChange == nil {
		return MutationResult{}, mutationValidationError("resolved_value", "missing_required_field")
	}
	patch := PatchRequest{
		ViewSchemaID:   claims.ViewSchemaID,
		BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        []PatchChange{*request.ResolvedChange},
	}
	return s.applyWorkbookPatch(ctx, actor, recordID, patch, requestHash, requestID, now, workbookConflictResolveRouteKey)
}

func (s *Store) clearWorkbookConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims workbookConflictTokenClaims, request ConflictResolveRequest, requestHash []byte) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookConflictResolveRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook conflict clear payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: claims.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook conflict clear idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook conflict clear transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, claims.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	row, err := s.loadGenericRowTx(ctx, tx, claims.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": claims.ViewSchemaID,
		"row":            row,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook conflict clear transaction: %w", err)
	}
	return MutationResult{
		Payload:      payload,
		StatusCode:   http.StatusOK,
		IncidentID:   meta.IncidentID,
		RecordID:     recordID,
		ClientTxnID:  request.ClientTxnID,
		RowVersion:   meta.RowVersion,
		ViewSchemaID: claims.ViewSchemaID,
	}, nil
}

func (s *Store) SupersedeDecision(ctx context.Context, actor authn.UserRecord, targetRecordID uuid.UUID, request timeline.SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ReplacementRecordID == nil {
		return MutationResult{}, mutationValidationError("replacement_record_id", "missing_required_field")
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookSupersedeRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    targetRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed decision supersede payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: targetRecordID, ViewSchemaID: DecisionsViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query decision supersede idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin decision supersede transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	targetMeta, err := loadRecordMetaForUpdateTx(ctx, tx, targetRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if targetMeta.RecordType != "decision" {
		return MutationResult{}, pgx.ErrNoRows
	}
	if err := incidents.EnsureIncidentOpenTx(ctx, tx, targetMeta.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if targetMeta.RowVersion != request.BaseRowVersion {
		return MutationResult{}, &RowVersionConflictError{RecordID: targetRecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: targetMeta.RowVersion}
	}

	sourceRecordID := *request.ReplacementRecordID
	if sourceRecordID == targetRecordID {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("superseding_decision_must_be_different"))
	}
	sourceMeta, err := loadRecordMetaForUpdateTx(ctx, tx, sourceRecordID)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision"))
	}
	if err != nil {
		return MutationResult{}, err
	}
	if sourceMeta.RecordType != "decision" || sourceMeta.IncidentID != targetMeta.IncidentID {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("superseding_decision_must_be_active_same_incident_decision"))
	}

	targetState, err := s.taskStore.LoadDecisionMachineStateForUpdateTx(ctx, tx, targetRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	sourceState, err := s.taskStore.LoadDecisionMachineStateForUpdateTx(ctx, tx, sourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := tasksdecisions.ValidateDecisionMachineState(targetState); err != nil {
		return MutationResult{}, adaptOwnerMutationError(err)
	}
	if err := tasksdecisions.ValidateDecisionMachineState(sourceState); err != nil {
		return MutationResult{}, adaptOwnerMutationError(err)
	}
	if sourceState.Status != "approved" && sourceState.Status != "executed" {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("superseding_decision_must_be_approved_or_executed"))
	}
	if targetState.Status != "proposed" && targetState.Status != "approved" && targetState.Status != "executed" {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("target_decision_must_be_proposed_approved_or_executed"))
	}
	if targetState.IncomingSupersedes > 0 {
		return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("target_must_not_have_active_replacement"))
	}

	if err := s.refreshWorkbookProjectionTx(ctx, tx, DecisionsViewSchemaID, targetRecordID); err != nil {
		return MutationResult{}, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, DecisionsViewSchemaID, sourceRecordID); err != nil {
		return MutationResult{}, err
	}
	beforeTargetRow, err := s.loadGenericRowTx(ctx, tx, DecisionsViewSchemaID, targetRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeSourceRow, err := s.loadGenericRowTx(ctx, tx, DecisionsViewSchemaID, sourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}

	linkID, err := s.taskStore.InsertDecisionSupersedesLinkTx(ctx, tx, targetMeta.IncidentID, sourceRecordID, targetRecordID, actor.ID, now.UTC())
	if err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, adaptOwnerMutationError(tasksdecisions.DecisionSupersedeValidationError("target_must_not_have_active_replacement"))
		}
		return MutationResult{}, err
	}

	sourceVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, sourceRecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	targetVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, targetRecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.taskStore.TouchSupersedingDecisionTx(ctx, tx, sourceRecordID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.taskStore.MarkSupersededDecisionTx(ctx, tx, targetRecordID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, DecisionsViewSchemaID, sourceRecordID); err != nil {
		return MutationResult{}, err
	}
	if err := s.refreshWorkbookProjectionTx(ctx, tx, DecisionsViewSchemaID, targetRecordID); err != nil {
		return MutationResult{}, err
	}
	afterSourceRow, err := s.loadGenericRowTx(ctx, tx, DecisionsViewSchemaID, sourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterTargetRow, err := s.loadGenericRowTx(ctx, tx, DecisionsViewSchemaID, targetRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  targetMeta.IncidentID,
		ActorUserID: actor.ID,
		Source:      workbookSupersedeRouteKey,
		Reason:      &request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	sourceBeforeVersionID := workbookVersionID(sourceRecordID, sourceMeta.RowVersion)
	sourceAfterVersionID := workbookVersionID(sourceRecordID, sourceVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        sourceRecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &sourceBeforeVersionID,
		AfterVersionID:  &sourceAfterVersionID,
		BeforeValue:     beforeSourceRow,
		AfterValue:      afterSourceRow,
	}); err != nil {
		return MutationResult{}, err
	}
	targetBeforeVersionID := workbookVersionID(targetRecordID, targetMeta.RowVersion)
	targetAfterVersionID := workbookVersionID(targetRecordID, targetVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      2,
		TargetKind:      "record",
		TargetID:        targetRecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &targetBeforeVersionID,
		AfterVersionID:  &targetAfterVersionID,
		BeforeValue:     beforeTargetRow,
		AfterValue:      afterTargetRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    3,
		TargetKind:    "record_link",
		TargetID:      linkID.String(),
		OperationKind: "create",
		AfterValue: map[string]any{
			"record_link_id": linkID.String(),
			"incident_id":    targetMeta.IncidentID.String(),
			"src_record_id":  sourceRecordID.String(),
			"dst_record_id":  targetRecordID.String(),
			"link_type":      "supersedes",
		},
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    sourceRecordID,
		RowVersion:  sourceVersion,
		BeforeValue: beforeSourceRow,
		AfterValue:  afterSourceRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    targetRecordID,
		RowVersion:  targetVersion,
		BeforeValue: beforeTargetRow,
		AfterValue:  afterTargetRow,
	}); err != nil {
		return MutationResult{}, err
	}

	targetStatus := decisionRowStatus(afterTargetRow)
	payload := map[string]any{
		"view_schema_id":          DecisionsViewSchemaID,
		"change_set_id":           changeSetID.String(),
		"target_record_id":        targetRecordID.String(),
		"superseding_record_id":   sourceRecordID.String(),
		"target_row_version":      targetVersion,
		"superseding_row_version": sourceVersion,
		"target_status":           targetStatus,
		"reason":                  request.Reason,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit decision supersede transaction: %w", err)
	}

	targetChange := MutationResult{
		Payload:          map[string]any{"row": afterTargetRow},
		StatusCode:       http.StatusOK,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         targetRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       targetVersion,
		ViewSchemaID:     DecisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeTargetRow, afterTargetRow),
	}
	sourceChange := MutationResult{
		Payload:          map[string]any{"row": afterSourceRow},
		StatusCode:       http.StatusOK,
		IncidentID:       targetMeta.IncidentID,
		RecordID:         sourceRecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       sourceVersion,
		ViewSchemaID:     DecisionsViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeSourceRow, afterSourceRow),
	}
	return MutationResult{
		Payload:                 payload,
		StatusCode:              http.StatusOK,
		IncidentID:              targetMeta.IncidentID,
		RecordID:                targetRecordID,
		ChangeSetID:             changeSetID,
		ClientTxnID:             request.ClientTxnID,
		RowVersion:              targetVersion,
		ViewSchemaID:            DecisionsViewSchemaID,
		ChangedFieldKeys:        targetChange.ChangedFieldKeys,
		AdditionalRecordChanges: []MutationResult{targetChange, sourceChange},
	}, nil
}

func (s *Store) RecordIncident(ctx context.Context, recordID uuid.UUID, viewSchemaID string) (uuid.UUID, error) {
	var incidentID uuid.UUID
	if viewSchemaID == "cartulary.view.timeline.v2" {
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type = 'timeline_event'
`, recordID).Scan(&incidentID)
		return incidentID, err
	}
	recordType := recordTypeForView(viewSchemaID)
	switch recordType {
	case "party":
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN parties p
    ON p.incident_id = r.incident_id
   AND p.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'party'
`, recordID).Scan(&incidentID)
		return incidentID, err
	case "artifact":
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN artifacts a
    ON a.incident_id = r.incident_id
   AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = $2
`, recordID, artifacts.ArtifactTypeForView(viewSchemaID)).Scan(&incidentID)
		return incidentID, err
	case "task_request", "decision":
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type = $2
`, recordID, recordType).Scan(&incidentID)
		return incidentID, err
	default:
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&incidentID)
		return incidentID, err
	}
}

type recordRouteTarget struct {
	IncidentID uuid.UUID
	RecordType string
	Deleted    bool
}

func (s *Store) RecordRouteTarget(ctx context.Context, recordID uuid.UUID) (recordRouteTarget, error) {
	var target recordRouteTarget
	var deletedAt sql.NullTime
	err := s.pool.QueryRow(ctx, `
SELECT incident_id, record_type, deleted_at
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&target.IncidentID, &target.RecordType, &deletedAt)
	if err != nil {
		return recordRouteTarget{}, err
	}
	target.Deleted = deletedAt.Valid
	return target, nil
}

type recordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func recordsInsertParams(incidentID uuid.UUID, recordType string, actorID uuid.UUID, now time.Time) records.InsertParams {
	return records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      recordType,
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	}
}

func loadRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (recordMeta, error) {
	var meta recordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return recordMeta{}, err
	}
	if deletedAt.Valid {
		return recordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func (s *Store) loadGenericRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if !projections.SupportsQuerySurface(viewSchemaID) {
		return nil, fmt.Errorf("workbook mutation surface %q not mapped", viewSchemaID)
	}
	return s.projectionStore.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

func validateCreateRequest(request CreateRequest) error {
	switch request.ViewSchemaID {
	case NotesViewSchemaID:
		if !hasTextValue(request.Values, "note.title") && !hasTextValue(request.Values, "note.body") {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
	case EvidenceViewSchemaID:
		if !hasTextValue(request.Values, "evidence.title") &&
			!hasTextValue(request.Values, "evidence.storage_ref") &&
			!hasTextValue(request.Values, "evidence.collector_party_text") &&
			!hasTextValue(request.Values, "evidence.source_party_text") {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
		if value, ok := request.Values["evidence.lifecycle_state"]; ok && !validEvidenceLifecycleState(derefText(value.Text)) {
			return mutationValidationError("evidence.lifecycle_state", "invalid_value")
		}
	case PartiesViewSchemaID:
		if !hasTextValue(request.Values, "party.display_name") {
			return mutationValidationError("party.display_name", "missing_required_field")
		}
		if !validValueText(request.Values, "party.party_kind", validPartyKind) {
			return mutationValidationError("party.party_kind", "missing_required_field")
		}
	case CommLogViewSchemaID:
		for _, field := range []string{"comm_log.comm_type", "comm_log.audience", "comm_log.channel_or_meeting", "comm_log.summary"} {
			if !hasTextValue(request.Values, field) {
				return mutationValidationError(field, "missing_required_field")
			}
		}
		if !validValueText(request.Values, "comm_log.comm_type", validCommType) {
			return mutationValidationError("comm_log.comm_type", "invalid_value")
		}
	case HandoffViewSchemaID:
		if !hasUUIDValue(request.Values, "handoff.incoming_owner_user_id") {
			return mutationValidationError("handoff.incoming_owner_user_id", "missing_required_field")
		}
		if !hasTextValue(request.Values, "handoff.current_state_summary") {
			return mutationValidationError("handoff.current_state_summary", "missing_required_field")
		}
	case StatusReviewViewSchemaID:
		if !hasTextValue(request.Values, "status_review.current_state_summary") {
			return mutationValidationError("status_review.current_state_summary", "missing_required_field")
		}
	case LessonViewSchemaID:
		if !hasTextValue(request.Values, "lesson.summary") {
			return mutationValidationError("lesson.summary", "missing_required_field")
		}
		if value, ok := request.Values["lesson.closure_state"]; ok && !validClosureState(derefText(value.Text)) {
			return mutationValidationError("lesson.closure_state", "invalid_value")
		}
	case FindingsViewSchemaID:
		if !hasTextValue(request.Values, "finding.statement") {
			return mutationValidationError("finding.statement", "missing_required_field")
		}
		if value, ok := request.Values["finding.kind"]; ok && !validFindingKind(derefText(value.Text)) {
			return mutationValidationError("finding.kind", "invalid_value")
		}
		if value, ok := request.Values["finding.state"]; ok && !validFindingState(derefText(value.Text)) {
			return mutationValidationError("finding.state", "invalid_value")
		}
		if value, ok := request.Values["finding.confidence_score"]; ok && value.Number != nil && !validConfidenceScore(*value.Number) {
			return mutationValidationError("finding.confidence_score", "invalid_value")
		}
	case InvestigativeQueriesViewSchemaID:
		for _, field := range []string{"investigative_query.platform", "investigative_query.purpose", "investigative_query.query_text"} {
			if !hasTextValue(request.Values, field) {
				return mutationValidationError(field, "missing_required_field")
			}
		}
	case ForensicKeywordsViewSchemaID:
		for _, field := range []string{"forensic_keyword.pattern", "forensic_keyword.reason"} {
			if !hasTextValue(request.Values, field) {
				return mutationValidationError(field, "missing_required_field")
			}
		}
		if value, ok := request.Values["forensic_keyword.match_mode"]; ok && !validForensicKeywordMatchMode(derefText(value.Text)) {
			return mutationValidationError("forensic_keyword.match_mode", "invalid_value")
		}
	case TaskRequestsViewSchemaID:
		if !hasTextValue(request.Values, "task.title") {
			return mutationValidationError("task.title", "missing_required_field")
		}
		if !validValueText(request.Values, "task.task_kind", validTaskKind) {
			return mutationValidationError("task.task_kind", "missing_required_field")
		}
		if value, ok := request.Values["task.status"]; ok && !validTaskStatus(derefText(value.Text)) {
			return mutationValidationError("task.status", "invalid_value")
		}
		if value, ok := request.Values["task.priority"]; ok && !validTaskPriority(derefText(value.Text)) {
			return mutationValidationError("task.priority", "invalid_value")
		}
	case DecisionsViewSchemaID:
		if !hasTextValue(request.Values, "decision.summary") {
			return mutationValidationError("decision.summary", "missing_required_field")
		}
		if !validValueText(request.Values, "decision.decision_type", validDecisionType) {
			return mutationValidationError("decision.decision_type", "missing_required_field")
		}
		if !hasTextValue(request.Values, "decision.rationale") {
			return mutationValidationError("decision.rationale", "missing_required_field")
		}
		if value, ok := request.Values["decision.status"]; ok {
			status := derefText(value.Text)
			if !validDecisionStatus(status) {
				return mutationValidationError("decision.status", "invalid_value")
			}
			if status == "superseded" {
				return &LifecycleValidationError{ToStatus: status, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
			}
		}
	default:
		schema, ok := viewschema.Lookup(request.ViewSchemaID)
		if ok && !schema.PermitsZeroFieldCreate && len(request.Values) == 0 && len(request.Collections) == 0 {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
	}
	return nil
}

func (s *Store) validatePatchLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request PatchRequest) error {
	if request.ViewSchemaID != EvidenceViewSchemaID {
		return nil
	}
	changes := make([]evidence.WorkbookLifecyclePatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		var text *string
		if change.Value != nil {
			text = change.Value.Text
		}
		changes = append(changes, evidence.WorkbookLifecyclePatchChange{
			FieldKey: change.FieldKey,
			Text:     text,
		})
	}
	return s.evidenceStore.ValidateWorkbookLifecyclePatchTx(ctx, tx, recordID, changes)
}

func validateCreateReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
		if value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, fieldKey, *value.UUID); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateCollectionPayloadTx(ctx, tx, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validatePatchReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request PatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Value != nil && change.Value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, change.FieldKey, *change.Value.UUID); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateCollectionPayloadTx(ctx, tx, incidentID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	for _, action := range payload.Actions {
		switch {
		case action.LinkedRecordID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.LinkedRecordID, expectedTargetType(fieldKey), fieldKey); err != nil {
				return err
			}
		case action.PartyID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.PartyID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDirectReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, recordID uuid.UUID) error {
	switch fieldKey {
	case "task.requester_party_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "party", fieldKey)
	case "task.decision_record_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "decision", fieldKey)
	case "evidence.collector_party_id", "evidence.source_party_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "party", fieldKey)
	default:
		return nil
	}
}

func validateActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
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
			return fmt.Errorf("validate collection target: %w", err)
		}
		if !exists {
			return mutationValidationError(field, "invalid_value")
		}
		return nil
	}
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func (s *Store) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request PatchRequest, now time.Time) (bool, error) {
	changed := false
	var beforeTask tasksdecisions.TaskLifecycleState
	var beforeDecisionStatus string
	var err error
	if request.ViewSchemaID == TaskRequestsViewSchemaID {
		beforeTask, err = s.taskStore.LoadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
	}
	if request.ViewSchemaID == DecisionsViewSchemaID {
		if err := s.taskStore.ValidateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, adaptOwnerMutationError(err)
		}
		if touchesField(request.Changes, "decision.status") {
			beforeDecisionStatus, err = s.taskStore.LoadDecisionStatusTx(ctx, tx, recordID)
			if err != nil {
				return false, err
			}
		}
	}
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, err := s.applyDirectChangeTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, err := s.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		applied, err := s.taskStore.NormalizeTaskLifecycleTx(ctx, tx, recordID, beforeTask, touchesField(request.Changes, "task.completed_at"), now)
		if err != nil {
			return false, adaptOwnerMutationError(err)
		}
		changed = changed || applied
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		afterDecisionStatus, err := s.taskStore.LoadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
		if err := tasksdecisions.ValidateDecisionStatusTransition(beforeDecisionStatus, afterDecisionStatus); err != nil {
			return false, adaptOwnerMutationError(err)
		}
		if err := s.taskStore.ValidateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, adaptOwnerMutationError(err)
		}
	}
	if request.ViewSchemaID == FindingsViewSchemaID && touchesField(request.Changes, "finding.state") {
		applied, err := s.artifactStore.NormalizeFindingLifecycleTx(ctx, tx, recordID, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func (s *Store) applyDirectChangeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, change PatchChange, now time.Time) (bool, error) {
	if err := validateDirectFieldValue(change); err != nil {
		return false, err
	}
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		changed, err := s.evidenceStore.ApplyWorkbookDirectChangeTx(ctx, tx, recordID, change.FieldKey, evidenceValueFromWorkbook(*change.Value), now)
		return changed, adaptOwnerMutationError(err)
	case PartiesViewSchemaID:
		return s.entityStore.ApplyPartyDirectChangeTx(ctx, tx, recordID, change.FieldKey, partyValueFromWorkbook(*change.Value), now)
	case TaskRequestsViewSchemaID:
		changed, err := s.taskStore.ApplyTaskDirectChangeTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, taskDecisionValueFromWorkbook(*change.Value), now)
		return changed, adaptOwnerMutationError(err)
	case DecisionsViewSchemaID:
		changed, err := s.taskStore.ApplyDecisionDirectChangeTx(ctx, tx, recordID, change.FieldKey, taskDecisionValueFromWorkbook(*change.Value), now)
		return changed, adaptOwnerMutationError(err)
	default:
		changed, err := s.artifactStore.ApplyDirectChangeTx(ctx, tx, recordID, change.FieldKey, artifactValueFromWorkbook(*change.Value), now)
		if err != nil && strings.Contains(err.Error(), "unsupported field key") {
			return false, mutationValidationError(change.FieldKey, "unsupported_field_key")
		}
		return changed, err
	}
}

func validateDirectFieldValue(change PatchChange) error {
	if change.Value == nil {
		return nil
	}
	if change.FieldKey == "finding.confidence_score" && change.Value.Number != nil && !validConfidenceScore(*change.Value.Number) {
		return mutationValidationError(change.FieldKey, "invalid_value")
	}
	if change.Value.Text == nil {
		return nil
	}
	value := *change.Value.Text
	switch change.FieldKey {
	case "task.status":
		if !validTaskStatus(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "task.task_kind":
		if !validTaskKind(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "task.priority":
		if !validTaskPriority(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "decision.status":
		if !validDecisionStatus(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "decision.decision_type":
		if !validDecisionType(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "evidence.lifecycle_state":
		if !validEvidenceLifecycleState(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "comm_log.comm_type":
		if !validCommType(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "lesson.closure_state":
		if !validClosureState(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "finding.kind":
		if !validFindingKind(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "finding.state":
		if !validFindingState(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "forensic_keyword.match_mode":
		if !validForensicKeywordMatchMode(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	}
	return nil
}

func (s *Store) applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) error {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if _, err := s.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, key, collections[key], now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			applied, err := s.linkStore.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_record_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "record_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := s.linkStore.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, fieldKey, expectedTargetType(fieldKey), actorID, now)
			if err != nil {
				if errors.Is(err, links.ErrFieldReferenceNotFound) {
					return false, mutationValidationError(fieldKey, "invalid_value")
				}
				return false, err
			}
			changed = changed || applied
		case "add_tag":
			applied, err := s.linkStore.UpsertTagTx(ctx, tx, incidentID, recordID, action.RawText, action.NormalizedText, actorID, now)
			if err != nil {
				if errors.Is(err, links.ErrInvalidTag) {
					return false, mutationValidationError("note.tags", "invalid_value")
				}
				return false, err
			}
			changed = changed || applied
		case "remove_tag":
			_, tagID, err := links.ParseRecordTagItemRef(action.ItemRef)
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := s.linkStore.TombstoneTagTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
			if err != nil {
				if errors.Is(err, links.ErrTagNotFound) {
					return false, mutationValidationError("note.tags", "invalid_value")
				}
				return false, err
			}
			changed = changed || applied
		case "add_party_ref":
			applied, err := s.linkStore.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.PartyID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_party_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "party_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := s.linkStore.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, fieldKey, "party", actorID, now)
			if err != nil {
				if errors.Is(err, links.ErrFieldReferenceNotFound) {
					return false, mutationValidationError(fieldKey, "invalid_value")
				}
				return false, err
			}
			changed = changed || applied
		case "add_risk_ref":
			applied, err := s.linkStore.UpsertRiskRefTx(ctx, tx, incidentID, recordID, action.RiskRefText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_risk_ref":
			riskRefID, err := uuidFromItemRef(action.ItemRef, "risk_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := s.linkStore.TombstoneRiskRefTx(ctx, tx, incidentID, recordID, riskRefID, actorID, now)
			if err != nil {
				if errors.Is(err, links.ErrRiskRefNotFound) {
					return false, mutationValidationError("handoff.open_risk_refs", "invalid_value")
				}
				return false, err
			}
			changed = changed || applied
		default:
			return false, mutationValidationError(fieldKey, "invalid_value")
		}
	}
	return changed, nil
}

func loadLinkedNoteSourceIncidentTx(ctx context.Context, tx pgx.Tx, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	var incidentID uuid.UUID
	err := tx.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type IN ('timeline_event', 'host', 'identity', 'evidence')
   AND deleted_at IS NULL
`, sourceRecordID).Scan(&incidentID)
	return incidentID, err
}

func (s *Store) touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		return s.evidenceStore.TouchWorkbookRowTx(ctx, tx, recordID, now)
	case PartiesViewSchemaID:
		return s.entityStore.TouchPartyTx(ctx, tx, recordID, now)
	case TaskRequestsViewSchemaID:
		return s.taskStore.TouchTaskRequestTx(ctx, tx, recordID, now)
	case DecisionsViewSchemaID:
		return s.taskStore.TouchDecisionTx(ctx, tx, recordID, now)
	default:
		if artifacts.IsArtifactBackedView(viewSchemaID) {
			return s.artifactStore.TouchRowTx(ctx, tx, recordID, now)
		}
		return mutationValidationError("view_schema_id", "unknown_view_schema")
	}
}

func (s *Store) refreshWorkbookProjectionTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		return s.projectionStore.RefreshEvidenceTx(ctx, tx, recordID)
	case PartiesViewSchemaID:
		return s.projectionStore.RefreshPartyTx(ctx, tx, recordID)
	case TaskRequestsViewSchemaID:
		return s.projectionStore.RefreshTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return s.projectionStore.RefreshDecisionTx(ctx, tx, recordID)
	default:
		if artifacts.IsArtifactBackedView(viewSchemaID) {
			return s.projectionStore.RefreshArtifactTx(ctx, tx, recordID)
		}
		return nil
	}
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

type workbookConflictTokenClaims = revisions.ConflictTokenClaims

var defaultWorkbookConflictTokenCodec = revisions.NewConflictTokenCodecForTesting("workbook")

func parseWorkbookConflictToken(token string) (workbookConflictTokenClaims, bool) {
	return parseWorkbookConflictTokenWithCodec(defaultWorkbookConflictTokenCodec, token)
}

func (s *Store) parseWorkbookConflictToken(token string) (workbookConflictTokenClaims, bool) {
	return parseWorkbookConflictTokenWithCodec(s.conflictTokens, token)
}

func parseWorkbookConflictTokenWithCodec(codec revisions.ConflictTokenCodec, token string) (workbookConflictTokenClaims, bool) {
	claims, ok := codec.Parse(token)
	if !ok {
		return workbookConflictTokenClaims{}, false
	}
	if claims.RouteKey != workbookConflictResolveRouteKey || claims.ViewSchemaID == timeline.TimelineViewSchemaID {
		return workbookConflictTokenClaims{}, false
	}
	return claims, true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractPayloadUUID(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	return uuid.Parse(text)
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func recordTypeMatchesView(recordType string, viewSchemaID string) bool {
	return recordType == recordTypeForView(viewSchemaID)
}

func recordTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		return "evidence"
	case PartiesViewSchemaID:
		return "party"
	case TaskRequestsViewSchemaID:
		return "task_request"
	case DecisionsViewSchemaID:
		return "decision"
	case NotesViewSchemaID, CommLogViewSchemaID, HandoffViewSchemaID, StatusReviewViewSchemaID, LessonViewSchemaID,
		FindingsViewSchemaID, InvestigativeQueriesViewSchemaID, ForensicKeywordsViewSchemaID:
		return "artifact"
	default:
		return ""
	}
}

func uuidFromItemRef(itemRef string, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	value := strings.TrimPrefix(itemRef, prefix)
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return parsed, nil
}

func hasTextValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && *value.Text != ""
}

func hasUUIDValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil
}

func validValueText(values map[string]ValueChange, field string, valid func(string) bool) bool {
	value, ok := values[field]
	return ok && value.Text != nil && valid(*value.Text)
}

func nullableUUIDValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.UUID == nil {
		return nil
	}
	return *value.UUID
}

func derefText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func expectedTargetType(fieldKey string) string {
	switch fieldKey {
	case "comm_log.decision_ids", "handoff.open_decision_ids", "status_review.open_decision_ids":
		return "decision"
	case "comm_log.action_task_ids", "handoff.open_task_ids", "status_review.blocked_task_ids", "lesson.follow_up_task_ids":
		return "task_request"
	case "status_review.pending_evidence_ids", "lesson.evidence_refs":
		return "evidence"
	case "task.linked_record_ids", "decision.support_refs", "decision.affected_record_ids",
		"finding.supporting_refs", "finding.contradictory_refs":
		return ""
	default:
		return ""
	}
}

func validPartyKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}

func validEvidenceLifecycleState(value string) bool {
	return evidence.ValidLifecycleState(value)
}

func validCommType(value string) bool {
	switch value {
	case "meeting", "notification", "approval", "briefing", "handoff":
		return true
	default:
		return false
	}
}

func validClosureState(value string) bool {
	return value == "open" || value == "closed"
}

func validFindingKind(value string) bool {
	return value == "finding" || value == "hypothesis"
}

func validFindingState(value string) bool {
	return value == "open" || value == "closed"
}

func validForensicKeywordMatchMode(value string) bool {
	return value == "literal" || value == "regex"
}

func validConfidenceScore(value int64) bool {
	return value >= 0 && value <= 100
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

func decisionRowStatus(row map[string]any) string {
	cells, _ := row["cells"].(map[string]any)
	cell, _ := cells["decision.status"].(map[string]any)
	status, _ := cell["value"].(string)
	return status
}

func nullableUUIDPointer(value any) *uuid.UUID {
	switch typed := value.(type) {
	case uuid.UUID:
		return &typed
	default:
		return nil
	}
}

func touchesField(changes []PatchChange, fieldKey string) bool {
	for _, change := range changes {
		if change.FieldKey == fieldKey {
			return true
		}
	}
	return false
}

func touchesAnyField(changes []PatchChange, fieldKeys ...string) bool {
	for _, fieldKey := range fieldKeys {
		if touchesField(changes, fieldKey) {
			return true
		}
	}
	return false
}
