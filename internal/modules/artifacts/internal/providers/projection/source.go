package projection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
)

type Source struct{}

func NewSource() *Source { return &Source{} }

func (*Source) LoadProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (artifactprojection.ProjectionInput, bool, error) {
	input, err := scanProjectionInput(tx.QueryRow(ctx, `
SELECT to_jsonb(source_row)
  FROM (`+artifactProjectionSourceSQL+`
 WHERE a.record_id = $1
) source_row
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return artifactprojection.ProjectionInput{}, false, nil
	}
	if err != nil {
		return artifactprojection.ProjectionInput{}, false, fmt.Errorf("load Artifact projection input: %w", err)
	}
	return input, true, nil
}

func (*Source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (artifactprojection.ProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return artifactprojection.ProjectionInputPage{}, fmt.Errorf("artifact projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, `
SELECT to_jsonb(source_row)
  FROM (`+artifactProjectionSourceSQL+`
 WHERE a.incident_id = $1
   AND ($2::uuid IS NULL OR a.record_id > $2)
 ORDER BY a.record_id
 LIMIT $3
) source_row
 ORDER BY source_row.record_id
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return artifactprojection.ProjectionInputPage{}, fmt.Errorf("list artifact projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]artifactprojection.ProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanProjectionInput(rows)
		if scanErr != nil {
			return artifactprojection.ProjectionInputPage{}, fmt.Errorf("scan Artifact projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return artifactprojection.ProjectionInputPage{}, fmt.Errorf("iterate Artifact projection inputs: %w", err)
	}

	page := artifactprojection.ProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanProjectionInput(scanner interface{ Scan(...any) error }) (artifactprojection.ProjectionInput, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		return artifactprojection.ProjectionInput{}, err
	}
	var input artifactprojection.ProjectionInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return artifactprojection.ProjectionInput{}, fmt.Errorf("decode typed Artifact projection input: %w", err)
	}
	if input.RecordID == uuid.Nil || input.IncidentID == uuid.Nil || input.ArtifactType == "" || input.AckState == "" {
		return artifactprojection.ProjectionInput{}, errors.New("typed Artifact projection input is incomplete")
	}
	return input, nil
}

const artifactProjectionSourceSQL = `
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.artifact_type,
    a.title,
    a.body,
    a.timestamp_utc,
    a.updated_at,
    a.created_at,
    a.created_by_user_id,
    a.comm_id,
    a.comm_type,
    a.audience,
    a.channel_or_meeting,
    a.summary,
    a.next_report_at,
    a.privilege_tag,
    a.handoff_id,
    a.outgoing_owner_user_id,
    a.incoming_owner_user_id,
    a.current_state_summary,
    a.next_checks,
    a.acknowledged_at,
    a.status_review_id,
    a.review_owner_user_id,
    a.active_risks_summary,
    a.lesson_id,
    a.owner_user_id,
    a.closure_state,
    f.statement AS finding_statement,
    f.kind AS finding_kind,
    f.state AS finding_state,
    f.owner_user_id AS finding_owner_user_id,
    f.confidence_score AS finding_confidence_score,
    f.closed_at AS finding_closed_at,
    GREATEST(a.updated_at, f.updated_at) AS finding_updated_at,
    cartulary_confidence_band(f.confidence_score) AS finding_confidence_band,
    iq.query_id AS investigative_query_query_id,
    iq.platform AS investigative_query_platform,
    iq.purpose AS investigative_query_purpose,
    iq.query_text AS investigative_query_query_text,
    iq.created_by_user_id AS investigative_query_created_by_user_id,
    iq.created_at AS investigative_query_created_at,
    iq.created_at::date AS investigative_query_created_day,
    fk.keyword_id AS forensic_keyword_keyword_id,
    fk.pattern AS forensic_keyword_pattern,
    fk.reason AS forensic_keyword_reason,
    fk.match_mode AS forensic_keyword_match_mode,
    fk.case_sensitive AS forensic_keyword_case_sensitive,
    fk.created_at AS forensic_keyword_created_at,
    fk.created_at::date AS forensic_keyword_created_day,
    a.timestamp_utc::date AS timestamp_day,
    a.next_report_at::date AS next_report_day,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END AS ack_state,
    COALESCE(links.linked_record_count, 0)::integer AS linked_record_count
  FROM artifacts a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN artifact_findings f
    ON f.incident_id = a.incident_id
   AND f.record_id = a.record_id
  LEFT JOIN artifact_investigative_queries iq
    ON iq.incident_id = a.incident_id
   AND iq.record_id = a.record_id
  LEFT JOIN artifact_forensic_keywords fk
    ON fk.incident_id = a.incident_id
   AND fk.record_id = a.record_id
  LEFT JOIN (
        SELECT incident_id, record_id, COUNT(*) AS linked_record_count
          FROM (
                SELECT incident_id, src_record_id AS record_id
                  FROM active_record_links_v1
                UNION ALL
                SELECT rl.incident_id, rl.dst_record_id AS record_id
                  FROM active_record_links_v1 rl
                  JOIN artifacts note_artifact
                    ON note_artifact.incident_id = rl.incident_id
                   AND note_artifact.record_id = rl.dst_record_id
                   AND note_artifact.artifact_type = 'note'
                 WHERE rl.link_type = 'references_artifact'
            ) counted_links
         GROUP BY incident_id, record_id
    ) links
    ON links.incident_id = a.incident_id
   AND links.record_id = a.record_id`
