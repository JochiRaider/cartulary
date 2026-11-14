package incidentbundle

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

const assessmentsBundlePath = "data/compromise_assessments.ndjson"

func NewSourcePort() sourceport.Port {
	binding := &preparationBinding{contractMajor: sourceport.ContractMajor}
	descriptor := sourceport.Descriptor{
		FamilyID: "assessments", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.assessments", OwnerRelationIDs: []string{"assessment-source"},
		Dependencies: []string{"evidence"},
		Paths: []sourceport.Path{{
			LogicalPath: assessmentsBundlePath, ContentRole: "source_rows",
			Versions: []int{8}, StableIdentity: []string{"record_id"},
			StableIdentityInvariantID: "assessments.source_identity_admitted",
		}},
		InvariantIDs: []string{
			"assessments.source_identity_admitted", "assessments.row_shape_exact",
			"assessments.subject_type_scope", "assessments.state_confidence_rationale_legal",
			"assessments.timestamps_lifecycle_legal", "assessments.attribution_exact",
			"assessments.source_state_exact",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export:     exportIncidentBundleFiles,
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext) (any, error) {
			return prepare(ctx, bundle, input, descriptor, binding)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, input sourceport.ImportContext) error {
			prepared, ok := value.(preparedImport)
			if !ok || !prepared.matches(input, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			return prepared.apply(ctx, tx, input, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, input sourceport.ImportContext) error {
			prepared, ok := value.(preparedImport)
			if !ok || !prepared.matches(input, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			return prepared.validate(ctx, tx, input, descriptor)
		},
	})
}
