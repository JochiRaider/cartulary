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
	specs := []incidentportability.FixedImportSpec{
		{LogicalBundlePath: "data/indicators.ndjson", AttributionTable: "indicators", StableIdentity: []string{"record_id"}, RequiredColumns: []string{"record_id", "incident_id"}, InsertSQL: `INSERT INTO indicators SELECT * FROM jsonb_populate_record(NULL::indicators, $1::jsonb)`},
		{LogicalBundlePath: "data/indicator_observations.ndjson", AttributionTable: "indicator_observations", StableIdentity: []string{"indicator_observation_id"}, RequiredColumns: []string{"indicator_observation_id"}, InsertSQL: `INSERT INTO indicator_observations SELECT * FROM jsonb_populate_record(NULL::indicator_observations, $1::jsonb)`},
		{LogicalBundlePath: "data/indicator_state_intervals.ndjson", AttributionTable: "indicator_state_intervals", StableIdentity: []string{"indicator_state_interval_id"}, RequiredColumns: []string{"indicator_state_interval_id"}, InsertSQL: `INSERT INTO indicator_state_intervals SELECT * FROM jsonb_populate_record(NULL::indicator_state_intervals, $1::jsonb)`},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
