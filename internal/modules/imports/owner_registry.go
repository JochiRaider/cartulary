package imports

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

func NewOwnerCreateRegistry(
	facades ...ownerfacade.ImportOwnerCreateFacade,
) (*ownerfacade.ImportOwnerCreateRegistry, error) {
	expected := make([]ownerfacade.ImportOwnerCreateBinding, 0)
	for _, target := range importtargetregistry.Targets {
		if target.TargetKind != ImportTargetKindViewSchema ||
			target.AvailabilityKind != "enabled" {
			continue
		}
		if target.TargetViewSchemaID == nil || target.FacadeID == nil {
			return nil, fmt.Errorf(
				"generated enabled import target %s has no owner-create binding",
				target.TargetID,
			)
		}
		expected = append(expected, ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: *target.TargetViewSchemaID,
			FacadeID:           *target.FacadeID,
		})
	}
	return ownerfacade.NewImportOwnerCreateRegistry(expected, facades...)
}
