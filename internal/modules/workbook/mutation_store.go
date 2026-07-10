package workbook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *Store) CreateWorkbookRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if artifacts.IsArtifactBackedView(request.ViewSchemaID) {
		result, err := s.artifactMutations.Create(ctx, artifacts.WorkbookCreateCommand{
			Actor:       actor,
			IncidentID:  incidentID,
			Request:     artifactCreateRequestFromWorkbook(request),
			RequestHash: requestHash,
			RequestID:   requestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         now,
		})
		if err != nil {
			return MutationResult{}, adaptArtifactWorkbookOwnerError(err)
		}
		return mutationResultFromArtifactWorkbook(result), nil
	}
	if request.ViewSchemaID == EvidenceViewSchemaID {
		result, err := s.evidenceMutations.Create(ctx, evidence.WorkbookCreateCommand{
			Actor:       actor,
			IncidentID:  incidentID,
			Request:     evidenceCreateRequestFromWorkbook(request),
			RequestHash: requestHash,
			RequestID:   requestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         now,
		})
		if err != nil {
			return MutationResult{}, adaptEvidenceWorkbookOwnerError(err)
		}
		return mutationResultFromEvidenceWorkbook(result), nil
	}
	if request.ViewSchemaID == PartiesViewSchemaID {
		result, err := s.partyMutations.Create(ctx, parties.WorkbookCreateCommand{
			Actor:       actor,
			IncidentID:  incidentID,
			Request:     partyCreateRequestFromWorkbook(request),
			RequestHash: requestHash,
			RequestID:   requestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         now,
		})
		if err != nil {
			return MutationResult{}, adaptPartyWorkbookOwnerError(err)
		}
		return mutationResultFromPartyWorkbook(result), nil
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID || request.ViewSchemaID == DecisionsViewSchemaID {
		result, err := s.taskMutations.Create(ctx, tasksdecisions.WorkbookCreateCommand{
			Actor:       actor,
			IncidentID:  incidentID,
			Request:     taskDecisionCreateRequestFromWorkbook(request),
			RequestHash: requestHash,
			RequestID:   requestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         now,
		})
		if err != nil {
			return MutationResult{}, adaptTaskDecisionWorkbookOwnerError(err)
		}
		return mutationResultFromTaskDecisionWorkbook(result), nil
	}
	return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
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

func artifactCreateRequestFromWorkbook(request CreateRequest) artifacts.WorkbookCreateRequest {
	return artifacts.WorkbookCreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       artifactValuesFromWorkbook(request.Values),
		Collections:  artifactWorkbookCollectionsFromWorkbook(request.Collections),
	}
}

