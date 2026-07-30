package records

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

func NewIncidentBundleSourcePort(subtypeCatalog *subtypepresence.Catalog) sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "records", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.records", OwnerRelationIDs: []string{"record-envelope"},
		Dependencies: []string{"incident"},
		Paths: []sourceport.Path{{
			LogicalPath: recordsBundlePath, ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.records.row.v1",
			Versions: []int{1, 2}, StableIdentity: []string{"record_id"},
		}},
		InvariantIDs: []string{
			"records.incident_scope", "records.envelope_legal",
			"records.subtype_complete",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor:       descriptor,
		ValidateContract: subtypeCatalog.ValidateContract,
		Export: func(
			ctx context.Context,
			exportContext sourceport.ExportContext,
		) ([]incidentportability.File, error) {
			return exportIncidentBundleFiles(ctx, recordsExportContext{
				Query:                exportContext.Query,
				IncidentID:           exportContext.IncidentID,
				PortableAttributions: exportContext.PortableAttributions,
			})
		},
		Prepare: func(
			_ context.Context,
			bundle sourceport.Bundle,
			importContext sourceport.ImportContext,
		) (any, error) {
			prepared, err := prepareRecordsImport(
				bundle,
				recordsPrivateImportContext(importContext),
			)
			return prepared, recordsPortError(err)
		},
		Apply: func(
			ctx context.Context,
			tx pgx.Tx,
			value any,
			importContext sourceport.ImportContext,
		) error {
			prepared, ok := value.(preparedRecordsImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return recordsPortError(applyPreparedRecordsImportTx(
				ctx,
				tx,
				prepared,
				recordsPrivateImportContext(importContext),
			))
		},
		Validate: func(
			ctx context.Context,
			tx pgx.Tx,
			value any,
			importContext sourceport.ImportContext,
		) error {
			prepared, ok := value.(preparedRecordsImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return recordsPortError(validatePreparedRecordsImportTx(
				ctx,
				tx,
				prepared,
				recordsPrivateImportContext(importContext),
				subtypeCatalog,
			))
		},
	})
}

func recordsPrivateImportContext(importContext sourceport.ImportContext) recordsImportContext {
	return recordsImportContext{
		IncidentID:   importContext.IncidentID,
		ActorUserID:  importContext.ActorUserID,
		Attributions: importContext.Attributions,
		ActorAdmitted: func(actorID string) bool {
			_, admitted := importContext.Actors.Lookup(actorID)
			return admitted
		},
	}
}

func recordsPortError(err error) error {
	if err == nil {
		return nil
	}
	var invariantFailure *recordsInvariantError
	if errors.As(err, &invariantFailure) {
		return &sourceport.Failure{
			FamilyID:    "records",
			InvariantID: invariantFailure.InvariantID,
		}
	}
	return err
}
