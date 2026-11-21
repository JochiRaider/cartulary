package incidentbundle

import (
	"context"
	"errors"
	"reflect"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
)

func NewSourcePort(paths []sourceport.Path, consumer reference_data.Consumer, assignments reference_data.RegistryAssignments) sourceport.Port {
	descriptor := indicatorSourceDescriptor(paths)
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export: func(ctx context.Context, input sourceport.ExportContext) error {
			evaluate, err := identity.FromConsumer(ctx, consumer)
			if err != nil {
				return err
			}
			return exportFiles(evaluate, ctx, input)
		},
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			evaluate, set, err := identity.FromConsumerSet(ctx, consumer)
			if err != nil {
				return nil, err
			}
			prepared, err := prepareIndicatorImport(ctx, evaluate, bundle, importContext)
			if err != nil {
				return nil, err
			}
			prepared.packSet = set
			return prepared, nil
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			if tx == nil || !prepared.binding.matches(importContext) {
				return indicatorSourceFailure(representationInvariant)
			}
			if prepared.indicatorCount == 0 && prepared.observationCount == 0 {
				return applyPreparedIndicatorImportTx(ctx, tx, prepared, importContext)
			}
			assignment, err := assignments.BeginTx(ctx, tx)
			if err != nil {
				return err
			}
			evaluate, err := lockedEvaluator(ctx, assignment, prepared.packSet)
			if err != nil {
				return err
			}
			if evaluate == nil {
				// Preparation's canonical results stand under the same locked set.
				if err := applyPreparedIndicatorImportTx(ctx, tx, prepared, importContext); err != nil {
					return err
				}
				return assignment.RecordUsage(ctx, "type_registry.indicator")
			}
			prepared.evaluate = evaluate
			if err := prepared.indicators(ctx, importContext, func(row portableIndicatorRow) error {
				if err := validatePortableIndicatorForExport(evaluate, row, importContext.IncidentID); err != nil {
					return indicatorSourceFailure(normalizationInvariant)
				}

				return nil
			}); err != nil {
				return err
			}
			if err := prepared.observations(ctx, importContext, func(row portableObservationRow) error {
				if err := validatePortableObservationForExport(evaluate, row, importContext.IncidentID); err != nil {
					return indicatorSourceFailure(normalizationInvariant)
				}

				return nil
			}); err != nil {
				return err
			}
			if err := applyPreparedIndicatorImportTx(ctx, tx, prepared, importContext); err != nil {
				return err
			}
			return assignment.RecordUsage(ctx, "type_registry.indicator")
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			if tx == nil || !prepared.binding.matches(importContext) {
				return indicatorSourceFailure(representationInvariant)
			}
			assignment, err := assignments.BeginTx(ctx, tx)
			if err != nil {
				return err
			}
			evaluate, err := lockedEvaluator(ctx, assignment, prepared.packSet)
			if err != nil {
				return err
			}
			return validatePreparedIndicatorImportTx(evaluate, ctx, tx, prepared, importContext)
		},
		// Attribution was staged while preparing; the core flushes it in step 9.
		Attributions: func(value any) []incidentportability.AttributionStream {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return nil
			}
			return []incidentportability.AttributionStream{prepared.attribution}
		},
	})
}

// lockedEvaluator returns nil when the registry-locked current pack set equals
// the set preparation evaluated under: evaluation is a pure function of the
// content-addressed set and its request, so the prepared canonical results
// stand. Otherwise it returns an evaluator over the locked set, and the import
// re-evaluates.
func lockedEvaluator(ctx context.Context, assignment reference_data.RegistryAssignment, prepared reference_data.PackSet) (identity.Evaluator, error) {
	locked := assignment.ResolveCurrentPackSet(ctx)
	if locked.Error != nil {
		return nil, locked.Error
	}
	if locked.Value == nil {
		return nil, errors.New("indicator set unavailable")
	}
	if samePackSet(prepared, *locked.Value) {
		return nil, nil
	}
	return identity.FromConsumer(ctx, assignment)
}

func samePackSet(left, right reference_data.PackSet) bool {
	return left.ID != "" && left.ID == right.ID && left.SHA256 == right.SHA256 && left.SchemaID == right.SchemaID && reflect.DeepEqual(left.Members, right.Members)
}

func indicatorSourceDescriptor(paths []sourceport.Path) sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "indicators", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.indicators", OwnerRelationIDs: []string{"indicator-source"},
		Dependencies: []string{"entities"},
		Paths:        paths,
		InvariantIDs: []string{
			"indicators.representation_legal", "indicators.normalization_exact",
			"indicators.identity_unique", "indicators.observation_same_incident",
			"indicators.observation_ordered", "indicators.observation_coherent",
			"indicators.interval_same_incident", "indicators.interval_ordered",
			"indicators.interval_coherent", "indicators.repeated_observations_preserved",
			"indicators.source_identity_admitted",
		},
	}
}

func indicatorSourceFailure(invariantID string) error {
	return indicatorSourceDescriptor(nil).DeclaredFailure(invariantID)
}
