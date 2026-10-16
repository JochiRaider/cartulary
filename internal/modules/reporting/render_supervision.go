package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

type candidateRenderResult struct {
	rendered RenderedRelease
	reason   string
	err      error
}

var errCandidateRendererPanic = errors.New("reporting renderer panicked")

// renderWhileCurrent keeps execution fencing active throughout an external
// render. An observation failure is control-plane failure, not a render result.
// Cancellation always drains the renderer before returning to finalization.
func (w *reportingJobWorker) renderWhileCurrent(ctx context.Context, execution jobs.Execution, input renderInput) (candidateRenderResult, error) {
	var result candidateRenderResult
	workErr, controlErr := w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		result.rendered, result.reason, result.err = w.renderer.Render(work, input)
		return result.err
	})
	if controlErr != nil {
		return interruptedRenderResult(result), controlErr
	}
	if errors.Is(workErr, errCandidateRendererPanic) {
		return candidateRenderResult{}, workErr
	}
	return result, nil
}

// The operation drains before this returns, including on cancellation or loss
// of execution. No database transition lock is held during owned work.
func (w *reportingJobWorker) whileExecutionCurrent(ctx context.Context, execution jobs.Execution, work func(context.Context) error) (error, error) {
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			if recover() != nil {
				err = errCandidateRendererPanic
			}
			done <- err
		}()
		err = work(workCtx)
	}()
	newTicks := w.renderObservationTicks
	if newTicks == nil {
		newTicks = func() (<-chan time.Time, func()) {
			ticker := time.NewTicker(time.Second)
			return ticker.C, ticker.Stop
		}
	}
	ticks, stop := newTicks()
	defer stop()
	interrupt := func(err error) (error, error) { cancel(); return <-done, err }
	for {
		select {
		case result := <-done:
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if errors.Is(result, errCandidateRendererPanic) {
				return result, errCandidateRendererPanic
			}
			return result, nil
		case <-ctx.Done():
			return interrupt(ctx.Err())
		case <-ticks:
			resource, err := w.jobManager.ObserveExecution(ctx, execution)
			if err != nil {
				return interrupt(err)
			}
			switch resource.Status {
			case jobs.StatusRunning:
			case jobs.StatusCancelRequested:
				return interrupt(jobs.ErrCancellationRequested)
			default:
				return interrupt(jobs.ErrExecutionLost)
			}
		}
	}
}

// Discard every output after interruption, while preserving a teardown defect
// for the owning lifecycle to record. Execution fencing still takes precedence
// over any attempt to finalize a Job whose ownership was lost.
func interruptedRenderResult(result candidateRenderResult) candidateRenderResult {
	if errors.Is(result.err, renderprocess.ErrCleanup) {
		return candidateRenderResult{reason: "partial_output_cleanup_failed", err: result.err}
	}
	return candidateRenderResult{}
}

func (w *reportingJobWorker) stopForRenderControl(ctx, parentCtx context.Context, execution jobs.Execution, progress jobs.Progress, err error) error {
	if stop, contextErr := w.stopForContext(ctx, parentCtx, execution, progress); stop {
		return contextErr
	}
	if errors.Is(err, jobs.ErrCancellationRequested) {
		return w.completeCanceled(parentCtx, execution, progress)
	}
	return err
}

// Render preparation and participant work share render failure accounting even
// when the total deadline expires before an artifact candidate exists.
func (w *reportingJobWorker) stopForRenderPreparationControl(ctx, parentCtx context.Context, execution jobs.Execution, progress jobs.Progress, err error) error {
	if parentCtx.Err() == nil && errors.Is(err, context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, err))
	}
	return w.stopForRenderControl(ctx, parentCtx, execution, progress, err)
}
