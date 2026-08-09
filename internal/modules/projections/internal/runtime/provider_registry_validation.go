package runtime

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func validateProvider(provider Provider) error {
	descriptor := provider.descriptor
	if descriptor.SchemaVersion != providercontract.DescriptorSchemaVersion {
		return fmt.Errorf("projection provider %q declares unsupported schema_version %q", descriptor.ProviderID, descriptor.SchemaVersion)
	}
	switch descriptor.Status {
	case providercontract.ProviderStatusActive, providercontract.ProviderStatusDeprecated, providercontract.ProviderStatusExperimental:
	default:
		return fmt.Errorf("projection provider %q declares unsupported status %q", descriptor.ProviderID, descriptor.Status)
	}
	if descriptor.ProviderID == "" {
		return fmt.Errorf("projection provider has empty provider_id")
	}
	if descriptor.SourceOwnerModule == "" {
		return fmt.Errorf("projection provider %q has empty source_owner_module", descriptor.ProviderID)
	}
	if len(descriptor.ViewSchemaIDs) == 0 {
		return fmt.Errorf("projection provider %q declares no view_schema_ids", descriptor.ProviderID)
	}
	if err := validateUniqueStrings(descriptor.ProviderID, "view_schema_ids", descriptor.ViewSchemaIDs); err != nil {
		return err
	}
	if len(descriptor.SourceRecordTypes) == 0 {
		return fmt.Errorf("projection provider %q declares no source_record_types", descriptor.ProviderID)
	}
	if err := validateUniqueStrings(descriptor.ProviderID, "source_record_types", descriptor.SourceRecordTypes); err != nil {
		return err
	}
	if len(descriptor.SourceAuthorityModules) == 0 {
		return fmt.Errorf("projection provider %q declares no source_authority_modules", descriptor.ProviderID)
	}
	if err := validateSourceAuthorityModules(descriptor); err != nil {
		return err
	}
	if len(descriptor.ProjectionTableIDs) == 0 {
		return fmt.Errorf("projection provider %q declares no projection_table_ids", descriptor.ProviderID)
	}
	if err := validateUniqueStrings(descriptor.ProviderID, "projection_table_ids", descriptor.ProjectionTableIDs); err != nil {
		return err
	}
	if err := validateUniqueStrings(descriptor.ProviderID, "rebuild_after", descriptor.RebuildAfter); err != nil {
		return err
	}
	if err := validateUniqueStrings(descriptor.ProviderID, "characterization_refs", descriptor.CharacterizationRefs); err != nil {
		return err
	}
	if descriptor.ProjectionStorageOwnerModule == "" {
		return fmt.Errorf("projection provider %q has empty projection_storage_owner_module", descriptor.ProviderID)
	}
	if descriptor.Capabilities.RefreshRow && provider.refreshRowTx == nil {
		return fmt.Errorf("projection provider %q declares refresh support without implementation", descriptor.ProviderID)
	}
	if !descriptor.Capabilities.RefreshRow && provider.refreshRowTx != nil {
		return fmt.Errorf("projection provider %q has refresh implementation without capability", descriptor.ProviderID)
	}
	if descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx == nil {
		return fmt.Errorf("projection provider %q declares incident rebuild support without implementation", descriptor.ProviderID)
	}
	if !descriptor.Capabilities.IncidentRebuild && provider.rebuildIncidentTx != nil {
		return fmt.Errorf("projection provider %q has incident rebuild implementation without capability", descriptor.ProviderID)
	}
	if descriptor.Capabilities.RestoreRebuild && !descriptor.Capabilities.IncidentRebuild {
		return fmt.Errorf("projection provider %q declares restore rebuild without incident rebuild capability", descriptor.ProviderID)
	}
	if descriptor.Capabilities.Query != (provider.queryStrategy != queryStrategyNone) {
		return fmt.Errorf("projection provider %q query capability does not match registered query strategy", descriptor.ProviderID)
	}
	switch provider.queryStrategy {
	case queryStrategyNone:
		if len(provider.queryPlans) != 0 {
			return fmt.Errorf("projection provider %q has compiled plans without a query strategy", descriptor.ProviderID)
		}
	case queryStrategyCompiledPlan:
		if len(provider.queryPlans) == 0 {
			return fmt.Errorf("projection provider %q compiled query strategy has no plans", descriptor.ProviderID)
		}
	case queryStrategySourceOwnerHydration:
		if len(provider.queryPlans) != 0 {
			return fmt.Errorf("projection provider %q source-owner hydration strategy has compiled plans", descriptor.ProviderID)
		}
	default:
		return fmt.Errorf("projection provider %q has unsupported query strategy %d", descriptor.ProviderID, provider.queryStrategy)
	}
	declaredViews := map[string]struct{}{}
	for _, viewSchemaID := range descriptor.ViewSchemaIDs {
		declaredViews[viewSchemaID] = struct{}{}
	}
	seenPlans := map[string]struct{}{}
	querySurfaces, err := providerPlans(provider)
	if err != nil {
		return err
	}
	for _, surface := range querySurfaces {
		if surface.viewSchemaID == "" {
			return fmt.Errorf("projection provider %q declares query surface with empty view_schema_id", descriptor.ProviderID)
		}
		if _, ok := declaredViews[surface.viewSchemaID]; !ok {
			return fmt.Errorf("projection provider %q query surface %q is not one of its view_schema_ids", descriptor.ProviderID, surface.viewSchemaID)
		}
		if _, exists := seenPlans[surface.viewSchemaID]; exists {
			return fmt.Errorf("projection provider %q declares duplicate query surface %q", descriptor.ProviderID, surface.viewSchemaID)
		}
		seenPlans[surface.viewSchemaID] = struct{}{}
	}
	switch descriptor.RestoreRebuild {
	case providercontract.RestoreRebuildRequired:
		if !descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares required restore rebuild without capability", descriptor.ProviderID)
		}
		if provider.rebuildIncidentTx == nil {
			return fmt.Errorf("projection provider %q declares required restore rebuild without implementation", descriptor.ProviderID)
		}
	case providercontract.RestoreRebuildNonparticipating:
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares nonparticipating restore rebuild with capability", descriptor.ProviderID)
		}
	case providercontract.RestoreRebuildUnsupported:
		if descriptor.Status == providercontract.ProviderStatusActive {
			return fmt.Errorf("projection provider %q is active but declares unsupported restore rebuild", descriptor.ProviderID)
		}
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares unsupported restore rebuild with capability", descriptor.ProviderID)
		}
	default:
		return fmt.Errorf("projection provider %q declares unsupported restore_rebuild %q", descriptor.ProviderID, descriptor.RestoreRebuild)
	}
	if len(descriptor.FacadePackages) == 0 {
		return fmt.Errorf("projection provider %q declares no facade_packages", descriptor.ProviderID)
	}
	seenFacadePackages := map[string]struct{}{}
	for _, packagePath := range descriptor.FacadePackages {
		if err := validateFacadePackagePath(packagePath); err != nil {
			return fmt.Errorf("projection provider %q facade package %q: %w", descriptor.ProviderID, packagePath, err)
		}
		if _, exists := seenFacadePackages[packagePath]; exists {
			return fmt.Errorf("projection provider %q declares duplicate facade package %q", descriptor.ProviderID, packagePath)
		}
		seenFacadePackages[packagePath] = struct{}{}
	}
	if descriptor.ProjectionStorageOwnerModule != "projections" {
		return fmt.Errorf("projection provider %q projection_storage_owner_module=%q must be projections", descriptor.ProviderID, descriptor.ProjectionStorageOwnerModule)
	}
	return nil
}

