package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/intervalhistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type intervalHistoryValidator struct{}

func (intervalHistoryValidator) ValidateHistoryMutation(mutation revisions.StoredMutation) error {
	if mutation.TargetKind != "indicator_state_interval" {
		return intervalhistory.ErrInvalidValue
	}
	return intervalhistory.ValidateMutation(mutation.TargetID, mutation.OperationKind, mutation.BeforeValue, mutation.AfterValue)
}
