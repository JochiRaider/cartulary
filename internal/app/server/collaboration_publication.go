package server

import (
	"fmt"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func buildCollaborationPublicationCatalog(
	contributions []revisions.ProviderContribution,
	descriptors providercontract.DescriptorSet,
) (*collaboration.PublicationCatalog, error) {
	canonicalViews, err := canonicalCollaborationPublicationViews(descriptors)
	if err != nil {
		return nil, err
	}
	publications := make([]collaboration.PublicationContribution, 0, len(contributions))
	for _, contribution := range contributions {
		if len(contribution.Records) == 0 {
			continue
		}
		publication := collaboration.PublicationContribution{
			ContributionID: "collaboration.publication." + string(contribution.SourceOwnerModule) + "@1",
			SourceOwnerID:  string(contribution.SourceOwnerModule),
		}
		recordTypesByView := map[string]map[string]struct{}{}
		for _, record := range contribution.Records {
			for _, route := range record.RecordViewRoutes {
				for _, viewSchemaID := range route.ViewSchemaIDs {
					if recordTypesByView[viewSchemaID] == nil {
						recordTypesByView[viewSchemaID] = map[string]struct{}{}
					}
					recordTypesByView[viewSchemaID][record.RecordType] = struct{}{}
				}
			}
		}
		viewSchemaIDs := make([]string, 0, len(recordTypesByView))
		for viewSchemaID := range recordTypesByView {
			viewSchemaIDs = append(viewSchemaIDs, viewSchemaID)
		}
		slices.Sort(viewSchemaIDs)
		for _, viewSchemaID := range viewSchemaIDs {
			recordTypeSet := recordTypesByView[viewSchemaID]
			resource, ok := viewschema.LookupPublicResource(viewSchemaID)
			if !ok {
				return nil, fmt.Errorf("compose Collaboration publication catalog: unknown view schema %q", viewSchemaID)
			}
			recordTypes := make([]string, 0, len(recordTypeSet))
			for recordType := range recordTypeSet {
				recordTypes = append(recordTypes, recordType)
			}
			slices.Sort(recordTypes)
			fieldKeys := make([]string, 0, len(resource.Fields))
			for _, field := range resource.Fields {
				fieldKeys = append(fieldKeys, field.FieldKey)
			}
			slices.Sort(fieldKeys)
			publication.AffectedViews = append(publication.AffectedViews, collaboration.ViewPublicationContribution{
				ViewSchemaID: viewSchemaID, RecordTypes: recordTypes,
				PublicFieldKeys: fieldKeys, PatchFieldKeys: slices.Clone(fieldKeys),
			})
		}
		slices.SortFunc(publication.AffectedViews, func(left collaboration.ViewPublicationContribution, right collaboration.ViewPublicationContribution) int {
			if left.ViewSchemaID < right.ViewSchemaID {
				return -1
			}
			if left.ViewSchemaID > right.ViewSchemaID {
				return 1
			}
			return 0
		})
		publications = append(publications, publication)
	}
	slices.SortFunc(publications, func(left collaboration.PublicationContribution, right collaboration.PublicationContribution) int {
		if left.ContributionID < right.ContributionID {
			return -1
		}
		if left.ContributionID > right.ContributionID {
			return 1
		}
		return 0
	})
	return collaboration.NewPublicationCatalog(publications, canonicalViews)
}

func canonicalCollaborationPublicationViews(
	descriptors providercontract.DescriptorSet,
) ([]collaboration.CanonicalPublicationView, error) {
	canonicalViews := make([]collaboration.CanonicalPublicationView, 0)
	for _, descriptor := range descriptors.All() {
		if descriptor.Status != providercontract.ProviderStatusActive {
			continue
		}
		descriptorRecordTypes := slices.Clone(descriptor.SourceRecordTypes)
		slices.Sort(descriptorRecordTypes)
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			resource, ok := viewschema.LookupPublicResource(viewSchemaID)
			if !ok {
				return nil, fmt.Errorf("compose Collaboration publication catalog: projection descriptor %q has unknown view schema %q", descriptor.ProviderID, viewSchemaID)
			}
			resourceRecordTypes := slices.Clone(resource.SourceRecordTypes)
			slices.Sort(resourceRecordTypes)
			if !slices.Equal(descriptorRecordTypes, resourceRecordTypes) {
				return nil, fmt.Errorf("compose Collaboration publication catalog: projection descriptor %q record types disagree with view schema %q", descriptor.ProviderID, viewSchemaID)
			}
			canonicalViews = append(canonicalViews, collaboration.CanonicalPublicationView{
				ViewSchemaID: viewSchemaID, SourceOwnerID: descriptor.SourceOwnerModule,
				RecordTypes: resourceRecordTypes,
			})
		}
	}
	slices.SortFunc(canonicalViews, func(left collaboration.CanonicalPublicationView, right collaboration.CanonicalPublicationView) int {
		if left.ViewSchemaID < right.ViewSchemaID {
			return -1
		}
		if left.ViewSchemaID > right.ViewSchemaID {
			return 1
		}
		return 0
	})
	if len(canonicalViews) == 0 {
		return nil, fmt.Errorf("compose Collaboration publication catalog: active projection descriptor set is empty")
	}
	return canonicalViews, nil
}
