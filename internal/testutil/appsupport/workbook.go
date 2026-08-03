package appsupport

import (
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// NewWorkbookStore composes the same code-backed projection catalog used by
// the server for focused module tests that do not need an HTTP runtime.
func NewWorkbookStore(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *workbook.Store {
	intents := collaboration.NewIntentAppender()
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		revisionassembly.CurrentProviderContributions()...,
	)
	if err != nil {
		panic(err)
	}
	appender := revisionRuntime.Appender()
	conflictFields := revisionRuntime.ConflictFieldResolver()
	evidenceAttachments := evidence.NewTimelineAttachmentContribution(pool)
	timelineBundle := timelineassembly.NewBundle(pool, conflictTokens, appender, intents, evidenceAttachments)
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    pool,
		Revisions:   appender,
		Projections: timelineBundle.ProjectionCoordinator,
	})
	if err != nil {
		panic(err)
	}
	evidenceContribution := evidence.NewWorkbookContribution(pool, conflictTokens, appender, intents, conflictFields, workbookassembly.NewConflictIdempotencyPort(pool))
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(pool, conflictTokens, appender, conflictFields)
	if err != nil {
		panic(err)
	}
	catalog, err := workbookassembly.NewContributionCatalog(
		pool,
		timelineBundle.ProjectionCatalog.Catalog,
		timelineBundle.ProjectionCatalog.Query,
		indicatorOwner,
		timelineBundle.Facade,
		evidenceContribution,
		taskDecisionMutation,
		conflictTokens,
		conflictFields,
		appender,
		intents,
	)
	if err != nil {
		panic(err)
	}
	store, err := workbookassembly.NewMutationStore(pool, catalog, appender, taskDecisionMutation)
	if err != nil {
		panic(err)
	}
	return store
}
