package tasksdecisions

import "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"

// These root wrappers preserve the adopted facade while the implementation is
// centralized in the private policy kernel.
func ValidateTaskCreateParams(params TaskCreateParams) error {
	return policy.ValidateTaskCreateParams(params)
}

func ValidateDecisionCreateParams(params DecisionCreateParams) error {
	return policy.ValidateDecisionCreateParams(params)
}

func ValidateTaskDirectPatchChange(fieldKey string, value FieldValue) error {
	return policy.ValidateTaskDirectPatchChange(fieldKey, value)
}

func ValidateDecisionDirectPatchChange(fieldKey string, value FieldValue) error {
	return policy.ValidateDecisionDirectPatchChange(fieldKey, value)
}
