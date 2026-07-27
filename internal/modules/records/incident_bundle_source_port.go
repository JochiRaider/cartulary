package records

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "records", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.records", OwnerRelationIDs: []string{"record-revisions"},
		Dependencies: []string{"incident"},
		Paths: []sourceport.Path{{
			LogicalPath: "data/records.ndjson", ContentRole: "source_rows",
			Versions: []int{1, 2}, StableIdentity: []string{"record_id"},
		}},
		InvariantIDs: []string{
			"records.incident_scope", "records.envelope_legal",
			"records.subtype_complete",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export:     ExportIncidentBundleFiles,
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, importContext sourceport.ImportContext) error {
			var invalid bool
			err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM records
     WHERE incident_id <> $1
       AND record_id IN (SELECT record_id FROM records WHERE incident_id = $1)
)`, importContext.IncidentID).Scan(&invalid)
			if err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "records", InvariantID: "records.incident_scope"}
			}
			return nil
		},
	})
}
