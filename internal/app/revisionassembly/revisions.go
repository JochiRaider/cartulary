package revisionassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewCommandService(db postgres.DB, attributionResolver revisions.ImportedAttributionResolver) (*revisions.CommandService, error) {
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Database:                    db,
		ImportedAttributionResolver: attributionResolver,
		ProjectionRebuilder:         projectionadapters.NewRowProjector(db),
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
