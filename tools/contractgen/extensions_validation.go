package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

var requiredExtensionDependencies = []string{
	"core00",
	"core01",
	"core02",
	"core03",
	"core04",
	"network_flow_activity",
	"opentelemetry",
	"report_composition",
	"reporting",
	"testing_harness",
}

var requiredExtensionProfiles = []string{
	"enterprise_authentication",
	"import",
	"incident_portability",
	"network_flow_activity",
	"reference_pack",
	"snapshot_reporting",
}

var requiredProfileScalarFacts = []string{
	"recognized_profile",
	"claim_configuration",
	"state_ownership",
	"admission_validation",
	"egress",
	"portability",
	"snapshot_reporting",
}

func validateExtensionArtifactShape(value any, relativePath string) error {
	object, err := asObject(value, "contracts/extensions/"+relativePath)
	if err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "contracts/extensions/"+relativePath)
	if err != nil {
		return err
	}
	switch schemaID {
	case "cartulary.extension_dependency_declaration_set.v3":
		return validateExtensionDependencyDeclarations(object)
	case "cartulary.extension_owner_fragment.v3":
		return validateExtensionOwnerFragment(object, relativePath)
	case "cartulary.extension_profile_configuration_contract.v3":
		return validateExtensionConfigurationContract(object, relativePath)
	case "cartulary.extension_validation_surface_declaration_set.v3":
		return validateExtensionValidationDeclarations(object)
	case "cartulary.base_route_reservation_registry.v3":
		return validateExtensionBaseReservations(object)
	case "cartulary.extension_implementation_binding_source_set.v1":
		return validateExtensionBindingSources(object)
	case "cartulary.extension_participant_contract.v1":
		return validateExtensionParticipantContract(object, relativePath)
	case "cartulary.extension_participant_specialization.v3":
		return validateExtensionParticipantSpecialization(object, relativePath)
	case "cartulary.extension_transaction_participant_contract.v3":
		return validateExtensionTransactionParticipantContract(object, relativePath)
	case "cartulary.extension_physical_state_binding.v1":
		return validateExtensionPhysicalStateBinding(object, relativePath)
	case "cartulary.extension_backup_binding_codec.v3":
		return validateExtensionBackupCodec(object, relativePath)
	case "cartulary.extension_generated_schema_source_set.v3":
		return validateExtensionGeneratedSchemaSources(object)
	default:
		return nil
	}
}

func validateExtensionTransactionParticipantContract(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet(
		"schema_id", "participant_id", "owner_profile_id", "participant_input_schema_id",
		"prepare_algorithm_id", "validation_algorithm_id", "write_algorithm_id",
		"serialization_key_kinds", "owned_state_family_ids",
	), relativePath); err != nil {
		return err
	}
	profileID, err := requiredString(object, "owner_profile_id", relativePath)
	if err != nil {
		return err
	}
	participantID, err := requiredString(object, "participant_id", relativePath)
	if err != nil || !strings.HasPrefix(participantID, profileID+".") {
		return fmt.Errorf("%s.participant_id must use the owner profile prefix", relativePath)
	}
	for _, key := range []string{"participant_input_schema_id", "prepare_algorithm_id", "validation_algorithm_id", "write_algorithm_id"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	keyKinds, err := sortedUniqueStringArray(object["serialization_key_kinds"], relativePath+".serialization_key_kinds", false)
	if err != nil || len(keyKinds) > 32 {
		return fmt.Errorf("%s.serialization_key_kinds must contain 1..32 sorted unique values", relativePath)
	}
	ownedFamilies, err := sortedUniqueStringArray(object["owned_state_family_ids"], relativePath+".owned_state_family_ids", false)
	if err != nil || len(ownedFamilies) > 64 {
		return fmt.Errorf("%s.owned_state_family_ids must contain 1..64 sorted unique values", relativePath)
	}
	for _, value := range append(keyKinds, ownedFamilies...) {
		if !strings.HasPrefix(value, profileID+".") {
			return fmt.Errorf("%s identity %s must use the owner profile prefix", relativePath, value)
		}
	}
	return nil
}

func validateExtensionGeneratedSchemaSources(object map[string]any) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "schemas"), "generated schema source set"); err != nil {
		return err
	}
	rows, err := objectArray(object["schemas"], "generated schema sources")
	if err != nil {
		return err
	}
	previous := ""
	for index, row := range rows {
		label := fmt.Sprintf("generated schema sources[%d]", index)
		if err := requireAllowedKeys(row, stringSet("target_schema_id", "members"), label); err != nil {
			return err
		}
		target, err := requiredString(row, "target_schema_id", label)
		if err != nil {
			return err
		}
		if previous != "" && previous >= target {
			return fmt.Errorf("generated schema sources must be sorted and unique")
		}
		previous = target
		members, err := asObject(row["members"], label+".members")
		if err != nil {
			return err
		}
		if members["schema_id"] != "string" {
			return fmt.Errorf("%s must declare schema_id as string", label)
		}
		allowedKinds := stringSet("string", "boolean", "integer", "array", "object", "string_or_null", "integer_or_null", "object_or_null")
		for member, rawKind := range members {
			kind, ok := rawKind.(string)
			_, allowed := allowedKinds[kind]
			if !ok || !allowed {
				return fmt.Errorf("%s member %s has invalid type source", label, member)
			}
		}
	}
	return nil
}

