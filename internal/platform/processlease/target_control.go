package processlease

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TargetControlTransaction cannot settle the transaction or release the lease.
// Recovery owns its statements; this package owns the original session fence.
type TargetControlTransaction interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// WithControlTransaction commits short installation-control changes on the
// session holding the exclusive serving lease. A disconnected lease session
// therefore cannot publish completion through a replacement pool connection.
func (a *RecoveryTargetAdmission) WithControlTransaction(ctx context.Context, run func(context.Context, TargetControlTransaction) error) (resultErr error) {
	if ctx == nil || run == nil {
		return ErrInvalidTransition
	}
	if err := a.AssertHeld(); err != nil {
		return err
	}
	if err := a.Context().Err(); err != nil {
		return err
	}
	a.lease.mu.RLock()
	session, ok := a.lease.session.(*postgresSession)
	identity := a.lease.sessionID
	budget := a.lease.lossDetection / 2
	a.lease.mu.RUnlock()
	if !ok || session.identity != identity || session.advisoryKey != ServingAdvisoryKey || session.mode != LockExclusive || budget <= 0 {
		return ErrRecoveryTargetLeaseLost
	}
	// Reserve half the proof interval for independent settlement, and never
	// let control work monopolize the session beyond its ownership budget.
	work, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	if err := session.enter(work); err != nil {
		return err
	}
	defer session.leave()
	if session.closed || session.connection == nil {
		return ErrRecoveryTargetLeaseLost
	}
	defer func() {
		if session.connection.Conn().IsClosed() {
			a.lease.transition(StateHeld, StateUncertain, "ownership_proof_lost")
			a.lease.transition(StateUncertain, StateLost, "session_loss_confirmed")
			a.cancel(ErrRecoveryTargetLeaseLost)
			resultErr = errors.Join(resultErr, ErrRecoveryTargetLeaseLost)
		}
	}()
	if err := a.AssertHeld(); err != nil {
		return err
	}
	if err := a.proveControlSession(work, session); err != nil {
		return err
	}
	tx, err := session.connection.BeginTx(work, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		settle, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
		defer cancel()
		if err := tx.Rollback(settle); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			resultErr = errors.Join(resultErr, err)
		}
	}()
	if err := run(work, tx); err != nil {
		return err
	}
	if err := work.Err(); err != nil {
		return err
	}
	if err := a.Context().Err(); err != nil {
		return err
	}
	if err := a.AssertHeld(); err != nil {
		return err
	}
	if err := a.proveControlSession(work, session); err != nil {
		return err
	}
	return tx.Commit(work)
}

func (a *RecoveryTargetAdmission) proveControlSession(ctx context.Context, session *postgresSession) error {
	proof := session.proveLocked(ctx)
	if proof == ProofContinuous {
		return nil
	}
	a.lease.transition(StateHeld, StateUncertain, "ownership_proof_uncertain")
	if proof == ProofLost {
		a.lease.transition(StateUncertain, StateLost, "session_loss_confirmed")
	}
	a.cancel(ErrRecoveryTargetLeaseLost)
	return ErrRecoveryTargetLeaseLost
}
