// Package references supplies bounded, source-owned Party reference labels.
package references

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LoadLabels reads only live Parties in the caller's incident and selected set.
// Mutation callers pass their transaction to preserve the admitted source view.
func LoadLabels(ctx context.Context, query interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, incidentID uuid.UUID, partyIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	labels := make(map[uuid.UUID]string, len(partyIDs))
	if len(partyIDs) == 0 {
		return labels, nil
	}
	rows, err := query.Query(ctx, `SELECT p.record_id, p.display_name FROM parties p
JOIN records r ON r.record_id=p.record_id AND r.incident_id=p.incident_id
WHERE p.incident_id=$1 AND p.record_id=ANY($2::uuid[]) AND r.deleted_at IS NULL`, incidentID, partyIDs)
	if err != nil {
		return nil, fmt.Errorf("read Party reference labels: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			return nil, err
		}
		labels[id] = label
	}
	return labels, rows.Err()
}
