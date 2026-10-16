package links

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProjectionDependentsTx includes both endpoints of current relationships.
// Endpoint deletion does not erase a relationship: using the active-links view
// here would miss precisely the counts that a record restore must refresh.
func (*Store) ProjectionDependentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
SELECT dst_record_id FROM record_links
 WHERE incident_id = $1 AND src_record_id = ANY($2::uuid[]) AND deleted_at IS NULL
UNION
SELECT src_record_id FROM record_links
 WHERE incident_id = $1 AND dst_record_id = ANY($2::uuid[]) AND deleted_at IS NULL
ORDER BY 1`, incidentID, changed)
	if err != nil {
		return nil, fmt.Errorf("load relationship projection dependents: %w", err)
	}
	defer rows.Close()
	var result []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}
