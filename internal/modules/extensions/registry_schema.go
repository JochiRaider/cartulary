package extensions

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/schemarefs"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"strings"
)

// The package is sealed before schema compilation. Resources come only from
// that immutable inventory; no URI may load a file or perform network I/O.
func validateRegistrySchemaBundle(p *admittedPackage) error {
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	schemas := map[string]*jsonschema.Schema{}
	for _, path := range p.paths {
		object := p.objects[path]
		id := stringValue(object["$id"])
		if strings.HasPrefix(id, "cartulary.extension_owner_fact_identity.") && id != registryfacts.IdentitySchemaID {
			return registryInvalid("owner_fact_identity_schema_invalid")
		}
		if retiredParticipantSchema(id) {
			return registryInvalid("retired_participant_schema")
		}
		if id == "" {
			continue
		}
		if err := compiler.AddResource("https://extensions.invalid/schemas/"+id, object); err != nil {
			return registryInvalid("schema_resource_invalid")
		}
	}
	for _, path := range p.paths {
		object := p.objects[path]
		id := stringValue(object["$id"])
		if id == "" {
			continue
		}
		schema, err := compiler.Compile("https://extensions.invalid/schemas/" + id)
		if err != nil {
			return registryInvalid("schema_definition_invalid")
		}
		schemas[id] = schema
	}
	if schemas[registryfacts.IdentitySchemaID] == nil {
		return registryInvalid("owner_fact_identity_definition_invalid")
	}
	schemaDefinitions := make([]map[string]any, 0, len(schemas))
	for _, object := range p.objects {
		if stringValue(object["$id"]) != "" {
			schemaDefinitions = append(schemaDefinitions, object)
		}
	}
	if registryfacts.CheckInputSchemaLimits(schemaDefinitions) != nil {
		return registryInvalid("owner_input_schema_limits_invalid")
	}
	for _, path := range p.paths {
		if err := checkRegistryInputLimits(p.objects[path]); err != nil {
			return err
		}
	}
	for _, path := range p.paths {
		object := p.objects[path]
		if stringValue(object["$id"]) != "" {
			continue
		}
		if retiredParticipantSchema(stringValue(object["schema_id"])) {
			return registryInvalid("retired_participant_contract")
		}
		schema := schemas[stringValue(object["schema_id"])]
		if schema == nil {
			return registryInvalid("artifact_schema_unresolved")
		}
		if err := schema.Validate(object); err != nil {
			return registryInvalid("artifact_schema_invalid")
		}
	}
	findingSchema := schemas["cartulary.extension_startup_finding.v1"]
	if findingSchema == nil {
		return registryInvalid("finding_schema_unresolved")
	}
	objects := make([]map[string]any, 0, len(p.objects))
	for _, object := range p.objects {
		objects = append(objects, object)
	}
	inventory := schemarefs.Collect(objects)
	p.requiredSchemas = inventory.Required
	if inventory.ValidateImports(p.objects[generatedExtensionsRoot+"dependency-snapshot.json"]) != nil {
		return registryInvalid("dependency_import_unresolved")
	}
	conditions, err := parseValidationConditions(p.objects[generatedExtensionsRoot+"validation-condition-registry.json"])
	if err != nil {
		return registryInvalid("validation_conditions_invalid")
	}
	if err := validateRegistryDiagnosticConditions(conditions); err != nil {
		return err
	}
	p.validationConditions = conditions
	p.validateFinding = findingSchema.Validate
	return validateParticipantStateFamilies(p)
}

