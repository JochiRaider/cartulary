package reference_data_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestOpenAPIAndErrorRegistriesExposeClosedReferencePackContract_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	schemas := openAPIObjectAt(t, document, "components", "schemas")
	requireEnum(t, openAPIObjectAt(t, schemas, "ReferencePackVersionState"), []string{"staged", "verified_available", "disabled", "failed", "missing"})
	resource := openAPIObjectAt(t, schemas, "ReferencePackVersionResource")
	requireClosedObject(t, resource, "ReferencePackVersionResource")
	var typed map[string]any
	for _, artifact := range contractreferencepacks.Artifacts {
		if artifact.Path == "contracts/reference-packs/administrative_version.v2.schema.json" {
			if err := json.Unmarshal([]byte(artifact.JSON), &typed); err != nil {
				t.Fatal(err)
			}
		}
	}
	if typed == nil {
		t.Fatal("missing administrative version projection")
	}
	delete(typed, "$id")
	delete(typed, "$schema")
	openAPIObjectAt(t, typed, "properties")["pack_version_state"] = map[string]any{"$ref": "#/components/schemas/ReferencePackVersionState"}
	if !reflect.DeepEqual(resource, typed) {
		t.Fatal("administrative version and public OpenAPI projections disagree")
	}
	requireRequired(t, resource, []string{"pack_key", "pack_kind", "pack_version", "pack_version_state", "health", "administratively_disabled", "distribution_kind", "active", "removed", "missing_reason", "last_failure_code", "source_identifier", "manifest_sha256", "payload_sha256", "pack_contract_version", "content_profile_id", "content_profile_version", "verification_method", "last_verified_at", "trust_valid_until", "verified_signer_key_ids", "imported_by_user_id", "imported_at", "previous_active_version", "activated_by_user_id", "activated_at", "pack_release_sequence", "source_profile_id", "source_profile_sha256", "source_version", "source_as_of", "license_expression", "redistribution", "trust_repository_id", "pending_work", "reproducibility_pinned", "fallback_from_version", "dependencies"})
	paths := openAPIObjectAt(t, document, "paths")
	listOperation := openAPIObjectAt(t, paths, "/api/v1/reference-packs", "get")
	requireResponseRef(t, listOperation, "200", "ReferencePackListEnvelope")
	requireParameterNames(t, listOperation, []string{"limit", "cursor_token", "search", "pack_version_state", "active"})
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}", "get"), "200", "ReferencePackVersionEnvelope")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/validation-summaries/{summary_id}", "get"), "200", "ReferencePackValidationSummaryEnvelope")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/import", "post"), "202", "JobEnvelope")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/import", "post"), "409", "ReferencePackImportConflictEnvelope")
	uploadDetails := openAPIObjectAt(t, schemas, "ReferencePackUploadLimitDetails")
	requireClosedObject(t, uploadDetails, "ReferencePackUploadLimitDetails")
	requireRequired(t, uploadDetails, []string{"reason_code", "check_id", "validation_summary"})
	if openAPIObjectAt(t, uploadDetails, "properties", "reason_code")["const"] != "container_bytes_exceeded" || openAPIObjectAt(t, uploadDetails, "properties", "check_id")["const"] != "container_bytes" || openAPIObjectAt(t, uploadDetails, "properties", "validation_summary")["$ref"] != "#/components/schemas/ReferencePackValidationSummary" {
		t.Fatal("early admission diagnostics diverge from the owner contract")
	}
	for _, action := range []string{"activate", "disable"} {
		requireRequestRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/"+action, "post"), "ReferencePackActionRequest")
		requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/"+action, "post"), "200", "ReferencePackActionEnvelope")
	}
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/reverify", "post"), "202", "JobEnvelope")
	requireRequestRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/refresh", "post"), "ReferencePackRefreshRequest")

	errorsDoc := contracttest.ErrorRegistryDocument(t)
	requireReasonRegistry(t, errorsDoc, "invalid_reference_pack_request", []string{
		"unsupported_upload_envelope", "missing_required_part", "duplicate_part", "unexpected_part", "invalid_part_content_type",
		"invalid_metadata_encoding", "malformed_metadata_json", "request_not_object", "missing_required_field", "field_not_nullable",
		"unknown_field", "invalid_activation_policy", "pack_version_required", "auto_activation_not_supported", "invalid_pack_keys", "empty_pack_keys", "invalid_reason", "reason_too_long", "request_too_large",
	})
	requireReasonRegistry(t, errorsDoc, "reference_pack_verification_failed", []string{
		"archive_compression_ratio_exceeded",
		"archive_extracted_bytes_exceeded",
		"archive_member_count_exceeded",
		"archive_structure_invalid",
		"bundle_hint_invalid",
		"bundle_hint_noncanonical",
		"checksum_mismatch",
		"container_bytes_exceeded",
		"content_schema_invalid",
		"content_semantic_invalid",
		"contract_incompatible",
		"dependency_cycle",
		"dependency_unsatisfied",
		"disallowed_content",
		"disallowed_member_type",
		"duplicate_object_member",
		"manifest_encoding_invalid",
		"manifest_json_invalid",
		"manifest_noncanonical",
		"manifest_schema_invalid",
		"metadata_expired",
		"metadata_expiry_policy_invalid",
		"metadata_mix_and_match_detected",
		"metadata_noncanonical",
		"metadata_rollback_detected",
		"missing_integrity_metadata",
		"pack_conflict",
		"pack_release_sequence_collision",
		"pack_release_sequence_rollback",
		"pack_version_collision",
		"path_collision",
		"path_traversal",
		"payload_missing",
		"required_member_missing",
		"signature_threshold_not_met",
		"target_length_mismatch",
		"target_not_declared",
		"tuf_metadata_invalid",
		"tuf_root_rotation_invalid",
		"tuf_root_untrusted",
		"type_registry_incompatible",
		"undeclared_member",
		"unexpected_target",
		"unsupported_container_format",
		"verification_timeout",
	})
	requireReasonRegistry(t, errorsDoc, "reference_pack_activation_rejected", []string{"already_active", "not_verified_available", "metadata_expired"})
	requireReasonRegistry(t, errorsDoc, "reference_pack_operation_rejected", []string{"verification_pending", "no_successful_verification", "stale_admission_state", "clock_untrusted", "packaged_builtin", "removed", "not_disableable", "active", "pinned", "dependency_unsatisfied", "dependency_cycle", "pack_conflict", "required_registry_gap", "contract_incompatible", "type_registry_incompatible"})
	requireRequestRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/remove", "post"), "ReferencePackRemovalRequest")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/remove", "post"), "200", "ReferencePackActionEnvelope")
}

