package evidence

import (
	"context"
)

// MutationContribution is the complete Evidence-owned mutation capability
// consumed by generic Workbook owner dispatch. It exposes no database
// handle and keeps Evidence request/result/error types owner-local.
type MutationContribution interface {
	Create(context.Context, CreateCommand) (MutationResult, error)
	Patch(context.Context, PatchCommand) (MutationResult, error)
	ResolveConflict(context.Context, ConflictCommand) (MutationResult, error)
}

var _ MutationContribution = (*mutationFacade)(nil)
