package projections

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

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
	if err := validateUniqueStrings(descriptor.ProviderKey, "view_schema_ids", descriptor.ViewSchemaIDs); err != nil {
		return err
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
	if err := validateUniqueStrings(descriptor.ProviderKey, "projection_table_families", descriptor.ProjectionTableFamilies); err != nil {
		return err
	}
	if err := validateUniqueStrings(descriptor.ProviderKey, "rebuild_after", descriptor.RebuildAfter); err != nil {
		return err
	}
	if err := validateUniqueStrings(descriptor.ProviderKey, "characterization_refs", descriptor.CharacterizationRefs); err != nil {
		return err
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
