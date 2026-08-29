package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mutationadmission"
	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	workbookCreateOperation          = "workbook.rows.create"
	workbookPatchOperation           = "workbook.records.patch"
	workbookConflictResolveOperation = "workbook.records.conflicts.resolve"
)

type entityProviderSet struct {
	hostCreate       workbook.CreateProvider
	identityCreate   workbook.CreateProvider
	hostPatch        workbook.PatchProvider
	identityPatch    workbook.PatchProvider
	hostConflict     workbook.ConflictProvider
	identityConflict workbook.ConflictProvider
}

func newEntityProviderSet(owner *hostidentity.Store) (entityProviderSet, error) {
	if owner == nil {
		return entityProviderSet{}, fmt.Errorf("compose Entity Workbook adapters: owner is required")
	}
	hostCreate, err := newEntityCreateProvider(
		entitycontract.HostsViewSchemaID,
		owner.CreateHostRow,
	)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityCreate, err := newEntityCreateProvider(
		entitycontract.IdentitiesViewSchemaID,
		owner.CreateIdentityRow,
	)
	if err != nil {
		return entityProviderSet{}, err
	}
	hostPatch, err := newEntityPatchProvider("host", entitycontract.HostsViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityPatch, err := newEntityPatchProvider("identity", entitycontract.IdentitiesViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	hostConflict, err := newEntityConflictProvider("host", entitycontract.HostsViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityConflict, err := newEntityConflictProvider("identity", entitycontract.IdentitiesViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	return entityProviderSet{
		hostCreate: hostCreate, identityCreate: identityCreate,
		hostPatch: hostPatch, identityPatch: identityPatch,
		hostConflict: hostConflict, identityConflict: identityConflict,
	}, nil
}

type entityCreateFunc func(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	hostidentity.CreateRequest,
	[]byte,
	string,
	time.Time,
) (hostidentity.MutationResult, error)

func newEntityCreateProvider(viewSchemaID string, create entityCreateFunc) (workbook.CreateProvider, error) {
	if !isEntityViewSchema(viewSchemaID) || create == nil {
		return nil, fmt.Errorf("compose Entity create adapter: complete binding is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (hostidentity.CreateRequest, bool, *workbook.MutationFailure, error) {
			request, failure := hostidentity.DecodeCreateRequest(viewSchemaID, reader)
			if failure != nil {
				return hostidentity.CreateRequest{}, false, entityAdmissionFailure(failure), nil
			}
			return request, true, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand, request hostidentity.CreateRequest) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := create(
				ctx,
				command.Actor,
				command.IncidentID,
				request,
				hostidentity.CreateRequestHash(viewSchemaID, request),
				command.RequestID,
				command.Now,
			)
			if failure, safe := entityMutationFailure(err, request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(entityCreateResult(
				result,
				viewSchemaID,
				command.IncidentID,
				request.ClientTxnID,
			)), nil
		},
	)
}

func newEntityPatchProvider(
	recordType string,
	viewSchemaID string,
	owner *hostidentity.Store,
) (workbook.PatchProvider, error) {
	if !validEntityBinding(recordType, viewSchemaID) || owner == nil {
		return nil, fmt.Errorf("compose Entity patch adapter: complete binding is required")
	}
	return workbook.NewPatchProvider(
		func(reader io.Reader) (hostidentity.PatchRequest, bool, *workbook.MutationFailure, error) {
			request, failure := hostidentity.DecodePatchRequest(reader)
			if failure != nil {
				return hostidentity.PatchRequest{}, false, entityAdmissionFailure(failure), nil
			}
			return request, true, nil, nil
		},
		func(request hostidentity.PatchRequest) string { return request.ViewSchemaID },
		func(ctx context.Context, command workbook.PatchCommand, request hostidentity.PatchRequest) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != recordType || request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.PatchEntityRow(
				ctx,
				command.Actor,
				command.RecordID,
				request,
				hostidentity.PatchRequestHash(request),
				command.RequestID,
				command.Now,
				workbookPatchOperation,
			)
			if failure, safe := entityMutationFailure(err, request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(entityPatchResult(result, viewSchemaID)), nil
		},
	)
}

type entityConflictValue struct {
	request hostidentity.WorkbookConflictResolveRequest
	claims  hostidentity.WorkbookConflictClaims
}

func newEntityConflictProvider(
	recordType string,
	viewSchemaID string,
	owner *hostidentity.Store,
) (workbook.ConflictProvider, error) {
	if !validEntityBinding(recordType, viewSchemaID) || owner == nil {
		return nil, fmt.Errorf("compose Entity conflict adapter: complete binding is required")
	}
	return workbook.NewConflictProvider(
		func(
			reader io.Reader,
			token string,
			claims workbook.ConflictClaims,
		) (entityConflictValue, bool, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation ||
				claims.ViewSchemaID != viewSchemaID {
				return entityConflictValue{}, false, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := hostidentity.WorkbookConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, failure := hostidentity.DecodeWorkbookConflictResolveRequest(reader, token, ownerClaims)
			if failure != nil {
				return entityConflictValue{}, false, entityAdmissionFailure(failure), nil
			}
			return entityConflictValue{request: request, claims: ownerClaims}, true, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand, admitted entityConflictValue) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != recordType ||
				command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation ||
				command.Claims.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := hostidentity.WorkbookConflictResolveRequestHash(admitted.claims, admitted.request)
			result, err := owner.ResolveWorkbookConflict(ctx, hostidentity.ConflictCommand{
				Mechanics: conflicttokens.Command{
					ActorUserID: command.Actor.ID,
					RecordID:    command.RecordID,
					Claims:      entityConflictClaims(command.Claims),
					ClientTxnID: admitted.request.ClientTxnID,
					RequestHash: requestHash,
					RequestID:   command.RequestID,
					RouteKey:    command.Claims.RouteKey,
				},
				Actor:          command.Actor,
				ResolutionKind: admitted.request.ResolutionKind,
				Patch:          admitted.request.Patch,
				Now:            command.Now,
			})
			if failure, safe := entityMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(entityPatchResult(result, viewSchemaID)), nil
		},
	)
}

type entityClipboardValue struct {
	request hostidentity.ClipboardPasteRequest
	plan    tabularingest.TabularRowPlanV1
}

func newEntityClipboardProvider(viewSchemaID string, owner *hostidentity.Store) (workbook.ClipboardProvider, error) {
	if !isEntityViewSchema(viewSchemaID) || owner == nil {
		return nil, fmt.Errorf("compose Entity clipboard adapter: complete binding is required")
	}
	return workbook.NewClipboardProvider(
		func(reader io.Reader) (entityClipboardValue, bool, *workbook.MutationFailure, error) {
			request, failure := hostidentity.DecodeClipboardPasteRequest(reader, viewSchemaID)
			if failure != nil {
				return entityClipboardValue{}, false, entityAdmissionFailure(failure), nil
			}
			plan, err := hostidentity.BuildClipboardPastePlan(request)
			if err != nil {
				return entityClipboardValue{}, false, workbook.InvalidPayloadFailure("clipboard_text", "invalid_value"), nil
			}
			return entityClipboardValue{request: request, plan: plan}, true, nil, nil
		},
		func(ctx context.Context, command workbook.ClipboardCommand, admitted entityClipboardValue) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != viewSchemaID || admitted.request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.ApplyClipboardPastePlan(
				ctx,
				command.Actor,
				command.IncidentID,
				viewSchemaID,
				admitted.plan,
				admitted.request.RequestHash(),
				command.RequestID,
				command.Now,
			)
			if failure, safe := entityMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulBatchMutation(entityBatchResult(result)), nil
		},
	)
}

func entityMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
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
	if errors.Is(err, hostidentity.ErrInvalidCreateRequest) {
		return workbook.InvalidPayloadFailure("payload", "at_least_one_value_required"), true
	}
	if errors.Is(err, hostidentity.ErrNoEffectivePatchChange) {
		return workbook.NoEffectiveChangeFailure("changes"), true
	}
	if errors.Is(err, hostidentity.ErrInvalidAliasReference) {
		return workbook.InvalidPayloadFailure("action_payload", "invalid_value"), true
	}
	var rowConflict *hostidentity.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(
			rowConflict.RecordID,
			rowConflict.BaseRowVersion,
			rowConflict.CurrentRowVersion,
		), true
	}
	var matchConflict *hostidentity.ExactMatchConflictError
	if errors.As(err, &matchConflict) {
		return workbook.EntityMatchConflictFailure(
			matchConflict.EntityType,
			matchConflict.IdentifierClass,
			matchConflict.CandidateRecords,
		), true
	}
	return nil, false
}

func entityAdmissionFailure(failure *mutationadmission.Failure) *workbook.MutationFailure {
	if failure == nil {
		return nil
	}
	field, _ := failure.Field()
	collectionField, hasCollectionField := failure.CollectionField()
	requestedCount, hasRequestedCount := failure.RequestedCount()
	maximumCount, hasMaximumCount := failure.MaximumCount()
	reason := string(failure.ReasonCode())
	if hasRequestedCount || hasMaximumCount {
		return workbook.InvalidPayloadLimitFailure(
			field,
			reason,
			requestedCount,
			maximumCount,
			collectionField,
		)
	}
	if hasCollectionField {
		return workbook.InvalidPayloadCollectionFailure(field, reason, collectionField)
	}
	return workbook.InvalidPayloadFailure(field, reason)
}

func entityCreateResult(
	result hostidentity.MutationResult,
	viewSchemaID string,
	incidentID uuid.UUID,
	clientTxnID string,
) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: incidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: clientTxnID, RowVersion: result.RowVersion, ViewSchemaID: viewSchemaID,
	}
}

func entityPatchResult(result hostidentity.PatchMutationResult, viewSchemaID string) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: viewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func entityBatchResult(result hostidentity.ClipboardPasteResult) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, ChangeSetID: result.ChangeSetID, ClientTxnID: result.ClientTxnID,
	}
}

func entityConflictClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(),
		ViewSchemaID: claims.ViewSchemaID, RouteKey: claims.RouteKey,
		FieldKey: claims.FieldKey, ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion: claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}

func validEntityBinding(recordType string, viewSchemaID string) bool {
	return (recordType == "host" && viewSchemaID == entitycontract.HostsViewSchemaID) ||
		(recordType == "identity" && viewSchemaID == entitycontract.IdentitiesViewSchemaID)
}

func isEntityViewSchema(viewSchemaID string) bool {
	return viewSchemaID == entitycontract.HostsViewSchemaID ||
		viewSchemaID == entitycontract.IdentitiesViewSchemaID
}
