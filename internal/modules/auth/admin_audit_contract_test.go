package auth

import (
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestDeploymentAdministrativeAuditOpenAPIContract_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	operation := administrativeAuditOpenAPIObjectAt(
		t,
		administrativeAuditOpenAPIObjectAt(
			t,
			administrativeAuditOpenAPIObjectAt(t, document, "paths"),
			"/api/v1/administrative-audit-events",
		),
		"get",
	)
	if operation["operationId"] != "listAdministrativeAuditEvents" {
		t.Fatalf("unexpected administrative-audit operationId: %#v", operation["operationId"])
	}

	parameters := administrativeAuditOpenAPIArray(t, operation["parameters"])
	parameterNames := make([]string, 0, len(parameters))
	for _, rawParameter := range parameters {
		parameter := administrativeAuditOpenAPIObject(t, rawParameter)
		parameterNames = append(parameterNames, parameter["name"].(string))
	}
	if want := []string{
		"limit",
		"cursor_token",
		"actor_user_id",
		"action_code",
		"target_kind",
		"target_id",
		"occurred_at_gte",
		"occurred_at_lt",
	}; !reflect.DeepEqual(parameterNames, want) {
		t.Fatalf("unexpected administrative-audit parameters: got %v want %v", parameterNames, want)
	}
	targetKindParameter := administrativeAuditOpenAPIObject(t, parameters[4])
	targetKindSchema := administrativeAuditOpenAPIObjectAt(t, targetKindParameter, "schema")
	if got, want := administrativeAuditOpenAPIStrings(t, targetKindSchema["enum"]), administrativeaudit.TargetKinds(administrativeaudit.ScopeDeployment); !reflect.DeepEqual(got, want) {
		t.Fatalf("deployment target-kind filter drifted from machine audit mappings: got %v want %v", got, want)
	}
	actionCodeParameter := administrativeAuditOpenAPIObject(t, parameters[3])
	actionCodeSchema := administrativeAuditOpenAPIObjectAt(t, actionCodeParameter, "schema")
	if got, want := administrativeAuditOpenAPIStrings(t, actionCodeSchema["enum"]), administrativeaudit.ActionCodes(administrativeaudit.ScopeDeployment); !reflect.DeepEqual(got, want) {
		t.Fatalf("deployment action-code filter drifted from machine audit mappings: got %v want %v", got, want)
	}

	responses := administrativeAuditOpenAPIObjectAt(t, operation, "responses")
	responseStatuses := make([]string, 0, len(responses))
	for status := range responses {
		responseStatuses = append(responseStatuses, status)
	}
	sort.Strings(responseStatuses)
	if want := []string{"200", "400", "401", "403", "409", "500"}; !reflect.DeepEqual(responseStatuses, want) {
		t.Fatalf("unexpected administrative-audit responses: got %v want %v", responseStatuses, want)
	}
	successSchema := administrativeAuditOpenAPIObjectAt(
		t,
		administrativeAuditOpenAPIObjectAt(
			t,
			administrativeAuditOpenAPIObjectAt(t, responses, "200"),
			"content",
			"application/json",
		),
		"schema",
	)
	if successSchema["$ref"] != "#/components/schemas/AdministrativeAuditEnvelope" {
		t.Fatalf("unexpected administrative-audit success schema: %#v", successSchema)
	}
	if want := []any{
		map[string]any{"sessionCookie": []any{}},
		map[string]any{"bearerSession": []any{}},
	}; !reflect.DeepEqual(operation["security"], want) {
		t.Fatalf("unexpected administrative-audit security: %#v", operation["security"])
	}

	schemas := administrativeAuditOpenAPIObjectAt(
		t,
		administrativeAuditOpenAPIObjectAt(t, document, "components"),
		"schemas",
	)
	change := administrativeAuditOpenAPIObjectAt(t, schemas, "AdministrativeAuditChange")
	requireClosedAdministrativeAuditSchema(t, change, []string{"field_path", "value_state", "before", "after"})
	resource := administrativeAuditOpenAPIObjectAt(t, schemas, "AdministrativeAuditResource")
	requireClosedAdministrativeAuditSchema(t, resource, []string{
		"audit_event_id",
		"scope_kind",
		"scope_id",
		"occurred_at",
		"actor_kind",
		"actor_user_id",
		"source",
		"action_code",
		"target_kind",
		"target_id",
		"changes",
		"reason_code",
	})
	resourceProperties := administrativeAuditOpenAPIObjectAt(t, resource, "properties")
	if _, closed := administrativeAuditOpenAPIObjectAt(t, resourceProperties, "action_code")["enum"]; closed {
		t.Fatal("administrative-audit action_code must remain forward-tolerant")
	}
	if _, closed := administrativeAuditOpenAPIObjectAt(t, resourceProperties, "target_kind")["enum"]; closed {
		t.Fatal("administrative-audit target_kind must remain forward-tolerant")
	}
	data := administrativeAuditOpenAPIObjectAt(t, schemas, "AdministrativeAuditData")
	requireClosedAdministrativeAuditSchema(t, data, []string{"audit_events"})
	if _, legacy := administrativeAuditOpenAPIObjectAt(t, data, "properties")["administrative_audit_events"]; legacy {
		t.Fatal("obsolete administrative_audit_events alias must not be accepted")
	}
}

func requireClosedAdministrativeAuditSchema(t testing.TB, schema map[string]any, required []string) {
	t.Helper()
	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("administrative-audit schema must be closed: %#v", schema)
	}
	if got := administrativeAuditOpenAPIStrings(t, schema["required"]); !reflect.DeepEqual(got, required) {
		t.Fatalf("unexpected required fields: got %v want %v", got, required)
	}
	properties := administrativeAuditOpenAPIObjectAt(t, schema, "properties")
	if len(properties) != len(required) {
		t.Fatalf("administrative-audit properties are not exact: %#v", properties)
	}
}

func administrativeAuditOpenAPIObjectAt(t testing.TB, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, part := range path {
		current = administrativeAuditOpenAPIObject(t, current[part])
	}
	return current
}

func administrativeAuditOpenAPIObject(t testing.TB, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI object, got %T", value)
	}
	return object
}

func administrativeAuditOpenAPIArray(t testing.TB, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("expected OpenAPI array, got %T", value)
	}
	return array
}

func administrativeAuditOpenAPIStrings(t testing.TB, value any) []string {
	t.Helper()
	array := administrativeAuditOpenAPIArray(t, value)
	result := make([]string, 0, len(array))
	for _, entry := range array {
		text, ok := entry.(string)
		if !ok {
			t.Fatalf("expected OpenAPI string, got %T", entry)
		}
		result = append(result, text)
	}
	return result
}