func validateExtensionBindingSources(object map[string]any) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "bindings"), "implementation binding source set"); err != nil {
		return err
	}
	bindings, err := objectArray(object["bindings"], "implementation binding sources")
	if err != nil {
		return err
	}
	if len(bindings) != len(requiredExtensionProfiles) {
		return fmt.Errorf("implementation binding sources must contain every recognized claimable profile")
	}
	for index, binding := range bindings {
		label := fmt.Sprintf("implementation binding sources[%d]", index)
		if err := requireAllowedKeys(binding, stringSet("profile_id", "contract_major", "implementation_id", "state_ownership_kind", "implemented_contribution_ids", "implemented_job_kinds", "implemented_worker_kinds", "algorithm_ids", "participant_implementations"), label); err != nil {
			return err
		}
		profileID, err := requiredString(binding, "profile_id", label)
		if err != nil {
			return err
		}
		if profileID != requiredExtensionProfiles[index] {
			return fmt.Errorf("implementation binding sources must be sorted and complete")
		}
		if _, err := positiveJSONInt(binding["contract_major"], label+".contract_major"); err != nil {
			return err
		}
		if _, err := requiredString(binding, "implementation_id", label); err != nil {
			return err
		}
		if kind := binding["state_ownership_kind"]; kind != "none" && kind != "core_managed" && kind != "extension_versioned" {
			return fmt.Errorf("%s.state_ownership_kind is invalid", label)
		}
		for _, key := range []string{"implemented_contribution_ids", "implemented_job_kinds", "implemented_worker_kinds", "algorithm_ids"} {
			values, err := sortedUniqueStringArray(binding[key], label+"."+key, true)
			if err != nil {
				return err
			}
			_ = values
		}
		rawParticipants, ok := binding["participant_implementations"].([]any)
		if !ok {
			return fmt.Errorf("%s.participant_implementations must be a present non-null array", label)
		}
		participants := make([]map[string]any, len(rawParticipants))
		for participantIndex, rawParticipant := range rawParticipants {
			participant, ok := rawParticipant.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.participant_implementations[%d] must be an object", label, participantIndex)
			}
			participants[participantIndex] = participant
		}
		previous := ""
		for participantIndex, participant := range participants {
			participantLabel := fmt.Sprintf("%s.participant_implementations[%d]", label, participantIndex)
			if err := requireAllowedKeys(participant, stringSet("participant_id", "algorithm_ids"), participantLabel); err != nil {
				return err
			}
			participantID, err := requiredString(participant, "participant_id", participantLabel)
			if err != nil {
				return err
			}
			if previous != "" && previous >= participantID {
				return fmt.Errorf("%s participant identities must be sorted and unique", label)
			}
			previous = participantID
			algorithms, err := sortedUniqueStringArray(participant["algorithm_ids"], participantLabel+".algorithm_ids", false)
			if err != nil || len(algorithms) == 0 || len(algorithms) > 16 {
				return fmt.Errorf("%s.algorithm_ids must contain 1..16 sorted unique values", participantLabel)
			}
		}
	}
	return nil
}

func validateExtensionParticipantSpecialization(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet(
		"schema_id", "profile_id", "participant_id", "participant_kind",
		"shared_context_schema_id", "operations",
	), relativePath); err != nil {
		return err
	}
	profileID, err := requiredString(object, "profile_id", relativePath)
	if err != nil {
		return err
	}
	participantID, err := requiredString(object, "participant_id", relativePath)
	if err != nil || !strings.HasPrefix(participantID, profileID+".") {
		return fmt.Errorf("%s.participant_id must use the profile prefix", relativePath)
	}
	participantKind, err := requiredString(object, "participant_kind", relativePath)
	if err != nil {
		return err
	}
	contextByKind := map[string]string{
		"incident_portability": "cartulary.extension_portability_participant_context.v1",
		"snapshot_reporting":   "cartulary.extension_snapshot_reporting_participant_context.v1",
		"backup_restore":       "cartulary.extension_backup_restore_participant_context.v1",
	}
	expectedContext, recognized := contextByKind[participantKind]
	if !recognized || object["shared_context_schema_id"] != expectedContext {
		return fmt.Errorf("%s has an invalid participant kind or shared context schema", relativePath)
	}
	resultByKindAndOperation := map[string]map[string]string{
		"incident_portability": {
			"export": "cartulary.extension_portability_export_result.v1",
			"import": "cartulary.extension_portability_import_preparation_result.v1",
		},
		"snapshot_reporting": {
			"emit": "cartulary.extension_snapshot_reporting_participant_result.v1",
		},
		"backup_restore": {
			"backup_enumerate": "cartulary.extension_backup_restore_participant_result.v1",
			"restore_rebuild":  "cartulary.extension_backup_restore_participant_result.v1",
			"restore_validate": "cartulary.extension_backup_restore_participant_result.v1",
		},
	}
	operations, err := objectArray(object["operations"], relativePath+".operations")
	if err != nil || len(operations) != len(resultByKindAndOperation[participantKind]) {
		return fmt.Errorf("%s.operations must contain the exact operation set", relativePath)
	}
	previous := ""
	for index, operation := range operations {
		label := fmt.Sprintf("%s.operations[%d]", relativePath, index)
		if err := requireAllowedKeys(operation, stringSet(
			"operation_kind", "result_schema_id", "algorithm_id", "output_schema_id",
			"ordering_algorithm_id", "state_family_ids", "max_input_bytes", "max_output_bytes",
			"max_items",
		), label); err != nil {
			return err
		}
		operationKind, err := requiredString(operation, "operation_kind", label)
		if err != nil || (previous != "" && previous >= operationKind) {
			return fmt.Errorf("%s operation identities must be sorted and unique", relativePath)
		}
		previous = operationKind
		if operation["result_schema_id"] != resultByKindAndOperation[participantKind][operationKind] {
			return fmt.Errorf("%s.result_schema_id does not match the participant operation", label)
		}
		for _, key := range []string{"algorithm_id", "output_schema_id", "ordering_algorithm_id"} {
			if _, err := requiredString(operation, key, label); err != nil {
				return err
			}
		}
		if _, err := sortedUniqueStringArray(operation["state_family_ids"], label+".state_family_ids", true); err != nil {
			return err
		}
		inputBytes, err := positiveJSONInt(operation["max_input_bytes"], label+".max_input_bytes")
		if err != nil || inputBytes > 67108864 {
			return fmt.Errorf("%s.max_input_bytes must be in 1..67108864", label)
		}
		outputBytes, err := nonnegativeJSONInt(operation["max_output_bytes"], label+".max_output_bytes")
		if err != nil || outputBytes > 67108864 {
			return fmt.Errorf("%s.max_output_bytes must be in 0..67108864", label)
		}
		items, err := nonnegativeJSONInt(operation["max_items"], label+".max_items")
		if err != nil || items > 1048576 {
			return fmt.Errorf("%s.max_items must be in 0..1048576", label)
		}
	}
	return nil
}

