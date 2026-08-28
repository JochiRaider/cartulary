package processlease

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRecoveryTargetLeaseLost = errors.New("recovery_target_lease_lost")

// RecoveryTargetAdmission holds the exclusive Recovery side of the serving
// fence and cancels its context if continuous ownership can no longer be
// proved.
type RecoveryTargetAdmission struct {
	lease       *RecoveryTargetLease
	ctx         context.Context
	cancel      context.CancelCauseFunc
	watchCancel context.CancelFunc
	watchDone   chan struct{}
	releaseOnce sync.Once
	releaseErr  error
}

func AcquireRecoveryTarget(
	ctx context.Context,
	pool *pgxpool.Pool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (*RecoveryTargetAdmission, error) {
	lease, err := AcquireRecoveryTargetLease(ctx, pool, acquireTimeout, lossDetection)
	if err != nil {
		return nil, err
	}
	admissionCtx, cancel := context.WithCancelCause(ctx)
	monitorCtx, watchCancel := context.WithCancel(context.Background())
	admission := &RecoveryTargetAdmission{
		lease:       lease,
		ctx:         admissionCtx,
		cancel:      cancel,
		watchCancel: watchCancel,
		watchDone:   make(chan struct{}),
	}
	lease.StartMonitor(monitorCtx)
	go admission.watch(monitorCtx)
	return admission, nil
}

func (a *RecoveryTargetAdmission) Context() context.Context {
	if a == nil || a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

func (a *RecoveryTargetAdmission) AssertHeld() error {
	if a == nil || a.lease == nil || a.lease.State() != StateHeld {
		return ErrRecoveryTargetLeaseLost
	}
	if errors.Is(context.Cause(a.ctx), ErrRecoveryTargetLeaseLost) {
		return ErrRecoveryTargetLeaseLost
	}
	return nil
}

func (a *RecoveryTargetAdmission) SessionIdentity() string {
	if a == nil || a.lease == nil {
		return ""
	}
	return a.lease.SessionIdentity()
}

func (a *RecoveryTargetAdmission) Release(ctx context.Context) error {
	if a == nil {
		return nil
	}
	a.releaseOnce.Do(func() {
		a.watchCancel()
		<-a.watchDone
		if a.lease.State() == StateHeld {
			a.releaseErr = a.lease.Release(ctx)
		} else {
			a.lease.Close()
		}
		a.cancel(context.Canceled)
	})
	return a.releaseErr
}

func (a *RecoveryTargetAdmission) watch(ctx context.Context) {
	defer close(a.watchDone)
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-a.lease.Events():
			switch event.State {
			case StateUncertain, StateLost:
				a.cancel(ErrRecoveryTargetLeaseLost)
				if event.State == StateLost {
					return
				}
			}
		}
	}
}
