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
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "records", supportRefs, []exportprovider.FieldQuery{{
		Prefix: "record_envelopes",
		SQL: `SELECT r.record_id::text, 'record_envelope'::text, 'derived_analytic'::text, to_jsonb(r) - 'incident_id'
  FROM records r
 WHERE r.incident_id = $1 AND r.deleted_at IS NULL`,
	}})
}
