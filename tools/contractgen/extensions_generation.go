package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/schemarefs"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"sort"
	"strings"
)

const generatedExtensionPrefix = "contracts/extensions/generated/"

// deriveExtensionArtifacts materializes every runtime-facing identity from the
// catalogued owner inputs. It deliberately receives only the repository root:
// all inputs are re-opened through the exact authored input catalog, never by
// implementation-package discovery or source-text search.
func deriveExtensionArtifacts(root string) ([]artifact, error) {
	indexed, err := readExtensionInputCatalog(root)
	if err != nil {
		return nil, err
	}

	if err := registryfacts.ValidateIdentityDefinition(indexed["specification/contract-definitions.json"]); err != nil {
		return nil, err
	}
	if err := validateExtensionProfileFactClosure(indexed); err != nil {
		return nil, err
	}
	if err := validateExtensionImportReferences(indexed); err != nil {
		return nil, err
	}
	profiles, err := recognizedExtensionProfiles(indexed)
	if err != nil {
		return nil, err
	}

	declarations := indexed["dependencies.json"]
	dependencySnapshot := cloneObject(declarations)
	dependencySnapshot["schema_id"] = "cartulary.extension_dependency_snapshot.v3"

	fragments, err := registryfacts.NormalizeFragments(collectExtensionObjects(indexed, "cartulary.extension_owner_fragment.v3", "owner_fragment_id"))
	if err != nil {
		return nil, err
	}
	ownerInput := map[string]any{
		"schema_id":           "cartulary.extension_owner_input_registry.v2",
		"dependency_snapshot": dependencySnapshot,
		"owner_fragments":     objectsToAny(fragments),
	}
	if err := registryfacts.CheckCanonicalLimit(ownerInput, "owner_input_bytes", registryfacts.RegistryByteLimit); err != nil {
		return nil, extensionInputLimitError(indexed, err)
	}

	factsByProfile := extensionFactsByProfile(fragments)
	for profile, facts := range factsByProfile {
		for _, contract := range extensionJobKindContracts(facts) {
			if err := validateExtensionJobKindContract(contract, profile, "job contract"); err != nil {
				return nil, err
			}
		}
	}
	configsByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_profile_configuration_contract.v4")
	descriptors := make([]map[string]any, 0, len(profiles))
	descriptorDigest := map[string]string{}
	for _, profileID := range profiles {
		descriptor, err := materializeExtensionDescriptor(profileID, factsByProfile[profileID], configsByProfile[profileID])
		if err != nil {
			return nil, err
		}
		digest, err := extensionCanonicalDigest(descriptor)
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, descriptor)
		descriptorDigest[profileID] = digest
	}
	registry := map[string]any{
		"schema_id": "cartulary.extension_profile_registry.v1",
		"profiles":  objectsToAny(descriptors),
	}

	bindings, err := materializeExtensionBindings(indexed, descriptors, descriptorDigest, factsByProfile)
	if err != nil {
		return nil, err
	}
	stateRegistry, backupRegistry, participantRegistry, err := materializeExtensionRuntimeRegistries(indexed, descriptors, bindings)
	if err != nil {
		return nil, err
	}

	validationRegistry, err := materializeValidationConditionRegistry(indexed["validation/surfaces.json"])
	if err != nil {
		return nil, err
	}
	clientSupport, err := materializeClientSupport(indexed["build/client-support.json"], descriptors)
	if err != nil {
		return nil, err
	}

	var generated []artifact
	appendGenerated := func(path string, value any) error {
		current, err := extensionGeneratedArtifact(path, value)
		if err != nil {
			return err
		}
		generated = append(generated, current)
		return nil
	}
	if err := appendGenerated("dependency-snapshot.json", dependencySnapshot); err != nil {
		return nil, err
	}
	if err := appendGenerated("owner-input-registry.json", ownerInput); err != nil {
		return nil, err
	}
	for _, descriptor := range descriptors {
		if err := appendGenerated("descriptors/"+descriptor["profile_id"].(string)+".json", descriptor); err != nil {
			return nil, err
		}
	}
	if err := appendGenerated("profile-registry.json", registry); err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		if err := appendGenerated("implementation-bindings/"+binding["profile_id"].(string)+".json", binding); err != nil {
			return nil, err
		}
	}
	for _, profileID := range profiles {
		jobs := extensionJobKindContracts(factsByProfile[profileID])
		for _, contract := range jobs {
			if err := appendGenerated("job-contracts/"+profileID+"/"+stringValue(contract["job_kind"])+".json", contract); err != nil {
				return nil, err
			}
		}
		workers, err := extensionWorkerRuntimeContracts(factsByProfile[profileID], jobs)
		if err != nil {
			return nil, err
		}
		for _, runtimeContract := range workers {
			if err := appendGenerated("worker-runtime-contracts/"+profileID+"/"+stringValue(runtimeContract["worker_kind"])+".json", runtimeContract); err != nil {
				return nil, err
			}
		}
	}
	if err := appendGenerated("state-registry.json", stateRegistry); err != nil {
		return nil, err
	}
	if err := appendGenerated("backup-registry.json", backupRegistry); err != nil {
		return nil, err
	}
	if err := appendGenerated("participant-registry.json", participantRegistry); err != nil {
		return nil, err
	}
	if err := appendGenerated("validation-condition-registry.json", validationRegistry); err != nil {
		return nil, err
	}
	if err := appendGenerated("client-support-registry.json", clientSupport); err != nil {
		return nil, err
	}
	transactionLimits, err := extensionTransactionProtocol(indexed)
	if err != nil {
		return nil, err
	}
	schemaSources := cloneObject(indexed["specification/generated-schema-sources.json"])
	if err := projectExtensionTransactionProtocol(schemaSources, transactionLimits); err != nil {
		return nil, err
	}
	generatedSchemas, err := materializeExtensionSchemas(schemaSources)
	if err != nil {
		return nil, err
	}
	for _, schema := range generatedSchemas {
		if err := appendGenerated("schemas/"+stringValue(schema["$id"])+".schema.json", schema); err != nil {
			return nil, err
		}
	}

	integrity, err := materializeRegistryIntegrity(root, indexed, generated, descriptors, bindings, dependencySnapshot, ownerInput, registry)
	if err != nil {
		return nil, err
	}
	if err := appendGenerated("registry-integrity.json", integrity); err != nil {
		return nil, err
	}
	if err := validateExtensionArtifactSchemas(indexed, generated, generatedSchemas); err != nil {
		return nil, err
	}
	sort.Slice(generated, func(i, j int) bool { return generated[i].Path < generated[j].Path })
	return generated, nil
}

