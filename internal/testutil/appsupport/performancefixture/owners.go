package performancefixture

import (
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type Owners struct {
	Entities    *hostidentity.Store
	Timeline    *timeline.PerformanceFixtureContribution
	Projections *projectionassembly.Runtime
}

// NewOwners composes owner mutation and query facades for a harness-owned
// database through the same revision, projection, and conflict boundaries as
// production application assembly, without starting HTTP transport.
func NewOwners(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) (*Owners, error) {
	intents := collaborationsupport.NewPublicationAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		return nil, err
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		return nil, err
	}
	projectionRuntime, err := projectionassembly.Build(pool)
	if err != nil {
		return nil, err
	}
	appender := revisionRuntime.Appender()
	conflictFields := revisionRuntime.ConflictFieldResolver()
	evidenceOwner := appsupport.NewEvidenceOwnerRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		appsupport.UnavailableEvidenceObjectStore(),
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime,
	)
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		ConflictFields:      conflictFields,
		Revisions:           appender,
		Collaboration:       intents,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projectionRuntime.TimelinePorts().Writer,
		EntityProjection:    projectionRuntime.EntityPorts().Writer,
		AssessmentRows:      projectionRuntime.AssessmentPorts().Rows,
	})
	if err != nil {
		return nil, err
	}
	entityPorts := projectionRuntime.EntityPorts()
	entityStore, err := hostidentity.NewStore(hostidentity.StoreDependencies{
		Postgres:             pool,
		Revisions:            appender,
		ProjectionWriter:     entityPorts.Writer,
		ProjectionReader:     entityPorts.Reader,
		KeepSavedIdempotency: workbookassembly.NewConflictIdempotencyPort(pool),
		Collaboration:        intents,
	})
	if err != nil {
		return nil, err
	}
	return &Owners{
		Entities:    entityStore,
		Timeline:    timelineBundle.PerformanceFixture,
		Projections: projectionRuntime,
	}, nil
}
