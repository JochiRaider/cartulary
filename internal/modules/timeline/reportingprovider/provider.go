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
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "timeline", supportRefs, []exportprovider.FieldQuery{{
		Prefix: "timeline",
		SQL: `SELECT t.record_id::text, 'timeline_event'::text, 'source_evidence'::text, to_jsonb(t) - 'incident_id'
  FROM timeline_events t
  JOIN records r ON r.incident_id = t.incident_id AND r.record_id = t.record_id AND r.deleted_at IS NULL
 WHERE t.incident_id = $1`,
	}})
}
