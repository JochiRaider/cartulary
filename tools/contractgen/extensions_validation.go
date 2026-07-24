package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	"conformance_manifest",
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
	case "cartulary.extension_dependency_declaration_set.v1":
		return validateExtensionDependencyDeclarations(object)
	case "cartulary.extension_owner_contract_manifest.v1":
		return validateExtensionOwnerManifest(object, relativePath)
	case "cartulary.extension_owner_fragment.v1":
		return validateExtensionOwnerFragment(object, relativePath)
	case "cartulary.extension_profile_configuration_contract.v1":
		return validateExtensionConfigurationContract(object, relativePath)
	case "cartulary.extension_traceability_mapping_source.v1":
		return validateExtensionTraceabilitySource(object)
	case "cartulary.extension_validation_surface_declaration_set.v1":
		return validateExtensionValidationDeclarations(object)
	case "cartulary.base_route_reservation_registry.v1":
		return validateExtensionBaseReservations(object)
	case "cartulary.extension_implementation_binding_source_set.v1":
		return validateExtensionBindingSources(object)
	case "cartulary.extension_participant_contract.v1":
		return validateExtensionParticipantContract(object, relativePath)
	case "cartulary.extension_transaction_participant_contract.v1":
		return validateExtensionTransactionParticipantContract(object, relativePath)
	case "cartulary.extension_physical_state_binding.v1":
		return validateExtensionPhysicalStateBinding(object, relativePath)
	case "cartulary.extension_backup_binding_codec.v1":
		return validateExtensionBackupCodec(object, relativePath)
	case "cartulary.extension_generated_schema_source_set.v1":
		return validateExtensionGeneratedSchemaSources(object)
	default:
		return nil
	}
}

