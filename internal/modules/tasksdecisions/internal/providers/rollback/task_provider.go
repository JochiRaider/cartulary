package rollback

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

type TaskRequestProvider struct{}

var _ rollbackcontract.RowSourceProvider = TaskRequestProvider{}

func NewTaskRequestProvider() TaskRequestProvider { return TaskRequestProvider{} }

func (TaskRequestProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := taskSourceForRollbackValue(value)
	if !ok || !validTaskSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

type taskLifecycle struct {
	incidentID    uuid.UUID
	createdAt     time.Time
	status        string
	blockedReason *string
	completedAt   *time.Time
	ownerUserID   *uuid.UUID
}

func (TaskRequestProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := taskSourceForRollbackValue(request.RetainedValue)
	if !ok || !validTaskSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	var lifecycle taskLifecycle
	if err := tx.QueryRow(ctx, `
SELECT incident_id, created_at, status, NULLIF(blocked_reason, ''), completed_at, owner_user_id
  FROM task_requests
 WHERE record_id = $1
`, request.RecordID).Scan(&lifecycle.incidentID, &lifecycle.createdAt, &lifecycle.status, &lifecycle.blockedReason, &lifecycle.completedAt, &lifecycle.ownerUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if raw, present := source["status"]; present {
		lifecycle.status = raw.(string)
	}
	if value, present, err := nullableText(source, "blocked_reason"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		lifecycle.blockedReason = value
	}
	if value, present, err := nullableTime(source, "completed_at"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		lifecycle.completedAt = value
	}
	if value, present, err := nullableUUID(source, "owner_user_id"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		lifecycle.ownerUserID = value
	}
	if !validTaskLifecycle(lifecycle) {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := validateTaskReferencesTx(ctx, tx, lifecycle.incidentID, source); err != nil {
		return err
	}
	values, err := typedValues(source, []fieldSpec{
		{"title", fieldText}, {"status", fieldText}, {"owner_user_id", fieldUUID}, {"priority", fieldText},
		{"task_kind", fieldText}, {"workstream", fieldText}, {"due_at", fieldTime},
		{"requester_party_text", fieldText}, {"requester_party_id", fieldUUID}, {"blocked_reason", fieldText},
		{"completed_at", fieldTime}, {"external_ticket_ref", fieldText}, {"closure_summary", fieldText},
		{"decision_record_id", fieldUUID},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE task_requests
   SET title = CASE WHEN $2 THEN $3::text ELSE title END,
       status = CASE WHEN $4 THEN $5::text ELSE status END,
       owner_user_id = CASE WHEN $6 THEN $7::uuid ELSE owner_user_id END,
       priority = CASE WHEN $8 THEN $9::text ELSE priority END,
       task_kind = CASE WHEN $10 THEN $11::text ELSE task_kind END,
       workstream = CASE WHEN $12 THEN $13::text ELSE workstream END,
       due_at = CASE WHEN $14 THEN $15::timestamptz ELSE due_at END,
       requester_party_text = CASE WHEN $16 THEN $17::text ELSE requester_party_text END,
       requester_party_id = CASE WHEN $18 THEN $19::uuid ELSE requester_party_id END,
       blocked_reason = CASE WHEN $20 THEN $21::text ELSE blocked_reason END,
       completed_at = CASE WHEN $22 THEN $23::timestamptz ELSE completed_at END,
       external_ticket_ref = CASE WHEN $24 THEN $25::text ELSE external_ticket_ref END,
       closure_summary = CASE WHEN $26 THEN $27::text ELSE closure_summary END,
       decision_record_id = CASE WHEN $28 THEN $29::uuid ELSE decision_record_id END,
       updated_at = $30
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func taskSourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	return sourceForRollbackValue(value)
}

func validTaskSource(source map[string]any) bool {
	if raw, present := source["title"]; present && !nonEmptyText(raw) {
		return false
	}
	if raw, present := source["status"]; present && !policyText(raw, policy.ValidTaskStatus) {
		return false
	}
	if raw, present := source["priority"]; present && raw != nil && !policyText(raw, policy.ValidTaskPriority) {
		return false
	}
	if raw, present := source["task_kind"]; present && !policyText(raw, policy.ValidTaskKind) {
		return false
	}
	return validTypedFields(source, []fieldSpec{{"owner_user_id", fieldUUID}, {"due_at", fieldTime}, {"requester_party_id", fieldUUID}, {"completed_at", fieldTime}, {"decision_record_id", fieldUUID}})
}

func validTaskLifecycle(state taskLifecycle) bool {
	if !policy.ValidTaskStatus(state.status) {
		return false
	}
	return policy.ValidateTaskState(policy.TaskLifecycleState{
		Status: state.status, BlockedReason: nullableSQLString(state.blockedReason),
		CompletedAt: nullableSQLTime(state.completedAt), OwnerUserID: nullableSQLUUID(state.ownerUserID),
		CreatedAt: state.createdAt,
	}) == nil
}

func validateTaskReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source map[string]any) error {
	for _, reference := range []struct {
		key        string
		recordType string
	}{
		{key: "requester_party_id", recordType: "party"},
		{key: "decision_record_id", recordType: "decision"},
	} {
		value, present, err := nullableUUID(source, reference.key)
		if err != nil {
			return rollbackcontract.ErrTargetNotReversible
		}
		if !present || value == nil {
			continue
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $2 AND record_type = $3)`, incidentID, *value, reference.recordType).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	return nil
}
