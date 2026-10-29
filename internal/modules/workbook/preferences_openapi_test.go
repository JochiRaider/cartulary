package workbook_test

import (
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"testing"
)

func TestOpenAPIWorkbookPreferencesExposeGetAndPutContracts(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)

	defaultPath := openAPIObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/workbook-preferences/default")
	requireOpenAPIOperation(t, defaultPath, "get", "getIncidentDefaultWorkbookPreferences")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, defaultPath, "get"), "DefaultWorkbookPreferencesEnvelope")
	requireOpenAPIOperation(t, defaultPath, "put", "putIncidentDefaultWorkbookPreferences")
	requireOpenAPIRequestSchemaRef(t, openAPIObjectAt(t, defaultPath, "put"), "DefaultWorkbookPreferencesPutRequest")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, defaultPath, "put"), "DefaultWorkbookPreferencesEnvelope")

	userPath := openAPIObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/workbook-preferences/me")
	requireOpenAPIOperation(t, userPath, "get", "getCurrentUserWorkbookPreferences")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, userPath, "get"), "UserWorkbookPreferencesEnvelope")
	requireOpenAPIOperation(t, userPath, "put", "putCurrentUserWorkbookPreferences")
	requireOpenAPIRequestSchemaRef(t, openAPIObjectAt(t, userPath, "put"), "UserWorkbookPreferencesPutRequest")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, userPath, "put"), "UserWorkbookPreferencesEnvelope")

	schemas := openAPIObjectAt(t, document, "components", "schemas")
	requireOpenAPISheetRefSchema(t, schemas)
	requireOpenAPIWorkbookPreferencesPutSchema(t, openAPIObjectAt(t, schemas, "DefaultWorkbookPreferencesPutRequest"), "default_sheet_ref")
	requireOpenAPIWorkbookPreferencesPutSchema(t, openAPIObjectAt(t, schemas, "UserWorkbookPreferencesPutRequest"), "home_sheet_ref")
	requireOpenAPIEnvelopeSchema(t, openAPIObjectAt(t, schemas, "DefaultWorkbookPreferencesEnvelope"), "DefaultWorkbookPreferencesResource")
	requireOpenAPIEnvelopeSchema(t, openAPIObjectAt(t, schemas, "UserWorkbookPreferencesEnvelope"), "UserWorkbookPreferencesResource")
}

func toStrings(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]string)
	if ok {
		return append([]string(nil), items...)
	}
	interfaces, ok := raw.([]any)
	if !ok {
		t.Fatalf("unexpected string slice payload: %T", raw)
	}
	values := make([]string, 0, len(interfaces))
	for _, item := range interfaces {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("unexpected string item payload: %T", item)
		}
		values = append(values, value)
	}
	return values
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func openAPIObjectAt(t testing.TB, root any, path ...string) map[string]any {
	t.Helper()

	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected OpenAPI object before %q, got %T", segment, current)
		}
		var found bool
		current, found = object[segment]
		if !found {
			t.Fatalf("missing OpenAPI path segment %q in %#v", segment, object)
		}
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI object at %v, got %T", path, current)
	}
	return object
}

func requireOpenAPIOperation(t testing.TB, path map[string]any, method string, wantOperationID string) {
	t.Helper()

	operation := openAPIObjectAt(t, path, method)
	if operation["operationId"] != wantOperationID {
		t.Fatalf("unexpected %s operationId: got %v want %q", method, operation["operationId"], wantOperationID)
	}
}

