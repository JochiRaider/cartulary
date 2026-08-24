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
	case errors.Is(err, rollbackcontract.ErrEntityIdentifierConflict):
		return &RollbackPreconditionError{ReasonCode: "active_entity_identifier_conflict"}
	case errors.Is(err, rollbackcontract.ErrPartyExactMatchKeyClaimed):
		return &RollbackPreconditionError{ReasonCode: "exact_match_key_claimed"}
	default:
		return err
	}
}