func validateExtensionTransactionParticipantContract(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet(
		"schema_id", "participant_id", "owner_profile_id", "participant_input_schema_id",
		"prepare_algorithm_id", "validation_algorithm_id", "write_algorithm_id",
		"serialization_key_kinds", "owned_state_family_ids", "error_contract_ref",
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
	for _, key := range []string{"participant_input_schema_id", "prepare_algorithm_id", "validation_algorithm_id", "write_algorithm_id", "error_contract_ref"} {
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
		if err := requireAllowedKeys(binding, stringSet("profile_id", "contract_major", "implementation_id", "state_ownership_kind", "implemented_contribution_ids", "algorithm_ids", "participant_implementations"), label); err != nil {
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
		for _, key := range []string{"implemented_contribution_ids", "algorithm_ids"} {
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
	if err := requireAllowedKeys(object, stringSet("schema_id", "backup_codec_id", "binding_id", "storage_kind", "codec_contract_ref", "logical_identity_algorithm_id", "content_encoding_algorithm_id", "max_items", "max_entry_bytes", "max_binding_bytes", "historical_restore_codecs"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"backup_codec_id", "binding_id", "storage_kind", "codec_contract_ref", "logical_identity_algorithm_id", "content_encoding_algorithm_id"} {
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
	allowed := stringSet("schema_id", "extensions_document_version", "dependencies")
	if err := requireAllowedKeys(object, allowed, "extension dependency declarations"); err != nil {
		return err
	}
	if object["extensions_document_version"] != "0.6.1" {
		return fmt.Errorf("extension dependency declarations have stale extensions_document_version")
	}
	dependencies, err := objectArray(object["dependencies"], "dependencies")
	if err != nil {
		return err
	}
	if len(dependencies) != len(requiredExtensionDependencies) {
		return fmt.Errorf("dependencies must contain exactly %d rows", len(requiredExtensionDependencies))
	}
	rowKeys := stringSet(
		"dependency_id", "owner_document_ref", "owner_document_schema_id", "owner_document_version",
		"owner_document_sha256", "owner_contract_manifest_ref", "owner_contract_manifest_id",
		"owner_contract_manifest_sha256", "imported_anchor_refs", "imported_schema_ids",
		"imported_algorithm_ids", "imported_artifacts", "required_status",
	)
	for index, dependency := range dependencies {
		label := fmt.Sprintf("dependencies[%d]", index)
		if err := requireAllowedKeys(dependency, rowKeys, label); err != nil {
			return err
		}
		dependencyID, err := requiredString(dependency, "dependency_id", label)
		if err != nil {
			return err
		}
		if dependencyID != requiredExtensionDependencies[index] {
			return fmt.Errorf("dependencies must contain the exact sorted Table 1-B identities")
		}
		for _, key := range []string{"owner_document_ref", "owner_document_schema_id", "owner_document_version", "owner_contract_manifest_ref", "owner_contract_manifest_id"} {
			if _, err := requiredString(dependency, key, label); err != nil {
				return err
			}
		}
		for _, key := range []string{"owner_document_sha256", "owner_contract_manifest_sha256"} {
			digest, err := requiredString(dependency, key, label)
			if err != nil {
				return err
			}
			if !isLowerSHA256(digest) {
				return fmt.Errorf("%s.%s must be lowercase SHA-256", label, key)
			}
		}
		if dependency["required_status"] != "adopted/current" {
			return fmt.Errorf("%s.required_status must be adopted/current", label)
		}
		for _, key := range []string{"imported_anchor_refs", "imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
			array, ok := dependency[key].([]any)
			if !ok {
				return fmt.Errorf("%s.%s must be a present non-null array", label, key)
			}
			if key == "imported_anchor_refs" && len(array) == 0 {
				return fmt.Errorf("%s.imported_anchor_refs must not be empty", label)
			}
		}
	}
	return nil
}

func validateExtensionOwnerManifest(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "owner_contract_manifest_id", "owner_id", "owner_document", "anchors", "owner_fragments"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"owner_contract_manifest_id", "owner_id"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	document, err := asObject(object["owner_document"], relativePath+".owner_document")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(document, stringSet("owner_document_ref", "owner_document_schema_id", "owner_document_version", "owner_document_sha256", "byte_length"), relativePath+".owner_document"); err != nil {
		return err
	}
	for _, key := range []string{"owner_document_ref", "owner_document_schema_id", "owner_document_version"} {
		if _, err := requiredString(document, key, relativePath+".owner_document"); err != nil {
			return err
		}
	}
	digest, err := requiredString(document, "owner_document_sha256", relativePath+".owner_document")
	if err != nil || !isLowerSHA256(digest) {
		return fmt.Errorf("%s.owner_document.owner_document_sha256 must be lowercase SHA-256", relativePath)
	}
	if _, err := positiveJSONInt(document["byte_length"], relativePath+".owner_document.byte_length"); err != nil {
		return err
	}
	anchors, err := objectArray(object["anchors"], relativePath+".anchors")
	if err != nil {
		return err
	}
	if len(anchors) == 0 || len(anchors) > 4096 {
		return fmt.Errorf("%s.anchors must contain 1..4096 rows", relativePath)
	}
	fragments, ok := object["owner_fragments"].([]any)
	if !ok {
		return fmt.Errorf("%s.owner_fragments must be a present non-null array", relativePath)
	}
	previous := ""
	for index, rawFragment := range fragments {
		label := fmt.Sprintf("%s.owner_fragments[%d]", relativePath, index)
		fragment, ok := rawFragment.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be an object", label)
		}
		if err := requireAllowedKeys(fragment, stringSet("owner_fragment_id", "owner_fragment_ref", "owner_fragment_sha256"), label); err != nil {
			return err
		}
		id, err := requiredString(fragment, "owner_fragment_id", label)
		if err != nil {
			return err
		}
		if previous != "" && previous >= id {
			return fmt.Errorf("%s.owner_fragments must be sorted and unique", relativePath)
		}
		previous = id
		ref, err := requiredString(fragment, "owner_fragment_ref", label)
		if err != nil || !strings.HasPrefix(ref, "contracts/extensions/fragments/") || !validExtensionCatalogPath(strings.TrimPrefix(ref, "contracts/extensions/")) {
			return fmt.Errorf("%s.owner_fragment_ref must name a normalized extension fragment", label)
		}
		digest, err := requiredString(fragment, "owner_fragment_sha256", label)
		if err != nil || !isLowerSHA256(digest) {
			return fmt.Errorf("%s.owner_fragment_sha256 must be lowercase SHA-256", label)
		}
	}
	return nil
}

func validateExtensionOwnerFragment(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "owner_fragment_id", "owner_id", "owner_document_ref", "owner_document_schema_id", "owner_document_version", "owner_document_sha256", "facts"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"owner_fragment_id", "owner_id", "owner_document_ref", "owner_document_schema_id", "owner_document_version"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	digest, err := requiredString(object, "owner_document_sha256", relativePath)
	if err != nil || !isLowerSHA256(digest) {
		return fmt.Errorf("%s.owner_document_sha256 must be lowercase SHA-256", relativePath)
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
		for _, key := range []string{"fact_kind", "profile_id", "owner_contract_ref"} {
			if _, err := requiredString(fact, key, label); err != nil {
				return err
			}
		}
		if fact["fact_kind"] == "capability" {
			return fmt.Errorf("%s capability facts are prohibited in contract major 1", label)
		}
	}
	return nil
}

func validateExtensionConfigurationContract(object map[string]any, relativePath string) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "configuration_contract_id", "profile_id", "configuration_contract_major", "namespace_schema_id", "keys"), relativePath); err != nil {
		return err
	}
	for _, key := range []string{"configuration_contract_id", "profile_id", "namespace_schema_id"} {
		if _, err := requiredString(object, key, relativePath); err != nil {
			return err
		}
	}
	if _, err := positiveJSONInt(object["configuration_contract_major"], relativePath+".configuration_contract_major"); err != nil {
		return err
	}
	keys, ok := object["keys"].([]any)
	if !ok {
		return fmt.Errorf("%s.keys must be a present non-null array", relativePath)
	}
	for index, rawKey := range keys {
		key, ok := rawKey.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.keys[%d] must be an object", relativePath, index)
		}
		label := fmt.Sprintf("%s.keys[%d]", relativePath, index)
		policy, err := requiredString(key, "inactive_policy", label)
		if err != nil {
			return err
		}
		ref, hasRef := key["inactive_value_schema_ref"]
		if policy == "syntax_only" {
			if !hasRef || ref == nil {
				return fmt.Errorf("%s syntax_only requires a non-null inactive_value_schema_ref", label)
			}
		} else if policy == "forbidden" && (!hasRef || ref != nil) {
			return fmt.Errorf("%s forbidden requires explicit null inactive_value_schema_ref", label)
		}
	}
	return nil
}

func validateExtensionTraceabilitySource(object map[string]any) error {
	if err := requireAllowedKeys(object, stringSet("schema_id", "extensions_document_sha256", "mappings"), "extension traceability source"); err != nil {
		return err
	}
	digest, err := requiredString(object, "extensions_document_sha256", "extension traceability source")
	if err != nil || !isLowerSHA256(digest) {
		return fmt.Errorf("extension traceability source digest must be lowercase SHA-256")
	}
	mappings, err := objectArray(object["mappings"], "traceability mappings")
	if err != nil {
		return err
	}
	if len(mappings) == 0 || len(mappings) > 65536 {
		return fmt.Errorf("traceability mappings must contain 1..65536 clauses")
	}
	previousEnd := int64(-1)
	for index, mapping := range mappings {
		label := fmt.Sprintf("traceability mappings[%d]", index)
		if err := requireAllowedKeys(mapping, stringSet("source_start_byte", "source_end_byte", "parent_anchor_kind", "parent_anchor_id", "clause_kind", "requirement_ids", "acceptance_criterion_ids", "verification_ids"), label); err != nil {
			return err
		}
		start, err := nonnegativeJSONInt(mapping["source_start_byte"], "mapping source_start_byte")
		if err != nil {
			return err
		}
		end, err := positiveJSONInt(mapping["source_end_byte"], "mapping source_end_byte")
		if err != nil {
			return err
		}
		if start >= end || (previousEnd >= 0 && start < previousEnd) {
			return fmt.Errorf("traceability mapping ranges must be nonempty, sorted, and nonoverlapping")
		}
		previousEnd = end
		parentKind, err := requiredString(mapping, "parent_anchor_kind", label)
		if err != nil || !extensionStringIn(parentKind, "document", "h1", "requirement") {
			return fmt.Errorf("%s.parent_anchor_kind is invalid", label)
		}
		if _, err := requiredString(mapping, "parent_anchor_id", label); err != nil {
			return err
		}
		clauseKind, err := requiredString(mapping, "clause_kind", label)
		if err != nil || !extensionStringIn(clauseKind, "frontmatter_member", "normative_table_caption", "normative_table_row", "list_item", "fenced_literal", "prose_block", "acceptance_row") {
			return fmt.Errorf("%s.clause_kind is invalid", label)
		}
		requirements, err := sortedUniqueStringArray(mapping["requirement_ids"], label+".requirement_ids", false)
		if err != nil {
			return err
		}
		criteria, err := sortedUniqueStringArray(mapping["acceptance_criterion_ids"], label+".acceptance_criterion_ids", false)
		if err != nil {
			return err
		}
		verifications, err := sortedUniqueStringArray(mapping["verification_ids"], label+".verification_ids", false)
		if err != nil {
			return err
		}
		for _, requirementID := range requirements {
			if !regexp.MustCompile(`^EXT-REQ-[0-9]{3}$`).MatchString(requirementID) {
				return fmt.Errorf("%s has invalid requirement ID %q", label, requirementID)
			}
		}
		for _, criterionID := range criteria {
			if !regexp.MustCompile(`^EXT-AC-[0-9]{3}$`).MatchString(criterionID) {
				return fmt.Errorf("%s has invalid acceptance criterion ID %q", label, criterionID)
			}
		}
		for _, verificationID := range verifications {
			if verificationID != "module.extensions.verification.behavior_contract" && verificationID != "module.extensions.verification.contract_accounting" {
				return fmt.Errorf("%s has unresolved verification ID %q", label, verificationID)
			}
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
		if err := requireAllowedKeys(declaration, stringSet("schema_id", "owner_contract_ref", "schema_surfaces", "procedural_surfaces"), label); err != nil {
			return err
		}
		if declaration["schema_id"] != "cartulary.extension_validation_surface_declaration.v1" {
			return fmt.Errorf("%s.schema_id is invalid", label)
		}
		if _, err := requiredString(declaration, "owner_contract_ref", label); err != nil {
			return err
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

func validateExtensionOwnerBindings(root string, indexed map[string]map[string]any) error {
	manifestByRef := map[string]map[string]any{}
	fragmentRefs := map[string]string{}
	for relativePath, object := range indexed {
		if object["schema_id"] != "cartulary.extension_owner_contract_manifest.v1" {
			continue
		}
		manifestRef := "contracts/extensions/" + relativePath
		manifestByRef[manifestRef] = object
		if err := validateExtensionManifestDocument(root, relativePath, object); err != nil {
			return err
		}
		ownerID, _ := requiredString(object, "owner_id", relativePath)
		fragments, ok := object["owner_fragments"].([]any)
		if !ok {
			return fmt.Errorf("%s.owner_fragments must be a present non-null array", relativePath)
		}
		for _, rawFragment := range fragments {
			fragment, ok := rawFragment.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.owner_fragments item must be an object", relativePath)
			}
			ref, _ := requiredString(fragment, "owner_fragment_ref", relativePath)
			digest, _ := requiredString(fragment, "owner_fragment_sha256", relativePath)
			if _, duplicate := fragmentRefs[ref]; duplicate {
				return fmt.Errorf("owner fragment %s is associated more than once", ref)
			}
			fragmentRefs[ref] = ownerID
			fragmentRelative := strings.TrimPrefix(ref, "contracts/extensions/")
			fragmentObject, ok := indexed[fragmentRelative]
			if !ok {
				return fmt.Errorf("owner manifest references missing fragment %s", ref)
			}
			if fragmentObject["owner_id"] != ownerID {
				return fmt.Errorf("owner fragment %s owner_id does not match manifest", ref)
			}
			actualDigest, err := extensionCanonicalDigest(fragmentObject)
			if err != nil {
				return err
			}
			if actualDigest != digest {
				return fmt.Errorf("owner fragment %s digest is stale", ref)
			}
		}
	}
	for relativePath, object := range indexed {
		if object["schema_id"] == "cartulary.extension_owner_fragment.v1" {
			ref := "contracts/extensions/" + relativePath
			if _, ok := fragmentRefs[ref]; !ok {
				return fmt.Errorf("owner fragment %s is not adopted by a manifest", ref)
			}
		}
	}
	dependencies := indexed["dependencies.json"]
	rows, _ := objectArray(dependencies["dependencies"], "dependencies")
	for _, dependency := range rows {
		manifestRef, _ := requiredString(dependency, "owner_contract_manifest_ref", "dependency")
		manifest, ok := manifestByRef[manifestRef]
		if !ok {
			return fmt.Errorf("dependency references missing manifest %s", manifestRef)
		}
		manifestDigest, err := extensionCanonicalDigest(manifest)
		if err != nil {
			return err
		}
		if dependency["owner_contract_manifest_sha256"] != manifestDigest {
			return fmt.Errorf("dependency manifest digest is stale for %s", dependency["dependency_id"])
		}
		document, _ := asObject(manifest["owner_document"], manifestRef+".owner_document")
		for _, key := range []string{"owner_document_ref", "owner_document_schema_id", "owner_document_version", "owner_document_sha256"} {
			if dependency[key] != document[key] {
				return fmt.Errorf("dependency %s does not match its owner manifest %s", dependency["dependency_id"], key)
			}
		}
		if dependency["owner_contract_manifest_id"] != manifest["owner_contract_manifest_id"] {
			return fmt.Errorf("dependency %s manifest ID mismatch", dependency["dependency_id"])
		}
		anchors, _ := objectArray(manifest["anchors"], manifestRef+".anchors")
		anchorSet := map[string]struct{}{}
		for _, anchor := range anchors {
			ownerRef, _ := requiredString(document, "owner_document_ref", manifestRef)
			pathPart := strings.SplitN(ownerRef, "#", 2)[0]
			anchorSet[pathPart+"#"+anchor["anchor_kind"].(string)+":"+anchor["anchor_id"].(string)] = struct{}{}
		}
		importedAnchors, _ := dependency["imported_anchor_refs"].([]any)
		for _, rawRef := range importedAnchors {
			ref, ok := rawRef.(string)
			if !ok {
				return fmt.Errorf("dependency imported anchor must be a string")
			}
			if _, ok := anchorSet[ref]; !ok {
				return fmt.Errorf("dependency imported anchor %s is absent from its manifest", ref)
			}
		}
	}
	if err := validateExtensionProfileFactClosure(indexed); err != nil {
		return err
	}
	return validateExtensionSupportingContractParity(indexed)
}

func validateExtensionSupportingContractParity(indexed map[string]map[string]any) error {
	participantDigests := map[string]string{}
	codecsByID := map[string]map[string]any{}
	statePresenceByProfile := map[string]map[string]any{}
	for _, object := range indexed {
		switch object["schema_id"] {
		case "cartulary.extension_participant_contract.v1", "cartulary.extension_transaction_participant_contract.v1":
			participantID := stringValue(object["participant_id"])
			digest, err := extensionCanonicalDigest(object)
			if err != nil {
				return err
			}
			if _, duplicate := participantDigests[participantID]; duplicate {
				return fmt.Errorf("duplicate participant contract %s", participantID)
			}
			participantDigests[participantID] = digest
		case "cartulary.extension_backup_binding_codec.v1":
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
		if object["schema_id"] != "cartulary.extension_owner_fragment.v1" {
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

func validateExtensionManifestDocument(root, relativePath string, manifest map[string]any) error {
	document, _ := asObject(manifest["owner_document"], relativePath+".owner_document")
	ownerRef, _ := requiredString(document, "owner_document_ref", relativePath)
	documentPath := strings.SplitN(ownerRef, "#", 2)[0]
	if !validRepositoryRelativePath(documentPath) {
		return fmt.Errorf("%s has invalid owner document path", relativePath)
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(documentPath)))
	if err != nil {
		return fmt.Errorf("read owner document for %s: %w", relativePath, err)
	}
	actualDigest := sha256.Sum256(data)
	if document["owner_document_sha256"] != hex.EncodeToString(actualDigest[:]) {
		return fmt.Errorf("%s owner document digest is stale", relativePath)
	}
	byteLength, _ := positiveJSONInt(document["byte_length"], relativePath+".byte_length")
	if byteLength != int64(len(data)) {
		return fmt.Errorf("%s owner document byte length is stale", relativePath)
	}
	anchors, _ := objectArray(manifest["anchors"], relativePath+".anchors")
	documentAnchors := 0
	seen := map[string]struct{}{}
	previousOrder := -1
	previousID := ""
	kindOrder := map[string]int{"document": 0, "req": 1, "table": 2, "schema": 3, "algorithm": 4}
	for index, anchor := range anchors {
		kind, err := requiredString(anchor, "anchor_kind", relativePath+".anchor")
		if err != nil {
			return err
		}
		order, ok := kindOrder[kind]
		if !ok {
			return fmt.Errorf("%s anchor kind is invalid", relativePath)
		}
		id, err := requiredString(anchor, "anchor_id", relativePath+".anchor")
		if err != nil {
			return err
		}
		if order < previousOrder || (order == previousOrder && previousID >= id) {
			return fmt.Errorf("%s anchors are not in canonical order", relativePath)
		}
		previousOrder, previousID = order, id
		identity := kind + "\x00" + id
		if _, duplicate := seen[identity]; duplicate {
			return fmt.Errorf("%s has duplicate anchor %s", relativePath, identity)
		}
		seen[identity] = struct{}{}
		start, err := nonnegativeJSONInt(anchor["start_byte"], relativePath+".anchor.start_byte")
		if err != nil {
			return err
		}
		end, err := positiveJSONInt(anchor["end_byte"], relativePath+".anchor.end_byte")
		if err != nil {
			return err
		}
		if start >= end || end > int64(len(data)) {
			return fmt.Errorf("%s anchor %d range is invalid", relativePath, index)
		}
		anchorDigest := sha256.Sum256(data[start:end])
		if anchor["anchor_sha256"] != hex.EncodeToString(anchorDigest[:]) {
			return fmt.Errorf("%s anchor %s digest is stale", relativePath, id)
		}
		if kind == "document" {
			documentAnchors++
			if start != 0 || end != int64(len(data)) || id != document["owner_document_schema_id"] {
				return fmt.Errorf("%s document anchor is not the exact whole document", relativePath)
			}
		}
	}
	if documentAnchors != 1 {
		return fmt.Errorf("%s must contain exactly one document anchor", relativePath)
	}
	return nil
}

func validateExtensionProfileFactClosure(indexed map[string]map[string]any) error {
	counts := map[string]map[string]int{}
	configDigests := map[string]string{}
	for _, profileID := range requiredExtensionProfiles {
		counts[profileID] = map[string]int{}
	}
	for relativePath, object := range indexed {
		schemaID, _ := object["schema_id"].(string)
		if schemaID == "cartulary.extension_profile_configuration_contract.v1" {
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
		if schemaID != "cartulary.extension_owner_fragment.v1" {
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
	}
	return nil
}

func validateExtensionTraceabilityRanges(root string, indexed map[string]map[string]any) error {
	object := indexed["traceability/mapping-source.json"]
	if object == nil {
		return fmt.Errorf("extension traceability mapping source is missing")
	}
	data, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	if err != nil {
		return err
	}
	expected, err := buildExtensionTraceabilityMappingSource(data)
	if err != nil {
		return err
	}
	expectedCanonical, err := canonicalizeDecoded(expected)
	if err != nil {
		return err
	}
	actualCanonical, err := canonicalizeDecoded(object)
	if err != nil {
		return err
	}
	if actualCanonical != expectedCanonical {
		return fmt.Errorf("extension traceability mapping source does not exactly cover the normative clause extraction")
	}
	return nil
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

func validRepositoryRelativePath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == path && path != "." && path != ".." && !strings.HasPrefix(path, "../")
}
