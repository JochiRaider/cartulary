package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestReportingWorkerInterruptsActiveRenderer(t *testing.T) {
	for _, kind := range []string{reportingJobKindSnapshotCreate, reportingJobKindReleaseCreate, reportingJobKindCompositionPreview} {
		t.Run(kind+" preparation drains before cancellation", func(t *testing.T) {
			started, drained := make(chan struct{}), make(chan struct{})
			ticks := make(chan time.Time)
			manager := &fakeReportingJobManager{status: jobs.StatusRunning, observe: func() error {
				select {
				case <-started:
					return jobs.ErrCancellationRequested
				default:
					return nil
				}
			}}
			store := &fakeReportingJobStore{kind: kind, prepare: func(ctx context.Context) error { close(started); <-ctx.Done(); close(drained); return ctx.Err() }}
			worker := newFakeReportingJobWorker(store, manager, releaseRendererFunc(func(context.Context, renderInput) (RenderedRelease, string, error) {
				return RenderedRelease{}, "", errors.New("renderer must not start")
			}))
			worker.renderObservationTicks = func() (<-chan time.Time, func()) { return ticks, func() {} }
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- worker.Handle(ctx, jobs.Execution{}) }()
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("preparation did not start")
			}
			select {
			case ticks <- time.Time{}:
			case <-ctx.Done():
				t.Fatal("preparation was not observed")
			}
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("preparation did not drain")
			}
			select {
			case <-drained:
			default:
				t.Fatal("terminal transition preceded preparation drain")
			}
			if manager.status != jobs.StatusCanceled || store.completeSnapshotCalls != 0 || store.completeReleaseCalls != 0 || store.completePreviewCalls != 0 {
				t.Fatal("preparation cancellation published output")
			}
		})
	}

	observationFailure := errors.New("synthetic observation failure")
	for _, test := range []struct {
		name, status     string
		observeErr, want error
	}{
		{name: "user cancellation", status: jobs.StatusCancelRequested, want: jobs.ErrCancellationRequested},
		{name: "lost execution", status: jobs.StatusRunning, observeErr: jobs.ErrExecutionLost, want: jobs.ErrExecutionLost},
		{name: "observation unavailable", status: jobs.StatusRunning, observeErr: observationFailure, want: observationFailure},
		{name: "terminal execution", status: jobs.StatusCanceled, want: jobs.ErrExecutionLost},
	} {
		t.Run(test.name, func(t *testing.T) {
			started, drained := make(chan struct{}), make(chan struct{})
			ticks := make(chan time.Time)
			manager := &fakeReportingJobManager{status: test.status, observeErr: test.observeErr}
			worker := newFakeReportingJobWorker(&fakeReportingJobStore{}, manager, releaseRendererFunc(func(ctx context.Context, _ renderInput) (RenderedRelease, string, error) {
				close(started)
				<-ctx.Done()
				close(drained)
				// Even a renderer returning success after cancellation cannot publish.
				return RenderedRelease{ProfileSHA256: "must-not-publish"}, "", nil
			}))
			worker.renderObservationTicks = func() (<-chan time.Time, func()) { return ticks, func() {} }
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				rendered, err := worker.renderWhileCurrent(ctx, jobs.Execution{}, renderInput{})
				if rendered.rendered.ProfileSHA256 != "" {
					result <- errors.New("canceled output escaped")
					return
				}
				result <- err
			}()
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("renderer did not start")
			}
			select {
			case ticks <- time.Time{}:
			case <-ctx.Done():
				t.Fatal("supervisor did not observe")
			}
			select {
			case err := <-result:
				if !errors.Is(err, test.want) {
					t.Fatalf("got %v want %v", err, test.want)
				}
			case <-ctx.Done():
				t.Fatal("supervisor failed to drain renderer")
			}
			select {
			case <-drained:
			default:
				t.Fatal("returned before renderer cleanup")
			}
		})
	}
	t.Run("panic has no output or raw diagnostic", func(t *testing.T) {
		worker := newFakeReportingJobWorker(&fakeReportingJobStore{}, &fakeReportingJobManager{status: jobs.StatusRunning}, releaseRendererFunc(func(context.Context, renderInput) (RenderedRelease, string, error) {
			panic("private renderer diagnostic")
		}))
		result, err := worker.renderWhileCurrent(context.Background(), jobs.Execution{}, renderInput{})
		if !errors.Is(err, errCandidateRendererPanic) || result.err != nil || result.reason != "" || result.rendered.ProfileSHA256 != "" {
			t.Fatalf("unexpected panic result: %#v, %v", result, err)
		}
	})
}
