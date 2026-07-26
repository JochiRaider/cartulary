package revisions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RecordRevisionWindowEntry struct {
	RowVersion  int64
	BeforeJSON  []byte
	AfterJSON   []byte
	ActorUserID uuid.UUID
	CreatedAt   time.Time
}

type Reader struct{}

func NewReader() Reader {
	return Reader{}
}

func (Reader) ListRecordRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, firstVersion int64, lastVersion int64) ([]RecordRevisionWindowEntry, error) {
	rows, err := tx.Query(ctx, `
SELECT rr.row_version, rr.before_json, rr.after_json, cs.actor_user_id, cs.created_at
  FROM record_revisions rr
  JOIN change_sets cs ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND rr.row_version >= $2
   AND rr.row_version <= $3
 ORDER BY rr.row_version ASC
`, recordID, firstVersion, lastVersion)
	if err != nil {
		return nil, fmt.Errorf("query record revision window: %w", err)
	}
	defer rows.Close()
	result := make([]RecordRevisionWindowEntry, 0)
	for rows.Next() {
		var entry RecordRevisionWindowEntry
		if err := rows.Scan(&entry.RowVersion, &entry.BeforeJSON, &entry.AfterJSON, &entry.ActorUserID, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan record revision window: %w", err)
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record revision window: %w", err)
	}
	return result, nil
}
