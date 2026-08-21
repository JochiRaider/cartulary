package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"

	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type partyProviderSet struct {
	create   workbook.CreateProvider
	patch    workbook.PatchProvider
	conflict workbook.ConflictProvider
}

func newPartyProviderSet(owner *parties.MutationFacade) (partyProviderSet, error) {
	if owner == nil {
		return partyProviderSet{}, fmt.Errorf("compose Party Workbook adapters: owner is required")
	}
	create, err := newPartyCreateProvider(owner)
	if err != nil {
		return partyProviderSet{}, err
	}
	patch, err := newPartyPatchProvider(owner)
	if err != nil {
		return partyProviderSet{}, err
	}
	conflict, err := newPartyConflictProvider(owner)
	if err != nil {
		return partyProviderSet{}, err
	}
	return partyProviderSet{create: create, patch: patch, conflict: conflict}, nil
}

type partyCreateAdmission struct{ request parties.CreateRequest }

func (value partyCreateAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newPartyCreateProvider(owner *parties.MutationFacade) (workbook.CreateProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Party create adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			request, apiErr := parties.DecodeCreateRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return partyCreateAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(partyCreateAdmission)
			if !ok || command.ViewSchemaID != parties.ViewSchemaID || admitted.request.ViewSchemaID != parties.ViewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.Create(ctx, parties.CreateCommand{
				Actor: command.Actor, IncidentID: command.IncidentID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, parties.CreateRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookCreateOperation, Now: command.Now,
			})
			if failure, safe := partyMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(partyMutationResult(result)), nil
		},
	)
}

type partyPatchAdmission struct{ request parties.PatchRequest }

func (value partyPatchAdmission) ClientTransactionID() string   { return value.request.ClientTxnID }
func (value partyPatchAdmission) AdmittedViewSchemaID() string  { return value.request.ViewSchemaID }
func (value partyPatchAdmission) AdmittedBaseRowVersion() int64 { return value.request.BaseRowVersion }

func newPartyPatchProvider(owner *parties.MutationFacade) (workbook.PatchProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Party patch adapter: owner is required")
	}
	return workbook.NewPatchProvider(
		func(reader io.Reader) (workbook.PatchAdmission, *workbook.MutationFailure, error) {
			request, apiErr := parties.DecodePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return partyPatchAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.PatchCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(partyPatchAdmission)
			if !ok || command.AuthoritativeRecordType != "party" || admitted.request.ViewSchemaID != parties.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.Patch(ctx, parties.PatchCommand{
				Actor: command.Actor, RecordID: command.RecordID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, parties.PatchRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookPatchOperation,
				ConflictRouteKey: workbookConflictResolveOperation, Now: command.Now,
			})
			if failure, safe := partyMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(partyMutationResult(result)), nil
		},
	)
}

type partyConflictAdmission struct {
	request parties.ConflictResolveRequest
	claims  parties.ConflictClaims
}

func (value partyConflictAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newPartyConflictProvider(owner *parties.MutationFacade) (workbook.ConflictProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Party conflict adapter: owner is required")
	}
	return workbook.NewConflictProvider(
		func(
			reader io.Reader,
			token string,
			claims workbook.ConflictClaims,
		) (workbook.ConflictAdmission, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation || claims.ViewSchemaID != parties.ViewSchemaID {
				return nil, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := parties.ConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, apiErr := parties.DecodeConflictResolveRequest(reader, token, ownerClaims)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return partyConflictAdmission{request: request, claims: ownerClaims}, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(partyConflictAdmission)
			if !ok || command.AuthoritativeRecordType != "party" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation || command.Claims.ViewSchemaID != parties.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := preferredRequestHash(
				command.RequestHash,
				parties.ConflictResolveRequestHash(admitted.claims, admitted.request),
			)
			result, err := owner.ResolveConflict(ctx, parties.ConflictCommand{
				Mechanics: conflicttokens.Command{
					ActorUserID: command.Actor.ID, RecordID: command.RecordID,
					Claims: partyConflictClaims(command.Claims), ClientTxnID: admitted.request.ClientTxnID,
					RequestHash: requestHash, RequestID: command.RequestID, RouteKey: command.Claims.RouteKey,
				},
				Actor: command.Actor, ResolutionKind: admitted.request.ResolutionKind,
				Patch: admitted.request.Patch, Now: command.Now,
			})
			if failure, safe := partyMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(partyMutationResult(result)), nil
		},
	)
}

func partyMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, authn.ErrClientTxnConflict) || errors.Is(err, conflicttokens.ErrClientTxnConflict) {
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
	var validation *parties.ValidationError
	if errors.As(err, &validation) {
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	var rowConflict *parties.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(
			rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion,
		), true
	}
	var sameConflict *parties.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		conflict := sameConflict.Conflict
		failure, conversionErr := workbook.SameFieldConflictFailure(workbook.SameFieldConflictInput{
			ConflictToken: conflict.ConflictToken, RecordID: conflict.RecordID,
			FieldKey:                conflict.FieldKey,
			ConflictResolutionClass: workbook.SameFieldConflictClass(conflict.ConflictResolutionClass),
			BaseRowVersion:          conflict.BaseRowVersion, CurrentRowVersion: conflict.CurrentRowVersion,
			ClientValue: conflict.ClientValue, ServerValue: conflict.ServerValue,
			BaseValue: workbook.OptionalConflictValue{
				Present: conflict.BaseValue.Present, Value: conflict.BaseValue.Value,
			},
			ServerUpdatedBy: conflict.ServerUpdatedBy, ServerUpdatedAt: conflict.ServerUpdatedAt,
			SuggestedMergedValue: workbook.OptionalConflictValue{
				Present: conflict.SuggestedMergedValue.Present, Value: conflict.SuggestedMergedValue.Value,
			},
		})
		return failure, conversionErr == nil
	}
	return nil, false
}

func partyMutationResult(result parties.MutationResult) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func partyConflictClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(),
		ViewSchemaID: claims.ViewSchemaID, RouteKey: claims.RouteKey,
		FieldKey: claims.FieldKey, ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion: claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}
