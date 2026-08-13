package evidence

import (
	"strings"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
)

func validateCreateParams(params createParams) error {
	if value, ok := params.Values["evidence.storage_ref"]; ok &&
		value.Text != nil &&
		evidencepolicy.IsServerManagedStorageRef(strings.TrimSpace(*value.Text)) {
		return &ValidationError{Field: "evidence.storage_ref", ReasonCode: "reserved_server_managed_ref"}
	}
	lifecycle, lifecyclePresent := params.Values["evidence.lifecycle_state"]
	if lifecyclePresent && (lifecycle.Text == nil || !evidencepolicy.ValidEvidenceLifecycle(*lifecycle.Text)) {
		return &ValidationError{Field: "evidence.lifecycle_state", ReasonCode: "invalid_value"}
	}
	effectiveLifecycle := "requested"
	if lifecyclePresent && lifecycle.Text != nil {
		effectiveLifecycle = *lifecycle.Text
	}
	hasBlob := params.InitialBlobFinalized
	switch evidencepolicy.InitialEvidenceLifecycleDisposition(effectiveLifecycle, params.InitialBlobWasSupplied, hasBlob) {
	case evidencepolicy.InitialLifecycleGuardViolation:
		return &LifecycleValidationError{
			FromStatus:     "",
			ToStatus:       effectiveLifecycle,
			ReasonCode:     "violated_lifecycle_guards",
			ViolatedGuards: []string{"evidence.lifecycle_state", "object_blobs.upload_state"},
		}
	case evidencepolicy.InitialLifecycleIllegalTransition:
		return &LifecycleValidationError{
			FromStatus:     "",
			ToStatus:       effectiveLifecycle,
			ReasonCode:     "illegal_status_transition",
			ViolatedGuards: []string{"evidence.lifecycle_state"},
		}
	}
	if !hasWorkbookText(params.Values, "evidence.title") &&
		!hasWorkbookText(params.Values, "evidence.storage_ref") &&
		!hasWorkbookText(params.Values, "evidence.collector_party_text") &&
		!hasWorkbookText(params.Values, "evidence.source_party_text") &&
		!lifecyclePresent &&
		!hasWorkbookTimestamp(params.Values, "evidence.requested_at") &&
		!hasWorkbookTimestamp(params.Values, "evidence.received_at") &&
		!hasBlob {
		if params.InitialBlobWasSupplied {
			return nil
		}
		return &ValidationError{Field: "payload", ReasonCode: "minimum_create_signal_missing"}
	}
	return nil
}

func validateDirectPatchChange(fieldKey string, value FieldValue) error {
	if fieldKey == "evidence.storage_ref" &&
		value.Text != nil &&
		evidencepolicy.IsServerManagedStorageRef(strings.TrimSpace(*value.Text)) {
		return &ValidationError{Field: fieldKey, ReasonCode: "reserved_server_managed_ref"}
	}
	if fieldKey == "evidence.lifecycle_state" && value.Text != nil && !evidencepolicy.ValidEvidenceLifecycle(*value.Text) {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	return nil
}

func hasWorkbookText(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func hasWorkbookTimestamp(values map[string]FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Timestamp != nil
}