func readExtensionInputCatalog(root string) (map[string]map[string]any, error) {
	data, err := readExtensionSourceFile(root, "contracts/extensions/index.json")
	if err != nil {
		return nil, err
	}
	decoded, err := decodeContract(data)
	if err != nil {
		return nil, err
	}
	catalog, err := asObject(decoded, "contracts/extensions/index.json")
	if err != nil {
		return nil, err
	}
	rows, err := objectArray(catalog["artifacts"], "extension authored input catalog")
	if err != nil {
		return nil, err
	}
	indexed := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		path, _ := row["path"].(string)
		input, err := readExtensionSourceFile(root, "contracts/extensions/"+path)
		if err != nil {
			return nil, fmt.Errorf("read indexed extension input %s: %w", path, err)
		}
		value, err := decodeContract(input)
		if err != nil {
			return nil, fmt.Errorf("decode indexed extension input %s: %w", path, err)
		}
		object, err := asObject(value, path)
		if err != nil {
			return nil, err
		}
		indexed[path] = object
	}
	if err := admitExtensionRegistryDeclarations(indexed); err != nil {
		return nil, err
	}

	return indexed, nil
}

// admitExtensionRegistryDeclarations is shared by the complete producer and
// direct materialization. Logical diagnosis never precedes typed admission.
func admitExtensionRegistryDeclarations(indexed map[string]map[string]any) error {
	// Admit typed owner shapes before deriving identity or diagnostic positions.
	// Resources come exclusively from this authored catalog; no external loader.
	schemaRows, err := objectArray(indexed["specification/generated-schema-sources.json"]["schemas"], "schema sources")
	if err != nil {
		return err
	}
	if err := registryfacts.CheckInputSchemaLimits(schemaRows); err != nil {
		return err
	}
	schemas, err := compileExtensionSchemas(indexed, schemaRows)
	if err != nil {
		return err
	}
	for path, id := range map[string]string{"build/implementation-bindings.json": "cartulary.extension_implementation_binding_source_set.v1", "build/base-route-reservations.json": "cartulary.base_route_reservation_registry.v3"} {
		if indexed[path]["schema_id"] != id {
			return fmt.Errorf("registry declaration schema unresolved")
		}
	}
	fragments := []map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
			fragments = append(fragments, object)
		}
	}
	if err := registryfacts.CheckOwnerInputLimits(fragments); err != nil {
		return extensionInputLimitError(indexed, err)
	}
	for _, object := range indexed {
		id := stringValue(object["schema_id"])
		switch id {
		case "cartulary.extension_owner_fragment.v3", "cartulary.extension_profile_configuration_contract.v4", "cartulary.extension_implementation_binding_source_set.v1", "cartulary.base_route_reservation_registry.v3":
			schema := schemas[id]
			if schema == nil || schema.Validate(object) != nil {
				return fmt.Errorf("registry declaration schema invalid")
			}
		}
	}
	for _, fragment := range fragments {
		if err := normalizeExtensionArtifact(fragment); err != nil {
			return err
		}
	}
	return nil
}

func collectExtensionObjects(indexed map[string]map[string]any, schemaID, identity string) []map[string]any {
	objects := make([]map[string]any, 0)
	for _, object := range indexed {
		if object["schema_id"] == schemaID {
			objects = append(objects, cloneObject(object))
		}
	}
	sort.Slice(objects, func(i, j int) bool {
		left, _ := objects[i][identity].(string)
		right, _ := objects[j][identity].(string)
		return left < right
	})
	return objects
}

func extensionObjectsByProfile(indexed map[string]map[string]any, schemaID string) map[string]map[string]any {
	result := map[string]map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] != schemaID {
			continue
		}
		profileID, _ := object["profile_id"].(string)
		result[profileID] = object
	}
	return result
}

func objectsToAny(objects []map[string]any) []any {
	values := make([]any, len(objects))
	for index := range objects {
		values[index] = objects[index]
	}
	return values
}

func cloneObject(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = cloneExtensionValue(value)
	}
	return clone
}

func cloneExtensionValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneObject(typed)
	case []any:
		clone := make([]any, len(typed))
		for index := range typed {
			clone[index] = cloneExtensionValue(typed[index])
		}
		return clone
	default:
		return value
	}
}

func extensionFactsByProfile(fragments []map[string]any) map[string][]map[string]any {
	result := map[string][]map[string]any{}
	for _, fragment := range fragments {
		facts, _ := fragment["facts"].([]any)
		for _, rawFact := range facts {
			fact := rawFact.(map[string]any)
			profileID, _ := fact["profile_id"].(string)
			result[profileID] = append(result[profileID], fact)
		}
	}
	return result
}

func extensionMigrationFacts(indexed map[string]map[string]any, profileID string) []map[string]any {
	result := []map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
			continue
		}
		facts, _ := object["facts"].([]any)
		for _, rawFact := range facts {
			fact, _ := rawFact.(map[string]any)
			if fact["fact_kind"] == "migration_definition" && fact["profile_id"] == profileID {
				result = append(result, fact)
			}
		}
	}
	return result
}

func extensionMigrationLedgerFacts(indexed map[string]map[string]any, profileID string) []map[string]any {
	result := []map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
			continue
		}
		facts, _ := object["facts"].([]any)
		for _, rawFact := range facts {
			fact, _ := rawFact.(map[string]any)
			if fact["fact_kind"] == "migration_ledger_definition" && fact["profile_id"] == profileID {
				result = append(result, fact)
			}
		}
	}
	return result
}

