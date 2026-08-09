package appsupport

import (
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func ArtifactProjectionRows(pool postgres.DB) artifactprojection.Rows {
	return mustBuildProjectionRuntime(pool).ArtifactPorts().Rows
}

func EvidenceProjectionRows(pool postgres.DB) evidenceprojection.Rows {
	return mustBuildProjectionRuntime(pool).EvidencePorts().Rows
}

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
	projectionRuntime := mustBuildProjectionRuntime(pool)
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		Revisions:           appender,
		Collaboration:       intents,
		EvidenceAttachments: evidenceAttachments,
		TimelineProjection:  projectionRuntime.TimelinePorts().Writer,
		EntityProjection:    projectionRuntime.EntityPorts().Writer,
		AssessmentRows:      projectionRuntime.AssessmentPorts().Rows,
	})
	if err != nil {
		panic(err)
	}
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    pool,
		Revisions:   appender,
		Projections: projectionRuntime.IndicatorPorts().Rows,
		SourceText:  indicatorassembly.NewSourceTextPort(projectionRuntime.SourceTextRows()),
	})
	if err != nil {
		panic(err)
	}
	evidenceContribution := evidence.NewWorkbookContribution(
		pool,
		conflictTokens,
		appender,
		intents,
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime.EvidencePorts().Rows,
	)
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projectionRuntime.TaskDecisionPorts().Rows,
	)
	if err != nil {
		panic(err)
	}
	catalog, err := workbookassembly.NewContributionCatalog(
		pool,
		projectionRuntime.DescriptorSet(),
		projectionRuntime,
		projectionRuntime.EntityPorts(),
		projectionRuntime.AssessmentPorts().Rows,
		projectionRuntime.ArtifactPorts().Rows,
		projectionRuntime.PartyPorts().Rows,
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
	store, err := workbookassembly.NewMutationStore(
		pool,
		catalog,
		appender,
		taskDecisionMutation,
		projectionRuntime.ArtifactPorts().Rows,
	)
	if err != nil {
		panic(err)
	}
	return store
}

func mustBuildProjectionRuntime(pool postgres.DB) *projectionassembly.Runtime {
	runtime, err := projectionassembly.Build(pool)
	if err != nil {
		panic(err)
	}
	return runtime
}
