package incidents

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportIncidentBundleIncident(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]byte, string, error) {
	var raw []byte
	var incidentKey string
	err := q.QueryRow(ctx, `SELECT to_jsonb(i), incident_key FROM incidents i WHERE id = $1`, incidentID).Scan(&raw, &incidentKey)
	if err != nil {
		return nil, "", err
	}
	canonical, err := incidentportability.CanonicalRawJSON(raw)
	return canonical, incidentKey, err
}

func importIncidentBundleIncidentTx(ctx context.Context, tx pgx.Tx, payload []byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	var row map[string]any
	if err := json.Unmarshal(payload, &row); err != nil {
		return &incidentportability.MalformedPayloadError{Err: err}
	}
	if err := incidentportability.ValidateRequiredColumns(row, []string{"id", "incident_key"}, []string{"id"}); err != nil {
		return err
	}
	if err := incidentportability.RemapTopLevelUserFields(row, "incidents", []string{"id"}, actorUserID, attributions); err != nil {
		return err
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO incidents SELECT * FROM jsonb_populate_record(NULL::incidents, $1::jsonb)`, raw)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return incidentportability.FixedImportFailure("data/incident.json")
	}
	return nil
}
