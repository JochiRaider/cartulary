package projection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

	pageInputs := inputs
	var nextRecordID *uuid.UUID
	if len(inputs) > limit {
		pageInputs = inputs[:limit]
		next := pageInputs[len(pageInputs)-1].Envelope().RecordID()
		nextRecordID = &next
	}
	page, err := artifactprojection.NewProjectionInputPage(pageInputs, nextRecordID)
	if err != nil {
		return artifactprojection.ProjectionInputPage{}, fmt.Errorf("construct Artifact projection input page: %w", err)
	}
	return page, nil
}

func scanProjectionInput(scanner interface{ Scan(...any) error }) (artifactprojection.ProjectionInput, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		return artifactprojection.ProjectionInput{}, err
	}
	var row projectionScanRow
	if err := json.Unmarshal(raw, &row); err != nil {
		return artifactprojection.ProjectionInput{}, fmt.Errorf("decode typed Artifact projection input: %w", err)
	}
	input, err := projectionInputFromScanRow(row)
	if err != nil {
		return artifactprojection.ProjectionInput{}, fmt.Errorf("construct typed Artifact projection input: %w", err)
	}
	return input, nil
}

// projectionScanRow is private SQL-decoding state. It mirrors the current
// source query only inside Artifacts and never crosses the owner boundary.
type projectionScanRow struct {
	RecordID                     uuid.UUID  `json:"record_id"`
	IncidentID                   uuid.UUID  `json:"incident_id"`
	RowVersion                   int64      `json:"row_version"`
	ArtifactType                 string     `json:"artifact_type"`
	Title                        *string    `json:"title"`
	Body                         *string    `json:"body"`
	TimestampUTC                 *time.Time `json:"timestamp_utc"`
	UpdatedAt                    time.Time  `json:"updated_at"`
	CreatedAt                    time.Time  `json:"created_at"`
	CreatedByUserID              *uuid.UUID `json:"created_by_user_id"`
	CommID                       *string    `json:"comm_id"`
	CommType                     *string    `json:"comm_type"`
	Audience                     *string    `json:"audience"`
	ChannelOrMeeting             *string    `json:"channel_or_meeting"`
	Summary                      *string    `json:"summary"`
	NextReportAt                 *time.Time `json:"next_report_at"`
	PrivilegeTag                 *string    `json:"privilege_tag"`
	HandoffID                    *string    `json:"handoff_id"`
	OutgoingOwnerUserID          *uuid.UUID `json:"outgoing_owner_user_id"`
	IncomingOwnerUserID          *uuid.UUID `json:"incoming_owner_user_id"`
	CurrentStateSummary          *string    `json:"current_state_summary"`
	NextChecks                   *string    `json:"next_checks"`
	AcknowledgedAt               *time.Time `json:"acknowledged_at"`
	StatusReviewID               *string    `json:"status_review_id"`
	ReviewOwnerUserID            *uuid.UUID `json:"review_owner_user_id"`
	ActiveRisksSummary           *string    `json:"active_risks_summary"`
	LessonID                     *string    `json:"lesson_id"`
	OwnerUserID                  *uuid.UUID `json:"owner_user_id"`
	ClosureState                 *string    `json:"closure_state"`
	FindingStatement             *string    `json:"finding_statement"`
	FindingKind                  *string    `json:"finding_kind"`
	FindingState                 *string    `json:"finding_state"`
	FindingOwnerUserID           *uuid.UUID `json:"finding_owner_user_id"`
	FindingConfidenceScore       *int       `json:"finding_confidence_score"`
	FindingClosedAt              *time.Time `json:"finding_closed_at"`
	FindingUpdatedAt             *time.Time `json:"finding_updated_at"`
	FindingConfidenceBand        *string    `json:"finding_confidence_band"`
	InvestigativeQueryQueryID    *string    `json:"investigative_query_query_id"`
	InvestigativeQueryPlatform   *string    `json:"investigative_query_platform"`
	InvestigativeQueryPurpose    *string    `json:"investigative_query_purpose"`
	InvestigativeQueryQueryText  *string    `json:"investigative_query_query_text"`
	InvestigativeQueryCreatedBy  *uuid.UUID `json:"investigative_query_created_by_user_id"`
	InvestigativeQueryCreatedAt  *time.Time `json:"investigative_query_created_at"`
	InvestigativeQueryCreatedDay *string    `json:"investigative_query_created_day"`
	ForensicKeywordKeywordID     *string    `json:"forensic_keyword_keyword_id"`
	ForensicKeywordPattern       *string    `json:"forensic_keyword_pattern"`
	ForensicKeywordReason        *string    `json:"forensic_keyword_reason"`
	ForensicKeywordMatchMode     *string    `json:"forensic_keyword_match_mode"`
	ForensicKeywordCaseSensitive *bool      `json:"forensic_keyword_case_sensitive"`
	ForensicKeywordCreatedAt     *time.Time `json:"forensic_keyword_created_at"`
	ForensicKeywordCreatedDay    *string    `json:"forensic_keyword_created_day"`
	TimestampDay                 *string    `json:"timestamp_day"`
	NextReportDay                *string    `json:"next_report_day"`
	AckState                     string     `json:"ack_state"`
	LinkedRecordCount            int        `json:"linked_record_count"`
}

