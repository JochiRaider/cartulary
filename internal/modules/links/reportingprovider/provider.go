package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectSupportRefsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string][]string, error) {
	rows, err := tx.Query(ctx, `
SELECT src_record_id::text, dst_record_id::text
  FROM record_links
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND link_type IN ('supported_by', 'references_record', 'attached_evidence')
 ORDER BY src_record_id::text ASC, dst_record_id::text ASC
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var src string
		var dst string
		if err := rows.Scan(&src, &dst); err != nil {
			return nil, err
		}
		out[src] = append(out[src], "/record_envelopes/"+dst)
	}
	return out, rows.Err()
}

func CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) ([]exportprovider.Field, error) {
	return exportprovider.CollectQueryFieldsTx(ctx, tx, incidentID, supportRefs, []exportprovider.FieldQuery{
		{
			Prefix: "relationships",
			SQL: `SELECT rl.record_link_id::text, 'record_link'::text, 'derived_analytic'::text, to_jsonb(rl) - 'incident_id'
  FROM record_links rl
 WHERE rl.incident_id = $1 AND rl.deleted_at IS NULL`,
		},
		{
			Prefix: "tags",
			SQL: `SELECT rt.record_tag_id::text, 'record_tag'::text, 'derived_analytic'::text, to_jsonb(rt) - 'incident_id'
  FROM record_tags rt
 WHERE rt.incident_id = $1 AND rt.deleted_at IS NULL`,
		},
	})
}
