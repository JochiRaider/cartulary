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
	specs := []incidentportability.FixedImportSpec{
		{
			"data/change_sets.ndjson", "change_sets", []string{"change_set_id"}, []string{"change_set_id", "incident_id", "actor_user_id"},
			`INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at)
SELECT (payload->>'change_set_id')::uuid, (payload->>'incident_id')::uuid,
       (payload->>'actor_user_id')::uuid, payload->>'source', payload->>'reason',
       payload->>'client_txn_id', payload->>'request_id',
       (payload->>'created_at')::timestamp with time zone
FROM (SELECT $1::jsonb AS payload) AS input`,
		},
		{
			"data/change_set_mutations.ndjson", "change_set_mutations", []string{"change_set_id", "sequence_no"}, []string{"change_set_id", "sequence_no"},
			`INSERT INTO change_set_mutations (change_set_id, sequence_no, target_kind, target_id, operation_kind, before_version_id, after_version_id, before_value, after_value)
SELECT (payload->>'change_set_id')::uuid, (payload->>'sequence_no')::integer,
       payload->>'target_kind', payload->>'target_id', payload->>'operation_kind',
       payload->>'before_version_id', payload->>'after_version_id',
       payload->'before_value', payload->'after_value'
FROM (SELECT $1::jsonb AS payload) AS input`,
		},
		{
			"data/record_revisions.ndjson", "record_revisions", []string{"revision_id"}, []string{"revision_id", "change_set_id", "record_id"},
			`INSERT INTO record_revisions (revision_id, change_set_id, record_id, row_version, before_json, after_json, created_at)
SELECT (payload->>'revision_id')::bigint, (payload->>'change_set_id')::uuid,
       (payload->>'record_id')::uuid, (payload->>'row_version')::bigint,
       payload->'before_json', payload->'after_json',
       (payload->>'created_at')::timestamp with time zone
FROM (SELECT $1::jsonb AS payload) AS input`,
		},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
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
