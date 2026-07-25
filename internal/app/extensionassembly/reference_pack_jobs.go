package extensionassembly

import (
	"context"
	"errors"
	"fmt"

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
	resource, err := adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Transition:    request.Transition,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
	if errors.Is(err, extensionstore.ErrIndeterminateCommit) {
		return resource, fmt.Errorf("%w: %v", reference_data.ErrJobFinalizationIndeterminate, err)
	}
	return resource, err
}
