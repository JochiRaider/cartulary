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
	if !ok || finalization.tx == nil {
		return jobs.Resource{}, fmt.Errorf(
			"%w: incident bundle finalization capability",
			crossownertransaction.ErrWrite,
		)
	}
	resource, err := adapter.finalizer.FinalizeSuccessTx(
		ctx,
		finalization.tx,
		extensionstore.JobFinalizationRequest{
			Execution:     request.Execution,
			Completion:    request.Completion,
			FinalCommitID: request.FinalCommitID,
			Mutate:        extensionstore.OwnerMutation(request.Mutate),
		},
		adapter.now().UTC(),
	)
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

func mapIncidentBundleFinalizationError(err error) error {
	if errors.Is(err, extensionstore.ErrIndeterminateCommit) {
		return fmt.Errorf("%w: %v", incidentbundles.ErrJobFinalizationIndeterminate, err)
	}
	return err
}
