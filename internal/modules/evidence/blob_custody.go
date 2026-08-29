package evidence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type evidenceCustodyEventParams struct {
	IncidentID       uuid.UUID
	EvidenceRecordID uuid.UUID
	CustodyEventType string
	ActorUserID      *uuid.UUID
	OccurredAt       time.Time
	LocationText     *string
	Note             *string
	Metadata         map[string]any
}

func insertEvidenceCustodyEventTx(ctx context.Context, tx pgx.Tx, params evidenceCustodyEventParams) error {
	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO evidence_custody_events (
    incident_id,
    evidence_record_id,
    custody_event_type,
    actor_user_id,
    occurred_at,
    location_text,
    note,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
`, params.IncidentID, params.EvidenceRecordID, params.CustodyEventType, params.ActorUserID, params.OccurredAt.UTC(), params.LocationText, params.Note, metadataJSON)
	return err
}
