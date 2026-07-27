package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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

	declarations := indexed["dependencies.json"]
	dependencySnapshot := cloneObject(declarations)
	dependencySnapshot["schema_id"] = "cartulary.extension_dependency_snapshot.v3"

	fragments := collectExtensionObjects(indexed, "cartulary.extension_owner_fragment.v3", "owner_fragment_id")
	for _, fragment := range fragments {
		facts, _ := fragment["facts"].([]any)
		sort.Slice(facts, func(i, j int) bool {
			return extensionFactSortKey(facts[i].(map[string]any)) < extensionFactSortKey(facts[j].(map[string]any))
		})
	}
	ownerInput := map[string]any{
		"schema_id":           "cartulary.extension_owner_input_registry.v2",
		"dependency_snapshot": dependencySnapshot,
		"owner_fragments":     objectsToAny(fragments),
	}

	factsByProfile := extensionFactsByProfile(fragments)
	configsByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_profile_configuration_contract.v3")
	descriptors := make([]map[string]any, 0, len(requiredExtensionProfiles))
	descriptorDigest := map[string]string{}
	for _, profileID := range requiredExtensionProfiles {
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
	for _, profileID := range requiredExtensionProfiles {
		for _, contract := range extensionJobKindContracts(factsByProfile[profileID]) {
			if err := appendGenerated("job-contracts/"+profileID+"/"+stringValue(contract["job_kind"])+".json", contract); err != nil {
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
	generatedSchemas, err := materializeExtensionSchemas(indexed["specification/generated-schema-sources.json"])
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
	sort.Slice(generated, func(i, j int) bool { return generated[i].Path < generated[j].Path })
	return generated, nil
}

func readExtensionInputCatalog(root string) (map[string]map[string]any, error) {
	base := filepath.Join(root, "contracts", "extensions")
	data, err := os.ReadFile(filepath.Join(base, "index.json"))
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
		input, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(path)))
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
	return indexed, nil
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

func extensionFactSortKey(fact map[string]any) string {
	profileID, _ := fact["profile_id"].(string)
	factKind, _ := fact["fact_kind"].(string)
	identity := profileID
	switch factKind {
	case "runtime_dependency":
		identity += "\x00" + stringValue(fact["dependency_profile_id"])
	case "route_family":
		identity += "\x00" + stringValue(fact["route_family"])
	case "workspace":
		identity += "\x00" + stringValue(fact["workspace_key"])
	case "public_schema":
		identity = stringValue(fact["public_schema_id"])
	case "contribution":
		contribution, _ := fact["contribution"].(map[string]any)
		identity = stringValue(contribution["contribution_id"])
	case "migration_definition":
		definition, _ := fact["migration_definition"].(map[string]any)
		identity = stringValue(definition["migration_id"])
	case "worker_kind":
		identity = stringValue(fact["worker_kind"])
	case "job_kind":
		contract, _ := fact["job_kind_contract"].(map[string]any)
		identity = stringValue(contract["job_kind"])
	}
	return profileID + "\x00" + factKind + "\x00" + identity
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

func materializeExtensionBindings(indexed map[string]map[string]any, descriptors []map[string]any, descriptorDigests map[string]string, factsByProfile map[string][]map[string]any) ([]map[string]any, error) {
	source := indexed["build/implementation-bindings.json"]
	rawSources, err := objectArray(source["bindings"], "implementation binding sources")
	if err != nil {
		return nil, err
	}
	sources := map[string]map[string]any{}
	for _, row := range rawSources {
		profileID := stringValue(row["profile_id"])
		sources[profileID] = row
	}
	participantContracts := map[string]map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] == "cartulary.extension_participant_contract.v1" ||
			object["schema_id"] == "cartulary.extension_participant_specialization.v3" ||
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
			return nil, fmt.Errorf("profile %s has no implementation binding source", profileID)
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
				digest, err := extensionCanonicalDigest(contract)
				if err != nil {
					return nil, err
				}
				if digest != contribution["participant_contract_sha256"] {
					return nil, fmt.Errorf("profile %s participant %s digest is stale", profileID, participantID)
				}
				algorithmIDs := anyToStrings(contract["algorithm_ids"])
				if contract["schema_id"] == "cartulary.extension_participant_specialization.v3" {
					algorithmIDs = extensionSpecializationAlgorithmIDs(contract)
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
		for _, requiredAlgorithm := range []any{initializationAlgorithmID, finalValidation} {
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
		jobKinds := make([]string, len(jobContracts))
		for index, contract := range jobContracts {
			jobKinds[index] = stringValue(contract["job_kind"])
		}
		if !equalStringAny(workerKinds, bindingSource["implemented_worker_kinds"]) {
			return nil, fmt.Errorf("profile %s implementation worker set is incomplete or extra", profileID)
		}
		if !equalStringAny(jobKinds, bindingSource["implemented_job_kinds"]) {
			return nil, fmt.Errorf("profile %s implementation job set is incomplete or extra", profileID)
		}
		bindings = append(bindings, map[string]any{
			"schema_id":                           "cartulary.extension_implementation_binding.v1",
			"profile_id":                          profileID,
			"contract_major":                      descriptor["contract_major"],
			"descriptor_sha256":                   descriptorDigests[profileID],
			"implemented_contribution_ids":        stringsToAny(requiredContributionIDs),
			"supported_capability_ids":            []any{},
			"state_ownership_kind":                state["kind"],
			"preflight_algorithm_id":              extensionAlgorithmID(admission["preflight_algorithm_ref"]),
			"post_migration_algorithm_id":         extensionAlgorithmID(admission["post_migration_algorithm_ref"]),
			"initialization_definition_sha256":    initializationDigest,
			"initialization_algorithm_id":         initializationAlgorithmID,
			"final_state_validation_algorithm_id": finalValidation,
			"dependency_probe_ids":                stringsToAny(dependencyProbeIDs),
			"migration_definitions":               objectsToAny(migrationBindings),
			"physical_state_binding_sha256":       physicalDigest,
			"backup_codec_bindings":               objectsToAny(backupCodecBindings),
			"rebuild_algorithm_ids":               stringsToAny(rebuildAlgorithms),
			"transaction_participant_limits": map[string]any{
				"participant_input_bytes":            67108864,
				"aggregate_input_bytes":              67108864,
				"serialization_keys_per_participant": 1024,
				"aggregate_serialization_keys":       4096,
				"result_bytes":                       1048576,
				"validation_findings":                256,
			},
			"supporting_schema_ids": stringsToAny(extensionSupportingSchemas(profileID, indexed, participantBindings, jobContracts)),
			"worker_kinds":          stringsToAny(workerKinds),
			"job_kind_contracts":    objectsToAny(jobContracts),
			"participant_contracts": objectsToAny(participantBindings),
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
	case "http_route_family", "incident_workspace", "deployment_admin_panel", "authentication_entry", "import_target", "websocket_invalidation", "cross_owner_transaction_participant", "snapshot_reporting_participant", "incident_portability_participant", "backup_restore_participant":
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
		if presence == nil || initialization == nil || physical == nil || binding == nil {
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
		if presence["profile_id"] != profileID || presence["migration_lineage_id"] != state["migration_lineage_id"] || presence["empty_state_policy"] != state["empty_state_policy"] || initializationDigest != state["initialization_definition_sha256"] || physicalDigest != binding["physical_state_binding_sha256"] {
			return nil, nil, nil, fmt.Errorf("profile %s runtime state registry disagrees with admitted descriptor/binding", profileID)
		}
		initializationVariant, _ := initialization["initialization"].(map[string]any)
		bindingDigest, err := extensionCanonicalDigest(binding)
		if err != nil {
			return nil, nil, nil, err
		}
		runtimeMigrations := []map[string]any{}
		rawMigrations, _ := objectArray(binding["migration_definitions"], profileID+" migration bindings")
		for _, rawMigration := range rawMigrations {
			migration := cloneObject(rawMigration)
			migration["migration_lineage_id"] = state["migration_lineage_id"]
			migration["implementation_binding_profile_id"] = profileID
			migration["implementation_binding_sha256"] = bindingDigest
			runtimeMigrations = append(runtimeMigrations, migration)
		}
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
			"final_state_validation_algorithm_id":        binding["final_state_validation_algorithm_id"],
			"physical_state_binding_sha256":              physicalDigest,
			"implementation_binding_sha256":              bindingDigest,
		})

		codecRows := []map[string]any{}
		physicalBindings, _ := objectArray(physical["bindings"], profileID+" physical bindings")
		for _, physicalBinding := range physicalBindings {
			codecID := stringValue(physicalBinding["backup_codec_id"])
			var codec map[string]any
			for _, candidate := range indexed {
				if candidate["schema_id"] == "cartulary.extension_backup_binding_codec.v3" && candidate["backup_codec_id"] == codecID {
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
		backupRows = append(backupRows, map[string]any{
			"profile_id":                    profileID,
			"physical_state_binding_sha256": physicalDigest,
			"physical_state_binding":        cloneObject(physical),
			"codecs":                        objectsToAny(codecRows),
		})
	}
	sortObjectRows(stateRows, "profile_id")
	sortObjectRows(backupRows, "profile_id")

	participantRows := []map[string]any{}
	for _, contract := range indexed {
		schemaID := stringValue(contract["schema_id"])
		if schemaID != "cartulary.extension_participant_contract.v1" &&
			schemaID != "cartulary.extension_participant_specialization.v3" &&
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

func extensionAlgorithmID(value any) any {
	if value == nil {
		return nil
	}
	ref, ok := value.(string)
	if !ok {
		return nil
	}
	parts := strings.SplitN(ref, "#algorithm:", 2)
	if len(parts) != 2 {
		return nil
	}
	return parts[1]
}

func extensionSupportingSchemas(profileID string, indexed map[string]map[string]any, participants []map[string]any, jobContracts []map[string]any) []string {
	set := map[string]struct{}{}
	for _, object := range indexed {
		if object["profile_id"] != profileID {
			continue
		}
		if schemaID, ok := object["schema_id"].(string); ok {
			set[schemaID] = struct{}{}
		}
		for _, key := range []string{"context_schema_id", "result_schema_id", "shared_context_schema_id"} {
			if schemaID, ok := object[key].(string); ok {
				set[schemaID] = struct{}{}
			}
		}
		if operations, ok := object["operations"].([]any); ok {
			for _, rawOperation := range operations {
				operation, _ := rawOperation.(map[string]any)
				for _, key := range []string{"result_schema_id", "output_schema_id"} {
					if schemaID, ok := operation[key].(string); ok {
						set[schemaID] = struct{}{}
					}
				}
			}
		}
	}
	for _, contract := range jobContracts {
		for _, key := range []string{"schema_id", "idempotency_identity_schema_id", "terminal_result_schema_id"} {
			if schemaID, ok := contract[key].(string); ok {
				set[schemaID] = struct{}{}
			}
		}
		resourceRefs, _ := objectArray(contract["resource_ref_contracts"], "job resource reference contracts")
		for _, resourceRef := range resourceRefs {
			if schemaID, ok := resourceRef["resource_id_schema_id"].(string); ok {
				set[schemaID] = struct{}{}
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
					condition["surface_id"] = surface["surface_id"]
					conditions = append(conditions, condition)
				}
			}
		}
	}
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i]["condition_id"].(string) < conditions[j]["condition_id"].(string)
	})
	return map[string]any{
		"schema_id":  "cartulary.extension_validation_condition_registry.v2",
		"conditions": objectsToAny(conditions),
	}, nil
}

func materializeClientSupport(source map[string]any, descriptors []map[string]any) (map[string]any, error) {
	rows, err := objectArray(source["rows"], "client support source rows")
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

func sortedUniqueStrings(values []string) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func materializeExtensionSchemas(source map[string]any) ([]map[string]any, error) {
	rows, err := objectArray(source["schemas"], "generated schema sources")
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(rows))
	previous := ""
	for _, row := range rows {
		targetID := stringValue(row["target_schema_id"])
		if targetID == "" || (previous != "" && previous >= targetID) {
			return nil, fmt.Errorf("generated schema sources must be sorted and unique")
		}
		previous = targetID
		members, err := asObject(row["members"], "generated schema source members")
		if err != nil {
			return nil, err
		}
		memberNames := make([]string, 0, len(members))
		properties := map[string]any{}
		for memberName, rawKind := range members {
			memberNames = append(memberNames, memberName)
			kind := stringValue(rawKind)
			var property map[string]any
			switch kind {
			case "string", "boolean", "integer", "array", "object":
				property = map[string]any{"type": kind}
			case "string_or_null":
				property = map[string]any{"type": []any{"string", "null"}}
			case "integer_or_null":
				property = map[string]any{"type": []any{"integer", "null"}}
			case "object_or_null":
				property = map[string]any{"type": []any{"object", "null"}}
			default:
				return nil, fmt.Errorf("generated schema %s has unknown member kind %s", targetID, kind)
			}
			if memberName == "schema_id" {
				property["const"] = targetID
			}
			properties[memberName] = property
		}
		sort.Strings(memberNames)
		result = append(result, map[string]any{
			"$schema":              "https://json-schema.org/draft/2020-12/schema",
			"$id":                  targetID,
			"type":                 "object",
			"additionalProperties": false,
			"required":             stringsToAny(memberNames),
			"properties":           properties,
		})
	}
	return result, nil
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
			if path == "dependencies.json" || strings.HasPrefix(path, "specification/") || strings.HasPrefix(path, "validation/") || path == "build/implementation-bindings.json" {
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
	generatorSources := []map[string]any{}
	for _, sourceRef := range []string{"tools/contractgen/extensions_generation.go", "tools/contractgen/extensions_validation.go", "tools/contractgen/main.go", "tools/contractgen/validation.go"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sourceRef)))
		if err != nil {
			return nil, err
		}
		digest := sha256.Sum256(data)
		generatorSources = append(generatorSources, map[string]any{"source_ref": sourceRef, "source_sha256": hex.EncodeToString(digest[:])})
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
	canonical, err := canonicalizeDecoded(value)
	if err != nil {
		return artifact{}, err
	}
	canonical += "\n"
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
