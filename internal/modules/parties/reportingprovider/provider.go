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
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "parties", supportRefs, []exportprovider.FieldQuery{{
		Prefix:                       "parties",
		DisclosurePartitionRefPrefix: "party:",
		SQL: `SELECT p.record_id::text, 'party'::text, 'source_evidence'::text, to_jsonb(p) - 'incident_id'
  FROM parties p
  JOIN records r ON r.incident_id = p.incident_id AND r.record_id = p.record_id AND r.deleted_at IS NULL
 WHERE p.incident_id = $1`,
	}})
}