func validateExtensionParticipantContract(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "participant_id", "profile_id", "contribution_kind", "context_schema_id", "result_schema_id", "maximum_input_bytes", "maximum_result_bytes", "preparation_side_effects", "mutation_protocol", "staged_output_capability", "algorithm_ids"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"participant_id", "profile_id", "contribution_kind", "context_schema_id", "result_schema_id", "mutation_protocol"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	for _, key := range []string{"maximum_input_bytes", "maximum_result_bytes"} {
		value, err := positiveJSONInt(object[key], relativePath+"."+key)
		if err != nil || value > 67108864 {
			return fmt.Errorf("%s.%s must be in 1..67108864", relativePath, key)
		}
	}
	if object["preparation_side_effects"] != "forbidden" {
		return fmt.Errorf("%s preparation side effects must be forbidden", relativePath)
	}
	if _, err := sortedUniqueStringArray(object["algorithm_ids"], relativePath+".algorithm_ids", false); err != nil {
		return err
	}
	return nil
}

func validateExtensionPhysicalStateBinding(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "profile_id", "state_presence_manifest_sha256", "bindings"), relativePath); err != nil {
		return err
	}
	profileID, err := requiredString(object, "profile_id", relativePath)
	if err != nil {
		return err
	}
	digest, err := requiredString(object, "state_presence_manifest_sha256", relativePath)
	if err != nil || !isLowerSHA256(digest) {
		return fmt.Errorf("%s.state_presence_manifest_sha256 must be lowercase SHA-256", relativePath)
	}
	bindings, err := objectArray(object["bindings"], relativePath+".bindings")
	if err != nil || len(bindings) == 0 || len(bindings) > 4096 {
		return fmt.Errorf("%s.bindings must contain 1..4096 rows", relativePath)
	}
	seenIDs := map[string]struct{}{}
	seenFamilies := map[string]struct{}{}
	seenPhysical := map[string]struct{}{}
	previous := ""
	for index, binding := range bindings {
		label := fmt.Sprintf("%s.bindings[%d]", relativePath, index)
		if err := requireAllowedKeys(binding, stringSet("binding_id", "logical_family_id", "storage_kind", "physical_ref", "state_class", "backup_inclusion", "restore_order_group", "backup_codec_id", "backup_codec_sha256", "post_restore_validation_algorithm_id", "rebuild_algorithm_id"), label); err != nil {
			return err
		}
		bindingID, err := requiredString(binding, "binding_id", label)
		if err != nil || !strings.HasPrefix(bindingID, profileID+".") {
			return fmt.Errorf("%s.binding_id must use the profile prefix", label)
		}
		familyID, err := requiredString(binding, "logical_family_id", label)
		if err != nil {
			return err
		}
		physicalRef, err := requiredString(binding, "physical_ref", label)
		if err != nil {
			return err
		}
		if _, duplicate := seenIDs[bindingID]; duplicate {
			return fmt.Errorf("%s contains duplicate binding ID %s", relativePath, bindingID)
		}
		seenIDs[bindingID] = struct{}{}
		if _, duplicate := seenFamilies[familyID]; duplicate {
			return fmt.Errorf("%s contains duplicate logical family %s", relativePath, familyID)
		}
		seenFamilies[familyID] = struct{}{}
		if _, duplicate := seenPhysical[physicalRef]; duplicate {
			return fmt.Errorf("%s contains duplicate physical reference", relativePath)
		}
		seenPhysical[physicalRef] = struct{}{}
		storage := stringValue(binding["storage_kind"])
		if storage != "postgres" && storage != "object_store" && storage != "filesystem" {
			return fmt.Errorf("%s.storage_kind is invalid", label)
		}
		stateClass := stringValue(binding["state_class"])
		if stateClass != "authoritative" && stateClass != "derived" {
			return fmt.Errorf("%s.state_class is invalid", label)
		}
		if stateClass == "authoritative" && (binding["backup_inclusion"] != "required" || binding["rebuild_algorithm_id"] != nil) {
			return fmt.Errorf("%s authoritative binding must be required and non-rebuildable", label)
		}
		group, err := nonnegativeJSONInt(binding["restore_order_group"], label+".restore_order_group")
		if err != nil || group > 1024 {
			return fmt.Errorf("%s.restore_order_group must be in 0..1024", label)
		}
		codecDigest, err := requiredString(binding, "backup_codec_sha256", label)
		if err != nil || !isLowerSHA256(codecDigest) {
			return fmt.Errorf("%s.backup_codec_sha256 must be lowercase SHA-256", label)
		}
		key := fmt.Sprintf("%04d\x00%s\x00%s", group, familyID, bindingID)
		if previous != "" && previous >= key {
			return fmt.Errorf("%s.bindings must be sorted by restore group, family, and binding", relativePath)
		}
		previous = key
	}
	return nil
}

func validateExtensionBackupCodec(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "backup_codec_id", "binding_id", "storage_kind", "logical_identity_algorithm_id", "content_encoding_algorithm_id", "max_items", "max_entry_bytes", "max_binding_bytes", "historical_restore_codecs"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"backup_codec_id", "binding_id", "storage_kind", "logical_identity_algorithm_id", "content_encoding_algorithm_id"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	maxItems, err := positiveJSONInt(object["max_items"], relativePath+".max_items")
	if err != nil || maxItems > 1048576 {
		return fmt.Errorf("%s.max_items must be in 1..1048576", relativePath)
	}
	maxEntry, err := positiveJSONInt(object["max_entry_bytes"], relativePath+".max_entry_bytes")
	if err != nil {
		return err
	}
	maxBinding, err := positiveJSONInt(object["max_binding_bytes"], relativePath+".max_binding_bytes")
	if err != nil || maxEntry > maxBinding {
		return fmt.Errorf("%s codec byte bounds are invalid", relativePath)
	}
	historical, ok := object["historical_restore_codecs"].([]any)
	if !ok || len(historical) > 16 {
		return fmt.Errorf("%s.historical_restore_codecs must be a present array with at most 16 rows", relativePath)
	}
	return nil
}

