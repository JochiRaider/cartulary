package evidence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
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
}

// OwnerRuntime is the immutable application-facing Evidence capability set.
// It intentionally exposes no Store or database handle.
type OwnerRuntime struct {
	routes      RouteService
	workbook    WorkbookContribution
	attachments TimelineAttachmentContribution
}

func NewOwnerRuntime(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	objects objectstore.Store,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRows evidenceprojection.Rows,
	options ...StoreOption,
) *OwnerRuntime {
	if appender == nil {
		panic("compose Evidence owner runtime: Revisions appender is required")
	}
	if intents == nil {
		panic("compose Evidence owner runtime: Collaboration intent appender is required")
	}
	if projectionRows == nil {
		panic("compose Evidence owner runtime: projection rows are required")
	}
	storeOptions := append([]StoreOption{
		WithRevisionAppender(appender),
		WithCollaborationIntents(intents),
		WithWorkbookProjections(projectionRows),
	}, options...)
	store := NewStore(pool, storeOptions...)
	workbook := newWorkbookFacade(pool, conflictTokens, appender, intents, store, objects, conflictFields, keepSaved, projectionRows)
	return &OwnerRuntime{
		routes:      store,
		workbook:    workbook,
		attachments: timelineAttachmentReader{},
	}
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

// NewTimelineAttachmentContribution is an explicit narrow composition helper
// for focused tests that do not construct the Server runtime.
func NewTimelineAttachmentContribution(_ postgres.DB) TimelineAttachmentContribution {
	return timelineAttachmentReader{}
}
