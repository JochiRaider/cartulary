package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const projectionProviderDescriptorSchemaVersion = providercontract.DescriptorSchemaVersion

type ProviderStatus string

const (
	ProviderStatusActive       ProviderStatus = "active"
	ProviderStatusDeprecated   ProviderStatus = "deprecated"
	ProviderStatusExperimental ProviderStatus = "experimental"
)

type RestoreRebuildParticipation string

const (
	RestoreRebuildRequired         RestoreRebuildParticipation = "required"
	RestoreRebuildNonparticipating RestoreRebuildParticipation = "nonparticipating"
	RestoreRebuildUnsupported      RestoreRebuildParticipation = "unsupported"
)

type ProviderCapabilities struct {
	Query           bool
	RefreshRow      bool
	RestoreRebuild  bool
	IncidentRebuild bool
}

type ProviderDescriptor struct {
	SchemaVersion             string
	Status                    ProviderStatus
	ProviderKey               string
	SourceOwnerKey            string
	ViewSchemaIDs             []string
	SourceRecordTypes         []string
	SourceAuthorityModules    []string
	ProjectionTableFamilies   []string
	ProjectionStorageOwnerKey string
	Capabilities              ProviderCapabilities
	QuerySurfaces             []providercontract.QuerySurface
	RestoreRebuild            RestoreRebuildParticipation
	FacadePackages            []string
	RebuildAfter              []string
	CharacterizationRefs      []string
}

type projectionProvider struct {
	descriptor        ProviderDescriptor
	refreshRowTx      func(context.Context, *Store, pgx.Tx, uuid.UUID) error
	rebuildIncidentTx func(context.Context, *Store, pgx.Tx, uuid.UUID) error
}

type providerRegistry struct {
	providers     []*projectionProvider
	byViewSchema  map[string]*projectionProvider
	querySurfaces map[string]genericSurface
	rebuildOrder  []*projectionProvider
}

var requiredProjectionViewSchemaIDs = map[string]struct{}{
	timelineViewSchemaID:     {},
	hostsViewSchemaID:        {},
	identitiesViewSchemaID:   {},
	indicatorsViewSchemaID:   {},
	assessmentsViewSchemaID:  {},
	evidenceViewSchemaID:     {},
	notesViewSchemaID:        {},
	partiesViewSchemaID:      {},
	taskRequestsViewSchemaID: {},
	decisionsViewSchemaID:    {},
	commLogViewSchemaID:      {},
	handoffViewSchemaID:      {},
	statusReviewViewSchemaID: {},
	lessonViewSchemaID:       {},
}

var projectionTableSchemaOwners = map[string]string{
	"timeline_grid_projection":     "projections",
	"host_grid_projection":         "projections",
	"identity_grid_projection":     "projections",
	"indicator_grid_projection":    "projections",
	"assessment_grid_projection":   "projections",
	"artifact_grid_projection":     "projections",
	"evidence_grid_projection":     "projections",
	"party_grid_projection":        "projections",
	"task_request_grid_projection": "projections",
	"decision_grid_projection":     "projections",
}

var projectionSourceAuthorityModules = map[string]struct{}{
	"assessments":         {},
	"artifacts":           {},
	"auth":                {},
	"collaboration":       {},
	"database_migrations": {},
	"deployment_admin":    {},
	"entities":            {},
	"evidence":            {},
	"graphprojection":     {},
	"harness_support":     {},
	"imports":             {},
	"incidentbundles":     {},
	"incidents":           {},
	"indicators":          {},
	"jobapi":              {},
	"links":               {},
	"networkflow":         {},
	"parties":             {},
	"platform_jobs":       {},
	"projections":         {},
	"recovery":            {},
	"reference_data":      {},
	"reportcomposition":   {},
	"reporting":           {},
	"revisions":           {},
	"savedviews":          {},
	"tasksdecisions":      {},
	"timeline":            {},
	"viewschemas":         {},
	"workbook":            {},
}

func defaultProviderRegistry() *providerRegistry {
	registry, err := newProviderRegistry(builtInProjectionProviders())
	if err != nil {
		panic(fmt.Sprintf("invalid built-in projection provider registry: %v", err))
	}
	return registry
}

func (s *Store) providerRegistry() *providerRegistry {
	if s == nil || s.registry == nil {
		return defaultProviderRegistry()
	}
	return s.registry
}

