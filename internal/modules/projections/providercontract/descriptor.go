package providercontract

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
)

const DescriptorSchemaVersion = "projection_provider_descriptor.v3"

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

// ProviderDescriptor contains immutable declarative provider facts. It never
// carries executable callbacks, persistence handles, SQL, or query-engine
// values.
type ProviderDescriptor struct {
	SchemaVersion                string
	Status                       ProviderStatus
	ProviderID                   string
	SourceOwnerModule            string
	ViewSchemaIDs                []string
	SourceRecordTypes            []string
	SourceAuthorityModules       []string
	ProjectionTableIDs           []string
	ProjectionStorageOwnerModule string
	Capabilities                 ProviderCapabilities
	RestoreRebuild               RestoreRebuildParticipation
	FacadePackages               []string
	RebuildAfter                 []string
	CharacterizationRefs         []string
}

func (descriptor ProviderDescriptor) Clone() ProviderDescriptor {
	descriptor.ViewSchemaIDs = cloneStrings(descriptor.ViewSchemaIDs)
	descriptor.SourceRecordTypes = cloneStrings(descriptor.SourceRecordTypes)
	descriptor.SourceAuthorityModules = cloneStrings(descriptor.SourceAuthorityModules)
	descriptor.ProjectionTableIDs = cloneStrings(descriptor.ProjectionTableIDs)
	descriptor.FacadePackages = cloneStrings(descriptor.FacadePackages)
	descriptor.RebuildAfter = cloneStrings(descriptor.RebuildAfter)
	descriptor.CharacterizationRefs = cloneStrings(descriptor.CharacterizationRefs)
	return descriptor
}

type DescriptorSet struct {
	descriptors  []ProviderDescriptor
	byProviderID map[string]ProviderDescriptor
	rebuildOrder []string
}

func NewDescriptorSet(descriptors []ProviderDescriptor) (DescriptorSet, error) {
	if len(descriptors) == 0 {
		return DescriptorSet{}, fmt.Errorf("projection descriptor set is empty")
	}
	set := DescriptorSet{
		descriptors:  make([]ProviderDescriptor, 0, len(descriptors)),
		byProviderID: make(map[string]ProviderDescriptor, len(descriptors)),
		rebuildOrder: make([]string, 0, len(descriptors)),
	}
	viewOwners := make(map[string]string)
	tableOwners := make(map[string]string)
	for _, descriptor := range descriptors {
		if err := validateDescriptor(descriptor); err != nil {
			return DescriptorSet{}, err
		}
		if _, exists := set.byProviderID[descriptor.ProviderID]; exists {
			return DescriptorSet{}, fmt.Errorf("duplicate projection provider_id %q", descriptor.ProviderID)
		}
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			if existing := viewOwners[viewSchemaID]; existing != "" {
				return DescriptorSet{}, fmt.Errorf(
					"duplicate projection view ownership for %q: %q and %q",
					viewSchemaID,
					existing,
					descriptor.ProviderID,
				)
			}
			viewOwners[viewSchemaID] = descriptor.ProviderID
		}
		for _, tableID := range descriptor.ProjectionTableIDs {
			if existing := tableOwners[tableID]; existing != "" {
				return DescriptorSet{}, fmt.Errorf(
					"duplicate projection table ownership for %q: %q and %q",
					tableID,
					existing,
					descriptor.ProviderID,
				)
			}
			tableOwners[tableID] = descriptor.ProviderID
		}
		cloned := descriptor.Clone()
		set.descriptors = append(set.descriptors, cloned)
		set.byProviderID[cloned.ProviderID] = cloned
	}
	rebuildOrder, err := descriptorRebuildOrder(set.descriptors, set.byProviderID)
	if err != nil {
		return DescriptorSet{}, err
	}
	set.rebuildOrder = rebuildOrder
	return set, nil
}

func (set DescriptorSet) Len() int {
	return len(set.descriptors)
}

func (set DescriptorSet) All() []ProviderDescriptor {
	result := make([]ProviderDescriptor, 0, len(set.descriptors))
	for _, descriptor := range set.descriptors {
		result = append(result, descriptor.Clone())
	}
	return result
}

func (set DescriptorSet) Lookup(providerID string) (ProviderDescriptor, bool) {
	descriptor, ok := set.byProviderID[providerID]
	if !ok {
		return ProviderDescriptor{}, false
	}
	return descriptor.Clone(), true
}

// RebuildOrder returns defensive descriptor copies in deterministic dependency
// order. It is the sole declarative rebuild-order projection.
func (set DescriptorSet) RebuildOrder() []ProviderDescriptor {
	result := make([]ProviderDescriptor, 0, len(set.rebuildOrder))
	for _, providerID := range set.rebuildOrder {
		if descriptor, ok := set.byProviderID[providerID]; ok {
			result = append(result, descriptor.Clone())
		}
	}
	return result
}

