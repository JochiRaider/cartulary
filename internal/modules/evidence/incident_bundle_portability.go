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
	for _, spec := range []struct {
		target incidentportability.ImportTargetDescriptor
	}{
		{incidentportability.TargetObjectBlobs},
		{incidentportability.TargetEvidenceRecords},
		{incidentportability.TargetEvidenceCustodyEvents},
	} {
		if err := incidentportability.ImportBundleFileNDJSON(ctx, tx, spec.target, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
