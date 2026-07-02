package savedviews

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, "data/saved_views.ndjson", `SELECT to_jsonb(t) FROM saved_views t WHERE incident_id = $1 ORDER BY saved_view_id`)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{file}, nil
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	return incidentportability.ImportNDJSON(ctx, tx, "saved_views", files["data/saved_views.ndjson"], actorUserID, attributions)
}
