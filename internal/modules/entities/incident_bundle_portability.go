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
	specs := []incidentportability.FixedImportSpec{
		{"data/hosts.ndjson", "hosts", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO hosts SELECT * FROM jsonb_populate_record(NULL::hosts, $1::jsonb)`},
		{"data/identities.ndjson", "identities", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO identities SELECT * FROM jsonb_populate_record(NULL::identities, $1::jsonb)`},
		{"data/entity_preserved_identifiers.ndjson", "entity_preserved_identifiers", []string{"entity_preserved_identifier_id"}, []string{"entity_preserved_identifier_id", "record_id"}, `INSERT INTO entity_preserved_identifiers SELECT * FROM jsonb_populate_record(NULL::entity_preserved_identifiers, $1::jsonb)`},
		{"data/entity_aliases.ndjson", "entity_aliases", []string{"entity_alias_id"}, []string{"entity_alias_id", "record_id"}, `INSERT INTO entity_aliases SELECT * FROM jsonb_populate_record(NULL::entity_aliases, $1::jsonb)`},
		{"data/entity_mentions.ndjson", "entity_mentions", []string{"entity_mention_id"}, []string{"entity_mention_id", "source_record_id"}, `INSERT INTO entity_mentions SELECT * FROM jsonb_populate_record(NULL::entity_mentions, $1::jsonb)`},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
