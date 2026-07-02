package incidents

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func ExportIncidentBundleIncident(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]byte, string, error) {
	var raw []byte
	var incidentKey string
	err := q.QueryRow(ctx, `SELECT to_jsonb(i), incident_key FROM incidents i WHERE id = $1`, incidentID).Scan(&raw, &incidentKey)
	if err != nil {
		return nil, "", err
	}
	canonical, err := incidentportability.CanonicalRawJSON(raw)
	return canonical, incidentKey, err
}

func ImportIncidentBundleIncidentTx(ctx context.Context, tx pgx.Tx, payload []byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	var row map[string]any
	if err := json.Unmarshal(payload, &row); err != nil {
		return &incidentportability.MalformedPayloadError{Err: err}
	}
	incidentportability.RemapTopLevelUserFields(row, "incidents", actorUserID, attributions)
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO incidents SELECT * FROM jsonb_populate_record(NULL::incidents, $1::jsonb)`, raw)
	return err
}
