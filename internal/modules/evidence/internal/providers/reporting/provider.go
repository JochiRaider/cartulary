package reporting

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type fieldContribution struct{}

func NewFieldContribution() exportprovider.FieldProvider {
	return fieldContribution{}
}

func (fieldContribution) ProviderKey() string { return "evidence" }

func (fieldContribution) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "evidence", supportRefs, []exportprovider.FieldQuery{{
		Prefix: "evidence",
		SQL: `SELECT e.record_id::text, 'evidence'::text, 'source_evidence'::text,
       jsonb_build_object(
           'record_id', e.record_id,
           'title', e.title,
           'lifecycle_state', e.lifecycle_state,
           'requested_at', e.requested_at,
           'received_at', e.received_at,
           'collector_party_text', e.collector_party_text,
           'collector_party_id', e.collector_party_id,
           'source_party_text', e.source_party_text,
           'source_party_id', e.source_party_id,
           'upload_state', e.upload_state,
           'created_at', e.created_at,
           'updated_at', e.updated_at
       )
  FROM evidence e
  JOIN records r ON r.incident_id = e.incident_id AND r.record_id = e.record_id AND r.deleted_at IS NULL
 WHERE e.incident_id = $1`,
	}})
}

type logicalTargetContribution struct{}

func NewLogicalTargetContribution() exportprovider.LogicalSupportTargetProvider {
	return logicalTargetContribution{}
}

func (logicalTargetContribution) ProviderKey() string { return "evidence" }

func (logicalTargetContribution) CollectLogicalSupportTargetsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) (map[string]string, error) {
	rows, err := tx.Query(ctx, `
SELECT e.record_id::text
  FROM evidence e
  JOIN records r
    ON r.incident_id = e.incident_id
   AND r.record_id = e.record_id
   AND r.deleted_at IS NULL
 WHERE e.incident_id = $1
 ORDER BY e.record_id::text ASC
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	targets := map[string]string{}
	for rows.Next() {
		var recordID string
		if err := rows.Scan(&recordID); err != nil {
			return nil, err
		}
		targets[recordID] = "/evidence/" + recordID
	}
	return targets, rows.Err()
}
