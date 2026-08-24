package incidentbundle

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewContribution() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "parties", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.parties", OwnerRelationIDs: []string{"parties"},
		Dependencies: []string{"timeline"},
		Paths: []sourceport.Path{{
			LogicalPath: partyIncidentBundlePath, ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.parties.row.v1",
			Versions: []int{3}, StableIdentity: []string{"record_id"},
			StableIdentityInvariantID: "parties.source_identity_admitted",
		}},
		InvariantIDs: []string{
			"parties.source_identity_admitted",
			"parties.version_shape_exact",
			"parties.envelope_type_scope",
			"parties.identity_lifecycle",
			"parties.normalization_exact",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export:     sourceport.QueryExport(exportIncidentBundleFiles),
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return preparePartyImport(ctx, bundle, importContext, descriptor)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedPartyImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return applyPreparedPartyImportTx(ctx, tx, prepared, importContext, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedPartyImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return validatePreparedPartyImportTx(ctx, tx, prepared, importContext, descriptor)
		},
	})
}
