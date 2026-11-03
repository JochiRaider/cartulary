package extensions

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/validation"
)

func (r *StateRuntime) classifyStateResult(profile, phase, algorithm, schema string, result StateValidationResult, invocationErr error, invocationFailure StartupFinding) error {
	var fatal *FatalStateIntegrityError
	if errors.As(invocationErr, &fatal) {
		return invocationErr
	}
	disposition := validation.Classify(invocationErr, result.Findings != nil, len(result.Findings), 256, func() bool {
		if result.SchemaID != schema || (result.Status != "valid" && result.Status != "invalid") || (result.Status == "valid") != (len(result.Findings) == 0) {
			return false
		}
		// The currently registered generic state conditions identify the whole
		// profile or migration operation. A callback cannot substitute data or
		// a storage locator into that registered diagnostic frame.
		for i, finding := range result.Findings {
			if !r.coordinator.validStartupFinding(finding) || !equalJSONValue(finding, invocationFailure) || (i > 0 && compareStartupFinding(result.Findings[i-1], finding) > 0) {
				return false
			}
		}
		return true
	})
	var findings []StartupFinding
	switch disposition {
	case validation.Success:
		return nil
	case validation.InvocationFailed:
		findings = []StartupFinding{invocationFailure}
	case validation.Overflow:
		findings = []StartupFinding{r.coordinator.validationOverflow(phase)}
	case validation.Findings:
		findings = make([]StartupFinding, len(result.Findings))
		for i, finding := range result.Findings {
			finding.Details = cloneObject(finding.Details)
			findings[i] = finding
		}
		sortStartupFindings(findings)
	default:
		condition := "extension_validation_remaining_schema_defect"
		if disposition == validation.StructureInvalid {
			condition = "extension_validation_result_structural_invalid"
		}
		findings = []StartupFinding{r.coordinator.registeredResultFinding(resultConditionID(condition, phase), invalidValidationFinding(profile, phase, algorithm))}
	}
	return &StartupValidationError{Findings: findings, cause: errors.Join(ErrStateValidationFailed, invocationErr)}
}
