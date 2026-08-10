package links

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LoadTimelineCollectionFieldsChangedTx exposes Links-owned historical facts
// without allowing Timeline to query Links source tables directly.
func (*Store) LoadTimelineCollectionFieldsChangedTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changedAt time.Time) ([]string, error) {
	rows, err := tx.Query(ctx, `
SELECT field_key
  FROM (
        SELECT rl.field_key
          FROM record_links rl
         WHERE rl.src_record_id = $1
           AND (rl.created_at = $2 OR rl.deleted_at = $2)
        UNION
        SELECT 'timeline.tags'
          FROM record_tags rt
         WHERE rt.record_id = $1
           AND (rt.created_at = $2 OR rt.deleted_at = $2)
       ) changed
 ORDER BY field_key
`, recordID, changedAt.UTC())
	if err != nil {
		return nil, fmt.Errorf("query Links-owned Timeline history facts: %w", err)
	}
	defer rows.Close()
	fields := make([]string, 0)
	for rows.Next() {
		var field string
		if err := rows.Scan(&field); err != nil {
			return nil, fmt.Errorf("scan Links-owned Timeline history fact: %w", err)
		}
		fields = append(fields, field)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Links-owned Timeline history facts: %w", err)
	}
	return fields, nil
}
