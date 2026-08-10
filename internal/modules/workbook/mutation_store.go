package workbook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func (s *Store) CreateWorkbookRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if s == nil || s.contributions == nil {
		return MutationResult{}, fmt.Errorf("workbook contribution catalog is required")
	}
	provider, ok := s.contributions.CreateFor(request.ViewSchemaID)
	if !ok {
		return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
	}
	return provider.Create(ctx, CreateCommand{
		Actor:        actor,
		IncidentID:   incidentID,
		ViewSchemaID: request.ViewSchemaID,
		Admission: CreateAdmission{
			ClientTxnID: request.ClientTxnID,
			normalized:  request,
		},
		RequestHash: requestHash,
		RequestID:   requestID,
		Now:         now,
	})
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
	var partyValidation *parties.ValidationError
	if errors.As(err, &partyValidation) {
		return mutationValidationError(partyValidation.Field, partyValidation.ReasonCode)
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

func artifactCreateRequestFromWorkbook(request CreateRequest) artifacts.CreateRequest {
	return artifacts.CreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       artifactValuesFromWorkbook(request.Values),
		Collections:  artifactWorkbookCollectionsFromWorkbook(request.Collections),
	}
}

func artifactPatchRequestFromWorkbook(request PatchRequest) artifacts.PatchRequest {
	changes := make([]artifacts.PatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := artifacts.PatchChange{
			FieldKey:       change.FieldKey,
			CanonicalValue: change.CanonicalAny,
		}
		if change.Value != nil {
			value := artifactValueFromWorkbook(*change.Value)
			converted.Value = &value
		}
		if change.Collection != nil {
			payload := artifactWorkbookCollectionPayloadFromWorkbook(*change.Collection)
			converted.Collection = &payload
		}
		changes = append(changes, converted)
	}
	return artifacts.PatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}
}

func artifactWorkbookCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]artifacts.CollectionActionPayload {
	result := make(map[string]artifacts.CollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = artifactWorkbookCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func artifactWorkbookCollectionPayloadFromWorkbook(payload CollectionActionPayload) artifacts.CollectionActionPayload {
	actions := make([]artifacts.CollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, artifacts.CollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return artifacts.CollectionActionPayload{Actions: actions}
}

func mutationResultFromArtifactWorkbook(result artifacts.MutationResult) MutationResult {
	payload := map[string]any{
		"view_schema_id": result.ViewSchemaID,
		"row":            result.Row,
	}
	if result.ChangeSetID != uuid.Nil {
		payload["change_set_id"] = result.ChangeSetID.String()
	}
	if result.ContextualLink != nil {
		payload["source_record_id"] = result.ContextualLink.SourceRecordID.String()
		payload["link_type"] = result.ContextualLink.LinkType
	}
	statusCode := http.StatusOK
	if result.Created {
		statusCode = http.StatusCreated
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       statusCode,
		Replayed:         result.Replayed,
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		RowVersion:       result.RowVersion,
		ViewSchemaID:     result.ViewSchemaID,
		ChangedFieldKeys: result.ChangedFieldKeys,
	}
}

func adaptArtifactWorkbookOwnerError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, artifacts.ErrClientTxnConflict) {
		return authn.ErrClientTxnConflict
	}
	var validation *artifacts.ValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	var collectionValidation *links.CollectionValidationError
	if errors.As(err, &collectionValidation) {
		return mutationValidationError(collectionValidation.Field, collectionValidation.ReasonCode)
	}
	var rowConflict *artifacts.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	var sameConflict *artifacts.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		return &SameFieldConflictError{Conflict: sameConflict.Conflict}
	}
	return err
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

func evidenceCreateRequestFromWorkbook(request CreateRequest) evidence.WorkbookCreateRequest {
	converted := evidence.WorkbookCreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       evidenceValuesFromWorkbook(request.Values),
	}
	if value, ok := request.Inputs["evidence.initial_object_blob_id"]; ok {
		objectBlobID := value.UUID
		converted.InitialObjectBlobID = &objectBlobID
	}
	return converted
}