func validateSemanticIntents(
	registry *providerRegistry,
	intents []providercontract.SurfaceIntent,
	intentOwners map[string]string,
) error {
	if registry == nil || len(intents) == 0 {
		return fmt.Errorf("projection query intents are empty")
	}
	intentByView := make(map[string]providercontract.SurfaceIntent, len(intents))
	for _, intent := range intents {
		if strings.TrimSpace(intent.ViewSchemaID) == "" {
			return fmt.Errorf("projection query intent has empty view_schema_id")
		}
		if len(intent.FieldKeys) == 0 {
			return fmt.Errorf("projection query intent %q has no field keys", intent.ViewSchemaID)
		}
		if _, exists := intentByView[intent.ViewSchemaID]; exists {
			return fmt.Errorf("duplicate projection query intent %q", intent.ViewSchemaID)
		}
		if err := validateUniqueStrings(intent.ViewSchemaID, "field_keys", intent.FieldKeys); err != nil {
			return err
		}
		provider, exists := registry.providerForView(intent.ViewSchemaID)
		if !exists {
			return fmt.Errorf("projection surface intent %q has no provider", intent.ViewSchemaID)
		}
		if owner := intentOwners[intent.ViewSchemaID]; owner != provider.descriptor.SourceOwnerModule {
			return fmt.Errorf(
				"projection surface intent %q is supplied by owner %q, want %q",
				intent.ViewSchemaID,
				owner,
				provider.descriptor.SourceOwnerModule,
			)
		}
		if !provider.descriptor.Capabilities.Query {
			return fmt.Errorf("projection provider %q has semantic intent without query capability", provider.descriptor.ProviderID)
		}
		schema, exists := viewschema.Lookup(intent.ViewSchemaID)
		if !exists {
			return fmt.Errorf("projection query intent %q has no view schema", intent.ViewSchemaID)
		}
		schemaFields := make([]string, 0, len(schema.Fields()))
		for fieldKey := range schema.Fields() {
			schemaFields = append(schemaFields, fieldKey)
		}
		if !equalStringSets(intent.FieldKeys, schemaFields) {
			return fmt.Errorf(
				"projection query intent %q fields do not match its view schema",
				intent.ViewSchemaID,
			)
		}
		if provider.queryStrategy == queryStrategyCompiledPlan {
			plan, exists := registry.querySurfaces[intent.ViewSchemaID]
			if !exists {
				return fmt.Errorf("projection query intent %q has no private compiled plan", intent.ViewSchemaID)
			}
			planFields := make([]string, 0, len(plan.fields))
			for _, field := range plan.fields {
				planFields = append(planFields, field.key)
			}
			if !equalStringSets(intent.FieldKeys, planFields) {
				return fmt.Errorf(
					"projection query intent %q fields do not match its private compiled plan",
					intent.ViewSchemaID,
				)
			}
		}
		intentByView[intent.ViewSchemaID] = intent.Clone()
	}
	for _, provider := range registry.providers {
		for _, viewSchemaID := range provider.descriptor.ViewSchemaIDs {
			_, hasIntent := intentByView[viewSchemaID]
			if provider.descriptor.Capabilities.Query && !hasIntent {
				return fmt.Errorf(
					"projection provider %q query surface %q has no semantic intent",
					provider.descriptor.ProviderID,
					viewSchemaID,
				)
			}
			if !provider.descriptor.Capabilities.Query && hasIntent {
				return fmt.Errorf(
					"projection provider %q has semantic intent without query capability",
					provider.descriptor.ProviderID,
				)
			}
		}
	}
	for viewSchemaID := range registry.querySurfaces {
		if _, exists := intentByView[viewSchemaID]; !exists {
			return fmt.Errorf("private compiled plan %q has no semantic intent", viewSchemaID)
		}
	}
	return nil
}

