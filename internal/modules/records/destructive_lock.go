package records

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrRecordEnvelopeNotFound           = errors.New("records: record envelope not found")
	ErrDestructiveOperationRecordLocked = errors.New("records: destructive operation record locked")
)

type DestructiveOperationRecordLockedError struct {
	RecordID uuid.UUID
}

func (e *DestructiveOperationRecordLockedError) Error() string {
	return ErrDestructiveOperationRecordLocked.Error()
}

func (e *DestructiveOperationRecordLockedError) Unwrap() error {
	return ErrDestructiveOperationRecordLocked
}

func LockDestructiveOperationRecordsNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	ordered := append([]uuid.UUID(nil), recordIDs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].String() < ordered[j].String()
	})
	for i := 0; i < len(ordered); i++ {
		if i > 0 && ordered[i] == ordered[i-1] {
			continue
		}
		var locked uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE NOWAIT`, ordered[i]).Scan(&locked); err != nil {
			if isLockUnavailable(err) {
				return &DestructiveOperationRecordLockedError{RecordID: ordered[i]}
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRecordEnvelopeNotFound
			}
			return err
		}
	}
	return nil
}

func isLockUnavailable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}
