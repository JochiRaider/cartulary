package tasksdecisions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

type Store struct {
	linkStore        taskDecisionLinkPort
	revisionAppender *revisions.Appender
}

const TaskDecisionRecordFieldKey = policy.TaskDecisionRecordField

type taskDecisionLinkPort interface {
	SyncFieldReferenceCommandTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (bool, error)
	InsertSupersedesCommandTx(context.Context, pgx.Tx, links.InsertSupersedesCommand) (links.SupersedesLink, error)
}

type FieldValue = policy.FieldValue
type TaskCreateParams = policy.TaskCreateParams
type DecisionCreateParams = policy.DecisionCreateParams
type LifecycleValidationError = policy.LifecycleValidationError
type ValidationError = policy.ValidationError
type TaskLifecycleState = policy.TaskLifecycleState
type DecisionMachineState = policy.DecisionMachineState

func NewStore(appender *revisions.Appender) *Store {
	return newStoreWithLinks(links.NewStore(), appender)
}

func newStoreWithLinks(linkStore taskDecisionLinkPort, appender *revisions.Appender) *Store {
	if linkStore == nil {
		panic("tasksdecisions link capability is required")
	}
	return &Store{linkStore: linkStore, revisionAppender: appender}
}

func ValidTaskKind(value string) bool {
	return policy.ValidTaskKind(value)
}

func ValidTaskStatus(value string) bool {
	return policy.ValidTaskStatus(value)
}

func ValidTaskPriority(value string) bool {
	return policy.ValidTaskPriority(value)
}

func ValidDecisionType(value string) bool {
	return policy.ValidDecisionType(value)
}

func ValidDecisionStatus(value string) bool {
	return policy.ValidDecisionStatus(value)
}

func (s *Store) InsertTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params TaskCreateParams, now time.Time) error {
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
	if err := ValidateTaskCreateState(TaskLifecycleState{
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

func (s *Store) InsertDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params DecisionCreateParams, now time.Time) error {
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

func (s *Store) ApplyTaskDirectChangeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, value FieldValue, now time.Time) (bool, error) {
	if err := policy.ValidateTaskDirectPatchChange(fieldKey, value); err != nil {
		return false, err
	}
	scalarChanged, err := tasksource.ApplyTaskDirectChangeTx(ctx, tx, recordID, fieldKey, value, now)
	if err != nil {
		return false, err
	}
	if fieldKey == TaskDecisionRecordFieldKey {
		linkChanged, err := s.linkStore.SyncFieldReferenceCommandTx(ctx, tx, links.SyncFieldReferenceCommand{
			IncidentID:  incidentID,
			SrcRecordID: recordID,
			TargetID:    value.UUID,
			FieldKey:    TaskDecisionRecordFieldKey,
			LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
			ActorUserID: actorID,
			Now:         now,
		})
		if err != nil {
			return false, err
		}
		return scalarChanged || linkChanged, nil
	}
	return scalarChanged, nil
}

func (s *Store) ApplyDecisionDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string, value FieldValue, now time.Time) (bool, error) {
	if err := policy.ValidateDecisionDirectPatchChange(fieldKey, value); err != nil {
		return false, err
	}
	return tasksource.ApplyDecisionDirectChangeTx(ctx, tx, recordID, fieldKey, value, now)
}

func (s *Store) TouchTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE task_requests SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch task request row: %w", err)
	}
	return nil
}

func (s *Store) TouchDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE decisions SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch decision row: %w", err)
	}
	return nil
}

func (s *Store) LoadTaskLifecycleStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (TaskLifecycleState, error) {
	return tasksource.LoadTaskLifecycleStateTx(ctx, tx, recordID)
}

func (s *Store) NormalizeTaskLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, before TaskLifecycleState, explicitCompletedAt bool, now time.Time) (bool, error) {
	changed := false
	after, err := s.LoadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if !policy.ValidTaskStatus(after.Status) {
		return false, &ValidationError{Field: "task.status", ReasonCode: "invalid_value"}
	}
	if before.Status != after.Status && !policy.ValidTaskTransition(before.Status, after.Status) {
		return false, &LifecycleValidationError{FromStatus: before.Status, ToStatus: after.Status, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"task.status"}}
	}
	if before.Status != after.Status && before.Status == "blocked" && after.Status != "blocked" && after.BlockedReason.Valid {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET blocked_reason = NULL WHERE record_id = $1`, recordID); err != nil {
			return false, fmt.Errorf("clear blocked reason: %w", err)
		}
		changed = true
	}
	if before.Status != after.Status && before.Status == "done" && after.Status != "done" && after.CompletedAt.Valid {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET completed_at = NULL WHERE record_id = $1`, recordID); err != nil {
			return false, fmt.Errorf("clear completed at: %w", err)
		}
		changed = true
	}
	if after.Status == "done" && !after.CompletedAt.Valid && !explicitCompletedAt {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET completed_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
			return false, fmt.Errorf("fill completed at: %w", err)
		}
		changed = true
	}
	after, err = s.LoadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if err := ValidateTaskCreateState(after); err != nil {
		return false, err
	}
	return changed, nil
}

func ValidateTaskCreateState(state TaskLifecycleState) error {
	return policy.ValidateTaskState(state)
}

func (s *Store) LoadDecisionStatusTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status FROM decisions WHERE record_id = $1`, recordID).Scan(&status)
	return status, err
}

func ValidateDecisionStatusTransition(from string, to string) error {
	return policy.ValidateDecisionStatusTransition(from, to)
}

func (s *Store) LoadDecisionMachineStateForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (DecisionMachineState, error) {
	return tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, recordID)
}

func (s *Store) ValidateDecisionMachineConsistentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	state, err := s.LoadDecisionMachineStateForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	return ValidateDecisionMachineState(state)
}

func ValidateDecisionMachineState(state DecisionMachineState) error {
	return policy.ValidateDecisionMachineState(state)
}

func DecisionSupersedeValidationError(guards ...string) error {
	return policy.DecisionSupersedeValidationError(guards...)
}

func (s *Store) InsertDecisionSupersedesLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, actorID uuid.UUID, now time.Time) (uuid.UUID, error) {
	link, err := s.linkStore.InsertSupersedesCommandTx(ctx, tx, links.InsertSupersedesCommand{
		IncidentID:          incidentID,
		ReplacementRecordID: sourceID,
		SupersededRecordID:  targetID,
		OwnerUserID:         actorID,
		Now:                 now,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return link.RecordLinkID, nil
}

func (s *Store) TouchSupersedingDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
UPDATE decisions
   SET updated_at = $2
 WHERE record_id = $1
`, recordID, now); err != nil {
		return fmt.Errorf("update superseding decision: %w", err)
	}
	return nil
}

func (s *Store) MarkSupersededDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
UPDATE decisions
   SET status = CASE WHEN status IN ('proposed', 'approved') THEN 'superseded' ELSE status END,
       updated_at = $2
 WHERE record_id = $1
`, recordID, now); err != nil {
		return fmt.Errorf("update superseded decision: %w", err)
	}
	return nil
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

func NullableUUIDPointer(value any) *uuid.UUID {
	if id, ok := value.(uuid.UUID); ok {
		return &id
	}
	return nil
}
