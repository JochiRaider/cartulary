package reference_data_test

import (
	"encoding/json"
	"reflect"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestPhase11_U_11_REFERENCE_PACK_05_OpenAPIAndErrorRegistriesExposeClosedReferencePackContract(t *testing.T) {
	artifact, ok := gencontracts.OpenAPIArtifactsIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("generated OpenAPI artifact missing")
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI artifact: %v", err)
	}
	schemas := openAPIObjectAt(t, document, "components", "schemas")
	requireEnum(t, openAPIObjectAt(t, schemas, "ReferencePackVersionState"), []string{"staged", "verified_available", "disabled", "failed", "missing"})
	requireEnum(t, openAPIObjectAt(t, schemas, "ReferencePackVerificationResult"), []string{"pending", "passed", "failed"})
	resource := openAPIObjectAt(t, schemas, "ReferencePackVersionResource")
	requireClosedObject(t, resource, "ReferencePackVersionResource")
	requireRequired(t, resource, []string{
		"pack_key",
		"pack_kind",
		"pack_version",
		"pack_version_state",
		"active",
		"source_identifier",
		"manifest_sha256",
		"payload_sha256",
		"pack_contract_version",
		"verification_method",
		"verification_result",
		"signer_key_id",
		"previous_active_version",
		"imported_by_user_id",
		"imported_at",
		"activated_by_user_id",
		"activated_at",
	})
	paths := openAPIObjectAt(t, document, "paths")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs", "get"), "200", "ReferencePackListEnvelope")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}", "get"), "200", "ReferencePackVersionEnvelope")
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/import", "post"), "202", "JobEnvelope")
	for _, action := range []string{"activate", "disable"} {
		requireRequestRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/"+action, "post"), "ReferencePackActionRequest")
		requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/"+action, "post"), "200", "ReferencePackActionEnvelope")
	}
	requireResponseRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/{pack_key}/{pack_version}/reverify", "post"), "202", "JobEnvelope")
	requireRequestRef(t, openAPIObjectAt(t, paths, "/api/v1/reference-packs/refresh", "post"), "ReferencePackRefreshRequest")

	errorArtifact, ok := gencontracts.ErrorArtifactsIndex["contracts/errors/index.json"]
	if !ok {
		t.Fatal("generated error artifact missing")
	}
	var errorsDoc map[string]any
	if err := json.Unmarshal([]byte(errorArtifact.JSON), &errorsDoc); err != nil {
		t.Fatalf("decode generated error artifact: %v", err)
	}
	requireReasonRegistry(t, errorsDoc, "invalid_reference_pack_request", []string{
		"unsupported_upload_envelope", "missing_required_part", "duplicate_part", "unexpected_part", "invalid_part_content_type",
		"invalid_metadata_encoding", "malformed_metadata_json", "request_not_object", "missing_required_field", "field_not_nullable",
		"unknown_field", "invalid_activation_policy", "pack_version_required", "auto_activation_not_supported", "invalid_pack_keys", "empty_pack_keys",
	})
	requireReasonRegistry(t, errorsDoc, "reference_pack_verification_failed", []string{
		"checksum_mismatch", "signature_mismatch", "missing_integrity_metadata", "contract_incompatible", "path_traversal",
		"disallowed_content", "payload_missing", "archive_extracted_bytes_exceeded", "archive_compression_ratio_exceeded", "archive_member_count_exceeded",
	})
	requireReasonRegistry(t, errorsDoc, "reference_pack_activation_rejected", []string{"already_active", "not_verified_available"})
	requireReasonRegistry(t, errorsDoc, "reference_pack_state_conflict", []string{"already_disabled", "not_disableable", "verification_pending"})
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