func materializeExtensionDescriptor(profileID string, facts []map[string]any, config map[string]any) (map[string]any, error) {
	byKind := map[string][]map[string]any{}
	for _, fact := range facts {
		kind, _ := fact["fact_kind"].(string)
		byKind[kind] = append(byKind[kind], fact)
	}
	requireOne := func(kind string) (map[string]any, error) {
		if len(byKind[kind]) != 1 {
			return nil, fmt.Errorf("profile %s requires exactly one %s fact", profileID, kind)
		}
		return byKind[kind][0], nil
	}
	recognized, err := requireOne("recognized_profile")
	if err != nil {
		return nil, err
	}
	claim, err := requireOne("claim_configuration")
	if err != nil {
		return nil, err
	}
	state, err := requireOne("state_ownership")
	if err != nil {
		return nil, err
	}
	admission, err := requireOne("admission_validation")
	if err != nil {
		return nil, err
	}
	egress, err := requireOne("egress")
	if err != nil {
		return nil, err
	}
	portability, err := requireOne("portability")
	if err != nil {
		return nil, err
	}
	reporting, err := requireOne("snapshot_reporting")
	if err != nil {
		return nil, err
	}
	routes := extensionFactStrings(byKind["route_family"], "route_family")
	workspaces := extensionFactStrings(byKind["workspace"], "workspace_key")
	publicSchemas := extensionFactStrings(byKind["public_schema"], "public_schema_id")
	dependencies := make([]map[string]any, 0, len(byKind["runtime_dependency"]))
	for _, fact := range byKind["runtime_dependency"] {
		dependencies = append(dependencies, map[string]any{
			"profile_id":              fact["dependency_profile_id"],
			"required_contract_major": fact["required_contract_major"],
		})
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i]["profile_id"].(string) < dependencies[j]["profile_id"].(string)
	})
	contributions := make([]map[string]any, 0, len(byKind["contribution"]))
	for _, fact := range byKind["contribution"] {
		contribution, ok := fact["contribution"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("profile %s has malformed contribution fact", profileID)
		}
		contributions = append(contributions, cloneObject(contribution))
	}
	sort.Slice(contributions, func(i, j int) bool {
		left := stringValue(contributions[i]["contribution_id"]) + "\x00" + stringValue(contributions[i]["kind"])
		right := stringValue(contributions[j]["contribution_id"]) + "\x00" + stringValue(contributions[j]["kind"])
		return left < right
	})

	contributionRoutes := extensionContributionStrings(contributions, "http_route_family", "route_family")
	contributionWorkspaces := extensionContributionStrings(contributions, "incident_workspace", "workspace_key")
	if strings.Join(routes, "\x00") != strings.Join(contributionRoutes, "\x00") {
		return nil, fmt.Errorf("profile %s route-family facts do not equal contribution routes", profileID)
	}
	if strings.Join(workspaces, "\x00") != strings.Join(contributionWorkspaces, "\x00") {
		return nil, fmt.Errorf("profile %s workspace facts do not equal contribution workspaces", profileID)
	}
	prestage := []string{}
	if config == nil {
		return nil, fmt.Errorf("profile %s configuration contract is missing", profileID)
	}
	keys, _ := config["keys"].([]any)
	for _, rawKey := range keys {
		key := rawKey.(map[string]any)
		if key["inactive_policy"] == "syntax_only" {
			prestage = append(prestage, stringValue(key["key"]))
		}
	}
	sort.Strings(prestage)
	if claim["claim_config_key"] != profileID+".claimed" {
		return nil, fmt.Errorf("profile %s claim key is not canonical", profileID)
	}
	return map[string]any{
		"schema_id":                 "cartulary.extension_profile_descriptor.v3",
		"profile_id":                profileID,
		"claimable":                 recognized["claimable"],
		"contract_major":            recognized["contract_major"],
		"owner_id":                  recognized["primary_owner_id"],
		"claim_config_key":          claim["claim_config_key"],
		"route_families":            stringsToAny(routes),
		"workspace_keys":            stringsToAny(workspaces),
		"capability_ids":            []any{},
		"runtime_dependencies":      objectsToAny(dependencies),
		"contributions":             objectsToAny(contributions),
		"public_schema_ids":         stringsToAny(publicSchemas),
		"prestage_config_keys":      stringsToAny(prestage),
		"state_ownership":           state["state_ownership"],
		"admission_validation":      admission["admission_validation"],
		"egress_mode":               egress["egress_mode"],
		"incident_portability_mode": portability["incident_portability_mode"],
		"snapshot_reporting_mode":   reporting["snapshot_reporting_mode"],
	}, nil
}

func extensionFactStrings(facts []map[string]any, key string) []string {
	values := make([]string, 0, len(facts))
	for _, fact := range facts {
		values = append(values, stringValue(fact[key]))
	}
	sort.Strings(values)
	return values
}

func extensionContributionStrings(contributions []map[string]any, kind, key string) []string {
	values := []string{}
	for _, contribution := range contributions {
		if contribution["kind"] == kind {
			values = append(values, stringValue(contribution[key]))
		}
	}
	sort.Strings(values)
	return values
}

func extensionDeclaredMigrationBindings(indexed map[string]map[string]any, profileID string) ([]map[string]any, error) {
	migrationBindings := []map[string]any{}
	for _, fact := range extensionMigrationFacts(indexed, profileID) {
		definition, _ := fact["migration_definition"].(map[string]any)
		definitionDigest, digestErr := extensionCanonicalDigest(definition)
		if digestErr != nil {
			return nil, digestErr
		}
		migrationBindings = append(migrationBindings, map[string]any{
			"migration_id":                definition["migration_id"],
			"from_state_version":          definition["from_state_version"],
			"to_state_version":            definition["to_state_version"],
			"migration_definition_sha256": definitionDigest,
			"apply_algorithm_id":          definition["apply_algorithm_id"],
			"validation_algorithm_id":     definition["validation_algorithm_id"],
		})
	}
	sort.Slice(migrationBindings, func(i, j int) bool {
		leftFrom, _ := positiveJSONInt(migrationBindings[i]["from_state_version"], "migration from version")
		rightFrom, _ := positiveJSONInt(migrationBindings[j]["from_state_version"], "migration from version")
		if leftFrom != rightFrom {
			return leftFrom < rightFrom
		}
		leftTo, _ := positiveJSONInt(migrationBindings[i]["to_state_version"], "migration to version")
		rightTo, _ := positiveJSONInt(migrationBindings[j]["to_state_version"], "migration to version")
		if leftTo != rightTo {
			return leftTo < rightTo
		}
		return stringValue(migrationBindings[i]["migration_id"]) < stringValue(migrationBindings[j]["migration_id"])
	})
	return migrationBindings, nil
}

