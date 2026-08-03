package revisions

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rollbackRecordLocker struct {
	envelopes RecordEnvelopePort
}

func (rollbackRecordLocker) orderedRecordIDs(recordIDs []uuid.UUID) []uuid.UUID {
	ordered := append([]uuid.UUID(nil), recordIDs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].String() < ordered[j].String()
	})
	result := ordered[:0]
	for _, recordID := range ordered {
		if recordID == uuid.Nil || (len(result) > 0 && recordID == result[len(result)-1]) {
			continue
		}
		result = append(result, recordID)
	}
	return result
}

func (l rollbackRecordLocker) lockTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := l.envelopes.LockDestructiveRecordsNowaitTx(ctx, tx, l.orderedRecordIDs(recordIDs)); err != nil {
		var locked *EnvelopeLockError
		if errors.As(err, &locked) {
			return &RecordLockedError{RecordID: locked.RecordID}
		}
		if errors.Is(err, ErrEnvelopeNotFound) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}
