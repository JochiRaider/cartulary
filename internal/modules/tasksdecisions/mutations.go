package tasksdecisions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type Store struct {
	linkStore taskDecisionLinkPort
}

const TaskDecisionRecordFieldKey = "task.decision_record_id"

type taskDecisionLinkPort interface {
	SyncFieldReferenceCommandTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (bool, error)
	InsertSupersedesCommandTx(context.Context, pgx.Tx, links.InsertSupersedesCommand) (links.SupersedesLink, error)
}

type FieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type TaskCreateParams struct {
	Values map[string]FieldValue
}

type DecisionCreateParams struct {
	Values map[string]FieldValue
}

type LifecycleValidationError struct {
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
	ReasonCode     string
}

func (e *LifecycleValidationError) Error() string {
	return "tasksdecisions: illegal transition"
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return "tasksdecisions: invalid mutation request"
}

type TaskLifecycleState struct {
	Status        string
	BlockedReason sql.NullString
	CompletedAt   sql.NullTime
	OwnerUserID   sql.NullString
	CreatedAt     time.Time
}

type DecisionMachineState struct {
	RecordID             uuid.UUID
	IncidentID           uuid.UUID
	Status               string
	OwnerUserID          sql.NullString
	DecidedAt            sql.NullTime
	IncomingSupersedes   int
	OutgoingSupersedes   int
	IncomingSupersederID sql.NullString
	OutgoingTargetID     sql.NullString
}

func NewStore() *Store {
	return &Store{linkStore: links.NewStore()}
}

func ValidTaskKind(value string) bool {
	switch value {
	case "question", "request", "collection", "containment", "follow_up":
		return true
	default:
		return false
	}
}

func ValidTaskStatus(value string) bool {
	switch value {
	case "open", "in_progress", "blocked", "done", "canceled":
		return true
	default:
		return false
	}
}

func ValidTaskPriority(value string) bool {
	switch value {
	case "low", "normal", "high", "urgent":
		return true
	default:
		return false
	}
}

func ValidDecisionType(value string) bool {
	switch value {
	case "scope", "containment", "communication", "evidence", "reporting":
		return true
	default:
		return false
	}
}

func ValidDecisionStatus(value string) bool {
	switch value {
	case "proposed", "approved", "rejected", "superseded", "executed":
		return true
	default:
		return false
	}
}

