package registryfacts

import (
	"fmt"
	"sort"
)

const (
	OwnerMissingCondition       = "extension_owner_required_fact_missing"
	OwnerUnknownCondition       = "extension_owner_fact_profile_unrecognized"
	OwnerOverflowCondition      = "extension_registry_diagnostic_overflow"
	DependencyOverflowCondition = "extension_dependency_diagnostic_overflow"
	RegistryLimitCondition      = "extension_registry_limit"
	DiagnosticLimit             = 4096
)

// OwnerFinding contains only admitted diagnostic identities and formatter
// outputs. Unknown profile values and their hashes never cross this boundary.
// The caller resolves the condition against its registered definition and schema.
type OwnerFinding struct {
	ConditionID string
	Path        string
	Details     map[string]any
}

func RequiredScalarKinds() []string {
	return []string{"recognized_profile", "claim_configuration", "state_ownership", "admission_validation", "egress", "portability", "snapshot_reporting"}
}

// FactOwner projects the owner allocation shared by admission and fact joins.
func FactOwner(kind, primary string) string {
	switch kind {
	case "recognized_profile", "runtime_dependency":
		return "core00"
	case "route_family", "workspace", "public_schema":
		return "core01"
	case "claim_configuration":
		return "core04"
	default:
		return primary
	}
}

// ownerInputFindings uses the normalized, structurally admitted registry. Only
// Core 00 recognition grants permission to expose a profile identity.
func ownerInputFindings(normalized []map[string]any, recognized map[string]bool) ([]OwnerFinding, error) {
	counts := map[string]map[string]int{}
	unknown := map[string]bool{}
	findings := []OwnerFinding{}
	overflow := func() ([]OwnerFinding, error) {
		return []OwnerFinding{{ConditionID: OwnerOverflowCondition, Path: "$", Details: map[string]any{"phase": "registry_generation", "limit": DiagnosticLimit, "actual": DiagnosticLimit + 1}}}, nil
	}
	for fi, fragment := range normalized {
		for i, raw := range fragment["facts"].([]any) {
			fact := raw.(map[string]any)
			identity, err := Identify(fact)
			if err != nil {
				return nil, err
			}
			if !recognized[identity.profile] {
				if unknown[identity.Key()] {
					continue
				}
				unknown[identity.Key()] = true
				findings = append(findings, OwnerFinding{ConditionID: OwnerUnknownCondition, Path: fmt.Sprintf("$.owner_fragments[%d].facts[%d].profile_id", fi, i), Details: map[string]any{"artifact_kind": "owner_fragment", "safe_ref": "cartulary:owner_fragment:" + fragment["owner_fragment_id"].(string), "expected": "exact:recognized_profile_id", "actual": "redacted"}})
				if len(findings) > DiagnosticLimit {
					return overflow()
				}
				continue
			}
			if counts[identity.profile] == nil {
				counts[identity.profile] = map[string]int{}
			}
			counts[identity.profile][identity.kind]++
		}
	}
	profiles := make([]string, 0, len(recognized))
	for profile := range recognized {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)
	for _, profile := range profiles {
		for _, kind := range RequiredScalarKinds() {
			if counts[profile][kind] != 0 {
				continue
			}
			findings = append(findings, OwnerFinding{ConditionID: OwnerMissingCondition, Path: "$.owner_fragments", Details: map[string]any{"artifact_kind": "owner_input_registry", "safe_ref": "cartulary:profile:" + profile, "expected": "exact:" + kind, "actual": "missing"}})
			if len(findings) > DiagnosticLimit {
				return overflow()
			}
		}
	}
	return findings, nil
}

// DiagnosticDefinition is the fixed owner-registry projection. Adapters must
// resolve and match its registered condition before emitting the finding.
type DiagnosticDefinition struct {
	ReasonCode, Message, Phase, ConditionClass                     string
	ExpectedFormatter, ActualFormatter, Multiplicity, SecretPolicy string
}

func (f OwnerFinding) Definition() (DiagnosticDefinition, bool) {
	d := DiagnosticDefinition{ReasonCode: "extension_registry_invalid", Message: "The extension contract artifact set is invalid.", Phase: "owner_input", ConditionClass: "missing_required_member", ExpectedFormatter: "diagnostic_exact_token_v1", ActualFormatter: "diagnostic_missing_v1", Multiplicity: "one_per_identity", SecretPolicy: "safe_value"}
	switch f.ConditionID {
	case OwnerMissingCondition:
	case OwnerUnknownCondition:
		d.ConditionClass = "invalid_scalar"
		d.ActualFormatter = "diagnostic_redacted_v1"
		d.SecretPolicy = "redacted"
	case OwnerOverflowCondition, DependencyOverflowCondition, RegistryLimitCondition:
		d.ReasonCode = "extension_diagnostic_overflow"
		d.Message = "Extension validation produced too many findings."
		d.Phase = "registry_generation"
		if f.ConditionID == DependencyOverflowCondition {
			d.Phase = "dependency_validation"
		}
		if f.ConditionID == RegistryLimitCondition {
			d.ReasonCode = "extension_registry_limit_exceeded"
			d.Message = "An extension registry limit was exceeded."
		}
		d.ConditionClass = "limit_exceeded"
		d.ExpectedFormatter = "diagnostic_count_v1"
		d.ActualFormatter = "diagnostic_count_v1"
		d.Multiplicity = "single"
	case CollisionCondition:
		d.ReasonCode = "extension_registry_conflict"
		d.Message = "The extension registry contains conflicting declarations."
		d.Phase = "collision_admission"
		d.ConditionClass = "collision"
		d.ActualFormatter = "diagnostic_redacted_v1"
		d.Multiplicity = "one_per_occurrence"
		d.SecretPolicy = "redacted"
	default:
		return DiagnosticDefinition{}, false
	}
	return d, true
}

func RegistryConditionIDs() []string {
	return []string{DependencyOverflowCondition, OwnerUnknownCondition, OwnerMissingCondition, CollisionCondition, OwnerOverflowCondition, RegistryLimitCondition}
}
