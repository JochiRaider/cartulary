package workbook_test

import (
	"encoding/json"
	"slices"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestWorkbookOpenAPIRecordMutationContracts(t *testing.T) {
	document := loadWorkbookOpenAPIContract(t)

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

	schemas := workbookObjectAt(t, document, "components", "schemas")
	request := workbookObjectAt(t, schemas, "LinkedNoteCreateRequest")
	required := workbookStringArrayAt(t, request, "required")
	for _, field := range []string{"client_txn_id"} {
		if !slices.Contains(required, field) {
			t.Fatalf("LinkedNoteCreateRequest missing required field %q; got %v", field, required)
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
