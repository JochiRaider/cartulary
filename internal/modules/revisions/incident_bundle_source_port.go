package revisions

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "revisions", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.revisions", OwnerRelationIDs: []string{"record-revisions"},
		Dependencies: []string{"links_tags"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/change_sets.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"change_set_id"}},
			{LogicalPath: "data/change_set_mutations.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"change_set_id", "sequence_no"}},
			{LogicalPath: "data/record_revisions.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"revision_id"}},
		},
		InvariantIDs: []string{
			"revisions.references_complete", "revisions.actor_references_complete",
			"revisions.mutation_sequence_contiguous", "revisions.record_version_unique",
			"revisions.history_reconstruction", "revisions.sequence_repair_after_validation",
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
    SELECT 1
      FROM change_sets change_set
     WHERE change_set.incident_id = $1
       AND EXISTS (
           SELECT 1
             FROM change_set_mutations mutation
            WHERE mutation.change_set_id = change_set.change_set_id
            GROUP BY mutation.change_set_id
           HAVING min(mutation.sequence_no) <> 1
               OR max(mutation.sequence_no) <> count(*)
       )
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "revisions", InvariantID: "revisions.mutation_sequence_contiguous"}
			}
			return nil
		},
	})
}
