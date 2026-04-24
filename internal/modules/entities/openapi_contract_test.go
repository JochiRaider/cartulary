package entities_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestOpenAPIPhase4MutationContractShape(t *testing.T) {
	document := loadOpenAPIContract(t)
	schemas := objectAt(t, document, "components", "schemas")

	info := objectAt(t, document, "info")
	title := stringAt(t, info, "title")
	if strings.Contains(title, "Phase 3") || strings.Contains(title, "Timeline Mutation") {
		t.Fatalf("OpenAPI title still advertises stale Phase 3 Timeline-only scope: %q", title)
	}
	version := stringAt(t, info, "version")
	if strings.Contains(version, "phase3") {
		t.Fatalf("OpenAPI version still advertises Phase 3 scope: %q", version)
	}

	assertViewRowCreateContract(t, document)
	assertMentionResolveContract(t, document, schemas)
	assertRecordPatchChangeContract(t, schemas)
	assertCollectionActionsContract(t, schemas)
}

func assertViewRowCreateContract(t *testing.T, document map[string]any) {
	t.Helper()

	post := objectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", "post")
	if got := stringAt(t, post, "operationId"); got != "createViewRow" {
		t.Fatalf("unexpected row-create operationId: %q", got)
	}

	requestSchema := objectAt(t, post, "requestBody", "content", "application/json", "schema")
	refs := refsFromOneOf(t, requestSchema)
	for _, ref := range []string{
		"#/components/schemas/TimelineCreateRequest",
		"#/components/schemas/HostCreateRequest",
		"#/components/schemas/IdentityCreateRequest",
		"#/components/schemas/IndicatorCreateRequest",
	} {
		if !slices.Contains(refs, ref) {
			t.Fatalf("row-create request schema missing %s; got %v", ref, refs)
		}
	}

	for _, status := range []string{"200", "201"} {
		responseSchema := objectAt(t, post, "responses", status, "content", "application/json", "schema")
		if got := stringAt(t, responseSchema, "$ref"); got != "#/components/schemas/ViewMutationEnvelope" {
			t.Fatalf("row-create response %s should use ViewMutationEnvelope, got %q", status, got)
		}
	}
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

func assertRecordPatchChangeContract(t *testing.T, schemas map[string]any) {
	t.Helper()

	change := schema(t, schemas, "RecordPatchChange")
	properties := objectAt(t, change, "properties")
	if _, ok := properties["action_payload"]; !ok {
		t.Fatal("RecordPatchChange must expose action_payload")
	}
	if got := stringAt(t, objectAt(t, properties, "action_payload"), "$ref"); got != "#/components/schemas/CollectionActionsV1" {
		t.Fatalf("RecordPatchChange action_payload should use CollectionActionsV1, got %q", got)
	}
	required := requiredFields(t, change)
	if !slices.Equal(required, []string{"field_key"}) {
		t.Fatalf("RecordPatchChange should require only field_key before exactly-one validation, got %v", required)
	}
	if _, ok := change["oneOf"].([]any); !ok {
		t.Fatal("RecordPatchChange must enforce exactly one of value or action_payload with oneOf")
	}
}

func assertCollectionActionsContract(t *testing.T, schemas map[string]any) {
	t.Helper()

	collection := schema(t, schemas, "CollectionActionsV1")
	if got := stringAt(t, objectAt(t, collection, "properties", "kind"), "const"); got != "collection_actions_v1" {
		t.Fatalf("CollectionActionsV1 kind const mismatch: %q", got)
	}
	actions := objectAt(t, collection, "properties", "actions")
	if got := intAt(t, actions, "minItems"); got != 1 {
		t.Fatalf("CollectionActionsV1 actions minItems = %d, want 1", got)
	}
	if got := intAt(t, actions, "maxItems"); got != 64 {
		t.Fatalf("CollectionActionsV1 actions maxItems = %d, want 64", got)
	}

	refs := refsFromOneOf(t, objectAt(t, actions, "items"))
	actionSchemas := map[string]string{
		"CollectionAddTokenAction":           "add_token",
		"CollectionAddResolvedRefAction":     "add_resolved_ref",
		"CollectionResolveItemAction":        "resolve_item",
		"CollectionDismissItemAction":        "dismiss_item",
		"CollectionRevertToUnresolvedAction": "revert_to_unresolved",
	}
	for name, op := range actionSchemas {
		ref := "#/components/schemas/" + name
		if !slices.Contains(refs, ref) {
			t.Fatalf("CollectionActionsV1 missing action schema %s; got %v", ref, refs)
		}
		action := schema(t, schemas, name)
		if got := stringAt(t, objectAt(t, action, "properties", "op"), "const"); got != op {
			t.Fatalf("%s op const = %q, want %q", name, got, op)
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

func intAt(t *testing.T, root map[string]any, key string) int {
	t.Helper()

	value, ok := root[key].(float64)
	if !ok {
		t.Fatalf("key %q is %T, want number", key, root[key])
	}
	return int(value)
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

func refsFromOneOf(t *testing.T, root map[string]any) []string {
	t.Helper()

	raw, ok := root["oneOf"].([]any)
	if !ok {
		t.Fatalf("oneOf is %T, want array", root["oneOf"])
	}
	refs := make([]string, 0, len(raw))
	for _, item := range raw {
		object, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("oneOf item is %T, want object", item)
		}
		refs = append(refs, stringAt(t, object, "$ref"))
	}
	return refs
}
