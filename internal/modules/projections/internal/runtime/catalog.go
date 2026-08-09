package runtime

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

var requiredProviderOwners = map[string]string{
	"timeline":     "timeline",
	"host":         "entities",
	"identity":     "entities",
	"indicator":    "indicators",
	"assessment":   "assessments",
	"artifact":     "artifacts",
	"evidence":     "evidence",
	"party":        "parties",
	"task_request": "tasksdecisions",
	"decision":     "tasksdecisions",
}

var requiredContributionOwners = []string{
	"artifacts",
	"assessments",
	"entities",
	"evidence",
	"indicators",
	"parties",
	"tasksdecisions",
	"timeline",
}

type DeclarativeCatalog struct {
	descriptors  providercontract.DescriptorSet
	intents      []providercontract.SurfaceIntent
	rebuildOrder []string
}

func NewDeclarativeCatalog(contributions []providercontract.Contribution) (*DeclarativeCatalog, error) {
	contributionsByOwner := make(map[string]providercontract.Contribution, len(contributions))
	descriptors := make([]providercontract.ProviderDescriptor, 0, len(requiredProviderOwners))
	intents := make([]providercontract.SurfaceIntent, 0)
	for _, contribution := range contributions {
		if contribution.IsZero() {
			return nil, fmt.Errorf("projection contribution is required")
		}
		owner := contribution.SourceOwnerModule()
		if _, exists := contributionsByOwner[owner]; exists {
			return nil, fmt.Errorf("duplicate projection contribution owner %q", owner)
		}
		contributionsByOwner[owner] = contribution
		descriptors = append(descriptors, contribution.Descriptors()...)
		intents = append(intents, contribution.SurfaceIntents()...)
	}
	for _, owner := range requiredContributionOwners {
		if _, exists := contributionsByOwner[owner]; !exists {
			return nil, fmt.Errorf("missing projection contribution owner %q", owner)
		}
	}
	if len(contributionsByOwner) != len(requiredContributionOwners) {
		return nil, fmt.Errorf("projection contribution owner set has %d entries, want %d", len(contributionsByOwner), len(requiredContributionOwners))
	}

	byProviderID := make(map[string]providercontract.ProviderDescriptor, len(descriptors))
	viewOwners := map[string]string{}
	tableOwners := map[string]string{}
	for _, descriptor := range descriptors {
		if err := validateDescriptor(descriptor); err != nil {
			return nil, err
		}
		if _, exists := byProviderID[descriptor.ProviderID]; exists {
			return nil, fmt.Errorf("duplicate projection provider_id %q", descriptor.ProviderID)
		}
		expectedOwner, required := requiredProviderOwners[descriptor.ProviderID]
		if !required {
			return nil, fmt.Errorf("unexpected active projection provider %q", descriptor.ProviderID)
		}
		if descriptor.SourceOwnerModule != expectedOwner {
			return nil, fmt.Errorf("projection provider %q source owner %q, want %q", descriptor.ProviderID, descriptor.SourceOwnerModule, expectedOwner)
		}
		contribution, exists := contributionsByOwner[descriptor.SourceOwnerModule]
		if !exists || !contributionDeclaresProvider(contribution, descriptor.ProviderID) {
			return nil, fmt.Errorf("projection provider %q is not supplied by source owner %q", descriptor.ProviderID, descriptor.SourceOwnerModule)
		}
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			if existing := viewOwners[viewSchemaID]; existing != "" {
				return nil, fmt.Errorf("duplicate projection view ownership for %q: %q and %q", viewSchemaID, existing, descriptor.ProviderID)
			}
			viewOwners[viewSchemaID] = descriptor.ProviderID
		}
		for _, tableID := range descriptor.ProjectionTableIDs {
			if existing := tableOwners[tableID]; existing != "" {
				return nil, fmt.Errorf("duplicate projection table ownership for %q: %q and %q", tableID, existing, descriptor.ProviderID)
			}
			tableOwners[tableID] = descriptor.ProviderID
		}
		byProviderID[descriptor.ProviderID] = descriptor
	}
	for providerID := range requiredProviderOwners {
		if _, exists := byProviderID[providerID]; !exists {
			return nil, fmt.Errorf("missing active projection provider %q", providerID)
		}
	}
	if len(byProviderID) != len(requiredProviderOwners) {
		return nil, fmt.Errorf("active projection provider set has %d entries, want %d", len(byProviderID), len(requiredProviderOwners))
	}

	intentOwners := map[string]string{}
	for _, contribution := range contributions {
		for _, intent := range contribution.SurfaceIntents() {
			providerID := viewOwners[intent.ViewSchemaID]
			if providerID == "" {
				return nil, fmt.Errorf("projection surface intent %q has no provider", intent.ViewSchemaID)
			}
			if byProviderID[providerID].SourceOwnerModule != contribution.SourceOwnerModule() {
				return nil, fmt.Errorf("projection surface intent %q is supplied by owner %q, want %q", intent.ViewSchemaID, contribution.SourceOwnerModule(), byProviderID[providerID].SourceOwnerModule)
			}
			if existing := intentOwners[intent.ViewSchemaID]; existing != "" {
				return nil, fmt.Errorf("duplicate projection surface intent %q", intent.ViewSchemaID)
			}
			intentOwners[intent.ViewSchemaID] = contribution.SourceOwnerModule()
		}
	}
	for viewSchemaID, providerID := range viewOwners {
		descriptor := byProviderID[providerID]
		if descriptor.Capabilities.Query && intentOwners[viewSchemaID] == "" {
			return nil, fmt.Errorf("projection provider %q query surface %q has no semantic intent", providerID, viewSchemaID)
		}
		if !descriptor.Capabilities.Query && intentOwners[viewSchemaID] != "" {
			return nil, fmt.Errorf("projection provider %q has semantic intent without query capability", providerID)
		}
	}

	rebuildOrder, err := topologicalOrder(byProviderID)
	if err != nil {
		return nil, err
	}
	descriptorSet, err := providercontract.NewDescriptorSet(descriptors)
	if err != nil {
		return nil, err
	}
	return &DeclarativeCatalog{
		descriptors:  descriptorSet,
		intents:      cloneIntents(intents),
		rebuildOrder: rebuildOrder,
	}, nil
}

