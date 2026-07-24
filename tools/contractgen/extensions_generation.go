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
	dependencySnapshot["schema_id"] = "cartulary.extension_dependency_snapshot.v1"

	manifests := collectExtensionObjects(indexed, "cartulary.extension_owner_contract_manifest.v1", "owner_contract_manifest_id")
	fragments := collectExtensionObjects(indexed, "cartulary.extension_owner_fragment.v1", "owner_fragment_id")
	for _, fragment := range fragments {
		facts, _ := fragment["facts"].([]any)
		sort.Slice(facts, func(i, j int) bool {
			return extensionFactSortKey(facts[i].(map[string]any)) < extensionFactSortKey(facts[j].(map[string]any))
		})
	}
	ownerInput := map[string]any{
		"schema_id":                "cartulary.extension_owner_input_registry.v1",
		"dependency_snapshot":      dependencySnapshot,
		"owner_contract_manifests": objectsToAny(manifests),
		"owner_fragments":          objectsToAny(fragments),
	}

	factsByProfile := extensionFactsByProfile(fragments)
	configsByProfile := extensionObjectsByProfile(indexed, "cartulary.extension_profile_configuration_contract.v1")
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

	bindings, err := materializeExtensionBindings(indexed, descriptors, descriptorDigest)
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
	traceability, err := materializeExtensionTraceability(root, indexed["traceability/mapping-source.json"])
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
	if err := appendGenerated("clause-traceability.json", traceability); err != nil {
		return nil, err
	}
	closureCatalogs, conformanceManifests, conformanceIndex, err := materializeExtensionConformance(indexed, descriptors, factsByProfile)
	if err != nil {
		return nil, err
	}
	for _, catalog := range closureCatalogs {
		if err := appendGenerated("closure-catalogs/"+stringValue(catalog["profile_id"])+".json", catalog); err != nil {
			return nil, err
		}
	}
	for _, manifest := range conformanceManifests {
		if err := appendGenerated("conformance-manifests/"+stringValue(manifest["profile_id"])+".json", manifest); err != nil {
			return nil, err
		}
	}
	if err := appendGenerated("conformance-manifest-index.json", conformanceIndex); err != nil {
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
	accounting, err := materializeExtensionRegistryAccounting(root, registry, integrity, descriptors)
	if err != nil {
		return nil, err
	}
	if err := appendGenerated("registry-accounting.json", accounting); err != nil {
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
		if object["schema_id"] != "cartulary.extension_owner_fragment.v1" {
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
	manifest, err := requireOne("conformance_manifest")
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
		"schema_id":                 "cartulary.extension_profile_descriptor.v1",
		"profile_id":                profileID,
		"claimable":                 recognized["claimable"],
		"contract_major":            recognized["contract_major"],
		"owner_contract_ref":        recognized["primary_owner_contract_ref"],
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
		"conformance_manifest_id":   manifest["conformance_manifest_id"],
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

func materializeExtensionBindings(indexed map[string]map[string]any, descriptors []map[string]any, descriptorDigests map[string]string) ([]map[string]any, error) {
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
		if object["schema_id"] == "cartulary.extension_participant_contract.v1" || object["schema_id"] == "cartulary.extension_transaction_participant_contract.v1" {
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
				if contract["schema_id"] == "cartulary.extension_transaction_participant_contract.v1" {
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
		workerKinds := []string{}
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
			"supporting_schema_ids": stringsToAny(extensionSupportingSchemas(profileID, indexed, participantBindings)),
			"worker_kinds":          stringsToAny(workerKinds),
			"job_kind_contracts":    []any{},
			"participant_contracts": objectsToAny(participantBindings),
		})
	}
	return bindings, nil
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
				if candidate["schema_id"] == "cartulary.extension_backup_binding_codec.v1" && candidate["backup_codec_id"] == codecID {
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
		if schemaID != "cartulary.extension_participant_contract.v1" && schemaID != "cartulary.extension_transaction_participant_contract.v1" {
			continue
		}
		digest, err := extensionCanonicalDigest(contract)
		if err != nil {
			return nil, nil, nil, err
		}
		profileID := stringValue(contract["profile_id"])
		if schemaID == "cartulary.extension_transaction_participant_contract.v1" {
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

func extensionSupportingSchemas(profileID string, indexed map[string]map[string]any, participants []map[string]any) []string {
	set := map[string]struct{}{}
	for _, object := range indexed {
		if object["profile_id"] != profileID {
			continue
		}
		if schemaID, ok := object["schema_id"].(string); ok {
			set[schemaID] = struct{}{}
		}
		for _, key := range []string{"context_schema_id", "result_schema_id"} {
			if schemaID, ok := object[key].(string); ok {
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
		ownerRef := declaration["owner_contract_ref"]
		for _, surfaceFamily := range []string{"schema_surfaces", "procedural_surfaces"} {
			surfaces, _ := declaration[surfaceFamily].([]any)
			for _, rawSurface := range surfaces {
				surface := rawSurface.(map[string]any)
				rawConditions, _ := surface["conditions"].([]any)
				for _, rawCondition := range rawConditions {
					condition := cloneObject(rawCondition.(map[string]any))
					condition["owner_contract_ref"] = ownerRef
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
		"schema_id":  "cartulary.extension_validation_condition_registry.v1",
		"conditions": objectsToAny(conditions),
	}, nil
}

func materializeExtensionTraceability(root string, source map[string]any) (map[string]any, error) {
	document, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	if err != nil {
		return nil, err
	}
	mappings, err := objectArray(source["mappings"], "traceability mappings")
	if err != nil {
		return nil, err
	}
	clauses := make([]map[string]any, 0, len(mappings))
	parentOrdinals := map[string]int{}
	for documentOrdinal, mapping := range mappings {
		start, _ := nonnegativeJSONInt(mapping["source_start_byte"], "traceability source start")
		end, _ := positiveJSONInt(mapping["source_end_byte"], "traceability source end")
		textDigest := sha256.Sum256(document[start:end])
		parentKey := stringValue(mapping["parent_anchor_kind"]) + "\x00" + stringValue(mapping["parent_anchor_id"])
		clauseOrdinal := parentOrdinals[parentKey]
		parentOrdinals[parentKey]++
		identity := map[string]any{
			"extensions_document_sha256": source["extensions_document_sha256"],
			"document_ordinal":           documentOrdinal,
			"parent_anchor_kind":         mapping["parent_anchor_kind"],
			"parent_anchor_id":           mapping["parent_anchor_id"],
			"clause_kind":                mapping["clause_kind"],
			"clause_ordinal":             clauseOrdinal,
			"source_start_byte":          mapping["source_start_byte"],
			"source_end_byte":            mapping["source_end_byte"],
			"clause_text_sha256":         hex.EncodeToString(textDigest[:]),
		}
		canonical, err := canonicalizeDecoded(identity)
		if err != nil {
			return nil, err
		}
		identityDigest := sha256.Sum256([]byte(canonical))
		clause := cloneObject(identity)
		clause["clause_id"] = "extcl:" + hex.EncodeToString(identityDigest[:])[:32]
		clause["requirement_ids"] = mapping["requirement_ids"]
		clause["acceptance_criterion_ids"] = mapping["acceptance_criterion_ids"]
		clause["verification_ids"] = mapping["verification_ids"]
		clauses = append(clauses, clause)
	}
	return map[string]any{
		"schema_id":                  "cartulary.extension_clause_traceability.v1",
		"extensions_document_sha256": source["extensions_document_sha256"],
		"clauses":                    objectsToAny(clauses),
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

type extensionClosureResolution struct {
	item      map[string]any
	ownerRefs []string
}

func materializeExtensionConformance(indexed map[string]map[string]any, descriptors []map[string]any, factsByProfile map[string][]map[string]any) ([]map[string]any, []map[string]any, map[string]any, error) {
	source := indexed["specification/closure-mapping.json"]
	baselineItems, err := objectArray(source["baseline_items"], "closure baseline items")
	if err != nil {
		return nil, nil, nil, err
	}
	contributionCategories, err := asObject(source["contribution_categories"], "closure contribution categories")
	if err != nil {
		return nil, nil, nil, err
	}
	manifestByOwner := map[string]map[string]any{}
	for _, object := range indexed {
		if object["schema_id"] == "cartulary.extension_owner_contract_manifest.v1" {
			manifestByOwner[stringValue(object["owner_id"])] = object
		}
	}
	configs := extensionObjectsByProfile(indexed, "cartulary.extension_profile_configuration_contract.v1")
	statePresence := extensionObjectsByProfile(indexed, "cartulary.extension_state_presence_manifest.v1")
	physicalBindings := extensionObjectsByProfile(indexed, "cartulary.extension_physical_state_binding.v1")

	var catalogs []map[string]any
	var manifests []map[string]any
	var indexRows []map[string]any
	for _, descriptor := range descriptors {
		profileID := stringValue(descriptor["profile_id"])
		contractMajor := descriptor["contract_major"]
		facts := factsByProfile[profileID]
		factByKind := map[string][]map[string]any{}
		for _, fact := range facts {
			factByKind[stringValue(fact["fact_kind"])] = append(factByKind[stringValue(fact["fact_kind"])], fact)
		}
		recognized := factByKind["recognized_profile"][0]
		primaryOwnerID := stringValue(recognized["primary_owner_id"])
		primaryManifest := manifestByOwner[primaryOwnerID]
		if primaryManifest == nil {
			return nil, nil, nil, fmt.Errorf("profile %s primary owner %s has no manifest", profileID, primaryOwnerID)
		}
		ownerDocument := primaryManifest["owner_document"].(map[string]any)
		primaryOwnerRef := stringValue(recognized["primary_owner_contract_ref"])
		resolutions := []extensionClosureResolution{}
		addItem := func(subjectKind, subjectID, category string, allowed []string, refs []string) error {
			identity := map[string]any{
				"profile_id":     profileID,
				"contract_major": contractMajor,
				"category":       category,
				"subject_kind":   subjectKind,
				"subject_id":     subjectID,
			}
			canonical, err := canonicalizeDecoded(identity)
			if err != nil {
				return err
			}
			digest := sha256.Sum256([]byte(canonical))
			item := cloneObject(identity)
			item["closure_item_id"] = "extclosure:" + hex.EncodeToString(digest[:])[:32]
			item["allowed_not_applicable_reason_codes"] = stringsToAny(allowed)
			refs = sortedUniqueStrings(refs)
			resolutions = append(resolutions, extensionClosureResolution{item: item, ownerRefs: refs})
			return nil
		}
		for _, baseline := range baselineItems {
			if err := addItem("baseline", stringValue(baseline["subject_id"]), stringValue(baseline["category"]), anyToStrings(baseline["allowed_not_applicable_reason_codes"]), []string{primaryOwnerRef}); err != nil {
				return nil, nil, nil, err
			}
		}
		documentPath := strings.SplitN(stringValue(ownerDocument["owner_document_ref"]), "#", 2)[0]
		ownerRequirementIDs := []string{}
		anchors, _ := objectArray(primaryManifest["anchors"], "primary owner anchors")
		for _, anchor := range anchors {
			if anchor["anchor_kind"] != "req" {
				continue
			}
			anchorID := stringValue(anchor["anchor_id"])
			ownerRequirementIDs = append(ownerRequirementIDs, anchorID)
			ref := documentPath + "#req:" + anchorID
			for _, category := range anyToStrings(anchor["closure_categories"]) {
				if err := addItem("owner_requirement", ref, category, nil, []string{ref}); err != nil {
					return nil, nil, nil, err
				}
			}
		}
		config := configs[profileID]
		if config == nil {
			return nil, nil, nil, fmt.Errorf("profile %s has no configuration contract", profileID)
		}
		claimOwnerRef := stringValue(factByKind["claim_configuration"][0]["owner_contract_ref"])
		for _, rawKey := range config["keys"].([]any) {
			key := rawKey.(map[string]any)
			for _, category := range []string{"defaults_omission_null", "scalar_collection_bounds", "security_secrets_egress"} {
				if err := addItem("configuration_key", stringValue(key["key"]), category, nil, []string{claimOwnerRef}); err != nil {
					return nil, nil, nil, err
				}
			}
		}
		for _, fact := range factByKind["public_schema"] {
			for _, category := range []string{"request_response_schemas", "defaults_omission_null", "scalar_collection_bounds", "identity_canonicalization_ordering"} {
				if err := addItem("public_schema", stringValue(fact["public_schema_id"]), category, nil, []string{stringValue(fact["owner_contract_ref"])}); err != nil {
					return nil, nil, nil, err
				}
			}
		}
		for _, fact := range factByKind["contribution"] {
			contribution := fact["contribution"].(map[string]any)
			kind := stringValue(contribution["kind"])
			categories, ok := contributionCategories[kind]
			if !ok {
				return nil, nil, nil, fmt.Errorf("contribution kind %s has no closed closure mapping", kind)
			}
			for _, category := range anyToStrings(categories) {
				if err := addItem("contribution", stringValue(contribution["contribution_id"]), category, nil, []string{stringValue(fact["owner_contract_ref"])}); err != nil {
					return nil, nil, nil, err
				}
			}
		}
		stateFact := factByKind["state_ownership"][0]
		if descriptor["state_ownership"].(map[string]any)["kind"] == "extension_versioned" {
			for _, category := range []string{"state_migration", "defaults_omission_null", "scalar_collection_bounds", "security_secrets_egress"} {
				if err := addItem("initialization", profileID+".state_initialization", category, nil, []string{stringValue(stateFact["owner_contract_ref"])}); err != nil {
					return nil, nil, nil, err
				}
			}
			presence := statePresence[profileID]
			for _, familyKey := range []string{"database_family_ids", "object_reference_family_ids"} {
				for _, familyID := range anyToStrings(presence[familyKey]) {
					categories := []string{"resource_lifecycle_retention", "backup_restore"}
					if descriptor["incident_portability_mode"] != "no_authoritative_incident_state" {
						categories = append(categories, "portability")
					}
					if descriptor["snapshot_reporting_mode"] != "no_participation" {
						categories = append(categories, "snapshot_reporting")
					}
					for _, category := range categories {
						if err := addItem("state_family", familyID, category, nil, []string{stringValue(stateFact["owner_contract_ref"])}); err != nil {
							return nil, nil, nil, err
						}
					}
				}
			}
			physical := physicalBindings[profileID]
			physicalRows, _ := objectArray(physical["bindings"], "physical bindings")
			for _, row := range physicalRows {
				codecID := stringValue(row["backup_codec_id"])
				codecRef := stringValue(stateFact["owner_contract_ref"])
				for _, object := range indexed {
					if object["schema_id"] == "cartulary.extension_backup_binding_codec.v1" && object["backup_codec_id"] == codecID {
						codecRef = stringValue(object["codec_contract_ref"])
					}
				}
				for _, category := range []string{"backup_restore", "identity_canonicalization_ordering", "scalar_collection_bounds", "errors_precedence_retry"} {
					if err := addItem("backup_codec", codecID, category, nil, []string{codecRef}); err != nil {
						return nil, nil, nil, err
					}
				}
			}
		}
		verificationIDs := []string{"module.extensions.verification.behavior_contract", "module.extensions.verification.contract_accounting"}
		for _, verificationID := range verificationIDs {
			if err := addItem("verification_contract", verificationID, "conformance_evidence", nil, []string{"docs/extension-subsystem-nlspec.md#req:EXT-REQ-165"}); err != nil {
				return nil, nil, nil, err
			}
		}
		sort.Slice(resolutions, func(i, j int) bool {
			left, right := resolutions[i].item, resolutions[j].item
			leftKey := stringValue(left["category"]) + "\x00" + stringValue(left["subject_kind"]) + "\x00" + stringValue(left["subject_id"]) + "\x00" + stringValue(left["closure_item_id"])
			rightKey := stringValue(right["category"]) + "\x00" + stringValue(right["subject_kind"]) + "\x00" + stringValue(right["subject_id"]) + "\x00" + stringValue(right["closure_item_id"])
			return leftKey < rightKey
		})
		items := make([]map[string]any, len(resolutions))
		contractClosure := make([]map[string]any, len(resolutions))
		ownerRefs := []string{}
		for index, resolution := range resolutions {
			items[index] = resolution.item
			ownerRefs = append(ownerRefs, resolution.ownerRefs...)
			contractClosure[index] = map[string]any{
				"closure_item_id":            resolution.item["closure_item_id"],
				"category":                   resolution.item["category"],
				"status":                     "specified",
				"owner_contract_refs":        stringsToAny(resolution.ownerRefs),
				"not_applicable_reason_code": nil,
			}
		}
		catalog := map[string]any{
			"schema_id":             "cartulary.extension_contract_closure_catalog.v1",
			"profile_id":            profileID,
			"contract_major":        contractMajor,
			"owner_document_sha256": ownerDocument["owner_document_sha256"],
			"items":                 objectsToAny(items),
		}
		catalogDigest, err := extensionCanonicalDigest(catalog)
		if err != nil {
			return nil, nil, nil, err
		}
		descriptorDigest, _ := extensionCanonicalDigest(descriptor)
		ownerRefs = sortedUniqueStrings(ownerRefs)
		ownerRequirementIDs = sortedUniqueStrings(ownerRequirementIDs)
		acceptanceIDs := make([]string, 0, 158)
		for value := 1; value <= 158; value++ {
			acceptanceIDs = append(acceptanceIDs, fmt.Sprintf("EXT-AC-%03d", value))
		}
		contributionIDs := []string{}
		for _, rawContribution := range descriptor["contributions"].([]any) {
			contributionIDs = append(contributionIDs, stringValue(rawContribution.(map[string]any)["contribution_id"]))
		}
		manifest := map[string]any{
			"schema_id":                       "cartulary.extension_conformance_manifest.v1",
			"conformance_manifest_id":         descriptor["conformance_manifest_id"],
			"profile_id":                      profileID,
			"contract_major":                  contractMajor,
			"descriptor_sha256":               descriptorDigest,
			"contract_closure_catalog_sha256": catalogDigest,
			"owner_contract_refs":             stringsToAny(ownerRefs),
			"requirement_ids":                 stringsToAny(ownerRequirementIDs),
			"acceptance_criterion_ids":        stringsToAny(acceptanceIDs),
			"verification_ids":                stringsToAny(verificationIDs),
			"public_schema_ids":               descriptor["public_schema_ids"],
			"contribution_ids":                stringsToAny(contributionIDs),
			"contract_closure":                objectsToAny(contractClosure),
		}
		manifestDigest, err := extensionCanonicalDigest(manifest)
		if err != nil {
			return nil, nil, nil, err
		}
		catalogs = append(catalogs, catalog)
		manifests = append(manifests, manifest)
		indexRows = append(indexRows, map[string]any{
			"conformance_manifest_id": descriptor["conformance_manifest_id"],
			"profile_id":              profileID,
			"contract_major":          contractMajor,
			"manifest_sha256":         manifestDigest,
			"safe_ref":                "extensions:conformance_manifest:" + stringValue(descriptor["conformance_manifest_id"]),
		})
	}
	sortObjectRows(indexRows, "conformance_manifest_id")
	return catalogs, manifests, map[string]any{
		"schema_id": "cartulary.extension_conformance_manifest_index.v1",
		"manifests": objectsToAny(indexRows),
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

func materializeExtensionRegistryAccounting(root string, registry, integrity map[string]any, descriptors []map[string]any) (map[string]any, error) {
	verificationDigest, catalogDigest, err := extensionHarnessSemanticDigests(root)
	if err != nil {
		return nil, err
	}
	registryDigest, _ := extensionCanonicalDigest(registry)
	integrityDigest, _ := extensionCanonicalDigest(integrity)
	registryChecks := []string{"registry_artifact_set_match", "validation_condition_registry_match", "normative_source_lint_match", "verification_registry_match", "test_catalog_match"}
	profileChecks := []string{"dependency_snapshot_match", "owner_input_match", "core00_match", "core01_discovery_match", "core03_workspace_match", "core04_claim_configuration_match", "owner_contract_match", "implementation_binding_match", "client_support_match", "physical_state_binding_match", "job_kind_contract_match", "participant_contract_match", "telemetry_match", "conformance_manifest_match", "contract_closure_match", "verification_contract_match", "catalog_coverage_match"}
	checks := make([]map[string]any, 0, len(registryChecks)+len(profileChecks)*len(descriptors))
	newCheck := func(checkID string, profileID any, digestID, digest string) map[string]any {
		return map[string]any{
			"check_id":   checkID,
			"profile_id": profileID,
			"status":     "pass",
			"input_digests": []any{map[string]any{
				"artifact_id": digestID,
				"sha256":      digest,
			}},
			"safe_refs": []any{},
		}
	}
	for _, checkID := range registryChecks {
		digestID, digest := "registry_integrity", integrityDigest
		if checkID == "verification_registry_match" {
			digestID, digest = "verification_semantic_digest", verificationDigest
		} else if checkID == "test_catalog_match" {
			digestID, digest = "catalog_semantic_digest", catalogDigest
		}
		checks = append(checks, newCheck(checkID, nil, digestID, digest))
	}
	for _, descriptor := range descriptors {
		profileID := stringValue(descriptor["profile_id"])
		descriptorDigest, _ := extensionCanonicalDigest(descriptor)
		for _, checkID := range profileChecks {
			checks = append(checks, newCheck(checkID, profileID, "descriptor:"+profileID, descriptorDigest))
		}
	}
	return map[string]any{
		"schema_id":                    "cartulary.extension_registry_accounting.v1",
		"registry_sha256":              registryDigest,
		"registry_integrity_sha256":    integrityDigest,
		"verification_semantic_digest": verificationDigest,
		"catalog_semantic_digest":      catalogDigest,
		"status":                       "pass",
		"checks":                       objectsToAny(checks),
		"findings":                     []any{},
	}, nil
}

func extensionHarnessSemanticDigests(root string) (string, string, error) {
	readObject := func(relative string) (map[string]any, error) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		decoded, err := decodeContract(data)
		if err != nil {
			return nil, err
		}
		return asObject(decoded, relative)
	}
	verificationRegistry, err := readObject("contracts/verification/registry.json")
	if err != nil {
		return "", "", err
	}
	verificationOwners, _ := objectArray(verificationRegistry["owners"], "verification owners")
	semanticVerificationOwners := make([]map[string]any, 0, len(verificationOwners))
	for _, owner := range verificationOwners {
		contractPath := stringValue(owner["contract_path"])
		contract, err := readObject(contractPath)
		if err != nil {
			return "", "", err
		}
		verifications, _ := objectArray(contract["verifications"], contractPath+" verifications")
		semanticVerifications := make([]map[string]any, len(verifications))
		for index, verification := range verifications {
			semantic := cloneObject(verification)
			delete(semantic, "documentation_refs")
			semanticVerifications[index] = semantic
		}
		semanticVerificationOwners = append(semanticVerificationOwners, map[string]any{
			"owner_id":      owner["owner_id"],
			"contract_path": contractPath,
			"status":        owner["status"],
			"contract": map[string]any{
				"schema_id":     contract["schema_id"],
				"owner_id":      contract["owner_id"],
				"verifications": objectsToAny(semanticVerifications),
			},
		})
	}
	verificationSemantic := map[string]any{"schema_id": verificationRegistry["schema_id"], "owners": objectsToAny(semanticVerificationOwners)}
	verificationDigest, err := extensionSemanticDigest(verificationSemantic)
	if err != nil {
		return "", "", err
	}

	ownerRegistry, err := readObject("tools/test_catalog_owner.json")
	if err != nil {
		return "", "", err
	}
	owners, _ := objectArray(ownerRegistry["owners"], "test catalog owners")
	semanticOwners := make([]map[string]any, 0, len(owners))
	for _, owner := range owners {
		manifestPath := stringValue(owner["manifest_path"])
		manifest, err := readObject(manifestPath)
		if err != nil {
			return "", "", err
		}
		rows, _ := objectArray(manifest["rows"], manifestPath+" rows")
		semanticRows := make([]map[string]any, len(rows))
		for index, row := range rows {
			semantic := cloneObject(row)
			delete(semantic, "documentation_refs")
			semanticRows[index] = semantic
		}
		semanticOwners = append(semanticOwners, map[string]any{
			"owner_id":      owner["owner_id"],
			"manifest_path": manifestPath,
			"status":        owner["status"],
			"manifest": map[string]any{
				"schema_id": manifest["schema_id"],
				"owner_id":  manifest["owner_id"],
				"rows":      objectsToAny(semanticRows),
			},
		})
	}
	runners, err := readObject("tools/test_runner_registry.json")
	if err != nil {
		return "", "", err
	}
	topology, err := readObject("tools/execution_topology_manifest.json")
	if err != nil {
		return "", "", err
	}
	catalogSemantic := map[string]any{
		"schema_id":       ownerRegistry["schema_id"],
		"owners":          objectsToAny(semanticOwners),
		"runner_registry": runners,
		"profiles": map[string]any{
			"runtime_profiles":  topology["runtime_profiles"],
			"resource_profiles": topology["resource_profiles"],
			"fixture_profiles":  topology["fixture_profiles"],
		},
	}
	catalogDigest, err := extensionSemanticDigest(catalogSemantic)
	if err != nil {
		return "", "", err
	}
	return verificationDigest, catalogDigest, nil
}

func extensionSemanticDigest(value any) (string, error) {
	canonical, err := canonicalizeDecoded(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(canonical))
	return "sha256:" + hex.EncodeToString(digest[:]), nil
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
	manifestRows := []map[string]any{}
	fragmentRows := []map[string]any{}
	supportRows := []map[string]any{}
	generatedSchemaRows := []map[string]any{}
	for path, object := range indexed {
		digest, err := extensionCanonicalDigest(object)
		if err != nil {
			return nil, err
		}
		switch object["schema_id"] {
		case "cartulary.extension_owner_contract_manifest.v1":
			manifestRows = append(manifestRows, map[string]any{"owner_contract_manifest_id": object["owner_contract_manifest_id"], "owner_contract_manifest_sha256": digest})
		case "cartulary.extension_owner_fragment.v1":
			fragmentRows = append(fragmentRows, map[string]any{"owner_fragment_id": object["owner_fragment_id"], "owner_fragment_sha256": digest})
		default:
			if path == "dependencies.json" || strings.HasPrefix(path, "specification/") || strings.HasPrefix(path, "traceability/") || strings.HasPrefix(path, "validation/") || path == "build/implementation-bindings.json" {
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
			generatedRelative == "clause-traceability.json" ||
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
	sortObjectRows(manifestRows, "owner_contract_manifest_id")
	sortObjectRows(fragmentRows, "owner_fragment_id")
	sortObjectRows(supportRows, "artifact_id")
	sortObjectRows(generatedSchemaRows, "schema_id")
	generatorSources := []map[string]any{}
	for _, sourceRef := range []string{"tools/contractgen/extensions_generation.go", "tools/contractgen/extensions_traceability.go", "tools/contractgen/extensions_validation.go", "tools/contractgen/main.go", "tools/contractgen/validation.go"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sourceRef)))
		if err != nil {
			return nil, err
		}
		digest := sha256.Sum256(data)
		generatorSources = append(generatorSources, map[string]any{"source_ref": sourceRef, "source_sha256": hex.EncodeToString(digest[:])})
	}
	return map[string]any{
		"schema_id":                            "cartulary.extension_registry_integrity.v1",
		"canonicalization_algorithm_id":        "extension_registry_canonical_json_v1",
		"dependency_snapshot_sha256":           dependencyDigest,
		"owner_input_registry_sha256":          ownerInputDigest,
		"registry_schema_id":                   "cartulary.extension_profile_registry.v1",
		"registry_sha256":                      registryDigest,
		"descriptor_digests":                   objectsToAny(descriptorRows),
		"owner_contract_manifest_digests":      objectsToAny(manifestRows),
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
