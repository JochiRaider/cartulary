package indicators

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
		{"data/indicators.ndjson", `SELECT to_jsonb(t) FROM indicators t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/indicator_observations.ndjson", `SELECT to_jsonb(t) FROM indicator_observations t WHERE incident_id = $1 ORDER BY indicator_observation_id`},
		{"data/indicator_state_intervals.ndjson", `SELECT to_jsonb(t) FROM indicator_state_intervals t WHERE incident_id = $1 ORDER BY indicator_state_interval_id`},
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
	for _, spec := range []struct {
		target incidentportability.ImportTargetDescriptor
	}{
		{incidentportability.TargetIndicators},
		{incidentportability.TargetIndicatorObservations},
		{incidentportability.TargetIndicatorStateIntervals},
	} {
		if err := incidentportability.ImportBundleFileNDJSON(ctx, tx, spec.target, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
