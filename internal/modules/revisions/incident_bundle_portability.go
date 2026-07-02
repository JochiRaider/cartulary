package revisions

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
		{"data/change_sets.ndjson", `SELECT to_jsonb(t) FROM change_sets t WHERE incident_id = $1 ORDER BY change_set_id`},
		{"data/change_set_mutations.ndjson", `SELECT to_jsonb(t) FROM change_set_mutations t JOIN change_sets c ON c.change_set_id = t.change_set_id WHERE c.incident_id = $1 ORDER BY t.change_set_id, t.sequence_no`},
		{"data/record_revisions.ndjson", `SELECT to_jsonb(t) FROM record_revisions t JOIN change_sets c ON c.change_set_id = t.change_set_id WHERE c.incident_id = $1 ORDER BY t.revision_id`},
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
		path  string
		table string
	}{
		{"data/change_sets.ndjson", "change_sets"},
		{"data/change_set_mutations.ndjson", "change_set_mutations"},
		{"data/record_revisions.ndjson", "record_revisions"},
	} {
		if err := incidentportability.ImportNDJSON(ctx, tx, spec.table, files[spec.path], actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}

func RepairIncidentBundleImportedSequencesTx(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
SELECT setval(
    pg_get_serial_sequence('record_revisions', 'revision_id'),
    GREATEST(COALESCE((SELECT MAX(revision_id) FROM record_revisions), 0), 1),
    true
)
`)
	return err
}
