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
	specs := []incidentportability.FixedImportSpec{
		{"data/artifacts.ndjson", "artifacts", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO artifacts SELECT * FROM jsonb_populate_record(NULL::artifacts, $1::jsonb)`},
		{"data/artifact_findings.ndjson", "artifact_findings", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO artifact_findings SELECT * FROM jsonb_populate_record(NULL::artifact_findings, $1::jsonb)`},
		{"data/artifact_investigative_queries.ndjson", "artifact_investigative_queries", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO artifact_investigative_queries SELECT * FROM jsonb_populate_record(NULL::artifact_investigative_queries, $1::jsonb)`},
		{"data/artifact_forensic_keywords.ndjson", "artifact_forensic_keywords", []string{"record_id"}, []string{"record_id", "incident_id"}, `INSERT INTO artifact_forensic_keywords SELECT * FROM jsonb_populate_record(NULL::artifact_forensic_keywords, $1::jsonb)`},
		{"data/handoff_risk_refs.ndjson", "handoff_risk_refs", []string{"risk_ref_id"}, []string{"risk_ref_id", "handoff_record_id"}, `INSERT INTO handoff_risk_refs SELECT * FROM jsonb_populate_record(NULL::handoff_risk_refs, $1::jsonb)`},
	}
	for _, spec := range specs {
		if err := incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
