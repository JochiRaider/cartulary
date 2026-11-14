package references

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SourceLocker protects typed Party presence in the borrowed transaction.
// The caller protects and admits the authoritative envelope first.
type SourceLocker struct{}

func (SourceLocker) LockSourceTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID, nowait bool) (bool, error) {
	query := "SELECT record_id FROM parties WHERE incident_id=$1 AND record_id=$2 FOR SHARE"
	if nowait {
		query += " NOWAIT"
	}
	var found uuid.UUID
	err := tx.QueryRow(ctx, query, incidentID, recordID).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
