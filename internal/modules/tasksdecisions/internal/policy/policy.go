// Package policy owns the closed Task/Decision vocabulary and every
// entry-path-independent source rule.
package policy

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultTaskStatus       = "open"
	DefaultTaskPriority     = "normal"
	DefaultDecisionStatus   = "proposed"
	TaskDecisionRecordField = "task.decision_record_id"
	InvariantEnvelope       = "tasks_decisions.envelope_type_scope"
	InvariantLifecycle      = "tasks_decisions.lifecycle_legal"
	InvariantDependent      = "tasks_decisions.dependent_fields_legal"
	InvariantReferences     = "tasks_decisions.references_same_incident"
	InvariantSourceIdentity = "tasks_decisions.source_identity_admitted"
)

func PortabilityInvariantIDs() []string {
	return []string{InvariantEnvelope, InvariantLifecycle, InvariantDependent, InvariantReferences, InvariantSourceIdentity}
}

func IsPortabilityInvariant(value string) bool {
	for _, candidate := range PortabilityInvariantIDs() {
		if value == candidate {
			return true
		}
	}
	return false
}

type FieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type TaskCreateParams struct{ Values map[string]FieldValue }
type DecisionCreateParams struct{ Values map[string]FieldValue }

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (*ValidationError) Error() string { return "tasksdecisions: invalid mutation request" }

type LifecycleValidationError struct {
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
	ReasonCode     string
}

func (*LifecycleValidationError) Error() string { return "tasksdecisions: illegal transition" }

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

func ValidateTaskCreateParams(params TaskCreateParams) error {
	values := params.Values
	if !hasText(values, "task.title") {
		return &ValidationError{Field: "task.title", ReasonCode: "missing_required_field"}
	}
	if !validText(values, "task.task_kind", ValidTaskKind) {
		return &ValidationError{Field: "task.task_kind", ReasonCode: "missing_required_field"}
	}
	if value, ok := values["task.status"]; ok && !ValidTaskStatus(derefText(value.Text)) {
		return &ValidationError{Field: "task.status", ReasonCode: "invalid_value"}
	}
	if value, ok := values["task.priority"]; ok && !ValidTaskPriority(derefText(value.Text)) {
		return &ValidationError{Field: "task.priority", ReasonCode: "invalid_value"}
	}
	return nil
}

func ValidateDecisionCreateParams(params DecisionCreateParams) error {
	values := params.Values
	if !hasText(values, "decision.summary") {
		return &ValidationError{Field: "decision.summary", ReasonCode: "missing_required_field"}
	}
	if !validText(values, "decision.decision_type", ValidDecisionType) {
		return &ValidationError{Field: "decision.decision_type", ReasonCode: "missing_required_field"}
	}
	if !hasText(values, "decision.rationale") {
		return &ValidationError{Field: "decision.rationale", ReasonCode: "missing_required_field"}
	}
	if value, ok := values["decision.status"]; ok {
		status := derefText(value.Text)
		if !ValidDecisionStatus(status) {
			return &ValidationError{Field: "decision.status", ReasonCode: "invalid_value"}
		}
		if status == "superseded" {
			return &LifecycleValidationError{ToStatus: status, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
		}
	}
	return nil
}

func ValidateTaskDirectPatchChange(fieldKey string, value FieldValue) error {
	switch fieldKey {
	case "task.title", "task.owner_user_id", "task.workstream", "task.due_at",
		"task.requester_party_text", "task.requester_party_id", "task.blocked_reason",
		"task.completed_at", "task.external_ticket_ref", "task.closure_summary",
		TaskDecisionRecordField:
		return nil
	case "task.status":
		if value.Text != nil && !ValidTaskStatus(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "task.task_kind":
		if value.Text != nil && !ValidTaskKind(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "task.priority":
		if value.Text != nil && !ValidTaskPriority(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	default:
		return &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	return nil
}

func ValidateDecisionDirectPatchChange(fieldKey string, value FieldValue) error {
	switch fieldKey {
	case "decision.summary", "decision.owner_user_id", "decision.decided_at", "decision.rationale":
		return nil
	case "decision.status":
		if value.Text != nil && !ValidDecisionStatus(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "decision.decision_type":
		if value.Text != nil && !ValidDecisionType(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	default:
		return &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	return nil
}

func ValidateTaskState(state TaskLifecycleState) error {
	violated := TaskLifecycleViolations(state)
	if len(violated) == 0 {
		return nil
	}
	return &LifecycleValidationError{ToStatus: state.Status, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: violated}
}

func TaskLifecycleViolations(state TaskLifecycleState) []string {
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

func ValidTaskTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "open", "in_progress", "blocked":
		return to == "open" || to == "in_progress" || to == "blocked" || to == "done" || to == "canceled"
	case "done", "canceled":
		return to == "open" || to == "in_progress" || to == "blocked"
	default:
		return false
	}
}

func ValidateDecisionStatusTransition(from, to string) error {
	if !ValidDecisionStatus(to) {
		return &ValidationError{Field: "decision.status", ReasonCode: "invalid_value"}
	}
	if to == "superseded" {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
	}
	if from == to || (from == "proposed" && (to == "approved" || to == "rejected" || to == "executed")) || (from == "approved" && to == "executed") {
		return nil
	}
	return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"decision.status"}}
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
	return &LifecycleValidationError{ToStatus: state.Status, ReasonCode: "inconsistent_decision_machine", ViolatedGuards: violated}
}

func DecisionSupersedeValidationError(guards ...string) error {
	return &LifecycleValidationError{ReasonCode: "decision_supersede_not_allowed", ViolatedGuards: append([]string(nil), guards...)}
}

func IsMemberUserReferenceField(fieldKey string) bool {
	return strings.HasSuffix(fieldKey, "_user_id")
}

func DirectReferenceRecordType(fieldKey string) (string, bool) {
	switch fieldKey {
	case "task.requester_party_id":
		return "party", true
	case TaskDecisionRecordField:
		return "decision", true
	default:
		return "", false
	}
}

func hasText(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func validText(values map[string]FieldValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	return ok && value.Text != nil && predicate(*value.Text)
}

func derefText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
