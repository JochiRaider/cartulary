package revisionassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
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

func NewCommandService(db postgres.DB, attributionResolver revisions.ImportedAttributionResolver, projectionRebuilder revisions.ProjectionRebuilder) (*revisions.CommandService, error) {
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Database:                    db,
		ImportedAttributionResolver: attributionResolver,
		ProjectionRebuilder:         projectionRebuilder,
		CollaborationIntents:        collaboration.NewStore(db, nil),
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
