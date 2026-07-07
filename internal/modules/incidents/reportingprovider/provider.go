package reportingprovider

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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

func ResolveSourceBoundaryStateTx(ctx context.Context, tx pgx.Tx, incident exportprovider.IncidentSnapshot) (exportprovider.SourceBoundaryState, error) {
	incidentID, err := uuid.Parse(incident.ID)
	if err != nil {
		return exportprovider.SourceBoundaryState{}, err
	}
	var latestID pgtype.Text
	var latestCreated pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
SELECT change_set_id::text, created_at
  FROM change_sets
 WHERE incident_id = $1
 ORDER BY created_at DESC, change_set_id DESC
 LIMIT 1
`, incidentID).Scan(&latestID, &latestCreated)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return exportprovider.SourceBoundaryState{}, err
	}

	var latestIDPtr *string
	var latestCreatedPtr *string
	if err == nil {
		id := latestID.String
		latestIDPtr = &id
		created := latestCreated.Time.UTC().Format(time.RFC3339Nano)
		latestCreatedPtr = &created
	}
	return exportprovider.SourceBoundaryState{
		IncidentID:               incident.ID,
		IncidentVersion:          incident.Version,
		LatestChangeSetID:        latestIDPtr,
		LatestChangeSetCreatedAt: latestCreatedPtr,
	}, nil
}
