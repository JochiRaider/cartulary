package evidence

import (
	"context"
)

// WorkbookContribution is the complete Evidence-owned mutation capability
// consumed by generic Workbook dispatch. It exposes no Store or database
// handle and keeps Evidence request/result/error types owner-local.
type WorkbookContribution interface {
	Create(context.Context, WorkbookCreateCommand) (WorkbookMutationResult, error)
	Patch(context.Context, WorkbookPatchCommand) (WorkbookMutationResult, error)
	ResolveConflict(context.Context, WorkbookConflictCommand) (WorkbookMutationResult, error)
}

var _ WorkbookContribution = (*WorkbookFacade)(nil)
