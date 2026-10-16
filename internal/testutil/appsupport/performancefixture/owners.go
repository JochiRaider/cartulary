package performancefixture

import (
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type Owners struct {
	Evidence          *evidence.PerformanceFixtureContribution
	Entities          *hostidentity.Store
	Timeline          *timeline.PerformanceFixtureContribution
	TimelineMutations *timeline.Facade
	Projections       *projectionassembly.Runtime
}

// NewOwners composes owner mutation and query facades for a harness-owned
// database through the same revision, projection, and conflict boundaries as
// production application assembly, without starting HTTP transport.
func NewOwners(pool postgres.DB, objects objectstore.TypedStore, conflictTokens conflicttokens.ConflictTokenCodec) (*Owners, error) {
	intents := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
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
	evidenceOwner := appsupport.NewEvidenceSourceRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
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
		EntityProjection:    projectionRuntime.EntityMutationRows(),
		AssessmentRows:      projectionRuntime.AssessmentPorts().Rows,
	})
	if err != nil {
		return nil, err
	}
	entityStore, err := hostidentity.NewStore(hostidentity.StoreDependencies{
		Postgres:               pool,
		Revisions:              appender,
		ProjectionMutationRows: projectionRuntime.EntityMutationRows(),
		ProjectionQueryReader:  projectionRuntime.EntityQueryReader(),
		KeepSavedIdempotency:   workbookassembly.NewConflictIdempotencyPort(pool),
		Collaboration:          intents,
	})
	if err != nil {
		return nil, err
	}
	return &Owners{
		Evidence:          evidence.NewPerformanceFixtureContribution(pool, objects, evidenceOwner),
		Entities:          entityStore,
		Timeline:          timelineBundle.PerformanceFixture,
		TimelineMutations: timelineBundle.Facade,
		Projections:       projectionRuntime,
	}, nil
}
