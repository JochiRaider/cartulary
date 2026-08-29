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
		SQL: `SELECT em.entity_mention_id::text, 'entity_mention'::text, 'source_evidence'::text, to_jsonb(em)
  FROM entity_mentions em
  JOIN records r ON r.record_id = em.source_record_id AND r.deleted_at IS NULL
 WHERE r.incident_id = $1`,
	}})
}
