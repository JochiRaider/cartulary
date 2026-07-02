package workbook_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestWorkbookOpenAPIRecordMutationContracts(t *testing.T) {
	document := loadWorkbookOpenAPIContract(t)
	schemas := workbookObjectAt(t, document, "components", "schemas")

	info := workbookObjectAt(t, document, "info")
	title := workbookStringAt(t, info, "title")
	if strings.Contains(title, "Phase 3") || strings.Contains(title, "Timeline Mutation") {
		t.Fatalf("OpenAPI title still advertises stale Phase 3 Timeline-only scope: %q", title)
	}
	version := workbookStringAt(t, info, "version")
	if strings.Contains(version, "phase3") {
		t.Fatalf("OpenAPI version still advertises Phase 3 scope: %q", version)
	}

	assertWorkbookQueryContract(t, document, schemas)
	assertViewRowCreateContract(t, document)

	patch := workbookObjectAt(t, document, "paths", "/api/v1/records/{record_id}", "patch")
	if got := workbookStringAt(t, patch, "operationId"); got != "patchRecord" {
		t.Fatalf("record patch operationId = %q, want patchRecord", got)
	}
	tags := workbookStringArrayAt(t, patch, "tags")
	if slices.Contains(tags, "timeline") {
		t.Fatalf("record patch tags should not advertise timeline-only ownership: %v", tags)
	}

	linkedNote := workbookObjectAt(t, document, "paths", "/api/v1/records/{record_id}/linked-notes", "post")
	if got := workbookStringAt(t, linkedNote, "operationId"); got != "createRecordLinkedNote" {
		t.Fatalf("linked-note operationId = %q, want createRecordLinkedNote", got)
	}
	requestSchema := workbookObjectAt(t, linkedNote, "requestBody", "content", "application/json", "schema")
	if got := workbookStringAt(t, requestSchema, "$ref"); got != "#/components/schemas/LinkedNoteCreateRequest" {
		t.Fatalf("linked-note request schema = %q, want LinkedNoteCreateRequest", got)
	}
	for _, status := range []string{"200", "201"} {
		responseSchema := workbookObjectAt(t, linkedNote, "responses", status, "content", "application/json", "schema")
		if got := workbookStringAt(t, responseSchema, "$ref"); got != "#/components/schemas/ViewMutationEnvelope" {
			t.Fatalf("linked-note response %s schema = %q, want ViewMutationEnvelope", status, got)
		}
	}

	assertRecordPatchChangeContract(t, schemas)
	assertCollectionActionsContract(t, schemas)

	request := workbookObjectAt(t, schemas, "LinkedNoteCreateRequest")
	required := workbookStringArrayAt(t, request, "required")
	for _, field := range []string{"client_txn_id"} {
		if !slices.Contains(required, field) {
			t.Fatalf("LinkedNoteCreateRequest missing required field %q; got %v", field, required)
		}
	}
}

func assertWorkbookQueryContract(t *testing.T, document map[string]any, schemas map[string]any) {
	t.Helper()

	post := workbookObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/query", "post")
	if got := workbookStringAt(t, post, "operationId"); got != "queryWorkbookView" {
		t.Fatalf("unexpected workbook query operationId: %q", got)
	}

	requestSchema := workbookObjectAt(t, post, "requestBody", "content", "application/json", "schema")
	if got := workbookStringAt(t, requestSchema, "$ref"); got != "#/components/schemas/WorkbookQueryRequest" {
		t.Fatalf("workbook query request should use WorkbookQueryRequest, got %q", got)
	}
	responseSchema := workbookObjectAt(t, post, "responses", "200", "content", "application/json", "schema")
	if got := workbookStringAt(t, responseSchema, "$ref"); got != "#/components/schemas/WorkbookQueryEnvelope" {
		t.Fatalf("workbook query response should use WorkbookQueryEnvelope, got %q", got)
	}

	queryData := workbookSchema(t, schemas, "WorkbookQueryData")
	viewSchemaID := workbookObjectAt(t, queryData, "properties", "view_schema_id")
	if _, ok := viewSchemaID["const"]; ok {
		t.Fatal("WorkbookQueryData.view_schema_id must not be constrained to the Timeline schema")
	}
	if got := workbookStringAt(t, viewSchemaID, "type"); got != "string" {
		t.Fatalf("WorkbookQueryData.view_schema_id type = %q, want string", got)
	}

	for _, staleSchema := range []string{"TimelineQueryRequest", "TimelineQueryData", "TimelineQueryEnvelope"} {
		if _, ok := schemas[staleSchema]; ok {
			t.Fatalf("stale Timeline-only query schema %q should not be present", staleSchema)
		}
	}
}