func materializeExtensionBindings(indexed map[string]map[string]any, descriptors []map[string]any, descriptorDigests map[string]string, factsByProfile map[string][]map[string]any) ([]map[string]any, error) {
	supportingSchemas := extensionSupportingSchemas(indexed, descriptors)
	transactionLimits, err := extensionTransactionProtocol(indexed)
	if err != nil {
		return nil, err
	}
	source := indexed["build/implementation-bindings.json"]
	if err := validateExtensionBindingSources(source); err != nil {
		return nil, err
	}
	rawSources, err := objectArrayAllowEmpty(source["bindings"], "implementation binding sources")
	if err != nil {
		return nil, err
	}
	sources := map[string]map[string]any{}
	recognized := map[string]bool{}
	for _, descriptor := range descriptors {
		recognized[stringValue(descriptor["profile_id"])] = true
	}
	for _, row := range rawSources {
		profileID := stringValue(row["profile_id"])
		if !recognized[profileID] {
			return nil, fmt.Errorf("implementation binding names unrecognized profile %s", profileID)
		}
		sources[profileID] = row
	}
	participantContracts := map[string]map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] == "cartulary.extension_reporting_source_contract.v1" || object["schema_id"] == "cartulary.extension_participant_specialization.v5" ||
			object["schema_id"] == "cartulary.extension_transaction_participant_contract.v3" {
			participantContracts[stringValue(object["participant_id"])] = object
		}
	}
	physicalByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_physical_state_binding.v1")
	initializationByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_state_initialization_definition.v1")
	bindings := make([]map[string]any, 0, len(descriptors))
	for _, descriptor := range descriptors {
		profileID := stringValue(descriptor["profile_id"])
		bindingSource := sources[profileID]
		if bindingSource == nil {
			continue
		}
		if bindingSource["contract_major"] != descriptor["contract_major"] || bindingSource["state_ownership_kind"] != descriptor["state_ownership"].(map[string]any)["kind"] {
			return nil, fmt.Errorf("profile %s implementation binding source disagrees with descriptor", profileID)
		}
		contributions := descriptor["contributions"].([]any)
		requiredContributionIDs := []string{}
		participantBindings := []map[string]any{}
		for _, rawContribution := range contributions {
			contribution := rawContribution.(map[string]any)
			kind := stringValue(contribution["kind"])
			if extensionContributionNeedsBinding(kind) {
				requiredContributionIDs = append(requiredContributionIDs, stringValue(contribution["contribution_id"]))
			}
			if strings.HasSuffix(kind, "_participant") {
				participantID := stringValue(contribution["participant_id"])
				contract := participantContracts[participantID]
				if contract == nil {
					return nil, fmt.Errorf("profile %s participant %s has no contract", profileID, participantID)
				}
				if err := validateParticipantContributionReference(profileID, contribution, contract); err != nil {
					return nil, err
				}
				digest, err := extensionCanonicalDigest(contract)
				if err != nil {
					return nil, err
				}
				if digest != contribution["participant_contract_sha256"] {
					return nil, fmt.Errorf("profile %s participant %s digest is stale", profileID, participantID)
				}
				algorithmIDs := []string{}
				if contract["schema_id"] == "cartulary.extension_participant_specialization.v5" {
					algorithmIDs = extensionSpecializationAlgorithmIDs(contract)
				}
				if contract["schema_id"] == "cartulary.extension_reporting_source_contract.v1" {
					algorithmIDs = []string{stringValue(contract["read_algorithm_id"]), stringValue(contract["ordering_algorithm_id"])}
					sort.Strings(algorithmIDs)
				}
				if contract["schema_id"] == "cartulary.extension_transaction_participant_contract.v3" {
					algorithmIDs = []string{
						stringValue(contract["prepare_algorithm_id"]),
						stringValue(contract["validation_algorithm_id"]),
						stringValue(contract["write_algorithm_id"]),
					}
					sort.Strings(algorithmIDs)
				}
				participantBindings = append(participantBindings, map[string]any{
					"participant_id":              participantID,
					"participant_contract_sha256": digest,
					"algorithm_ids":               stringsToAny(algorithmIDs),
				})
			}
		}
		sort.Strings(requiredContributionIDs)
		if !equalStringAny(requiredContributionIDs, bindingSource["implemented_contribution_ids"]) {
			return nil, fmt.Errorf("profile %s implementation contribution set is incomplete or extra", profileID)
		}
		sort.Slice(participantBindings, func(i, j int) bool {
			return participantBindings[i]["participant_id"].(string) < participantBindings[j]["participant_id"].(string)
		})
		rawParticipantSources, ok := bindingSource["participant_implementations"].([]any)
		if !ok || len(rawParticipantSources) != len(participantBindings) {
			return nil, fmt.Errorf("profile %s implementation participant set is incomplete or extra", profileID)
		}
		participantSources := make([]map[string]any, len(rawParticipantSources))
		for index, rawParticipantSource := range rawParticipantSources {
			participantSource, ok := rawParticipantSource.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("profile %s participant implementation source must be an object", profileID)
			}
			participantSources[index] = participantSource
		}
		for index, participantBinding := range participantBindings {
			participantSource := participantSources[index]
			if participantSource["participant_id"] != participantBinding["participant_id"] ||
				!equalStringAny(anyToStrings(participantBinding["algorithm_ids"]), participantSource["algorithm_ids"]) {
				return nil, fmt.Errorf("profile %s participant implementation binding is incomplete or extra", profileID)
			}
		}

		state := descriptor["state_ownership"].(map[string]any)
		var initializationDigest any
		var initializationAlgorithmID any
		var physicalDigest any
		var finalValidation any
		migrationBindings := []map[string]any{}
		backupCodecBindings := []map[string]any{}
		rebuildAlgorithms := []string{}
		if state["kind"] == "extension_versioned" {
			initialization := initializationByProfile[profileID]
			physical := physicalByProfile[profileID]
			if initialization == nil || physical == nil {
				return nil, fmt.Errorf("profile %s has incomplete versioned-state build inputs", profileID)
			}
			initializationDigest, err = extensionCanonicalDigest(initialization)
			if err != nil {
				return nil, err
			}
			if initializationDigest != state["initialization_definition_sha256"] {
				return nil, fmt.Errorf("profile %s initialization digest is stale", profileID)
			}
			initializationVariant, _ := initialization["initialization"].(map[string]any)
			if initializationVariant["kind"] == "algorithm" {
				initializationAlgorithmID = initializationVariant["algorithm_id"]
			} else {
				initializationAlgorithmID = nil
			}
			physicalDigest, err = extensionCanonicalDigest(physical)
			if err != nil {
				return nil, err
			}
			finalValidation = state["final_state_validation_algorithm_id"]
			migrationBindings, err = extensionDeclaredMigrationBindings(indexed, profileID)
			if err != nil {
				return nil, err
			}
			rows, _ := objectArray(physical["bindings"], "physical state bindings")
			for _, row := range rows {
				backupCodecBindings = append(backupCodecBindings, map[string]any{
					"binding_id":          row["binding_id"],
					"backup_codec_id":     row["backup_codec_id"],
					"backup_codec_sha256": row["backup_codec_sha256"],
				})
				if row["rebuild_algorithm_id"] != nil {
					rebuildAlgorithms = append(rebuildAlgorithms, stringValue(row["rebuild_algorithm_id"]))
				}
			}
		} else {
			initializationDigest = nil
			initializationAlgorithmID = nil
			physicalDigest = nil
			finalValidation = nil
		}
		implementedAlgorithms := map[string]struct{}{}
		for _, algorithmID := range anyToStrings(bindingSource["algorithm_ids"]) {
			implementedAlgorithms[algorithmID] = struct{}{}
		}
		for _, participant := range participantBindings {
			for _, algorithmID := range anyToStrings(participant["algorithm_ids"]) {
				if _, ok := implementedAlgorithms[algorithmID]; !ok {
					return nil, fmt.Errorf("profile %s implementation omits participant algorithm %s", profileID, algorithmID)
				}
			}
		}
		for _, requiredAlgorithm := range []any{initializationAlgorithmID, finalValidation, descriptor["admission_validation"].(map[string]any)["preflight_algorithmid"], descriptor["admission_validation"].(map[string]any)["post_migration_algorithmid"]} {
			if requiredAlgorithm == nil {
				continue
			}
			if _, ok := implementedAlgorithms[stringValue(requiredAlgorithm)]; !ok {
				return nil, fmt.Errorf("profile %s implementation omits state algorithm %s", profileID, requiredAlgorithm)
			}
		}
		for _, migration := range migrationBindings {
			for _, key := range []string{"apply_algorithm_id", "validation_algorithm_id"} {
				algorithmID := stringValue(migration[key])
				if _, ok := implementedAlgorithms[algorithmID]; !ok {
					return nil, fmt.Errorf("profile %s implementation omits migration algorithm %s", profileID, algorithmID)
				}
			}
		}
		admission := descriptor["admission_validation"].(map[string]any)
		dependencyProbeIDs := []string{}
		for _, rawProbe := range admission["dependency_probes"].([]any) {
			dependencyProbeIDs = append(dependencyProbeIDs, stringValue(rawProbe.(map[string]any)["probe_id"]))
		}
		workerKinds := extensionFactStringsByKind(factsByProfile[profileID], "worker_kind", "worker_kind")
		jobContracts := extensionJobKindContracts(factsByProfile[profileID])
		workerRuntimeContracts, err := extensionWorkerRuntimeContracts(factsByProfile[profileID], jobContracts)
		if err != nil {
			return nil, err
		}
		if len(jobContracts) > 64 {
			return nil, fmt.Errorf("profile %s exceeds the job binding limit", profileID)
		}
		jobReferences := make([]map[string]any, 0, len(jobContracts))
		jobKinds := make([]string, len(jobContracts))
		for index, contract := range jobContracts {
			jobKinds[index] = stringValue(contract["job_kind"])
			digest, err := extensionCanonicalDigest(contract)
			if err != nil {
				return nil, err
			}
			jobReferences = append(jobReferences, map[string]any{"job_kind": contract["job_kind"], "job_kind_contract_sha256": digest})
		}
		if !equalStringAny(workerKinds, bindingSource["implemented_worker_kinds"]) {
			return nil, fmt.Errorf("profile %s implementation worker set is incomplete or extra", profileID)
		}
		if !equalStringAny(jobKinds, bindingSource["implemented_job_kinds"]) {
			return nil, fmt.Errorf("profile %s implementation job set is incomplete or extra", profileID)
		}
		bindings = append(bindings, map[string]any{
			"schema_id":                           "cartulary.extension_implementation_binding.v2",
			"profile_id":                          profileID,
			"contract_major":                      descriptor["contract_major"],
			"descriptor_sha256":                   descriptorDigests[profileID],
			"implemented_contribution_ids":        stringsToAny(requiredContributionIDs),
			"supported_capability_ids":            []any{},
			"state_ownership_kind":                state["kind"],
			"preflight_algorithm_id":              admission["preflight_algorithmid"],
			"post_migration_algorithm_id":         admission["post_migration_algorithmid"],
			"initialization_definition_sha256":    initializationDigest,
			"initialization_algorithm_id":         initializationAlgorithmID,
			"final_state_validation_algorithm_id": finalValidation,
			"dependency_probe_ids":                stringsToAny(dependencyProbeIDs),
			"migration_definitions":               objectsToAny(migrationBindings),
			"physical_state_binding_sha256":       physicalDigest,
			"backup_codec_bindings":               objectsToAny(backupCodecBindings),
			"rebuild_algorithm_ids":               stringsToAny(rebuildAlgorithms),
			"transaction_participant_limits":      extensionBindingTransactionLimits(transactionLimits),
			"supporting_schema_ids":               stringsToAny(supportingSchemas[profileID]),
			"worker_kinds":                        stringsToAny(workerKinds),
			"job_kind_contracts":                  objectsToAny(jobReferences),
			"participant_contracts":               objectsToAny(participantBindings),
			"worker_runtime_contracts":            objectsToAny(workerRuntimeContracts),
		})
	}
	return bindings, nil
}

