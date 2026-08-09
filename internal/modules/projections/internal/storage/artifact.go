package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
)

func (store *Store) InsertArtifactTx(
	ctx context.Context,
	tx pgx.Tx,
	input artifactprojection.ProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO artifact_grid_projection (
    record_id, incident_id, row_version, artifact_type, title, body,
    timestamp_utc, updated_at, created_at, created_by_user_id, comm_id,
    comm_type, audience, channel_or_meeting, summary, next_report_at,
    privilege_tag, handoff_id, outgoing_owner_user_id,
    incoming_owner_user_id, current_state_summary, next_checks,
    acknowledged_at, status_review_id, review_owner_user_id,
    active_risks_summary, lesson_id, owner_user_id, closure_state,
    finding_statement, finding_kind, finding_state, finding_owner_user_id,
    finding_confidence_score, finding_closed_at, finding_updated_at,
    finding_confidence_band, investigative_query_query_id,
    investigative_query_platform, investigative_query_purpose,
    investigative_query_query_text, investigative_query_created_by_user_id,
    investigative_query_created_at, investigative_query_created_day,
    forensic_keyword_keyword_id, forensic_keyword_pattern,
    forensic_keyword_reason, forensic_keyword_match_mode,
    forensic_keyword_case_sensitive, forensic_keyword_created_at,
    forensic_keyword_created_day, timestamp_day, next_report_day,
    ack_state, linked_record_count
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
    $31, $32, $33, $34, $35, $36, $37, $38, $39, $40,
    $41, $42, $43, $44, $45, $46, $47, $48, $49, $50,
    $51, $52, $53, $54, $55
)
`, input.RecordID, input.IncidentID, input.RowVersion, input.ArtifactType,
		input.Title, input.Body, input.TimestampUTC, input.UpdatedAt.UTC(),
		input.CreatedAt.UTC(), input.CreatedByUserID, input.CommID, input.CommType,
		input.Audience, input.ChannelOrMeeting, input.Summary, input.NextReportAt,
		input.PrivilegeTag, input.HandoffID, input.OutgoingOwnerUserID,
		input.IncomingOwnerUserID, input.CurrentStateSummary, input.NextChecks,
		input.AcknowledgedAt, input.StatusReviewID, input.ReviewOwnerUserID,
		input.ActiveRisksSummary, input.LessonID, input.OwnerUserID, input.ClosureState,
		input.FindingStatement, input.FindingKind, input.FindingState,
		input.FindingOwnerUserID, input.FindingConfidenceScore, input.FindingClosedAt,
		input.FindingUpdatedAt, input.FindingConfidenceBand,
		input.InvestigativeQueryQueryID, input.InvestigativeQueryPlatform,
		input.InvestigativeQueryPurpose, input.InvestigativeQueryQueryText,
		input.InvestigativeQueryCreatedBy, input.InvestigativeQueryCreatedAt,
		input.InvestigativeQueryCreatedDay, input.ForensicKeywordKeywordID,
		input.ForensicKeywordPattern, input.ForensicKeywordReason,
		input.ForensicKeywordMatchMode, input.ForensicKeywordCaseSensitive,
		input.ForensicKeywordCreatedAt, input.ForensicKeywordCreatedDay,
		input.TimestampDay, input.NextReportDay, input.AckState,
		input.LinkedRecordCount); err != nil {
		return fmt.Errorf("insert Artifact projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteArtifactRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM artifact_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Artifact projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteArtifactIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM artifact_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Artifact projection rows: %w", err)
	}
	return nil
}
