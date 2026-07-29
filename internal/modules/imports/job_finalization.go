package imports

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type JobSuccessMutation func(context.Context, pgx.Tx) error

type JobSuccessFinalization struct {
	Transition    jobs.TransitionParams
	FinalCommitID string
	Mutate        JobSuccessMutation
}

// JobSuccessFinalizer is the Import-owned terminal-success port. Application
// assembly adapts the shared proof store without exposing that store to Import.
type JobSuccessFinalizer interface {
	FinalizeImportJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	FinalizeImportJobFailure(context.Context, JobTerminalFinalization) (jobs.Resource, error)
	FinalizeImportJobCancellation(context.Context, JobTerminalFinalization) (jobs.Resource, error)
}

type JobTerminalFinalization struct {
	Transition jobs.TransitionParams
	Mutate     JobSuccessMutation
}
