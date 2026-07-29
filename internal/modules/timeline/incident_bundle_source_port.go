package timeline

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "timeline", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.timeline", OwnerRelationIDs: []string{"timeline-source"},
		Dependencies: []string{"records"},
		Paths: []sourceport.Path{
			{LogicalPath: timelineBundleV1ProfilesPath, ContentRole: "source_rows", Versions: []int{1}, StableIdentity: []string{"incident_id"}},
			{LogicalPath: timelineBundleV1RecordsPath, ContentRole: "source_rows", Versions: []int{1}, StableIdentity: []string{"record_id"}},
			{LogicalPath: timelineBundleProfilesPath, ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"incident_id"}},
			{LogicalPath: timelineBundleRecordsPath, ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"record_id"}},
			{LogicalPath: timelineBundleProvenancePath, ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"record_id", "source_row_ordinal", "source_column_ordinal", "source_identity_sha256"}},
		},
		InvariantIDs: []string{
			"timeline.version_shape_exact", "timeline.envelope_type_scope",
			"timeline.lifecycle_coherent", "timeline.generated_time_coherent",
			"timeline.paired_time_coherent", "timeline.provenance_unique",
			"timeline.provenance_non_orphaned", "timeline.v1_translation_lossless",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(ExportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return ImportIncidentBundleFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.BundleVersion, importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM timeline_events event
      LEFT JOIN records record ON record.record_id = event.record_id
     WHERE event.incident_id = $1
       AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'timeline_event')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "timeline", InvariantID: "timeline.envelope_type_scope"}
			}
			return nil
		},
	})
}
