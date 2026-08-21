package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type evidenceProviderSet struct {
	create   workbook.CreateProvider
	patch    workbook.PatchProvider
	conflict workbook.ConflictProvider
}

func newEvidenceProviderSet(owner evidence.MutationContribution) (evidenceProviderSet, error) {
	if nilEvidenceOwner(owner) {
		return evidenceProviderSet{}, fmt.Errorf("compose Evidence Workbook adapters: owner is required")
	}
	create, err := newEvidenceCreateProvider(owner)
	if err != nil {
		return evidenceProviderSet{}, err
	}
	patch, err := newEvidencePatchProvider(owner)
	if err != nil {
		return evidenceProviderSet{}, err
	}
	conflict, err := newEvidenceConflictProvider(owner)
	if err != nil {
		return evidenceProviderSet{}, err
	}
	return evidenceProviderSet{create: create, patch: patch, conflict: conflict}, nil
}

func nilEvidenceOwner(owner evidence.MutationContribution) bool {
	if owner == nil {
		return true
	}
	value := reflect.ValueOf(owner)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

type evidenceCreateAdmission struct{ request evidence.CreateRequest }

func (value evidenceCreateAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newEvidenceCreateProvider(owner evidence.MutationContribution) (workbook.CreateProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence create adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			request, apiErr := evidence.DecodeCreateRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return evidenceCreateAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(evidenceCreateAdmission)
			if !ok || command.ViewSchemaID != evidence.ViewSchemaID || admitted.request.ViewSchemaID != evidence.ViewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.Create(ctx, evidence.CreateCommand{
				Actor: command.Actor, IncidentID: command.IncidentID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, evidence.CreateRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookCreateOperation, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

type evidencePatchAdmission struct{ request evidence.PatchRequest }

func (value evidencePatchAdmission) ClientTransactionID() string  { return value.request.ClientTxnID }
func (value evidencePatchAdmission) AdmittedViewSchemaID() string { return value.request.ViewSchemaID }
func (value evidencePatchAdmission) AdmittedBaseRowVersion() int64 {
	return value.request.BaseRowVersion
}

func newEvidencePatchProvider(owner evidence.MutationContribution) (workbook.PatchProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence patch adapter: owner is required")
	}
	return workbook.NewPatchProvider(
		func(reader io.Reader) (workbook.PatchAdmission, *workbook.MutationFailure, error) {
			request, apiErr := evidence.DecodePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return evidencePatchAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.PatchCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(evidencePatchAdmission)
			if !ok || command.AuthoritativeRecordType != "evidence" || admitted.request.ViewSchemaID != evidence.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.Patch(ctx, evidence.PatchCommand{
				Actor: command.Actor, RecordID: command.RecordID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, evidence.PatchRequestHash(admitted.request)),
				RequestID:   command.RequestID, RouteKey: workbookPatchOperation,
				ConflictRouteKey: workbookConflictResolveOperation, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

type evidenceConflictAdmission struct {
	request evidence.ConflictResolveRequest
	claims  evidence.ConflictClaims
}

func (value evidenceConflictAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newEvidenceConflictProvider(owner evidence.MutationContribution) (workbook.ConflictProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence conflict adapter: owner is required")
	}
	return workbook.NewConflictProvider(
		func(
			reader io.Reader,
			token string,
			claims workbook.ConflictClaims,
		) (workbook.ConflictAdmission, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation || claims.ViewSchemaID != evidence.ViewSchemaID {
				return nil, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerClaims := evidence.ConflictClaims{
				RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				FieldKey: claims.FieldKey, CurrentRowVersion: claims.CurrentRowVersion,
			}
			request, apiErr := evidence.DecodeConflictResolveRequest(reader, token, ownerClaims)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return evidenceConflictAdmission{request: request, claims: ownerClaims}, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(evidenceConflictAdmission)
			if !ok || command.AuthoritativeRecordType != "evidence" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation || command.Claims.ViewSchemaID != evidence.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			requestHash := preferredRequestHash(
				command.RequestHash,
				evidence.ConflictResolveRequestHash(admitted.claims, admitted.request),
			)
			result, err := owner.ResolveConflict(ctx, evidence.ConflictCommand{
				Mechanics: conflicttokens.Command{
					ActorUserID: command.Actor.ID, RecordID: command.RecordID,
					Claims: evidenceConflictClaims(command.Claims), ClientTxnID: admitted.request.ClientTxnID,
					RequestHash: requestHash, RequestID: command.RequestID, RouteKey: command.Claims.RouteKey,
				},
				Actor: command.Actor, ResolutionKind: admitted.request.ResolutionKind,
				Patch: admitted.request.Patch, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admitted.request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

func evidenceMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
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
	var validation *evidence.ValidationError
	if errors.As(err, &validation) {
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	var lifecycle *evidence.LifecycleValidationError
	if errors.As(err, &lifecycle) {
		return workbook.IllegalTransitionFailure(
			lifecycle.FromStatus,
			lifecycle.ToStatus,
			lifecycle.ReasonCode,
			append([]string(nil), lifecycle.ViolatedGuards...),
		), true
	}
	var rowConflict *evidence.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(
			rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion,
		), true
	}
	var sameConflict *evidence.SameFieldConflictError
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
	var attachRejected evidence.AttachRejectedError
	if errors.As(err, &attachRejected) {
		return workbook.EvidenceAttachFailure(attachRejected.ReasonCode), true
	}
	if errors.Is(err, evidence.ErrBlobNotFound) || errors.Is(err, evidence.ErrIncidentMismatch) {
		return workbook.EvidenceAttachFailure(evidence.AttachReasonBlobNotVisible), true
	}
	if errors.Is(err, evidence.ErrObjectStoreUnavailable) {
		return workbook.ObjectStoreUnavailableFailure("dependency_unavailable"), true
	}
	if reasonCode, ok := evidence.PersistedObjectBlobStorageKeyErrorReason(err); ok {
		return workbook.ObjectStoreInvalidFailure(reasonCode), true
	}
	adapterError, ok := objectstore.AsAdapterError(err)
	if !ok {
		return nil, false
	}
	switch adapterError.Code {
	case objectstore.ErrorCodeAccessRejected:
		return workbook.ObjectStoreAccessRejectedFailure(evidenceAccessRejectedReason(adapterError)), true
	case objectstore.ErrorCodeUnavailable:
		return workbook.ObjectStoreUnavailableFailure(evidenceUnavailableReason(adapterError)), true
	case objectstore.ErrorCodeDeadlineExceeded, objectstore.ErrorCodeRetryExhausted:
		return workbook.ObjectStoreUnavailableFailure("retry_exhausted"), true
	default:
		return workbook.ObjectStoreInvalidFailure("invalid_request"), true
	}
}

func evidenceUnavailableReason(adapterError *objectstore.AdapterError) string {
	switch adapterError.Reason {
	case objectstore.ReasonBucketMissing:
		return "bucket_missing"
	case objectstore.ReasonRetryExhausted, objectstore.ReasonDeadlineExceeded:
		return "retry_exhausted"
	default:
		return "endpoint_unreachable"
	}
}

func evidenceAccessRejectedReason(adapterError *objectstore.AdapterError) string {
	switch adapterError.Reason {
	case objectstore.ReasonCredentialDenied:
		return "credential_denied"
	case objectstore.ReasonCORSRejected:
		return "cors_rejected"
	default:
		return "capability_missing"
	}
}

func evidenceMutationResult(result evidence.MutationResult) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func evidenceConflictClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(),
		ViewSchemaID: claims.ViewSchemaID, RouteKey: claims.RouteKey,
		FieldKey: claims.FieldKey, ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion: claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}
