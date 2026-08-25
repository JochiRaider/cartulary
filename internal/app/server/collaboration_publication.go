package server

import (
	"fmt"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func buildCollaborationPublicationCatalog(contributions []revisions.ProviderContribution) (*collaboration.PublicationCatalog, error) {
	publications := make([]collaboration.PublicationContribution, 0, len(contributions))
	for _, contribution := range contributions {
		if len(contribution.Records) == 0 {
			continue
		}
		publication := collaboration.PublicationContribution{
			ContributionID: "collaboration.publication." + string(contribution.SourceOwnerModule) + "@1",
			SourceOwnerID:  string(contribution.SourceOwnerModule),
		}
		seenRecords := map[string]struct{}{}
		seenViews := map[string]struct{}{}
		for _, record := range contribution.Records {
			if _, seen := seenRecords[record.RecordType]; !seen {
				publication.RecordTypes = append(publication.RecordTypes, record.RecordType)
				seenRecords[record.RecordType] = struct{}{}
			}
			for _, route := range record.RecordViewRoutes {
				for _, viewSchemaID := range route.ViewSchemaIDs {
					if _, seen := seenViews[viewSchemaID]; seen {
						continue
					}
					resource, ok := viewschema.LookupPublicResource(viewSchemaID)
					if !ok {
						return nil, fmt.Errorf("compose Collaboration publication catalog: unknown view schema %q", viewSchemaID)
					}
					fieldKeys := make([]string, 0, len(resource.Fields))
					for _, field := range resource.Fields {
						fieldKeys = append(fieldKeys, field.FieldKey)
					}
					slices.Sort(fieldKeys)
					publication.AffectedViews = append(publication.AffectedViews, collaboration.ViewPublicationContribution{
						ViewSchemaID: viewSchemaID, PublicFieldKeys: fieldKeys, PatchFieldKeys: slices.Clone(fieldKeys),
					})
					seenViews[viewSchemaID] = struct{}{}
				}
			}
		}
		slices.Sort(publication.RecordTypes)
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
	return collaboration.NewPublicationCatalog(publications)
}
