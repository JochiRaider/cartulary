package incidentbundle

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
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
		Descriptor: descriptor, Export: sourceport.QueryExport(exportFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareIndicatorFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return importFilesTx(ctx, tx, map[string][]byte(value.(sourceport.PreparedFiles)), importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
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

func prepareIndicatorFiles(descriptor sourceport.Descriptor, bundle sourceport.Bundle, bundleVersion int) (sourceport.PreparedFiles, error) {
	prepared, err := sourceport.PrepareFiles(descriptor, bundle, bundleVersion)
	if err != nil {
		return nil, err
	}
	rows, err := incidentportability.DecodeNDJSON(prepared["data/indicators.ndjson"])
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		canonical, err := identity.Canonicalize(identity.Input{
			IndicatorType:   portableRequiredString(row, "indicator_type"),
			ValueKind:       portableRequiredString(row, "value_kind"),
			DisplayValue:    portableRequiredString(row, "display_value"),
			NormalizedValue: portableNullableString(row, "normalized_value"),
			DefangedValue:   portableNullableString(row, "defanged_value"),
			HashAlgorithm:   portableNullableString(row, "hash_algorithm"),
			HashValue:       portableNullableString(row, "hash_value"),
			STIXPattern:     portableNullableString(row, "stix_pattern"),
		})
		if err != nil {
			return nil, indicatorSourceFailure("indicators.representation_legal")
		}
		if !portableIdentityIsCanonical(row, canonical) {
			return nil, indicatorSourceFailure("indicators.normalization_exact")
		}
		key := canonical.IndicatorType + "\x00" + canonical.DedupeKey
		if _, duplicate := seen[key]; duplicate {
			return nil, indicatorSourceFailure("indicators.identity_unique")
		}
		seen[key] = struct{}{}
	}
	observationRows, err := incidentportability.DecodeNDJSON(prepared["data/indicator_observations.ndjson"])
	if err != nil {
		return nil, err
	}
	for _, row := range observationRows {
		raw, ok := row["origin_kind"].(string)
		if !ok {
			return nil, indicatorSourceFailure("indicators.representation_legal")
		}
		if _, err := indicatororigin.Parse(raw); err != nil {
			return nil, indicatorSourceFailure("indicators.representation_legal")
		}
	}
	return prepared, nil
}

func portableIdentityIsCanonical(row map[string]any, canonical identity.Canonical) bool {
	return portableRequiredString(row, "indicator_type") == canonical.IndicatorType &&
		portableRequiredString(row, "value_kind") == canonical.ValueKind &&
		portableRequiredString(row, "display_value") == canonical.DisplayValue &&
		portableStringPointersEqual(portableNullableString(row, "normalized_value"), canonical.NormalizedValue) &&
		portableStringPointersEqual(portableNullableString(row, "hash_algorithm"), canonical.HashAlgorithm) &&
		portableStringPointersEqual(portableNullableString(row, "hash_value"), canonical.HashValue) &&
		portableRequiredString(row, "dedupe_key") == canonical.DedupeKey
}

func portableRequiredString(row map[string]any, key string) string {
	value, _ := row[key].(string)
	return value
}

func portableNullableString(row map[string]any, key string) *string {
	raw, present := row[key]
	if !present || raw == nil {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	return &value
}

func indicatorSourceFailure(invariantID string) error {
	return &sourceport.Failure{FamilyID: "indicators", InvariantID: invariantID}
}

func portableStringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
