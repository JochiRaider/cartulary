package tasksdecisions

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "tasks_decisions", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.tasksdecisions", OwnerRelationIDs: []string{"tasks-and-decisions"},
		Dependencies: []string{"artifacts"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/task_requests.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}},
			{LogicalPath: "data/decisions.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}},
		},
		InvariantIDs: []string{
			"tasks_decisions.envelope_type_scope", "tasks_decisions.lifecycle_legal",
			"tasks_decisions.dependent_fields_legal", "tasks_decisions.references_same_incident",
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
    SELECT 1 FROM (
        SELECT task.record_id, task.incident_id, 'task_request'::text AS required_type FROM task_requests task
        UNION ALL
        SELECT decision.record_id, decision.incident_id, 'decision'::text FROM decisions decision
    ) item
    LEFT JOIN records record ON record.record_id = item.record_id
    WHERE item.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> item.required_type)
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return &sourceport.Failure{FamilyID: "tasks_decisions", InvariantID: "tasks_decisions.envelope_type_scope"}
			}
			return nil
		},
	})
}
