package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) ([]exportprovider.Field, error) {
	return exportprovider.CollectQueryFieldsTx(ctx, tx, incidentID, supportRefs, []exportprovider.FieldQuery{
		{
			Prefix: "notes",
			SQL: `SELECT a.record_id::text, 'note'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'note'`,
		},
		{
			Prefix: "findings",
			SQL: `SELECT a.record_id::text, 'finding_hypothesis'::text,
       CASE WHEN a.finding_kind = 'finding' THEN 'curated_narrative'::text ELSE 'working_material'::text END,
       to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'finding'`,
		},
		{
			Prefix: "comm_log",
			SQL: `SELECT a.record_id::text, 'comm_log'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'comm_log'`,
		},
		{
			Prefix: "handoffs",
			SQL: `SELECT a.record_id::text, 'handoff'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'handoff'`,
		},
		{
			Prefix: "status_reviews",
			SQL: `SELECT a.record_id::text, 'status_review'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'status_review'`,
		},
		{
			Prefix: "lessons",
			SQL: `SELECT a.record_id::text, 'lesson'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'lesson'`,
		},
	})
}
