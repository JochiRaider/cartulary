package entities

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProjectionDependencies supplies the mention-derived part of the projection
// refresh scope. Entity aliases can change displayed auto-match provenance
// even when no explicit record link changes.
type ProjectionDependencies struct{}

func (ProjectionDependencies) ProjectionDependentsTx(ctx context.Context, tx pgx.Tx, _ uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	// Record IDs are global identities. Revisions validates the complete returned
	// scope through the Records envelope port before locks or mutations; this
	// source owner reads only its own mention relationships.
	rows, err := tx.Query(ctx, `
SELECT DISTINCT m.source_record_id
  FROM entity_mentions m
 WHERE m.resolved_record_id = ANY($1::uuid[])
   AND m.resolution_status = 'resolved'
 ORDER BY m.source_record_id`, changed)
	if err != nil {
		return nil, fmt.Errorf("load mention projection dependents: %w", err)
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
