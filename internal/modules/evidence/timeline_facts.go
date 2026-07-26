package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type TimelineFactReader struct{}

func (TimelineFactReader) LoadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) ([]workbookprojection.EvidenceFact, error) {
	if len(recordIDs) == 0 {
		return []workbookprojection.EvidenceFact{}, nil
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
	facts := make([]workbookprojection.EvidenceFact, 0, len(recordIDs))
	for rows.Next() {
		var fact workbookprojection.EvidenceFact
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
