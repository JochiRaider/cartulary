package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

type extensionRegistryError struct{ findings []map[string]any }

func extensionInputLimitError(indexed map[string]map[string]any, err error) error {
	var limit *registryfacts.LimitError
	if errors.As(err, &limit) {
		return extensionRegistryDiagnostics(indexed, []registryfacts.OwnerFinding{limit.Finding()})
	}
	return err
}

func (e *extensionRegistryError) Error() string {
	encoded, err := canonicalizeExtension(map[string]any{"findings": objectsToAny(e.findings)})
	if err != nil {
		return "extension_registry_invalid"
	}
	return strings.TrimSuffix(encoded, "\n")
}

func validateExtensionRegistryFindings(indexed map[string]map[string]any) error {
	fragments := collectExtensionObjects(indexed, "cartulary.extension_owner_fragment.v3", "owner_fragment_id")
	bindings, err := objectArrayAllowEmpty(indexed["build/implementation-bindings.json"]["bindings"], "implementation declarations")
	if err != nil {
		return err
	}
	reservations, err := objectArray(indexed["build/base-route-reservations.json"]["reservations"], "Base reservations")
	if err != nil {
		return err
	}
	issues, err := registryfacts.RegistryFindings(registryfacts.RegistryInput{Fragments: fragments, ConfigurationContracts: collectExtensionObjects(indexed, "cartulary.extension_profile_configuration_contract.v4", "configuration_contract_id"), ImplementationBindings: bindings, BaseReservations: reservations})
	if err != nil {
		return err
	}
	return extensionRegistryDiagnostics(indexed, issues)
}

func extensionRegistryDiagnostics(indexed map[string]map[string]any, issues []registryfacts.OwnerFinding) error {
	registry, err := materializeValidationConditionRegistry(indexed["validation/surfaces.json"])
	if err != nil {
		return fmt.Errorf("owner diagnostic conditions unresolved")
	}
	conditions := map[string]map[string]any{}
	for _, raw := range registry["conditions"].([]any) {
		row := raw.(map[string]any)
		id := stringValue(row["condition_id"])
		if conditions[id] != nil {
			return fmt.Errorf("duplicate diagnostic condition")
		}
		conditions[id] = row
	}
	for _, id := range registryfacts.RegistryConditionIDs() {
		definition, _ := (registryfacts.OwnerFinding{ConditionID: id}).Definition()
		if !registryConditionMatches(conditions[id], definition) {
			return fmt.Errorf("registry diagnostic definition invalid")
		}
	}
	if len(issues) == 0 {
		return nil
	}
	rows, err := objectArray(indexed["specification/generated-schema-sources.json"]["schemas"], "schema sources")
	if err != nil {
		return err
	}
	schemas, err := compileExtensionSchemas(indexed, rows)
	if err != nil {
		return err
	}
	schema := schemas["cartulary.extension_startup_finding.v1"]
	if schema == nil {
		return fmt.Errorf("owner diagnostic schema unresolved")
	}
	findings := make([]map[string]any, 0, len(issues))
	for _, issue := range issues {
		definition, valid := issue.Definition()
		condition := conditions[issue.ConditionID]
		if !valid || !registryConditionMatches(condition, definition) {
			return fmt.Errorf("owner diagnostic definition invalid")
		}
		finding := map[string]any{"path": issue.Path, "reason_code": definition.ReasonCode, "message": definition.Message, "details": issue.Details}
		encoded, err := canonicalizeExtension(finding)
		if err != nil {
			return fmt.Errorf("registry diagnostic encoding invalid")
		}
		decoded, err := decodeContract([]byte(encoded))
		if err != nil || schema.Validate(decoded) != nil {
			return fmt.Errorf("registry diagnostic projection invalid")
		}
		findings = append(findings, finding)
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		for _, key := range []string{"path", "reason_code"} {
			if a[key] != b[key] {
				return stringValue(a[key]) < stringValue(b[key])
			}
		}
		left, _ := canonicalizeExtension(a["details"])
		right, _ := canonicalizeExtension(b["details"])
		return left < right
	})
	return &extensionRegistryError{findings: findings}
}

func registryConditionMatches(condition map[string]any, definition registryfacts.DiagnosticDefinition) bool {
	return condition["phase"] == definition.Phase && condition["condition_class"] == definition.ConditionClass && condition["path_algorithm_id"] == "extension_diagnostic_path_v1" && condition["reason_code"] == definition.ReasonCode && condition["expected_formatter_id"] == definition.ExpectedFormatter && condition["actual_formatter_id"] == definition.ActualFormatter && condition["multiplicity"] == definition.Multiplicity && condition["secret_policy"] == definition.SecretPolicy
}