func sortedUniqueStringArray(value any, label string, allowEmpty bool) ([]string, error) {
	rawValues, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a present non-null array", label)
	}
	if !allowEmpty && len(rawValues) == 0 {
		return nil, fmt.Errorf("%s must not be empty", label)
	}
	values := make([]string, len(rawValues))
	for index, rawValue := range rawValues {
		current, ok := rawValue.(string)
		if !ok || current == "" {
			return nil, fmt.Errorf("%s items must be nonempty strings", label)
		}
		if index > 0 && values[index-1] >= current {
			return nil, fmt.Errorf("%s must be sorted and unique", label)
		}
		values[index] = current
	}
	return values, nil
}

func validateExtensionDependencyDeclarations(object map[string]any) error {
	allowed := stringSet("schema_id", "dependencies")
	if err := requireExtensionExactKeys(object, allowed, "extension dependency declarations"); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.extension_dependency_declaration_set.v3" {
		return fmt.Errorf("extension dependency declarations have an invalid schema_id")
	}
	dependencies, err := objectArray(object["dependencies"], "dependencies")
	if err != nil {
		return err
	}
	if len(dependencies) != len(requiredExtensionDependencies) {
		return fmt.Errorf("dependencies must contain exactly %d rows", len(requiredExtensionDependencies))
	}
	rowKeys := stringSet("dependency_id", "imported_schema_ids", "imported_algorithm_ids", "imported_artifacts")
	for index, dependency := range dependencies {
		label := fmt.Sprintf("dependencies[%d]", index)
		if err := requireExtensionExactKeys(dependency, rowKeys, label); err != nil {
			return err
		}
		dependencyID, err := requiredString(dependency, "dependency_id", label)
		if err != nil {
			return err
		}
		if dependencyID != requiredExtensionDependencies[index] {
			return fmt.Errorf("dependencies must contain the exact sorted Table 1-B identities")
		}
		importedSchemas, err := sortedUniqueStringArray(dependency["imported_schema_ids"], label+".imported_schema_ids", true)
		if err != nil {
			return err
		}
		if _, err := sortedUniqueStringArray(dependency["imported_algorithm_ids"], label+".imported_algorithm_ids", true); err != nil {
			return err
		}
		artifacts, err := objectArrayAllowEmpty(dependency["imported_artifacts"], label+".imported_artifacts")
		if err != nil {
			return err
		}
		previousArtifact := ""
		for artifactIndex, artifact := range artifacts {
			artifactLabel := fmt.Sprintf("%s.imported_artifacts[%d]", label, artifactIndex)
			if err := requireExtensionExactKeys(artifact, stringSet("artifact_id", "schema_id", "artifact_sha256", "safe_ref"), artifactLabel); err != nil {
				return err
			}
			artifactID, err := requiredString(artifact, "artifact_id", artifactLabel)
			if err != nil {
				return err
			}
			schemaID, err := requiredString(artifact, "schema_id", artifactLabel)
			if err != nil {
				return err
			}
			digest, err := requiredString(artifact, "artifact_sha256", artifactLabel)
			if err != nil || !isLowerSHA256(digest) {
				return fmt.Errorf("%s.artifact_sha256 must be lowercase SHA-256", artifactLabel)
			}
			if artifact["safe_ref"] != "artifact:"+artifactID {
				return fmt.Errorf("%s.safe_ref must derive from artifact_id", artifactLabel)
			}
			if !containsString(importedSchemas, schemaID) {
				return fmt.Errorf("%s.schema_id must be declared in imported_schema_ids", artifactLabel)
			}
			identity := artifactID + "\x00" + schemaID + "\x00" + digest
			if previousArtifact != "" && previousArtifact >= identity {
				return fmt.Errorf("%s.imported_artifacts must be sorted and unique", label)
			}
			previousArtifact = identity
		}
	}
	return nil
}

func validateExtensionOwnerFragment(object map[string]any, relativePath string) error {
	if err := requireExtensionExactKeys(object, stringSet("schema_id", "owner_fragment_id", "owner_id", "facts"), relativePath); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
		return fmt.Errorf("%s.schema_id must be cartulary.extension_owner_fragment.v3", relativePath)
	}
	for _, key := range []string{"owner_fragment_id", "owner_id"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	facts, err := objectArray(object["facts"], relativePath+".facts")
	if err != nil {
		return err
	}
	if len(facts) == 0 || len(facts) > 4096 {
		return fmt.Errorf("%s.facts must contain 1..4096 rows", relativePath)
	}
	for index, fact := range facts {
		label := fmt.Sprintf("%s.facts[%d]", relativePath, index)
		for _, key := range []string{"fact_kind", "profile_id"} {
			if _, err := requiredString(fact, key, label); err != nil {
				return err
			}
		}
		if fact["fact_kind"] == "capability" {
			return fmt.Errorf("%s capability facts are prohibited in contract major 1", label)
		}
		switch fact["fact_kind"] {
		case "claim_configuration":
			if err := requireExtensionExactKeys(fact, stringSet("fact_kind", "profile_id", "claim_config_key", "configuration_contract_sha256"), label); err != nil {
				return err
			}
			profileID := stringValue(fact["profile_id"])
			if fact["claim_config_key"] != profileID+".claimed" {
				return fmt.Errorf("%s.claim_config_key is not canonical", label)
			}
			digest, err := requiredString(fact, "configuration_contract_sha256", label)
			if err != nil || !isLowerSHA256(digest) {
				return fmt.Errorf("%s.configuration_contract_sha256 must be lowercase SHA-256", label)
			}
		case "job_kind":
			contract, ok := fact["job_kind_contract"].(map[string]any)
			if !ok {
				return fmt.Errorf("%s.job_kind_contract must be an object", label)
			}
			if err := validateExtensionJobKindContract(contract, stringValue(fact["profile_id"]), label+".job_kind_contract"); err != nil {
				return err
			}
		case "worker_kind":
			workerKind, err := requiredString(fact, "worker_kind", label)
			if err != nil || !strings.HasPrefix(workerKind, stringValue(fact["profile_id"])+".") {
				return fmt.Errorf("%s.worker_kind must use the profile prefix", label)
			}
		}
	}
	return nil
}

