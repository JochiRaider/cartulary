package savedviews

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "saved_views", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.savedviews", OwnerRelationIDs: []string{"savedviews"},
		Dependencies: []string{"revisions"},
		Paths: []sourceport.Path{{
			LogicalPath: "data/saved_views.ndjson", ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.saved_views.row.v1",
			Versions: []int{1, 2}, StableIdentity: []string{"saved_view_id"},
		}},
		InvariantIDs: []string{
			"saved_views.row_shape_exact", "saved_views.identity_scope_legal", "saved_views.owner_tuple_legal",
			"saved_views.display_name_normalized", "saved_views.query_layout_legal",
			"saved_views.version_timestamps_legal", "saved_views.reference_pack_degradation_bounded",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) ([]incidentportability.File, error) {
			return exportIncidentBundleFiles(ctx, savedViewExportContext{
				Query:                exportContext.Query,
				IncidentID:           exportContext.IncidentID,
				PortableAttributions: exportContext.PortableAttributions,
			})
		},
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			prepared, err := prepareSavedViewImport(bundle, savedViewPrivateImportContext(importContext))
			return prepared, savedViewSourcePortError(err)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedSavedViewImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return savedViewSourcePortError(applyPreparedSavedViewImportTx(
				ctx,
				tx,
				prepared,
				savedViewPrivateImportContext(importContext),
			))
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedSavedViewImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return savedViewSourcePortError(validatePreparedSavedViewImportTx(
				ctx,
				tx,
				prepared,
				savedViewPrivateImportContext(importContext),
			))
		},
	})
}

func savedViewPrivateImportContext(importContext sourceport.ImportContext) savedViewImportContext {
	return savedViewImportContext{
		IncidentID:   importContext.IncidentID,
		ActorUserID:  importContext.ActorUserID,
		Attributions: importContext.Attributions,
		ActorAdmitted: func(actorID string) bool {
			_, ok := importContext.Actors.Lookup(actorID)
			return ok
		},
	}
}

func savedViewSourcePortError(err error) error {
	if err == nil {
		return nil
	}
	var invariantFailure *savedViewInvariantError
	if errors.As(err, &invariantFailure) {
		return &sourceport.Failure{
			FamilyID:    "saved_views",
			InvariantID: invariantFailure.InvariantID,
		}
	}
	return err
}
