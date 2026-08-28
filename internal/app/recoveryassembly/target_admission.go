package recoveryassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
)

type targetServingAdmission struct {
	admission *processlease.RecoveryTargetAdmission
	ctx       context.Context
	cancel    context.CancelCauseFunc
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
	admission, err := processlease.AcquireRecoveryTarget(
		ctx,
		concretePool,
		acquireTimeout,
		lossDetection,
	)
	if err != nil {
		return nil, err
	}
	admissionCtx, cancel := context.WithCancelCause(ctx)
	target := &targetServingAdmission{
		admission: admission,
		ctx:       admissionCtx,
		cancel:    cancel,
	}
	go target.propagateCancellation()
	return target, nil
}

func (admission *targetServingAdmission) Context() context.Context {
	if admission == nil || admission.ctx == nil {
		return context.Background()
	}
	return admission.ctx
}

func (admission *targetServingAdmission) AssertHeld() error {
	if admission == nil || admission.admission == nil {
		return application.ErrTargetServingLeaseLost
	}
	if err := admission.admission.AssertHeld(); err != nil {
		return application.ErrTargetServingLeaseLost
	}
	if errors.Is(context.Cause(admission.ctx), application.ErrTargetServingLeaseLost) {
		return application.ErrTargetServingLeaseLost
	}
	return nil
}

func (admission *targetServingAdmission) Release(ctx context.Context) error {
	if admission == nil || admission.admission == nil {
		return nil
	}
	err := admission.admission.Release(ctx)
	admission.cancel(context.Canceled)
	return err
}

func (admission *targetServingAdmission) propagateCancellation() {
	<-admission.admission.Context().Done()
	cause := context.Cause(admission.admission.Context())
	if errors.Is(cause, processlease.ErrRecoveryTargetLeaseLost) {
		cause = application.ErrTargetServingLeaseLost
	}
	admission.cancel(cause)
}
