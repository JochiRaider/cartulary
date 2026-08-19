package reportingprovider

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func GetIncidentSnapshotTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (exportprovider.IncidentSnapshot, error) {
	row := tx.QueryRow(ctx, `
SELECT id, title, description, status, severity, tlp, current_phase, incident_version
  FROM incidents
 WHERE id = $1
`, incidentID)
	var id uuid.UUID
	var record exportprovider.IncidentSnapshot
	if err := row.Scan(
		&id,
		&record.Title,
		&record.Description,
		&record.Status,
		&record.Severity,
		&record.TLP,
		&record.CurrentPhase,
		&record.Version,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return exportprovider.IncidentSnapshot{}, exportprovider.ErrNotFound
		}
		return exportprovider.IncidentSnapshot{}, err
	}
	record.ID = id.String()
	return record, nil
}
