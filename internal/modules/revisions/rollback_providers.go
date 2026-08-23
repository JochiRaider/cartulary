package revisions

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func adaptRowRollbackProviderError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, rollbackcontract.ErrTargetNotFound):
		return ErrRollbackTargetNotFound
	case errors.Is(err, rollbackcontract.ErrStaleTarget):
		return &RollbackPreconditionError{ReasonCode: "stale_target"}
	case errors.Is(err, rollbackcontract.ErrTargetNotReversible):
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	case errors.Is(err, rollbackcontract.ErrIdentifierConflict):
		return &RollbackPreconditionError{ReasonCode: "active_entity_identifier_conflict"}
	default:
		return err
	}
}
