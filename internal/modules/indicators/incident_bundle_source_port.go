package indicators

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "indicators", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.indicators", OwnerRelationIDs: []string{"indicator-source"},
		Dependencies: []string{"entities"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/indicators.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}},
			{LogicalPath: "data/indicator_observations.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"indicator_observation_id"}},
			{LogicalPath: "data/indicator_state_intervals.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"indicator_state_interval_id"}},
		},
		InvariantIDs: []string{
			"indicators.representation_legal", "indicators.normalization_exact",
			"indicators.identity_unique", "indicators.observation_same_incident",
			"indicators.observation_ordered", "indicators.observation_coherent",
			"indicators.interval_same_incident", "indicators.interval_ordered",
			"indicators.interval_coherent", "indicators.repeated_observations_preserved",
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
    SELECT 1 FROM indicators indicator
    LEFT JOIN records record ON record.record_id = indicator.record_id
    WHERE indicator.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'indicator')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "indicators", InvariantID: "indicators.representation_legal"}
			}
			return nil
		},
	})
}
