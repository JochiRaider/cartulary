package reporting

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// JobSuccessMutation is the Reporting-owned authoritative state mutation that
// application composition commits with terminal job state and its immutable
// extension proof.
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
	TimedOut   bool
}

type JobCancellationFinalization struct {
	Execution  jobs.Execution
	Completion jobs.CancellationCompletion
	Mutate     JobSuccessMutation
}

// JobSuccessFinalizer is Reporting's narrow port to the shared extension job
// finalization transaction.
// Success publication uses the supplied work deadline. A context error may be
// returned only when no commit occurred. A proven commit returns success even
// after the deadline; an indeterminate commit returns an opaque error that does
// not wrap a context error, leaving reconciliation to the shared Jobs owner.
type JobSuccessFinalizer interface {
	FinalizeReportingJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	FinalizeReportingJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error)
	FinalizeReportingJobCancellation(context.Context, JobCancellationFinalization) (jobs.Resource, error)
}
