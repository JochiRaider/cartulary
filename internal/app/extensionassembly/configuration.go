// Package extensionassembly contains application composition edges that adapt
// immutable Extensions catalogs to narrower owner-local ports.
package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/config/extensioninactive"
)

func GeneratedInactiveConfigurationCatalog() (extensioninactive.Catalog, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return extensioninactive.Catalog{}, err
	}
	return InactiveConfigurationCatalog(coordinator)
}

func InactiveConfigurationCatalog(coordinator *extensions.Coordinator) (extensioninactive.Catalog, error) {
	if coordinator == nil {
		return extensioninactive.Catalog{}, fmt.Errorf("extension coordinator is required")
	}
	projection := coordinator.InactiveConfigurationPolicies()
	policies := make([]extensioninactive.Policy, len(projection))
	for index, policy := range projection {
		policies[index] = extensioninactive.Policy{
			ProfileID: policy.ProfileID,
			ClaimKey:  policy.ClaimKey,
			Key:       policy.Key,
			Kind:      extensioninactive.PolicyKind(policy.Kind),
			Schema:    policy.Schema,
		}
	}
	return extensioninactive.NewCatalog(policies)
}
