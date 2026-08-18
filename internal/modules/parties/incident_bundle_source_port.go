package parties

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "parties", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.parties", OwnerRelationIDs: []string{"parties"},
		Dependencies: []string{"timeline"},
		Paths:        []sourceport.Path{{LogicalPath: "data/parties.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "parties.source_identity_admitted"}},
		InvariantIDs: []string{"parties.envelope_type_scope", "parties.identity_lifecycle", "parties.normalization_exact", "parties.source_identity_admitted"},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(ExportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM parties party
    LEFT JOIN records record ON record.record_id = party.record_id
    WHERE party.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'party')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("parties.envelope_type_scope")
			}
			return nil
		},
	})
}
