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
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	ValidateCancellationExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	CompleteFailedTx(context.Context, pgx.Tx, jobs.Execution, jobs.FailureCompletion, time.Time) (jobs.Resource, error)
	CompleteCanceledTx(context.Context, pgx.Tx, jobs.Execution, jobs.CancellationCompletion, time.Time) (jobs.Resource, error)
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
		Execution:     request.Execution,
		Completion:    request.Completion,
		FinalCommitID: request.FinalCommitID,
		Mutate:        extensionstore.OwnerMutation(request.Mutate),
	})
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobFailure(
	ctx context.Context,
	request imports.JobFailureFinalization,
) (jobs.Resource, error) {
	return adapter.finalizeFailure(ctx, request)
}

func (adapter importJobSuccessFinalizer) FinalizeImportJobCancellation(
	ctx context.Context,
	request imports.JobCancellationFinalization,
) (jobs.Resource, error) {
	return adapter.finalizeCancellation(ctx, request)
}

func (adapter importJobSuccessFinalizer) finalizeFailure(
	ctx context.Context,
	request imports.JobFailureFinalization,
) (jobs.Resource, error) {
	if adapter.pool == nil || adapter.transactions == nil || adapter.now == nil {
		return jobs.Resource{}, errors.New("import terminal finalizer is unavailable")
	}
	tx, err := adapter.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := adapter.transactions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
		return jobs.Resource{}, err
	}
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	resource, err := adapter.transactions.CompleteFailedTx(ctx, tx, request.Execution, request.Completion, adapter.now().UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}

func (adapter importJobSuccessFinalizer) finalizeCancellation(
	ctx context.Context,
	request imports.JobCancellationFinalization,
) (jobs.Resource, error) {
	if adapter.pool == nil || adapter.transactions == nil || adapter.now == nil {
		return jobs.Resource{}, errors.New("import terminal finalizer is unavailable")
	}
	tx, err := adapter.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := adapter.transactions.ValidateCancellationExecutionTx(ctx, tx, request.Execution); err != nil {
		return jobs.Resource{}, err
	}
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	resource, err := adapter.transactions.CompleteCanceledTx(ctx, tx, request.Execution, request.Completion, adapter.now().UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}
