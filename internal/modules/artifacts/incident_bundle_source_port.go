package artifacts

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() (sourceport.Port, error) {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return nil, err
	}
	descriptor := sourceport.Descriptor{
		FamilyID: "artifacts", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.artifacts", OwnerRelationIDs: []string{"artifacts-and-optional-surfaces"},
		Dependencies: []string{"indicators"},
		Paths:        manifest.sourcePortPaths(),
		InvariantIDs: []string{
			"artifacts.envelope_type_scope", "artifacts.subtype_exact",
			"artifacts.lifecycle_fields_legal", "artifacts.handoff_risk_target",
			"artifacts.references_same_incident", "artifacts.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(exportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return importIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM artifacts artifact
    LEFT JOIN records record ON record.record_id = artifact.record_id
    WHERE artifact.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'artifact')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("artifacts.envelope_type_scope")
			}
			return nil
		},
	}), nil
}
