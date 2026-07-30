package indicators

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

type incidentBundleSubtypeSource struct{}

func IncidentBundleSubtypeContribution() subtypepresence.Contribution {
	return subtypepresence.Contribution{
		FamilyID: "indicators",
		Source:   incidentBundleSubtypeSource{},
	}
}

func (incidentBundleSubtypeSource) SupportedRecordTypes() []subtypepresence.RecordType {
	return []subtypepresence.RecordType{subtypepresence.RecordTypeIndicator}
}

func (incidentBundleSubtypeSource) ListSubtypeBindingsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]subtypepresence.Binding, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, 'indicator'::text
  FROM indicators
 WHERE incident_id = $1
 ORDER BY record_id
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