func extensionFactStringsByKind(facts []map[string]any, factKind, key string) []string {
	values := []string{}
	for _, fact := range facts {
		if fact["fact_kind"] == factKind {
			values = append(values, stringValue(fact[key]))
		}
	}
	sort.Strings(values)
	return values
}

func extensionJobKindContracts(facts []map[string]any) []map[string]any {
	contracts := []map[string]any{}
	for _, fact := range facts {
		if fact["fact_kind"] != "job_kind" {
			continue
		}
		if contract, ok := fact["job_kind_contract"].(map[string]any); ok {
			contracts = append(contracts, cloneObject(contract))
		}
	}
	sortObjectRows(contracts, "job_kind")
	return contracts
}

func extensionWorkerRuntimeContracts(facts []map[string]any, jobContracts []map[string]any) ([]map[string]any, error) {
	declaredJobs := make(map[string]struct{}, len(jobContracts))
	for _, contract := range jobContracts {
		declaredJobs[stringValue(contract["job_kind"])] = struct{}{}
	}
	assignedJobs := make(map[string]string, len(jobContracts))
	runtimeContracts := []map[string]any{}
	for _, fact := range facts {
		if fact["fact_kind"] != "worker_kind" {
			continue
		}
		workerKind := stringValue(fact["worker_kind"])
		jobKinds, err := sortedUniqueStringArray(fact["job_kinds"], workerKind+".job_kinds", true)
		if err != nil {
			return nil, err
		}
		for _, jobKind := range jobKinds {
			if _, ok := declaredJobs[jobKind]; !ok {
				return nil, fmt.Errorf("worker %s maps undeclared job kind %s", workerKind, jobKind)
			}
			if prior, duplicate := assignedJobs[jobKind]; duplicate {
				return nil, fmt.Errorf("job kind %s is mapped to both %s and %s", jobKind, prior, workerKind)
			}
			assignedJobs[jobKind] = workerKind
		}
		runtimeContracts = append(runtimeContracts, map[string]any{
			"schema_id":                       "cartulary.extension_worker_runtime_contract.v1",
			"profile_id":                      fact["profile_id"],
			"worker_kind":                     workerKind,
			"job_kinds":                       stringsToAny(jobKinds),
			"max_active_attempts_per_process": fact["max_active_attempts_per_process"],
		})
	}
	if len(assignedJobs) != len(declaredJobs) {
		missing := []string{}
		for jobKind := range declaredJobs {
			if _, ok := assignedJobs[jobKind]; !ok {
				missing = append(missing, jobKind)
			}
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("worker runtime contracts omit job kinds %s", strings.Join(missing, ", "))
	}
	sortObjectRows(runtimeContracts, "worker_kind")
	return runtimeContracts, nil
}

func extensionSpecializationAlgorithmIDs(contract map[string]any) []string {
	set := map[string]struct{}{}
	operations, _ := objectArray(contract["operations"], "participant specialization operations")
	for _, operation := range operations {
		for _, key := range []string{"algorithm_id", "ordering_algorithm_id"} {
			if algorithmID := stringValue(operation[key]); algorithmID != "" {
				set[algorithmID] = struct{}{}
			}
		}
	}
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func extensionContributionNeedsBinding(kind string) bool {
	switch kind {
	case "http_route_family", "incident_workspace", "deployment_admin_panel", "authentication_entry", "import_target", "websocket_invalidation", "cross_owner_transaction_participant", "snapshot_reporting_participant", "incident_portability_participant":
		return true
	default:
		return false
	}
}

func materializeExtensionRuntimeRegistries(indexed map[string]map[string]any, descriptors, bindings []map[string]any) (map[string]any, map[string]any, map[string]any, error) {
	presenceByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_state_presence_manifest.v1")
	initializationByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_state_initialization_definition.v1")
	physicalByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_physical_state_binding.v1")
	bindingByProfile := map[string]map[string]any{}
	for _, binding := range bindings {
		bindingByProfile[stringValue(binding["profile_id"])] = binding
	}
	stateRows := []map[string]any{}
	backupRows := []map[string]any{}
	for _, descriptor := range descriptors {
		profileID := stringValue(descriptor["profile_id"])
		state := descriptor["state_ownership"].(map[string]any)
		if state["kind"] != "extension_versioned" {
			continue
		}
		presence := presenceByProfile[profileID]
		initialization := initializationByProfile[profileID]
		physical := physicalByProfile[profileID]
		binding := bindingByProfile[profileID]
		if presence == nil || initialization == nil || physical == nil {
			return nil, nil, nil, fmt.Errorf("profile %s runtime state registry input is incomplete", profileID)
		}
		presenceDigest, err := extensionCanonicalDigest(presence)
		if err != nil {
			return nil, nil, nil, err
		}
		initializationDigest, err := extensionCanonicalDigest(initialization)
		if err != nil {
			return nil, nil, nil, err
		}
		physicalDigest, err := extensionCanonicalDigest(physical)
		if err != nil {
			return nil, nil, nil, err
		}
		if presence["profile_id"] != profileID || !strings.HasPrefix(stringValue(presence["state_presence_manifest_id"]), profileID+".") || presence["migration_lineage_id"] != state["migration_lineage_id"] || presence["empty_state_policy"] != state["empty_state_policy"] || initializationDigest != state["initialization_definition_sha256"] || binding != nil && physicalDigest != binding["physical_state_binding_sha256"] {
			return nil, nil, nil, fmt.Errorf("profile %s runtime state registry disagrees with admitted descriptor/binding", profileID)
		}
		initializationVariant, _ := initialization["initialization"].(map[string]any)
		var bindingDigest any
		if binding != nil {
			bindingDigest, err = extensionCanonicalDigest(binding)
			if err != nil {
				return nil, nil, nil, err
			}
		}
		runtimeMigrations := []map[string]any{}
		ledgerDefinitions := []map[string]any{}
		rawMigrations, err := extensionDeclaredMigrationBindings(indexed, profileID)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, rawMigration := range rawMigrations {
			migration := cloneObject(rawMigration)
			migration["migration_lineage_id"] = state["migration_lineage_id"]
			migration["implementation_binding_profile_id"] = profileID
			migration["implementation_binding_sha256"] = bindingDigest
			runtimeMigrations = append(runtimeMigrations, migration)
			ledgerDefinitions = append(ledgerDefinitions, map[string]any{
				"schema_id":                   "cartulary.extension_migration_ledger_definition.v1",
				"profile_id":                  profileID,
				"migration_lineage_id":        state["migration_lineage_id"],
				"migration_id":                migration["migration_id"],
				"from_state_version":          migration["from_state_version"],
				"to_state_version":            migration["to_state_version"],
				"migration_definition_sha256": migration["migration_definition_sha256"],
			})
		}
		for _, fact := range extensionMigrationLedgerFacts(indexed, profileID) {
			definition, _ := fact["migration_ledger_definition"].(map[string]any)
			ledgerDefinitions = append(ledgerDefinitions, cloneObject(definition))
		}
		sort.Slice(ledgerDefinitions, func(i, j int) bool {
			leftFrom, _ := positiveJSONInt(ledgerDefinitions[i]["from_state_version"], "ledger definition from version")
			rightFrom, _ := positiveJSONInt(ledgerDefinitions[j]["from_state_version"], "ledger definition from version")
			if leftFrom != rightFrom {
				return leftFrom < rightFrom
			}
			return stringValue(ledgerDefinitions[i]["migration_id"]) < stringValue(ledgerDefinitions[j]["migration_id"])
		})
		stateRows = append(stateRows, map[string]any{
			"profile_id":                                 profileID,
			"contract_major":                             descriptor["contract_major"],
			"migration_lineage_id":                       state["migration_lineage_id"],
			"current_state_version":                      state["current_state_version"],
			"minimum_migratable_state_version":           state["minimum_migratable_state_version"],
			"empty_state_policy":                         state["empty_state_policy"],
			"database_family_ids":                        presence["database_family_ids"],
			"object_reference_family_ids":                presence["object_reference_family_ids"],
			"state_presence_manifest_sha256":             presenceDigest,
			"initialization_kind":                        initializationVariant["kind"],
			"initialization_definition_sha256":           initializationDigest,
			"initialization_algorithm_id":                initializationVariant["algorithm_id"],
			"initialization_algorithm_definition_sha256": initializationVariant["algorithm_definition_sha256"],
			"migration_definitions":                      objectsToAny(runtimeMigrations),
			"migration_ledger_definitions":               objectsToAny(ledgerDefinitions),
			"final_state_validation_algorithm_id":        state["final_state_validation_algorithm_id"],
			"physical_state_binding_sha256":              physicalDigest,
			"implementation_binding_sha256":              bindingDigest,
		})

		codecRows := []map[string]any{}
		physicalBindings, _ := objectArray(physical["bindings"], profileID+" physical bindings")
		for _, physicalBinding := range physicalBindings {
			codecID := stringValue(physicalBinding["backup_codec_id"])
			var codec map[string]any
			for _, candidate := range indexed {
				if candidate["schema_id"] == "cartulary.extension_backup_binding_codec.v4" && candidate["backup_codec_id"] == codecID {
					codec = candidate
					break
				}
			}
			if codec == nil {
				return nil, nil, nil, fmt.Errorf("profile %s backup codec %s is missing", profileID, codecID)
			}
			codecDigest, err := extensionCanonicalDigest(codec)
			if err != nil {
				return nil, nil, nil, err
			}
			if codecDigest != physicalBinding["backup_codec_sha256"] {
				return nil, nil, nil, fmt.Errorf("profile %s backup codec %s digest is stale", profileID, codecID)
			}
			codecRows = append(codecRows, map[string]any{
				"backup_codec_id":     codecID,
				"backup_codec_sha256": codecDigest,
				"codec":               cloneObject(codec),
			})
		}
		sortObjectRows(codecRows, "backup_codec_id")
		logicalContracts, postgresMappings, err := extensionLogicalBackupArtifacts(profileID, physicalBindings, indexed)
		if err != nil {
			return nil, nil, nil, err
		}
		backupRows = append(backupRows, map[string]any{
			"profile_id":                    profileID,
			"physical_state_binding_sha256": physicalDigest,
			"physical_state_binding":        cloneObject(physical),
			"codecs":                        objectsToAny(codecRows),
			"logical_contracts":             objectsToAny(logicalContracts),
			"postgres_mappings":             objectsToAny(postgresMappings),
		})
	}
	sortObjectRows(stateRows, "profile_id")
	sortObjectRows(backupRows, "profile_id")

	participantRows := []map[string]any{}
	for _, contract := range indexed {
		schemaID := stringValue(contract["schema_id"])
		if schemaID != "cartulary.extension_reporting_source_contract.v1" && schemaID != "cartulary.extension_participant_specialization.v5" &&
			schemaID != "cartulary.extension_transaction_participant_contract.v3" {
			continue
		}
		digest, err := extensionCanonicalDigest(contract)
		if err != nil {
			return nil, nil, nil, err
		}
		profileID := stringValue(contract["profile_id"])
		if schemaID == "cartulary.extension_transaction_participant_contract.v3" {
			profileID = stringValue(contract["owner_profile_id"])
		}
		participantRows = append(participantRows, map[string]any{
			"participant_id":              contract["participant_id"],
			"profile_id":                  profileID,
			"contract_schema_id":          schemaID,
			"participant_contract_sha256": digest,
			"contract":                    cloneObject(contract),
		})
	}
	sortObjectRows(participantRows, "participant_id")
	return map[string]any{
		"schema_id": "cartulary.extension_state_registry.v1",
		"profiles":  objectsToAny(stateRows),
	}, map[string]any{
		"schema_id": "cartulary.extension_backup_registry.v1",
		"profiles":  objectsToAny(backupRows),
	}, map[string]any{
		"schema_id":    "cartulary.extension_participant_registry.v1",
		"participants": objectsToAny(participantRows),
	}, nil
}

func extensionSupportingSchemas(indexed map[string]map[string]any, descriptors []map[string]any) map[string][]string {
	objects := make([]map[string]any, 0, len(indexed)+len(descriptors))
	for _, object := range indexed {
		objects = append(objects, object)
	}
	objects = append(objects, descriptors...)
	return schemarefs.Collect(objects).Required
}

func materializeValidationConditionRegistry(source map[string]any) (map[string]any, error) {
	declarations, err := objectArray(source["declarations"], "validation declarations")
	if err != nil {
		return nil, err
	}
	conditions := []map[string]any{}
	for _, declaration := range declarations {
		for _, surfaceFamily := range []string{"schema_surfaces", "procedural_surfaces"} {
			surfaces, _ := declaration[surfaceFamily].([]any)
			for _, rawSurface := range surfaces {
				surface := rawSurface.(map[string]any)
				rawConditions, _ := surface["conditions"].([]any)
				for _, rawCondition := range rawConditions {
					condition := cloneObject(rawCondition.(map[string]any))
					delete(condition, "location")
					delete(condition, "constraint")
					conditions = append(conditions, condition)
				}
			}
		}
	}
	sort.Slice(conditions, func(i, j int) bool {
		if conditions[i]["phase"] != conditions[j]["phase"] {
			return stringValue(conditions[i]["phase"]) < stringValue(conditions[j]["phase"])
		}
		return conditions[i]["condition_id"].(string) < conditions[j]["condition_id"].(string)
	})
	return map[string]any{
		"schema_id":  "cartulary.extension_validation_condition_registry.v2",
		"conditions": objectsToAny(conditions),
	}, nil
}

func materializeClientSupport(source map[string]any, descriptors []map[string]any) (map[string]any, error) {
	rows, err := objectArrayAllowEmpty(source["rows"], "client support source rows")
	if err != nil {
		return nil, err
	}
	descriptorByProfile := map[string]map[string]any{}
	for _, descriptor := range descriptors {
		descriptorByProfile[stringValue(descriptor["profile_id"])] = descriptor
	}
	profiles := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		profileID := stringValue(row["profile_id"])
		descriptor := descriptorByProfile[profileID]
		if descriptor == nil || row["contract_major"] != descriptor["contract_major"] || !equalStringAny(anyToStrings(row["workspace_keys"]), descriptor["workspace_keys"]) {
			return nil, fmt.Errorf("client support source for %s disagrees with descriptor", profileID)
		}
		if len(anyToStrings(row["capability_ids"])) != 0 {
			return nil, fmt.Errorf("client support source for %s advertises a prohibited capability", profileID)
		}
		if !equalStringAny(anyToStrings(row["public_schema_ids"]), descriptor["public_schema_ids"]) {
			return nil, fmt.Errorf("client support source for %s disagrees with descriptor public schemas", profileID)
		}
		profiles = append(profiles, map[string]any{
			"profile_id":                profileID,
			"supported_contract_majors": []any{row["contract_major"]},
			"workspace_keys":            row["workspace_keys"],
			"capability_ids":            []any{},
			"public_schema_ids":         row["public_schema_ids"],
		})
	}
	// The asset-set digest is finalized by the web build in ES-12. Binding the
	// exact source object here prevents an untracked semantic support change in
	// the generated contract family without pretending source files are served
	// assets.
	sourceDigest, err := extensionCanonicalDigest(source)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"schema_id":          "cartulary.client_extension_support_registry.v1",
		"client_build_id":    "cartulary.web.standard.source.v1",
		"client_build_class": source["client_build_class"],
		"asset_set_sha256":   sourceDigest,
		"profiles":           objectsToAny(profiles),
	}, nil
}

