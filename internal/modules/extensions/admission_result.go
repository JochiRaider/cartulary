package extensions

import (
	"sort"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/validation"
)

func admissionPhase(phase string) string {
	if phase == "post_migration" {
		return "post_migration_validation"
	}
	if phase == "dependency_probe" {
		return "dependency_probe"
	}
	return "profile_preflight"
}

func (c *Coordinator) classifyAdmissionResult(input AdmissionValidationContext, algorithmID string, result AdmissionValidationResult, invocationErr error, timedOut bool) []StartupFinding {
	disposition := validation.Classify(invocationErr, result.Findings != nil, len(result.Findings), validation.FindingLimit, func() bool {
		if result.SchemaID != "cartulary.extension_admission_validation_result.v1" || result.Phase != input.Phase || result.AlgorithmID != algorithmID {
			return false
		}
		for _, finding := range result.Findings {
			if !c.validStartupFinding(finding) || !equalJSONValue(finding, admissionFailure(input, algorithmID, false)) {
				return false
			}
		}
		return true
	})
	switch disposition {
	case validation.Success:
		return nil
	case validation.Findings:
		findings := make([]StartupFinding, len(result.Findings))
		for i, finding := range result.Findings {
			finding.Details = cloneObject(finding.Details)
			findings[i] = finding
		}
		sortStartupFindings(findings)
		return findings
	case validation.InvocationFailed:
		finding := admissionFailure(input, algorithmID, timedOut)
		return []StartupFinding{c.registeredResultFinding(resultConditionID("extension_validation_invocation_failure", admissionPhase(input.Phase)), finding)}
	case validation.Overflow:
		return []StartupFinding{c.validationOverflow(admissionPhase(input.Phase))}
	default:
		condition := "extension_validation_remaining_schema_defect"
		if disposition == validation.StructureInvalid {
			condition = "extension_validation_result_structural_invalid"
		}
		return []StartupFinding{c.registeredResultFinding(resultConditionID(condition, admissionPhase(input.Phase)), invalidValidationFinding(input.ProfileID, admissionPhase(input.Phase), algorithmID))}
	}
}

func invalidValidationFinding(profile, phase, algorithm string) StartupFinding {
	var profileValue, algorithmValue any
	if profile != "" {
		profileValue = profile
	}
	if algorithm != "" {
		algorithmValue = algorithm
	}
	return StartupFinding{Path: "$", ReasonCode: "extension_validation_result_invalid", Message: "An extension validation algorithm returned an invalid result.", Details: map[string]any{"profile_id": profileValue, "phase": phase, "algorithm_id": algorithmValue, "actual": "redacted"}}
}

func (c *Coordinator) registeredResultFinding(conditionID string, finding StartupFinding) StartupFinding {
	condition, err := c.RequireRegisteredCondition(conditionID)
	if phase, ok := finding.Details["phase"].(string); ok && phase != condition.Phase {
		return invalidValidationFinding("", "profile_preflight", "")
	}
	if err != nil || condition.ReasonCode != finding.ReasonCode || !c.validStartupFinding(finding) {
		return invalidValidationFinding("", "profile_preflight", "")
	}
	return finding
}

func (c *Coordinator) validationOverflow(phase string) StartupFinding {
	return c.registeredResultFinding(resultConditionID("extension_validation_result_overflow", phase), StartupFinding{Path: "$", ReasonCode: "extension_diagnostic_overflow", Message: "Extension validation produced too many findings.", Details: map[string]any{"phase": phase, "limit": validation.FindingLimit, "actual": validation.FindingLimit + 1}})
}

func sortStartupFindings(findings []StartupFinding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].ReasonCode != findings[j].ReasonCode {
			return findings[i].ReasonCode < findings[j].ReasonCode
		}
		left, _ := canonicalJSON(findings[i].Details, false)
		right, _ := canonicalJSON(findings[j].Details, false)
		return string(left) < string(right)
	})
}

