package revisionassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type projectionServices interface {
	revisions.ProjectionServices
}

func NewCommandService(db postgres.DB, attributionResolver revisions.ImportedAttributionResolver, projections projectionServices) (*revisions.CommandService, error) {
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Database:                    db,
		ImportedAttributionResolver: attributionResolver,
		Projections:                 projections,
		ProviderContributions: []revisions.ProviderContribution{
			artifacts.RevisionProviderContribution(),
			assessments.RevisionProviderContribution(),
			entities.RevisionProviderContribution(),
			evidence.RevisionProviderContribution(),
			indicators.RevisionProviderContribution(),
			links.RevisionProviderContribution(),
			parties.RevisionProviderContribution(),
			tasksdecisions.RevisionProviderContribution(),
			timeline.RevisionProviderContribution(),
		},
	})
}