func materializeExtensionSchemas(source map[string]any) ([]map[string]any, error) {
	if err := requireAllowedKeys(source, stringSet("schema_id", "schemas"), "schema sources"); err != nil {
		return nil, err
	}
	if source["schema_id"] != "cartulary.extension_generated_schema_source_set.v4" {
		return nil, fmt.Errorf("unsupported schema source format")
	}
	rows, err := objectArray(source["schemas"], "generated schema sources")
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	currentIdentity := false
	previous := ""
	for _, row := range rows {
		id := stringValue(row["$id"])
		if strings.HasPrefix(id, "cartulary.extension_owner_fact_identity.") {
			if id != registryfacts.IdentitySchemaID {
				return nil, registryfacts.ErrInvalidFact
			}
			currentIdentity = true
		}
		if retiredExtensionParticipantSchema(id) {
			return nil, fmt.Errorf("retired participant schema")
		}
		if id == "" || id <= previous || row["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
			return nil, fmt.Errorf("schema sources must use sorted unique current identities")
		}
		previous = id
		if err := compiler.AddResource("https://extensions.invalid/schemas/"+id, row); err != nil {
			return nil, fmt.Errorf("invalid schema resource %s", id)
		}
	}
	if !currentIdentity {
		return nil, registryfacts.ErrInvalidFact
	}
	for _, row := range rows {
		if _, err := compiler.Compile("https://extensions.invalid/schemas/" + stringValue(row["$id"])); err != nil {
			return nil, fmt.Errorf("invalid or unresolved schema %s", row["$id"])
		}
	}
	return rows, nil
}