func validateExtensionJobKindContract(object map[string]any, profileID, label string) error {
	if err := requireAllowedKeys(object, stringSet(
		"schema_id", "profile_id", "job_kind", "operation_kind", "proof_policy",
		"idempotency_policy", "idempotency_identity_schema_id", "terminal_result_schema_id",
		"resource_ref_contracts", "cancellation_policy", "max_proof_bytes",
	), label); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.extension_job_kind_contract.v1" || object["profile_id"] != profileID {
		return fmt.Errorf("%s must bind the extension job schema and owner profile", label)
	}
	for _, key := range []string{"job_kind", "operation_kind", "idempotency_identity_schema_id", "terminal_result_schema_id"} {
		if _, err := requiredString(object, key, label); err != nil {
			return err
		}
	}
	if !strings.HasPrefix(stringValue(object["job_kind"]), profileID+".") ||
		!strings.HasPrefix(stringValue(object["operation_kind"]), profileID+".") {
		return fmt.Errorf("%s job and operation identities must use the profile prefix", label)
	}
	if object["proof_policy"] != "required_on_terminal_success" ||
		object["idempotency_policy"] != "required" ||
		object["cancellation_policy"] != "precommit_observable" {
		return fmt.Errorf("%s must use the canonical proof, idempotency, and cancellation policies", label)
	}
	if object["idempotency_identity_schema_id"] != "cartulary.route_scoped_idempotency_identity.v1" ||
		object["terminal_result_schema_id"] != "cartulary.common_job_terminal_success.v1" {
		return fmt.Errorf("%s must use the Core common-job schemas", label)
	}
	proofBytes, err := positiveJSONInt(object["max_proof_bytes"], label+".max_proof_bytes")
	if err != nil || proofBytes > 1048576 {
		return fmt.Errorf("%s.max_proof_bytes must be in 1..1048576", label)
	}
	rawResourceRefs, ok := object["resource_ref_contracts"].([]any)
	if !ok || len(rawResourceRefs) > 64 {
		return fmt.Errorf("%s.resource_ref_contracts must contain 0..64 rows", label)
	}
	resourceRefs := make([]map[string]any, len(rawResourceRefs))
	for index, rawResourceRef := range rawResourceRefs {
		resourceRef, ok := rawResourceRef.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.resource_ref_contracts[%d] must be an object", label, index)
		}
		resourceRefs[index] = resourceRef
	}
	previous := ""
	for index, resourceRef := range resourceRefs {
		rowLabel := fmt.Sprintf("%s.resource_ref_contracts[%d]", label, index)
		if err := requireAllowedKeys(resourceRef, stringSet("resource_ref_kind", "resource_id_schema_id", "max_refs"), rowLabel); err != nil {
			return err
		}
		kind, err := requiredString(resourceRef, "resource_ref_kind", rowLabel)
		if err != nil || (previous != "" && previous >= kind) {
			return fmt.Errorf("%s resource reference kinds must be sorted and unique", label)
		}
		previous = kind
		if resourceRef["resource_id_schema_id"] != "cartulary.common_job_resource_ref_id.v1" {
			return fmt.Errorf("%s must use the Core resource-reference identity schema", rowLabel)
		}
		maxRefs, err := positiveJSONInt(resourceRef["max_refs"], rowLabel+".max_refs")
		if err != nil || maxRefs > 1024 {
			return fmt.Errorf("%s.max_refs must be in 1..1024", rowLabel)
		}
	}
	return nil
}

func validateExtensionConfigurationContract(object map[string]any, relativePath string) error {
	if err := requireExtensionExactKeys(object, stringSet("schema_id", "configuration_contract_id", "profile_id", "configuration_contract_major", "namespace_schema_id", "keys"), relativePath); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.extension_profile_configuration_contract.v3" {
		return fmt.Errorf("%s.schema_id must be cartulary.extension_profile_configuration_contract.v3", relativePath)
	}
	profileID, err := requiredString(object, "profile_id", relativePath)
	if err != nil {
		return err
	}
	major, err := positiveJSONInt(object["configuration_contract_major"], relativePath+".configuration_contract_major")
	if err != nil {
		return err
	}
	if object["configuration_contract_id"] != fmt.Sprintf("%s.configuration.v%d", profileID, major) {
		return fmt.Errorf("%s.configuration_contract_id must match profile_id and configuration_contract_major", relativePath)
	}
	if object["namespace_schema_id"] != fmt.Sprintf("cartulary.%s.configuration_namespace.v%d", profileID, major) {
		return fmt.Errorf("%s.namespace_schema_id must match the profile namespace and contract major", relativePath)
	}
	keys, ok := object["keys"].([]any)
	if !ok || len(keys) > 256 {
		return fmt.Errorf("%s.keys must be a present non-null array with at most 256 rows", relativePath)
	}
	previousKey := ""
	for index, rawKey := range keys {
		key, ok := rawKey.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.keys[%d] must be an object", relativePath, index)
		}
		label := fmt.Sprintf("%s.keys[%d]", relativePath, index)
		if err := requireExtensionExactKeys(key, stringSet(
			"key", "omission_policy", "inactive_policy", "resolution_kind",
			"diagnostic_policy", "value_schema_id", "inactive_value_schema_id",
		), label); err != nil {
			return err
		}
		keyPath, err := requiredString(key, "key", label)
		if err != nil || !strings.HasPrefix(keyPath, profileID+".") || keyPath == profileID+".claimed" {
			return fmt.Errorf("%s.key must be a non-claim key inside the profile namespace", label)
		}
		if previousKey != "" && previousKey >= keyPath {
			return fmt.Errorf("%s.keys must be sorted and unique by key", relativePath)
		}
		previousKey = keyPath
		if _, err := requiredString(key, "value_schema_id", label); err != nil {
			return err
		}
		omission, ok := key["omission_policy"].(map[string]any)
		if !ok {
			return fmt.Errorf("%s.omission_policy must be an object", label)
		}
		omissionKind, err := requiredString(omission, "kind", label+".omission_policy")
		if err != nil {
			return err
		}
		switch omissionKind {
		case "required", "absent":
			if err := requireExtensionExactKeys(omission, stringSet("kind"), label+".omission_policy"); err != nil {
				return err
			}
		case "default":
			if err := requireExtensionExactKeys(omission, stringSet("kind", "value"), label+".omission_policy"); err != nil {
				return err
			}
			if omission["value"] == nil {
				return fmt.Errorf("%s.omission_policy default value must be non-null", label)
			}
		default:
			return fmt.Errorf("%s.omission_policy.kind is invalid", label)
		}
		policy, err := requiredString(key, "inactive_policy", label)
		if err != nil {
			return err
		}
		ref, hasRef := key["inactive_value_schema_id"]
		if policy == "syntax_only" {
			if !hasRef || ref == nil {
				return fmt.Errorf("%s syntax_only requires a non-null inactive_value_schema_id", label)
			}
			if _, ok := ref.(string); !ok || ref == "" {
				return fmt.Errorf("%s inactive_value_schema_id must be a nonempty string", label)
			}
		} else if policy == "forbidden" && (!hasRef || ref != nil) {
			return fmt.Errorf("%s forbidden requires explicit null inactive_value_schema_id", label)
		} else if policy != "forbidden" {
			return fmt.Errorf("%s.inactive_policy is invalid", label)
		}
		if resolution, err := requiredString(key, "resolution_kind", label); err != nil ||
			(resolution != "plain" && resolution != "secret_ref" && resolution != "regular_file_ref" && resolution != "trust_material_ref") {
			return fmt.Errorf("%s.resolution_kind is invalid", label)
		}
		if diagnostic, err := requiredString(key, "diagnostic_policy", label); err != nil ||
			(diagnostic != "name_only" && diagnostic != "safe_value") {
			return fmt.Errorf("%s.diagnostic_policy is invalid", label)
		}
	}
	return nil
}

