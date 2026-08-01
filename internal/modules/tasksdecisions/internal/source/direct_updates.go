// Package source owns transaction-bound Task/Decision persistence mechanics.
package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

type Error struct {
	Operation string
	Err       error
}

func (e *Error) Error() string { return "tasksdecisions source: " + e.Operation }
func (e *Error) Unwrap() error { return e.Err }

func ApplyTaskDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	fieldKey string,
	value policy.FieldValue,
	now time.Time,
) (bool, error) {
	query, ok := taskUpdateStatement(fieldKey)
	if !ok {
		return false, &policy.ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	tag, err := tx.Exec(ctx, query, recordID, DirectDBValue(value), now.UTC())
	if err != nil {
		return false, &Error{Operation: "apply task direct change", Err: err}
	}
	return tag.RowsAffected() > 0, nil
}

func ApplyDecisionDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	fieldKey string,
	value policy.FieldValue,
	now time.Time,
) (bool, error) {
	query, ok := decisionUpdateStatement(fieldKey)
	if !ok {
		return false, &policy.ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	tag, err := tx.Exec(ctx, query, recordID, DirectDBValue(value), now.UTC())
	if err != nil {
		return false, &Error{Operation: "apply decision direct change", Err: err}
	}
	return tag.RowsAffected() > 0, nil
}

func taskUpdateStatement(fieldKey string) (string, bool) {
	switch fieldKey {
	case "task.title":
		return `UPDATE task_requests SET title = $2, updated_at = $3 WHERE record_id = $1 AND title IS DISTINCT FROM $2`, true
	case "task.status":
		return `UPDATE task_requests SET status = $2, updated_at = $3 WHERE record_id = $1 AND status IS DISTINCT FROM $2`, true
	case "task.owner_user_id":
		return `UPDATE task_requests SET owner_user_id = $2, updated_at = $3 WHERE record_id = $1 AND owner_user_id IS DISTINCT FROM $2`, true
	case "task.priority":
		return `UPDATE task_requests SET priority = $2, updated_at = $3 WHERE record_id = $1 AND priority IS DISTINCT FROM $2`, true
	case "task.task_kind":
		return `UPDATE task_requests SET task_kind = $2, updated_at = $3 WHERE record_id = $1 AND task_kind IS DISTINCT FROM $2`, true
	case "task.workstream":
		return `UPDATE task_requests SET workstream = $2, updated_at = $3 WHERE record_id = $1 AND workstream IS DISTINCT FROM $2`, true
	case "task.due_at":
		return `UPDATE task_requests SET due_at = $2, updated_at = $3 WHERE record_id = $1 AND due_at IS DISTINCT FROM $2`, true
	case "task.requester_party_text":
		return `UPDATE task_requests SET requester_party_text = $2, updated_at = $3 WHERE record_id = $1 AND requester_party_text IS DISTINCT FROM $2`, true
	case "task.requester_party_id":
		return `UPDATE task_requests SET requester_party_id = $2, updated_at = $3 WHERE record_id = $1 AND requester_party_id IS DISTINCT FROM $2`, true
	case "task.blocked_reason":
		return `UPDATE task_requests SET blocked_reason = $2, updated_at = $3 WHERE record_id = $1 AND blocked_reason IS DISTINCT FROM $2`, true
	case "task.completed_at":
		return `UPDATE task_requests SET completed_at = $2, updated_at = $3 WHERE record_id = $1 AND completed_at IS DISTINCT FROM $2`, true
	case "task.external_ticket_ref":
		return `UPDATE task_requests SET external_ticket_ref = $2, updated_at = $3 WHERE record_id = $1 AND external_ticket_ref IS DISTINCT FROM $2`, true
	case "task.closure_summary":
		return `UPDATE task_requests SET closure_summary = $2, updated_at = $3 WHERE record_id = $1 AND closure_summary IS DISTINCT FROM $2`, true
	case policy.TaskDecisionRecordField:
		return `UPDATE task_requests SET decision_record_id = $2, updated_at = $3 WHERE record_id = $1 AND decision_record_id IS DISTINCT FROM $2`, true
	default:
		return "", false
	}
}

func decisionUpdateStatement(fieldKey string) (string, bool) {
	switch fieldKey {
	case "decision.summary":
		return `UPDATE decisions SET summary = $2, updated_at = $3 WHERE record_id = $1 AND summary IS DISTINCT FROM $2`, true
	case "decision.status":
		return `UPDATE decisions SET status = $2, updated_at = $3 WHERE record_id = $1 AND status IS DISTINCT FROM $2`, true
	case "decision.owner_user_id":
		return `UPDATE decisions SET owner_user_id = $2, updated_at = $3 WHERE record_id = $1 AND owner_user_id IS DISTINCT FROM $2`, true
	case "decision.decision_type":
		return `UPDATE decisions SET decision_type = $2, updated_at = $3 WHERE record_id = $1 AND decision_type IS DISTINCT FROM $2`, true
	case "decision.decided_at":
		return `UPDATE decisions SET decided_at = $2, updated_at = $3 WHERE record_id = $1 AND decided_at IS DISTINCT FROM $2`, true
	case "decision.rationale":
		return `UPDATE decisions SET rationale = $2, updated_at = $3 WHERE record_id = $1 AND rationale IS DISTINCT FROM $2`, true
	default:
		return "", false
	}
}

func DirectDBValue(value policy.FieldValue) any {
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

func IsUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

func ClassifyUniqueConflict(err error, operation string) error {
	if IsUniqueViolation(err) {
		return &Error{Operation: operation, Err: err}
	}
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
