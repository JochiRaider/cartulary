package records

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EnvelopeIdentity struct {
	RecordID   uuid.UUID
	RecordType string
}

// LockIncidentEnvelopeIdentitiesTx serializes dependent refresh without changing
// key columns or owning the caller's transaction. Missing IDs are omitted.
func LockIncidentEnvelopeIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, ids []uuid.UUID) ([]EnvelopeIdentity, error) {
	rows, err := tx.Query(ctx, `SELECT record_id, record_type FROM records WHERE incident_id=$1 AND record_id=ANY($2::uuid[]) ORDER BY record_id FOR NO KEY UPDATE`, incidentID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	endpoints := []EnvelopeIdentity{}
	for rows.Next() {
		var value EnvelopeIdentity
		if err := rows.Scan(&value.RecordID, &value.RecordType); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return endpoints, nil
}
