package appsupport

import (
	"context"
	"errors"
	"io"
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
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var errUnavailableEvidenceObjectStore = errors.New("test Evidence object store is unavailable")

func ArtifactProjectionRows(pool postgres.DB) artifactprojection.Rows {
	return mustBuildProjectionRuntime(pool).ArtifactPorts().Rows
}

func EvidenceProjectionRows(pool postgres.DB) evidenceprojection.Rows {
	return mustBuildProjectionRuntime(pool).EvidencePorts().Rows
}

func NewEvidenceOwnerRuntime(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	objects objectstore.TypedStore,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRuntime *projectionassembly.Runtime,
) *evidence.OwnerRuntime {
	runtime, err := evidence.NewOwnerRuntime(evidence.OwnerRuntimeDependencies{
		Postgres:            pool,
		ConflictTokens:      &conflictTokens,
		Revisions:           appender,
		Collaboration:       intents,
		ObjectStore:         objects,
		ConflictFields:      conflictFields,
		ConflictIdempotency: keepSaved,
		Projections:         projectionRuntime.EvidencePorts(),
		CleanupObserver:     evidenceCleanupObserver{},
		Now:                 time.Now,
	})
	if err != nil {
		panic(err)
	}
	return runtime
}

func NewEvidenceOwnerRuntimeForTimeline(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	projectionRuntime *projectionassembly.Runtime,
) *evidence.OwnerRuntime {
	conflictFields, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		panic(err)
	}
	return NewEvidenceOwnerRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		UnavailableEvidenceObjectStore(),
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime,
	)
}

// NewEvidenceMutationOwner composes the Evidence mutation owner for tests that
// exercise Evidence semantics without depending on Workbook's legacy Store.
func NewEvidenceMutationOwner(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) evidence.MutationContribution {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	runtime := NewEvidenceOwnerRuntime(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		intents,
		UnavailableEvidenceObjectStore(),
		revisionRuntime.ConflictFieldResolver(),
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime,
	)
	return runtime.MutationContribution()
}

func UnavailableEvidenceObjectStore() objectstore.TypedStore {
	return unavailableEvidenceObjectStore{}
}

// NewTaskDecisionOwner composes the source-owner mutation facade for tests
// that verify Task/Decision lifecycle behavior without the legacy Store.
func NewTaskDecisionOwner(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *tasksdecisions.MutationFacade {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
	if err != nil {
		panic(err)
	}
	owner, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		mustBuildProjectionRuntime(pool).TaskDecisionPorts().Rows,
	)
	if err != nil {
		panic(err)
	}
	return owner
}

// NewWorkbookCatalog composes the same immutable provider catalog used by the
// server for focused generic coordination tests.
func NewWorkbookCatalog(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *workbook.WorkbookContributionCatalog {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
	if err != nil {
		panic(err)
	}
	appender := revisionRuntime.Appender()
	conflictFields := revisionRuntime.ConflictFieldResolver()
	projectionRuntime := mustBuildProjectionRuntime(pool)
	evidenceOwner := NewEvidenceOwnerRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		UnavailableEvidenceObjectStore(),
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
	artifactMutation, err := workbookassembly.NewArtifactMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projectionRuntime.ArtifactPorts().Rows,
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
		projectionRuntime.PartyPorts().Rows,
		indicatorOwner,
		timelineBundle.Facade,
		evidenceOwner.MutationContribution(),
		artifactMutation,
		taskDecisionMutation,
		conflictTokens,
		conflictFields,
		appender,
		intents,
	)
	if err != nil {
		panic(err)
	}
	return catalog
}

// NewAssessmentOwner composes the Assessment owner for tests that exercise
// source semantics directly rather than through Workbook's legacy Store.
func NewAssessmentOwner(pool postgres.DB) *assessments.Facade {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	entityStore := hostidentity.NewStore(
		pool,
		revisionRuntime.Appender(),
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime.EntityPorts().Writer,
		hostidentity.WithProjectionReader(projectionRuntime.EntityPorts().Reader),
	)
	owner, err := workbookassembly.NewAssessmentMutationContribution(
		pool,
		projectionRuntime.AssessmentPorts().Rows,
		entityStore,
		revisionRuntime.Appender(),
	)
	if err != nil {
		panic(err)
	}
	return owner
}

// NewPartyOwner composes the Party mutation owner for tests that exercise
// Party semantics without depending on Workbook's transitional Store.
func NewPartyOwner(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) *parties.MutationFacade {
	intents := collaboration.NewIntentAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		panic(err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		contributions...,
	)
	if err != nil {
		panic(err)
	}
	projectionRuntime := mustBuildProjectionRuntime(pool)
	return parties.NewMutationFacade(
		pool,
		conflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		workbookassembly.NewConflictIdempotencyPort(pool),
		projectionRuntime.PartyPorts().Rows,
	)
}

type unavailableEvidenceObjectStore struct{}

type evidenceCleanupObserver struct{}

func (evidenceCleanupObserver) ObserveCleanupSweep(context.Context, evidence.CleanupSweepObservation) {
}

func (unavailableEvidenceObjectStore) UploadTarget(context.Context, string, time.Time) (objectstore.UploadTarget, error) {
	return objectstore.UploadTarget{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) CompleteUploadTarget(context.Context, string, io.Reader, string) error {
	return errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) PutObject(context.Context, string, io.Reader, int64, string) error {
	return errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) ReadObject(context.Context, string, objectstore.ReadOptions) (io.ReadCloser, objectstore.ObjectInfo, error) {
	return nil, objectstore.ObjectInfo{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) StatObject(context.Context, string) (objectstore.ObjectInfo, error) {
	return objectstore.ObjectInfo{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) ListObjects(context.Context, string) ([]objectstore.ObjectInfo, error) {
	return nil, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) DeleteObject(context.Context, string) error {
	return errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) CreateUploadTarget(context.Context, objectstore.UploadTargetRequest) (objectstore.UploadTarget, error) {
	return objectstore.UploadTarget{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) Put(context.Context, objectstore.PutObjectRequest) (objectstore.PutObjectResult, error) {
	return objectstore.PutObjectResult{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) Head(context.Context, objectstore.HeadObjectRequest) (objectstore.ObjectInfo, error) {
	return objectstore.ObjectInfo{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) Get(context.Context, objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error) {
	return nil, objectstore.ObjectInfo{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) ListPrefix(context.Context, objectstore.ListPrefixRequest) (objectstore.ListPrefixResult, error) {
	return objectstore.ListPrefixResult{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) Delete(context.Context, objectstore.DeleteObjectRequest) error {
	return errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) EnsureBucketForDevTest(context.Context, objectstore.EnsureBucketRequest) (objectstore.EnsureBucketResult, error) {
	return objectstore.EnsureBucketResult{}, errUnavailableEvidenceObjectStore
}

func (unavailableEvidenceObjectStore) Close() error { return nil }

func mustBuildProjectionRuntime(pool postgres.DB) *projectionassembly.Runtime {
	runtime, err := projectionassembly.Build(pool)
	if err != nil {
		panic(err)
	}
	return runtime
}
