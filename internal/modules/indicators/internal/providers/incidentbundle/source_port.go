package incidentbundle

import (
	"context"

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
		Export: func(ctx context.Context, input sourceport.ExportContext) ([]incidentportability.File, error) {
			evaluate, err := identity.FromConsumer(ctx, consumer)
			if err != nil {
				return nil, err
			}
			return exportFiles(evaluate, ctx, input)
		},
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			evaluate, err := identity.FromConsumer(ctx, consumer)
			if err != nil {
				return nil, err
			}
			return prepareIndicatorImport(evaluate, bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedIndicatorImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			if tx == nil || !prepared.binding.matches(importContext) {
				return indicatorSourceFailure(representationInvariant)
			}
			if len(prepared.indicators) == 0 && len(prepared.observations) == 0 {
				return applyPreparedIndicatorImportTx(ctx, tx, prepared, importContext)
			}
			assignment, err := assignments.BeginTx(ctx, tx)
			if err != nil {
				return err
			}
			evaluate, err := identity.FromConsumer(ctx, assignment)
			if err != nil {
				return err
			}
			for _, row := range prepared.indicators {
				if err := validatePortableIndicatorForExport(evaluate, row, importContext.IncidentID); err != nil {
					return indicatorSourceFailure(normalizationInvariant)
				}
			}
			for _, row := range prepared.observations {
				if err := validatePortableObservationForExport(evaluate, row, importContext.IncidentID); err != nil {
					return indicatorSourceFailure(normalizationInvariant)
				}
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
			evaluate, err := identity.FromConsumer(ctx, assignment)
			if err != nil {
				return err
			}
			return validatePreparedIndicatorImportTx(evaluate, ctx, tx, prepared, importContext)
		},
	})
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
