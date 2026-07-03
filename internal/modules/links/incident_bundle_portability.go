package links

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
		{"data/record_links.ndjson", `SELECT to_jsonb(t) FROM record_links t WHERE incident_id = $1 ORDER BY record_link_id`},
		{"data/tags.ndjson", `SELECT jsonb_build_object('tag_name', tag_name, 'normalized_tag_name', normalized_tag_name) FROM (SELECT DISTINCT tag_name, normalized_tag_name FROM record_tags WHERE incident_id = $1 ORDER BY normalized_tag_name, tag_name) tags`},
		{"data/record_tags.ndjson", `SELECT to_jsonb(t) FROM record_tags t WHERE incident_id = $1 ORDER BY record_id, record_tag_id`},
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
	if err := incidentportability.ImportBundleFileNDJSON(ctx, tx, incidentportability.TargetRecordLinks, files, actorUserID, attributions); err != nil {
		return err
	}
	return incidentportability.ImportBundleFileNDJSON(ctx, tx, incidentportability.TargetRecordTags, files, actorUserID, attributions)
}
