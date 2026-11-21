package records

import (
	"context"
	"errors"

	"github.com/google/uuid"
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
			Versions: []int{8}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "records.source_identity_admitted",
		}},
		InvariantIDs: []string{
			"records.incident_scope", "records.envelope_legal",
			"records.subtype_complete", "records.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor:       descriptor,
		ValidateContract: subtypeCatalog.ValidateContract,
		Export: func(ctx context.Context, input sourceport.ExportContext) error {
			return exportIncidentBundleFiles(ctx, recordsExportContext{Actors: input.Actors, Query: input.Query, IncidentID: input.IncidentID, PortableAttributions: input.PortableAttributions}, input.Members)
		},
		Prepare: func(
			ctx context.Context,
			bundle sourceport.Bundle,
			importContext sourceport.ImportContext,
		) (any, error) {
			prepared, err := prepareRecordsImport(ctx,
				bundle,
				recordsPrivateImportContext(importContext),
			)
			if err != nil {
				return prepared, recordsPortError(descriptor, err)
			}
			return prepared, publishRecordAttribution(ctx, importContext, prepared)
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
			return recordsPortError(descriptor, applyPreparedRecordsImportTx(
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
			return recordsPortError(descriptor, validatePreparedRecordsImportTx(
				ctx,
				tx,
				prepared,
				recordsPrivateImportContext(importContext),
				subtypeCatalog,
			))
		},
		// Attribution was staged while preparing; the core flushes it in step 9.
		Attributions: func(value any) []incidentportability.AttributionStream {
			prepared, ok := value.(preparedRecordsImport)
			if !ok {
				return nil
			}
			return []incidentportability.AttributionStream{prepared.attribution}
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
		Preparation: importContext.Preparation,
	}
}

func recordsPortError(descriptor sourceport.Descriptor, err error) error {
	if err == nil {
		return nil
	}
	var invariantFailure *recordsInvariantError
	if errors.As(err, &invariantFailure) {
		return descriptor.DeclaredFailure(invariantFailure.InvariantID)
	}
	return err
}

// publishRecordAttribution publishes every admitted historical envelope actor
// as a read-only fact for later families' preparation (Core 01 REQ-01-641),
// so their cross-family actor checks are proven before the final transaction.
func publishRecordAttribution(ctx context.Context, importContext sourceport.ImportContext, prepared preparedRecordsImport) error {
	if importContext.Facts == nil {
		return nil
	}
	return importContext.Facts.Publish(ctx, sourceport.RecordAttributionFact, func(add func(string) error) error {
		return incidentportability.ReadPreparedRows(ctx, importContext.Preparation, prepared.rows, recordRowCodec{}, func(row recordsPortableRow) error {
			for _, attribution := range []struct {
				column string
				actor  *uuid.UUID
			}{{"created_by_user_id", &row.PortableCreatedByID}, {"updated_by_user_id", &row.PortableUpdatedByID}, {"deleted_by_user_id", row.PortableDeletedByID}} {
				if attribution.actor == nil {
					continue
				}
				if err := add(sourceport.RecordAttributionKey(row.RecordID, attribution.column, *attribution.actor)); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