func equalStringSets(left []string, right []string) bool {
	leftCopy := slices.Clone(left)
	rightCopy := slices.Clone(right)
	slices.Sort(leftCopy)
	slices.Sort(rightCopy)
	return slices.Equal(leftCopy, rightCopy)
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

func validateSourceAuthorityModules(descriptor providercontract.ProviderDescriptor) error {
	if err := validateUniqueStrings(descriptor.ProviderID, "source_authority_modules", descriptor.SourceAuthorityModules); err != nil {
		return err
	}
	includesSourceOwner := false
	for _, module := range descriptor.SourceAuthorityModules {
		if module == descriptor.SourceOwnerModule {
			includesSourceOwner = true
		}
	}
	if !includesSourceOwner {
		return fmt.Errorf("projection provider %q source_authority_modules omit source_owner_module %q", descriptor.ProviderID, descriptor.SourceOwnerModule)
	}
	return nil
}

func providerPlans(provider Provider) ([]genericSurface, error) {
	surfaces := make([]genericSurface, 0, len(provider.queryPlans))
	for _, surface := range provider.queryPlans {
		converted, err := genericSurfaceFromPlan(surface)
		if err != nil {
			return nil, fmt.Errorf("projection provider %q query surface: %w", provider.descriptor.ProviderID, err)
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

func topologicalProviderOrder(providers []*Provider, byProviderKey map[string]*Provider) ([]*Provider, error) {
	remaining := map[string]*Provider{}
	indegree := map[string]int{}
	outgoing := map[string][]string{}
	for _, provider := range providers {
		key := provider.descriptor.ProviderID
		remaining[key] = provider
		indegree[key] = 0
	}
	for _, provider := range providers {
		key := provider.descriptor.ProviderID
		for _, dependency := range provider.descriptor.RebuildAfter {
			if byProviderKey[dependency] == nil {
				return nil, fmt.Errorf("projection provider %q rebuild_after references unknown provider %q", key, dependency)
			}
			indegree[key]++
			outgoing[dependency] = append(outgoing[dependency], key)
		}
	}
	ordered := make([]*Provider, 0, len(providers))
	for len(remaining) > 0 {
		ready := make([]*Provider, 0)
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
		nextKey := next.descriptor.ProviderID
		ordered = append(ordered, next)
		delete(remaining, nextKey)
		for _, dependent := range outgoing[nextKey] {
			indegree[dependent]--
		}
	}
	return ordered, nil
}

func providerSortKey(provider *Provider) string {
	viewIDs := append([]string(nil), provider.descriptor.ViewSchemaIDs...)
	sort.Strings(viewIDs)
	firstView := ""
	if len(viewIDs) > 0 {
		firstView = viewIDs[0]
	}
	return provider.descriptor.ProviderID + "\x00" + firstView
}
