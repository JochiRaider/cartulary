package networkflow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestGraphViewMaterializationDeadlineIsTerminalBeforePublication_Unit(t *testing.T) {
	manager := &deadlineGraphViewJobManager{}
	finalizer := &deadlineGraphViewFinalizer{}
	module := &Module{
		store: &Store{}, limits: EffectiveLimits{GraphMaterializationTimeoutSeconds: 0}, now: time.Now,
		graphProjection: newGraphProjectionAdapter(), jobManager: manager, jobFinalizer: finalizer,
	}
	started := time.Now()
	if err := module.handleGraphViewMaterialization(context.Background(), jobs.Execution{}); err != nil {
		t.Fatalf("terminalize materialization timeout: %v", err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("materialization deadline did not start at handler invocation")
	}
	if finalizer.successCalled || !finalizer.failureCalled {
		t.Fatalf("deadline finalization success=%v failure=%v", finalizer.successCalled, finalizer.failureCalled)
	}
	summary := finalizer.failure.Completion.ErrorSummary
	if summary.Code != "network_flow_graph_materialization_failed" || summary.Retryable || summary.Details["reason_code"] != "timeout" {
		t.Fatalf("materialization timeout summary = %#v", summary)
	}
	if graphViewMaterializationFailureCode("timeout") != "network_flow_graph_materialization_timeout" {
		t.Fatal("materialization timeout did not map to the closed safe declaration failure code")
	}
}

type deadlineGraphViewJobManager struct{}

func (*deadlineGraphViewJobManager) Get(context.Context, uuid.UUID) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*deadlineGraphViewJobManager) ObserveExecution(ctx context.Context, _ jobs.Execution) (jobs.Resource, error) {
	<-ctx.Done()
	return jobs.Resource{}, ctx.Err()
}

func (*deadlineGraphViewJobManager) HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error) {
	return nil, nil
}

func (*deadlineGraphViewJobManager) RetainedHandlerPayload(context.Context, uuid.UUID) (json.RawMessage, error) {
	return nil, nil
}

func (*deadlineGraphViewJobManager) CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

type deadlineGraphViewFinalizer struct {
	successCalled bool
	failureCalled bool
	failure       GraphViewJobFailureFinalization
}

func (f *deadlineGraphViewFinalizer) FinalizeGraphViewJobSuccess(context.Context, GraphViewJobSuccessFinalization) (jobs.Resource, error) {
	f.successCalled = true
	return jobs.Resource{}, nil
}

func (f *deadlineGraphViewFinalizer) FinalizeGraphViewJobFailure(_ context.Context, finalization GraphViewJobFailureFinalization) (jobs.Resource, error) {
	f.failureCalled = true
	f.failure = finalization
	return jobs.Resource{}, nil
}