func (catalog *DeclarativeCatalog) DescriptorSet() providercontract.DescriptorSet {
	if catalog == nil {
		return providercontract.DescriptorSet{}
	}
	return catalog.descriptors
}

func (catalog *DeclarativeCatalog) SurfaceIntents() []providercontract.SurfaceIntent {
	if catalog == nil {
		return nil
	}
	return cloneIntents(catalog.intents)
}

func (catalog *DeclarativeCatalog) Ready() bool {
	return catalog != nil && catalog.descriptors.Len() == len(requiredProviderOwners) && len(catalog.rebuildOrder) == len(requiredProviderOwners)
}

func validateDescriptor(descriptor providercontract.ProviderDescriptor) error {
	if descriptor.SchemaVersion != providercontract.DescriptorSchemaVersion {
		return fmt.Errorf("projection provider %q has unsupported schema version %q", descriptor.ProviderID, descriptor.SchemaVersion)
	}
	if descriptor.Status != providercontract.ProviderStatusActive {
		return fmt.Errorf("projection provider %q has non-active status %q", descriptor.ProviderID, descriptor.Status)
	}
	if strings.TrimSpace(descriptor.ProviderID) == "" || strings.TrimSpace(descriptor.SourceOwnerModule) == "" {
		return fmt.Errorf("projection provider identity and source owner are required")
	}
	if descriptor.ProjectionStorageOwnerModule != "projections" {
		return fmt.Errorf("projection provider %q storage owner %q, want projections", descriptor.ProviderID, descriptor.ProjectionStorageOwnerModule)
	}
	if len(descriptor.ViewSchemaIDs) == 0 || len(descriptor.ProjectionTableIDs) == 0 || len(descriptor.SourceRecordTypes) == 0 || len(descriptor.SourceAuthorityModules) == 0 {
		return fmt.Errorf("projection provider %q has incomplete ownership facts", descriptor.ProviderID)
	}
	if !contains(descriptor.SourceAuthorityModules, descriptor.SourceOwnerModule) {
		return fmt.Errorf("projection provider %q source authorities omit source owner %q", descriptor.ProviderID, descriptor.SourceOwnerModule)
	}
	for field, values := range map[string][]string{
		"view_schema_ids":          descriptor.ViewSchemaIDs,
		"projection_table_ids":     descriptor.ProjectionTableIDs,
		"source_record_types":      descriptor.SourceRecordTypes,
		"source_authority_modules": descriptor.SourceAuthorityModules,
		"facade_packages":          descriptor.FacadePackages,
		"rebuild_after":            descriptor.RebuildAfter,
	} {
		if err := validateUnique(field, descriptor.ProviderID, values); err != nil {
			return err
		}
	}
	for _, tableID := range descriptor.ProjectionTableIDs {
		if !strings.HasSuffix(tableID, "_grid_projection") || strings.ContainsAny(tableID, " ./\\") {
			return fmt.Errorf("projection provider %q has invalid projection table id %q", descriptor.ProviderID, tableID)
		}
	}
	if len(descriptor.FacadePackages) == 0 {
		return fmt.Errorf("projection provider %q has no facade packages", descriptor.ProviderID)
	}
	for _, facadePackage := range descriptor.FacadePackages {
		if path.Clean(facadePackage) != facadePackage || !strings.HasPrefix(facadePackage, "internal/modules/") || strings.HasPrefix(facadePackage, "internal/modules/projections") {
			return fmt.Errorf("projection provider %q has invalid facade package %q", descriptor.ProviderID, facadePackage)
		}
	}
	if !descriptor.Capabilities.RefreshRow || !descriptor.Capabilities.RestoreRebuild || !descriptor.Capabilities.IncidentRebuild {
		return fmt.Errorf("projection provider %q has incomplete active capabilities", descriptor.ProviderID)
	}
	if descriptor.RestoreRebuild != providercontract.RestoreRebuildRequired {
		return fmt.Errorf("projection provider %q restore participation %q, want required", descriptor.ProviderID, descriptor.RestoreRebuild)
	}
	return nil
}

