package historyquery

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RevisionWindowRow struct {
	RowVersion  int64
	BeforeJSON  []byte
	AfterJSON   []byte
	ActorUserID uuid.UUID
	CreatedAt   time.Time
}

// Reader exposes only retained record-revision windows. The caller owns the
// transaction and all policy for interpreting or resolving conflicts.
type Reader struct{}

func NewReader() Reader { return Reader{} }

func (Reader) LoadRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) ([]RevisionWindowRow, error) {
	rows, err := tx.Query(ctx, `
SELECT rr.row_version, rr.before_json, rr.after_json, cs.actor_user_id, cs.created_at
  FROM record_revisions rr
  JOIN change_sets cs
    ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND rr.row_version >= $2
   AND rr.row_version <= $3
 ORDER BY rr.row_version ASC
`, recordID, baseRowVersion, currentRowVersion)
	if err != nil {
		return nil, fmt.Errorf("query record revision window: %w", err)
	}
	defer rows.Close()

	result := make([]RevisionWindowRow, 0)
	for rows.Next() {
		var row RevisionWindowRow
		if err := rows.Scan(&row.RowVersion, &row.BeforeJSON, &row.AfterJSON, &row.ActorUserID, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan record revision window: %w", err)
		}
		row.CreatedAt = row.CreatedAt.UTC()
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record revision window: %w", err)
	}
	return result, nil
}
