package projections

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
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

func validateProvider(provider projectionProvider) error {
	descriptor := provider.descriptor
	if descriptor.SchemaVersion != projectionProviderDescriptorSchemaVersion {
		return fmt.Errorf("projection provider %q declares unsupported schema_version %q", descriptor.ProviderKey, descriptor.SchemaVersion)
	}
	switch descriptor.Status {
	case ProviderStatusActive, ProviderStatusDeprecated, ProviderStatusExperimental:
	default:
		return fmt.Errorf("projection provider %q declares unsupported status %q", descriptor.ProviderKey, descriptor.Status)
	}
	if descriptor.ProviderKey == "" {
		return fmt.Errorf("projection provider has empty provider_key")
	}
	if descriptor.SourceOwnerKey == "" {
		return fmt.Errorf("projection provider %q has empty source_owner_key", descriptor.ProviderKey)
	}
	if len(descriptor.ViewSchemaIDs) == 0 {
		return fmt.Errorf("projection provider %q declares no view_schema_ids", descriptor.ProviderKey)
	}
	if len(descriptor.SourceRecordTypes) == 0 {
		return fmt.Errorf("projection provider %q declares no source_record_types", descriptor.ProviderKey)
	}
	if err := validateUniqueStrings(descriptor.ProviderKey, "source_record_types", descriptor.SourceRecordTypes); err != nil {
		return err
	}
	if len(descriptor.SourceAuthorityModules) == 0 {
		return fmt.Errorf("projection provider %q declares no source_authority_modules", descriptor.ProviderKey)
	}
	if err := validateSourceAuthorityModules(descriptor); err != nil {
		return err
	}
	if len(descriptor.ProjectionTableFamilies) == 0 {
		return fmt.Errorf("projection provider %q declares no projection_table_families", descriptor.ProviderKey)
	}
	if descriptor.ProjectionStorageOwnerKey == "" {
		return fmt.Errorf("projection provider %q has empty projection_storage_owner_key", descriptor.ProviderKey)
	}
	if descriptor.Capabilities.RefreshRow && provider.refreshRowTx == nil {
		return fmt.Errorf("projection provider %q declares refresh support without implementation", descriptor.ProviderKey)
	}
	if !descriptor.Capabilities.RefreshRow && provider.refreshRowTx != nil {
		return fmt.Errorf("projection provider %q has refresh implementation without capability", descriptor.ProviderKey)
	}
	if descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection provider %q declares incident rebuild support without implementation", descriptor.ProviderKey)
	}
	if !descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx != nil {
		return fmt.Errorf("projection provider %q has incident rebuild implementation without capability", descriptor.ProviderKey)
	}
	if descriptor.Capabilities.RestoreRebuild && !descriptor.Capabilities.IncidentRebuild {
		return fmt.Errorf("projection provider %q declares restore rebuild without incident rebuild capability", descriptor.ProviderKey)
	}
	if descriptor.Capabilities.Query != (len(descriptor.QuerySurfaces) > 0) {
		return fmt.Errorf("projection provider %q query capability does not match registered query surfaces", descriptor.ProviderKey)
	}
	declaredViews := map[string]struct{}{}
	for _, viewSchemaID := range descriptor.ViewSchemaIDs {
		declaredViews[viewSchemaID] = struct{}{}
	}
	seenQuerySurfaces := map[string]struct{}{}
	querySurfaces, err := providerQuerySurfaces(provider)
	if err != nil {
		return err
	}
	for _, surface := range querySurfaces {
		if surface.viewSchemaID == "" {
			return fmt.Errorf("projection provider %q declares query surface with empty view_schema_id", descriptor.ProviderKey)
		}
		if _, ok := declaredViews[surface.viewSchemaID]; !ok {
			return fmt.Errorf("projection provider %q query surface %q is not one of its view_schema_ids", descriptor.ProviderKey, surface.viewSchemaID)
		}
		if _, exists := seenQuerySurfaces[surface.viewSchemaID]; exists {
			return fmt.Errorf("projection provider %q declares duplicate query surface %q", descriptor.ProviderKey, surface.viewSchemaID)
		}
		seenQuerySurfaces[surface.viewSchemaID] = struct{}{}
	}
	switch descriptor.RestoreRebuild {
	case RestoreRebuildRequired:
		if !descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares required restore rebuild without capability", descriptor.ProviderKey)
		}
		if provider.rebuildIncidentTx == nil {
			return fmt.Errorf("projection provider %q declares required restore rebuild without implementation", descriptor.ProviderKey)
		}
	case RestoreRebuildNonparticipating:
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares nonparticipating restore rebuild with capability", descriptor.ProviderKey)
		}
	case RestoreRebuildUnsupported:
		if descriptor.Status == ProviderStatusActive {
			return fmt.Errorf("projection provider %q is active but declares unsupported restore rebuild", descriptor.ProviderKey)
		}
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares unsupported restore rebuild with capability", descriptor.ProviderKey)
		}
	default:
		return fmt.Errorf("projection provider %q declares unsupported restore_rebuild %q", descriptor.ProviderKey, descriptor.RestoreRebuild)
	}
	if len(descriptor.FacadePackages) == 0 {
		return fmt.Errorf("projection provider %q declares no facade_packages", descriptor.ProviderKey)
	}
	seenFacadePackages := map[string]struct{}{}
	for _, packagePath := range descriptor.FacadePackages {
		if err := validateFacadePackagePath(packagePath); err != nil {
			return fmt.Errorf("projection provider %q facade package %q: %w", descriptor.ProviderKey, packagePath, err)
		}
		if _, exists := seenFacadePackages[packagePath]; exists {
			return fmt.Errorf("projection provider %q declares duplicate facade package %q", descriptor.ProviderKey, packagePath)
		}
		seenFacadePackages[packagePath] = struct{}{}
	}
	for _, family := range descriptor.ProjectionTableFamilies {
		owner, ok := projectionTableSchemaOwners[family]
		if !ok {
			return fmt.Errorf("projection provider %q declares unknown projection table family %q", descriptor.ProviderKey, family)
		}
		if owner != descriptor.ProjectionStorageOwnerKey {
			return fmt.Errorf("projection provider %q projection_storage_owner_key=%q does not match %s owner %q", descriptor.ProviderKey, descriptor.ProjectionStorageOwnerKey, family, owner)
		}
	}
	return nil
}

