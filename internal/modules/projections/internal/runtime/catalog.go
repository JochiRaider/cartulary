package runtime

import (
	"fmt"

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

type contractFacts struct {
	descriptors  providercontract.DescriptorSet
	intents      []providercontract.SurfaceIntent
	intentOwners map[string]string
}

func collectContractFacts(contributions []providercontract.Contribution) (contractFacts, error) {
	contributionsByOwner := make(map[string]providercontract.Contribution, len(contributions))
	descriptors := make([]providercontract.ProviderDescriptor, 0, len(requiredProviderOwners))
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
	for _, owner := range requiredContributionOwners {
		if _, exists := contributionsByOwner[owner]; !exists {
			return contractFacts{}, fmt.Errorf("missing projection contribution owner %q", owner)
		}
	}
	if len(contributionsByOwner) != len(requiredContributionOwners) {
		return contractFacts{}, fmt.Errorf(
			"projection contribution owner set has %d entries, want %d",
			len(contributionsByOwner),
			len(requiredContributionOwners),
		)
	}

	descriptorSet, err := providercontract.NewDescriptorSet(descriptors)
	if err != nil {
		return contractFacts{}, err
	}
	for providerID, expectedOwner := range requiredProviderOwners {
		descriptor, exists := descriptorSet.Lookup(providerID)
		if !exists {
			return contractFacts{}, fmt.Errorf("missing active projection provider %q", providerID)
		}
		if descriptor.SourceOwnerModule != expectedOwner {
			return contractFacts{}, fmt.Errorf(
				"projection provider %q source owner %q, want %q",
				providerID,
				descriptor.SourceOwnerModule,
				expectedOwner,
			)
		}
		if !contributionDeclaresProvider(contributionsByOwner[expectedOwner], providerID) {
			return contractFacts{}, fmt.Errorf(
				"projection provider %q is not supplied by source owner %q",
				providerID,
				expectedOwner,
			)
		}
	}
	if descriptorSet.Len() != len(requiredProviderOwners) {
		return contractFacts{}, fmt.Errorf(
			"active projection provider set has %d entries, want %d",
			descriptorSet.Len(),
			len(requiredProviderOwners),
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
