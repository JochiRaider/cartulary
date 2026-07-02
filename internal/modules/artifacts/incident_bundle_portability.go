package artifacts

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
		{"data/artifacts.ndjson", `SELECT to_jsonb(t) FROM artifacts t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/artifact_findings.ndjson", `SELECT to_jsonb(t) FROM artifact_findings t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/artifact_investigative_queries.ndjson", `SELECT to_jsonb(t) FROM artifact_investigative_queries t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/artifact_forensic_keywords.ndjson", `SELECT to_jsonb(t) FROM artifact_forensic_keywords t WHERE incident_id = $1 ORDER BY record_id`},
		{"data/handoff_risk_refs.ndjson", `SELECT to_jsonb(t) FROM handoff_risk_refs t WHERE incident_id = $1 ORDER BY risk_ref_id`},
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
		{"data/artifacts.ndjson", "artifacts"},
		{"data/artifact_findings.ndjson", "artifact_findings"},
		{"data/artifact_investigative_queries.ndjson", "artifact_investigative_queries"},
		{"data/artifact_forensic_keywords.ndjson", "artifact_forensic_keywords"},
		{"data/handoff_risk_refs.ndjson", "handoff_risk_refs"},
	} {
		if err := incidentportability.ImportNDJSON(ctx, tx, spec.table, files[spec.path], actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
