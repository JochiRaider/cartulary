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
	valueSchemas  []config.NamespaceSchemaBindings
}

// RequestedClaims is the immutable configuration request admitted before
// dependency resolution. It is intentionally distinct from both
// extensions.ClaimResolution and extensions.ResolvedClaimSet.
type RequestedClaims struct {
	profileIDs []string
}

// GeneratedConfigurationPolicy admits packaged Extensions artifacts before
// exposing their configuration projection.
func GeneratedConfigurationPolicy() (ConfigurationPolicy, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return ConfigurationPolicy{}, err
	}
	return ConfigurationPolicyFromCoordinator(coordinator)
}

// ConfigurationPolicyFromCoordinator derives an exact claim/inactive policy
// from one already admitted descriptor and configuration catalog.
func ConfigurationPolicyFromCoordinator(coordinator *extensions.Coordinator) (ConfigurationPolicy, error) {
	if coordinator == nil {
		return ConfigurationPolicy{}, fmt.Errorf("extension coordinator is required")
	}
	policy, err := newConfigurationPolicy(coordinator.Descriptors(), coordinator.InactiveConfigurationPolicies())
	if err != nil {
		return ConfigurationPolicy{}, err
	}
	policy.valueSchemas = make([]config.NamespaceSchemaBindings, 0)
	for _, contract := range coordinator.ConfigurationSchemaContracts() {
		claimable := false
		for i := range policy.registrations {
			if policy.registrations[i].ID == contract.ProfileID {
				policy.registrations[i].Namespace = contract.Namespace
				claimable = policy.registrations[i].Claimable
				break
			}
		}
		if !claimable {
			continue
		}
		group := config.NamespaceSchemaBindings{Namespace: contract.Namespace, Values: []config.ValueSchemaBinding{}}
		for _, value := range contract.Values {
			group.Values = append(group.Values, config.ValueSchemaBinding{Path: value.Key, SchemaID: value.SchemaID})
		}
		policy.valueSchemas = append(policy.valueSchemas, group)
	}
	return policy, nil
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
		if descriptor.ProfileID == "" || descriptor.ClaimConfigKey == "" {
			return ConfigurationPolicy{}, fmt.Errorf("recognized extension descriptor identity is incomplete")
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
			ID:        descriptor.ProfileID,
			Path:      descriptor.ClaimConfigKey,
			Claimable: descriptor.Claimable,
		})
	}
	sort.Slice(policy.registrations, func(i, j int) bool {
		return policy.registrations[i].ID < policy.registrations[j].ID
	})

	namespaces := map[string]string{}
	profileNamespaces := map[string]string{}
	for _, inactivePolicy := range inactivePolicies {
		if _, known := policy.profileIDs[inactivePolicy.ProfileID]; !known {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q has no recognized profile", inactivePolicy.Key)
		}
		if claimPaths[inactivePolicy.ClaimKey] != inactivePolicy.ProfileID {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q has stale claim path %q", inactivePolicy.Key, inactivePolicy.ClaimKey)
		}
		if inactivePolicy.Namespace == "" || !strings.HasPrefix(inactivePolicy.Key, inactivePolicy.Namespace+".") || inactivePolicy.Key == inactivePolicy.ClaimKey {
			return ConfigurationPolicy{}, fmt.Errorf("inactive configuration policy %q is outside profile namespace %q", inactivePolicy.Key, inactivePolicy.ProfileID)
		}
		if prior := namespaces[inactivePolicy.Namespace]; prior != "" && prior != inactivePolicy.ProfileID {
			return ConfigurationPolicy{}, fmt.Errorf("configuration namespace has multiple profile owners")
		}
		if prior := profileNamespaces[inactivePolicy.ProfileID]; prior != "" && prior != inactivePolicy.Namespace {
			return ConfigurationPolicy{}, fmt.Errorf("profile configuration has multiple namespaces")
		}
		namespaces[inactivePolicy.Namespace] = inactivePolicy.ProfileID
		profileNamespaces[inactivePolicy.ProfileID] = inactivePolicy.Namespace
	}

	inactive, err := extensions.NewInactiveConfigurationCatalog(inactivePolicies)
	if err != nil {
		return ConfigurationPolicy{}, err
	}
	for i := range policy.registrations {
		policy.registrations[i].Namespace = profileNamespaces[policy.registrations[i].ID]
	}
	policy.inactive = inactive
	return policy, nil
}

// ValidateAdmission rejects the zero value before configuration inputs are read.
func (policy ConfigurationPolicy) ValidateAdmission() error {
	if policy.profileIDs == nil {
		return fmt.Errorf("admitted Extensions configuration policy is required")
	}
	return nil
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

// ValueSchemaBindings returns defensive references for the actual configuration
// catalog to admit before input loading. It executes no profile code.
func (policy ConfigurationPolicy) ValueSchemaBindings() []config.NamespaceSchemaBindings {
	result := make([]config.NamespaceSchemaBindings, len(policy.valueSchemas))
	for i, group := range policy.valueSchemas {
		result[i] = group
		result[i].Values = append([]config.ValueSchemaBinding(nil), group.Values...)
	}
	return result
}