func compileExtensionSchemas(indexed map[string]map[string]any, schemas []map[string]any) (map[string]*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	resources := append([]map[string]any(nil), schemas...)
	for _, object := range indexed {
		if stringValue(object["$id"]) != "" {
			resources = append(resources, object)
		}
	}
	for _, resource := range resources {
		if err := compiler.AddResource("https://extensions.invalid/schemas/"+stringValue(resource["$id"]), resource); err != nil {
			return nil, fmt.Errorf("invalid schema resource")
		}
	}
	compiled := map[string]*jsonschema.Schema{}
	for _, resource := range resources {
		id := stringValue(resource["$id"])
		if compiled[id] != nil {
			return nil, fmt.Errorf("duplicate schema resource %s", id)
		}
		schema, err := compiler.Compile("https://extensions.invalid/schemas/" + id)
		if err != nil {
			return nil, fmt.Errorf("unresolved schema resource %s", id)
		}
		compiled[id] = schema
	}
	return compiled, nil
}

func validateExtensionArtifactSchemas(indexed map[string]map[string]any, generated []artifact, schemas []map[string]any) error {
	compiled, err := compileExtensionSchemas(indexed, schemas)
	if err != nil {
		return err
	}
	validate := func(path string, object map[string]any) error {
		if stringValue(object["$id"]) != "" {
			return nil
		}
		id := stringValue(object["schema_id"])
		schema := compiled[id]
		if schema == nil {
			return fmt.Errorf("artifact %s has unresolved schema %s", path, id)
		}
		if err := schema.Validate(object); err != nil {
			return fmt.Errorf("artifact %s does not satisfy schema %s", path, id)
		}
		return nil
	}
	paths := make([]string, 0, len(indexed))
	for path := range indexed {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := validate(path, indexed[path]); err != nil {
			return err
		}
	}
	for _, artifact := range generated {
		decoded, err := decodeContract([]byte(artifact.JSON))
		if err != nil {
			return fmt.Errorf("generated artifact is not canonical JSON")
		}
		if err := validate(artifact.Path, decoded.(map[string]any)); err != nil {
			return err
		}
	}
	return nil
}

