package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/observationhistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type observationHistoryValidator struct{}

func (observationHistoryValidator) ValidateHistoryMutation(m revisions.StoredMutation) error {
	if m.TargetKind != "indicator_observation" {
		return observationhistory.ErrInvalidValue
	}
	return observationhistory.ValidateMutation(m.TargetID, m.OperationKind, m.BeforeValue, m.AfterValue)
}
