package imports

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type JobSuccessMutation func(context.Context, pgx.Tx) error

type JobSuccessFinalization struct {
	Execution     jobs.Execution
	Completion    jobs.SuccessCompletion
	FinalCommitID string
	Mutate        JobSuccessMutation
}

// JobSuccessFinalizer is the Import-owned terminal-success port. Application
// assembly adapts the shared proof store without exposing that store to Import.
type JobSuccessFinalizer interface {
	FinalizeImportJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	FinalizeImportJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error)
	FinalizeImportJobCancellation(context.Context, JobCancellationFinalization) (jobs.Resource, error)
}

type JobFailureFinalization struct {
	Execution  jobs.Execution
	Completion jobs.FailureCompletion
	Mutate     JobSuccessMutation
}

type JobCancellationFinalization struct {
	Execution  jobs.Execution
	Completion jobs.CancellationCompletion
	Mutate     JobSuccessMutation
}
