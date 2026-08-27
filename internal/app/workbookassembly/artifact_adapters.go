package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var artifactViewSchemaIDs = []string{
	artifacts.CommLogViewSchemaID,
	artifacts.FindingsViewSchemaID,
	artifacts.ForensicKeywordsViewSchemaID,
	artifacts.HandoffViewSchemaID,
	artifacts.InvestigativeQueriesViewSchemaID,
	artifacts.LessonViewSchemaID,
	artifacts.NotesViewSchemaID,
	artifacts.StatusReviewViewSchemaID,
}

type artifactProviderSet struct {
	creates  map[string]workbook.CreateProvider
	patch    workbook.PatchProvider
	conflict workbook.ConflictProvider
}

func newArtifactProviderSet(owner *artifacts.MutationFacade) (artifactProviderSet, error) {
	if owner == nil {
		return artifactProviderSet{}, fmt.Errorf("compose Artifact Workbook adapters: owner is required")
	}
	creates := make(map[string]workbook.CreateProvider, len(artifactViewSchemaIDs))
	for _, viewSchemaID := range artifactViewSchemaIDs {
		provider, err := newArtifactCreateProvider(viewSchemaID, owner)
		if err != nil {
			return artifactProviderSet{}, err
		}
		creates[viewSchemaID] = provider
	}
	patch, err := newArtifactPatchProvider(owner)
	if err != nil {
		return artifactProviderSet{}, err
	}
	conflict, err := newArtifactConflictProvider(owner)
	if err != nil {
		return artifactProviderSet{}, err
	}
	return artifactProviderSet{creates: creates, patch: patch, conflict: conflict}, nil
}

func newArtifactCreateProvider(viewSchemaID string, owner *artifacts.MutationFacade) (workbook.CreateProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Artifact create adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (artifacts.CreateRequest, bool, *workbook.MutationFailure, error) {
			request, apiErr := artifacts.DecodeCreateRequest(viewSchemaID, reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return artifacts.CreateRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand, request artifacts.CreateRequest) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != viewSchemaID || request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.Create(ctx, artifacts.CreateCommand{
				ActorUserID: command.Actor.ID, IncidentID: command.IncidentID, Request: request,
				RequestHash: artifacts.CreateRequestHash(request),
				RequestID:   command.RequestID, OperationID: artifacts.OperationCreate, Now: command.Now,
			})
			if failure, safe := artifactMutationFailure(err, request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(artifactMutationResult(result)), nil
		},
	)
}

func newArtifactPatchProvider(owner *artifacts.MutationFacade) (workbook.PatchProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Artifact patch adapter: owner is required")
	}
	return workbook.NewPatchProvider(
		func(reader io.Reader) (artifacts.PatchRequest, bool, *workbook.MutationFailure, error) {
			request, apiErr := artifacts.DecodePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return artifacts.PatchRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(request artifacts.PatchRequest) string { return request.ViewSchemaID },
		func(ctx context.Context, command workbook.PatchCommand, request artifacts.PatchRequest) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "artifact" {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.Patch(ctx, artifacts.PatchCommand{
				ActorUserID: command.Actor.ID, RecordID: command.RecordID, Request: request,
				RequestHash: artifacts.PatchRequestHash(request),
				RequestID:   command.RequestID, OperationID: artifacts.OperationPatch,
				ConflictOperationID: artifacts.OperationConflictResolve, Now: command.Now,
			})
			if failure, safe := artifactMutationFailure(err, request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(artifactMutationResult(result)), nil
		},
	)
}

type artifactConflictValue struct {
	request artifacts.ConflictResolveRequest
	claims  artifacts.ConflictClaims
}

func newArtifactConflictProvider(owner *artifacts.MutationFacade) (workbook.ConflictProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Artifact conflict adapter: owner is required")
	}
	return workbook.NewConflictProvider(
		func(
			reader io.Reader,
			token string,
			claims workbook.ConflictClaims,
		) (artifactConflictValue, bool, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation || !artifactViewSchemaID(claims.ViewSchemaID) {
				return artifactConflictValue{}, false, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := artifacts.ConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, apiErr := artifacts.DecodeConflictResolveRequest(reader, token, ownerClaims)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return artifactConflictValue{}, false, failure, err
			}
			return artifactConflictValue{request: request, claims: ownerClaims}, true, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand, admitted artifactConflictValue) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "artifact" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation || !artifactViewSchemaID(command.Claims.ViewSchemaID) {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := artifacts.ConflictResolveRequestHash(admitted.claims, admitted.request)
			result, err := owner.ResolveConflict(ctx, artifacts.ConflictCommand{
				Mechanics: conflicttokens.Command{
					ActorUserID: command.Actor.ID, RecordID: command.RecordID,
					Claims: artifactConflictTokenClaims(command.Claims), ClientTxnID: admitted.request.ClientTxnID,
					RequestHash: requestHash, RequestID: command.RequestID, RouteKey: command.Claims.RouteKey,
				},
				ActorUserID: command.Actor.ID, OperationID: artifacts.OperationConflictResolve,
				ResolutionKind: admitted.request.ResolutionKind, Patch: admitted.request.Patch, Now: command.Now,
			})
			if failure, safe := artifactMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(artifactMutationResult(result)), nil
		},
	)
}

func artifactViewSchemaID(viewSchemaID string) bool {
	for _, candidate := range artifactViewSchemaIDs {
		if viewSchemaID == candidate {
			return true
		}
	}
	return false
}

func artifactMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, artifacts.ErrClientTxnConflict) || errors.Is(err, authn.ErrClientTxnConflict) ||
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
	var validation *artifacts.ValidationError
	if errors.As(err, &validation) {
		if validation.ReasonCode == "no_effective_change" {
			return workbook.NoEffectiveChangeFailure(validation.Field), true
		}
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	var collection *links.CollectionValidationError
	if errors.As(err, &collection) {
		return workbook.InvalidPayloadFailure(collection.Field, collection.ReasonCode), true
	}
	var rowConflict *artifacts.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(
			rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion,
		), true
	}
	var sameConflict *artifacts.SameFieldConflictError
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

func artifactConflictTokenClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(), ViewSchemaID: claims.ViewSchemaID,
		RouteKey: claims.RouteKey, FieldKey: claims.FieldKey,
		ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion:          claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}