func requireParameterNames(t testing.TB, operation map[string]any, want []string) {
	t.Helper()
	parameters, ok := operation["parameters"].([]any)
	if !ok {
		t.Fatalf("operation parameters are %T, want array", operation["parameters"])
	}
	got := make([]string, 0, len(parameters))
	for _, rawParameter := range parameters {
		parameter, ok := rawParameter.(map[string]any)
		if !ok {
			t.Fatalf("operation parameter is %T, want object", rawParameter)
		}
		name, _ := parameter["name"].(string)
		got = append(got, name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("operation parameter names mismatch: got %v want %v", got, want)
	}
}

func requireEnum(t testing.TB, schema map[string]any, want []string) {
	t.Helper()
	if schema["type"] != "string" {
		t.Fatalf("enum schema must be string: %#v", schema)
	}
	if got := stringArray(schema["enum"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("enum mismatch: got %v want %v", got, want)
	}
}

func requireClosedObject(t testing.TB, schema map[string]any, name string) {
	t.Helper()
	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("%s must be a closed object schema: %#v", name, schema)
	}
}

func requireRequired(t testing.TB, schema map[string]any, want []string) {
	t.Helper()
	if got := stringArray(schema["required"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("required members mismatch: got %v want %v", got, want)
	}
}

func requireRequestRef(t testing.TB, operation map[string]any, wantSchema string) {
	t.Helper()
	schema := openAPIObjectAt(t, operation, "requestBody", "content", "application/json", "schema")
	want := "#/components/schemas/" + wantSchema
	if schema["$ref"] != want {
		t.Fatalf("request ref mismatch: got %#v want %q", schema, want)
	}
}

func requireResponseRef(t testing.TB, operation map[string]any, status string, wantSchema string) {
	t.Helper()
	schema := openAPIObjectAt(t, operation, "responses", status, "content", "application/json", "schema")
	want := "#/components/schemas/" + wantSchema
	if schema["$ref"] != want {
		t.Fatalf("response %s ref mismatch: got %#v want %q", status, schema, want)
	}
}

func requireReasonRegistry(t testing.TB, document map[string]any, errorCode string, want []string) {
	t.Helper()
	registries := document["reason_registries"].([]any)
	for _, raw := range registries {
		registry := raw.(map[string]any)
		if registry["error_code"] != errorCode {
			continue
		}
		var got []string
		for _, rawReason := range registry["reason_codes"].([]any) {
			got = append(got, rawReason.(map[string]any)["code"].(string))
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("reason registry %s mismatch: got %v want %v", errorCode, got, want)
		}
		return
	}
	t.Fatalf("missing reason registry %s", errorCode)
}

func openAPIObjectAt(t testing.TB, root any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("path %v parent for %q is %T", path, segment, current)
		}
		value, ok := object[segment]
		if !ok {
			t.Fatalf("path %v missing %q", path, segment)
		}
		current = value
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("path %v is %T, want object", path, current)
	}
	return object
}

func stringArray(raw any) []string {
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, _ := value.(string)
		out = append(out, text)
	}
	return out
}
