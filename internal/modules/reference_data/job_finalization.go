package reference_data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrJobFinalizationIndeterminate = errors.New("reference pack job finalization is indeterminate")

type JobSuccessMutation func(context.Context, pgx.Tx) error

type JobSuccessFinalization struct {
	Execution     jobs.Execution
	Completion    jobs.SuccessCompletion
	FinalCommitID string
	Mutate        JobSuccessMutation
}

type JobFailureFinalization struct {
	Execution  jobs.Execution
	Completion jobs.FailureCompletion
	Mutate     JobSuccessMutation
}

// JobSuccessFinalizer is the Reference Pack-owned terminal-success port.
// Application assembly adapts the shared proof store without exposing that
// store or Extensions-owned persistence to the profile implementation.
type JobSuccessFinalizer interface {
	FinalizeReferencePackJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	FinalizeReferencePackJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error)
}
