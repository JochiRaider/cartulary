package recoveryassembly

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
)

type targetServingAdmission struct {
	lease       *processlease.Lease
	ctx         context.Context
	cancel      context.CancelCauseFunc
	watchCancel context.CancelFunc
	watchDone   chan struct{}
	releaseOnce sync.Once
	releaseErr  error
}

func AcquireTargetServingAdmission(
	ctx context.Context,
	pool application.PostgresPool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (application.TargetServingAdmission, error) {
	concretePool, ok := pool.(*pgxpool.Pool)
	if !ok || concretePool == nil {
		return nil, fmt.Errorf("restore target serving lease requires a concrete PostgreSQL pool")
	}
	lease, err := processlease.Acquire(
		ctx,
		processlease.PostgresBackend{
			Pool:        concretePool,
			AdvisoryKey: processlease.ServingAdvisoryKey,
			Purpose:     "restore target",
			Mode:        processlease.LockExclusive,
		},
		acquireTimeout,
		lossDetection,
	)
	if err != nil {
		return nil, err
	}
	admissionCtx, cancel := context.WithCancelCause(ctx)
	monitorCtx, watchCancel := context.WithCancel(context.Background())
	admission := &targetServingAdmission{
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

func (admission *targetServingAdmission) Context() context.Context {
	if admission == nil || admission.ctx == nil {
		return context.Background()
	}
	return admission.ctx
}

func (admission *targetServingAdmission) AssertHeld() error {
	if admission == nil || admission.lease == nil || admission.lease.State() != processlease.StateHeld {
		return application.ErrTargetServingLeaseLost
	}
	if errors.Is(context.Cause(admission.ctx), application.ErrTargetServingLeaseLost) {
		return application.ErrTargetServingLeaseLost
	}
	return nil
}

func (admission *targetServingAdmission) Release(ctx context.Context) error {
	if admission == nil {
		return nil
	}
	admission.releaseOnce.Do(func() {
		admission.watchCancel()
		<-admission.watchDone
		if admission.lease.State() == processlease.StateHeld {
			admission.releaseErr = admission.lease.Release(ctx)
		} else {
			admission.lease.Close()
		}
		admission.cancel(context.Canceled)
	})
	return admission.releaseErr
}

func (admission *targetServingAdmission) watch(ctx context.Context) {
	defer close(admission.watchDone)
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-admission.lease.Events():
			switch event.State {
			case processlease.StateUncertain, processlease.StateLost:
				admission.cancel(application.ErrTargetServingLeaseLost)
				if event.State == processlease.StateLost {
					return
				}
			}
		}
	}
}
