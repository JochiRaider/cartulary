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
	default:
		return err
	}
}
