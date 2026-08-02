package evidence

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
		{"data/evidence_records.ndjson", `SELECT to_jsonb(t) FROM evidence t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/evidence_custody_events.ndjson", `SELECT to_jsonb(t) FROM evidence_custody_events t WHERE incident_id = $1 ORDER BY evidence_record_id, occurred_at, custody_event_id`},
		{"data/object_blobs.ndjson", `SELECT to_jsonb(t) FROM object_blobs t WHERE incident_id = $1 ORDER BY object_blob_id`},
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
		{LogicalBundlePath: "data/object_blobs.ndjson", AttributionTable: "object_blobs", StableIdentity: []string{"object_blob_id"}, RequiredColumns: []string{"object_blob_id", "incident_id"}, InsertSQL: `INSERT INTO object_blobs SELECT * FROM jsonb_populate_record(NULL::object_blobs, $1::jsonb)`},
		{LogicalBundlePath: "data/evidence_records.ndjson", AttributionTable: "evidence", StableIdentity: []string{"record_id"}, RequiredColumns: []string{"record_id", "incident_id"}, InsertSQL: `INSERT INTO evidence SELECT * FROM jsonb_populate_record(NULL::evidence, $1::jsonb)`},
		{LogicalBundlePath: "data/evidence_custody_events.ndjson", AttributionTable: "evidence_custody_events", StableIdentity: []string{"custody_event_id"}, RequiredColumns: []string{"custody_event_id", "evidence_record_id"}, InsertSQL: `INSERT INTO evidence_custody_events SELECT * FROM jsonb_populate_record(NULL::evidence_custody_events, $1::jsonb)`},
	}
	for _, spec := range specs {
		recorder := attributions
		if spec.LogicalBundlePath == "data/object_blobs.ndjson" {
			// Blob staging already remapped and recorded the source owner before
			// replacing the storage reference. Applying the rewritten rows must
			// not reinterpret the target-local actor as a second source actor.
			recorder = nil
		}
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, recorder); err != nil {
			return err
		}
	}
	return nil
}
