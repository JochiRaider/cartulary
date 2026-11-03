package extensions

import (
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func (p *admittedPackage) registryDiagnostics(fragments []map[string]any) error {
	input := registryfacts.RegistryInput{Fragments: fragments}
	for _, object := range p.objects {
		switch object["schema_id"] {
		case "cartulary.extension_profile_configuration_contract.v4":
			input.ConfigurationContracts = append(input.ConfigurationContracts, object)
		case "cartulary.extension_implementation_binding.v2":
			input.ImplementationBindings = append(input.ImplementationBindings, object)
		}
	}
	base, exists := p.objects["contracts/extensions/build/base-route-reservations.json"]
	if !exists || base["schema_id"] != "cartulary.base_route_reservation_registry.v3" {
		return registryInvalid("base_reservation_missing")
	}
	reservations, ok := objectSlice(base["reservations"])
	if !ok {
		return registryInvalid("base_reservation_invalid")
	}
	input.BaseReservations = reservations
	findings, err := registryfacts.RegistryFindings(input)
	if err != nil {
		return registryInvalid("owner_fact_identity_invalid")
	}
	if len(findings) == 0 {
		return nil
	}
	diagnostics := make([]StartupFinding, 0, len(findings))
	causes := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		definition, valid := finding.Definition()
		condition, ok := p.validationConditions[finding.ConditionID]
		if !valid || !ok || !registryConditionMatches(condition, definition) {
			return registryInvalid("owner_condition_definition_invalid")
		}
		diagnostic := StartupFinding{Path: finding.Path, ReasonCode: definition.ReasonCode, Message: definition.Message, Details: cloneObject(finding.Details)}
		body, err := canonicaljson.MarshalExtension(diagnostic)
		if err != nil {
			return registryInvalid("owner_diagnostic_invalid")
		}
		value, err := canonicaljson.DecodeStrict(body)
		if err != nil || p.validateFinding == nil || p.validateFinding(value) != nil {
			return registryInvalid("owner_diagnostic_invalid")
		}
		diagnostics = append(diagnostics, diagnostic)
		cause := Finding{Code: definition.ReasonCode, Phase: "registry_generation", Actual: "redacted"}
		if finding.ConditionID == registryfacts.CollisionCondition {
			cause.Phase = "registry_collision"
			cause.CollisionClass = stringValue(finding.Details["collision_class"])
			cause.ConflictingProfileIDs = append([]string{}, finding.Details["conflicting_profile_ids"].([]string)...)
			cause.ConflictingTokens = append([]string{}, finding.Details["conflicting_tokens"].([]string)...)
			cause.DuplicateCount, _ = finding.Details["duplicate_count"].(int)
		}
		causes = append(causes, cause)
	}
	sortStartupFindings(diagnostics)
	return &StartupValidationError{Findings: diagnostics, cause: &ValidationError{findings: causes}}
}

func registryConditionMatches(condition ValidationCondition, definition registryfacts.DiagnosticDefinition) bool {
	return condition.Phase == definition.Phase && condition.ConditionClass == definition.ConditionClass && condition.PathAlgorithmID == "extension_diagnostic_path_v1" && condition.ReasonCode == definition.ReasonCode && condition.ExpectedFormatter == definition.ExpectedFormatter && condition.ActualFormatter == definition.ActualFormatter && condition.Multiplicity == definition.Multiplicity && condition.SecretPolicy == definition.SecretPolicy
}

func validateRegistryDiagnosticConditions(conditions map[string]ValidationCondition) error {
	for _, id := range registryfacts.RegistryConditionIDs() {
		definition, _ := (registryfacts.OwnerFinding{ConditionID: id}).Definition()
		condition, ok := conditions[id]
		if !ok || !registryConditionMatches(condition, definition) {
			return registryInvalid("registry_condition_unresolved")
		}
	}
	return nil
}
