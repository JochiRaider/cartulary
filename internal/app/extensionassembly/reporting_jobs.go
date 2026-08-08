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
		Execution:     request.Execution,
		Completion:    request.Completion,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}

func (adapter reportingJobSuccessFinalizer) FinalizeReportingJobFailure(
	ctx context.Context,
	request reporting.JobFailureFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeFailure(ctx, extensionstore.JobFailureFinalizationRequest{
		Execution:  request.Execution,
		Completion: request.Completion,
		Mutate:     extensionstore.OwnerMutation(request.Mutate),
	})
}
