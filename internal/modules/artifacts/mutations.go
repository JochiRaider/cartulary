package artifacts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type Store struct {
	revisionAppender *revisions.Appender
}

type FieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type CreateParams struct {
	ViewSchemaID string
	Values       map[string]FieldValue
}

func NewStore(appender *revisions.Appender) *Store {
	return &Store{revisionAppender: appender}
}

func (s *Store) InsertRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params CreateParams, now time.Time) error {
	artifactType := ArtifactTypeForView(params.ViewSchemaID)
	timestamp := now
	if value, ok := params.Values[artifactType+".timestamp_utc"]; ok && value.Timestamp != nil {
		timestamp = value.Timestamp.UTC()
	}
	commID, handoffID, statusReviewID, lessonID := any(nil), any(nil), any(nil), any(nil)
	switch params.ViewSchemaID {
	case CommLogViewSchemaID:
		commID = uuid.NewString()
	case HandoffViewSchemaID:
		handoffID = uuid.NewString()
	case StatusReviewViewSchemaID:
		statusReviewID = uuid.NewString()
	case LessonViewSchemaID:
		lessonID = uuid.NewString()
	}
	outgoingOwner := nullableUUIDValue(params.Values, "handoff.outgoing_owner_user_id")
	if params.ViewSchemaID == HandoffViewSchemaID && outgoingOwner == nil {
		outgoingOwner = actorID
	}
	currentStateSummary := nullableTextValue(params.Values, "handoff.current_state_summary")
	if params.ViewSchemaID == StatusReviewViewSchemaID {
		currentStateSummary = nullableTextValue(params.Values, "status_review.current_state_summary")
	}
	summary := nullableTextValue(params.Values, "comm_log.summary")
	if params.ViewSchemaID == LessonViewSchemaID {
		summary = nullableTextValue(params.Values, "lesson.summary")
	}
	nextReportAt := nullableTimestampValue(params.Values, "comm_log.next_report_at")
	if params.ViewSchemaID == StatusReviewViewSchemaID {
		nextReportAt = nullableTimestampValue(params.Values, "status_review.next_report_at")
	}
	reviewOwner := nullableUUIDValue(params.Values, "status_review.review_owner_user_id")
	if params.ViewSchemaID == StatusReviewViewSchemaID && reviewOwner == nil {
		reviewOwner = actorID
	}
	lessonOwner := nullableUUIDValue(params.Values, "lesson.owner_user_id")
	if params.ViewSchemaID == LessonViewSchemaID && lessonOwner == nil {
		lessonOwner = actorID
	}
	closureState := nullableTextValue(params.Values, "lesson.closure_state")
	if params.ViewSchemaID == LessonViewSchemaID && closureState == nil {
		closureState = "open"
	}
	_, err := tx.Exec(ctx, `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, timestamp_utc, updated_at, created_at,
    title, body,
    comm_id, comm_type, audience, channel_or_meeting, summary, next_report_at, privilege_tag,
    handoff_id, outgoing_owner_user_id, incoming_owner_user_id, current_state_summary, next_checks, acknowledged_at,
    status_review_id, review_owner_user_id, active_risks_summary,
    lesson_id, owner_user_id, closure_state, created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $5,
    $6, $7,
    $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20,
    $21, $22, $23,
    $24, $25, $26, $27
)
	`, recordID, incidentID, artifactType, timestamp, now,
		nullableTextValue(params.Values, "note.title"), nullableTextValue(params.Values, "note.body"),
		commID, nullableTextValue(params.Values, "comm_log.comm_type"), nullableTextValue(params.Values, "comm_log.audience"), nullableTextValue(params.Values, "comm_log.channel_or_meeting"), summary, nextReportAt, nullableTextValue(params.Values, "comm_log.privilege_tag"),
		handoffID, outgoingOwner, nullableUUIDValue(params.Values, "handoff.incoming_owner_user_id"), currentStateSummary, nullableTextValue(params.Values, "handoff.next_checks"), nullableTimestampValue(params.Values, "handoff.acknowledged_at"),
		statusReviewID, reviewOwner, nullableTextValue(params.Values, "status_review.active_risks_summary"),
		lessonID, lessonOwner, closureState, actorID)
	if err != nil {
		return fmt.Errorf("insert artifact: %w", err)
	}
	switch params.ViewSchemaID {
	case FindingsViewSchemaID:
		if err := s.insertFindingTx(ctx, tx, recordID, incidentID, actorID, params, now); err != nil {
			return err
		}
	case InvestigativeQueriesViewSchemaID:
		if err := s.insertInvestigativeQueryTx(ctx, tx, recordID, incidentID, actorID, params, now); err != nil {
			return err
		}
	case ForensicKeywordsViewSchemaID:
		if err := s.insertForensicKeywordTx(ctx, tx, recordID, incidentID, params, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) insertFindingTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params CreateParams, now time.Time) error {
	kind := nullableTextValue(params.Values, "finding.kind")
	if kind == nil {
		kind = "finding"
	}
	state := nullableTextValue(params.Values, "finding.state")
	if state == nil {
		state = "open"
	}
	owner := nullableUUIDValue(params.Values, "finding.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	var closedAt any
	if state == "closed" {
		closedAt = now
	}
	_, err := tx.Exec(ctx, `
INSERT INTO artifact_findings (
    record_id, incident_id, kind, statement, state, confidence_score,
    owner_user_id, closed_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $9
)
`, recordID, incidentID, kind, textValue(params.Values, "finding.statement"), state,
		nullableNumberValue(params.Values, "finding.confidence_score"), owner, closedAt, now)
	if err != nil {
		return fmt.Errorf("insert finding subtype: %w", err)
	}
	return nil
}

func (s *Store) insertInvestigativeQueryTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params CreateParams, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO artifact_investigative_queries (
    record_id, incident_id, query_id, platform, purpose, query_text,
    created_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $8
)
`, recordID, incidentID, uuid.NewString(),
		textValue(params.Values, "investigative_query.platform"),
		textValue(params.Values, "investigative_query.purpose"),
		textValue(params.Values, "investigative_query.query_text"),
		actorID, now)
	if err != nil {
		return fmt.Errorf("insert investigative query subtype: %w", err)
	}
	return nil
}

func (s *Store) insertForensicKeywordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, params CreateParams, now time.Time) error {
	matchMode := nullableTextValue(params.Values, "forensic_keyword.match_mode")
	if matchMode == nil {
		matchMode = "literal"
	}
	caseSensitive := nullableBoolValue(params.Values, "forensic_keyword.case_sensitive")
	if caseSensitive == nil {
		caseSensitive = false
	}
	_, err := tx.Exec(ctx, `
INSERT INTO artifact_forensic_keywords (
    record_id, incident_id, keyword_id, pattern, reason, match_mode,
    case_sensitive, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $8
)
`, recordID, incidentID, uuid.NewString(),
		textValue(params.Values, "forensic_keyword.pattern"),
		textValue(params.Values, "forensic_keyword.reason"),
		matchMode, caseSensitive, now)
	if err != nil {
		return fmt.Errorf("insert forensic keyword subtype: %w", err)
	}
	return nil
}

func (s *Store) ApplyDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string, value FieldValue, now time.Time) (bool, error) {
	table, column := tableColumnForField(fieldKey)
	if table == "" || column == "" {
		return false, fmt.Errorf("artifacts: unsupported field key %q", fieldKey)
	}
	dbValue := directDBValue(value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, table, column, column), recordID, dbValue, now)
	if err != nil {
		return false, fmt.Errorf("apply artifact direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) NormalizeFindingLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE artifact_findings
   SET closed_at = CASE
           WHEN state = 'closed' AND closed_at IS NULL THEN $2
           WHEN state = 'open' AND closed_at IS NOT NULL THEN NULL
           ELSE closed_at
       END,
       updated_at = CASE
           WHEN (state = 'closed' AND closed_at IS NULL)
             OR (state = 'open' AND closed_at IS NOT NULL)
           THEN $2
           ELSE updated_at
       END
 WHERE record_id = $1
   AND ((state = 'closed' AND closed_at IS NULL)
     OR (state = 'open' AND closed_at IS NOT NULL))
`, recordID, now)
	if err != nil {
		return false, fmt.Errorf("normalize finding lifecycle: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) TouchRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE artifacts SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch artifact row: %w", err)
	}
	return nil
}

