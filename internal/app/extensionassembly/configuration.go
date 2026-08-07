// Package extensionassembly contains application composition edges that adapt
// immutable Extensions catalogs to narrower owner-local ports.
package extensionassembly

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

// ConfigurationPolicy is the single immutable projection through which the
// admitted Extensions catalog supplies claim registrations and inactive-value
// behavior to the generic configuration kernel.
type ConfigurationPolicy struct {
	registrations []config.ClaimRegistration
	profileIDs    map[string]struct{}
	inactive      extensions.InactiveConfigurationCatalog
}

// RequestedClaims is the immutable configuration request admitted before
// dependency resolution. It is intentionally distinct from both
// extensions.ClaimResolution and extensions.ResolvedClaimSet.
type RequestedClaims struct {
	profileIDs []string
}

// ClaimConfiguration is the Extensions assembly-owned wire value for a
// profile whose deployment namespace contains only its claim request.
type ClaimConfiguration struct {
	Claimed bool `toml:"claimed"`
}

// GeneratedConfigurationPolicy admits packaged Extensions artifacts before
// exposing their configuration projection.
func GeneratedConfigurationPolicy() (ConfigurationPolicy, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return ConfigurationPolicy{}, err
	}
	return configurationPolicyFromCoordinator(coordinator)
}

// configurationPolicyFromCoordinator derives an exact claim/inactive policy
// from one already admitted descriptor and configuration catalog.
func configurationPolicyFromCoordinator(coordinator *extensions.Coordinator) (ConfigurationPolicy, error) {
	if coordinator == nil {
		return ConfigurationPolicy{}, fmt.Errorf("extension coordinator is required")
	}
	return newConfigurationPolicy(coordinator.Descriptors(), coordinator.InactiveConfigurationPolicies())
}

// newConfigurationPolicy validates the owner seam independently of the kernel.
// Malformed and future-profile fixtures exercise it inside this package; only
// admitted generated catalogs cross the production composition boundary.
func newConfigurationPolicy(
	descriptors []extensions.Descriptor,
	inactivePolicies []extensions.InactiveConfigurationPolicy,
) (ConfigurationPolicy, error) {
	policy := ConfigurationPolicy{profileIDs: make(map[string]struct{}, len(descriptors))}
	claimPaths := make(map[string]string, len(descriptors))
	for _, descriptor := range descriptors {
		if !descriptor.Claimable {
			continue
		}
		if descriptor.ProfileID == "" || descriptor.ClaimConfigKey == "" {
			return ConfigurationPolicy{}, fmt.Errorf("claimable extension descriptor identity is incomplete")
		}
		if descriptor.ClaimConfigKey != descriptor.ProfileID+".claimed" {
			return ConfigurationPolicy{}, fmt.Errorf("extension profile %q does not own canonical claim path %q", descriptor.ProfileID, descriptor.ClaimConfigKey)
		}
		if _, duplicate := policy.profileIDs[descriptor.ProfileID]; duplicate {
			return ConfigurationPolicy{}, fmt.Errorf("duplicate extension profile %q", descriptor.ProfileID)
		}
		if prior := claimPaths[descriptor.ClaimConfigKey]; prior != "" {
			return ConfigurationPolicy{}, fmt.Errorf("extension profiles %q and %q share claim path %q", prior, descriptor.ProfileID, descriptor.ClaimConfigKey)
		}
		policy.profileIDs[descriptor.ProfileID] = struct{}{}
		claimPaths[descriptor.ClaimConfigKey] = descriptor.ProfileID
		policy.registrations = append(policy.registrations, config.ClaimRegistration{
			ID:   descriptor.ProfileID,
			Path: descriptor.ClaimConfigKey,
		})
	}
	sort.Slice(policy.registrations, func(i, j int) bool {
		return policy.registrations[i].ID < policy.registrations[j].ID
	})

	for _, inactivePolicy := range inactivePolicies {
		if _, known := policy.profileIDs[inactivePolicy.ProfileID]; !known {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q has no claimable profile", inactivePolicy.Key)
		}
		if claimPaths[inactivePolicy.ClaimKey] != inactivePolicy.ProfileID {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q has stale claim path %q", inactivePolicy.Key, inactivePolicy.ClaimKey)
		}
		if !strings.HasPrefix(inactivePolicy.Key, inactivePolicy.ProfileID+".") || inactivePolicy.Key == inactivePolicy.ClaimKey {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q is outside profile namespace %q", inactivePolicy.Key, inactivePolicy.ProfileID)
		}
	}

	inactive, err := extensions.NewInactiveConfigurationCatalog(inactivePolicies)
	if err != nil {
		return ConfigurationPolicy{}, err
	}
	policy.inactive = inactive
	return policy, nil
}

func (policy ConfigurationPolicy) ClaimRegistrations() []config.ClaimRegistration {
	return append([]config.ClaimRegistration(nil), policy.registrations...)
}

func (policy ConfigurationPolicy) Keys() []string {
	return policy.inactive.Keys()
}

func (policy ConfigurationPolicy) ClaimKey(key string) (string, bool) {
	return policy.inactive.ClaimKey(key)
}

func (policy ConfigurationPolicy) ParseOverlay(key string, raw string) (any, error) {
	return policy.inactive.ParseOverlay(key, raw)
}

func (policy ConfigurationPolicy) ValidateAndDiscard(values map[string]any) [][2]string {
	return policy.inactive.ValidateAndDiscard(values)
}

// MaterializeRequestedClaims converts the kernel's closed registration
// identities into the only request type accepted by application composition.
func (policy ConfigurationPolicy) MaterializeRequestedClaims(registrationIDs []string) (RequestedClaims, error) {
	requested := make([]string, 0, len(registrationIDs))
	seen := make(map[string]struct{}, len(registrationIDs))
	for _, profileID := range registrationIDs {
		if _, known := policy.profileIDs[profileID]; !known {
			return RequestedClaims{}, fmt.Errorf("extension claim registration %q is not admitted", profileID)
		}
		if _, duplicate := seen[profileID]; duplicate {
			return RequestedClaims{}, fmt.Errorf("extension claim registration %q is duplicated", profileID)
		}
		seen[profileID] = struct{}{}
		requested = append(requested, profileID)
	}
	sort.Strings(requested)
	return RequestedClaims{profileIDs: requested}, nil
}

// ProfileIDs returns a defensive canonical request projection for coordinator
// resolution. Resolution never accepts configuration paths or Boolean maps.
func (claims RequestedClaims) ProfileIDs() []string {
	return append([]string(nil), claims.profileIDs...)
}
