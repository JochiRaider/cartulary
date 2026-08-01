package reporting

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "tasksdecisions", supportRefs, []exportprovider.FieldQuery{
		{
			Prefix: "task_requests",
			SQL: `SELECT t.record_id::text, 'task_request'::text, 'working_material'::text, to_jsonb(t) - 'incident_id'
  FROM task_request_grid_projection t
  JOIN records r ON r.incident_id = t.incident_id AND r.record_id = t.record_id AND r.deleted_at IS NULL
 WHERE t.incident_id = $1`,
		},
		{
			Prefix: "decisions",
			SQL: `SELECT d.record_id::text, 'decision'::text, 'working_material'::text, to_jsonb(d) - 'incident_id'
  FROM decision_grid_projection d
  JOIN records r ON r.incident_id = d.incident_id AND r.record_id = d.record_id AND r.deleted_at IS NULL
 WHERE d.incident_id = $1`,
		},
	})
}
