package conflicts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidRevisionWindow = errors.New("revisions conflicts: invalid revision window")

type RevisionWindowRow struct {
	RowVersion  int64
	BeforeJSON  []byte
	AfterJSON   []byte
	ActorUserID uuid.UUID
	CreatedAt   time.Time
}

// RevisionWindowReader reads retained history in the caller-owned transaction.
// It owns no policy for interpreting or resolving a conflict.
type RevisionWindowReader struct{}

func NewRevisionWindowReader() RevisionWindowReader { return RevisionWindowReader{} }

func (RevisionWindowReader) LoadRevisionWindowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	baseRowVersion int64,
	currentRowVersion int64,
) ([]RevisionWindowRow, error) {
	if tx == nil || recordID == uuid.Nil || baseRowVersion < 1 || currentRowVersion < baseRowVersion {
		return nil, ErrInvalidRevisionWindow
	}
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
