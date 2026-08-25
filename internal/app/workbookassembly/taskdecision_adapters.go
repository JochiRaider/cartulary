package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type taskDecisionProviderSet struct {
	creates   map[string]workbook.CreateProvider
	patches   map[string]workbook.PatchProvider
	conflicts map[string]workbook.ConflictProvider
}

func newTaskDecisionProviderSet(owner *tasksdecisions.MutationFacade) (taskDecisionProviderSet, error) {
	if owner == nil {
		return taskDecisionProviderSet{}, fmt.Errorf("compose Tasks/Decisions Workbook adapters: owner is required")
	}
	result := taskDecisionProviderSet{
		creates:   make(map[string]workbook.CreateProvider, 2),
		patches:   make(map[string]workbook.PatchProvider, 2),
		conflicts: make(map[string]workbook.ConflictProvider, 2),
	}
	for _, binding := range []struct {
		viewSchemaID string
		recordType   string
	}{
		{tasksdecisions.TaskRequestsViewSchemaID, "task_request"},
		{tasksdecisions.DecisionsViewSchemaID, "decision"},
	} {
		create, err := newTaskDecisionCreateProvider(binding.viewSchemaID, owner)
		if err != nil {
			return taskDecisionProviderSet{}, err
		}
		patch, err := newTaskDecisionPatchProvider(binding.recordType, binding.viewSchemaID, owner)
		if err != nil {
			return taskDecisionProviderSet{}, err
		}
		conflict, err := newTaskDecisionConflictProvider(binding.recordType, binding.viewSchemaID, owner)
		if err != nil {
			return taskDecisionProviderSet{}, err
		}
		result.creates[binding.viewSchemaID] = create
		result.patches[binding.recordType] = patch
		result.conflicts[binding.recordType] = conflict
	}
	return result, nil
}

type taskDecisionCreateAdmission struct{ request tasksdecisions.CreateRequest }

func (value taskDecisionCreateAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
}

func newTaskDecisionCreateProvider(viewSchemaID string, owner *tasksdecisions.MutationFacade) (workbook.CreateProvider, error) {
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			request, admissionFailure := tasksdecisions.AdmitCreateJSON(viewSchemaID, reader)
			if admissionFailure != nil {
				return nil, taskDecisionAdmissionFailure(admissionFailure), nil
			}
			return taskDecisionCreateAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(taskDecisionCreateAdmission)
			if !ok || command.ViewSchemaID != viewSchemaID || admitted.request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id")), nil
			}
			result, err := owner.Create(ctx, tasksdecisions.CreateCommand{
				ActorUserID: command.Actor.ID, IncidentID: command.IncidentID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, tasksdecisions.CreateRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookCreateOperation, Now: command.Now,
			})
			if failure, safe := taskDecisionMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(taskDecisionMutationResult(result)), nil
		},
	)
}

type taskDecisionPatchAdmission struct{ request tasksdecisions.PatchRequest }

func (value taskDecisionPatchAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
}
func (value taskDecisionPatchAdmission) AdmittedViewSchemaID() string {
	return value.request.ViewSchemaID
}
func (value taskDecisionPatchAdmission) AdmittedBaseRowVersion() int64 {
	return value.request.BaseRowVersion
}

func newTaskDecisionPatchProvider(recordType, viewSchemaID string, owner *tasksdecisions.MutationFacade) (workbook.PatchProvider, error) {
	return workbook.NewPatchProvider(
		func(reader io.Reader) (workbook.PatchAdmission, *workbook.MutationFailure, error) {
			request, admissionFailure := tasksdecisions.AdmitPatchJSON(reader)
			if admissionFailure != nil {
				return nil, taskDecisionAdmissionFailure(admissionFailure), nil
			}
			return taskDecisionPatchAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.PatchCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(taskDecisionPatchAdmission)
			if !ok || command.AuthoritativeRecordType != recordType || admitted.request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.Patch(ctx, tasksdecisions.PatchCommand{
				ActorUserID: command.Actor.ID, RecordID: command.RecordID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, tasksdecisions.PatchRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookPatchOperation,
				ConflictRouteKey: workbookConflictResolveOperation, Now: command.Now,
			})
			if failure, safe := taskDecisionMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(taskDecisionMutationResult(result)), nil
		},
	)
}

type taskDecisionConflictAdmission struct {
	request tasksdecisions.ConflictResolveRequest
	claims  tasksdecisions.ConflictClaims
}

func (value taskDecisionConflictAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
}

