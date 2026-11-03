package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestReportingRenderDeadlineOwnsMaterializationAndPublication(t *testing.T) {
	t.Run("model admission honors inherited deadline", func(t *testing.T) {
		store := &fakeReportingJobStore{}
		payload, err := store.ReleasePayloadForJob(context.Background(), jobs.Execution{})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()
		_, err = admitFrozenRenderInput(ctx, payload.renderInput)
		failure := renderFailureDetails(err)
		if err == nil || failure.FailureCode != "render_timeout" || failure.ReasonCode != "render_timeout" {
			t.Fatal("inherited timeout lost owner classification", err)
		}
	})

	for _, kind := range []string{reportingJobKindReleaseCreate, reportingJobKindCompositionPreview} {
		t.Run(kind+"/renderer observes total attempt deadline", func(t *testing.T) {
			manager := &fakeReportingJobManager{status: jobs.StatusRunning}
			store := &fakeReportingJobStore{kind: kind, declaredLimits: map[string]json.RawMessage{"timeout.total_attempt": json.RawMessage(`1`)}}
			rendered := false
			worker := newFakeReportingJobWorker(store, manager, releaseRendererFunc(func(ctx context.Context, _ renderInput) (RenderedRelease, string, error) {
				rendered = true
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > time.Second {
					t.Fatal("renderer escaped total attempt deadline")
				}
				<-ctx.Done()
				return RenderedRelease{}, "", ctx.Err()
			}))
			worker.timeout = 5 * time.Second

			if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
				t.Fatal(err)
			}
			if !rendered || manager.failed == nil || !manager.renderTimedOut || manager.failed.ErrorSummary.Retryable || manager.failed.ErrorSummary.Details["reason_code"] != "render_timeout" || store.completePreviewCalls != 0 || store.completeReleaseCalls != 0 {
				t.Fatal("render timeout published output or lost its terminal identity")
			}
		})
		for _, outcome := range []string{"proven absent", "proven committed", "indeterminate"} {
			t.Run(kind+"/publication/"+outcome, func(t *testing.T) {
				manager := &fakeReportingJobManager{status: jobs.StatusRunning}
				store := &fakeReportingJobStore{kind: kind}
				worker := newFakeReportingJobWorker(store, manager, nil)
				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()
				deadline, _ := ctx.Deadline()
				uncertain := errors.New("commit outcome remains indeterminate")
				finalizer := &deadlineReportingFinalizer{fakeReportingJobFinalizer: fakeReportingJobFinalizer{manager: manager}}
				finalizer.success = func(work context.Context, request JobSuccessFinalization) (jobs.Resource, error) {
					if actual, ok := work.Deadline(); !ok || !actual.Equal(deadline) {
						t.Fatal("publication replaced the original work deadline")
					}
					<-work.Done()
					switch outcome {
					case "proven committed":
						// Simulate the adapter's independent commit proof resolving
						// an acknowledgement lost when the work deadline fired.
						return manager.CompleteSucceeded(context.Background(), request.Execution, request.Completion)
					case "indeterminate":
						return jobs.Resource{}, uncertain
					default:
						return jobs.Resource{}, work.Err()
					}
				}
				worker.jobFinalizer = finalizer
				err := worker.execute(ctx, context.Background(), jobs.Execution{})
				if outcome == "indeterminate" {
					if !errors.Is(err, uncertain) || manager.status != jobs.StatusRunning || manager.failed != nil || store.renderFailedReleaseCalls != 0 {
						t.Fatal("uncertain publication invented a terminal failure", err)
					}
				} else if outcome == "proven committed" {
					if err != nil || manager.status != jobs.StatusSucceeded || manager.failed != nil || store.renderFailedReleaseCalls != 0 {
						t.Fatal("committed publication was replaced by timeout", err)
					}
				} else {
					wantFailedRelease := 0
					if kind == reportingJobKindReleaseCreate {
						wantFailedRelease = 1
					}
					if err != nil || manager.status != jobs.StatusFailed || !manager.renderTimedOut || manager.failed == nil || manager.failed.ErrorSummary.Details["reason_code"] != "render_timeout" || store.renderFailedReleaseCalls != wantFailedRelease {
						t.Fatal("proven-uncommitted deadline did not retain timeout", err)
					}
				}
				if finalizer.successCalls != 1 || store.completeReleaseCalls != 0 || store.completePreviewCalls != 0 {
					t.Fatal("publication was retried or output escaped the test transaction")
				}
			})
		}
	}
}

type deadlineReportingFinalizer struct {
	fakeReportingJobFinalizer
	success      func(context.Context, JobSuccessFinalization) (jobs.Resource, error)
	successCalls int
}

func (f *deadlineReportingFinalizer) FinalizeReportingJobSuccess(ctx context.Context, request JobSuccessFinalization) (jobs.Resource, error) {
	f.successCalls++
	return f.success(ctx, request)
}