func IsArtifactBackedField(fieldKey string) bool {
	table, column := tableColumnForField(fieldKey)
	return table != "" && column != ""
}

func tableColumnForField(fieldKey string) (string, string) {
	switch {
	case fieldKey == "note.title":
		return "artifacts", "title"
	case fieldKey == "note.body":
		return "artifacts", "body"
	case strings.HasPrefix(fieldKey, "comm_log."):
		return "artifacts", strings.TrimPrefix(fieldKey, "comm_log.")
	case strings.HasPrefix(fieldKey, "handoff."):
		return "artifacts", strings.TrimPrefix(fieldKey, "handoff.")
	case strings.HasPrefix(fieldKey, "status_review."):
		return "artifacts", strings.TrimPrefix(fieldKey, "status_review.")
	case strings.HasPrefix(fieldKey, "lesson."):
		return "artifacts", strings.TrimPrefix(fieldKey, "lesson.")
	case strings.HasPrefix(fieldKey, "finding."):
		return "artifact_findings", strings.TrimPrefix(fieldKey, "finding.")
	case strings.HasPrefix(fieldKey, "investigative_query."):
		return "artifact_investigative_queries", strings.TrimPrefix(fieldKey, "investigative_query.")
	case strings.HasPrefix(fieldKey, "forensic_keyword."):
		return "artifact_forensic_keywords", strings.TrimPrefix(fieldKey, "forensic_keyword.")
	default:
		return "", ""
	}
}

func directDBValue(value FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC()
	case value.UUID != nil:
		return *value.UUID
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func textValue(values map[string]FieldValue, field string) string {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return ""
}

func nullableTextValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return nil
}

func nullableUUIDValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return nil
}

func nullableTimestampValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Timestamp != nil {
		return value.Timestamp.UTC()
	}
	return nil
}

func nullableNumberValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Number != nil {
		return *value.Number
	}
	return nil
}

func nullableBoolValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Bool != nil {
		return *value.Bool
	}
	return nil
}
