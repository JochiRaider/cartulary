package evidence

import "strings"

func ValidateWorkbookCreateParams(params WorkbookCreateParams) error {
	if !hasWorkbookText(params.Values, "evidence.title") &&
		!hasWorkbookText(params.Values, "evidence.storage_ref") &&
		!hasWorkbookText(params.Values, "evidence.collector_party_text") &&
		!hasWorkbookText(params.Values, "evidence.source_party_text") {
		return &ValidationError{Field: "payload", ReasonCode: "missing_minimum_create_signal"}
	}
	if value, ok := params.Values["evidence.lifecycle_state"]; ok && !ValidLifecycleState(derefWorkbookText(value.Text)) {
		return &ValidationError{Field: "evidence.lifecycle_state", ReasonCode: "invalid_value"}
	}
	return nil
}

func ValidateWorkbookDirectPatchChange(fieldKey string, value WorkbookFieldValue) error {
	if fieldKey == "evidence.lifecycle_state" && value.Text != nil && !ValidLifecycleState(*value.Text) {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	return nil
}

func hasWorkbookText(values map[string]WorkbookFieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func derefWorkbookText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