func validateUnique(field string, providerID string, values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("projection provider %q has empty %s value", providerID, field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("projection provider %q has duplicate %s value %q", providerID, field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func contributionDeclaresProvider(contribution providercontract.Contribution, providerID string) bool {
	for _, descriptor := range contribution.Descriptors() {
		if descriptor.ProviderID == providerID {
			return true
		}
	}
	return false
}

func topologicalOrder(byProviderID map[string]providercontract.ProviderDescriptor) ([]string, error) {
	indegree := make(map[string]int, len(byProviderID))
	outgoing := map[string][]string{}
	for providerID := range byProviderID {
		indegree[providerID] = 0
	}
	for providerID, descriptor := range byProviderID {
		for _, dependency := range descriptor.RebuildAfter {
			if _, exists := byProviderID[dependency]; !exists {
				return nil, fmt.Errorf("projection provider %q rebuild dependency %q is missing", providerID, dependency)
			}
			indegree[providerID]++
			outgoing[dependency] = append(outgoing[dependency], providerID)
		}
	}
	order := make([]string, 0, len(byProviderID))
	for len(order) < len(byProviderID) {
		ready := make([]string, 0)
		for providerID, degree := range indegree {
			if degree == 0 && !contains(order, providerID) {
				ready = append(ready, providerID)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("projection provider rebuild graph has a cycle")
		}
		sort.Strings(ready)
		next := ready[0]
		order = append(order, next)
		indegree[next] = -1
		for _, dependent := range outgoing[next] {
			indegree[dependent]--
		}
	}
	return order, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func cloneIntents(intents []providercontract.SurfaceIntent) []providercontract.SurfaceIntent {
	result := make([]providercontract.SurfaceIntent, 0, len(intents))
	for _, intent := range intents {
		result = append(result, intent.Clone())
	}
	return result
}
