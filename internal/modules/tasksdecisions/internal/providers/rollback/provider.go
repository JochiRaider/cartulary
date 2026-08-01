package rollback

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

type TaskRequestProvider struct{}
type DecisionProvider struct{}

var _ rollbackcontract.RowSourceProvider = TaskRequestProvider{}
var _ rollbackcontract.RowSourceProvider = DecisionProvider{}

func NewTaskRequestProvider() TaskRequestProvider { return TaskRequestProvider{} }
func NewDecisionProvider() DecisionProvider       { return DecisionProvider{} }

func (TaskRequestProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := taskSourceForRollbackValue(value)
	if !ok || !validTaskSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (DecisionProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := decisionSourceForRollbackValue(value)
	if !ok || !validDecisionSource(source) {
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

type decisionMachine struct {
	incidentID         uuid.UUID
	status             string
	ownerUserID        *uuid.UUID
	decidedAt          *time.Time
	incomingSupersedes int
	outgoingSupersedes int
}

func (DecisionProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := decisionSourceForRollbackValue(request.RetainedValue)
	if !ok || !validDecisionSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	var machine decisionMachine
	if err := tx.QueryRow(ctx, `
SELECT d.incident_id, d.status, d.owner_user_id, d.decided_at,
       (SELECT COUNT(*) FROM active_record_links_v1 rl WHERE rl.incident_id = d.incident_id AND rl.dst_record_id = d.record_id AND rl.link_type = 'supersedes'),
       (SELECT COUNT(*) FROM active_record_links_v1 rl WHERE rl.incident_id = d.incident_id AND rl.src_record_id = d.record_id AND rl.link_type = 'supersedes')
  FROM decisions d
 WHERE d.record_id = $1
`, request.RecordID).Scan(&machine.incidentID, &machine.status, &machine.ownerUserID, &machine.decidedAt, &machine.incomingSupersedes, &machine.outgoingSupersedes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if raw, present := source["status"]; present {
		machine.status = raw.(string)
	}
	if value, present, err := nullableUUID(source, "owner_user_id"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		machine.ownerUserID = value
	}
	if value, present, err := nullableTime(source, "decided_at"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		machine.decidedAt = value
	}
	if !validDecisionMachine(machine) {
		return rollbackcontract.ErrTargetNotReversible
	}
	values, err := typedValues(source, []fieldSpec{
		{"summary", fieldText}, {"status", fieldText}, {"owner_user_id", fieldUUID},
		{"decision_type", fieldText}, {"decided_at", fieldTime}, {"rationale", fieldText},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE decisions
   SET summary = CASE WHEN $2 THEN $3::text ELSE summary END,
       status = CASE WHEN $4 THEN $5::text ELSE status END,
       owner_user_id = CASE WHEN $6 THEN $7::uuid ELSE owner_user_id END,
       decision_type = CASE WHEN $8 THEN $9::text ELSE decision_type END,
       decided_at = CASE WHEN $10 THEN $11::timestamptz ELSE decided_at END,
       rationale = CASE WHEN $12 THEN $13::text ELSE rationale END,
       updated_at = $14
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func taskSourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	return sourceForRollbackValue(value, map[string]string{
		"task.title": "title", "task.status": "status", "task.owner_user_id": "owner_user_id",
		"task.priority": "priority", "task.task_kind": "task_kind", "task.workstream": "workstream",
		"task.due_at": "due_at", "task.requester_party_text": "requester_party_text",
		"task.requester_party_id": "requester_party_id", "task.blocked_reason": "blocked_reason",
		"task.completed_at": "completed_at", "task.external_ticket_ref": "external_ticket_ref",
		"task.closure_summary": "closure_summary", "task.decision_record_id": "decision_record_id",
	})
}

func decisionSourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	return sourceForRollbackValue(value, map[string]string{
		"decision.summary": "summary", "decision.status": "status", "decision.owner_user_id": "owner_user_id",
		"decision.decision_type": "decision_type", "decision.decided_at": "decided_at",
		"decision.rationale": "rationale",
	})
}

func sourceForRollbackValue(value map[string]any, mapping map[string]string) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	if cells, ok := objectMap(value, "cells"); ok {
		source := map[string]any{}
		for fieldKey, sourceKey := range mapping {
			if cell, present := objectMap(cells, fieldKey); present {
				source[sourceKey] = cell["value"]
			}
		}
		return source, len(source) > 0
	}
	if _, ok := value["record_id"]; ok {
		for _, sourceKey := range mapping {
			if _, present := value[sourceKey]; present {
				return value, true
			}
		}
	}
	return nil, false
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

func validDecisionSource(source map[string]any) bool {
	if raw, present := source["summary"]; present && !nonEmptyText(raw) {
		return false
	}
	if raw, present := source["status"]; present && !policyText(raw, policy.ValidDecisionStatus) {
		return false
	}
	if raw, present := source["decision_type"]; present && !policyText(raw, policy.ValidDecisionType) {
		return false
	}
	return validTypedFields(source, []fieldSpec{{"owner_user_id", fieldUUID}, {"decided_at", fieldTime}})
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

func validDecisionMachine(state decisionMachine) bool {
	return policy.ValidateDecisionMachineState(policy.DecisionMachineState{
		Status: state.status, OwnerUserID: nullableSQLUUID(state.ownerUserID),
		DecidedAt: nullableSQLTime(state.decidedAt), IncomingSupersedes: state.incomingSupersedes,
		OutgoingSupersedes: state.outgoingSupersedes,
	}) == nil
}

func policyText(value any, valid func(string) bool) bool {
	text, ok := value.(string)
	return ok && valid(text)
}

func nullableSQLString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullableSQLTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}

func nullableSQLUUID(value *uuid.UUID) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
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

type fieldKind int

const (
	fieldText fieldKind = iota
	fieldUUID
	fieldTime
)

type fieldSpec struct {
	key  string
	kind fieldKind
}

func typedValues(source map[string]any, fields []fieldSpec) ([]any, error) {
	values := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		raw, present := source[field.key]
		var err error
		switch field.kind {
		case fieldUUID:
			raw, _, err = nullableUUID(source, field.key)
		case fieldTime:
			raw, _, err = nullableTime(source, field.key)
		case fieldText:
			if raw != nil {
				_, ok := raw.(string)
				if !ok {
					err = errors.New("invalid text")
				}
			}
		}
		if err != nil {
			return nil, err
		}
		values = append(values, present, raw)
	}
	return values, nil
}

func validTypedFields(source map[string]any, fields []fieldSpec) bool {
	_, err := typedValues(source, fields)
	return err == nil
}

func nullableUUID(value map[string]any, key string) (*uuid.UUID, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, errors.New("invalid uuid")
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	return &parsed, true, nil
}

func nullableTime(value map[string]any, key string) (*time.Time, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	if timestamp, ok := raw.(time.Time); ok {
		utc := timestamp.UTC()
		return &utc, true, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, errors.New("invalid timestamp")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return nil, true, err
	}
	utc := parsed.UTC()
	return &utc, true, nil
}

func nullableText(value map[string]any, key string) (*string, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok {
		return nil, true, errors.New("invalid text")
	}
	if strings.TrimSpace(text) == "" {
		return nil, true, nil
	}
	return &text, true, nil
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func nonEmptyText(value any) bool {
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}