func artifactPatchRequestFromWorkbook(request PatchRequest) artifacts.WorkbookPatchRequest {
	changes := make([]artifacts.WorkbookPatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := artifacts.WorkbookPatchChange{
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
	return artifacts.WorkbookPatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}
}

func artifactWorkbookCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]artifacts.WorkbookCollectionActionPayload {
	result := make(map[string]artifacts.WorkbookCollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = artifactWorkbookCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func artifactWorkbookCollectionPayloadFromWorkbook(payload CollectionActionPayload) artifacts.WorkbookCollectionActionPayload {
	actions := make([]artifacts.WorkbookCollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, artifacts.WorkbookCollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return artifacts.WorkbookCollectionActionPayload{Actions: actions}
}

func mutationResultFromArtifactWorkbook(result artifacts.WorkbookMutationResult) MutationResult {
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

func adaptArtifactWorkbookOwnerError(err error) error {
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
	return evidence.WorkbookCreateRequest{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       evidenceValuesFromWorkbook(request.Values),
	}
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

func adaptTaskDecisionWorkbookOwnerError(err error) error {
	if err == nil {
		return nil
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

func artifactCollectionsFromWorkbook(collections map[string]CollectionActionPayload) map[string]linkednotes.CollectionActionPayload {
	result := make(map[string]linkednotes.CollectionActionPayload, len(collections))
	for fieldKey, payload := range collections {
		result[fieldKey] = linkedNoteCollectionPayloadFromWorkbook(payload)
	}
	return result
}

func linkedNoteCollectionPayloadFromWorkbook(payload CollectionActionPayload) linkednotes.CollectionActionPayload {
	actions := make([]linkednotes.CollectionAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, linkednotes.CollectionAction{
			Op:             action.Op,
			RawText:        action.RawText,
			LinkedRecordID: action.LinkedRecordID,
			PartyID:        action.PartyID,
			ItemRef:        action.ItemRef,
			NormalizedText: action.NormalizedText,
		})
	}
	return linkednotes.CollectionActionPayload{Actions: actions}
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
	if artifacts.IsArtifactBackedView(request.ViewSchemaID) {
		result, err := s.artifactMutations.Patch(ctx, artifacts.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          artifactPatchRequestFromWorkbook(request),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptArtifactWorkbookOwnerError(err)
		}
		return mutationResultFromArtifactWorkbook(result), nil
	}
	if request.ViewSchemaID == EvidenceViewSchemaID {
		result, err := s.evidenceMutations.Patch(ctx, evidence.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          evidencePatchRequestFromWorkbook(request),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptEvidenceWorkbookOwnerError(err)
		}
		return mutationResultFromEvidenceWorkbook(result), nil
	}
	if request.ViewSchemaID == PartiesViewSchemaID {
		result, err := s.partyMutations.Patch(ctx, parties.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          partyPatchRequestFromWorkbook(request),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptPartyWorkbookOwnerError(err)
		}
		return mutationResultFromPartyWorkbook(result), nil
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID || request.ViewSchemaID == DecisionsViewSchemaID {
		result, err := s.taskMutations.Patch(ctx, tasksdecisions.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          taskDecisionPatchRequestFromWorkbook(request),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptTaskDecisionWorkbookOwnerError(err)
		}
		return mutationResultFromTaskDecisionWorkbook(result), nil
	}
	if request.ViewSchemaID == hostidentity.HostsViewSchemaID || request.ViewSchemaID == hostidentity.IdentitiesViewSchemaID {
		entityRequest, err := entityPatchRequestFromWorkbook(request)
		if err != nil {
			return MutationResult{}, err
		}
		result, err := s.entityStore.PatchEntityRow(ctx, actor, recordID, entityRequest, requestHash, requestID, now, workbookPatchRouteKey)
		if err := adaptEntityPatchOwnerError(err); err != nil {
			return MutationResult{}, err
		}
		return mutationResultFromEntityPatch(result, request.ViewSchemaID), nil
	}
	return s.applyWorkbookPatch(ctx, actor, recordID, request, requestHash, requestID, now, workbookPatchRouteKey)
}

func entityPatchRequestFromWorkbook(request PatchRequest) (hostidentity.PatchRequest, error) {
	changes := make([]hostidentity.PatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		if !isEntityDirectPatchField(request.ViewSchemaID, change.FieldKey) {
			return hostidentity.PatchRequest{}, mutationValidationError("field_key", "unsupported_field_key")
		}
		if change.Collection != nil || change.Value == nil {
			return hostidentity.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
		}
		var value *string
		switch change.Value.Kind {
		case "text":
			if change.Value.Text == nil {
				return hostidentity.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
			}
			value = change.Value.Text
		case "null":
			value = nil
		default:
			return hostidentity.PatchRequest{}, mutationValidationError(change.FieldKey, "invalid_value")
		}
		changes = append(changes, hostidentity.PatchChange{
			FieldKey: change.FieldKey,
			Value:    value,
		})
	}
	return hostidentity.PatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}, nil
}

func isEntityDirectPatchField(viewSchemaID string, fieldKey string) bool {
	switch viewSchemaID {
	case hostidentity.HostsViewSchemaID:
		switch fieldKey {
		case "host.display_name", "host.hostname", "host.aad_device_id", "host.fqdn",
			"host.location", "host.os_platform", "host.business_owner", "host.criticality", "host.containment_status":
			return true
		default:
			return false
		}
	case hostidentity.IdentitiesViewSchemaID:
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
	return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
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
	if artifacts.IsArtifactBackedView(claims.ViewSchemaID) {
		result, err := s.artifactMutations.Patch(ctx, artifacts.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          artifactPatchRequestFromWorkbook(patch),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookConflictResolveRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptArtifactWorkbookOwnerError(err)
		}
		return mutationResultFromArtifactWorkbook(result), nil
	}
	if claims.ViewSchemaID == EvidenceViewSchemaID {
		result, err := s.evidenceMutations.Patch(ctx, evidence.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          evidencePatchRequestFromWorkbook(patch),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookConflictResolveRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptEvidenceWorkbookOwnerError(err)
		}
		return mutationResultFromEvidenceWorkbook(result), nil
	}
	if claims.ViewSchemaID == PartiesViewSchemaID {
		result, err := s.partyMutations.Patch(ctx, parties.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          partyPatchRequestFromWorkbook(patch),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookConflictResolveRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptPartyWorkbookOwnerError(err)
		}
		return mutationResultFromPartyWorkbook(result), nil
	}
	if claims.ViewSchemaID == TaskRequestsViewSchemaID || claims.ViewSchemaID == DecisionsViewSchemaID {
		result, err := s.taskMutations.Patch(ctx, tasksdecisions.WorkbookPatchCommand{
			Actor:            actor,
			RecordID:         recordID,
			Request:          taskDecisionPatchRequestFromWorkbook(patch),
			RequestHash:      requestHash,
			RequestID:        requestID,
			RouteKey:         workbookConflictResolveRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              now,
		})
		if err != nil {
			return MutationResult{}, adaptTaskDecisionWorkbookOwnerError(err)
		}
		return mutationResultFromTaskDecisionWorkbook(result), nil
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

	target, err := s.recordTargets.ResolveTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if target.Deleted {
		return MutationResult{}, revisions.ErrRecordDeletedUseRestore
	}
	if !recordTypeMatchesView(target.RecordType, claims.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	if !s.projectionRows.Supports(claims.ViewSchemaID) {
		return MutationResult{}, fmt.Errorf("workbook mutation surface %q not mapped", claims.ViewSchemaID)
	}
	row, err := s.projectionRows.LoadRowTx(ctx, tx, claims.ViewSchemaID, recordID)
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
		IncidentID:   target.IncidentID,
		RecordID:     recordID,
		ClientTxnID:  request.ClientTxnID,
		RowVersion:   target.RowVersion,
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

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
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