func newProviderRegistry(providers []projectionProvider) (*providerRegistry, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("projection provider registry is empty")
	}
	registry := &providerRegistry{
		providers:     make([]*projectionProvider, 0, len(providers)),
		byViewSchema:  map[string]*projectionProvider{},
		querySurfaces: map[string]genericSurface{},
	}
	byProviderKey := map[string]*projectionProvider{}
	for index := range providers {
		provider := providers[index]
		if err := validateProvider(provider); err != nil {
			return nil, err
		}
		if _, exists := byProviderKey[provider.descriptor.ProviderKey]; exists {
			return nil, fmt.Errorf("duplicate projection provider key %q", provider.descriptor.ProviderKey)
		}
		providerCopy := provider
		providerPointer := &providerCopy
		byProviderKey[provider.descriptor.ProviderKey] = providerPointer
		registry.providers = append(registry.providers, providerPointer)
		seenViews := map[string]struct{}{}
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			if _, exists := seenViews[viewSchemaID]; exists {
				return nil, fmt.Errorf("projection provider %q declares duplicate view_schema_id %q", provider.descriptor.ProviderKey, viewSchemaID)
			}
			seenViews[viewSchemaID] = struct{}{}
			if existing := registry.byViewSchema[viewSchemaID]; existing != nil {
				return nil, fmt.Errorf("duplicate projection provider ownership for %q: %q and %q", viewSchemaID, existing.descriptor.ProviderKey, provider.descriptor.ProviderKey)
			}
			registry.byViewSchema[viewSchemaID] = providerPointer
		}
		querySurfaces, err := providerQuerySurfaces(provider)
		if err != nil {
			return nil, err
		}
		for _, surface := range querySurfaces {
			if existing, exists := registry.querySurfaces[surface.viewSchemaID]; exists {
				return nil, fmt.Errorf("duplicate projection query surface ownership for %q: %q and %q", surface.viewSchemaID, existing.viewSchemaID, provider.descriptor.ProviderKey)
			}
			registry.querySurfaces[surface.viewSchemaID] = surface
		}
	}
	for viewSchemaID := range requiredProjectionViewSchemaIDs {
		if registry.byViewSchema[viewSchemaID] == nil {
			return nil, fmt.Errorf("required projection surface %q has no provider", viewSchemaID)
		}
	}
	rebuildOrder, err := topologicalProviderOrder(registry.providers, byProviderKey)
	if err != nil {
		return nil, err
	}
	registry.rebuildOrder = rebuildOrder
	return registry, nil
}

func (r *providerRegistry) providerForView(viewSchemaID string) (*projectionProvider, bool) {
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
	provider, ok := s.providerRegistry().providerForView(viewSchemaID)
	if !ok || !provider.descriptor.Capabilities.RefreshRow || provider.refreshRowTx == nil {
		return fmt.Errorf("projection refresh surface %q not mapped", viewSchemaID)
	}
	return provider.refreshRowTx(ctx, s, tx, recordID)
}

func (s *Store) RefreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, viewSchemaID, recordID)
}

func (s *Store) rebuildProjectionIncidentTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	provider, ok := s.providerRegistry().providerForView(viewSchemaID)
	if !ok || !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection rebuild surface %q not mapped", viewSchemaID)
	}
	return provider.rebuildIncidentTx(ctx, s, tx, incidentID)
}

func (s *Store) RebuildIncidentViewTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, viewSchemaID, incidentID)
}

func (s *Store) RebuildIncidentViewsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, viewSchemaIDs []string) error {
	registry := s.providerRegistry()
	selectedViews := make(map[string]struct{}, len(viewSchemaIDs))
	selectedProviders := make(map[string]struct{}, len(viewSchemaIDs))
	for _, viewSchemaID := range viewSchemaIDs {
		provider, ok := registry.providerForView(viewSchemaID)
		if !ok || !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
			return fmt.Errorf("projection rebuild surface %q not mapped", viewSchemaID)
		}
		selectedViews[viewSchemaID] = struct{}{}
		selectedProviders[provider.descriptor.ProviderKey] = struct{}{}
	}
	for _, provider := range registry.rebuildOrder {
		if _, ok := selectedProviders[provider.descriptor.ProviderKey]; !ok {
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
	for _, provider := range s.providerRegistry().rebuildOrder {
		if !provider.descriptor.Capabilities.IncidentRebuild || provider.rebuildIncidentTx == nil {
			continue
		}
		if err := provider.rebuildIncidentTx(ctx, s, tx, incidentID); err != nil {
			return err
		}
	}
	return nil
}
