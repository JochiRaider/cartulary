package runtime

import (
	"fmt"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

type contractFacts struct {
	descriptors  providercontract.DescriptorSet
	intents      []providercontract.SurfaceIntent
	intentOwners map[string]string
}

func collectContractFacts(contributions []providercontract.Contribution) (contractFacts, error) {
	contributionsByOwner := make(map[string]providercontract.Contribution, len(contributions))
	descriptors := make([]providercontract.ProviderDescriptor, 0, len(requiredProviderCatalog))
	intents := make([]providercontract.SurfaceIntent, 0)
	for _, contribution := range contributions {
		if contribution.IsZero() {
			return contractFacts{}, fmt.Errorf("projection contribution is required")
		}
		owner := contribution.SourceOwnerModule()
		if _, exists := contributionsByOwner[owner]; exists {
			return contractFacts{}, fmt.Errorf("duplicate projection contribution owner %q", owner)
		}
		contributionsByOwner[owner] = contribution
		descriptors = append(descriptors, contribution.Descriptors()...)
		intents = append(intents, contribution.SurfaceIntents()...)
	}
	requiredOwners := requiredContributionOwners()
	for _, owner := range requiredOwners {
		if _, exists := contributionsByOwner[owner]; !exists {
			return contractFacts{}, fmt.Errorf("missing projection contribution owner %q", owner)
		}
	}
	if len(contributionsByOwner) != len(requiredOwners) {
		return contractFacts{}, fmt.Errorf(
			"projection contribution owner set has %d entries, want %d",
			len(contributionsByOwner),
			len(requiredOwners),
		)
	}

	declaredProviderIDs := make(map[string]int, len(descriptors))
	hasDuplicateProviderID := false
	for _, descriptor := range descriptors {
		declaredProviderIDs[descriptor.ProviderID]++
		if declaredProviderIDs[descriptor.ProviderID] > 1 {
			hasDuplicateProviderID = true
		}
	}
	if !hasDuplicateProviderID {
		for _, required := range requiredProviderCatalog {
			if declaredProviderIDs[required.providerID] == 0 {
				return contractFacts{}, fmt.Errorf("missing active projection provider %q", required.providerID)
			}
		}
	}
	descriptorSet, err := providercontract.NewDescriptorSet(descriptors)
	if err != nil {
		return contractFacts{}, err
	}
	for _, required := range requiredProviderCatalog {
		descriptor, exists := descriptorSet.Lookup(required.providerID)
		if !exists {
			return contractFacts{}, fmt.Errorf("missing active projection provider %q", required.providerID)
		}
		if descriptor.SourceOwnerModule != required.owner {
			return contractFacts{}, fmt.Errorf(
				"projection provider %q source owner %q, want %q",
				required.providerID,
				descriptor.SourceOwnerModule,
				required.owner,
			)
		}
		if !contributionDeclaresProvider(contributionsByOwner[required.owner], required.providerID) {
			return contractFacts{}, fmt.Errorf(
				"projection provider %q is not supplied by source owner %q",
				required.providerID,
				required.owner,
			)
		}
	}
	if descriptorSet.Len() != len(requiredProviderCatalog) {
		return contractFacts{}, fmt.Errorf(
			"active projection provider set has %d entries, want %d",
			descriptorSet.Len(),
			len(requiredProviderCatalog),
		)
	}
	intentOwners := make(map[string]string, len(intents))
	for _, contribution := range contributions {
		for _, intent := range contribution.SurfaceIntents() {
			if existing := intentOwners[intent.ViewSchemaID]; existing != "" {
				return contractFacts{}, fmt.Errorf("duplicate projection surface intent %q", intent.ViewSchemaID)
			}
			intentOwners[intent.ViewSchemaID] = contribution.SourceOwnerModule()
		}
	}
	return contractFacts{
		descriptors:  descriptorSet,
		intents:      cloneIntents(intents),
		intentOwners: intentOwners,
	}, nil
}

func requiredContributionOwners() []string {
	seen := make(map[string]struct{}, len(requiredProviderCatalog))
	owners := make([]string, 0, len(requiredProviderCatalog))
	for _, required := range requiredProviderCatalog {
		if _, exists := seen[required.owner]; exists {
			continue
		}
		seen[required.owner] = struct{}{}
		owners = append(owners, required.owner)
	}
	slices.Sort(owners)
	return owners
}

func contributionDeclaresProvider(contribution providercontract.Contribution, providerID string) bool {
	for _, descriptor := range contribution.Descriptors() {
		if descriptor.ProviderID == providerID {
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
