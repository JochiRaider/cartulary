package records

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// VisitIncidentRecordIDsTx includes retained tombstones. Bounded keyset pages
// are closed before callbacks so callers can use the same borrowed transaction.
func (s *Store) VisitIncidentRecordIDsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, visit func(uuid.UUID) error) error {
	if visit == nil {
		return errors.New("records: missing incident record visitor")
	}
	const batchSize = 128
	var after *uuid.UUID
	for {
		rows, err := tx.Query(ctx, `SELECT record_id FROM records WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT $3`, incidentID, after, batchSize)
		if err != nil {
			return err
		}
		ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := visit(id); err != nil {
				return err
			}
		}
		if len(ids) < batchSize {
			return nil
		}
		last := ids[len(ids)-1]
		after = &last
	}
}
