package httpapi

import (
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestOpenAPIIncidentMembershipAuditOperationIsCompleteAndExact_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObjectAt(t, document, "paths")
	pathItem := openAPIObjectAt(t, paths, "/api/v1/incidents/{incident_id}/membership-audit-events")
	operation := openAPIObjectAt(t, pathItem, "get")
	if operation["operationId"] != "listIncidentMembershipAuditEvents" {
		t.Fatalf("membership audit operationId got %#v", operation["operationId"])
	}
	if !reflect.DeepEqual(operation["security"], []any{
		map[string]any{"sessionCookie": []any{}},
		map[string]any{"bearerSession": []any{}},
	}) {
		t.Fatalf("membership audit security got %#v", operation["security"])
	}
	pathParameters := toObjects(t, pathItem["parameters"])
	if len(pathParameters) != 1 || pathParameters[0]["$ref"] != "#/components/parameters/IncidentIDPathParameter" {
		t.Fatalf("membership audit path parameter got %#v", pathParameters)
	}

	parameters := toObjects(t, operation["parameters"])
	parameterNames := make([]string, 0, len(parameters))
	parameterByName := make(map[string]map[string]any, len(parameters))
	for _, parameter := range parameters {
		name, ok := parameter["name"].(string)
		if !ok {
			t.Fatalf("membership audit parameter has invalid name: %#v", parameter)
		}
		parameterNames = append(parameterNames, name)
		parameterByName[name] = parameter
	}
	sort.Strings(parameterNames)
	wantParameterNames := []string{
		"action_code",
		"actor_user_id",
		"cursor_token",
		"limit",
		"occurred_at_gte",
		"occurred_at_lt",
		"target_id",
		"target_kind",
	}
	if !equalStringSlices(parameterNames, wantParameterNames) {
		t.Fatalf("membership audit parameters got %v want %v", parameterNames, wantParameterNames)
	}
	actionSchema := openAPIObjectAt(t, parameterByName["action_code"], "schema")
	if got := toStrings(t, actionSchema["enum"]); !equalStringSlices(got, administrativeaudit.ActionCodes(administrativeaudit.ScopeIncident)) {
		t.Fatalf("membership audit action filter enum got %v", got)
	}
	targetSchema := openAPIObjectAt(t, parameterByName["target_kind"], "schema")
	if got := toStrings(t, targetSchema["enum"]); !equalStringSlices(got, administrativeaudit.TargetKinds(administrativeaudit.ScopeIncident)) {
		t.Fatalf("membership audit target filter enum got %v", got)
	}

	responses := openAPIObjectAt(t, operation, "responses")
	responseStatuses := make([]string, 0, len(responses))
	for status := range responses {
		responseStatuses = append(responseStatuses, status)
	}
	sort.Strings(responseStatuses)
	if want := []string{"200", "400", "401", "403", "404", "409", "500"}; !equalStringSlices(responseStatuses, want) {
		t.Fatalf("membership audit response statuses got %v want %v", responseStatuses, want)
	}
	successMedia := openAPIObjectAt(
		t,
		openAPIObjectAt(t, openAPIObjectAt(t, responses, "200"), "content"),
		"application/json",
	)
	if openAPIObjectAt(t, successMedia, "schema")["$ref"] != "#/components/schemas/AdministrativeAuditEnvelope" {
		t.Fatalf("membership audit success schema got %#v", successMedia)
	}
	wantErrorCodes := map[string][]string{
		"400": {"invalid_list_query", "invalid_pagination_request"},
		"401": {"session_required"},
		"403": {"authorization_denied"},
		"404": {"incident_not_found"},
		"409": {"credential_bootstrap_rejected"},
		"500": {"internal_error"},
	}
	for status, wantCodes := range wantErrorCodes {
		response := openAPIObjectAt(t, responses, status)
		if got := toStrings(t, response["x-cartulary-error-codes"]); !equalStringSlices(got, wantCodes) {
			t.Fatalf("membership audit %s error codes got %v want %v", status, got, wantCodes)
		}
		errorMedia := openAPIObjectAt(t, openAPIObjectAt(t, response, "content"), "application/json")
		if openAPIObjectAt(t, errorMedia, "schema")["$ref"] != "#/components/schemas/ErrorEnvelope" {
			t.Fatalf("membership audit %s error schema got %#v", status, errorMedia)
		}
	}

	schemas := openAPIObjectAt(t, openAPIObjectAt(t, document, "components"), "schemas")
	dataSchema := openAPIObjectAt(t, schemas, "AdministrativeAuditData")
	if dataSchema["type"] != "object" || dataSchema["additionalProperties"] != false {
		t.Fatalf("administrative audit data must be closed: %#v", dataSchema)
	}
	dataProperties := openAPIObjectAt(t, dataSchema, "properties")
	if len(dataProperties) != 1 || dataProperties["audit_events"] == nil {
		t.Fatalf("administrative audit data members must be exact: %#v", dataProperties)
	}
	resource := openAPIObjectAt(t, schemas, "AdministrativeAuditResource")
	if resource["type"] != "object" || resource["additionalProperties"] != false {
		t.Fatalf("administrative audit resource must be closed: %#v", resource)
	}
	resourceProperties := openAPIObjectAt(t, resource, "properties")
	if len(resourceProperties) != 12 || len(toStrings(t, resource["required"])) != 12 {
		t.Fatalf("administrative audit resource member closure drifted: %#v", resource)
	}
	for _, field := range []string{"action_code", "target_kind"} {
		fieldSchema := openAPIObjectAt(t, resourceProperties, field)
		if fieldSchema["type"] != "string" {
			t.Fatalf("%s read field must be a string: %#v", field, fieldSchema)
		}
		if _, closed := fieldSchema["enum"]; closed {
			t.Fatalf("%s read field must tolerate additive future values: %#v", field, fieldSchema)
		}
	}
}

func toObjects(t testing.TB, raw any) []map[string]any {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected object array, got %T", raw)
	}
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("expected object array member, got %T", value)
		}
		result = append(result, object)
	}
	return result
}
