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
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "entities.hostidentity", supportRefs, []exportprovider.FieldQuery{
		{
			Prefix: "hosts",
			SQL: `SELECT h.record_id::text, 'host'::text, 'derived_analytic'::text, to_jsonb(h) - 'incident_id'
  FROM host_grid_projection h
  JOIN records r ON r.incident_id = h.incident_id AND r.record_id = h.record_id AND r.deleted_at IS NULL
 WHERE h.incident_id = $1`,
		},
		{
			Prefix: "identities",
			SQL: `SELECT i.record_id::text, 'identity'::text, 'derived_analytic'::text, to_jsonb(i) - 'incident_id'
  FROM identity_grid_projection i
  JOIN records r ON r.incident_id = i.incident_id AND r.record_id = i.record_id AND r.deleted_at IS NULL
 WHERE i.incident_id = $1`,
		},
	})
}
