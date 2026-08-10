package mentions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LoadTimelineCollectionFieldsChangedTx exposes Entities-owned mention facts
// without allowing Timeline to query the entity_mentions source table.
func (*Store) LoadTimelineCollectionFieldsChangedTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changedAt time.Time) ([]string, error) {
	rows, err := tx.Query(ctx, `
SELECT DISTINCT source_field_key
  FROM entity_mentions
 WHERE source_record_id = $1
   AND created_at = $2
 ORDER BY source_field_key
`, recordID, changedAt.UTC())
	if err != nil {
		return nil, fmt.Errorf("query Entities-owned Timeline mention history facts: %w", err)
	}
	defer rows.Close()
	fields := make([]string, 0)
	for rows.Next() {
		var field string
		if err := rows.Scan(&field); err != nil {
			return nil, fmt.Errorf("scan Entities-owned Timeline mention history fact: %w", err)
		}
		fields = append(fields, field)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Entities-owned Timeline mention history facts: %w", err)
	}
	return fields, nil
}
