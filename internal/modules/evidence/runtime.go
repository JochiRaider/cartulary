package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
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
	routes       RouteService
	workbook     WorkbookContribution
	attachments  TimelineAttachmentContribution
	importCreate ownerfacade.ImportOwnerCreateFacade
}

type OwnerRuntimeDependencies struct {
	Postgres            postgres.DB
	ConflictTokens      *conflicttokens.ConflictTokenCodec
	Revisions           *revisions.Appender
	Collaboration       collaboration.IntentAppender
	ObjectStore         objectstore.Store
	ConflictFields      conflicttokens.FieldResolver
	ConflictIdempotency conflicttokens.IdempotencyPort
	Projections         evidenceprojection.Ports
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
	case dependencies.ConflictFields == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: conflict field resolver is required")
	case dependencies.ConflictIdempotency == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: conflict idempotency is required")
	case dependencies.Projections.Rows == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: projection rows are required")
	case dependencies.Projections.SupportEffects == nil:
		return nil, fmt.Errorf("compose Evidence owner runtime: support projection effects are required")
	}
	sourceMutations, err := newSourceMutationService(
		dependencies.Postgres,
		dependencies.Projections.Rows,
		dependencies.Revisions,
		dependencies.Collaboration,
	)
	if err != nil {
		return nil, err
	}
	blobs, err := NewBlobLifecycleService(BlobLifecycleDependencies{
		Postgres:       dependencies.Postgres,
		Revisions:      dependencies.Revisions,
		Projections:    dependencies.Projections.Rows,
		SupportEffects: dependencies.Projections.SupportEffects,
		Collaboration:  dependencies.Collaboration,
	})
	if err != nil {
		return nil, err
	}
	access, err := NewAccessHandleService(dependencies.Postgres)
	if err != nil {
		return nil, err
	}
	routes, err := NewRouteOperations(blobs, access)
	if err != nil {
		return nil, err
	}
	workbook, err := newWorkbookFacade(
		dependencies.Postgres,
		*dependencies.ConflictTokens,
		dependencies.Revisions,
		dependencies.Collaboration,
		sourceMutations,
		dependencies.ObjectStore,
		dependencies.ConflictFields,
		dependencies.ConflictIdempotency,
		dependencies.Projections.Rows,
		dependencies.Projections.SupportEffects,
	)
	if err != nil {
		return nil, err
	}
	importCreate, err := newImportCreateFacade("evidence.import_create", sourceMutations)
	if err != nil {
		return nil, err
	}
	return &OwnerRuntime{
		routes:       routes,
		workbook:     workbook,
		attachments:  timelineAttachmentReader{projections: dependencies.Projections.Rows},
		importCreate: importCreate,
	}, nil
}

func (runtime *OwnerRuntime) RouteService() RouteService {
	return runtime.routes
}

func (runtime *OwnerRuntime) WorkbookContribution() WorkbookContribution {
	return runtime.workbook
}

func (runtime *OwnerRuntime) TimelineAttachmentContribution() TimelineAttachmentContribution {
	return runtime.attachments
}

func (runtime *OwnerRuntime) ImportCreateFacade() ownerfacade.ImportOwnerCreateFacade {
	return runtime.importCreate
}

// NewTimelineAttachmentContribution is an explicit narrow composition helper
// for focused tests that do not construct the Server runtime.
func NewTimelineAttachmentContribution(projectionRows evidenceprojection.Rows) TimelineAttachmentContribution {
	return timelineAttachmentReader{projections: projectionRows}
}