func projectionInputFromScanRow(row projectionScanRow) (artifactprojection.ProjectionInput, error) {
	envelope, err := artifactprojection.NewProjectionEnvelope(
		row.RecordID, row.IncidentID, row.RowVersion, row.Title, row.Body, row.TimestampUTC,
		row.UpdatedAt, row.CreatedAt, row.CreatedByUserID, row.TimestampDay, row.LinkedRecordCount,
	)
	if err != nil {
		return artifactprojection.ProjectionInput{}, err
	}
	if err := validateJoinedSubtype(row); err != nil {
		return artifactprojection.ProjectionInput{}, err
	}
	switch row.ArtifactType {
	case "note":
		return artifactprojection.NewNoteProjectionInput(envelope, artifactprojection.NoteVariant{})
	case "comm_log":
		return artifactprojection.NewCommunicationLogProjectionInput(envelope, artifactprojection.CommunicationLogVariant{
			CommID: requiredString(row.CommID), CommType: requiredString(row.CommType),
			Audience: requiredString(row.Audience), ChannelOrMeeting: requiredString(row.ChannelOrMeeting),
			Summary: requiredString(row.Summary), NextReportAt: row.NextReportAt,
			PrivilegeTag: row.PrivilegeTag, NextReportDay: row.NextReportDay,
		})
	case "handoff":
		return artifactprojection.NewHandoffProjectionInput(envelope, artifactprojection.HandoffVariant{
			HandoffID: requiredString(row.HandoffID), OutgoingOwnerUserID: requiredUUID(row.OutgoingOwnerUserID),
			IncomingOwnerUserID: requiredUUID(row.IncomingOwnerUserID), CurrentStateSummary: requiredString(row.CurrentStateSummary),
			NextChecks: row.NextChecks, AcknowledgedAt: row.AcknowledgedAt, AckState: row.AckState,
		})
	case "status_review":
		return artifactprojection.NewStatusReviewProjectionInput(envelope, artifactprojection.StatusReviewVariant{
			StatusReviewID: requiredString(row.StatusReviewID), ReviewOwnerUserID: requiredUUID(row.ReviewOwnerUserID),
			CurrentStateSummary: requiredString(row.CurrentStateSummary), ActiveRisksSummary: row.ActiveRisksSummary,
			NextReportAt: row.NextReportAt, NextReportDay: row.NextReportDay,
		})
	case "lesson":
		return artifactprojection.NewLessonProjectionInput(envelope, artifactprojection.LessonVariant{
			LessonID: requiredString(row.LessonID), Summary: requiredString(row.Summary),
			OwnerUserID: requiredUUID(row.OwnerUserID), ClosureState: requiredString(row.ClosureState),
		})
	case "finding":
		return artifactprojection.NewFindingProjectionInput(envelope, artifactprojection.FindingVariant{
			Statement: requiredString(row.FindingStatement), Kind: requiredString(row.FindingKind),
			State: requiredString(row.FindingState), OwnerUserID: requiredUUID(row.FindingOwnerUserID),
			ConfidenceScore: row.FindingConfidenceScore, ClosedAt: row.FindingClosedAt,
			UpdatedAt: requiredTime(row.FindingUpdatedAt), ConfidenceBand: requiredString(row.FindingConfidenceBand),
		})
	case "investigative_query":
		return artifactprojection.NewInvestigativeQueryProjectionInput(envelope, artifactprojection.InvestigativeQueryVariant{
			QueryID: requiredString(row.InvestigativeQueryQueryID), Platform: requiredString(row.InvestigativeQueryPlatform),
			Purpose: requiredString(row.InvestigativeQueryPurpose), QueryText: requiredString(row.InvestigativeQueryQueryText),
			CreatedByUserID: requiredUUID(row.InvestigativeQueryCreatedBy), CreatedAt: requiredTime(row.InvestigativeQueryCreatedAt),
			CreatedDay: requiredString(row.InvestigativeQueryCreatedDay),
		})
	case "forensic_keyword":
		if row.ForensicKeywordCaseSensitive == nil {
			return artifactprojection.ProjectionInput{}, errors.New("forensic-keyword case-sensitivity fact is missing")
		}
		return artifactprojection.NewForensicKeywordProjectionInput(envelope, artifactprojection.ForensicKeywordVariant{
			KeywordID: requiredString(row.ForensicKeywordKeywordID), Pattern: requiredString(row.ForensicKeywordPattern),
			Reason: requiredString(row.ForensicKeywordReason), MatchMode: requiredString(row.ForensicKeywordMatchMode),
			CaseSensitive: *row.ForensicKeywordCaseSensitive, CreatedAt: requiredTime(row.ForensicKeywordCreatedAt),
			CreatedDay: requiredString(row.ForensicKeywordCreatedDay),
		})
	default:
		return artifactprojection.ProjectionInput{}, fmt.Errorf("unknown Artifact projection type %q", row.ArtifactType)
	}
}

