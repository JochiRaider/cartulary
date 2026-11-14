package artifacts

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() (sourceport.Port, error) {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return nil, err
	}
	binding := &artifactPreparationBinding{contractMajor: sourceport.ContractMajor}
	descriptor := sourceport.Descriptor{
		FamilyID: "artifacts", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.artifacts", OwnerRelationIDs: []string{"artifacts-and-optional-surfaces"},
		Dependencies: []string{"indicators"},
		Paths:        manifest.sourcePortPaths(),
		InvariantIDs: []string{
			"artifacts.source_identity_admitted", "artifacts.row_shape_exact",
			"artifacts.envelope_type_scope", "artifacts.subtype_exact",
			"artifacts.lifecycle_fields_legal", "artifacts.handoff_risk_target",
			"artifacts.handoff_risk_legal", "artifacts.references_same_incident",
			"artifacts.attribution_exact", "artifacts.source_state_exact",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: exportIncidentBundleFiles,
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareArtifactImport(ctx, bundle, importContext, descriptor, manifest, binding)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedArtifactImport)
			if !ok || !prepared.matches(importContext, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			return prepared.apply(ctx, tx, importContext, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedArtifactImport)
			if !ok || !prepared.matches(importContext, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			return prepared.validate(ctx, tx, importContext, descriptor)
		},
	}), nil
}
