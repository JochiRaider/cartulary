package incidentbundle

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewSourcePort(paths []sourceport.Path) sourceport.Port {
	descriptor := indicatorSourceDescriptor(paths)
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
