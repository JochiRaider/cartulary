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
	specs := []incidentportability.FixedImportSpec{
		{LogicalBundlePath: "data/record_links.ndjson", AttributionTable: "record_links", StableIdentity: []string{"record_link_id"}, RequiredColumns: []string{"record_link_id", "incident_id"}, InsertSQL: `INSERT INTO record_links SELECT * FROM jsonb_populate_record(NULL::record_links, $1::jsonb)`},
		{LogicalBundlePath: "data/record_tags.ndjson", AttributionTable: "record_tags", StableIdentity: []string{"record_tag_id"}, RequiredColumns: []string{"record_tag_id", "record_id", "incident_id"}, InsertSQL: `INSERT INTO record_tags SELECT * FROM jsonb_populate_record(NULL::record_tags, $1::jsonb)`},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