func requireOpenAPIRequestSchemaRef(t testing.TB, operation map[string]any, wantSchemaName string) {
	t.Helper()

	requestBody := openAPIObjectAt(t, operation, "requestBody")
	if requestBody["required"] != true {
		t.Fatalf("OpenAPI requestBody must be required: %#v", requestBody)
	}
	schema := openAPIObjectAt(t, requestBody, "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchemaName
	if schema["$ref"] != wantRef {
		t.Fatalf("unexpected request schema ref: got %v want %q", schema["$ref"], wantRef)
	}
}

func requireOpenAPIResponseSchemaRef(t testing.TB, operation map[string]any, wantSchemaName string) {
	t.Helper()

	requireOpenAPIStatusResponseSchemaRef(t, operation, "200", wantSchemaName)
}

func requireOpenAPIStatusResponseSchemaRef(t testing.TB, operation map[string]any, status string, wantSchemaName string) {
	t.Helper()

	schema := openAPIObjectAt(t, operation, "responses", status, "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchemaName
	if schema["$ref"] != wantRef {
		t.Fatalf("unexpected response %s schema ref: got %v want %q", status, schema["$ref"], wantRef)
	}
}

func requireOpenAPISheetRefSchema(t testing.TB, schemas map[string]any) {
	t.Helper()

	schema := openAPIObjectAt(t, schemas, "SheetRef")
	oneOf, ok := schema["oneOf"].([]any)
	if !ok || len(oneOf) != 3 {
		t.Fatalf("SheetRef must be a closed three-variant union: %#v", schema)
	}
	wantRefs := []string{
		"#/components/schemas/ViewSchemaSheetRef",
		"#/components/schemas/SavedViewSheetRef",
		"#/components/schemas/ExtensionWorkspaceSheetRef",
	}
	for index, wantRef := range wantRefs {
		ref, ok := oneOf[index].(map[string]any)
		if !ok || ref["$ref"] != wantRef {
			t.Fatalf("unexpected SheetRef oneOf[%d]: got %#v want %q", index, oneOf[index], wantRef)
		}
	}
	discriminator := openAPIObjectAt(t, schema, "discriminator")
	if discriminator["propertyName"] != "kind" {
		t.Fatalf("SheetRef discriminator must use kind: %#v", discriminator)
	}
	mapping := openAPIObjectAt(t, discriminator, "mapping")
	wantMapping := map[string]string{
		"view_schema":         "#/components/schemas/ViewSchemaSheetRef",
		"saved_view":          "#/components/schemas/SavedViewSheetRef",
		"extension_workspace": "#/components/schemas/ExtensionWorkspaceSheetRef",
	}
	if len(mapping) != len(wantMapping) {
		t.Fatalf("unexpected SheetRef discriminator mapping: %#v", mapping)
	}
	for kind, wantRef := range wantMapping {
		if mapping[kind] != wantRef {
			t.Fatalf("unexpected SheetRef mapping for %q: got %v want %q", kind, mapping[kind], wantRef)
		}
	}
	requireOpenAPISheetRefVariant(t, openAPIObjectAt(t, schemas, "ViewSchemaSheetRef"), "view_schema", []string{"kind", "id"})
	requireOpenAPISheetRefVariant(t, openAPIObjectAt(t, schemas, "SavedViewSheetRef"), "saved_view", []string{"kind", "id"})
	requireOpenAPISheetRefVariant(t, openAPIObjectAt(t, schemas, "ExtensionWorkspaceSheetRef"), "extension_workspace", []string{"kind", "extension_profile_id", "workspace_key"})
}

func requireOpenAPISheetRefVariant(t testing.TB, schema map[string]any, kindValue string, requiredFields []string) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("SheetRef variant must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, requiredFields) {
		t.Fatalf("unexpected %s required fields: %v", kindValue, required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	if len(properties) != len(requiredFields) {
		t.Fatalf("%s must expose exactly its required fields: %#v", kindValue, properties)
	}
	kind := openAPIObjectAt(t, properties, "kind")
	if kind["const"] != kindValue {
		t.Fatalf("unexpected %s kind schema: %#v", kindValue, kind)
	}
	for _, field := range requiredFields[1:] {
		value := openAPIObjectAt(t, properties, field)
		if value["type"] != "string" || value["minLength"] != float64(1) {
			t.Fatalf("unexpected %s.%s schema: %#v", kindValue, field, value)
		}
	}
}

func requireOpenAPIWorkbookPreferencesPutSchema(t testing.TB, schema map[string]any, field string) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("workbook preferences PUT request must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, []string{field}) {
		t.Fatalf("unexpected workbook preferences PUT required fields: %v", required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	if len(properties) != 1 {
		t.Fatalf("workbook preferences PUT request must expose exactly one property: %#v", properties)
	}
	fieldSchema := openAPIObjectAt(t, properties, field)
	oneOf, ok := fieldSchema["oneOf"].([]any)
	if !ok || len(oneOf) != 2 {
		t.Fatalf("workbook preferences PUT field must be SheetRef or null: %#v", fieldSchema)
	}
	sheetRef, ok := oneOf[0].(map[string]any)
	if !ok || sheetRef["$ref"] != "#/components/schemas/SheetRef" {
		t.Fatalf("workbook preferences PUT field must reference SheetRef first: %#v", oneOf[0])
	}
	nullRef, ok := oneOf[1].(map[string]any)
	if !ok || nullRef["type"] != "null" {
		t.Fatalf("workbook preferences PUT field must allow explicit null: %#v", oneOf[1])
	}
}

func requireOpenAPIEnvelopeSchema(t testing.TB, schema map[string]any, wantDataSchemaName string) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("workbook preferences envelope must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, []string{"data", "meta"}) {
		t.Fatalf("unexpected workbook preferences envelope required fields: %v", required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	data := openAPIObjectAt(t, properties, "data")
	if data["$ref"] != "#/components/schemas/"+wantDataSchemaName {
		t.Fatalf("unexpected workbook preferences envelope data ref: %#v", data)
	}
	meta := openAPIObjectAt(t, properties, "meta")
	if meta["$ref"] != "#/components/schemas/EnvelopeMeta" {
		t.Fatalf("unexpected workbook preferences envelope meta ref: %#v", meta)
	}
}
