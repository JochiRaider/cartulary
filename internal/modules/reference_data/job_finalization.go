package reference_data

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

// JobSuccessFinalizer is the Reference Pack-owned terminal-success port.
// Application assembly adapts the shared proof store without exposing that
// store or Extensions-owned persistence to the profile implementation.
type JobSuccessFinalizer interface {
	FinalizeReferencePackJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
}
