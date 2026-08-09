package providercontract

import "fmt"

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
}

func NewDescriptorSet(descriptors []ProviderDescriptor) (DescriptorSet, error) {
	if len(descriptors) == 0 {
		return DescriptorSet{}, fmt.Errorf("projection descriptor set is empty")
	}
	set := DescriptorSet{
		descriptors:  make([]ProviderDescriptor, 0, len(descriptors)),
		byProviderID: make(map[string]ProviderDescriptor, len(descriptors)),
	}
	for _, descriptor := range descriptors {
		if descriptor.ProviderID == "" {
			return DescriptorSet{}, fmt.Errorf("projection descriptor has empty provider_id")
		}
		if _, exists := set.byProviderID[descriptor.ProviderID]; exists {
			return DescriptorSet{}, fmt.Errorf("duplicate projection provider_id %q", descriptor.ProviderID)
		}
		cloned := descriptor.Clone()
		set.descriptors = append(set.descriptors, cloned)
		set.byProviderID[cloned.ProviderID] = cloned
	}
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

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string(nil), values...)
}