func requireExtensionExactKeys(object map[string]any, expected map[string]struct{}, label string) error {
	if err := requireAllowedKeys(object, expected, label); err != nil {
		return err
	}
	for key := range expected {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%s is missing required key %s", label, key)
		}
	}
	return nil
}

func validateExtensionValidationDeclarations(object map[string]any) error {
	declarations, err := objectArray(object["declarations"], "validation declarations")
	if err != nil {
		return err
	}
	if len(declarations) == 0 {
		return fmt.Errorf("validation declarations must not be empty")
	}
	seenConditions := map[string]struct{}{}
	requiredReasons := stringSet("extension_transaction_timeout", "backup_binding_codec_unsupported", "extension_staged_object_cleanup_dependency_failed")
	seenReasons := map[string]struct{}{}
	for declarationIndex, declaration := range declarations {
		label := fmt.Sprintf("declarations[%d]", declarationIndex)
		if err := requireAllowedKeys(declaration, stringSet("schema_id", "schema_surfaces", "procedural_surfaces"), label); err != nil {
			return err
		}
		if declaration["schema_id"] != "cartulary.extension_validation_surface_declaration.v3" {
			return fmt.Errorf("%s.schema_id is invalid", label)
		}
		for _, surfaceKey := range []string{"schema_surfaces", "procedural_surfaces"} {
			surfaces, ok := declaration[surfaceKey].([]any)
			if !ok {
				return fmt.Errorf("%s.%s must be a present non-null array", label, surfaceKey)
			}
			for _, rawSurface := range surfaces {
				surface, ok := rawSurface.(map[string]any)
				if !ok {
					return fmt.Errorf("%s.%s surface must be an object", label, surfaceKey)
				}
				conditions, ok := surface["conditions"].([]any)
				if !ok {
					return fmt.Errorf("%s.%s.conditions must be a present non-null array", label, surfaceKey)
				}
				for _, rawCondition := range conditions {
					condition, ok := rawCondition.(map[string]any)
					if !ok {
						return fmt.Errorf("validation condition must be an object")
					}
					conditionID, err := requiredString(condition, "condition_id", "validation condition")
					if err != nil {
						return err
					}
					if _, duplicate := seenConditions[conditionID]; duplicate {
						return fmt.Errorf("duplicate validation condition %s", conditionID)
					}
					seenConditions[conditionID] = struct{}{}
					reasonCode, err := requiredString(condition, "reason_code", "validation condition")
					if err != nil {
						return err
					}
					seenReasons[reasonCode] = struct{}{}
					if condition["secret_policy"] == "redacted" && condition["actual_formatter_id"] != "diagnostic_redacted_v1" {
						return fmt.Errorf("redacted condition %s must use diagnostic_redacted_v1", conditionID)
					}
				}
			}
		}
	}
	for reason := range requiredReasons {
		if _, ok := seenReasons[reason]; !ok {
			return fmt.Errorf("validation declarations omit required reason %s", reason)
		}
	}
	return nil
}

func validateExtensionBaseReservations(object map[string]any) error {
	reservations, err := objectArray(object["reservations"], "base route reservations")
	if err != nil {
		return err
	}
	seenIDs := map[string]struct{}{}
	seenPaths := map[string]struct{}{}
	previousKey := ""
	for index, reservation := range reservations {
		label := fmt.Sprintf("reservations[%d]", index)
		id, err := requiredString(reservation, "reservation_id", label)
		if err != nil {
			return err
		}
		path, err := requiredString(reservation, "path_template", label)
		if err != nil {
			return err
		}
		scope, err := requiredString(reservation, "match_scope", label)
		if err != nil || (scope != "exact" && scope != "descendants") {
			return fmt.Errorf("%s.match_scope must be exact or descendants", label)
		}
		key := path + "\x00" + scope + "\x00" + id
		if previousKey != "" && previousKey >= key {
			return fmt.Errorf("base route reservations must be sorted by path, scope, and ID")
		}
		previousKey = key
		if _, duplicate := seenIDs[id]; duplicate {
			return fmt.Errorf("duplicate base route reservation ID %s", id)
		}
		seenIDs[id] = struct{}{}
		pathKey := path + "\x00" + scope
		if _, duplicate := seenPaths[pathKey]; duplicate {
			return fmt.Errorf("duplicate base route reservation %s", pathKey)
		}
		seenPaths[pathKey] = struct{}{}
	}
	return nil
}

