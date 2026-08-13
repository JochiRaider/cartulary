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
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
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

func NewEvidenceBlobLifecycleService(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) *evidence.BlobLifecycleService {
	projectionRuntime := mustBuildProjectionRuntime(pool)
	service, err := evidence.NewBlobLifecycleService(evidence.BlobLifecycleDependencies{
		Postgres:       pool,
		Revisions:      appender,
		Projections:    projectionRuntime.EvidencePorts().Rows,
		SupportEffects: projectionRuntime.EvidencePorts().SupportEffects,
		Collaboration:  intents,
	})
	if err != nil {
		panic(err)
	}
	return service
}

func NewEvidenceRouteOperations(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) *evidence.RouteOperations {
	blobs := NewEvidenceBlobLifecycleService(pool, appender, intents)
	access, err := evidence.NewAccessHandleService(pool)
	if err != nil {
		panic(err)
	}
	operations, err := evidence.NewRouteOperations(blobs, access)
	if err != nil {
		panic(err)
	}
	return operations
}

func NewEvidenceOwnerRuntime(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	objects objectstore.Store,
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
	})
	if err != nil {
		panic(err)
	}
	return runtime
}

func UnavailableEvidenceObjectStore() objectstore.Store {
	return unavailableEvidenceObjectStore{}
}

// NewWorkbookStore composes the same code-backed projection catalog used by
// the server for focused module tests that do not need an HTTP runtime.
func NewWorkbookStore(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *workbook.Store {
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
	evidenceAttachments := evidence.NewTimelineAttachmentContribution(projectionRuntime.EvidencePorts().Rows)
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
		evidenceOwner.WorkbookContribution(),
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
	store, err := workbookassembly.NewMutationStore(
		pool,
		catalog,
		artifactMutation,
		taskDecisionMutation,
	)
	if err != nil {
		panic(err)
	}
	return store
}

type unavailableEvidenceObjectStore struct{}

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

func (unavailableEvidenceObjectStore) Close() error { return nil }

func mustBuildProjectionRuntime(pool postgres.DB) *projectionassembly.Runtime {
	runtime, err := projectionassembly.Build(pool)
	if err != nil {
		panic(err)
	}
	return runtime
}
