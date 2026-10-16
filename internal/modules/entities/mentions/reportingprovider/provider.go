package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type provider struct{}

func New() exportprovider.FieldProvider { return provider{} }

func (provider) ProviderKey() string { return "entities.mentions" }

func (p provider) CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, p.ProviderKey(), supportRefs, []exportprovider.FieldQuery{{
		Prefix: "entity_mentions",
		SQL: `SELECT em.entity_mention_id::text, 'entity_mention'::text, 'entity_mention'::text, 'source_evidence'::text, jsonb_build_object(
       'entity_mention_id', em.entity_mention_id,
       'source_record_id', em.source_record_id,
       'entity_type', em.entity_type,
       'source_field_key', em.source_field_key,
       'origin_kind', em.origin_kind,
       'origin_locator', em.origin_locator,
       'raw_text', em.raw_text,
       'normalized_text', em.normalized_text,
       'resolution_status', em.resolution_status,
       'row_version', em.row_version,
       'ordinal', em.ordinal,
       'created_by_user_id', em.created_by_user_id,
       'created_at', em.created_at,
       'resolved_record_id', em.resolved_record_id,
       'resolved_by_user_id', em.resolved_by_user_id,
       'resolved_at', em.resolved_at,
       'resolution_method', em.resolution_method
   )
  FROM entity_mentions em
  JOIN records r ON r.record_id = em.source_record_id AND r.deleted_at IS NULL
 WHERE r.incident_id = $1`,
	}})
}