func newTaskDecisionConflictProvider(recordType, viewSchemaID string, owner *tasksdecisions.MutationFacade) (workbook.ConflictProvider, error) {
	return workbook.NewConflictProvider(
		func(reader io.Reader, token string, claims workbook.ConflictClaims) (workbook.ConflictAdmission, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation || claims.ViewSchemaID != viewSchemaID {
				return nil, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := tasksdecisions.ConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, admissionFailure := tasksdecisions.AdmitConflictResolveJSON(reader, token, ownerClaims)
			if admissionFailure != nil {
				return nil, taskDecisionAdmissionFailure(admissionFailure), nil
			}
			return taskDecisionConflictAdmission{request: request, claims: ownerClaims}, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(taskDecisionConflictAdmission)
			if !ok || command.AuthoritativeRecordType != recordType || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation || command.Claims.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := preferredRequestHash(command.RequestHash, tasksdecisions.ConflictResolveRequestHash(admitted.claims, admitted.request))
			result, err := owner.ResolveConflict(ctx, tasksdecisions.ConflictCommand{
				Mechanics: conflicttokens.Command{
					ActorUserID: command.Actor.ID, RecordID: command.RecordID,
					Claims: taskDecisionConflictTokenClaims(command.Claims), ClientTxnID: admitted.request.ClientTxnID,
					RequestHash: requestHash, RequestID: command.RequestID, RouteKey: command.Claims.RouteKey,
				},
				ResolutionKind: admitted.request.ResolutionKind, Patch: admitted.request.Patch, Now: command.Now,
			})
			if failure, safe := taskDecisionMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(taskDecisionMutationResult(result)), nil
		},
	)
}

func taskDecisionAdmissionFailure(failure *tasksdecisions.AdmissionFailure) *workbook.MutationFailure {
	if failure == nil {
		return nil
	}
	fieldKey, hasFieldKey := failure.CollectionFieldKey()
	if requestedCount, maxCount, hasLimit := failure.CountLimit(); hasLimit {
		return workbook.InvalidPayloadLimitFailure(
			failure.Field(), failure.ReasonCode(), requestedCount, maxCount, fieldKey,
		)
	}
	if hasFieldKey {
		return workbook.InvalidPayloadCollectionFailure(
			failure.Field(), failure.ReasonCode(), fieldKey,
		)
	}
	return workbook.InvalidPayloadFailure(failure.Field(), failure.ReasonCode())
}

func taskDecisionMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, tasksdecisions.ErrClientTxnConflict) || errors.Is(err, authn.ErrClientTxnConflict) ||
		errors.Is(err, conflicttokens.ErrClientTxnConflict) {
		return workbook.ClientTxnConflictFailure(clientTxnID), true
	}
	if incidentadmission.IsDenied(err, incidentadmission.DenialIncidentClosed) {
		return workbook.IncidentClosedFailure(), true
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return workbook.TargetNotFoundFailure(), true
	}
	if errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
		return workbook.RecordDeletedFailure(), true
	}
	var validation *tasksdecisions.ValidationError
	if errors.As(err, &validation) {
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	var collection *links.CollectionValidationError
	if errors.As(err, &collection) {
		return workbook.InvalidPayloadFailure(collection.Field, collection.ReasonCode), true
	}
	var lifecycle *tasksdecisions.LifecycleValidationError
	if errors.As(err, &lifecycle) {
		return workbook.IllegalTransitionFailure(lifecycle.FromStatus, lifecycle.ToStatus, lifecycle.ReasonCode, lifecycle.ViolatedGuards), true
	}
	var rowConflict *tasksdecisions.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion), true
	}
	var sameConflict *tasksdecisions.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		conflict := sameConflict.Conflict
		failure, conversionErr := workbook.SameFieldConflictFailure(workbook.SameFieldConflictInput{
			ConflictToken: conflict.ConflictToken, RecordID: conflict.RecordID, FieldKey: conflict.FieldKey,
			ConflictResolutionClass: workbook.SameFieldConflictClass(conflict.ConflictResolutionClass),
			BaseRowVersion:          conflict.BaseRowVersion, CurrentRowVersion: conflict.CurrentRowVersion,
			ClientValue: conflict.ClientValue, ServerValue: conflict.ServerValue,
			BaseValue:       workbook.OptionalConflictValue{Present: conflict.BaseValue.Present, Value: conflict.BaseValue.Value},
			ServerUpdatedBy: conflict.ServerUpdatedBy, ServerUpdatedAt: conflict.ServerUpdatedAt,
			SuggestedMergedValue: workbook.OptionalConflictValue{
				Present: conflict.SuggestedMergedValue.Present, Value: conflict.SuggestedMergedValue.Value,
			},
		})
		return failure, conversionErr == nil
	}
	return nil, false
}

func taskDecisionMutationResult(result tasksdecisions.MutationResult) workbook.MutationResult {
	payload := map[string]any{"view_schema_id": result.ViewSchemaID, "row": result.Row}
	if result.ChangeSetID != uuid.Nil {
		payload["change_set_id"] = result.ChangeSetID.String()
	}
	status := http.StatusOK
	if result.Created && !result.Replayed {
		status = http.StatusCreated
	}
	return workbook.MutationResult{
		Payload: payload, StatusCode: status, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func taskDecisionConflictTokenClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(), ViewSchemaID: claims.ViewSchemaID,
		RouteKey: claims.RouteKey, FieldKey: claims.FieldKey,
		ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion:          claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}
