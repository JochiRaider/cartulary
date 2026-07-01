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
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
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
		if err := s.partyStore.InsertPartyTx(ctx, tx, recordID, incidentID, parties.CreateParams{Values: partyValuesFromWorkbook(request.Values)}, now.UTC()); err != nil {
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
	recordID, found, err := s.partyStore.FindReusablePartyTx(ctx, tx, incidentID, parties.CreateParams{Values: partyValuesFromWorkbook(request.Values)})
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
	var collectionValidation *links.CollectionValidationError
	if errors.As(err, &collectionValidation) {
		return mutationValidationError(collectionValidation.Field, collectionValidation.ReasonCode)
	}
	var evidenceValidation *evidence.ValidationError
	if errors.As(err, &evidenceValidation) {
		return mutationValidationError(evidenceValidation.Field, evidenceValidation.ReasonCode)
	}
	var artifactValidation *artifacts.ValidationError
	if errors.As(err, &artifactValidation) {
		return mutationValidationError(artifactValidation.Field, artifactValidation.ReasonCode)
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

func partyValuesFromWorkbook(values map[string]ValueChange) map[string]parties.FieldValue {
	result := make(map[string]parties.FieldValue, len(values))
	for field, value := range values {
		result[field] = partyValueFromWorkbook(value)
	}
	return result
}

func partyValueFromWorkbook(value ValueChange) parties.FieldValue {
	return parties.FieldValue{
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
	result, err := s.linkedNoteStore.Create(ctx, linkednotes.CreateCommand{
		Actor:          actor,
		SourceRecordID: sourceRecordID,
		Request:        linkedNoteCreateRequestToArtifacts(request),
		RequestHash:    requestHash,
		RequestID:      requestID,
		RouteKey:       workbookLinkedNoteRouteKey,
		Now:            now,
	})
	if err != nil {
		return MutationResult{}, adaptLinkedNoteOwnerError(err)
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
		ViewSchemaID:     result.ViewSchemaID,
		ChangedFieldKeys: result.ChangedFieldKeys,
	}, nil
}

func (s *Store) LinkedNoteSourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	return s.linkedNoteStore.SourceIncident(ctx, sourceRecordID)
}

func linkedNoteCreateRequestToArtifacts(request LinkedNoteCreateRequest) linkednotes.CreateRequest {
	return linkednotes.CreateRequest{
		ClientTxnID: request.ClientTxnID,
		Values:      artifactValuesFromWorkbook(request.Values),
		Collections: artifactCollectionsFromWorkbook(request.Collections),
	}
}

func artifactCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]links.CollectionActionPayload {
	return linkCollectionsFromWorkbook(collections)
}

func linkCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]links.CollectionActionPayload {
	result := make(map[string]links.CollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = linkCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func linkCollectionPayloadFromWorkbook(payload CollectionActionPayload) links.CollectionActionPayload {
	actions := make([]links.CollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, links.CollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return links.CollectionActionPayload{Actions: actions}
}

func adaptLinkedNoteOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var validation *linkednotes.MutationValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	return err
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
	result, err := s.supersedeStore.SupersedeDecision(ctx, tasksdecisions.SupersedeCommand{
		Actor:          actor,
		TargetRecordID: targetRecordID,
		Request: tasksdecisions.SupersedeRequest{
			BaseRowVersion:      request.BaseRowVersion,
			ClientTxnID:         request.ClientTxnID,
			Reason:              request.Reason,
			ReplacementRecordID: request.ReplacementRecordID,
		},
		RequestHash: requestHash,
		RequestID:   requestID,
		RouteKey:    workbookSupersedeRouteKey,
		Now:         now,
	})
	if err != nil {
		return MutationResult{}, adaptDecisionSupersedeOwnerError(err)
	}
	return mutationResultFromDecisionSupersede(result), nil
}

func mutationResultFromDecisionSupersede(result tasksdecisions.SupersedeMutationResult) MutationResult {
	additional := make([]MutationResult, 0, len(result.AdditionalRecordChanges))
	for _, change := range result.AdditionalRecordChanges {
		additional = append(additional, mutationResultFromDecisionSupersede(change))
	}
	return MutationResult{
		Payload:                 result.Payload,
		StatusCode:              result.StatusCode,
		Replayed:                result.Replayed,
		IncidentID:              result.IncidentID,
		RecordID:                result.RecordID,
		ChangeSetID:             result.ChangeSetID,
		ClientTxnID:             result.ClientTxnID,
		RowVersion:              result.RowVersion,
		ViewSchemaID:            result.ViewSchemaID,
		ChangedFieldKeys:        result.ChangedFieldKeys,
		AdditionalRecordChanges: additional,
	}
}

func adaptDecisionSupersedeOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var rowConflict *tasksdecisions.SupersedeRowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	return adaptOwnerMutationError(err)
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
	if !s.projectionRows.Supports(viewSchemaID) {
		return nil, fmt.Errorf("workbook mutation surface %q not mapped", viewSchemaID)
	}
	return s.projectionRows.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

func validateCreateRequest(request CreateRequest) error {
	switch request.ViewSchemaID {
	case EvidenceViewSchemaID:
		return adaptOwnerCreateValidationError(evidence.ValidateWorkbookCreateParams(evidence.WorkbookCreateParams{Values: evidenceValuesFromWorkbook(request.Values)}))
	case PartiesViewSchemaID:
		return adaptOwnerCreateValidationError(parties.ValidateCreateParams(parties.CreateParams{Values: partyValuesFromWorkbook(request.Values)}))
	case TaskRequestsViewSchemaID:
		return adaptOwnerCreateValidationError(tasksdecisions.ValidateTaskCreateParams(tasksdecisions.TaskCreateParams{Values: taskDecisionValuesFromWorkbook(request.Values)}))
	case DecisionsViewSchemaID:
		return adaptOwnerCreateValidationError(tasksdecisions.ValidateDecisionCreateParams(tasksdecisions.DecisionCreateParams{Values: taskDecisionValuesFromWorkbook(request.Values)}))
	default:
		if artifacts.IsArtifactBackedView(request.ViewSchemaID) {
			return adaptOwnerCreateValidationError(artifacts.ValidateCreateParams(artifacts.CreateParams{ViewSchemaID: request.ViewSchemaID, Values: artifactValuesFromWorkbook(request.Values)}))
		}
		schema, ok := viewschema.Lookup(request.ViewSchemaID)
		if ok && !schema.PermitsZeroFieldCreate && len(request.Values) == 0 && len(request.Collections) == 0 {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
	}
	return nil
}

func adaptOwnerCreateValidationError(err error) error {
	if err == nil {
		return nil
	}
	var evidenceValidation *evidence.ValidationError
	if errors.As(err, &evidenceValidation) {
		return mutationValidationError(evidenceValidation.Field, evidenceValidation.ReasonCode)
	}
	var partyValidation *parties.ValidationError
	if errors.As(err, &partyValidation) {
		return mutationValidationError(partyValidation.Field, partyValidation.ReasonCode)
	}
	var artifactValidation *artifacts.ValidationError
	if errors.As(err, &artifactValidation) {
		return mutationValidationError(artifactValidation.Field, artifactValidation.ReasonCode)
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
	return adaptOwnerMutationError(links.NewStore().ValidateCollectionPayloadTx(ctx, tx, incidentID, fieldKey, linkCollectionPayloadFromWorkbook(payload)))
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
		return s.partyStore.ApplyDirectChangeTx(ctx, tx, recordID, change.FieldKey, partyValueFromWorkbook(*change.Value), now)
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
	switch {
	case strings.HasPrefix(change.FieldKey, "evidence."):
		return adaptOwnerMutationError(evidence.ValidateWorkbookDirectPatchChange(change.FieldKey, evidenceValueFromWorkbook(*change.Value)))
	case strings.HasPrefix(change.FieldKey, "task."):
		return adaptOwnerMutationError(tasksdecisions.ValidateTaskDirectPatchChange(change.FieldKey, taskDecisionValueFromWorkbook(*change.Value)))
	case strings.HasPrefix(change.FieldKey, "decision."):
		return adaptOwnerMutationError(tasksdecisions.ValidateDecisionDirectPatchChange(change.FieldKey, taskDecisionValueFromWorkbook(*change.Value)))
	case artifacts.IsArtifactBackedField(change.FieldKey):
		return adaptOwnerMutationError(artifacts.ValidateDirectPatchChange(change.FieldKey, artifactValueFromWorkbook(*change.Value)))
	default:
		return nil
	}
}

func (s *Store) applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) error {
	_, err := s.linkStore.ApplyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actorID, linkCollectionsFromWorkbook(collections), now)
	return adaptOwnerMutationError(err)
}

func (s *Store) applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, error) {
	changed, err := s.linkStore.ApplyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, fieldKey, linkCollectionPayloadFromWorkbook(payload), now)
	return changed, adaptOwnerMutationError(err)
}

func (s *Store) touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		return s.evidenceStore.TouchWorkbookRowTx(ctx, tx, recordID, now)
	case PartiesViewSchemaID:
		return s.partyStore.TouchPartyTx(ctx, tx, recordID, now)
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
		return s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.EvidenceViewSchemaID, recordID)
	case PartiesViewSchemaID:
		return s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.PartiesViewSchemaID, recordID)
	case TaskRequestsViewSchemaID:
		return s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.TaskRequestsViewSchemaID, recordID)
	case DecisionsViewSchemaID:
		return s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, recordID)
	default:
		if artifacts.IsArtifactBackedView(viewSchemaID) {
			return s.rowProjector.RefreshRowTx(ctx, tx, viewSchemaID, recordID)
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

func nullableUUIDValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.UUID == nil {
		return nil
	}
	return *value.UUID
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
