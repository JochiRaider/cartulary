package source

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

func InsertTaskRequestTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	params policy.TaskCreateParams,
	now time.Time,
) error {
	status := nullableTextValue(params.Values, "task.status")
	if status == nil {
		status = policy.DefaultTaskStatus
	}
	owner := nullableUUIDValue(params.Values, "task.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	priority := nullableTextValue(params.Values, "task.priority")
	if priority == nil {
		priority = policy.DefaultTaskPriority
	}
	completedAt := nullableTimestampValue(params.Values, "task.completed_at")
	if status == "done" && completedAt == nil {
		completedAt = now
	}
	if err := policy.ValidateTaskState(policy.TaskLifecycleState{
		Status:        status.(string),
		BlockedReason: nullableStringFromAny(nullableTextValue(params.Values, "task.blocked_reason")),
		CompletedAt:   nullableTimeFromAny(completedAt),
		OwnerUserID:   nullableUUIDStringFromAny(owner),
		CreatedAt:     now,
	}); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority, task_kind,
    workstream, due_at, requester_party_text, requester_party_id, blocked_reason,
    completed_at, external_ticket_ref, closure_summary, decision_record_id,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15, $16,
    $17, $17
)
`, recordID, incidentID,
		textValue(params.Values, "task.title"),
		status,
		owner,
		priority,
		textValue(params.Values, "task.task_kind"),
		nullableTextValue(params.Values, "task.workstream"),
		nullableTimestampValue(params.Values, "task.due_at"),
		nullableTextValue(params.Values, "task.requester_party_text"),
		nullableUUIDValue(params.Values, "task.requester_party_id"),
		nullableTextValue(params.Values, "task.blocked_reason"),
		completedAt,
		nullableTextValue(params.Values, "task.external_ticket_ref"),
		nullableTextValue(params.Values, "task.closure_summary"),
		nullableUUIDValue(params.Values, "task.decision_record_id"),
		now)
	if err != nil {
		return fmt.Errorf("insert task request: %w", err)
	}
	return nil
}

func InsertDecisionTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	params policy.DecisionCreateParams,
	now time.Time,
) error {
	status := nullableTextValue(params.Values, "decision.status")
	if status == nil {
		status = policy.DefaultDecisionStatus
	}
	owner := nullableUUIDValue(params.Values, "decision.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	decidedAt := nullableTimestampValue(params.Values, "decision.decided_at")
	if decidedAt == nil {
		decidedAt = now
	}
	_, err := tx.Exec(ctx, `
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id, decision_type,
    decided_at, rationale, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $9
)
`, recordID, incidentID,
		textValue(params.Values, "decision.summary"),
		status,
		owner,
		textValue(params.Values, "decision.decision_type"),
		decidedAt,
		textValue(params.Values, "decision.rationale"),
		now)
	if err != nil {
		return fmt.Errorf("insert decision: %w", err)
	}
	return nil
}

func textValue(values map[string]policy.FieldValue, field string) string {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return ""
}

func nullableTextValue(values map[string]policy.FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return nil
}

func nullableUUIDValue(values map[string]policy.FieldValue, field string) any {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return nil
}

func nullableTimestampValue(values map[string]policy.FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Timestamp != nil {
		return value.Timestamp.UTC()
	}
	return nil
}

func nullableStringFromAny(value any) sql.NullString {
	if text, ok := value.(string); ok && text != "" {
		return sql.NullString{String: text, Valid: true}
	}
	return sql.NullString{}
}

func nullableTimeFromAny(value any) sql.NullTime {
	if timestamp, ok := value.(time.Time); ok {
		return sql.NullTime{Time: timestamp.UTC(), Valid: true}
	}
	return sql.NullTime{}
}

func nullableUUIDStringFromAny(value any) sql.NullString {
	if id, ok := value.(uuid.UUID); ok {
		return sql.NullString{String: id.String(), Valid: true}
	}
	return sql.NullString{}
}
