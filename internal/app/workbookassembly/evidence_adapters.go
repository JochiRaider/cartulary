package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/google/uuid"
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

func newEvidenceCreateProvider(owner evidence.MutationContribution) (workbook.CreateProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence create adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (evidence.CreateAdmission, bool, *workbook.MutationFailure, error) {
			admission, admissionFailure := evidence.AdmitCreateJSON(reader)
			if admissionFailure != nil {
				return evidence.CreateAdmission{}, false, evidenceAdmissionFailure(admissionFailure), nil
			}
			return admission, true, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand, admission evidence.CreateAdmission) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != evidence.ViewSchemaID || admission.ViewSchemaID() != evidence.ViewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.Create(ctx, evidence.CreateCommand{
				ActorUserID: command.Actor.ID, IncidentID: command.IncidentID, Admission: admission,
				RequestID: command.RequestID, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admission.ClientTxnID()); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

func newEvidencePatchProvider(owner evidence.MutationContribution) (workbook.PatchProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence patch adapter: owner is required")
	}
	return workbook.NewPatchProvider(
		func(reader io.Reader) (evidence.PatchAdmission, bool, *workbook.MutationFailure, error) {
			admission, admissionFailure := evidence.AdmitPatchJSON(reader)
			if admissionFailure != nil {
				return evidence.PatchAdmission{}, false, evidenceAdmissionFailure(admissionFailure), nil
			}
			return admission, true, nil, nil
		},
		func(admission evidence.PatchAdmission) string { return admission.ViewSchemaID() },
		func(ctx context.Context, command workbook.PatchCommand, admission evidence.PatchAdmission) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "evidence" || admission.ViewSchemaID() != evidence.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.Patch(ctx, evidence.PatchCommand{
				ActorUserID: command.Actor.ID, RecordID: command.RecordID, Admission: admission,
				RequestID: command.RequestID, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admission.ClientTxnID()); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

type evidenceConflictValue struct {
	admission evidence.ConflictResolveAdmission
}

func newEvidenceConflictProvider(owner evidence.MutationContribution) (workbook.ConflictProvider, error) {
	if nilEvidenceOwner(owner) {
		return nil, fmt.Errorf("compose Evidence conflict adapter: owner is required")
	}
	return workbook.NewConflictProvider(
		func(
			reader io.Reader,
			token string,
			claims workbook.ConflictClaims,
		) (evidenceConflictValue, bool, *workbook.MutationFailure, error) {
			if claims.RouteKey != workbookConflictResolveOperation || claims.ViewSchemaID != evidence.ViewSchemaID {
				return evidenceConflictValue{}, false, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			ownerContext := evidence.ConflictAdmissionContext{
				Version: claims.Version, RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
				RouteKey: claims.RouteKey, FieldKey: claims.FieldKey,
				ConflictResolutionClass: claims.ConflictResolutionClass,
				BaseRowVersion:          claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
				OriginalRequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
			}
			admission, admissionFailure := evidence.AdmitConflictResolveJSON(reader, token, ownerContext)
			if admissionFailure != nil {
				return evidenceConflictValue{}, false, evidenceAdmissionFailure(admissionFailure), nil
			}
			return evidenceConflictValue{admission: admission}, true, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand, admitted evidenceConflictValue) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "evidence" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != workbookConflictResolveOperation || command.Claims.ViewSchemaID != evidence.ViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.ResolveConflict(ctx, evidence.ConflictCommand{
				ActorUserID: command.Actor.ID, Admission: admitted.admission,
				RequestID: command.RequestID, Now: command.Now,
			})
			if failure, safe := evidenceMutationFailure(err, admitted.admission.ClientTxnID()); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(evidenceMutationResult(result)), nil
		},
	)
}

func evidenceAdmissionFailure(admissionFailure *evidence.AdmissionFailure) *workbook.MutationFailure {
	if admissionFailure == nil {
		return nil
	}
	field, _ := admissionFailure.Field()
	reason := string(admissionFailure.ReasonCode())
	collectionField, hasCollectionField := admissionFailure.CollectionField()
	requested, hasRequested := admissionFailure.RequestedCount()
	maximum, hasMaximum := admissionFailure.MaximumCount()
	if hasRequested || hasMaximum {
		return workbook.InvalidPayloadLimitFailure(field, reason, requested, maximum, collectionField)
	}
	if hasCollectionField {
		return workbook.InvalidPayloadCollectionFailure(field, reason, collectionField)
	}
	return workbook.InvalidPayloadFailure(field, reason)
}

func evidenceMutationFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, evidence.ErrClientTxnConflict) || errors.Is(err, authn.ErrClientTxnConflict) ||
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
	status := http.StatusOK
	replayed := result.Outcome == evidence.MutationOutcomeReplayed
	if result.Outcome == evidence.MutationOutcomeCreated {
		status = http.StatusCreated
	}
	payload := map[string]any{"view_schema_id": result.ViewSchemaID, "row": result.Row}
	var changeSetID uuid.UUID
	if result.ChangeSetID != nil {
		changeSetID = *result.ChangeSetID
		payload["change_set_id"] = changeSetID.String()
	}
	return workbook.MutationResult{
		Payload: payload, StatusCode: status, Replayed: replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: changeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}
