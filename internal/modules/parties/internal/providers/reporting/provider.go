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
	output, err := exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "parties", supportRefs, []exportprovider.FieldQuery{{
		Prefix:                       "parties",
		DisclosurePartitionRefPrefix: "party:",
		SQL: `SELECT p.record_id::text, 'party'::text, r.record_type::text, 'source_evidence'::text, jsonb_build_object(
       'record_id', p.record_id,
       'display_name', p.display_name,
       'party_kind', p.party_kind,
       'organization_name', p.organization_name,
       'role_title', p.role_title,
       'primary_email', p.primary_email,
       'timezone_name', p.timezone_name,
       'external_ref', p.external_ref,
       'notes', p.notes,
       'created_at', r.created_at,
       'updated_at', r.updated_at
   )
  FROM parties p
  JOIN records r ON r.incident_id = p.incident_id AND r.record_id = p.record_id AND r.deleted_at IS NULL
 WHERE p.incident_id = $1`,
	}})
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	// Record binding comes from the Party-owned query above, in the caller's
	// snapshot transaction. The current source has no affirmative directory
	// eligibility field; absence is false, never inferred from display content.
	for _, record := range output.RecordFacts {
		output.SourceFacts.Parties = append(output.SourceFacts.Parties, exportprovider.PartyFact{PartyID: record.RecordID, PartitionSegment: record.RecordID, PublicDirectoryEligible: false})
	}
	output.SourceFacts.Sort()
	if err := output.Validate(); err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	return output, nil
}

var _ exportprovider.FieldProvider = (*Provider)(nil)
