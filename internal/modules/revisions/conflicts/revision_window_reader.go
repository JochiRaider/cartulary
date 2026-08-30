package conflicts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errInvalidRevisionWindow = errors.New("revisions conflicts: invalid revision window")

type RevisionWindowRow struct {
	ChangeSetID   uuid.UUID
	RowVersion    int64
	BeforeJSON    []byte
	AfterJSON     []byte
	ActorUserID   uuid.UUID
	CreatedAt     time.Time
	ConflictFacts []RevisionConflictFact
}

type RevisionConflictFact struct {
	FieldKey      string
	BeforePresent bool
	BeforeValue   []byte
	AfterPresent  bool
	AfterValue    []byte
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
		return nil, errInvalidRevisionWindow
	}
	rows, err := tx.Query(ctx, `
SELECT rr.revision_id, cs.change_set_id, rr.row_version, rr.before_json, rr.after_json, cs.actor_user_id, cs.created_at
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
	byRevisionID := make(map[int64]int)
	for rows.Next() {
		var row RevisionWindowRow
		var revisionID int64
		if err := rows.Scan(&revisionID, &row.ChangeSetID, &row.RowVersion, &row.BeforeJSON, &row.AfterJSON, &row.ActorUserID, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan record revision window: %w", err)
		}
		row.CreatedAt = row.CreatedAt.UTC()
		byRevisionID[revisionID] = len(result)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record revision window: %w", err)
	}
	if len(result) == 0 {
		return result, nil
	}
	factRows, err := tx.Query(ctx, `
SELECT fact.revision_id, fact.field_key,
       fact.before_present, fact.before_value,
       fact.after_present, fact.after_value
  FROM record_revision_conflict_facts fact
 WHERE fact.revision_id = ANY($1)
 ORDER BY fact.revision_id, fact.field_key
`, revisionIDs(byRevisionID))
	if err != nil {
		return nil, fmt.Errorf("query record revision conflict facts: %w", err)
	}
	defer factRows.Close()
	for factRows.Next() {
		var revisionID int64
		var fact RevisionConflictFact
		if err := factRows.Scan(&revisionID, &fact.FieldKey, &fact.BeforePresent, &fact.BeforeValue, &fact.AfterPresent, &fact.AfterValue); err != nil {
			return nil, fmt.Errorf("scan record revision conflict fact: %w", err)
		}
		index, ok := byRevisionID[revisionID]
		if !ok {
			return nil, errInvalidRevisionWindow
		}
		result[index].ConflictFacts = append(result[index].ConflictFacts, fact)
	}
	if err := factRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record revision conflict facts: %w", err)
	}
	return result, nil
}

func revisionIDs(indexes map[int64]int) []int64 {
	result := make([]int64, 0, len(indexes))
	for revisionID := range indexes {
		result = append(result, revisionID)
	}
	return result
}
