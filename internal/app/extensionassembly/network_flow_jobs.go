package extensionassembly

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type networkFlowGraphViewJobFinalizer struct {
	finalizer *extensionstore.OwnerFinalizer
}

func NewNetworkFlowGraphViewJobFinalizer(finalizer *extensionstore.OwnerFinalizer) networkflow.GraphViewJobFinalizer {
	if finalizer == nil {
		return nil
	}
	return networkFlowGraphViewJobFinalizer{finalizer: finalizer}
}

func (adapter networkFlowGraphViewJobFinalizer) FinalizeGraphViewJobSuccess(
	ctx context.Context,
	request networkflow.GraphViewJobSuccessFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Execution: request.Execution, Completion: request.Completion,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}

func (adapter networkFlowGraphViewJobFinalizer) FinalizeGraphViewJobFailure(
	ctx context.Context,
	request networkflow.GraphViewJobFailureFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeFailure(ctx, extensionstore.JobFailureFinalizationRequest{
		Execution: request.Execution, Completion: request.Completion,
		Mutate: extensionstore.OwnerMutation(request.Mutate),
	})
}
