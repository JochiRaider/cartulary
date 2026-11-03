package extensions

import (
	"github.com/JochiRaider/cartulary/internal/modules/extensions/validation"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"sort"
)

// Byte-vector fixture for the precedence protocol. Actual typed owner callers
// are covered separately by the admission, state, and transaction regressions.
const (
	OwnerFindingSchemaLimit = 256
	OwnerFindingOverflowAt  = 4097
)

type OwnerFinding struct {
	Path       string
	ReasonCode string
	Message    string
	Details    map[string]any
}

type OwnerValidationDisposition string

const (
	OwnerValidationInvocationFailure OwnerValidationDisposition = "invocation_failure"
	OwnerValidationResultInvalid     OwnerValidationDisposition = "validation_result_invalid"
	OwnerValidationOverflow          OwnerValidationDisposition = "diagnostic_overflow"
	OwnerValidationFindings          OwnerValidationDisposition = "valid_findings"
	OwnerValidationSuccess           OwnerValidationDisposition = "valid_empty"
)

type OwnerValidationOutcome struct {
	Disposition OwnerValidationDisposition
	ReasonCode  string
	Findings    []OwnerFinding
	Limit       int
	Actual      int
}

// classifyProtocolByteVector supplies a fixture schema to the production kernel.
func classifyProtocolByteVector(invocationErr error, result []byte) OwnerValidationOutcome {
	raw, decodeErr := canonicaljson.DecodeStrict(result)
	object, objectOK := raw.(map[string]any)
	findings, arrayOK := object["findings"].([]any)
	normalized := make([]OwnerFinding, 0)
	disposition := validation.Classify(invocationErr, decodeErr == nil && objectOK && arrayOK, len(findings), OwnerFindingSchemaLimit, func() bool {
		if len(object) != 1 {
			return false
		}
		for _, raw := range findings {
			finding, ok := normalizeOwnerFinding(raw)
			if !ok {
				return false
			}
			normalized = append(normalized, finding)
		}
		return true
	})
	switch disposition {
	case validation.InvocationFailed:
		return OwnerValidationOutcome{Disposition: OwnerValidationInvocationFailure, ReasonCode: "extension_admission_validation_failed"}
	case validation.Overflow:
		return OwnerValidationOutcome{Disposition: OwnerValidationOverflow, ReasonCode: "extension_diagnostic_overflow", Limit: 4096, Actual: 4097}
	case validation.Success:
		return OwnerValidationOutcome{Disposition: OwnerValidationSuccess}
	case validation.Findings:
		sort.Slice(normalized, func(i, j int) bool {
			a, b := normalized[i], normalized[j]
			return compareStartupFinding(StartupFinding{Path: a.Path, ReasonCode: a.ReasonCode, Details: a.Details}, StartupFinding{Path: b.Path, ReasonCode: b.ReasonCode, Details: b.Details}) < 0
		})
		return OwnerValidationOutcome{Disposition: OwnerValidationFindings, Findings: normalized}
	default:
		return invalidOwnerValidationResult()
	}
}

func invalidOwnerValidationResult() OwnerValidationOutcome {
	return OwnerValidationOutcome{Disposition: OwnerValidationResultInvalid, ReasonCode: "extension_validation_result_invalid"}
}

func normalizeOwnerFinding(raw any) (OwnerFinding, bool) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 4 {
		return OwnerFinding{}, false
	}
	path, pathOK := object["path"].(string)
	reasonCode, reasonOK := object["reason_code"].(string)
	message, messageOK := object["message"].(string)
	details, detailsOK := object["details"].(map[string]any)
	if !pathOK || !reasonOK || !messageOK || !detailsOK || path == "" || path[0] != '$' || reasonCode == "" || message == "" {
		return OwnerFinding{}, false
	}
	return OwnerFinding{Path: path, ReasonCode: reasonCode, Message: message, Details: cloneObject(details)}, true
}
