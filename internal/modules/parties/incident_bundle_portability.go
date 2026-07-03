package parties

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, "data/parties.ndjson", `SELECT to_jsonb(t) FROM parties t WHERE incident_id = $1 ORDER BY record_id`)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{file}, nil
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	return incidentportability.ImportBundleFileNDJSON(ctx, tx, incidentportability.TargetParties, files, actorUserID, attributions)
}
