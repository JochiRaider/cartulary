package harnesscontrol

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Faults are consumed before their effect. Correlation is the import-unit ID or
// job ID, independent of retries and handler-attempt leases.
func (c *Controls) fault(ctx context.Context, cancel context.CancelFunc, boundary, correlation string, cancelJob func() error) error {
	fault, ok := c.Faults.ConsumeNetworkFlowFaultFor(boundary, correlation)
	if !ok {
		return nil
	}
	switch fault.FaultKind {
	case NetworkFlowFaultKindReturnError:
		return fmt.Errorf("network flow harness fault: %s", fault.ErrorCode)
	case NetworkFlowFaultKindPanic:
		panic("network flow harness fault")
	case NetworkFlowFaultKindCancelContext:
		cancel()
		return ctx.Err()
	case NetworkFlowFaultKindWorkerCancel:
		if cancelJob == nil {
			return errors.New("network flow harness cancellation unavailable")
		}
		return cancelJob()
	case NetworkFlowFaultKindWorkerCrash:
		if c.Crash == nil {
			return errors.New("network flow harness owned process unavailable")
		}
		c.Crash()
		return errors.New("network flow harness crash returned")
	default:
		return errors.New("network flow harness unknown fault")
	}
}

type faultDatabase struct {
	postgres.DB
	controls *Controls
}
type faultTransaction struct {
	pgx.Tx
	controls    *Controls
	correlation string
}

func (c *Controls) WrapDatabase(db postgres.DB) postgres.DB {
	return &faultDatabase{DB: db, controls: c}
}
func (db *faultDatabase) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &faultTransaction{Tx: tx, controls: db.controls}, nil
}
func (tx *faultTransaction) Commit(ctx context.Context) error {
	if tx.correlation == "" {
		return tx.Tx.Commit(ctx)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := tx.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryImportBeforeTransactionCommit, tx.correlation, nil); err != nil {
		_ = tx.Tx.Rollback(context.WithoutCancel(ctx))
		return err
	}
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	return tx.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryImportAfterTransactionCommitBeforeReply, tx.correlation, nil)
}

type faultImportFacade struct {
	imports.ExtensionImportFacade
	controls *Controls
}

func (c *Controls) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return &faultImportFacade{ExtensionImportFacade: f, controls: c}
}
func (f *faultImportFacade) ApplyImportUnitTx(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error) {
	var empty imports.ExtensionImportApplyResult
	decorated, ok := tx.(*faultTransaction)
	if !ok {
		return empty, errors.New("network flow harness requires decorated import transaction")
	}
	decorated.correlation = request.ImportUnitID.String()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := f.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryImportBeforeOwnerApply, decorated.correlation, nil); err != nil {
		return empty, err
	}
	result, err := f.ExtensionImportFacade.ApplyImportUnitTx(ctx, tx, request)
	if err != nil {
		return empty, err
	}
	if err := f.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryImportAfterOwnerApply, decorated.correlation, nil); err != nil {
		return empty, err
	}
	return result, nil
}

type jobCanceler interface {
	Cancel(context.Context, jobs.CancelParams) (jobs.CancelResult, error)
}
type faultJobRunner struct {
	networkflow.GraphViewJobRunner
	controls *Controls
}
type faultJobManager struct {
	networkflow.GraphViewJobManager
	controls *Controls
	canceler jobCanceler
}
type faultJobFinalizer struct {
	networkflow.GraphViewJobFinalizer
	controls *Controls
}

func (c *Controls) ConfigureNetworkFlow(d networkflow.ModuleDependencies) networkflow.ModuleDependencies {
	d.TableIDEntropy = c.TableIDEntropy()
	d.CursorNonceEntropy = c.CursorNonceEntropy()
	canceler, _ := d.JobManager.(jobCanceler)
	if d.JobRunner != nil {
		d.JobRunner = &faultJobRunner{d.JobRunner, c}
	}
	if d.JobManager != nil {
		d.JobManager = &faultJobManager{d.JobManager, c, canceler}
	}
	if d.JobFinalizer != nil {
		d.JobFinalizer = &faultJobFinalizer{d.JobFinalizer, c}
	}
	return d
}
func (r *faultJobRunner) RegisterHandler(kind string, handler jobs.HandlerFunc) error {
	return r.GraphViewJobRunner.RegisterHandler(kind, func(ctx context.Context, execution jobs.Execution) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		if err := r.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryWorkerBeforeHandlerStart, execution.JobID().String(), nil); err != nil {
			return err
		}
		return handler(ctx, execution)
	})
}
func (m *faultJobManager) ObserveExecution(ctx context.Context, execution jobs.Execution) (jobs.Resource, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cancelJob := func() error {
		if m.canceler == nil {
			return errors.New("network flow harness Jobs cancellation unavailable")
		}
		job, err := m.GraphViewJobManager.Get(ctx, execution.JobID())
		if err != nil {
			return err
		}
		actor, err := uuid.Parse(job.SubmittedByUserID)
		if err != nil {
			return err
		}
		_, err = m.canceler.Cancel(ctx, jobs.CancelParams{JobID: execution.JobID(), ActorUserID: actor, ClientTxnID: "harness-cancel-" + execution.JobID().String(), NormalizedRequest: []byte(`{"operation":"harness_cancel"}`)})
		return err
	}
	if err := m.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryWorkerBeforeCancellationCheck, execution.JobID().String(), cancelJob); err != nil {
		return jobs.Resource{}, err
	}
	return m.GraphViewJobManager.ObserveExecution(ctx, execution)
}
func (f *faultJobFinalizer) FinalizeGraphViewJobSuccess(ctx context.Context, p networkflow.GraphViewJobSuccessFinalization) (jobs.Resource, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mutate := p.Mutate
	p.Mutate = func(ctx context.Context, tx pgx.Tx) error {
		if mutate != nil {
			if err := mutate(ctx, tx); err != nil {
				return err
			}
		}
		return f.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, p.Execution.JobID().String(), nil)
	}
	result, err := f.GraphViewJobFinalizer.FinalizeGraphViewJobSuccess(ctx, p)
	if err != nil {
		return result, err
	}
	if err := f.controls.fault(ctx, cancel, NetworkFlowFaultBoundaryWorkerAfterCompletedPublication, p.Execution.JobID().String(), nil); err != nil {
		return result, err
	}
	return result, nil
}
