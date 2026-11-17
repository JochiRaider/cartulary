// Package policy owns the closed Task/Decision vocabulary and every
// entry-path-independent source rule.
package policy

import (
	"database/sql"
	"fmt"
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
	InvariantSourceState    = "tasks_decisions.source_state_exact"
	InvariantAttribution    = "tasks_decisions.attribution_exact"
	InvariantLifecycle      = "tasks_decisions.lifecycle_legal"
	InvariantDependent      = "tasks_decisions.dependent_fields_legal"
	InvariantReferences     = "tasks_decisions.references_same_incident"
	InvariantSourceIdentity = "tasks_decisions.source_identity_admitted"
)

func PortabilityInvariantIDs() []string {
	return []string{InvariantEnvelope, InvariantSourceState, InvariantAttribution, InvariantLifecycle, InvariantDependent, InvariantReferences, InvariantSourceIdentity}
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

// Lifecycle-transition reason family registered in Core 01 §3.3.6.2.
const (
	ReasonTransitionNotAllowed = "transition_not_allowed"
	ReasonGuardsNotSatisfied   = "guards_not_satisfied"
	ReasonInconsistentMachine  = "inconsistent_machine_state"
)

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
			return &LifecycleValidationError{ToStatus: status, ReasonCode: ReasonTransitionNotAllowed, ViolatedGuards: []string{}}
		}
	}
	return nil
}

// ValidateTaskState reports whether a Task row satisfies REQ-02-103.
func ValidateTaskState(state TaskLifecycleState) error {
	violated := TaskLifecycleViolations(state)
	if len(violated) == 0 {
		return nil
	}
	return &LifecycleValidationError{FromStatus: state.Status, ToStatus: state.Status, ReasonCode: ReasonGuardsNotSatisfied, ViolatedGuards: violated}
}

// TaskResultError validates the row a Task mutation would commit. fromStatus is
// the persisted status, or "" on create. When the persisted row already
// violated REQ-02-103 it was inconsistent (REQ-02-109a): a still-illegal result
// reports inconsistent_machine_state, and a legal result is the corrective
// action that commits.
func TaskResultError(fromStatus string, beforeInconsistent bool, after TaskLifecycleState) error {
	violated := TaskLifecycleViolations(after)
	if len(violated) == 0 {
		return nil
	}
	reason := ReasonGuardsNotSatisfied
	if beforeInconsistent {
		reason = ReasonInconsistentMachine
	}
	return &LifecycleValidationError{FromStatus: fromStatus, ToStatus: after.Status, ReasonCode: reason, ViolatedGuards: violated}
}

func TaskLifecycleViolations(state TaskLifecycleState) []string {
	violated := []string{}
	// REQ-02-103: a blank blocked_reason is absent ("absent or empty").
	hasReason := state.BlockedReason.Valid && strings.TrimSpace(state.BlockedReason.String) != ""
	if state.Status == "blocked" && !hasReason {
		violated = append(violated, "blocked_requires_reason")
	}
	if state.Status != "blocked" && hasReason {
		violated = append(violated, "non_blocked_clears_reason")
	}
	if state.Status == "done" {
		if !state.CompletedAt.Valid {
			violated = append(violated, "done_requires_completed_at")
		} else if state.CompletedAt.Time.Before(state.CreatedAt) {
			violated = append(violated, "completed_at_not_before_created_at")
		}
	}
	if state.Status != "done" && state.CompletedAt.Valid {
		violated = append(violated, "non_done_clears_completed_at")
	}
	if (state.Status == "open" || state.Status == "in_progress" || state.Status == "blocked") && !state.OwnerUserID.Valid {
		violated = append(violated, "active_requires_owner")
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
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: ReasonTransitionNotAllowed, ViolatedGuards: []string{}}
	}
	if from == to || (from == "proposed" && (to == "approved" || to == "rejected" || to == "executed")) || (from == "approved" && to == "executed") {
		return nil
	}
	return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: ReasonTransitionNotAllowed, ViolatedGuards: []string{}}
}

// ValidateDecisionMachineState reports a persisted Decision whose inputs do not
// resolve to one legal machine condition (Core 02 §10.4.2.1). An out-of-
// vocabulary status is a storage invariant failure, not an illegal transition.
func ValidateDecisionMachineState(state DecisionMachineState) error {
	if !ValidDecisionStatus(state.Status) {
		return fmt.Errorf("tasksdecisions: invalid persisted Decision status %q", state.Status)
	}
	violated := []string{}
	if !state.OwnerUserID.Valid {
		violated = append(violated, "owner_required")
	}
	if !state.DecidedAt.Valid {
		violated = append(violated, "decided_at_required")
	}
	if state.IncomingSupersedes > 1 {
		violated = append(violated, "single_incoming_supersedes")
	}
	if state.OutgoingSupersedes > 1 {
		violated = append(violated, "single_outgoing_supersedes")
	}
	if state.Status == "superseded" && state.IncomingSupersedes != 1 {
		violated = append(violated, "superseded_requires_incoming_supersedes")
	}
	if state.IncomingSupersedes == 1 && state.Status != "superseded" && state.Status != "executed" {
		violated = append(violated, "incoming_supersedes_requires_superseded_or_executed")
	}
	if state.OutgoingSupersedes > 0 && state.Status != "approved" && state.Status != "executed" {
		violated = append(violated, "outgoing_supersedes_requires_approved_or_executed")
	}
	if len(violated) == 0 {
		return nil
	}
	return &LifecycleValidationError{FromStatus: state.Status, ToStatus: state.Status, ReasonCode: ReasonInconsistentMachine, ViolatedGuards: violated}
}

// DecisionSupersedeValidationError rejects a supersede whose guards fail;
// fromStatus is the target Decision's persisted status.
func DecisionSupersedeValidationError(fromStatus string, guards ...string) error {
	return &LifecycleValidationError{FromStatus: fromStatus, ToStatus: "superseded", ReasonCode: ReasonGuardsNotSatisfied, ViolatedGuards: append([]string{}, guards...)}
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
