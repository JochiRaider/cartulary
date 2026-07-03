package timeline

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	specs := []struct {
		path  string
		query string
	}{
		{"data/timeline_time_conversion_profiles.ndjson", `SELECT to_jsonb(t) FROM timeline_time_conversion_profiles t WHERE incident_id = $1 ORDER BY incident_id`},
		{"data/timeline_events.ndjson", `SELECT to_jsonb(t) FROM timeline_events t WHERE incident_id = $1 ORDER BY record_id`},
	}
	files := make([]incidentportability.File, 0, len(specs))
	for _, spec := range specs {
		file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, spec.path, spec.query)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	if err := incidentportability.ImportBundleFileNDJSON(ctx, tx, incidentportability.TargetTimelineTimeConversionProfiles, files, actorUserID, attributions); err != nil {
		return err
	}
	return incidentportability.ImportBundleFileNDJSON(ctx, tx, incidentportability.TargetTimelineEvents, files, actorUserID, attributions)
}
