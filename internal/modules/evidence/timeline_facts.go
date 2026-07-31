package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TimelineFact is Evidence-owned source data consumed by Timeline composition.
// It intentionally contains no Timeline projection type.
type TimelineFact struct {
	RecordID       uuid.UUID
	Title          string
	LifecycleState string
	UploadState    string
}

type TimelineFactReader struct{}

func (TimelineFactReader) LoadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) ([]TimelineFact, error) {
	return (sourceReadRepository{}).loadTimelineFactsTx(ctx, tx, incidentID, recordIDs)
}

// sourceReadRepository contains transaction-supplied source reads used by
// Evidence-owned consumer contributions.
type sourceReadRepository struct{}

func (sourceReadRepository) loadTimelineFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) ([]TimelineFact, error) {
	if len(recordIDs) == 0 {
		return []TimelineFact{}, nil
	}
	rows, err := tx.Query(ctx, `
SELECT
    ev.record_id,
    COALESCE(ev.title, ev.record_id::text) AS title,
    ev.lifecycle_state,
    COALESCE(b.upload_state, ev.upload_state, 'pending') AS upload_state
  FROM evidence ev
  LEFT JOIN object_blobs b ON b.object_blob_id = ev.object_blob_id
 WHERE ev.incident_id = $1
   AND ev.record_id = ANY($2)
 ORDER BY COALESCE(ev.title, ev.record_id::text) ASC, ev.record_id ASC
`, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load timeline evidence facts: %w", err)
	}
	defer rows.Close()
	facts := make([]TimelineFact, 0, len(recordIDs))
	for rows.Next() {
		var fact TimelineFact
		if err := rows.Scan(&fact.RecordID, &fact.Title, &fact.LifecycleState, &fact.UploadState); err != nil {
			return nil, fmt.Errorf("scan timeline evidence fact: %w", err)
		}
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline evidence facts: %w", err)
	}
	return facts, nil
}
