package incidentbundle

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

type incidentBundleSubtypeSource struct{}

func SubtypeContribution() subtypepresence.Contribution {
	return subtypepresence.Contribution{
		FamilyID: "indicators",
		Source:   incidentBundleSubtypeSource{},
	}
}

func (incidentBundleSubtypeSource) SupportedRecordTypes() []subtypepresence.RecordType {
	return []subtypepresence.RecordType{subtypepresence.RecordTypeIndicator}
}

func (incidentBundleSubtypeSource) VisitSubtypeBindingsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	visit func(subtypepresence.Binding) error,
) error {
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, 'indicator'::text
  FROM indicators
 WHERE incident_id = $1
 ORDER BY record_id
`, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var binding subtypepresence.Binding
		if err := rows.Scan(&binding.RecordID, &binding.IncidentID, &binding.RecordType); err != nil {
			return err
		}
		if err := visit(binding); err != nil {
			return err
		}
	}
	return rows.Err()
}
