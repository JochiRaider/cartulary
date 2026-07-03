package revisions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeImportedAttributionResolver struct{}

func (fakeImportedAttributionResolver) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestAttributionResolverRegistryRequiresIncidentPortabilityResolver(t *testing.T) {
	registry := NewAttributionResolverRegistry()
	err := registry.ValidateAttributionResolvers([]ExtensionClaim{{ProfileID: "incident_portability", Claimed: true}})
	if !errors.Is(err, ErrMissingAttributionResolver) {
		t.Fatalf("missing incident portability resolver error got %v", err)
	}
	if err := registry.RegisterImportedAttributionResolver("incident_portability", fakeImportedAttributionResolver{}); err != nil {
		t.Fatalf("register resolver: %v", err)
	}
	if err := registry.ValidateAttributionResolvers([]ExtensionClaim{{ProfileID: "incident_portability", Claimed: true}}); err != nil {
		t.Fatalf("validate registered resolver: %v", err)
	}
}

func TestAttributionResolverRegistryRejectsDuplicateProfile(t *testing.T) {
	registry := NewAttributionResolverRegistry()
	if err := registry.RegisterImportedAttributionResolver("incident_portability", fakeImportedAttributionResolver{}); err != nil {
		t.Fatalf("register resolver: %v", err)
	}
	err := registry.RegisterImportedAttributionResolver("incident_portability", fakeImportedAttributionResolver{})
	if !errors.Is(err, ErrDuplicateAttributionResolver) {
		t.Fatalf("duplicate resolver error got %v", err)
	}
}
