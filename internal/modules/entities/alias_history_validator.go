package entities

import (
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/aliashistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type aliasHistoryValidator struct{}

func (aliasHistoryValidator) ValidateHistoryMutation(mutation revisions.StoredMutation) error {
	if mutation.TargetKind != "entity_alias" {
		return aliashistory.ErrInvalidValue
	}
	return aliashistory.ValidateMutation(mutation.TargetID, mutation.OperationKind, mutation.BeforeValue, mutation.AfterValue)
}
