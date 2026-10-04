package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReadTerminalDispositionTx is an owner recovery projection, not a public Job
// read. Public expiry never deletes these lifecycle facts. Lock ordering is the
// same transition guard then Job row used by live completion; owner locks come
// after this call. A nonterminal Job is not permission to repair owner state.
func (s *TransactionService) ReadTerminalDispositionTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (TerminalDisposition, bool, error) {
	if s == nil || s.terminalEffects == nil || tx == nil || id == uuid.Nil {
		return TerminalDisposition{}, false, ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, id); err != nil {
		return TerminalDisposition{}, false, err
	}
	var status string
	var finished *time.Time
	err := tx.QueryRow(ctx, `SELECT status,finished_at FROM jobs WHERE job_id=$1 FOR UPDATE`, id).Scan(&status, &finished)
	if errors.Is(err, pgx.ErrNoRows) {
		return TerminalDisposition{}, false, ErrNotFound
	}
	if err != nil {
		return TerminalDisposition{}, false, err
	}
	if status != StatusSucceeded && status != StatusFailed && status != StatusCanceled {
		return TerminalDisposition{}, false, nil
	}
	if finished == nil || finished.IsZero() {
		return TerminalDisposition{}, false, ErrInvalidTransition
	}
	return TerminalDisposition{JobID: id, Status: status, FinishedAt: finished.UTC()}, true, nil
}
