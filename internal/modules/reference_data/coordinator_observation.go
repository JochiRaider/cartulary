package reference_data

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// executionObservation propagates Jobs cancellation and lease loss through
// bounded streaming reads and private index writes, including a single large
// cohort member. Terminal mutation still rechecks the lease and cancellation
// under Jobs' transaction lock; a polling result never authorizes publication.
type executionObservation struct {
	mu           sync.Mutex
	cancellation *int64
	err          error
	stop         context.CancelFunc
	done         chan struct{}
}

func (c *verificationService) observeDuringPreparation(ctx context.Context, execution jobs.Execution, started time.Time, cancel context.CancelCauseFunc) *executionObservation {
	watchCtx, stop := context.WithCancel(ctx)
	o := &executionObservation{stop: stop, done: make(chan struct{})}
	go func() {
		defer close(o.done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
			}
			job, err := c.operations.ObserveExecution(watchCtx, execution)
			if watchCtx.Err() != nil {
				return
			}
			if errors.Is(err, jobs.ErrExecutionLost) {
				// A terminal parent Job ends observation without interrupting
				// the scope: it is neither cancellation nor a lost lease, so a
				// proven commit's remaining completion receives no grace from it
				// (EXT-REQ-215, Reference Pack §23.2). Loss of a still-running
				// execution remains the Jobs runtime's to signal (REQ-01-453).
				current, getErr := c.operations.Get(watchCtx, execution.JobID())
				if watchCtx.Err() != nil {
					return
				}
				if getErr == nil && terminalJobStatus(current.Status) {
					return
				}
			}
			var sample *int64
			if err == nil && job.Status == jobs.StatusCancelRequested {
				value := time.Since(started).Nanoseconds()
				sample = &value
				err = jobs.ErrCancellationRequested
			}
			if err != nil {
				o.mu.Lock()
				o.err = err
				o.cancellation = sample
				o.mu.Unlock()
				cancel(err)
				return
			}
		}
	}()
	return o
}
func terminalJobStatus(status string) bool {
	switch status {
	case jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled:
		return true
	}
	return false
}

func (o *executionObservation) finish() (*int64, error) {
	o.stop()
	<-o.done
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.cancellation, o.err
}