func validateExtensionSupportingContractParity(indexed map[string]map[string]any) error {
	participantDigests := map[string]string{}
	codecsByID := map[string]map[string]any{}
	statePresenceByProfile := map[string]map[string]any{}
	schemaIDs := map[string]struct{}{}
	algorithmIDs := map[string]struct{}{}
	definitionSet := indexed["specification/contract-definitions.json"]
	definitions, _ := objectArray(definitionSet["definitions"], "extension contract definitions")
	for _, definition := range definitions {
		schemaIDs[stringValue(definition["schema_id"])] = struct{}{}
	}
	dependencies := indexed["dependencies.json"]
	dependencyRows, _ := objectArray(dependencies["dependencies"], "extension dependencies")
	for _, dependency := range dependencyRows {
		for _, schemaID := range anyToStrings(dependency["imported_schema_ids"]) {
			schemaIDs[schemaID] = struct{}{}
		}
		for _, algorithmID := range anyToStrings(dependency["imported_algorithm_ids"]) {
			algorithmIDs[algorithmID] = struct{}{}
		}
	}
	for _, object := range indexed {
		switch object["schema_id"] {
		case "cartulary.extension_participant_contract.v1", "cartulary.extension_participant_specialization.v3", "cartulary.extension_transaction_participant_contract.v3":
			participantID := stringValue(object["participant_id"])
			digest, err := extensionCanonicalDigest(object)
			if err != nil {
				return err
			}
			if _, duplicate := participantDigests[participantID]; duplicate {
				return fmt.Errorf("duplicate participant contract %s", participantID)
			}
			participantDigests[participantID] = digest
			if object["schema_id"] == "cartulary.extension_participant_specialization.v3" {
				if _, exists := schemaIDs[stringValue(object["shared_context_schema_id"])]; !exists {
					return fmt.Errorf("participant %s has unresolved shared context schema", participantID)
				}
				operations, _ := objectArray(object["operations"], "participant specialization operations")
				for _, operation := range operations {
					for _, key := range []string{"result_schema_id", "output_schema_id"} {
						if _, exists := schemaIDs[stringValue(operation[key])]; !exists {
							return fmt.Errorf("participant %s has unresolved %s", participantID, key)
						}
					}
					for _, key := range []string{"algorithm_id", "ordering_algorithm_id"} {
						if _, exists := algorithmIDs[stringValue(operation[key])]; !exists {
							return fmt.Errorf("participant %s has unresolved %s", participantID, key)
						}
					}
				}
			}
		case "cartulary.extension_backup_binding_codec.v3":
			codecID := stringValue(object["backup_codec_id"])
			if _, duplicate := codecsByID[codecID]; duplicate {
				return fmt.Errorf("duplicate backup codec %s", codecID)
			}
			codecsByID[codecID] = object
		case "cartulary.extension_state_presence_manifest.v1":
			profileID := stringValue(object["profile_id"])
			statePresenceByProfile[profileID] = object
		}
	}
	usedParticipants := map[string]struct{}{}
	for _, object := range indexed {
		if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
			continue
		}
		facts, _ := objectArray(object["facts"], "owner fragment facts")
		for _, fact := range facts {
			if fact["fact_kind"] != "contribution" {
				continue
			}
			contribution, _ := fact["contribution"].(map[string]any)
			kind := stringValue(contribution["kind"])
			if !strings.HasSuffix(kind, "_participant") {
				continue
			}
			participantID := stringValue(contribution["participant_id"])
			digest, exists := participantDigests[participantID]
			if !exists {
				return fmt.Errorf("participant contribution %s has no indexed contract", participantID)
			}
			if contribution["participant_contract_sha256"] != digest {
				return fmt.Errorf("participant contribution %s has a stale contract digest", participantID)
			}
			usedParticipants[participantID] = struct{}{}
		}
	}
	for participantID := range participantDigests {
		if _, used := usedParticipants[participantID]; !used {
			return fmt.Errorf("participant contract %s has no owner contribution", participantID)
		}
	}
	usedCodecs := map[string]struct{}{}
	for _, object := range indexed {
		if object["schema_id"] != "cartulary.extension_physical_state_binding.v1" {
			continue
		}
		profileID := stringValue(object["profile_id"])
		manifest := statePresenceByProfile[profileID]
		if manifest == nil {
			return fmt.Errorf("physical state binding for %s has no state-presence manifest", profileID)
		}
		manifestDigest, err := extensionCanonicalDigest(manifest)
		if err != nil {
			return err
		}
		if object["state_presence_manifest_sha256"] != manifestDigest {
			return fmt.Errorf("physical state binding for %s has stale state-presence digest", profileID)
		}
		families := map[string]struct{}{}
		for _, key := range []string{"database_family_ids", "object_reference_family_ids"} {
			for _, familyID := range anyToStrings(manifest[key]) {
				families[familyID] = struct{}{}
			}
		}
		rows, _ := objectArray(object["bindings"], "physical state bindings")
		for _, row := range rows {
			familyID := stringValue(row["logical_family_id"])
			if row["state_class"] == "authoritative" {
				if _, exists := families[familyID]; !exists {
					return fmt.Errorf("authoritative binding %s is absent from state presence", familyID)
				}
				delete(families, familyID)
			}
			codecID := stringValue(row["backup_codec_id"])
			codec := codecsByID[codecID]
			if codec == nil {
				return fmt.Errorf("physical binding references missing codec %s", codecID)
			}
			codecDigest, err := extensionCanonicalDigest(codec)
			if err != nil {
				return err
			}
			if row["backup_codec_sha256"] != codecDigest || codec["binding_id"] != row["binding_id"] || codec["storage_kind"] != row["storage_kind"] {
				return fmt.Errorf("physical binding and codec %s do not have exact parity", codecID)
			}
			usedCodecs[codecID] = struct{}{}
		}
		if len(families) != 0 {
			return fmt.Errorf("physical state binding for %s omits authoritative families", profileID)
		}
	}
	for codecID := range codecsByID {
		if _, used := usedCodecs[codecID]; !used {
			return fmt.Errorf("backup codec %s has no physical binding", codecID)
		}
	}
	return nil
}

