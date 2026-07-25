// Package extensionassembly contains application composition edges that adapt
// immutable Extensions catalogs to narrower owner-local ports.
package extensionassembly

import (
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

// ClaimConfigurationPaths returns the exact generated configuration projection
// needed to resolve the supplied descriptors. It deliberately carries no
// profile-specific configuration knowledge.
func ClaimConfigurationPaths(descriptors []extensions.Descriptor) ([]string, error) {
	paths := make([]string, 0, len(descriptors))
	profiles := make(map[string]struct{}, len(descriptors))
	seenPaths := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.ProfileID == "" || descriptor.ClaimConfigKey == "" {
			return nil, fmt.Errorf("extension claim descriptor identity is incomplete")
		}
		if _, duplicate := profiles[descriptor.ProfileID]; duplicate {
			return nil, fmt.Errorf("duplicate extension profile %q", descriptor.ProfileID)
		}
		if _, duplicate := seenPaths[descriptor.ClaimConfigKey]; duplicate {
			return nil, fmt.Errorf("duplicate extension claim configuration path %q", descriptor.ClaimConfigKey)
		}
		profiles[descriptor.ProfileID] = struct{}{}
		seenPaths[descriptor.ClaimConfigKey] = struct{}{}
		paths = append(paths, descriptor.ClaimConfigKey)
	}
	sort.Strings(paths)
	return paths, nil
}

// ResolveClaimRequest materializes the canonical profile request solely from
// generated descriptor claim keys and a narrow Boolean configuration projection.
func ResolveClaimRequest(descriptors []extensions.Descriptor, values map[string]bool) ([]string, error) {
	paths, err := ClaimConfigurationPaths(descriptors)
	if err != nil {
		return nil, err
	}
	if len(values) != len(paths) {
		return nil, fmt.Errorf("extension claim configuration projection has %d values; want %d", len(values), len(paths))
	}
	knownPaths := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		knownPaths[path] = struct{}{}
		if _, present := values[path]; !present {
			return nil, fmt.Errorf("extension claim configuration path %q is unresolved", path)
		}
	}
	for path := range values {
		if _, known := knownPaths[path]; !known {
			return nil, fmt.Errorf("extension claim configuration path %q is not generated", path)
		}
	}
	claimed := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if values[descriptor.ClaimConfigKey] {
			claimed = append(claimed, descriptor.ProfileID)
		}
	}
	sort.Strings(claimed)
	return claimed, nil
}

func GeneratedInactiveConfigurationPolicy() (config.InactivePolicy, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, err
	}
	return InactiveConfigurationPolicy(coordinator)
}

func InactiveConfigurationPolicy(coordinator *extensions.Coordinator) (config.InactivePolicy, error) {
	if coordinator == nil {
		return nil, fmt.Errorf("extension coordinator is required")
	}
	catalog, err := extensions.NewInactiveConfigurationCatalog(coordinator.InactiveConfigurationPolicies())
	if err != nil {
		return nil, err
	}
	return catalog, nil
}
