package reporting

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type Provider struct{}

func New() *Provider { return &Provider{} }

func (*Provider) ProviderKey() string { return "parties" }

func (*Provider) CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	return exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "parties", supportRefs, []exportprovider.FieldQuery{{
		Prefix:                       "parties",
		DisclosurePartitionRefPrefix: "party:",
		SQL: `SELECT p.record_id::text, 'party'::text, 'source_evidence'::text, to_jsonb(p) - 'incident_id'
  FROM parties p
  JOIN records r ON r.incident_id = p.incident_id AND r.record_id = p.record_id AND r.deleted_at IS NULL
 WHERE p.incident_id = $1`,
	}})
}

var _ exportprovider.FieldProvider = (*Provider)(nil)
