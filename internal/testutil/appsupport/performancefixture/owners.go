package performancefixture

import (
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

type Owners struct {
	Entities    *hostidentity.Store
	Timeline    *timeline.Facade
	Projections *projectionassembly.Runtime
}

// NewOwners composes owner mutation and query facades for a harness-owned
// database through the same revision, projection, and conflict boundaries as
// production application assembly, without starting HTTP transport.
func NewOwners(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) (*Owners, error) {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		return nil, err
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
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
	return &Owners{
		Entities: hostidentity.NewStore(
			pool,
			appender,
			workbookassembly.NewConflictIdempotencyPort(pool),
			entityPorts.Writer,
			hostidentity.WithProjectionReader(entityPorts.Reader),
		),
		Timeline:    timelineBundle.Facade,
		Projections: projectionRuntime,
	}, nil
}
