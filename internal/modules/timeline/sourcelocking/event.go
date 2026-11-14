package sourcelocking

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LockEventTx serializes source-owned writes through the borrowed transaction.
func LockEventTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	var locked uuid.UUID
	return tx.QueryRow(ctx, `SELECT record_id FROM timeline_events WHERE record_id=$1 FOR UPDATE`, recordID).Scan(&locked)
}