// The compiled schema belongs to the sealed registry. Library diagnostics never
// cross this boundary; only a registered, closed owner finding may be emitted.
func (c *Coordinator) validStartupFinding(finding StartupFinding) bool {
	if c == nil || c.validateFinding == nil || !validDiagnosticPath(finding.Path) {
		return false
	}
	encoded, err := canonicaljson.MarshalExtension(finding)
	if err != nil {
		return false
	}
	object, err := canonicaljson.DecodeStrict(encoded)
	return err == nil && c.validateFinding(object) == nil
}

// StateValidationFailure constructs the generic owner mapping for final or
// pending state findings, without carrying underlying storage errors.
func (c *Coordinator) StateValidationFailure(profile, phase, algorithm string) StartupFinding {
	return c.registeredResultFinding("extension_state_invocation", StartupFinding{Path: "$", ReasonCode: "extension_state_validation_failed", Message: "Extension state failed final validation.", Details: map[string]any{"profile_id": profile, "phase": phase, "algorithm_id": algorithm}})
}

// ValidateTransactionFinding admits only registered owner diagnostics with the
// bound profile/algorithm identities, using the sealed startup-finding schema.
func (c *Coordinator) ValidateTransactionFinding(profile, algorithm string, finding StartupFinding) bool {
	// Current participant semantic findings use the registered profile-wide
	// admission condition. Coordinator-owned deadlines are not owner findings.
	if finding.Path != "$" || finding.ReasonCode != "extension_admission_validation_failed" {
		return false
	}
	if !c.validStartupFinding(finding) {
		return false
	}
	if value, exists := finding.Details["profile_id"]; exists && value != profile {
		return false
	}
	if value, exists := finding.Details["algorithm_id"]; exists && value != algorithm {
		return false
	}
	for _, condition := range c.validationConditions {
		if condition.Phase == "transaction" && condition.ReasonCode == finding.ReasonCode {
			if phase, exists := finding.Details["phase"]; exists && phase != "transaction" {
				return false
			}
			return true
		}
	}
	return false
}

func (c *Coordinator) migrationValidationFailure(profile, migration string, from, to int) StartupFinding {
	return c.registeredResultFinding("extension_migration_validation_invocation_failure", StartupFinding{Path: "$", ReasonCode: "extension_migration_failed", Message: "Extension state migration failed.", Details: map[string]any{"profile_id": profile, "migration_id": migration, "from_state_version": from, "to_state_version": to}})
}

func compareStartupFinding(a, b StartupFinding) int {
	if a.Path < b.Path {
		return -1
	}
	if a.Path > b.Path {
		return 1
	}
	if a.ReasonCode < b.ReasonCode {
		return -1
	}
	if a.ReasonCode > b.ReasonCode {
		return 1
	}
	left, _ := canonicalJSON(a.Details, false)
	right, _ := canonicalJSON(b.Details, false)
	if string(left) < string(right) {
		return -1
	}
	if string(left) > string(right) {
		return 1
	}
	return 0
}

func resultConditionID(base, phase string) string {
	if phase == "profile_preflight" {
		return base
	}
	return base + "_" + phase
}

// TransactionValidationFailure projects a registered result-classification
// failure into the same closed diagnostic contract used at startup.
func (c *Coordinator) TransactionValidationFailure(conditionID, profile, algorithm string, timeoutSeconds int64, timedOut bool) (StartupFinding, bool) {
	condition, err := c.RequireRegisteredCondition(conditionID)
	if err != nil || condition.Phase != "transaction" {
		return StartupFinding{}, false
	}
	var finding StartupFinding
	switch condition.ReasonCode {
	case "extension_diagnostic_overflow":
		finding = c.validationOverflow("transaction")
	case "extension_validation_result_invalid":
		finding = invalidValidationFinding(profile, "transaction", algorithm)
	case "extension_admission_validation_failed":
		finding = StartupFinding{Path: "$", ReasonCode: condition.ReasonCode, Message: "Extension admission validation failed.", Details: map[string]any{"profile_id": profile, "phase": "transaction", "algorithm_id": algorithm, "timed_out": timedOut, "timeout_seconds": timeoutSeconds}}
	default:
		return StartupFinding{}, false
	}
	finding = c.registeredResultFinding(conditionID, finding)
	return finding, finding.ReasonCode == condition.ReasonCode
}
