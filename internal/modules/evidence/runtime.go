package evidence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// TimelineAttachmentContribution is the narrow Evidence-owned capability used
// by Timeline to validate attached Evidence record identities in its caller
// transaction.
type TimelineAttachmentContribution interface {
	ValidateTimelineAttachmentsTx(
		ctx context.Context,
		tx pgx.Tx,
		incidentID uuid.UUID,
		recordIDs []uuid.UUID,
	) error
	RefreshTimelineAttachmentProjectionsTx(
		ctx context.Context,
		tx pgx.Tx,
		recordIDs []uuid.UUID,
	) error
}

// OwnerRuntime is the immutable application-facing Evidence capability set.
// It intentionally exposes no Store or database handle.
type OwnerRuntime struct {
	postgres     postgres.DB
	objectStore  objectstore.TypedStore
	now          func() time.Time
	routes       routeService
	workbook     MutationContribution
	attachments  TimelineAttachmentContribution
	importCreate ownerfacade.ImportOwnerCreateFacade
	cleanup      *CleanupDispatcher
}

type OwnerRuntimeDependencies struct {
	Postgres        postgres.DB
	ConflictTokens  *conflicttokens.ConflictTokenCodec
	Revisions       *revisions.Appender
	Collaboration   collaboration.RecordChangedAppender
	ObjectStore     objectstore.TypedStore
	Mutation        MutationDependencies
	CleanupObserver CleanupObserver
	Now             func() time.Time
}

func NewOwnerRuntime(dependencies OwnerRuntimeDependencies) (*OwnerRuntime, error) {
	switch {
	case dependencies.Postgres == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: Postgres is required")
	case dependencies.ConflictTokens == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: conflict-token codec is required")
	case dependencies.Revisions == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: Revisions appender is required")
	case dependencies.Collaboration == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: Collaboration intent appender is required")
	case dependencies.ObjectStore == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: object store is required")
	case dependencies.CleanupObserver == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: cleanup observer is required")
	case dependencies.Now == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: clock is required")
	}
	if err := dependencies.Mutation.validate(); err != nil {
		return nil, fmt.Errorf("compose Evidence owner runtime: %w", err)
	}
	sourceMutations, err := newSourceMutationService(
		dependencies.Postgres,
		dependencies.Mutation.ProjectionRows,
		dependencies.Revisions,
		dependencies.Collaboration,
		dependencies.Mutation.IncidentState,
		dependencies.Mutation.RecordEnvelopes,
	)
	if err != nil {
		return nil, err
	}
	blobs, err := newBlobLifecycleService(blobLifecycleDependencies{
		Postgres:        dependencies.Postgres,
		Revisions:       newRevisionAppendAdapter(dependencies.Revisions),
		Projections:     dependencies.Mutation.ProjectionRows,
		SupportEffects:  dependencies.Mutation.AssociationEffects,
		Collaboration:   dependencies.Collaboration,
		IncidentState:   dependencies.Mutation.IncidentState,
		RecordEnvelopes: dependencies.Mutation.RecordEnvelopes,
		Idempotency:     dependencies.Mutation.LifecycleIdempotency,
	})
	if err != nil {
		return nil, err
	}
	access, err := newAccessHandleService(dependencies.Postgres)
	if err != nil {
		return nil, err
	}
	routes, err := newRouteOperations(blobs, access)
	if err != nil {
		return nil, err
	}
	workbook, err := newMutationFacade(
		dependencies.Postgres,
		*dependencies.ConflictTokens,
		sourceMutations,
		dependencies.ObjectStore,
		dependencies.Mutation,
	)
	if err != nil {
		return nil, err
	}
	importCreate, err := newImportCreateFacade("evidence.import_create", sourceMutations)
	if err != nil {
		return nil, err
	}
	cleanupService, err := newCleanupService(dependencies.Postgres)
	if err != nil {
		return nil, err
	}
	cleanupDeleter, err := newCleanupObjectDeleter(dependencies.ObjectStore)
	if err != nil {
		return nil, err
	}
	cleanup, err := newCleanupDispatcher(
		cleanupService,
		cleanupDeleter,
		dependencies.CleanupObserver,
		dependencies.Now,
	)
	if err != nil {
		return nil, err
	}
	return &OwnerRuntime{
		postgres:     dependencies.Postgres,
		objectStore:  dependencies.ObjectStore,
		now:          dependencies.Now,
		routes:       routes,
		workbook:     workbook,
		attachments:  timelineAttachmentReader{projections: dependencies.Mutation.ProjectionRows},
		importCreate: importCreate,
		cleanup:      cleanup,
	}, nil
}

func (runtime *OwnerRuntime) RouteRegistrar(settings Settings) httpapi.RouteRegistrar {
	return registerRoutes(settings, routeDependencies{
		objectStore: runtime.objectStore,
		now:         runtime.now,
		operations:  runtime.routes,
		admission: routeAdmission{
			incidents: admission.NewChecker(runtime.postgres),
			auth:      authn.NewStore(runtime.postgres),
			now:       runtime.now,
		},
	})
}

func (runtime *OwnerRuntime) MutationContribution() MutationContribution {
	return runtime.workbook
}

func (runtime *OwnerRuntime) TimelineAttachmentContribution() TimelineAttachmentContribution {
	return runtime.attachments
}

func (runtime *OwnerRuntime) ImportCreateFacade() ownerfacade.ImportOwnerCreateFacade {
	return runtime.importCreate
}

func (runtime *OwnerRuntime) CleanupDispatcher() *CleanupDispatcher {
	return runtime.cleanup
}
