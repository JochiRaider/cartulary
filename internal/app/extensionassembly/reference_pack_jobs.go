package extensionassembly

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type referencePackJobSuccessFinalizer struct {
	finalizer *extensionstore.OwnerFinalizer
}

func NewReferencePackJobSuccessFinalizer(finalizer *extensionstore.OwnerFinalizer) reference_data.JobSuccessFinalizer {
	if finalizer == nil {
		return nil
	}
	return referencePackJobSuccessFinalizer{finalizer: finalizer}
}

func (adapter referencePackJobSuccessFinalizer) FinalizeReferencePackJobSuccess(
	ctx context.Context,
	request reference_data.JobSuccessFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Transition:    request.Transition,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}
