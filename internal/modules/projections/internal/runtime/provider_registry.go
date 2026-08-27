package runtime

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

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
	descriptor                      providercontract.ProviderDescriptor
	queryStrategy                   queryStrategy
	queryPlans                      []queryengine.Surface
	evidenceAssociationEffectFields []string
	loadEvidenceAssociationStateTx  func(context.Context, *Store, pgx.Tx, uuid.UUID) (map[string]any, error)
	refreshRowTx                    func(context.Context, *Store, pgx.Tx, uuid.UUID) error
	rebuildIncidentTx               func(context.Context, *Store, pgx.Tx, uuid.UUID) error
}

type providerRegistry struct {
	providers     []*Provider
	byViewSchema  map[string]*Provider
	querySurfaces map[string]queryengine.Surface
	bySourceType  map[string][]*Provider
	rebuildOrder  []*Provider
}

type Catalog struct {
	descriptors providercontract.DescriptorSet
	registry    *providerRegistry
}

func newCatalog(descriptors providercontract.DescriptorSet, providers []Provider) (*Catalog, error) {
	registry, err := newProviderRegistry(descriptors, providers)
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

func newProviderRegistry(
	descriptors providercontract.DescriptorSet,
	providers []Provider,
) (*providerRegistry, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("projection provider registry is empty")
	}
	registry := &providerRegistry{
		providers:     make([]*Provider, 0, len(providers)),
		byViewSchema:  map[string]*Provider{},
		querySurfaces: map[string]queryengine.Surface{},
		bySourceType:  map[string][]*Provider{},
	}
	byProviderID := map[string]*Provider{}
	for index := range providers {
		provider := providers[index]
		declarativeDescriptor, exists := descriptors.Lookup(provider.descriptor.ProviderID)
		if !exists {
			return nil, fmt.Errorf("projection provider %q has no declarative descriptor", provider.descriptor.ProviderID)
		}
		if !reflect.DeepEqual(declarativeDescriptor, provider.descriptor) {
			return nil, fmt.Errorf("projection provider %q executable binding descriptor does not match declarative descriptor", provider.descriptor.ProviderID)
		}
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
		if len(provider.evidenceAssociationEffectFields) > 0 {
			if len(provider.descriptor.ViewSchemaIDs) != 1 {
				return nil, fmt.Errorf("projection provider %q Evidence association effects require exactly one view", provider.descriptor.ProviderID)
			}
			if provider.loadEvidenceAssociationStateTx == nil {
				return nil, fmt.Errorf("projection provider %q Evidence association effects require a state loader", provider.descriptor.ProviderID)
			}
			for _, recordType := range provider.descriptor.SourceRecordTypes {
				registry.bySourceType[recordType] = append(registry.bySourceType[recordType], providerPointer)
			}
		}
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
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
			if existing, exists := registry.querySurfaces[surface.ViewSchemaID]; exists {
				return nil, fmt.Errorf("duplicate projection query surface ownership for %q: %q and %q", surface.ViewSchemaID, existing.ViewSchemaID, provider.descriptor.ProviderID)
			}
			registry.querySurfaces[surface.ViewSchemaID] = surface
		}
	}
	for _, descriptor := range descriptors.RebuildOrder() {
		provider := byProviderID[descriptor.ProviderID]
		if provider == nil {
			return nil, fmt.Errorf("projection provider %q has no executable binding", descriptor.ProviderID)
		}
		registry.rebuildOrder = append(registry.rebuildOrder, provider)
	}
	for recordType := range registry.bySourceType {
		slices.SortFunc(registry.bySourceType[recordType], func(left *Provider, right *Provider) int {
			return strings.Compare(left.descriptor.ViewSchemaIDs[0], right.descriptor.ViewSchemaIDs[0])
		})
	}
	return registry, nil
}

func (r *providerRegistry) providerForView(viewSchemaID string) (*Provider, bool) {
	if r == nil {
		return nil, false
	}
	provider := r.byViewSchema[viewSchemaID]
	return provider, provider != nil
}

func (r *providerRegistry) querySurfaceForView(viewSchemaID string) (queryengine.Surface, bool) {
	if r == nil {
		return queryengine.Surface{}, false
	}
	surface, ok := r.querySurfaces[viewSchemaID]
	return surface, ok
}

func (r *providerRegistry) evidenceAssociationProviders(recordType string) []*Provider {
	if r == nil {
		return nil
	}
	return append([]*Provider(nil), r.bySourceType[recordType]...)
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
