package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
)

func (store *Store) InsertArtifactTx(
	ctx context.Context,
	tx pgx.Tx,
	input artifactprojection.ProjectionInput,
) error {
	row, err := artifactPhysicalRowFromInput(input)
	if err != nil {
		return err
	}
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
`, row.RecordID, row.IncidentID, row.RowVersion, row.ArtifactType,
		row.Title, row.Body, row.TimestampUTC, row.UpdatedAt.UTC(),
		row.CreatedAt.UTC(), row.CreatedByUserID, row.CommID, row.CommType,
		row.Audience, row.ChannelOrMeeting, row.Summary, row.NextReportAt,
		row.PrivilegeTag, row.HandoffID, row.OutgoingOwnerUserID,
		row.IncomingOwnerUserID, row.CurrentStateSummary, row.NextChecks,
		row.AcknowledgedAt, row.StatusReviewID, row.ReviewOwnerUserID,
		row.ActiveRisksSummary, row.LessonID, row.OwnerUserID, row.ClosureState,
		row.FindingStatement, row.FindingKind, row.FindingState,
		row.FindingOwnerUserID, row.FindingConfidenceScore, row.FindingClosedAt,
		row.FindingUpdatedAt, row.FindingConfidenceBand,
		row.InvestigativeQueryQueryID, row.InvestigativeQueryPlatform,
		row.InvestigativeQueryPurpose, row.InvestigativeQueryQueryText,
		row.InvestigativeQueryCreatedBy, row.InvestigativeQueryCreatedAt,
		row.InvestigativeQueryCreatedDay, row.ForensicKeywordKeywordID,
		row.ForensicKeywordPattern, row.ForensicKeywordReason,
		row.ForensicKeywordMatchMode, row.ForensicKeywordCaseSensitive,
		row.ForensicKeywordCreatedAt, row.ForensicKeywordCreatedDay,
		row.TimestampDay, row.NextReportDay, row.AckState,
		row.LinkedRecordCount); err != nil {
		return fmt.Errorf("insert Artifact projection row: %w", err)
	}
	return nil
}

// artifactPhysicalRow is private Projections-owned binding state. Its order
// mirrors the unchanged physical table and never crosses the owner boundary.
type artifactPhysicalRow struct {
	RecordID                     uuid.UUID
	IncidentID                   uuid.UUID
	RowVersion                   int64
	ArtifactType                 string
	Title                        *string
	Body                         *string
	TimestampUTC                 *time.Time
	UpdatedAt                    time.Time
	CreatedAt                    time.Time
	CreatedByUserID              *uuid.UUID
	CommID                       *string
	CommType                     *string
	Audience                     *string
	ChannelOrMeeting             *string
	Summary                      *string
	NextReportAt                 *time.Time
	PrivilegeTag                 *string
	HandoffID                    *string
	OutgoingOwnerUserID          *uuid.UUID
	IncomingOwnerUserID          *uuid.UUID
	CurrentStateSummary          *string
	NextChecks                   *string
	AcknowledgedAt               *time.Time
	StatusReviewID               *string
	ReviewOwnerUserID            *uuid.UUID
	ActiveRisksSummary           *string
	LessonID                     *string
	OwnerUserID                  *uuid.UUID
	ClosureState                 *string
	FindingStatement             *string
	FindingKind                  *string
	FindingState                 *string
	FindingOwnerUserID           *uuid.UUID
	FindingConfidenceScore       *int
	FindingClosedAt              *time.Time
	FindingUpdatedAt             *time.Time
	FindingConfidenceBand        *string
	InvestigativeQueryQueryID    *string
	InvestigativeQueryPlatform   *string
	InvestigativeQueryPurpose    *string
	InvestigativeQueryQueryText  *string
	InvestigativeQueryCreatedBy  *uuid.UUID
	InvestigativeQueryCreatedAt  *time.Time
	InvestigativeQueryCreatedDay *string
	ForensicKeywordKeywordID     *string
	ForensicKeywordPattern       *string
	ForensicKeywordReason        *string
	ForensicKeywordMatchMode     *string
	ForensicKeywordCaseSensitive *bool
	ForensicKeywordCreatedAt     *time.Time
	ForensicKeywordCreatedDay    *string
	TimestampDay                 *string
	NextReportDay                *string
	AckState                     string
	LinkedRecordCount            int
}

func artifactPhysicalRowFromInput(input artifactprojection.ProjectionInput) (artifactPhysicalRow, error) {
	envelope := input.Envelope()
	row := artifactPhysicalRow{
		RecordID: envelope.RecordID(), IncidentID: envelope.IncidentID(), RowVersion: envelope.RowVersion(),
		ArtifactType: input.ArtifactType(), Title: envelope.Title(), Body: envelope.Body(),
		TimestampUTC: envelope.TimestampUTC(), UpdatedAt: envelope.UpdatedAt(), CreatedAt: envelope.CreatedAt(),
		CreatedByUserID: envelope.CreatedByUserID(), TimestampDay: envelope.TimestampDay(),
		FindingUpdatedAt:      storageTimeValuePointer(envelope.UpdatedAt()),
		FindingConfidenceBand: storageStringPointer("unset"), AckState: "pending",
		LinkedRecordCount: envelope.LinkedRecordCount(),
	}
	if row.RecordID == uuid.Nil || row.IncidentID == uuid.Nil || row.RowVersion < 1 || row.ArtifactType == "" {
		return artifactPhysicalRow{}, fmt.Errorf("bind Artifact projection row: invalid input")
	}
	switch variant := input.Variant().(type) {
	case artifactprojection.NoteVariant:
	case artifactprojection.CommunicationLogVariant:
		row.CommID = storageStringPointer(variant.CommID)
		row.CommType = storageStringPointer(variant.CommType)
		row.Audience = storageStringPointer(variant.Audience)
		row.ChannelOrMeeting = storageStringPointer(variant.ChannelOrMeeting)
		row.Summary = storageStringPointer(variant.Summary)
		row.NextReportAt = storageTimePointer(variant.NextReportAt)
		row.PrivilegeTag = storageStringPointerFrom(variant.PrivilegeTag)
		row.NextReportDay = storageStringPointerFrom(variant.NextReportDay)
	case artifactprojection.HandoffVariant:
		row.HandoffID = storageStringPointer(variant.HandoffID)
		row.OutgoingOwnerUserID = storageUUIDPointer(variant.OutgoingOwnerUserID)
		row.IncomingOwnerUserID = storageUUIDPointer(variant.IncomingOwnerUserID)
		row.CurrentStateSummary = storageStringPointer(variant.CurrentStateSummary)
		row.NextChecks = storageStringPointerFrom(variant.NextChecks)
		row.AcknowledgedAt = storageTimePointer(variant.AcknowledgedAt)
		row.AckState = variant.AckState
	case artifactprojection.StatusReviewVariant:
		row.StatusReviewID = storageStringPointer(variant.StatusReviewID)
		row.ReviewOwnerUserID = storageUUIDPointer(variant.ReviewOwnerUserID)
		row.CurrentStateSummary = storageStringPointer(variant.CurrentStateSummary)
		row.ActiveRisksSummary = storageStringPointerFrom(variant.ActiveRisksSummary)
		row.NextReportAt = storageTimePointer(variant.NextReportAt)
		row.NextReportDay = storageStringPointerFrom(variant.NextReportDay)
	case artifactprojection.LessonVariant:
		row.LessonID = storageStringPointer(variant.LessonID)
		row.Summary = storageStringPointer(variant.Summary)
		row.OwnerUserID = storageUUIDPointer(variant.OwnerUserID)
		row.ClosureState = storageStringPointer(variant.ClosureState)
	case artifactprojection.FindingVariant:
		row.FindingStatement = storageStringPointer(variant.Statement)
		row.FindingKind = storageStringPointer(variant.Kind)
		row.FindingState = storageStringPointer(variant.State)
		row.FindingOwnerUserID = storageUUIDPointer(variant.OwnerUserID)
		row.FindingConfidenceScore = storageIntPointer(variant.ConfidenceScore)
		row.FindingClosedAt = storageTimePointer(variant.ClosedAt)
		row.FindingUpdatedAt = storageTimeValuePointer(variant.UpdatedAt)
		row.FindingConfidenceBand = storageStringPointer(variant.ConfidenceBand)
	case artifactprojection.InvestigativeQueryVariant:
		row.InvestigativeQueryQueryID = storageStringPointer(variant.QueryID)
		row.InvestigativeQueryPlatform = storageStringPointer(variant.Platform)
		row.InvestigativeQueryPurpose = storageStringPointer(variant.Purpose)
		row.InvestigativeQueryQueryText = storageStringPointer(variant.QueryText)
		row.InvestigativeQueryCreatedBy = storageUUIDPointer(variant.CreatedByUserID)
		row.InvestigativeQueryCreatedAt = storageTimeValuePointer(variant.CreatedAt)
		row.InvestigativeQueryCreatedDay = storageStringPointer(variant.CreatedDay)
	case artifactprojection.ForensicKeywordVariant:
		row.ForensicKeywordKeywordID = storageStringPointer(variant.KeywordID)
		row.ForensicKeywordPattern = storageStringPointer(variant.Pattern)
		row.ForensicKeywordReason = storageStringPointer(variant.Reason)
		row.ForensicKeywordMatchMode = storageStringPointer(variant.MatchMode)
		row.ForensicKeywordCaseSensitive = storageBoolPointer(variant.CaseSensitive)
		row.ForensicKeywordCreatedAt = storageTimeValuePointer(variant.CreatedAt)
		row.ForensicKeywordCreatedDay = storageStringPointer(variant.CreatedDay)
	default:
		return artifactPhysicalRow{}, fmt.Errorf("bind Artifact projection row: unknown variant")
	}
	return row, nil
}

func storageStringPointer(value string) *string { return &value }

func storageStringPointerFrom(value *string) *string {
	if value == nil {
		return nil
	}
	return storageStringPointer(*value)
}

func storageUUIDPointer(value uuid.UUID) *uuid.UUID { return &value }

func storageTimeValuePointer(value time.Time) *time.Time {
	result := value.UTC()
	return &result
}

func storageTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	return storageTimeValuePointer(*value)
}

func storageIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func storageBoolPointer(value bool) *bool { return &value }

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
