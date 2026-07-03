package revisions

import (
	"errors"
	"fmt"
)

var (
	ErrDuplicateAttributionResolver = errors.New("revisions: duplicate attribution resolver")
	ErrMissingAttributionResolver   = errors.New("revisions: missing attribution resolver")
)

type ExtensionClaim struct {
	ProfileID string
	Claimed   bool
}

type AttributionResolverRegistry struct {
	resolvers map[string]ImportedAttributionResolver
}

var requiredAttributionResolverProfiles = map[string]struct{}{
	"incident_portability": {},
}

func NewAttributionResolverRegistry() *AttributionResolverRegistry {
	return &AttributionResolverRegistry{resolvers: map[string]ImportedAttributionResolver{}}
}

func (r *AttributionResolverRegistry) RegisterImportedAttributionResolver(profileID string, resolver ImportedAttributionResolver) error {
	if r == nil {
		return errors.New("revisions: attribution resolver registry is nil")
	}
	if profileID == "" || resolver == nil {
		return fmt.Errorf("%w: %s", ErrMissingAttributionResolver, profileID)
	}
	if _, exists := r.resolvers[profileID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateAttributionResolver, profileID)
	}
	r.resolvers[profileID] = resolver
	return nil
}

func (r *AttributionResolverRegistry) ValidateAttributionResolvers(activeExtensionClaims []ExtensionClaim) error {
	if r == nil {
		return errors.New("revisions: attribution resolver registry is nil")
	}
	for _, claim := range activeExtensionClaims {
		if !claim.Claimed || !requiresAttributionResolver(claim.ProfileID) {
			continue
		}
		if _, ok := r.resolvers[claim.ProfileID]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingAttributionResolver, claim.ProfileID)
		}
	}
	return nil
}

func (r *AttributionResolverRegistry) ImportedAttributionResolver(profileID string) ImportedAttributionResolver {
	if r == nil {
		return noopImportedAttributionResolver{}
	}
	resolver, ok := r.resolvers[profileID]
	if !ok || resolver == nil {
		return noopImportedAttributionResolver{}
	}
	return resolver
}

func requiresAttributionResolver(profileID string) bool {
	_, ok := requiredAttributionResolverProfiles[profileID]
	return ok
}
