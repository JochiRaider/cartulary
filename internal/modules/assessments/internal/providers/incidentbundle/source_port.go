package incidentbundle

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

const assessmentsBundlePath = "data/compromise_assessments.ndjson"

func NewSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "assessments", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.assessments", OwnerRelationIDs: []string{"assessment-source"},
		Dependencies: []string{"evidence"},
		Paths: []sourceport.Path{{
			LogicalPath: assessmentsBundlePath, ContentRole: "source_rows",
			Versions: []int{3}, StableIdentity: []string{"record_id"},
			StableIdentityInvariantID: "assessments.source_identity_admitted",
		}},
		InvariantIDs: []string{
			"assessments.subject_type_scope",
			"assessments.state_confidence_rationale_legal",
			"assessments.timestamps_lifecycle_legal",
			"assessments.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export:     sourceport.QueryExport(exportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return importIncidentBundleFilesTx(
				ctx,
				tx,
				map[string][]byte(value.(sourceport.PreparedFiles)),
				importContext.ActorUserID,
				importContext.Attributions,
			)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM assessments assessment
    LEFT JOIN records record ON record.record_id = assessment.record_id
    LEFT JOIN records subject ON subject.record_id = assessment.subject_record_id
    WHERE assessment.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'assessment'
           OR subject.record_id IS NULL OR subject.incident_id <> $1 OR subject.record_type <> assessment.subject_type)
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("assessments.subject_type_scope")
			}
			return nil
		},
	})
}