func evidencePatchRequestFromWorkbook(request PatchRequest) evidence.WorkbookPatchRequest {
	changes := make([]evidence.WorkbookPatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := evidence.WorkbookPatchChange{
			FieldKey:       change.FieldKey,
			CanonicalValue: change.CanonicalAny,
		}
		if change.Value != nil {
			value := evidenceValueFromWorkbook(*change.Value)
			converted.Value = &value
		}
		changes = append(changes, converted)
	}
	return evidence.WorkbookPatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}
}

func mutationResultFromEvidenceWorkbook(result evidence.WorkbookMutationResult) MutationResult {
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
	}
}

func adaptEvidenceWorkbookOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var validation *evidence.ValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	var attachRejected evidence.AttachRejectedError
	if errors.As(err, &attachRejected) {
		return &publicMutationError{apiError: &httpapi.APIError{
			Status:  http.StatusConflict,
			Code:    "evidence_attach_rejected",
			Message: "evidence attach rejected",
			Details: map[string]any{"reason_code": attachRejected.ReasonCode},
		}}
	}
	if errors.Is(err, evidence.ErrBlobNotFound) || errors.Is(err, evidence.ErrIncidentMismatch) {
		return &publicMutationError{apiError: &httpapi.APIError{
			Status:  http.StatusConflict,
			Code:    "evidence_attach_rejected",
			Message: "evidence attach rejected",
			Details: map[string]any{"reason_code": evidence.AttachReasonBlobNotVisible},
		}}
	}
	if errors.Is(err, evidence.ErrObjectStoreUnavailable) {
		return workbookObjectStoreUnavailable("dependency_unavailable")
	}
	if reasonCode, ok := evidence.PersistedObjectBlobStorageKeyErrorReason(err); ok {
		return &publicMutationError{apiError: &httpapi.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "object_store_invalid_request",
			Details: map[string]any{"reason_code": reasonCode},
		}}
	}
	if adapterError, ok := objectstore.AsAdapterError(err); ok {
		switch adapterError.Code {
		case objectstore.ErrorCodeAccessRejected:
			return &publicMutationError{apiError: &httpapi.APIError{
				Status:    http.StatusServiceUnavailable,
				Code:      "object_store_access_rejected",
				Retryable: false,
				Details:   map[string]any{"reason_code": "credential_denied"},
			}}
		case objectstore.ErrorCodeUnavailable:
			return workbookObjectStoreUnavailable("endpoint_unreachable")
		case objectstore.ErrorCodeDeadlineExceeded, objectstore.ErrorCodeRetryExhausted:
			return workbookObjectStoreUnavailable("retry_exhausted")
		default:
			return &publicMutationError{apiError: &httpapi.APIError{
				Status:  http.StatusInternalServerError,
				Code:    "object_store_invalid_request",
				Details: map[string]any{"reason_code": "invalid_request"},
			}}
		}
	}
	var lifecycleValidation *evidence.LifecycleValidationError
	if errors.As(err, &lifecycleValidation) {
		return &LifecycleValidationError{
			FromStatus:     lifecycleValidation.FromStatus,
			ToStatus:       lifecycleValidation.ToStatus,
			ViolatedGuards: lifecycleValidation.ViolatedGuards,
			ReasonCode:     lifecycleValidation.ReasonCode,
		}
	}
	var rowConflict *evidence.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	var sameConflict *evidence.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		return &SameFieldConflictError{Conflict: sameConflict.Conflict}
	}
	return err
}

func workbookObjectStoreUnavailable(reasonCode string) error {
	return &publicMutationError{apiError: &httpapi.APIError{
		Status:    http.StatusServiceUnavailable,
		Code:      "object_store_unavailable",
		Retryable: true,
		Details:   map[string]any{"reason_code": reasonCode},
	}}
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

func partyCreateRequestFromWorkbook(request CreateRequest) parties.WorkbookCreateRequest {
	return parties.WorkbookCreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       partyValuesFromWorkbook(request.Values),
	}
}

