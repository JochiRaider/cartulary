package entities

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
		{"data/entity_mentions.ndjson", `SELECT to_jsonb(t) FROM entity_mentions t JOIN records r ON r.record_id = t.source_record_id WHERE r.incident_id = $1 ORDER BY t.entity_mention_id`},
		{"data/hosts.ndjson", `SELECT to_jsonb(t) FROM hosts t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/identities.ndjson", `SELECT to_jsonb(t) FROM identities t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/entity_preserved_identifiers.ndjson", `SELECT to_jsonb(t) FROM entity_preserved_identifiers t WHERE incident_id = $1 ORDER BY entity_preserved_identifier_id`},
		{"data/entity_aliases.ndjson", `SELECT to_jsonb(t) FROM entity_aliases t WHERE incident_id = $1 ORDER BY entity_alias_id`},
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
		{"data/hosts.ndjson", "hosts"},
		{"data/identities.ndjson", "identities"},
		{"data/entity_preserved_identifiers.ndjson", "entity_preserved_identifiers"},
		{"data/entity_aliases.ndjson", "entity_aliases"},
		{"data/entity_mentions.ndjson", "entity_mentions"},
	} {
		if err := incidentportability.ImportNDJSON(ctx, tx, spec.table, files[spec.path], actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