func assertViewRowCreateContract(t *testing.T, document map[string]any) {
	t.Helper()

	post := workbookObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", "post")
	if got := workbookStringAt(t, post, "operationId"); got != "createViewRow" {
		t.Fatalf("unexpected row-create operationId: %q", got)
	}

	requestSchema := workbookObjectAt(t, post, "requestBody", "content", "application/json", "schema")
	refs := workbookRefsFromOneOf(t, requestSchema)
	for _, ref := range []string{
		"#/components/schemas/TimelineCreateRequest",
		"#/components/schemas/HostCreateRequest",
		"#/components/schemas/IdentityCreateRequest",
		"#/components/schemas/IndicatorCreateRequest",
		"#/components/schemas/AssessmentCreateRequest",
		"#/components/schemas/EvidenceCreateRequest",
		"#/components/schemas/NoteCreateRequest",
		"#/components/schemas/TaskRequestCreateRequest",
		"#/components/schemas/DecisionCreateRequest",
		"#/components/schemas/PartyCreateRequest",
		"#/components/schemas/CommLogCreateRequest",
		"#/components/schemas/HandoffCreateRequest",
		"#/components/schemas/StatusReviewCreateRequest",
		"#/components/schemas/LessonCreateRequest",
	} {
		if !slices.Contains(refs, ref) {
			t.Fatalf("row-create request schema missing %s; got %v", ref, refs)
		}
	}

	for _, status := range []string{"200", "201"} {
		responseSchema := workbookObjectAt(t, post, "responses", status, "content", "application/json", "schema")
		if got := workbookStringAt(t, responseSchema, "$ref"); got != "#/components/schemas/ViewMutationEnvelope" {
			t.Fatalf("row-create response %s should use ViewMutationEnvelope, got %q", status, got)
		}
	}
}

func assertRecordPatchChangeContract(t *testing.T, schemas map[string]any) {
	t.Helper()

	change := workbookSchema(t, schemas, "RecordPatchChange")
	properties := workbookObjectAt(t, change, "properties")
	if _, ok := properties["action_payload"]; !ok {
		t.Fatal("RecordPatchChange must expose action_payload")
	}
	if got := workbookStringAt(t, workbookObjectAt(t, properties, "action_payload"), "$ref"); got != "#/components/schemas/CollectionActionsV1" {
		t.Fatalf("RecordPatchChange action_payload should use CollectionActionsV1, got %q", got)
	}
	required := workbookStringArrayAt(t, change, "required")
	if !slices.Equal(required, []string{"field_key"}) {
		t.Fatalf("RecordPatchChange should require only field_key before exactly-one validation, got %v", required)
	}
	if _, ok := change["oneOf"].([]any); !ok {
		t.Fatal("RecordPatchChange must enforce exactly one of value or action_payload with oneOf")
	}
}

func assertCollectionActionsContract(t *testing.T, schemas map[string]any) {
	t.Helper()

	collection := workbookSchema(t, schemas, "CollectionActionsV1")
	if got := workbookStringAt(t, workbookObjectAt(t, collection, "properties", "kind"), "const"); got != "collection_actions_v1" {
		t.Fatalf("CollectionActionsV1 kind const mismatch: %q", got)
	}
	actions := workbookObjectAt(t, collection, "properties", "actions")
	if got := workbookIntAt(t, actions, "minItems"); got != 1 {
		t.Fatalf("CollectionActionsV1 actions minItems = %d, want 1", got)
	}
	if got := workbookIntAt(t, actions, "maxItems"); got != 64 {
		t.Fatalf("CollectionActionsV1 actions maxItems = %d, want 64", got)
	}

	refs := workbookRefsFromOneOf(t, workbookObjectAt(t, actions, "items"))
	actionSchemas := map[string]string{
		"CollectionAddTokenAction":        "add_token",
		"CollectionAddAliasAction":        "add_alias",
		"CollectionRemoveAliasAction":     "remove_alias",
		"CollectionAddTagAction":          "add_tag",
		"CollectionRemoveTagAction":       "remove_tag",
		"CollectionAddRecordRefAction":    "add_record_ref",
		"CollectionRemoveRecordRefAction": "remove_record_ref",
		"CollectionAddPartyRefAction":     "add_party_ref",
		"CollectionRemovePartyRefAction":  "remove_party_ref",
		"CollectionAddRiskRefAction":      "add_risk_ref",
		"CollectionRemoveRiskRefAction":   "remove_risk_ref",
	}
	for name, op := range actionSchemas {
		ref := "#/components/schemas/" + name
		if !slices.Contains(refs, ref) {
			t.Fatalf("CollectionActionsV1 missing action schema %s; got %v", ref, refs)
		}
		action := workbookSchema(t, schemas, name)
		if got := workbookStringAt(t, workbookObjectAt(t, action, "properties", "op"), "const"); got != op {
			t.Fatalf("%s op const = %q, want %q", name, got, op)
		}
	}
}

func loadWorkbookOpenAPIContract(t *testing.T) map[string]any {
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

func workbookSchema(t *testing.T, schemas map[string]any, name string) map[string]any {
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

func workbookObjectAt(t *testing.T, root map[string]any, path ...string) map[string]any {
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

func workbookStringAt(t *testing.T, root map[string]any, key string) string {
	t.Helper()

	value, ok := root[key].(string)
	if !ok {
		t.Fatalf("key %q is %T, want string", key, root[key])
	}
	return value
}

func workbookIntAt(t *testing.T, root map[string]any, key string) int {
	t.Helper()

	value, ok := root[key].(float64)
	if !ok {
		t.Fatalf("key %q is %T, want number", key, root[key])
	}
	return int(value)
}

func workbookStringArrayAt(t *testing.T, root map[string]any, key string) []string {
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

func workbookRefsFromOneOf(t *testing.T, root map[string]any) []string {
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
		refs = append(refs, workbookStringAt(t, object, "$ref"))
	}
	return refs
}
