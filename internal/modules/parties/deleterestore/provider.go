package deleterestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

const ActiveIncomingPartyReferenceReason = "active_incoming_party_reference"

type Source struct{}

var _ deleterestorecontract.DeleteRestoreSource = Source{}

func NewSource() Source {
	return Source{}
}

func (Source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(p))
  FROM records r
  JOIN parties p
    ON p.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (Source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (Source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.parties.v1", nil
}

func (Source) ValidateDeletePreconditionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (string, bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
      JOIN records r
        ON r.incident_id = e.incident_id
       AND r.record_id = e.record_id
       AND r.deleted_at IS NULL
     WHERE e.incident_id = $1
       AND (e.collector_party_id = $2 OR e.source_party_id = $2)
    UNION ALL
    SELECT 1
      FROM task_requests t
      JOIN records r
        ON r.incident_id = t.incident_id
       AND r.record_id = t.record_id
       AND r.deleted_at IS NULL
     WHERE t.incident_id = $1
       AND t.requester_party_id = $2
    UNION ALL
    SELECT 1
      FROM active_record_links_v1 rl
      JOIN records src
        ON src.incident_id = rl.incident_id
       AND src.record_id = rl.src_record_id
       AND src.deleted_at IS NULL
     WHERE rl.incident_id = $1
       AND rl.dst_record_id = $2
       AND rl.link_type = 'references_record'
       AND rl.field_key IN ('comm_log.audience_party_ids', 'comm_log.attendee_party_ids')
)
`, incidentID, recordID).Scan(&exists); err != nil {
		return "", false, fmt.Errorf("validate party delete references: %w", err)
	}
	if exists {
		return ActiveIncomingPartyReferenceReason, true, nil
	}
	return "", false, nil
}
