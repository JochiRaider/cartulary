package workbook

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
)

func mutationResultFromEntityPatch(result hostidentity.PatchMutationResult, viewSchemaID string) MutationResult {
	return MutationResult{
		Payload:          result.Payload,
		StatusCode:       result.StatusCode,
		Replayed:         result.Replayed,
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		RowVersion:       result.RowVersion,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: result.ChangedFieldKeys,
	}
}

func adaptEntityPatchOwnerError(err error) error {
	if err == nil {
		return nil
	}
	var entityConflict *hostidentity.RowVersionConflictError
	switch {
	case errors.As(err, &entityConflict):
		return &RowVersionConflictError{
			RecordID:          entityConflict.RecordID,
			BaseRowVersion:    entityConflict.BaseRowVersion,
			CurrentRowVersion: entityConflict.CurrentRowVersion,
		}
	case errors.Is(err, hostidentity.ErrNoEffectivePatchChange):
		return mutationValidationError("changes", "no_effective_change")
	default:
		return err
	}
}