func partyPatchRequestFromWorkbook(request PatchRequest) parties.WorkbookPatchRequest {
	changes := make([]parties.WorkbookPatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := parties.WorkbookPatchChange{
			FieldKey:       change.FieldKey,
			CanonicalValue: change.CanonicalAny,
		}
		if change.Value != nil {
			value := partyValueFromWorkbook(*change.Value)
			converted.Value = &value
		}
		changes = append(changes, converted)
	}
	return parties.WorkbookPatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}
}

func mutationResultFromPartyWorkbook(result parties.WorkbookMutationResult) MutationResult {
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
	}
}

func adaptPartyWorkbookOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var validation *parties.ValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	var rowConflict *parties.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	var sameConflict *parties.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		return &SameFieldConflictError{Conflict: sameConflict.Conflict}
	}
	return err
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

func taskDecisionCreateRequestFromWorkbook(request CreateRequest) tasksdecisions.WorkbookCreateRequest {
	return tasksdecisions.WorkbookCreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       taskDecisionValuesFromWorkbook(request.Values),
		Collections:  taskDecisionCollectionsFromWorkbook(request.Collections),
	}
}

func taskDecisionPatchRequestFromWorkbook(request PatchRequest) tasksdecisions.WorkbookPatchRequest {
	changes := make([]tasksdecisions.WorkbookPatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := tasksdecisions.WorkbookPatchChange{
			FieldKey:       change.FieldKey,
			CanonicalValue: change.CanonicalAny,
		}
		if change.Value != nil {
			value := taskDecisionValueFromWorkbook(*change.Value)
			converted.Value = &value
		}
		if change.Collection != nil {
			payload := taskDecisionCollectionPayloadFromWorkbook(*change.Collection)
			converted.Collection = &payload
		}
		changes = append(changes, converted)
	}
	return tasksdecisions.WorkbookPatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}
}

func taskDecisionCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]tasksdecisions.WorkbookCollectionActionPayload {
	result := make(map[string]tasksdecisions.WorkbookCollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = taskDecisionCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func taskDecisionCollectionPayloadFromWorkbook(payload CollectionActionPayload) tasksdecisions.WorkbookCollectionActionPayload {
	actions := make([]tasksdecisions.WorkbookCollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, tasksdecisions.WorkbookCollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return tasksdecisions.WorkbookCollectionActionPayload{Actions: actions}
}

func mutationResultFromTaskDecisionWorkbook(result tasksdecisions.WorkbookMutationResult) MutationResult {
	payload := map[string]any{
		"view_schema_id": result.ViewSchemaID,
		"row":            result.Row,
	}
	if result.ChangeSetID != uuid.Nil {
		payload["change_set_id"] = result.ChangeSetID.String()
	}
	statusCode := http.StatusOK
	if result.Created && !result.Replayed {
		statusCode = http.StatusCreated
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       statusCode,
		Replayed:         result.Replayed,
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		RowVersion:       result.RowVersion,
		ViewSchemaID:     result.ViewSchemaID,
		ChangedFieldKeys: result.ChangedFieldKeys,
	}
}

func adaptTaskDecisionWorkbookOwnerError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, tasksdecisions.ErrClientTxnConflict) {
		return authn.ErrClientTxnConflict
	}
	var validation *tasksdecisions.ValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	var collectionValidation *links.CollectionValidationError
	if errors.As(err, &collectionValidation) {
		return mutationValidationError(collectionValidation.Field, collectionValidation.ReasonCode)
	}
	var lifecycle *tasksdecisions.LifecycleValidationError
	if errors.As(err, &lifecycle) {
		return &LifecycleValidationError{
			FromStatus:     lifecycle.FromStatus,
			ToStatus:       lifecycle.ToStatus,
			ReasonCode:     lifecycle.ReasonCode,
			ViolatedGuards: append([]string(nil), lifecycle.ViolatedGuards...),
		}
	}
	var rowConflict *tasksdecisions.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return &RowVersionConflictError{
			RecordID:          rowConflict.RecordID,
			BaseRowVersion:    rowConflict.BaseRowVersion,
			CurrentRowVersion: rowConflict.CurrentRowVersion,
		}
	}
	var sameConflict *tasksdecisions.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		return &SameFieldConflictError{Conflict: sameConflict.Conflict}
	}
	return err
}

