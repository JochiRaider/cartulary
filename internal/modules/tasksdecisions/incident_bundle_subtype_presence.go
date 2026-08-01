package tasksdecisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

type incidentBundleSubtypeSource struct{}

func incidentBundleSubtypeContribution() subtypepresence.Contribution {
	return subtypepresence.Contribution{
		FamilyID: "tasks_decisions",
		Source:   incidentBundleSubtypeSource{},
	}
}

func (incidentBundleSubtypeSource) SupportedRecordTypes() []subtypepresence.RecordType {
	return []subtypepresence.RecordType{
		subtypepresence.RecordTypeTaskRequest,
		subtypepresence.RecordTypeDecision,
	}
}

func (incidentBundleSubtypeSource) ListSubtypeBindingsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]subtypepresence.Binding, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, record_type
  FROM (
        SELECT record_id, incident_id, 'task_request'::text AS record_type FROM task_requests
        UNION ALL
        SELECT record_id, incident_id, 'decision'::text FROM decisions
  ) AS coordination_bindings
 WHERE incident_id = $1
 ORDER BY record_id, record_type
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bindings []subtypepresence.Binding
	for rows.Next() {
		var binding subtypepresence.Binding
		if err := rows.Scan(&binding.RecordID, &binding.IncidentID, &binding.RecordType); err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}
