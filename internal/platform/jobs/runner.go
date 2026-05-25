package jobs

import (
	"context"
	"errors"
	"sync"
)

var ErrRunnerClosed = errors.New("jobs: runner closed")

type Runner struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

func NewRunner() *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{ctx: ctx, cancel: cancel}
}

func (r *Runner) Dispatch(work func(context.Context)) error {
	if r == nil || work == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRunnerClosed
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		work(r.ctx)
	}()
	return nil
}

func (r *Runner) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		r.cancel()
	}
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
