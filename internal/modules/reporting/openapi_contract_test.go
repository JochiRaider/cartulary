package reporting_test

import (
	"encoding/json"
	"reflect"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestPhase11_U_11_REPORTING_06_OpenAPIReleaseEnumsAndExactResources(t *testing.T) {
	artifact, ok := gencontracts.OpenAPIArtifactsIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("generated OpenAPI artifact missing")
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI artifact: %v", err)
	}
	schemas := reportingOpenAPIObjectAt(t, document, "components", "schemas")

	outputKinds := []string{"slidev", "mermaid"}
	releaseScopes := []string{"internal_draft", "internal_review", "external_release"}
	requireOpenAPIEnumSchema(t, reportingOpenAPIObjectAt(t, schemas, "ReleaseOutputKind"), outputKinds)
	requireOpenAPIEnumSchema(t, reportingOpenAPIObjectAt(t, schemas, "ReleaseScope"), releaseScopes)

	createRequest := reportingOpenAPIObjectAt(t, schemas, "ReleaseCreateRequest")
	requireClosedOpenAPIObject(t, createRequest, "ReleaseCreateRequest")
	createProperties := reportingOpenAPIObjectAt(t, createRequest, "properties")
	requireOpenAPIPropertyRef(t, createProperties, "output_kind", "ReleaseOutputKind")
	requireOpenAPIPropertyRef(t, createProperties, "release_scope", "ReleaseScope")

	snapshotResource := reportingOpenAPIObjectAt(t, schemas, "SnapshotResource")
	requireClosedOpenAPIObject(t, snapshotResource, "SnapshotResource")
	requireOpenAPIRequired(t, snapshotResource, []string{
		"snapshot_id",
		"incident_id",
		"created_by_user_id",
		"created_at",
		"snapshot_at",
		"source_change_set_high_watermark",
		"derivation_version",
		"export_model_sha256",
	})

	releaseResource := reportingOpenAPIObjectAt(t, schemas, "ReleaseResource")
	requireClosedOpenAPIObject(t, releaseResource, "ReleaseResource")
	requireOpenAPIRequired(t, releaseResource, []string{
		"release_id",
		"incident_id",
		"snapshot_id",
		"snapshot_at",
		"source_change_set_high_watermark",
		"derivation_version",
		"export_model_sha256",
		"template_id",
		"template_version",
		"redaction_profile_id",
		"redaction_profile_version",
		"redaction_profile_sha256",
		"output_kind",
		"output_media_type",
		"release_scope",
		"recipient_partition_refs",
		"output_options",
		"graph_projection_refs",
		"composition_id",
		"composition_version",
		"composition_sha256",
		"render_admitted_at",
		"output_sha256",
		"redaction_manifest_sha256",
		"release_state",
		"render_failed_reason_code",
		"created_by_user_id",
		"created_at",
		"approved_at",
		"invalidated_at",
		"published_at",
		"invalidation_reason",
	})
	releaseProperties := reportingOpenAPIObjectAt(t, releaseResource, "properties")
	requireOpenAPIPropertyRef(t, releaseProperties, "output_kind", "ReleaseOutputKind")
	requireOpenAPIPropertyRef(t, releaseProperties, "release_scope", "ReleaseScope")

	paths := reportingOpenAPIObjectAt(t, document, "paths")
	getSnapshot := reportingOpenAPIObjectAt(t, paths, "/api/v1/snapshots/{snapshot_id}", "get")
	requireOpenAPIResponseRef(t, getSnapshot, "200", "SnapshotEnvelope")
	requireOpenAPIResponseRef(t, getSnapshot, "400", "ErrorEnvelope")
	createRelease := reportingOpenAPIObjectAt(t, paths, "/api/v1/releases", "post")
	requireOpenAPIRequestRef(t, createRelease, "ReleaseCreateRequest")
	requireOpenAPIResponseRef(t, createRelease, "202", "JobEnvelope")
	getRelease := reportingOpenAPIObjectAt(t, paths, "/api/v1/releases/{release_id}", "get")
	requireOpenAPIResponseRef(t, getRelease, "200", "ReleaseEnvelope")
	requireOpenAPIResponseRef(t, getRelease, "400", "ErrorEnvelope")
	for _, action := range []string{"approve", "publish", "invalidate"} {
		operation := reportingOpenAPIObjectAt(t, paths, "/api/v1/releases/{release_id}/"+action, "post")
		requireOpenAPIRequestRef(t, operation, "ReleaseActionRequest")
		requireOpenAPIResponseRef(t, operation, "200", "ReleaseEnvelope")
	}
}

func requireOpenAPIEnumSchema(t testing.TB, schema map[string]any, want []string) {
	t.Helper()
	if schema["type"] != "string" {
		t.Fatalf("enum schema must be string, got %#v", schema)
	}
	if got := reportingOpenAPIStringArray(t, schema["enum"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("enum schema mismatch: got %v want %v", got, want)
	}
}

func requireClosedOpenAPIObject(t testing.TB, schema map[string]any, name string) {
	t.Helper()
	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("%s must be a closed object schema, got %#v", name, schema)
	}
}

func requireOpenAPIRequired(t testing.TB, schema map[string]any, want []string) {
	t.Helper()
	if got := reportingOpenAPIStringArray(t, schema["required"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("required members mismatch: got %v want %v", got, want)
	}
}

func requireOpenAPIPropertyRef(t testing.TB, properties map[string]any, field string, wantSchema string) {
	t.Helper()
	property := reportingOpenAPIObjectAt(t, properties, field)
	wantRef := "#/components/schemas/" + wantSchema
	if property["$ref"] != wantRef {
		t.Fatalf("property %s ref mismatch: got %#v want %q", field, property, wantRef)
	}
}

func requireOpenAPIRequestRef(t testing.TB, operation map[string]any, wantSchema string) {
	t.Helper()
	schema := reportingOpenAPIObjectAt(t, operation, "requestBody", "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchema
	if schema["$ref"] != wantRef {
		t.Fatalf("request schema ref mismatch: got %#v want %q", schema, wantRef)
	}
}

func requireOpenAPIResponseRef(t testing.TB, operation map[string]any, status string, wantSchema string) {
	t.Helper()
	schema := reportingOpenAPIObjectAt(t, operation, "responses", status, "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchema
	if schema["$ref"] != wantRef {
		t.Fatalf("response %s schema ref mismatch: got %#v want %q", status, schema, wantRef)
	}
}

func reportingOpenAPIObjectAt(t testing.TB, root any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("path %v: parent for %q is %T, want object", path, segment, current)
		}
		value, ok := object[segment]
		if !ok {
			t.Fatalf("path %v missing segment %q", path, segment)
		}
		current = value
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("path %v is %T, want object", path, current)
	}
	return object
}

func reportingOpenAPIStringArray(t testing.TB, raw any) []string {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("value is %T, want array", raw)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		typed, ok := value.(string)
		if !ok {
			t.Fatalf("array member is %T, want string", value)
		}
		out = append(out, typed)
	}
	return out
}
