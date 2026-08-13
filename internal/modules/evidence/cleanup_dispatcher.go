package evidence

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	CleanupDispatcherIdentity = "evidence.failed_unattached_blob_cleanup.v1"
	cleanupDispatchInterval   = 15 * time.Minute
)

type CleanupSweeper interface {
	SweepFailedUnattachedBlobs(context.Context, CleanupObjectDeleter, time.Time) (CleanupSweepResult, error)
}

type CleanupSweepObservation struct {
	Operation             string
	Result                string
	ErrorClass            string
	Duration              time.Duration
	HealthSnapshotValid   bool
	OverdueBlobCount      int64
	OldestEligibleBlobAge time.Duration
}

type CleanupObserver interface {
	ObserveCleanupSweep(context.Context, CleanupSweepObservation)
}

type CleanupDispatcher struct {
	sweeper  CleanupSweeper
	deleter  CleanupObjectDeleter
	observer CleanupObserver
	now      func() time.Time
	interval time.Duration

	mu     sync.Mutex
	runMu  sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewCleanupDispatcher(
	sweeper CleanupSweeper,
	deleter CleanupObjectDeleter,
	observer CleanupObserver,
	now func() time.Time,
) (*CleanupDispatcher, error) {
	if sweeper == nil {
		return nil, errors.New("compose Evidence cleanup dispatcher: sweeper is required")
	}
	if deleter == nil {
		return nil, errors.New("compose Evidence cleanup dispatcher: object deleter is required")
	}
	if observer == nil {
		return nil, errors.New("compose Evidence cleanup dispatcher: observer is required")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CleanupDispatcher{
		sweeper:  sweeper,
		deleter:  deleter,
		observer: observer,
		now:      now,
		interval: cleanupDispatchInterval,
	}, nil
}

func (d *CleanupDispatcher) Identity() string {
	return CleanupDispatcherIdentity
}

// Start activates the private dispatcher. Application assembly calls Start
// only after serving readiness is published. The first sweep is asynchronous
// and immediate; subsequent sweeps use the fixed owner interval.
func (d *CleanupDispatcher) Start(parent context.Context) error {
	if d == nil || d.sweeper == nil || d.deleter == nil || d.observer == nil || d.interval <= 0 {
		return errors.New("evidence cleanup dispatcher is not configured")
	}
	if parent == nil {
		parent = context.Background()
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	d.cancel = cancel
	d.done = make(chan struct{})
	go d.run(ctx, d.done)
	return nil
}

func (d *CleanupDispatcher) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	cancel := d.cancel
	done := d.done
	d.cancel = nil
	d.done = nil
	d.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *CleanupDispatcher) RunOnce(ctx context.Context) (CleanupSweepResult, error) {
	if d == nil || d.sweeper == nil || d.deleter == nil || d.observer == nil {
		return CleanupSweepResult{}, errors.New("evidence cleanup dispatcher is not configured")
	}
	d.runMu.Lock()
	defer d.runMu.Unlock()
	started := time.Now()
	result, err := d.sweeper.SweepFailedUnattachedBlobs(ctx, d.deleter, d.now().UTC())
	observation := CleanupSweepObservation{
		Operation:             "cleanup_sweep",
		Result:                cleanupObservationResult(ctx, err),
		ErrorClass:            cleanupObservationErrorClass(ctx, err),
		Duration:              time.Since(started),
		HealthSnapshotValid:   result.HealthSnapshotValid,
		OverdueBlobCount:      result.OverdueBlobCount,
		OldestEligibleBlobAge: result.OldestEligibleAge,
	}
	d.observer.ObserveCleanupSweep(context.WithoutCancel(ctx), observation)
	return result, err
}

func (d *CleanupDispatcher) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	_, _ = d.RunOnce(ctx)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = d.RunOnce(ctx)
		}
	}
}

func cleanupObservationResult(ctx context.Context, err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "timeout"
	}
	return "failed"
}

func cleanupObservationErrorClass(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return "dependency_unavailable"
	}
	return "internal_error"
}
