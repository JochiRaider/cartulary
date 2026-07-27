package savedviews

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "saved_views", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.savedviews", OwnerRelationIDs: []string{"savedviews"},
		Dependencies: []string{"revisions"},
		Paths:        []sourceport.Path{{LogicalPath: "data/saved_views.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"saved_view_id"}}},
		InvariantIDs: []string{
			"saved_views.identity_scope_legal", "saved_views.owner_tuple_legal",
			"saved_views.display_name_normalized", "saved_views.query_layout_legal",
			"saved_views.version_timestamps_legal", "saved_views.reference_pack_degradation_bounded",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: ExportIncidentBundleFiles,
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM saved_views
     WHERE incident_id = $1
       AND ((scope IN ('private', 'shared') AND owner_user_id IS NULL)
         OR (scope = 'system' AND owner_user_id IS NOT NULL))
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "saved_views", InvariantID: "saved_views.owner_tuple_legal"}
			}
			return nil
		},
	})
}
