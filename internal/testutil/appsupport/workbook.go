package appsupport

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func ArtifactProjectionRows(pool postgres.DB) artifactprojection.Rows {
	return mustBuildProjectionRuntime(pool).ArtifactPorts().Rows
}

func EvidenceProjectionRows(pool postgres.DB) evidenceprojection.MutationRows {
	return mustBuildProjectionRuntime(pool).EvidenceMutationRows()
}

func NewEvidenceSourceRuntime(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRuntime *projectionassembly.Runtime,
) *evidence.SourceRuntime {
	mutationDependencies, err := workbookassembly.NewEvidenceMutationDependencies(
		pool,
		appender,
		conflictFields,
		keepSaved,
		projectionRuntime.EvidenceMutationRows(),
		projectionRuntime.EvidenceAssociationEffects(),
		intents,
	)
	if err != nil {
		panic(err)
	}
	runtime, err := evidence.NewSourceRuntime(evidence.SourceRuntimeDependencies{
		Postgres:       pool,
		ConflictTokens: &conflictTokens,
		Revisions:      appender,
		Collaboration:  intents,
		Mutation:       mutationDependencies,
	})
	if err != nil {
		panic(err)
	}
	return runtime
}

func NewEvidenceSourceRuntimeForTimeline(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
	projectionRuntime *projectionassembly.Runtime,
) *evidence.SourceRuntime {
	conflictFields, err := revisionsupport.NewConflictFieldResolver()
	if err != nil {
		panic(err)
	}
	return NewEvidenceSourceRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime,
	)
}

// NewEvidenceMutationOwner composes the Evidence mutation owner for tests that
// exercise source-owner semantics directly.
func NewEvidenceMutationOwner(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) evidence.MutationContribution {
	intents := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	runtime := NewEvidenceSourceRuntime(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		intents,
		revisionRuntime.ConflictFieldResolver(),
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime,
	)
	return runtime.MutationContribution()
}

// NewTaskDecisionOwner composes the source-owner mutation facade for tests
// that verify Task/Decision lifecycle behavior directly.
func NewTaskDecisionOwner(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *tasksdecisions.MutationFacade {
	intents := collaborationsupport.NewRecordChangedAppender()
	return NewTaskDecisionOwnerWithPublications(pool, conflictTokens, intents)
}

// NewTaskDecisionOwnerWithPublications composes the source-owner mutation
// facade with a caller-supplied Collaboration appender for transaction-bound
// failure and publication characterization tests.
func NewTaskDecisionOwnerWithPublications(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	intents collaboration.RecordChangedAppender,
) *tasksdecisions.MutationFacade {
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		panic(err)
	}
	owner, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		mustBuildProjectionRuntime(pool).TaskDecisionMutationRows(),
		intents,
	)
	if err != nil {
		panic(err)
	}
	return owner
}

// NewWorkbookCatalog composes the same immutable provider catalog used by the
// server for focused generic coordination tests.
func NewWorkbookCatalog(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *workbook.WorkbookContributionCatalog {
	intents := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		panic(err)
	}
	appender := revisionRuntime.Appender()
	conflictFields := revisionRuntime.ConflictFieldResolver()
	projectionRuntime := mustBuildProjectionRuntime(pool)
	evidenceOwner := NewEvidenceSourceRuntime(
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
		panic(err)
	}
	indicatorOwner, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referencefixture.IndicatorRegistry{},
		Postgres:        pool,
		Idempotency:     indicatorassembly.NewIdempotencyPort(authn.NewStore(pool)),
		IncidentState:   admission.NewChecker(pool),
		Revisions:       appender,
		RecordEnvelopes: records.NewStore(pool),
		Projections:     projectionRuntime.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projectionRuntime.SourceTextRows()),
		Collaboration:   intents,
		Clock:           func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		panic(err)
	}
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projectionRuntime.TaskDecisionMutationRows(),
		intents,
	)
	if err != nil {
		panic(err)
	}
	artifactMutation, err := workbookassembly.NewArtifactMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projectionRuntime.ArtifactPorts().Rows,
		intents,
	)
	if err != nil {
		panic(err)
	}
	catalog, err := workbookassembly.NewContributionCatalog(
		workbookassembly.ContributionDependencies{
			Postgres:              pool,
			ProjectionDescriptors: projectionRuntime.DescriptorSet(),
			ProjectionQueries:     projectionRuntime,
			EntityMutationRows:    projectionRuntime.EntityMutationRows(),
			EntityQueryReader:     projectionRuntime.EntityQueryReader(),
			AssessmentProjections: projectionRuntime.AssessmentPorts().Rows,
			PartyProjections:      projectionRuntime.PartyPorts().Rows,
			IndicatorOwner:        indicatorOwner,
			TimelineOwner:         timelineBundle.Facade,
			EvidenceOwner:         evidenceOwner.MutationContribution(),
			ArtifactOwner:         artifactMutation,
			TaskDecisionOwner:     taskDecisionMutation,
			ConflictTokens:        conflictTokens,
			ConflictFields:        conflictFields,
			Revisions:             appender,
			CollaborationIntents:  intents,
		},
	)
	if err != nil {
		panic(err)
	}
	return catalog
}

// NewAssessmentOwner composes the Assessment owner for tests that exercise
// source semantics through its owner facade.
func NewAssessmentOwner(pool postgres.DB) *assessments.Facade {
	intents := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	owner, err := workbookassembly.NewAssessmentMutationContribution(
		pool,
		projectionRuntime.AssessmentPorts().Rows,
		hostidentity.NewSourceFacts(),
		revisionRuntime.Appender(),
		intents,
	)
	if err != nil {
		panic(err)
	}
	return owner
}

// NewPartyOwner composes the Party mutation owner for tests that exercise
// Party source semantics directly.
func NewPartyOwner(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) *parties.MutationFacade {
	intents := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(contributions...)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	owner, err := workbookassembly.NewPartyMutationContribution(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		projectionRuntime.PartyPorts().Rows,
		intents,
	)
	if err != nil {
		panic(err)
	}
	return owner
}

func mustBuildProjectionRuntime(pool postgres.DB) *projectionassembly.Runtime {
	runtime, err := projectionassembly.Build(pool)
	if err != nil {
		panic(err)
	}
	return runtime
}
