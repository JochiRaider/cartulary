package entities_test

import (
	"encoding/json"
	"slices"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestOpenAPIPhase4MutationContractShape(t *testing.T) {
	document := loadOpenAPIContract(t)
	schemas := objectAt(t, document, "components", "schemas")

	assertMentionResolveContract(t, document, schemas)
	assertMergeContract(t, document, schemas)
}

func assertMentionResolveContract(t *testing.T, document map[string]any, schemas map[string]any) {
	t.Helper()

	post := objectAt(t, document, "paths", "/api/v1/entity-mentions/{entity_mention_id}/resolve", "post")
	requestSchema := objectAt(t, post, "requestBody", "content", "application/json", "schema")
	if got := stringAt(t, requestSchema, "$ref"); got != "#/components/schemas/MentionActionRequest" {
		t.Fatalf("mention resolve request should use MentionActionRequest, got %q", got)
	}
	responseSchema := objectAt(t, post, "responses", "200", "content", "application/json", "schema")
	if got := stringAt(t, responseSchema, "$ref"); got != "#/components/schemas/MentionActionEnvelope" {
		t.Fatalf("mention resolve response should use MentionActionEnvelope, got %q", got)
	}

	request := schema(t, schemas, "MentionActionRequest")
	required := requiredFields(t, request)
	for _, field := range []string{"base_mention_row_version", "client_txn_id", "action"} {
		if !slices.Contains(required, field) {
			t.Fatalf("MentionActionRequest missing required field %q; got %v", field, required)
		}
	}
	if slices.Contains(required, "resolved_record_id") {
		t.Fatalf("MentionActionRequest should not require resolved_record_id for non-resolve actions")
	}
	actionEnum := stringArrayAt(t, objectAt(t, request, "properties", "action"), "enum")
	for _, action := range []string{"resolve_item", "dismiss_item", "revert_to_unresolved"} {
		if !slices.Contains(actionEnum, action) {
			t.Fatalf("MentionActionRequest action enum missing %q; got %v", action, actionEnum)
		}
	}
	if got := stringAt(t, objectAt(t, request, "properties", "resolved_record_id"), "format"); got != "uuid" {
		t.Fatalf("MentionActionRequest resolved_record_id should be uuid, got %q", got)
	}

	data := schema(t, schemas, "MentionActionData")
	dataRequired := requiredFields(t, data)
	for _, field := range []string{"entity_mention", "source_record", "change_set_id"} {
		if !slices.Contains(dataRequired, field) {
			t.Fatalf("MentionActionData missing required field %q; got %v", field, dataRequired)
		}
	}
}

func assertMergeContract(t *testing.T, document map[string]any, schemas map[string]any) {
	t.Helper()

	post := objectAt(t, document, "paths", "/api/v1/records/{survivor_record_id}/merge", "post")
	if got := stringAt(t, post, "operationId"); got != "mergeEntityRecord" {
		t.Fatalf("unexpected merge operationId: %q", got)
	}
	tags := stringArrayAt(t, post, "tags")
	if !slices.Contains(tags, "entities") {
		t.Fatalf("merge route should advertise entities tag, got %v", tags)
	}
	requestSchema := objectAt(t, post, "requestBody", "content", "application/json", "schema")
	if got := stringAt(t, requestSchema, "$ref"); got != "#/components/schemas/RecordMergeRequest" {
		t.Fatalf("merge request should use RecordMergeRequest, got %q", got)
	}
	responseSchema := objectAt(t, post, "responses", "200", "content", "application/json", "schema")
	if got := stringAt(t, responseSchema, "$ref"); got != "#/components/schemas/RecordMergeEnvelope" {
		t.Fatalf("merge response should use RecordMergeEnvelope, got %q", got)
	}

	request := schema(t, schemas, "RecordMergeRequest")
	required := requiredFields(t, request)
	for _, field := range []string{"loser_record_id", "survivor_base_row_version", "loser_base_row_version", "client_txn_id"} {
		if !slices.Contains(required, field) {
			t.Fatalf("RecordMergeRequest missing required field %q; got %v", field, required)
		}
	}
	if got := stringAt(t, objectAt(t, request, "properties", "loser_record_id"), "format"); got != "uuid" {
		t.Fatalf("RecordMergeRequest loser_record_id should be uuid, got %q", got)
	}

	data := schema(t, schemas, "RecordMergeData")
	dataRequired := requiredFields(t, data)
	for _, field := range []string{"incident_id", "survivor_record_id", "loser_record_id", "survivor_row_version", "loser_row_version", "change_set_id", "merged_into_record_id", "merge_summary"} {
		if !slices.Contains(dataRequired, field) {
			t.Fatalf("RecordMergeData missing required field %q; got %v", field, dataRequired)
		}
	}
}

func loadOpenAPIContract(t *testing.T) map[string]any {
	t.Helper()

	artifact, ok := gencontracts.OpenAPIArtifactsIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("generated OpenAPI artifact missing from internal/gen/contracts")
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI artifact JSON: %v", err)
	}
	return document
}

func schema(t *testing.T, schemas map[string]any, name string) map[string]any {
	t.Helper()

	raw, ok := schemas[name]
	if !ok {
		t.Fatalf("schema %q missing", name)
	}
	typed, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("schema %q is %T, want object", name, raw)
	}
	return typed
}

func objectAt(t *testing.T, root map[string]any, path ...string) map[string]any {
	t.Helper()

	current := any(root)
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("path %v: %q parent is %T, want object", path, key, current)
		}
		value, ok := object[key]
		if !ok {
			t.Fatalf("path %v missing key %q", path, key)
		}
		current = value
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("path %v is %T, want object", path, current)
	}
	return object
}

func stringAt(t *testing.T, root map[string]any, key string) string {
	t.Helper()

	value, ok := root[key].(string)
	if !ok {
		t.Fatalf("key %q is %T, want string", key, root[key])
	}
	return value
}

func requiredFields(t *testing.T, root map[string]any) []string {
	t.Helper()
	return stringArrayAt(t, root, "required")
}

func stringArrayAt(t *testing.T, root map[string]any, key string) []string {
	t.Helper()

	raw, ok := root[key].([]any)
	if !ok {
		t.Fatalf("key %q is %T, want array", key, root[key])
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("key %q item is %T, want string", key, item)
		}
		values = append(values, value)
	}
	return values
}
