package extensionassembly

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type importJobSuccessFinalizer struct {
	finalizer    *extensionstore.OwnerFinalizer
	pool         postgres.DB
	transactions importTerminalCompleter
	now          func() time.Time
}

type importTerminalCompleter interface {
	CompleteFailedTx(context.Context, pgx.Tx, jobs.TransitionParams, time.Time) (jobs.Resource, error)
	CompleteCanceledTx(context.Context, pgx.Tx, jobs.TransitionParams, time.Time) (jobs.Resource, error)
}

func NewImportJobSuccessFinalizer(
	finalizer *extensionstore.OwnerFinalizer,
	pool postgres.DB,
	transactions importTerminalCompleter,
	now func() time.Time,
) imports.JobSuccessFinalizer {
	if finalizer == nil || pool == nil || transactions == nil {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return importJobSuccessFinalizer{
		finalizer:    finalizer,
		pool:         pool,
		transactions: transactions,
		now:          now,
	}
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobSuccess(
	ctx context.Context,
	request imports.JobSuccessFinalization,
) (jobs.Resource, error) {
	return adapter.finalizer.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{
		Transition:    request.Transition,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobFailure(
	ctx context.Context,
	request imports.JobTerminalFinalization,
) (jobs.Resource, error) {
	return adapter.finalizeTerminal(ctx, request, jobs.StatusFailed)
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobCancellation(
	ctx context.Context,
	request imports.JobTerminalFinalization,
) (jobs.Resource, error) {
	return adapter.finalizeTerminal(ctx, request, jobs.StatusCanceled)
}

func (adapter importJobSuccessFinalizer) finalizeTerminal(
	ctx context.Context,
	request imports.JobTerminalFinalization,
	status string,
) (jobs.Resource, error) {
	if adapter.pool == nil || adapter.transactions == nil || adapter.now == nil {
		return jobs.Resource{}, errors.New("import terminal finalizer is unavailable")
	}
	tx, err := adapter.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	var resource jobs.Resource
	switch status {
	case jobs.StatusFailed:
		resource, err = adapter.transactions.CompleteFailedTx(
			ctx,
			tx,
			request.Transition,
			adapter.now().UTC(),
		)
	case jobs.StatusCanceled:
		resource, err = adapter.transactions.CompleteCanceledTx(
			ctx,
			tx,
			request.Transition,
			adapter.now().UTC(),
		)
	default:
		return jobs.Resource{}, errors.New("unsupported import terminal finalizer status")
	}
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}
