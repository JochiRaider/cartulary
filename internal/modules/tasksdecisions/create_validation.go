package tasksdecisions

import "strings"

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
		"task.decision_record_id":
		return nil
	case "task.status":
		if value.Text != nil && !ValidTaskStatus(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
		return nil
	case "task.task_kind":
		if value.Text != nil && !ValidTaskKind(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
		return nil
	case "task.priority":
		if value.Text != nil && !ValidTaskPriority(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
		return nil
	default:
		return &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
}

func ValidateDecisionDirectPatchChange(fieldKey string, value FieldValue) error {
	switch fieldKey {
	case "decision.summary", "decision.owner_user_id", "decision.decided_at", "decision.rationale":
		return nil
	case "decision.status":
		if value.Text != nil && !ValidDecisionStatus(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
		return nil
	case "decision.decision_type":
		if value.Text != nil && !ValidDecisionType(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
		return nil
	default:
		return &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
}

func hasText(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func validText(values map[string]FieldValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return false
	}
	return predicate(*value.Text)
}

func derefText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
