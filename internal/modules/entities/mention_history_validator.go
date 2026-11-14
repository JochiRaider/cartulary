package entities

import (
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions/historyvalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type mentionHistoryValidator struct{}

func (mentionHistoryValidator) ValidateHistoryMutation(mutation revisions.StoredMutation) error {
	if mutation.TargetKind != "entity_mention" {
		return historyvalue.ErrInvalidValue
	}
	return historyvalue.ValidateMutation(mutation.TargetID, mutation.OperationKind, mutation.BeforeValue, mutation.AfterValue)
}
