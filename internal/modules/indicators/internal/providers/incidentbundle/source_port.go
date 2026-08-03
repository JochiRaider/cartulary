package incidentbundle

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "indicators", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.indicators", OwnerRelationIDs: []string{"indicator-source"},
		Dependencies: []string{"entities"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/indicators.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.indicators.row.v1", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}},
			{LogicalPath: "data/indicator_observations.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.indicator_observations.row.v1", Versions: []int{1, 2}, StableIdentity: []string{"indicator_observation_id"}},
			{LogicalPath: "data/indicator_state_intervals.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.indicator_state_intervals.row.v1", Versions: []int{1, 2}, StableIdentity: []string{"indicator_state_interval_id"}},
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
		Descriptor: descriptor,
		Export:     exportFiles,
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareIndicatorImport(bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return applyPreparedIndicatorImportTx(ctx, tx, prepared, importContext)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return validatePreparedIndicatorImportTx(ctx, tx, prepared, importContext)
		},
	})
}

func indicatorSourceFailure(invariantID string) error {
	if indicatorInvariantRank(invariantID) == 0 {
		return errors.New("indicators portability invariant is undeclared")
	}
	return &sourceport.Failure{FamilyID: "indicators", InvariantID: invariantID}
}
