package extensionassembly

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type importJobSuccessFinalizer struct {
	finalizer *extensionstore.OwnerFinalizer
}

func NewImportJobSuccessFinalizer(finalizer *extensionstore.OwnerFinalizer) imports.JobSuccessFinalizer {
	if finalizer == nil {
		return nil
	}
	return importJobSuccessFinalizer{finalizer: finalizer}
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobSuccess(
	ctx context.Context,
	request imports.JobSuccessFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Transition:    request.Transition,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}
