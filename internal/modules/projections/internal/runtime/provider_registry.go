package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

type queryStrategy uint8

const (
	queryStrategyNone queryStrategy = iota
	queryStrategyCompiledPlan
	queryStrategySourceOwnerHydration
)

type Provider struct {
	descriptor        providercontract.ProviderDescriptor
	queryStrategy     queryStrategy
	queryPlans        []queryengine.Surface
	refreshRowTx      func(context.Context, *Store, pgx.Tx, uuid.UUID) error
	rebuildIncidentTx func(context.Context, *Store, pgx.Tx, uuid.UUID) error
}

type providerRegistry struct {
	providers     []*Provider
	byViewSchema  map[string]*Provider
	querySurfaces map[string]genericSurface
	rebuildOrder  []*Provider
}

type Catalog struct {
	descriptors providercontract.DescriptorSet
	registry    *providerRegistry
}

func newCatalog(descriptors providercontract.DescriptorSet, providers []Provider) (*Catalog, error) {
	registry, err := newProviderRegistry(providers)
	if err != nil {
		return nil, err
	}
	return &Catalog{descriptors: descriptors, registry: registry}, nil
}

func (c *Catalog) DescriptorSet() providercontract.DescriptorSet {
	if c == nil {
		return providercontract.DescriptorSet{}
	}
	return c.descriptors
}

func (s *Store) providerRegistry() (*providerRegistry, error) {
	if s == nil || s.registry == nil {
		return nil, fmt.Errorf("projection catalog is required")
	}
	return s.registry, nil
}

func newProviderRegistry(providers []Provider) (*providerRegistry, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("projection provider registry is empty")
	}
	registry := &providerRegistry{
		providers:     make([]*Provider, 0, len(providers)),
		byViewSchema:  map[string]*Provider{},
		querySurfaces: map[string]genericSurface{},
	}
	byProviderID := map[string]*Provider{}
	tableOwners := map[string]string{}
	for index := range providers {
		provider := providers[index]
		if err := validateProvider(provider); err != nil {
			return nil, err
		}
		if _, exists := byProviderID[provider.descriptor.ProviderID]; exists {
			return nil, fmt.Errorf("duplicate projection provider_id %q", provider.descriptor.ProviderID)
		}
		providerCopy := provider
		providerPointer := &providerCopy
		byProviderID[provider.descriptor.ProviderID] = providerPointer
		registry.providers = append(registry.providers, providerPointer)
		for _, tableID := range provider.descriptor.ProjectionTableIDs {
			if existing := tableOwners[tableID]; existing != "" {
				return nil, fmt.Errorf(
					"duplicate projection table ownership for %q: %q and %q",
					tableID,
					existing,
					provider.descriptor.ProviderID,
				)
			}
			tableOwners[tableID] = provider.descriptor.ProviderID
		}
		seenViews := map[string]struct{}{}
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			if _, exists := seenViews[viewSchemaID]; exists {
				return nil, fmt.Errorf("projection provider %q declares duplicate view_schema_id %q", provider.descriptor.ProviderID, viewSchemaID)
			}
			seenViews[viewSchemaID] = struct{}{}
			if existing := registry.byViewSchema[viewSchemaID]; existing != nil {
				return nil, fmt.Errorf("duplicate projection view ownership for %q: %q and %q", viewSchemaID, existing.descriptor.ProviderID, provider.descriptor.ProviderID)
			}
			registry.byViewSchema[viewSchemaID] = providerPointer
		}
		querySurfaces, err := providerPlans(provider)
		if err != nil {
			return nil, err
		}
		for _, surface := range querySurfaces {
			if existing, exists := registry.querySurfaces[surface.viewSchemaID]; exists {
				return nil, fmt.Errorf("duplicate projection query surface ownership for %q: %q and %q", surface.viewSchemaID, existing.viewSchemaID, provider.descriptor.ProviderID)
			}
			registry.querySurfaces[surface.viewSchemaID] = surface
		}
	}
	rebuildOrder, err := topologicalProviderOrder(registry.providers, byProviderID)
	if err != nil {
		return nil, err
	}
	registry.rebuildOrder = rebuildOrder
	return registry, nil
}

func (r *providerRegistry) providerForView(viewSchemaID string) (*Provider, bool) {
	if r == nil {
		return nil, false
	}
	provider := r.byViewSchema[viewSchemaID]
	return provider, provider != nil
}

func (r *providerRegistry) querySurfaceForView(viewSchemaID string) (genericSurface, bool) {
	if r == nil {
		return genericSurface{}, false
	}
	surface, ok := r.querySurfaces[viewSchemaID]
	return surface, ok
}

func (s *Store) refreshProjectionRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	registry, err := s.providerRegistry()
	if err != nil {
		return err
	}
	provider, ok := registry.providerForView(viewSchemaID)
	if !ok || !provider.descriptor.Capabilities.RefreshRow || provider.refreshRowTx == nil {
		return fmt.Errorf("projection refresh surface %q not mapped", viewSchemaID)
	}
	return provider.refreshRowTx(ctx, s, tx, recordID)
}

func (s *Store) RefreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, viewSchemaID, recordID)
}

func (s *Store) rebuildProjectionIncidentTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	registry, err := s.providerRegistry()
	if err != nil {
		return err
	}
	provider, ok := registry.providerForView(viewSchemaID)
	if !ok || !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection rebuild surface %q not mapped", viewSchemaID)
	}
	return provider.rebuildIncidentTx(ctx, s, tx, incidentID)
}

func (s *Store) RebuildIncidentViewsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, viewSchemaIDs []string) error {
	registry, err := s.providerRegistry()
	if err != nil {
		return err
	}
	selectedViews := make(map[string]struct{}, len(viewSchemaIDs))
	selectedProviders := make(map[string]struct{}, len(viewSchemaIDs))
	for _, viewSchemaID := range viewSchemaIDs {
		provider, ok := registry.providerForView(viewSchemaID)
		if !ok || !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
			return fmt.Errorf("projection rebuild surface %q not mapped", viewSchemaID)
		}
		selectedViews[viewSchemaID] = struct{}{}
		selectedProviders[provider.descriptor.ProviderID] = struct{}{}
	}
	for _, provider := range registry.rebuildOrder {
		if _, ok := selectedProviders[provider.descriptor.ProviderID]; !ok {
			continue
		}
		if err := provider.rebuildIncidentTx(ctx, s, tx, incidentID); err != nil {
			return err
		}
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			delete(selectedViews, viewSchemaID)
		}
	}
	if len(selectedViews) > 0 {
		for viewSchemaID := range selectedViews {
			return fmt.Errorf("projection rebuild surface %q not reached by registry order", viewSchemaID)
		}
	}
	return nil
}

func (s *Store) RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	registry, err := s.providerRegistry()
	if err != nil {
		return err
	}
	for _, provider := range registry.rebuildOrder {
		if !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
			continue
		}
		if err := provider.rebuildIncidentTx(ctx, s, tx, incidentID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.RebuildIncidentTx(ctx, tx, incidentID)
}
