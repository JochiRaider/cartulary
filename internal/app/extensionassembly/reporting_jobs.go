package extensionassembly

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type reportingJobSuccessFinalizer struct {
	finalizer *extensionstore.OwnerFinalizer
}

func NewReportingJobSuccessFinalizer(finalizer *extensionstore.OwnerFinalizer) reporting.JobSuccessFinalizer {
	if finalizer == nil {
		return nil
	}
	return reportingJobSuccessFinalizer{finalizer: finalizer}
}

func (adapter reportingJobSuccessFinalizer) FinalizeReportingJobSuccess(
	ctx context.Context,
	request reporting.JobSuccessFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Transition:    request.Transition,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}
