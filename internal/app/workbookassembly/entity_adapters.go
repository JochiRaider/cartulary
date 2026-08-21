package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
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
		hostidentity.HostsViewSchemaID,
		owner.CreateHostRow,
	)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityCreate, err := newEntityCreateProvider(
		hostidentity.IdentitiesViewSchemaID,
		owner.CreateIdentityRow,
	)
	if err != nil {
		return entityProviderSet{}, err
	}
	hostPatch, err := newEntityPatchProvider("host", hostidentity.HostsViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityPatch, err := newEntityPatchProvider("identity", hostidentity.IdentitiesViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	hostConflict, err := newEntityConflictProvider("host", hostidentity.HostsViewSchemaID, owner)
	if err != nil {
		return entityProviderSet{}, err
	}
	identityConflict, err := newEntityConflictProvider("identity", hostidentity.IdentitiesViewSchemaID, owner)
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

type entityCreateAdmission struct {
	request hostidentity.CreateRequest
}

func (value entityCreateAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
}

func newEntityCreateProvider(viewSchemaID string, create entityCreateFunc) (workbook.CreateProvider, error) {
	if !isEntityViewSchema(viewSchemaID) || create == nil {
		return nil, fmt.Errorf("compose Entity create adapter: complete binding is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			request, apiErr := hostidentity.DecodeCreateRequest(viewSchemaID, reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return entityCreateAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(entityCreateAdmission)
			if !ok || command.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := create(
				ctx,
				command.Actor,
				command.IncidentID,
				admitted.request,
				preferredRequestHash(command.RequestHash, hostidentity.CreateRequestHash(viewSchemaID, admitted.request)),
				command.RequestID,
				command.Now,
			)
			if failure, safe := entityMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(entityCreateResult(
				result,
				viewSchemaID,
				command.IncidentID,
				admitted.request.ClientTxnID,
			)), nil
		},
	)
}

type entityPatchAdmission struct {
	request hostidentity.PatchRequest
}

func (value entityPatchAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
}

func (value entityPatchAdmission) AdmittedViewSchemaID() string {
	return value.request.ViewSchemaID
}

func (value entityPatchAdmission) AdmittedBaseRowVersion() int64 {
	return value.request.BaseRowVersion
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
		func(reader io.Reader) (workbook.PatchAdmission, *workbook.MutationFailure, error) {
			request, apiErr := hostidentity.DecodePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return entityPatchAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.PatchCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(entityPatchAdmission)
			if !ok || command.AuthoritativeRecordType != recordType || admitted.request.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.PatchEntityRow(
				ctx,
				command.Actor,
				command.RecordID,
				admitted.request,
				preferredRequestHash(command.RequestHash, hostidentity.PatchRequestHash(admitted.request)),
				command.RequestID,
				command.Now,
				workbookPatchOperation,
			)
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

type entityConflictAdmission struct {
	request hostidentity.WorkbookConflictResolveRequest
	claims  hostidentity.WorkbookConflictClaims
}

func (value entityConflictAdmission) ClientTransactionID() string {
	return value.request.ClientTxnID
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
		) (workbook.ConflictAdmission, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation ||
				claims.ViewSchemaID != viewSchemaID {
				return nil, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := hostidentity.WorkbookConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, apiErr := hostidentity.DecodeWorkbookConflictResolveRequest(reader, token, ownerClaims)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return entityConflictAdmission{request: request, claims: ownerClaims}, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(entityConflictAdmission)
			if !ok || command.AuthoritativeRecordType != recordType ||
				command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation ||
				command.Claims.ViewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := preferredRequestHash(
				command.RequestHash,
				hostidentity.WorkbookConflictResolveRequestHash(admitted.claims, admitted.request),
			)
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
		func(reader io.Reader) (workbook.ClipboardAdmission, *workbook.MutationFailure, error) {
			request, apiErr := hostidentity.DecodeClipboardPasteRequest(reader, viewSchemaID)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			plan, err := hostidentity.BuildClipboardPastePlan(request)
			if err != nil {
				return nil, workbook.InvalidPayloadFailure("clipboard_text", "invalid_value"), nil
			}
			return clipboardAdmission[entityClipboardValue]{
				clientTxnID:  request.ClientTxnID,
				viewSchemaID: request.ViewSchemaID,
				value:        entityClipboardValue{request: request, plan: plan},
			}, nil, nil
		},
		func(ctx context.Context, command workbook.ClipboardCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(clipboardAdmission[entityClipboardValue])
			if !ok || command.ViewSchemaID != viewSchemaID || admitted.viewSchemaID != viewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.ApplyClipboardPastePlan(
				ctx,
				command.Actor,
				command.IncidentID,
				viewSchemaID,
				admitted.value.plan,
				admitted.value.request.RequestHash(),
				command.RequestID,
				command.Now,
			)
			if failure, safe := entityMutationFailure(err, admitted.clientTxnID); safe {
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
	return (recordType == "host" && viewSchemaID == hostidentity.HostsViewSchemaID) ||
		(recordType == "identity" && viewSchemaID == hostidentity.IdentitiesViewSchemaID)
}

func isEntityViewSchema(viewSchemaID string) bool {
	return viewSchemaID == hostidentity.HostsViewSchemaID ||
		viewSchemaID == hostidentity.IdentitiesViewSchemaID
}
