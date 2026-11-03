package extensions

import (
	"errors"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

// Dependency diagnostics retain the complete graph finding in the error chain
// and emit only the registered, closed public projection.
func (c *Coordinator) dependencyDiagnostics(err error) error {
	var failure *ValidationError
	if !errors.As(err, &failure) {
		return err
	}
	diagnostics := make([]StartupFinding, 0, len(failure.findings))
	for _, finding := range failure.findings {
		var diagnostic StartupFinding
		switch finding.Code {
		case "extension_diagnostic_overflow":
			definition, _ := (registryfacts.OwnerFinding{ConditionID: registryfacts.DependencyOverflowCondition}).Definition()
			diagnostic = c.registeredResultFinding(registryfacts.DependencyOverflowCondition, StartupFinding{
				Path: "$", ReasonCode: definition.ReasonCode, Message: definition.Message,
				Details: map[string]any{"phase": definition.Phase, "limit": registryfacts.DiagnosticLimit, "actual": registryfacts.DiagnosticLimit + 1},
			})
		case "extension_registry_conflict":
			var count any
			if finding.DuplicateCount != 0 {
				count = finding.DuplicateCount
			}
			diagnostic = c.registeredResultFinding("extension_registry_collision", StartupFinding{
				Path: "$", ReasonCode: finding.Code, Message: "The extension registry contains conflicting declarations.",
				Details: map[string]any{"collision_class": finding.CollisionClass, "conflicting_profile_ids": append([]string{}, finding.ConflictingProfileIDs...), "conflicting_tokens": append([]string{}, finding.ConflictingTokens...), "duplicate_count": count},
			})
		case "extension_dependency_not_claimed":
			diagnostic = c.registeredResultFinding("extension_dependency_not_claimed", StartupFinding{Path: "$", ReasonCode: finding.Code, Message: "A required extension dependency was not explicitly claimed.", Details: map[string]any{"profile_id": finding.ProfileID, "dependency_profile_id": finding.DependencyProfileID}})
		case "extension_profile_not_claimable":
			diagnostic = c.registeredResultFinding("extension_profile_not_claimable", StartupFinding{Path: "$", ReasonCode: finding.Code, Message: "The requested extension profile is not claimable.", Details: map[string]any{"profile_id": finding.ProfileID}})
		case "extension_implementation_unavailable":
			expected := c.profiles[finding.ProfileID].descriptor.ContractMajor
			diagnostic = StartupFinding{Path: "$", ReasonCode: finding.Code, Message: "The requested extension implementation is unavailable or incompatible.", Details: map[string]any{"profile_id": finding.ProfileID, "expected_contract_major": expected, "actual_contract_major": nil}}
			diagnostic = c.registeredResultFinding("extension_implementation_unavailable", diagnostic)
		case "extension_dependency_incompatible":
			expected, expectedErr := strconv.Atoi(finding.Expected)
			actual, actualErr := strconv.Atoi(finding.Actual)
			if expectedErr != nil || actualErr != nil {
				return invalidArtifact("dependency", nil)
			}
			diagnostic = c.registeredResultFinding("extension_dependency_incompatible", StartupFinding{Path: "$", ReasonCode: finding.Code, Message: "A required extension dependency is incompatible.", Details: map[string]any{"profile_id": finding.ProfileID, "dependency_profile_id": finding.DependencyProfileID, "expected_contract_major": expected, "actual_contract_major": actual}})
		case "extension_profile_unrecognized":
			diagnostic = registryArtifactFinding("configuration")
		default:
			return err
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	sortStartupFindings(diagnostics)
	return &StartupValidationError{Findings: diagnostics, cause: err}
}
