package imports

import (
	"context"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// Observation interrupts expensive uncommitted work. It grants no commit
// authority: the unit still fences its exact execution under the Jobs lock
// immediately before publishing its outcome in the owner transaction.
func (s *service) observeApplyCancellation(ctx context.Context, cancelWork context.CancelCauseFunc, execution jobs.Execution) func() {
	observationCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			readCtx, cancelRead := context.WithTimeout(observationCtx, jobs.ProductionRuntimePolicy().AttemptOperationTimeout)
			job, err := s.jobManager.ObserveExecution(readCtx, execution)
			cancelRead()
			if observationCtx.Err() != nil {
				return
			}
			if err != nil {
				cancelWork(err)
				return
			}
			switch job.Status {
			case jobs.StatusRunning:
			case jobs.StatusCancelRequested:
				cancelWork(errImportUnitCanceled)
				return
			default:
				cancelWork(jobs.ErrExecutionLost)
				return
			}
			select {
			case <-observationCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { stop(); <-done }
}
