package queryengine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

// recordSummaryLabelSQL renders the REQ-01-333a linked record summary label for
// the record identified by recordID and recordType. Labels come from the grid
// projections, which refresh in each source transaction. An empty label source
// yields "Untitled <noun>". A missing projection row can arise only from an
// invariant breach; it reads as the type's Untitled label, or "Untitled record"
// for an Artifact whose subtype is then unknown.
func recordSummaryLabelSQL(recordID, recordType sqlshape.Fragment) sqlshape.Fragment {
	return sqlshape.Format(`CASE %s
    WHEN 'timeline_event' THEN COALESCE((SELECT NULLIF(l.activity_synopsis_text, '') FROM timeline_grid_projection l WHERE l.record_id = %s), 'Untitled timeline record')
    WHEN 'host' THEN COALESCE((SELECT NULLIF(l.display_name, '') FROM host_grid_projection l WHERE l.record_id = %s), 'Untitled host')
    WHEN 'identity' THEN COALESCE((SELECT NULLIF(l.display_name, '') FROM identity_grid_projection l WHERE l.record_id = %s), 'Untitled identity')
    WHEN 'party' THEN COALESCE((SELECT NULLIF(l.display_name, '') FROM party_grid_projection l WHERE l.record_id = %s), 'Untitled party')
    WHEN 'indicator' THEN COALESCE((SELECT NULLIF(l.display_value, '') FROM indicator_grid_projection l WHERE l.record_id = %s), 'Untitled indicator')
    WHEN 'task_request' THEN COALESCE((SELECT NULLIF(l.title, '') FROM task_request_grid_projection l WHERE l.record_id = %s), 'Untitled task')
    WHEN 'decision' THEN COALESCE((SELECT NULLIF(l.summary, '') FROM decision_grid_projection l WHERE l.record_id = %s), 'Untitled decision')
    WHEN 'evidence' THEN COALESCE((SELECT NULLIF(l.title, '') FROM evidence_grid_projection l WHERE l.record_id = %s), 'Untitled evidence')
    WHEN 'assessment' THEN COALESCE((
        SELECT 'Assessment of ' || CASE l.subject_type
            WHEN 'host' THEN COALESCE((SELECT NULLIF(s.display_name, '') FROM host_grid_projection s WHERE s.record_id = l.subject_ref), 'Untitled host')
            WHEN 'identity' THEN COALESCE((SELECT NULLIF(s.display_name, '') FROM identity_grid_projection s WHERE s.record_id = l.subject_ref), 'Untitled identity')
        END
          FROM assessment_grid_projection l
         WHERE l.record_id = %s), 'Untitled assessment')
    WHEN 'artifact' THEN COALESCE((
        SELECT CASE l.artifact_type
            WHEN 'finding' THEN COALESCE(NULLIF(l.finding_statement, ''), 'Untitled finding')
            WHEN 'note' THEN COALESCE(NULLIF(l.title, ''), 'Untitled note')
            WHEN 'comm_log' THEN COALESCE(NULLIF(l.summary, ''), 'Untitled communication')
            WHEN 'lesson' THEN COALESCE(NULLIF(l.summary, ''), 'Untitled lesson')
            WHEN 'handoff' THEN COALESCE(NULLIF(l.current_state_summary, ''), 'Untitled handoff')
            WHEN 'status_review' THEN COALESCE(NULLIF(l.current_state_summary, ''), 'Untitled status review')
            WHEN 'forensic_keyword' THEN COALESCE(NULLIF(l.forensic_keyword_pattern, ''), 'Untitled forensic keyword')
            WHEN 'investigative_query' THEN COALESCE(NULLIF(l.investigative_query_purpose, ''), 'Untitled investigative query')
        END
          FROM artifact_grid_projection l
         WHERE l.record_id = %s), 'Untitled record')
END`, recordType, recordID, recordID, recordID, recordID, recordID, recordID, recordID, recordID, recordID, recordID)
}

// RecordSummaryLabelsTx returns the REQ-01-333a label of each requested live
// record in the incident. Records that are deleted, absent, or in another
// incident are omitted, so a caller cannot learn their state.
func RecordSummaryLabelsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	labels := make(map[uuid.UUID]string, len(recordIDs))
	if len(recordIDs) == 0 {
		return labels, nil
	}
	query := sqlshape.Format(`SELECT r.record_id, %s
  FROM records r
 WHERE r.incident_id = $1
   AND r.record_id = ANY($2::uuid[])
   AND r.deleted_at IS NULL`, recordSummaryLabelSQL(sqlshape.Literal("r.record_id"), sqlshape.Literal("r.record_type")))
	rows, err := tx.Query(ctx, sqlshape.Text(query), incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("query record summary labels: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var recordID uuid.UUID
		var label string
		if err := rows.Scan(&recordID, &label); err != nil {
			return nil, fmt.Errorf("scan record summary label: %w", err)
		}
		labels[recordID] = label
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record summary labels: %w", err)
	}
	return labels, nil
}