func (s *Store) CreateLinkedNote(ctx context.Context, actor authn.UserRecord, sourceRecordID uuid.UUID, request LinkedNoteCreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	result, err := s.contextualNoteOwner.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{
		ActorUserID:    actor.ID,
		SourceRecordID: sourceRecordID,
		Request:        linkedNoteCreateRequestToArtifacts(request),
		RequestHash:    requestHash,
		RequestID:      requestID,
		OperationID:    artifacts.OperationLinkedNoteCreate,
		Now:            now,
	})
	if err != nil {
		return MutationResult{}, adaptLinkedNoteOwnerError(err)
	}
	return mutationResultFromArtifactWorkbook(result), nil
}

func (s *Store) LinkedNoteSourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	return s.contextualNoteOwner.SourceIncident(ctx, sourceRecordID)
}

func linkedNoteCreateRequestToArtifacts(request LinkedNoteCreateRequest) artifacts.ContextualNoteCreateRequest {
	return artifacts.ContextualNoteCreateRequest{
		ClientTxnID: request.ClientTxnID,
		Values:      artifactValuesFromWorkbook(request.Values),
		Collections: artifactCollectionsFromWorkbook(request.Collections),
	}
}

func artifactCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]artifacts.CollectionActionPayload {
	result := make(map[string]artifacts.CollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = linkedNoteCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func linkedNoteCollectionPayloadFromWorkbook(payload CollectionActionPayload) artifacts.CollectionActionPayload {
	actions := make([]artifacts.CollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, artifacts.CollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			NormalizedText: action.NormalizedText,
		})
	}
	return artifacts.CollectionActionPayload{Actions: actions}
}

func adaptLinkedNoteOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var validation *artifacts.ValidationError
	if errors.As(err, &validation) {
		return mutationValidationError(validation.Field, validation.ReasonCode)
	}
	var collectionValidation *links.CollectionValidationError
	if errors.As(err, &collectionValidation) {
		return mutationValidationError(collectionValidation.Field, collectionValidation.ReasonCode)
	}
	return err
}

func (s *Store) PatchWorkbookRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if s == nil || s.contributions == nil {
		return MutationResult{}, fmt.Errorf("workbook contribution catalog is required")
	}
	target, err := s.recordTargets.Resolve(ctx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	provider, ok := s.contributions.PatchFor(target.RecordType)
	if !ok {
		return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
	}
	return provider.Patch(ctx, PatchCommand{
		Actor:                   actor,
		RecordID:                recordID,
		AuthoritativeRecordType: target.RecordType,
		Admission: PatchAdmission{
			ViewSchemaID:   request.ViewSchemaID,
			BaseRowVersion: request.BaseRowVersion,
			ClientTxnID:    request.ClientTxnID,
			normalized:     request,
		},
		RequestHash: requestHash,
		RequestID:   requestID,
		Now:         now,
	})
}

func (s *Store) SupersedeDecision(ctx context.Context, actor authn.UserRecord, targetRecordID uuid.UUID, request timeline.SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	result, err := s.supersedeStore.SupersedeDecision(ctx, tasksdecisions.SupersedeCommand{
		ActorUserID:    actor.ID,
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
	payload := map[string]any{"row": result.Row}
	if result.Row == nil {
		payload = map[string]any{
			"view_schema_id":          tasksdecisions.DecisionsViewSchemaID,
			"change_set_id":           result.ChangeSetID.String(),
			"target_record_id":        result.Facts.TargetRecordID.String(),
			"superseding_record_id":   result.Facts.SupersedingRecordID.String(),
			"target_row_version":      result.Facts.TargetRowVersion,
			"superseding_row_version": result.Facts.SupersedingRowVersion,
			"target_status":           result.Facts.TargetStatus,
			"reason":                  result.Facts.Reason,
		}
	}
	return MutationResult{
		Payload:                 payload,
		StatusCode:              http.StatusOK,
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
	if errors.Is(err, tasksdecisions.ErrClientTxnConflict) {
		return authn.ErrClientTxnConflict
	}
	return adaptOwnerMutationError(err)
}
