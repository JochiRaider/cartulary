package incidentbundle

import (
	"context"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const (
	taskRequestsBundlePath = "data/task_requests.ndjson"
	decisionsBundlePath    = "data/decisions.ndjson"
)

func exportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	specs := []struct {
		path  string
		query string
	}{
		{taskRequestsBundlePath, `SELECT to_jsonb(t) FROM task_requests t WHERE incident_id = $1 ORDER BY record_id`},
		{decisionsBundlePath, `SELECT to_jsonb(t) FROM decisions t WHERE incident_id = $1 ORDER BY record_id`},
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
