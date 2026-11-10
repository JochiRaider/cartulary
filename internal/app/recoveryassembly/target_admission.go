package recoveryassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
)

type targetServingAdmission struct {
	admission *processlease.RecoveryTargetAdmission
	database  dbport.DB
	ctx       context.Context
	cancel    context.CancelCauseFunc
}

func (admission *targetServingAdmission) TargetDatabase() dbport.DB {
	if admission == nil {
		return nil
	}
	return admission.database
}

func (admission *targetServingAdmission) WithControlTransaction(ctx context.Context, run func(context.Context, recovery.TargetControlTransaction) error) error {
	if admission == nil || admission.admission == nil || run == nil {
		return application.ErrTargetServingLeaseLost
	}
	err := admission.admission.WithControlTransaction(ctx, func(work context.Context, tx processlease.TargetControlTransaction) error { return run(work, tx) })
	if errors.Is(err, processlease.ErrRecoveryTargetLeaseLost) {
		return errors.Join(application.ErrTargetServingLeaseLost, err)
	}
	return err
}

func InspectStoppedTarget(ctx context.Context, pool application.PostgresPool) error {
	admitted, ok := pool.(postgres.AdmittedPool)
	if !ok || admitted == nil || admitted.Pool() == nil {
		return errors.New("restore target preflight requires an admitted PostgreSQL pool")
	}
	return processlease.RequireRecoveryTargetStopped(ctx, admitted.Pool())
}

func AcquireTargetServingAdmission(
	ctx context.Context,
	pool application.PostgresPool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (application.TargetServingAdmission, error) {
	admitted, ok := pool.(postgres.AdmittedPool)
	if !ok || admitted == nil || admitted.Pool() == nil {
		return nil, fmt.Errorf("restore target serving lease requires an admitted PostgreSQL pool")
	}
	concretePool := admitted.Pool()
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
		database:  admitted,
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
	if errors.Is(err, processlease.ErrRecoveryTargetLeaseLost) {
		return application.ErrTargetServingLeaseLost
	}
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
