package tasksdecisions

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
		{"data/task_requests.ndjson", `SELECT to_jsonb(t) FROM task_requests t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/decisions.ndjson", `SELECT to_jsonb(t) FROM decisions t WHERE incident_id = $1 ORDER BY record_id`},
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
		{
			LogicalBundlePath: "data/task_requests.ndjson",
			AttributionTable:  "task_requests",
			StableIdentity:    []string{"record_id"},
			RequiredColumns:   []string{"record_id", "incident_id"},
			InsertSQL:         `INSERT INTO task_requests SELECT * FROM jsonb_populate_record(NULL::task_requests, $1::jsonb)`,
		},
		{
			LogicalBundlePath: "data/decisions.ndjson",
			AttributionTable:  "decisions",
			StableIdentity:    []string{"record_id"},
			RequiredColumns:   []string{"record_id", "incident_id"},
			InsertSQL:         `INSERT INTO decisions SELECT * FROM jsonb_populate_record(NULL::decisions, $1::jsonb)`,
		},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
