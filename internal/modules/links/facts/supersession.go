package facts

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SupersessionFact struct {
	SourceRecordID      uuid.UUID
	DestinationRecordID uuid.UUID
}

// LoadRetainedSupersessionTx includes endpoint-hidden relations, but excludes a
// relation's own tombstone. It never changes endpoint or relation state.
func (FactReader) LoadRetainedSupersessionTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID) ([]SupersessionFact, error) {
	rows, err := tx.Query(ctx, `SELECT src_record_id,dst_record_id FROM record_links
WHERE incident_id=$1 AND (src_record_id=$2 OR dst_record_id=$2)
  AND link_type='supersedes' AND deleted_at IS NULL ORDER BY record_link_id`, incidentID, recordID)
	if err != nil {
		return nil, &FactReadError{err: err}
	}
	defer rows.Close()
	result := make([]SupersessionFact, 0)
	for rows.Next() {
		var fact SupersessionFact
		if err := rows.Scan(&fact.SourceRecordID, &fact.DestinationRecordID); err != nil {
			return nil, &FactReadError{err: err}
		}
		result = append(result, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, &FactReadError{err: err}
	}
	return result, nil
}
