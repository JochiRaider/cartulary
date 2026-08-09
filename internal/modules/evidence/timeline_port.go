package evidence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type timelineAttachmentReader struct{}

func (timelineAttachmentReader) ValidateTimelineAttachmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) error {
	for _, recordID := range recordIDs {
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
      JOIN records r
        ON r.record_id = e.record_id
       AND r.incident_id = e.incident_id
     WHERE e.incident_id = $1
       AND e.record_id = $2
       AND r.record_type = 'evidence'
       AND r.deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrEvidenceNotFound
		}
	}
	return nil
}
