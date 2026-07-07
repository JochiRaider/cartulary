package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) ([]exportprovider.Field, error) {
	return exportprovider.CollectQueryFieldsTx(ctx, tx, incidentID, supportRefs, []exportprovider.FieldQuery{{
		Prefix: "record_envelopes",
		SQL: `SELECT r.record_id::text, 'record_envelope'::text, 'derived_analytic'::text, to_jsonb(r) - 'incident_id'
  FROM records r
 WHERE r.incident_id = $1 AND r.deleted_at IS NULL`,
	}})
}
