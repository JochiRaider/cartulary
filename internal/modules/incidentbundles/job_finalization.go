package incidentbundles

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrJobFinalizationIndeterminate = errors.New("incident bundle job finalization is indeterminate")

type JobSuccessMutation func(context.Context, pgx.Tx) error

type JobSuccessFinalization struct {
	Transition    jobs.TransitionParams
	FinalCommitID string
	Mutate        JobSuccessMutation
}

// JobSuccessFinalizer is the Incident Bundles-owned terminal-success port.
// Application assembly adapts the shared proof store while keeping physical
// transaction access out of the module.
type JobSuccessFinalizer interface {
	FinalizeIncidentBundleJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	FinalizeIncidentBundleJobSuccessTx(context.Context, crossownertransaction.FinalizationCapability, JobSuccessFinalization) (jobs.Resource, error)
}

type importJobTransactionFinalizer struct {
	finalizer      JobSuccessFinalizer
	jobID          uuid.UUID
	manifestSHA256 string
}

func (f importJobTransactionFinalizer) Publish(
	ctx context.Context,
	capability crossownertransaction.FinalizationCapability,
	values map[string]any,
) error {
	if f.finalizer == nil || f.jobID == uuid.Nil || f.manifestSHA256 == "" ||
		len(values) == 0 {
		return crossownertransaction.ErrWrite
	}
	result, ok := values[ImportTransactionParticipantID].(ImportTransactionResult)
	if !ok || result.IncidentID == uuid.Nil {
		return crossownertransaction.ErrWrite
	}
	_, err := f.finalizer.FinalizeIncidentBundleJobSuccessTx(
		ctx,
		capability,
		JobSuccessFinalization{
			Transition: importSuccessTransition(f.jobID, result.IncidentID),
			FinalCommitID: "incident_portability.import:" +
				result.IncidentID.String() + ":" + f.manifestSHA256,
		},
	)
	return err
}
