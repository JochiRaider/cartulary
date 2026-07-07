package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) ([]exportprovider.Field, error) {
	output, err := CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		return nil, err
	}
	return output.Fields(), nil
}

func CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "evidence", supportRefs, []exportprovider.FieldQuery{{
		Prefix: "evidence",
		SQL: `SELECT e.record_id::text, 'evidence'::text, 'source_evidence'::text, to_jsonb(e) - 'incident_id' - 'blob_hash' - 'storage_ref' - 'object_blob_id'
  FROM evidence e
  JOIN records r ON r.incident_id = e.incident_id AND r.record_id = e.record_id AND r.deleted_at IS NULL
 WHERE e.incident_id = $1`,
	}})
}
