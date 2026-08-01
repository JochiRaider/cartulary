package appsupport

import (
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
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
	evidenceAttachments := evidence.NewTimelineAttachmentContribution(pool)
	timelineBundle := timelineassembly.NewBundle(pool, conflictTokens, appender, intents, evidenceAttachments)
	evidenceContribution := evidence.NewWorkbookContribution(pool, conflictTokens, appender, intents)
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(pool, conflictTokens, appender)
	if err != nil {
		panic(err)
	}
	catalog, err := workbookassembly.NewContributionCatalog(
		pool,
		timelineBundle.ProjectionCatalog.Catalog,
		timelineBundle.ProjectionCatalog.Query,
		timelineBundle.Facade,
		evidenceContribution,
		taskDecisionMutation,
		conflictTokens,
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