func validateDescriptor(descriptor ProviderDescriptor) error {
	if descriptor.ProviderID == "" {
		return fmt.Errorf("projection descriptor has empty provider_id")
	}
	if descriptor.SchemaVersion != DescriptorSchemaVersion {
		return fmt.Errorf("projection provider %q declares unsupported schema_version %q", descriptor.ProviderID, descriptor.SchemaVersion)
	}
	switch descriptor.Status {
	case ProviderStatusActive, ProviderStatusDeprecated, ProviderStatusExperimental:
	default:
		return fmt.Errorf("projection provider %q declares unsupported status %q", descriptor.ProviderID, descriptor.Status)
	}
	if strings.TrimSpace(descriptor.SourceOwnerModule) == "" {
		return fmt.Errorf("projection provider %q has empty source_owner_module", descriptor.ProviderID)
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "view_schema_ids", descriptor.ViewSchemaIDs, true); err != nil {
		return err
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "source_record_types", descriptor.SourceRecordTypes, true); err != nil {
		return err
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "source_authority_modules", descriptor.SourceAuthorityModules, true); err != nil {
		return err
	}
	if !slices.Contains(descriptor.SourceAuthorityModules, descriptor.SourceOwnerModule) {
		return fmt.Errorf("projection provider %q source_authority_modules omit source_owner_module %q", descriptor.ProviderID, descriptor.SourceOwnerModule)
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "projection_table_ids", descriptor.ProjectionTableIDs, true); err != nil {
		return err
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "rebuild_after", descriptor.RebuildAfter, false); err != nil {
		return err
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "characterization_refs", descriptor.CharacterizationRefs, false); err != nil {
		return err
	}
	if descriptor.ProjectionStorageOwnerModule != "projections" {
		return fmt.Errorf("projection provider %q projection_storage_owner_module=%q must be projections", descriptor.ProviderID, descriptor.ProjectionStorageOwnerModule)
	}
	if descriptor.Capabilities.RestoreRebuild && !descriptor.Capabilities.IncidentRebuild {
		return fmt.Errorf("projection provider %q declares restore rebuild without incident rebuild capability", descriptor.ProviderID)
	}
	switch descriptor.RestoreRebuild {
	case RestoreRebuildRequired:
		if !descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares required restore rebuild without capability", descriptor.ProviderID)
		}
	case RestoreRebuildNonparticipating:
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares nonparticipating restore rebuild with capability", descriptor.ProviderID)
		}
	case RestoreRebuildUnsupported:
		if descriptor.Status == ProviderStatusActive {
			return fmt.Errorf("projection provider %q is active but declares unsupported restore rebuild", descriptor.ProviderID)
		}
		if descriptor.Capabilities.RestoreRebuild {
			return fmt.Errorf("projection provider %q declares unsupported restore rebuild with capability", descriptor.ProviderID)
		}
	default:
		return fmt.Errorf("projection provider %q declares unsupported restore_rebuild %q", descriptor.ProviderID, descriptor.RestoreRebuild)
	}
	if err := requireUniqueStrings(descriptor.ProviderID, "facade_packages", descriptor.FacadePackages, true); err != nil {
		return err
	}
	for _, packagePath := range descriptor.FacadePackages {
		if err := validateFacadePackagePath(packagePath); err != nil {
			return fmt.Errorf("projection provider %q facade package %q: %w", descriptor.ProviderID, packagePath, err)
		}
	}
	return nil
}

func requireUniqueStrings(providerID string, field string, values []string, required bool) error {
	if required && len(values) == 0 {
		return fmt.Errorf("projection provider %q declares no %s", providerID, field)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("projection provider %q declares empty %s value", providerID, field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("projection provider %q declares duplicate %s value %q", providerID, field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateFacadePackagePath(packagePath string) error {
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

func descriptorRebuildOrder(
	descriptors []ProviderDescriptor,
	byProviderID map[string]ProviderDescriptor,
) ([]string, error) {
	remaining := make(map[string]ProviderDescriptor, len(descriptors))
	indegree := make(map[string]int, len(descriptors))
	outgoing := make(map[string][]string, len(descriptors))
	for _, descriptor := range descriptors {
		remaining[descriptor.ProviderID] = descriptor
		indegree[descriptor.ProviderID] = 0
	}
	for _, descriptor := range descriptors {
		for _, dependency := range descriptor.RebuildAfter {
			if _, exists := byProviderID[dependency]; !exists {
				return nil, fmt.Errorf("projection provider %q rebuild_after references unknown provider %q", descriptor.ProviderID, dependency)
			}
			indegree[descriptor.ProviderID]++
			outgoing[dependency] = append(outgoing[dependency], descriptor.ProviderID)
		}
	}
	ordered := make([]string, 0, len(descriptors))
	for len(remaining) > 0 {
		ready := make([]ProviderDescriptor, 0)
		for providerID, descriptor := range remaining {
			if indegree[providerID] == 0 {
				ready = append(ready, descriptor)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("projection provider rebuild graph has a cycle")
		}
		sort.Slice(ready, func(left, right int) bool {
			return descriptorSortKey(ready[left]) < descriptorSortKey(ready[right])
		})
		next := ready[0].ProviderID
		ordered = append(ordered, next)
		delete(remaining, next)
		for _, dependent := range outgoing[next] {
			indegree[dependent]--
		}
	}
	return ordered, nil
}

func descriptorSortKey(descriptor ProviderDescriptor) string {
	viewIDs := slices.Clone(descriptor.ViewSchemaIDs)
	slices.Sort(viewIDs)
	firstView := ""
	if len(viewIDs) > 0 {
		firstView = viewIDs[0]
	}
	return descriptor.ProviderID + "\x00" + firstView
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string(nil), values...)
}