func materializeRegistryIntegrity(root string, indexed map[string]map[string]any, generated []artifact, descriptors, bindings []map[string]any, dependencySnapshot, ownerInput, registry map[string]any) (map[string]any, error) {
	dependencyDigest, _ := extensionCanonicalDigest(dependencySnapshot)
	ownerInputDigest, _ := extensionCanonicalDigest(ownerInput)
	registryDigest, _ := extensionCanonicalDigest(registry)
	descriptorRows := []map[string]any{}
	for _, descriptor := range descriptors {
		digest, _ := extensionCanonicalDigest(descriptor)
		descriptorRows = append(descriptorRows, map[string]any{"profile_id": descriptor["profile_id"], "descriptor_sha256": digest})
	}
	bindingRows := []map[string]any{}
	for _, binding := range bindings {
		digest, _ := extensionCanonicalDigest(binding)
		bindingRows = append(bindingRows, map[string]any{"profile_id": binding["profile_id"], "binding_sha256": digest})
	}
	fragmentRows := []map[string]any{}
	supportRows := []map[string]any{}
	generatedSchemaRows := []map[string]any{}
	for path, object := range indexed {
		digest, err := extensionCanonicalDigest(object)
		if err != nil {
			return nil, err
		}
		switch object["schema_id"] {
		case "cartulary.extension_owner_fragment.v3":
			fragmentRows = append(fragmentRows, map[string]any{"owner_fragment_id": object["owner_fragment_id"], "owner_fragment_sha256": digest})
		default:
			if !extensionRuntimeArtifactPath("contracts/extensions/" + path) {
				continue
			}
			supportRows = append(supportRows, map[string]any{
				"artifact_id":     strings.TrimSuffix(strings.ReplaceAll(path, "/", "."), ".json"),
				"schema_id":       object["schema_id"],
				"artifact_sha256": digest,
			})
		}
	}
	for _, current := range generated {
		generatedRelative := strings.TrimPrefix(current.Path, generatedExtensionPrefix)
		if strings.HasPrefix(generatedRelative, "schemas/") {
			generatedSchemaRows = append(generatedSchemaRows, map[string]any{
				"schema_id":     generatedSchemaID(current.JSON),
				"schema_sha256": current.SHA256,
			})
			continue
		}
		if generatedRelative == "dependency-snapshot.json" ||
			generatedRelative == "owner-input-registry.json" ||
			generatedRelative == "profile-registry.json" ||
			strings.HasPrefix(generatedRelative, "descriptors/") ||
			strings.HasPrefix(generatedRelative, "implementation-bindings/") {
			continue
		}
		supportRows = append(supportRows, map[string]any{
			"artifact_id":     strings.TrimSuffix(generatedRelative, ".json"),
			"schema_id":       generatedSchemaID(current.JSON),
			"artifact_sha256": current.SHA256,
		})
	}
	sortObjectRows(fragmentRows, "owner_fragment_id")
	sortObjectRows(supportRows, "artifact_id")
	sortObjectRows(generatedSchemaRows, "schema_id")
	generatorSources, err := extensionGeneratorSources(root, indexed)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"schema_id":                            "cartulary.extension_registry_integrity.v2",
		"canonicalization_algorithm_id":        "extension_registry_canonical_json_v1",
		"dependency_snapshot_sha256":           dependencyDigest,
		"owner_input_registry_sha256":          ownerInputDigest,
		"registry_schema_id":                   "cartulary.extension_profile_registry.v1",
		"registry_sha256":                      registryDigest,
		"descriptor_digests":                   objectsToAny(descriptorRows),
		"owner_fragment_digests":               objectsToAny(fragmentRows),
		"implementation_binding_digests":       objectsToAny(bindingRows),
		"supporting_contract_artifact_digests": objectsToAny(supportRows),
		"generator_id":                         "cartulary.tools.contractgen.extensions.v1",
		"generator_sources":                    objectsToAny(generatorSources),
		"generated_schema_digests":             objectsToAny(generatedSchemaRows),
	}, nil
}

func generatedSchemaID(canonical string) any {
	decoded, err := decodeContract([]byte(canonical))
	if err != nil {
		return "unknown"
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		return "unknown"
	}
	if schemaID, exists := object["schema_id"]; exists {
		return schemaID
	}
	return object["$id"]
}

func sortObjectRows(rows []map[string]any, key string) {
	sort.Slice(rows, func(i, j int) bool { return stringValue(rows[i][key]) < stringValue(rows[j][key]) })
}

func extensionGeneratedArtifact(relativePath string, value any) (artifact, error) {
	canonical, err := canonicalizeExtension(value)
	if err != nil {
		return artifact{}, err
	}
	digest := sha256.Sum256([]byte(canonical))
	return artifact{
		Path:   generatedExtensionPrefix + relativePath,
		JSON:   canonical,
		SHA256: hex.EncodeToString(digest[:]),
	}, nil
}

func stringValue(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func anyToStrings(value any) []string {
	values, _ := value.([]any)
	result := make([]string, 0, len(values))
	for _, current := range values {
		result = append(result, stringValue(current))
	}
	return result
}

func equalStringAny(expected []string, actual any) bool {
	actualStrings := anyToStrings(actual)
	if len(expected) != len(actualStrings) {
		return false
	}
	for index := range expected {
		if expected[index] != actualStrings[index] {
			return false
		}
	}
	return true
}

// extensionRuntimeArtifactPath selects the explicit build inventory. Inputs that
// only control generation are not exposed as runtime contracts.
func extensionRuntimeArtifactPath(path string) bool {
	const prefix = "contracts/extensions/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	relative := strings.TrimPrefix(path, prefix)
	if relative == "specification/inactive-value-schemas.json" || relative == "specification/shared-protocols.json" {
		return true
	}
	return relative != "index.json" && relative != "dependencies.json" &&
		relative != "build/implementation-bindings.json" &&
		!strings.HasPrefix(relative, "specification/") && !strings.HasPrefix(relative, "validation/")
}

// Owner fragments have one normalized fact order in every packaged occurrence.
// This is shared by authored artifact packaging and owner-input derivation.
func normalizeExtensionArtifact(value any) error {
	object, ok := value.(map[string]any)
	if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
		return nil
	}
	normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
	if err != nil {
		return err
	}
	object["facts"] = normalized[0]["facts"]
	return nil
}
