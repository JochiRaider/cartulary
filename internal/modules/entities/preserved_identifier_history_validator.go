package entities

import (
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/preservedhistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type preservedIdentifierHistoryValidator struct{}

func (preservedIdentifierHistoryValidator) ValidateHistoryMutation(mutation revisions.StoredMutation) error {
	if mutation.TargetKind != "entity_preserved_identifier" {
		return preservedhistory.ErrInvalidValue
	}
	return preservedhistory.ValidateMutation(mutation.TargetID, mutation.OperationKind, mutation.BeforeValue, mutation.AfterValue)
}
