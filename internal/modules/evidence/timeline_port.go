package evidence

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
)

type timelineAttachmentReader struct {
	projections evidenceprojection.MutationRows
}

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

func (reader timelineAttachmentReader) RefreshTimelineAttachmentProjectionsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if reader.projections == nil {
		return errors.New("evidence projection rows are required")
	}
	unique := make(map[uuid.UUID]struct{}, len(recordIDs))
	for _, recordID := range recordIDs {
		if recordID != uuid.Nil {
			unique[recordID] = struct{}{}
		}
	}
	ordered := make([]uuid.UUID, 0, len(unique))
	for recordID := range unique {
		ordered = append(ordered, recordID)
	}
	slices.SortFunc(ordered, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	for _, recordID := range ordered {
		if err := reader.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
			return err
		}
	}
	return nil
}
