package extensionassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type incidentBundleJobSuccessFinalizer struct {
	finalizer *extensionstore.OwnerFinalizer
	now       func() time.Time
}

func NewIncidentBundleJobSuccessFinalizer(
	finalizer *extensionstore.OwnerFinalizer,
	now func() time.Time,
) incidentbundles.JobSuccessFinalizer {
	if finalizer == nil {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return incidentBundleJobSuccessFinalizer{finalizer: finalizer, now: now}
}

func (adapter incidentBundleJobSuccessFinalizer) FinalizeIncidentBundleJobSuccess(
	ctx context.Context,
	request incidentbundles.JobSuccessFinalization,
) (jobs.Resource, error) {
	resource, err := adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Execution:     request.Execution,
		Completion:    request.Completion,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
	return resource, mapIncidentBundleFinalizationError(err)
}

func (adapter incidentBundleJobSuccessFinalizer) FinalizeIncidentBundleJobSuccessTx(
	ctx context.Context,
	capability crossownertransaction.FinalizationCapability,
	request incidentbundles.JobSuccessFinalization,
) (jobs.Resource, error) {
	finalization, ok := capability.(crossOwnerFinalization)
	if !ok || finalization.transaction == nil || finalization.transaction.tx == nil || finalization.transaction.closed || finalization.transaction.commitFinalized != nil {
		return jobs.Resource{}, fmt.Errorf(
			"%w: incident bundle finalization capability",
			crossownertransaction.ErrWrite,
		)
	}
	finalRequest := extensionstore.JobFinalizationRequest{
		Execution: request.Execution, Completion: request.Completion,
		FinalCommitID: request.FinalCommitID, Mutate: extensionstore.OwnerMutation(request.Mutate),
	}
	resource, err := adapter.finalizer.FinalizeSuccessTx(
		ctx,
		finalization.transaction.tx,
		finalRequest,
		adapter.now().UTC(),
	)
	if err == nil {
		finalization.transaction.commitFinalized = func(commitCtx context.Context) (extensionstore.CommitOutcome, error) {
			return adapter.finalizer.CommitSuccessTx(commitCtx, finalization.transaction.tx, finalRequest, resource)
		}
	}
	return resource, mapIncidentBundleFinalizationError(err)
}

func (adapter incidentBundleJobSuccessFinalizer) FinalizeIncidentBundleJobFailure(
	ctx context.Context,
	request incidentbundles.JobFailureFinalization,
) (jobs.Resource, error) {
	resource, err := adapter.finalizer.FinalizeFailure(ctx, extensionstore.JobFailureFinalizationRequest{
		Execution:  request.Execution,
		Completion: request.Completion,
		Mutate:     extensionstore.OwnerMutation(request.Mutate),
	})
	return resource, mapIncidentBundleFinalizationError(err)
}

func (adapter incidentBundleJobSuccessFinalizer) FinalizeIncidentBundleJobTimeout(ctx context.Context, request incidentbundles.JobFailureFinalization) (jobs.Resource, error) {
	resource, err := adapter.finalizer.FinalizeTimeout(ctx, extensionstore.JobFailureFinalizationRequest{Execution: request.Execution, Completion: request.Completion, Mutate: extensionstore.OwnerMutation(request.Mutate)})
	return resource, mapIncidentBundleFinalizationError(err)
}
func (adapter incidentBundleJobSuccessFinalizer) FinalizeIncidentBundleJobCancellation(ctx context.Context, request incidentbundles.JobCancellationFinalization) (jobs.Resource, error) {
	resource, err := adapter.finalizer.FinalizeCancellation(ctx, extensionstore.JobCancellationFinalizationRequest{Execution: request.Execution, Completion: request.Completion, Mutate: extensionstore.OwnerMutation(request.Mutate)})
	return resource, mapIncidentBundleFinalizationError(err)
}

func mapIncidentBundleFinalizationError(err error) error {
	if errors.Is(err, extensionstore.ErrIndeterminateCommit) {
		return fmt.Errorf("%w: %v", incidentbundles.ErrJobFinalizationIndeterminate, err)
	}
	return err
}