func (s *Store) InsertTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params TaskCreateParams, now time.Time) error {
	status := nullableTextValue(params.Values, "task.status")
	if status == nil {
		status = "open"
	}
	owner := nullableUUIDValue(params.Values, "task.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	priority := nullableTextValue(params.Values, "task.priority")
	if priority == nil {
		priority = "normal"
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
		status = "proposed"
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
	if err := ValidateTaskDirectPatchChange(fieldKey, value); err != nil {
		return false, err
	}
	column := strings.TrimPrefix(fieldKey, "task.")
	dbValue := directDBValue(value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE task_requests SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, column, column), recordID, dbValue, now)
	if err != nil {
		return false, fmt.Errorf("apply task direct change: %w", err)
	}
	scalarChanged := tag.RowsAffected() > 0
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
	if err := ValidateDecisionDirectPatchChange(fieldKey, value); err != nil {
		return false, err
	}
	column := strings.TrimPrefix(fieldKey, "decision.")
	dbValue := directDBValue(value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE decisions SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, column, column), recordID, dbValue, now)
	if err != nil {
		return false, fmt.Errorf("apply decision direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
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
	var state TaskLifecycleState
	err := tx.QueryRow(ctx, `
SELECT status, NULLIF(blocked_reason, ''), completed_at, owner_user_id::text, created_at
  FROM task_requests
 WHERE record_id = $1
`, recordID).Scan(&state.Status, &state.BlockedReason, &state.CompletedAt, &state.OwnerUserID, &state.CreatedAt)
	if err != nil {
		return TaskLifecycleState{}, err
	}
	return state, nil
}

func (s *Store) NormalizeTaskLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, before TaskLifecycleState, explicitCompletedAt bool, now time.Time) (bool, error) {
	changed := false
	after, err := s.LoadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if !ValidTaskStatus(after.Status) {
		return false, &ValidationError{Field: "task.status", ReasonCode: "invalid_value"}
	}
	if before.Status != after.Status && !validTaskTransition(before.Status, after.Status) {
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
	violated := taskLifecycleViolations(state)
	if len(violated) > 0 {
		return &LifecycleValidationError{ToStatus: state.Status, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: violated}
	}
	return nil
}

func (s *Store) LoadDecisionStatusTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status FROM decisions WHERE record_id = $1`, recordID).Scan(&status)
	return status, err
}

func ValidateDecisionStatusTransition(from string, to string) error {
	if !ValidDecisionStatus(to) {
		return &ValidationError{Field: "decision.status", ReasonCode: "invalid_value"}
	}
	if to == "superseded" {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
	}
	if from == to {
		return nil
	}
	if (from == "proposed" && (to == "approved" || to == "rejected" || to == "executed")) ||
		(from == "approved" && to == "executed") {
		return nil
	}
	return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"decision.status"}}
}

func (s *Store) LoadDecisionMachineStateForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (DecisionMachineState, error) {
	var state DecisionMachineState
	if err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, status, owner_user_id::text, decided_at
  FROM decisions
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&state.RecordID, &state.IncidentID, &state.Status, &state.OwnerUserID, &state.DecidedAt); err != nil {
		return DecisionMachineState{}, err
	}
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*), MIN(src_record_id::text)
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
`, state.IncidentID, recordID).Scan(&state.IncomingSupersedes, &state.IncomingSupersederID); err != nil {
		return DecisionMachineState{}, fmt.Errorf("load decision incoming supersedes: %w", err)
	}
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*), MIN(dst_record_id::text)
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND link_type = 'supersedes'
`, state.IncidentID, recordID).Scan(&state.OutgoingSupersedes, &state.OutgoingTargetID); err != nil {
		return DecisionMachineState{}, fmt.Errorf("load decision outgoing supersedes: %w", err)
	}
	return state, nil
}

func (s *Store) ValidateDecisionMachineConsistentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	state, err := s.LoadDecisionMachineStateForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	return ValidateDecisionMachineState(state)
}

func ValidateDecisionMachineState(state DecisionMachineState) error {
	violated := []string{}
	if !ValidDecisionStatus(state.Status) {
		violated = append(violated, "decision_status_invalid")
	}
	if !state.OwnerUserID.Valid {
		violated = append(violated, "decision_owner_required")
	}
	if !state.DecidedAt.Valid {
		violated = append(violated, "decision_decided_at_required")
	}
	if state.IncomingSupersedes > 1 {
		violated = append(violated, "decision_incoming_supersedes_ambiguous")
	}
	if state.OutgoingSupersedes > 1 {
		violated = append(violated, "decision_outgoing_supersedes_ambiguous")
	}
	if state.Status == "superseded" && state.IncomingSupersedes != 1 {
		violated = append(violated, "superseded_requires_incoming_supersedes")
	}
	if state.IncomingSupersedes == 1 && state.Status != "superseded" && state.Status != "executed" {
		violated = append(violated, "incoming_supersedes_requires_superseded_or_executed")
	}
	if state.OutgoingSupersedes > 0 && state.Status != "approved" && state.Status != "executed" {
		violated = append(violated, "superseding_decision_requires_approved_or_executed")
	}
	if len(violated) == 0 {
		return nil
	}
	return &LifecycleValidationError{
		ToStatus:       state.Status,
		ReasonCode:     "inconsistent_decision_machine",
		ViolatedGuards: violated,
	}
}

func DecisionSupersedeValidationError(guards ...string) error {
	return &LifecycleValidationError{
		ReasonCode:     "decision_supersede_not_allowed",
		ViolatedGuards: append([]string(nil), guards...),
	}
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

func taskLifecycleViolations(state TaskLifecycleState) []string {
	violated := []string{}
	if state.Status == "blocked" && !state.BlockedReason.Valid {
		violated = append(violated, "blocked_requires_reason")
	}
	if state.Status != "blocked" && state.BlockedReason.Valid {
		violated = append(violated, "non_blocked_clears_reason")
	}
	if state.Status == "done" {
		if !state.CompletedAt.Valid {
			violated = append(violated, "done_requires_completed_at")
		} else if state.CompletedAt.Time.Before(state.CreatedAt) {
			violated = append(violated, "completed_at_before_created_at")
		}
	}
	if state.Status != "done" && state.CompletedAt.Valid {
		violated = append(violated, "non_done_clears_completed_at")
	}
	if (state.Status == "open" || state.Status == "in_progress" || state.Status == "blocked") && !state.OwnerUserID.Valid {
		violated = append(violated, "active_task_requires_owner")
	}
	return violated
}

func validTaskTransition(from string, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "open", "in_progress", "blocked":
		return to == "open" || to == "in_progress" || to == "blocked" || to == "done" || to == "canceled"
	case "done":
		return to == "open" || to == "in_progress" || to == "blocked"
	case "canceled":
		return to == "open" || to == "in_progress" || to == "blocked"
	default:
		return false
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

func IsUniqueViolation(err error) bool {
	return authn.IsUniqueViolation(err)
}
