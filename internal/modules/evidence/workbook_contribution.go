package evidence

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// WorkbookContribution is the complete Evidence-owned mutation capability
// consumed by generic Workbook dispatch. It exposes no Store or database
// handle and keeps Evidence request/result/error types owner-local.
type WorkbookContribution interface {
	Create(context.Context, WorkbookCreateCommand) (WorkbookMutationResult, error)
	Patch(context.Context, WorkbookPatchCommand) (WorkbookMutationResult, error)
	ResolveConflict(context.Context, WorkbookConflictCommand) (WorkbookMutationResult, error)
}

func NewWorkbookContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
) WorkbookContribution {
	return NewWorkbookFacade(pool, conflictTokens, appender, intents, conflictFields, keepSaved)
}

var _ WorkbookContribution = (*WorkbookFacade)(nil)
