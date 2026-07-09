package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RefreshAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_score,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.subject_record_id,
    a.subject_type,
    a.assessment_state,
    a.confidence_score,
    cartulary_confidence_band(a.confidence_score),
    a.rationale,
    a.assessor_user_id,
    a.assessed_at,
    COALESCE(links.supporting_link_count, 0)::integer
  FROM assessments a
  JOIN records r
    ON r.record_id = a.record_id
  LEFT JOIN (
        SELECT
            rl.incident_id,
            rl.src_record_id,
            COUNT(*) AS supporting_link_count
          FROM active_record_links_v1 rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.deleted_at IS NULL
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.deleted_at IS NULL
         WHERE rl.link_type = 'supported_by'
           AND rl.src_record_id = $1
         GROUP BY rl.incident_id, rl.src_record_id
  ) links
    ON links.incident_id = a.incident_id
   AND links.src_record_id = a.record_id
 WHERE a.record_id = $1
   AND a.deleted_at IS NULL
   AND r.deleted_at IS NULL
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    subject_ref = EXCLUDED.subject_ref,
    subject_type = EXCLUDED.subject_type,
    assessment_state = EXCLUDED.assessment_state,
    confidence_score = EXCLUDED.confidence_score,
    confidence_band = EXCLUDED.confidence_band,
    rationale = EXCLUDED.rationale,
    assessor = EXCLUDED.assessor,
    assessed_at = EXCLUDED.assessed_at,
    supporting_link_count = EXCLUDED.supporting_link_count
`, recordID); err != nil {
		return fmt.Errorf("refresh assessment projection: %w", err)
	}
	return nil
}

func RebuildIncidentAssessmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM assessment_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear assessment projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_score,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.subject_record_id,
    a.subject_type,
    a.assessment_state,
    a.confidence_score,
    cartulary_confidence_band(a.confidence_score),
    a.rationale,
    a.assessor_user_id,
    a.assessed_at,
    COALESCE(links.supporting_link_count, 0)::integer
  FROM assessments a
  JOIN records r
    ON r.record_id = a.record_id
  LEFT JOIN (
        SELECT
            rl.incident_id,
            rl.src_record_id,
            COUNT(*) AS supporting_link_count
          FROM active_record_links_v1 rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.deleted_at IS NULL
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.deleted_at IS NULL
         WHERE rl.link_type = 'supported_by'
           AND rl.incident_id = $1
         GROUP BY rl.incident_id, rl.src_record_id
  ) links
    ON links.incident_id = a.incident_id
   AND links.src_record_id = a.record_id
 WHERE a.incident_id = $1
   AND a.deleted_at IS NULL
   AND r.deleted_at IS NULL
`, incidentID); err != nil {
		return fmt.Errorf("insert assessment projection rows: %w", err)
	}
	return nil
}