func validateExtensionProfileFactClosure(indexed map[string]map[string]any) error {
	counts := map[string]map[string]int{}
	configDigests := map[string]string{}
	claimConfigDigests := map[string]string{}
	for _, profileID := range requiredExtensionProfiles {
		counts[profileID] = map[string]int{}
	}
	for relativePath, object := range indexed {
		schemaID, _ := object["schema_id"].(string)
		if schemaID == "cartulary.extension_profile_configuration_contract.v3" {
			profileID, _ := object["profile_id"].(string)
			digest, err := extensionCanonicalDigest(object)
			if err != nil {
				return err
			}
			if _, duplicate := configDigests[profileID]; duplicate {
				return fmt.Errorf("duplicate profile configuration contract for %s", profileID)
			}
			configDigests[profileID] = digest
		}
		if schemaID != "cartulary.extension_owner_fragment.v3" {
			continue
		}
		facts, _ := objectArray(object["facts"], relativePath+".facts")
		for _, fact := range facts {
			profileID, _ := fact["profile_id"].(string)
			factKind, _ := fact["fact_kind"].(string)
			if _, recognized := counts[profileID]; !recognized {
				return fmt.Errorf("owner fact names unrecognized profile %s", profileID)
			}
			counts[profileID][factKind]++
			if factKind == "claim_configuration" {
				if fact["claim_config_key"] != profileID+".claimed" {
					return fmt.Errorf("profile %s has invalid claim_config_key", profileID)
				}
				claimConfigDigests[profileID] = stringValue(fact["configuration_contract_sha256"])
			}
		}
	}
	for _, profileID := range requiredExtensionProfiles {
		for _, factKind := range requiredProfileScalarFacts {
			if counts[profileID][factKind] != 1 {
				return fmt.Errorf("profile %s must have exactly one %s fact", profileID, factKind)
			}
		}
		if counts[profileID]["capability"] != 0 {
			return fmt.Errorf("profile %s declares a prohibited capability", profileID)
		}
		if _, ok := configDigests[profileID]; !ok {
			return fmt.Errorf("profile %s is missing its configuration contract", profileID)
		}
		if claimConfigDigests[profileID] != configDigests[profileID] {
			return fmt.Errorf("profile %s claim configuration has a stale contract digest", profileID)
		}
	}
	return nil
}

func validateExtensionConfigurationReferences(indexed map[string]map[string]any) error {
	schemaIDs := map[string]struct{}{}
	inactiveSchemaIDs := map[string]struct{}{}
	if definitions := indexed["specification/contract-definitions.json"]; definitions != nil {
		rows, _ := objectArray(definitions["definitions"], "extension contract definitions")
		for _, row := range rows {
			schemaIDs[stringValue(row["schema_id"])] = struct{}{}
		}
	}
	if dependencies := indexed["dependencies.json"]; dependencies != nil {
		rows, _ := objectArray(dependencies["dependencies"], "extension dependencies")
		for _, row := range rows {
			for _, schemaID := range anyToStrings(row["imported_schema_ids"]) {
				schemaIDs[schemaID] = struct{}{}
			}
		}
	}
	if inactiveSet := indexed["specification/inactive-value-schemas.json"]; inactiveSet != nil {
		rows, _ := objectArray(inactiveSet["schemas"], "inactive value schemas")
		for _, row := range rows {
			inactiveSchemaIDs[stringValue(row["inactive_value_schema_id"])] = struct{}{}
		}
	}
	for relativePath, object := range indexed {
		if object["schema_id"] != "cartulary.extension_profile_configuration_contract.v3" {
			continue
		}
		keys, _ := objectArrayAllowEmpty(object["keys"], relativePath+".keys")
		for index, key := range keys {
			label := fmt.Sprintf("%s.keys[%d]", relativePath, index)
			valueSchemaID := stringValue(key["value_schema_id"])
			if _, ok := schemaIDs[valueSchemaID]; !ok {
				return fmt.Errorf("%s.value_schema_id %s is unresolved", label, valueSchemaID)
			}
			if key["inactive_policy"] == "syntax_only" {
				inactiveSchemaID := stringValue(key["inactive_value_schema_id"])
				if _, ok := inactiveSchemaIDs[inactiveSchemaID]; !ok {
					return fmt.Errorf("%s.inactive_value_schema_id %s is unresolved", label, inactiveSchemaID)
				}
			}
		}
	}
	return nil
}

func validateExtensionOperationalBindings(indexed map[string]map[string]any) error {
	if err := validateExtensionProfileFactClosure(indexed); err != nil {
		return err
	}
	if err := validateExtensionConfigurationReferences(indexed); err != nil {
		return err
	}
	return validateExtensionSupportingContractParity(indexed)
}

func extensionCanonicalDigest(value any) (string, error) {
	canonical, err := canonicalizeDecoded(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(canonical + "\n"))
	return hex.EncodeToString(digest[:]), nil
}

func positiveJSONInt(value any, label string) (int64, error) {
	integer, err := jsonInt(value, label)
	if err != nil {
		return 0, err
	}
	if integer <= 0 {
		return 0, fmt.Errorf("%s must be positive", label)
	}
	return integer, nil
}

func nonnegativeJSONInt(value any, label string) (int64, error) {
	integer, err := jsonInt(value, label)
	if err != nil {
		return 0, err
	}
	if integer < 0 {
		return 0, fmt.Errorf("%s must be nonnegative", label)
	}
	return integer, nil
}

func jsonInt(value any, label string) (int64, error) {
	switch typed := value.(type) {
	case interface{ String() string }:
		integer, err := strconv.ParseInt(typed.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be an integer", label)
		}
		return integer, nil
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	default:
		return 0, fmt.Errorf("%s must be an integer", label)
	}
}

func isLowerSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}