// State-family ownership is a registry join. Implementation reference parity
// belongs to admitBinding so its failures retain the binding diagnostic.
func validateParticipantStateFamilies(p *admittedPackage) error {
	for _, path := range p.paths {
		object := p.objects[path]
		transaction := object["schema_id"] == "cartulary.extension_transaction_participant_contract.v3"
		if !transaction && object["schema_id"] != "cartulary.extension_reporting_source_contract.v1" && object["schema_id"] != "cartulary.extension_participant_specialization.v4" {
			continue
		}
		profile := stringValue(object["profile_id"])
		if transaction {
			profile = stringValue(object["owner_profile_id"])
		}
		authoritative := map[string]bool{}
		for _, candidate := range p.objects {
			if candidate["profile_id"] != profile || candidate["schema_id"] != "cartulary.extension_state_presence_manifest.v1" {
				continue
			}
			for _, key := range []string{"database_family_ids", "object_reference_family_ids"} {
				ids, ok := stringSlice(candidate[key])
				if !ok {
					return registryInvalid("presence_families_invalid")
				}
				for _, id := range ids {
					authoritative[id] = true
				}
			}
		}
		rows := []map[string]any{object}
		familyKey := "owned_state_family_ids"
		if object["schema_id"] == "cartulary.extension_reporting_source_contract.v1" {
			familyKey = "state_family_ids"
		} else if !transaction {
			var ok bool
			rows, ok = objectSlice(object["operations"])
			if !ok {
				return registryInvalid("participant_operations_invalid")
			}
			familyKey = "state_family_ids"
		}
		for _, row := range rows {
			families, ok := stringSlice(row[familyKey])
			if !ok || !strictlySortedUnique(families) {
				return registryInvalid("participant_families_invalid")
			}
			for _, family := range families {
				if !authoritative[family] {
					return registryInvalid("participant_family_unresolved")
				}
			}
		}
	}
	return nil
}

func validateParticipantBindingReferences(profile string, binding map[string]any, source *admittedPackage) error {
	rows, ok := objectSlice(binding["participant_contracts"])
	if !ok {
		return errors.New("participant bindings are malformed")
	}
	declared := map[string][]string{}
	for _, row := range rows {
		ids, ok := stringSlice(row["algorithm_ids"])
		if !ok || !strictlySortedUnique(ids) {
			return errors.New("participant algorithm bindings are malformed")
		}
		declared[stringValue(row["participant_id"])] = ids
	}
	for _, path := range source.paths {
		contract := source.objects[path]
		transaction := contract["schema_id"] == "cartulary.extension_transaction_participant_contract.v3"
		if !transaction && contract["schema_id"] != "cartulary.extension_reporting_source_contract.v1" && contract["schema_id"] != "cartulary.extension_participant_specialization.v4" {
			continue
		}
		owner := stringValue(contract["profile_id"])
		if transaction {
			owner = stringValue(contract["owner_profile_id"])
		}
		if owner != profile {
			continue
		}
		required := map[string]bool{}
		if transaction {
			for _, key := range []string{"prepare_algorithm_id", "validation_algorithm_id", "write_algorithm_id"} {
				required[stringValue(contract[key])] = true
			}
		} else if contract["schema_id"] == "cartulary.extension_reporting_source_contract.v1" {
			required[stringValue(contract["read_algorithm_id"])] = true
			required[stringValue(contract["ordering_algorithm_id"])] = true
		} else {
			operations, ok := objectSlice(contract["operations"])
			if !ok {
				return errors.New("participant operations are malformed")
			}
			for _, operation := range operations {
				for _, key := range []string{"algorithm_id", "ordering_algorithm_id"} {
					required[stringValue(operation[key])] = true
				}
			}
		}
		ids := declared[stringValue(contract["participant_id"])]
		if len(ids) != len(required) {
			return errors.New("participant algorithm binding is incomplete or extra")
		}
		for _, id := range ids {
			if !required[id] {
				return errors.New("participant algorithm binding differs from its contract")
			}
		}
	}
	return nil
}

func retiredParticipantSchema(id string) bool {
	switch id {
	case "cartulary.extension_participant_specialization.v3", "cartulary.extension_snapshot_reporting_participant_context.v1", "cartulary.extension_snapshot_reporting_participant_result.v1":
		return true
	case "cartulary.extension_transaction_participant_prepare_result.v1", "cartulary.extension_transaction_participant_validation_result.v1", "cartulary.extension_transaction_participant_write_result.v1", "cartulary.network_flow_activity.indicator_link_transaction_input.v1", "cartulary.network_flow_activity.import_apply_transaction_input.v1", "cartulary.incident_portability_import_transaction_input.v1", "cartulary.extension_shared_protocol_contract_set.v4":
		return true
	case "cartulary.extension_backup_participant_context.v1", "cartulary.extension_backup_participant_result.v1", "cartulary.extension_backup_restore_participant_context.v1", "cartulary.extension_backup_restore_participant_result.v1", "cartulary.extension_participant_contract.v1":
		return true
	}
	return false
}