func validateUniqueStrings(providerKey string, field string, values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("projection provider %q declares empty %s value", providerKey, field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("projection provider %q declares duplicate %s value %q", providerKey, field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateSourceAuthorityModules(descriptor ProviderDescriptor) error {
	if err := validateUniqueStrings(descriptor.ProviderKey, "source_authority_modules", descriptor.SourceAuthorityModules); err != nil {
		return err
	}
	includesSourceOwner := false
	for _, module := range descriptor.SourceAuthorityModules {
		if _, ok := projectionSourceAuthorityModules[module]; !ok {
			return fmt.Errorf("projection provider %q declares unknown source_authority_module %q", descriptor.ProviderKey, module)
		}
		if module == descriptor.SourceOwnerKey {
			includesSourceOwner = true
		}
	}
	if !includesSourceOwner {
		return fmt.Errorf("projection provider %q source_authority_modules omit source_owner_key %q", descriptor.ProviderKey, descriptor.SourceOwnerKey)
	}
	return nil
}

func providerQuerySurfaces(provider projectionProvider) ([]genericSurface, error) {
	surfaces := make([]genericSurface, 0, len(provider.descriptor.QuerySurfaces))
	for _, surface := range provider.descriptor.QuerySurfaces {
		converted, err := genericSurfaceFromContract(surface)
		if err != nil {
			return nil, fmt.Errorf("projection provider %q query surface: %w", provider.descriptor.ProviderKey, err)
		}
		surfaces = append(surfaces, converted)
	}
	return surfaces, nil
}

func validateFacadePackagePath(packagePath string) error {
	if packagePath == "" {
		return fmt.Errorf("empty package path")
	}
	if strings.Contains(packagePath, "\\") {
		return fmt.Errorf("must use slash-separated repo paths")
	}
	if strings.HasPrefix(packagePath, "/") || path.Clean(packagePath) != packagePath || strings.HasPrefix(packagePath, "../") {
		return fmt.Errorf("must be a normalized relative repo path")
	}
	if !strings.HasPrefix(packagePath, "internal/modules/") {
		return fmt.Errorf("must be module-owned")
	}
	if strings.HasPrefix(packagePath, "internal/modules/projections") {
		return fmt.Errorf("must not expose projection internals as a facade")
	}
	return nil
}

func topologicalProviderOrder(providers []*projectionProvider, byProviderKey map[string]*projectionProvider) ([]*projectionProvider, error) {
	remaining := map[string]*projectionProvider{}
	indegree := map[string]int{}
	outgoing := map[string][]string{}
	for _, provider := range providers {
		key := provider.descriptor.ProviderKey
		remaining[key] = provider
		indegree[key] = 0
	}
	for _, provider := range providers {
		key := provider.descriptor.ProviderKey
		for _, dependency := range provider.descriptor.RebuildAfter {
			if byProviderKey[dependency] == nil {
				return nil, fmt.Errorf("projection provider %q rebuild_after references unknown provider %q", key, dependency)
			}
			indegree[key]++
			outgoing[dependency] = append(outgoing[dependency], key)
		}
	}
	ordered := make([]*projectionProvider, 0, len(providers))
	for len(remaining) > 0 {
		ready := make([]*projectionProvider, 0)
		for key, provider := range remaining {
			if indegree[key] == 0 {
				ready = append(ready, provider)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("projection provider rebuild graph has a cycle")
		}
		sort.Slice(ready, func(left, right int) bool {
			return providerSortKey(ready[left]) < providerSortKey(ready[right])
		})
		next := ready[0]
		nextKey := next.descriptor.ProviderKey
		ordered = append(ordered, next)
		delete(remaining, nextKey)
		for _, dependent := range outgoing[nextKey] {
			indegree[dependent]--
		}
	}
	return ordered, nil
}

func providerSortKey(provider *projectionProvider) string {
	viewIDs := append([]string(nil), provider.descriptor.ViewSchemaIDs...)
	sort.Strings(viewIDs)
	firstView := ""
	if len(viewIDs) > 0 {
		firstView = viewIDs[0]
	}
	return provider.descriptor.ProviderKey + "\x00" + firstView
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

func builtInProjectionProviders() []projectionProvider {
	return []projectionProvider{
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "timeline",
				SourceOwnerKey:            "timeline",
				ViewSchemaIDs:             []string{timelineViewSchemaID},
				SourceRecordTypes:         []string{"timeline_event"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions", "timeline"},
				ProjectionTableFamilies:   []string{"timeline_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/timeline/workbookprojection"},
				CharacterizationRefs: []string{"internal/modules/timeline/phase3_projection_contract_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTimelineTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "host",
				SourceOwnerKey:            "entities",
				ViewSchemaIDs:             []string{hostsViewSchemaID},
				SourceRecordTypes:         []string{"host"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions"},
				ProjectionTableFamilies:   []string{"host_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/entities/hostidentity"},
				RebuildAfter:         []string{"timeline"},
				CharacterizationRefs: []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshHostTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentHostsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "identity",
				SourceOwnerKey:            "entities",
				ViewSchemaIDs:             []string{identitiesViewSchemaID},
				SourceRecordTypes:         []string{"identity"},
				SourceAuthorityModules:    []string{"entities", "evidence", "links", "revisions"},
				ProjectionTableFamilies:   []string{"identity_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/entities/hostidentity"},
				RebuildAfter:         []string{"host"},
				CharacterizationRefs: []string{"internal/modules/entities/phase4_integration_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshIdentityTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "indicator",
				SourceOwnerKey:            "indicators",
				ViewSchemaIDs:             []string{indicatorsViewSchemaID},
				SourceRecordTypes:         []string{"indicator"},
				SourceAuthorityModules:    []string{"indicators", "links", "revisions"},
				ProjectionTableFamilies:   []string{"indicator_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/indicators"},
				RebuildAfter:         []string{"identity"},
				CharacterizationRefs: []string{"internal/modules/indicators/phase9_indicators_test.go"},
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "assessment",
				SourceOwnerKey:            "assessments",
				ViewSchemaIDs:             []string{assessmentsViewSchemaID},
				SourceRecordTypes:         []string{"assessment"},
				SourceAuthorityModules:    []string{"assessments", "links", "revisions"},
				ProjectionTableFamilies:   []string{"assessment_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        assessmentprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/assessments"},
				RebuildAfter:         []string{"indicator"},
				CharacterizationRefs: []string{"internal/modules/assessments/phase9_assessment_contract_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshAssessmentTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:  projectionProviderDescriptorSchemaVersion,
				Status:         ProviderStatusActive,
				ProviderKey:    "artifact",
				SourceOwnerKey: "artifacts",
				ViewSchemaIDs: []string{
					notesViewSchemaID,
					commLogViewSchemaID,
					handoffViewSchemaID,
					statusReviewViewSchemaID,
					lessonViewSchemaID,
					findingsViewSchemaID,
					investigativeQueriesViewSchemaID,
					forensicKeywordsViewSchemaID,
				},
				SourceRecordTypes:         []string{"artifact"},
				SourceAuthorityModules:    []string{"artifacts", "links", "parties", "revisions"},
				ProjectionTableFamilies:   []string{"artifact_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        artifactprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/artifacts", "internal/modules/artifacts/linkednotes", "internal/modules/workbook"},
				RebuildAfter:         []string{"assessment"},
				CharacterizationRefs: []string{"internal/modules/workbook/phase9_coordination_surfaces_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshArtifactTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "evidence",
				SourceOwnerKey:            "evidence",
				ViewSchemaIDs:             []string{evidenceViewSchemaID},
				SourceRecordTypes:         []string{"evidence"},
				SourceAuthorityModules:    []string{"evidence", "revisions"},
				ProjectionTableFamilies:   []string{"evidence_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        evidenceprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/evidence"},
				RebuildAfter:         []string{"artifact"},
				CharacterizationRefs: []string{"internal/modules/evidence/phase5_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshEvidenceTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "party",
				SourceOwnerKey:            "parties",
				ViewSchemaIDs:             []string{partiesViewSchemaID},
				SourceRecordTypes:         []string{"party"},
				SourceAuthorityModules:    []string{"parties", "revisions"},
				ProjectionTableFamilies:   []string{"party_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        partyprojection.QuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/parties"},
				RebuildAfter:         []string{"evidence"},
				CharacterizationRefs: []string{"internal/modules/workbook/phase9_parties_integration_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshPartyTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "task_request",
				SourceOwnerKey:            "tasksdecisions",
				ViewSchemaIDs:             []string{taskRequestsViewSchemaID},
				SourceRecordTypes:         []string{"task_request"},
				SourceAuthorityModules:    []string{"links", "revisions", "tasksdecisions"},
				ProjectionTableFamilies:   []string{"task_request_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        taskdecisionprojection.TaskRequestQuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/tasksdecisions"},
				RebuildAfter:         []string{"party"},
				CharacterizationRefs: []string{"internal/modules/tasksdecisions/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshTaskRequestTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID)
			},
		},
		{
			descriptor: ProviderDescriptor{
				SchemaVersion:             projectionProviderDescriptorSchemaVersion,
				Status:                    ProviderStatusActive,
				ProviderKey:               "decision",
				SourceOwnerKey:            "tasksdecisions",
				ViewSchemaIDs:             []string{decisionsViewSchemaID},
				SourceRecordTypes:         []string{"decision"},
				SourceAuthorityModules:    []string{"links", "revisions", "tasksdecisions"},
				ProjectionTableFamilies:   []string{"decision_grid_projection"},
				ProjectionStorageOwnerKey: "projections",
				Capabilities: ProviderCapabilities{
					Query:           true,
					RefreshRow:      true,
					RestoreRebuild:  true,
					IncidentRebuild: true,
				},
				QuerySurfaces:        taskdecisionprojection.DecisionQuerySurfaces(),
				RestoreRebuild:       RestoreRebuildRequired,
				FacadePackages:       []string{"internal/modules/tasksdecisions"},
				RebuildAfter:         []string{"task_request"},
				CharacterizationRefs: []string{"internal/modules/tasksdecisions/phase9_task_decisions_store_test.go", "internal/modules/projections/query_test.go"},
			},
			refreshRowTx: func(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) error {
				return store.refreshDecisionTxCore(ctx, tx, recordID)
			},
			rebuildIncidentTx: func(ctx context.Context, store *Store, tx pgx.Tx, incidentID uuid.UUID) error {
				return store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID)
			},
		},
	}
}
