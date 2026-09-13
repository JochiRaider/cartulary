package artifacts

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

const coordinationSourceInput = "coordination.source_record_id"

func isCoordinationView(view string) bool {
	switch view {
	case CommLogViewSchemaID, HandoffViewSchemaID, StatusReviewViewSchemaID, LessonViewSchemaID:
		return true
	default:
		return false
	}
}

// Source context is an association, never a semantic target field. The incident
// comes exclusively from the authorized route; source lookup cannot retarget it.
func (f *MutationFacade) validateCoordinationSourceTx(ctx context.Context, tx pgx.Tx, incidentID, sourceID uuid.UUID, targetView string) error {
	invalid := func() error { return &ValidationError{Field: coordinationSourceInput, ReasonCode: "invalid_value"} }
	envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, sourceID, true)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return invalid()
	}
	if err != nil {
		return err
	}
	if envelope.IncidentID != incidentID || envelope.DeletedAt != nil || !isCoordinationView(targetView) {
		return invalid()
	}
	switch envelope.RecordType {
	case "timeline_event":
		return nil
	case "task_request":
		if targetView != HandoffViewSchemaID {
			return nil
		}
	case "decision":
		if targetView == CommLogViewSchemaID || targetView == StatusReviewViewSchemaID {
			return nil
		}
	case "artifact":
		var subtype string
		if err := tx.QueryRow(ctx, `SELECT artifact_type FROM artifacts WHERE record_id=$1 AND incident_id=$2`, sourceID, incidentID).Scan(&subtype); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return invalid()
			}
			return err
		}
		if (targetView == StatusReviewViewSchemaID && (subtype == "comm_log" || subtype == "handoff")) || (targetView == CommLogViewSchemaID && subtype == "status_review") {
			return nil
		}
	}
	return invalid()
}