func validateJoinedSubtype(row projectionScanRow) error {
	finding := row.FindingStatement != nil || row.FindingKind != nil || row.FindingState != nil ||
		row.FindingOwnerUserID != nil || row.FindingConfidenceScore != nil || row.FindingClosedAt != nil ||
		row.FindingUpdatedAt != nil || row.FindingConfidenceBand != nil
	query := row.InvestigativeQueryQueryID != nil || row.InvestigativeQueryPlatform != nil ||
		row.InvestigativeQueryPurpose != nil || row.InvestigativeQueryQueryText != nil ||
		row.InvestigativeQueryCreatedBy != nil || row.InvestigativeQueryCreatedAt != nil || row.InvestigativeQueryCreatedDay != nil
	keyword := row.ForensicKeywordKeywordID != nil || row.ForensicKeywordPattern != nil ||
		row.ForensicKeywordReason != nil || row.ForensicKeywordMatchMode != nil ||
		row.ForensicKeywordCaseSensitive != nil || row.ForensicKeywordCreatedAt != nil || row.ForensicKeywordCreatedDay != nil
	wantFinding := row.ArtifactType == "finding"
	wantQuery := row.ArtifactType == "investigative_query"
	wantKeyword := row.ArtifactType == "forensic_keyword"
	if finding != wantFinding || query != wantQuery || keyword != wantKeyword {
		return errors.New("artifact projection subtype facts mismatch artifact type")
	}
	if row.ArtifactType != "handoff" && row.AckState != "pending" {
		return errors.New("artifact projection acknowledgement state mismatches artifact type")
	}
	return nil
}

func requiredString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func requiredUUID(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func requiredTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.UTC()
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
    CASE
        WHEN f.record_id IS NULL THEN NULL
        ELSE GREATEST(a.updated_at, f.updated_at)
    END AS finding_updated_at,
    CASE
        WHEN f.record_id IS NULL THEN NULL
        ELSE cartulary_confidence_band(f.confidence_score)
    END AS finding_confidence_band,
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
