package artifacts

import "strings"

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return "artifacts: invalid mutation request"
}

func ValidateCreateParams(params CreateParams) error {
	values := params.Values
	switch params.ViewSchemaID {
	case NotesViewSchemaID:
		if !hasText(values, "note.title") && !hasText(values, "note.body") {
			return &ValidationError{Field: "payload", ReasonCode: "missing_minimum_create_signal"}
		}
	case CommLogViewSchemaID:
		for _, field := range []string{"comm_log.comm_type", "comm_log.audience", "comm_log.channel_or_meeting", "comm_log.summary"} {
			if !hasText(values, field) {
				return &ValidationError{Field: field, ReasonCode: "missing_required_field"}
			}
		}
		if !validText(values, "comm_log.comm_type", ValidCommType) {
			return &ValidationError{Field: "comm_log.comm_type", ReasonCode: "invalid_value"}
		}
	case HandoffViewSchemaID:
		if !hasUUID(values, "handoff.incoming_owner_user_id") {
			return &ValidationError{Field: "handoff.incoming_owner_user_id", ReasonCode: "missing_required_field"}
		}
		if !hasText(values, "handoff.current_state_summary") {
			return &ValidationError{Field: "handoff.current_state_summary", ReasonCode: "missing_required_field"}
		}
	case StatusReviewViewSchemaID:
		if !hasText(values, "status_review.current_state_summary") {
			return &ValidationError{Field: "status_review.current_state_summary", ReasonCode: "missing_required_field"}
		}
	case LessonViewSchemaID:
		if !hasText(values, "lesson.summary") {
			return &ValidationError{Field: "lesson.summary", ReasonCode: "missing_required_field"}
		}
		if value, ok := values["lesson.closure_state"]; ok && !ValidClosureState(derefText(value.Text)) {
			return &ValidationError{Field: "lesson.closure_state", ReasonCode: "invalid_value"}
		}
	case FindingsViewSchemaID:
		if !hasText(values, "finding.statement") {
			return &ValidationError{Field: "finding.statement", ReasonCode: "missing_required_field"}
		}
		if value, ok := values["finding.kind"]; ok && !ValidFindingKind(derefText(value.Text)) {
			return &ValidationError{Field: "finding.kind", ReasonCode: "invalid_value"}
		}
		if value, ok := values["finding.state"]; ok && !ValidFindingState(derefText(value.Text)) {
			return &ValidationError{Field: "finding.state", ReasonCode: "invalid_value"}
		}
		if value, ok := values["finding.confidence_score"]; ok && value.Number != nil && !ValidConfidenceScore(*value.Number) {
			return &ValidationError{Field: "finding.confidence_score", ReasonCode: "invalid_value"}
		}
	case InvestigativeQueriesViewSchemaID:
		for _, field := range []string{"investigative_query.platform", "investigative_query.purpose", "investigative_query.query_text"} {
			if !hasText(values, field) {
				return &ValidationError{Field: field, ReasonCode: "missing_required_field"}
			}
		}
	case ForensicKeywordsViewSchemaID:
		for _, field := range []string{"forensic_keyword.pattern", "forensic_keyword.reason"} {
			if !hasText(values, field) {
				return &ValidationError{Field: field, ReasonCode: "missing_required_field"}
			}
		}
		if value, ok := values["forensic_keyword.match_mode"]; ok && !ValidForensicKeywordMatchMode(derefText(value.Text)) {
			return &ValidationError{Field: "forensic_keyword.match_mode", ReasonCode: "invalid_value"}
		}
	}
	return nil
}

func ValidateDirectPatchChange(fieldKey string, value FieldValue) error {
	if fieldKey == "finding.confidence_score" && value.Number != nil && !ValidConfidenceScore(*value.Number) {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	if value.Text == nil {
		return nil
	}
	switch fieldKey {
	case "comm_log.comm_type":
		if !ValidCommType(*value.Text) {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "lesson.closure_state":
		if *value.Text != "open" && *value.Text != "closed" {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "finding.kind":
		if *value.Text != "finding" && *value.Text != "hypothesis" {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "finding.state":
		if *value.Text != "open" && *value.Text != "closed" {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	case "forensic_keyword.match_mode":
		if *value.Text != "literal" && *value.Text != "regex" {
			return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
		}
	}
	return nil
}

func ValidCommType(value string) bool {
	switch value {
	case "meeting", "notification", "approval", "briefing", "handoff":
		return true
	default:
		return false
	}
}

func ValidClosureState(value string) bool {
	switch value {
	case "open", "closed", "":
		return true
	default:
		return false
	}
}

func ValidFindingKind(value string) bool {
	switch value {
	case "finding", "hypothesis", "":
		return true
	default:
		return false
	}
}

func ValidFindingState(value string) bool {
	switch value {
	case "open", "closed", "":
		return true
	default:
		return false
	}
}

func ValidForensicKeywordMatchMode(value string) bool {
	switch value {
	case "literal", "regex", "":
		return true
	default:
		return false
	}
}

func ValidConfidenceScore(value int64) bool {
	return value >= 0 && value <= 100
}

func hasText(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func hasUUID(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil
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
