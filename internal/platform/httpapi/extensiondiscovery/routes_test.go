package extensiondiscovery

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	extensionroutetest "github.com/JochiRaider/cartulary/internal/platform/httpapi/extensiondiscovery/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
)

func TestExtensionDiscoveryResponseShapeAndOrder_Unit(t *testing.T) {
	provider := httpapiextensions.FromGeneratedRegistry(t)
	profiles := provider.ExtensionDiscoveryProfiles()
	data := buildResponseData(profiles)
	items, ok := data["extensions"].([]map[string]any)
	if !ok || len(items) != len(profiles) {
		t.Fatalf("discovery data = %#v", data)
	}
	for index, item := range items {
		if len(item) != 7 {
			t.Fatalf("item %d member count = %d, want 7", index, len(item))
		}
		if item["profile_id"] != profiles[index].ProfileID {
			t.Fatalf("item %d profile_id = %v, want %s", index, item["profile_id"], profiles[index].ProfileID)
		}
		capabilities, ok := item["capabilities"].([]string)
		if !ok || len(capabilities) != 0 {
			t.Fatalf("item %d capabilities = %#v, want present empty array", index, item["capabilities"])
		}
	}
}

func TestExtensionDiscoveryRejectsSingletonQueryBeforeAuthentication_Unit(t *testing.T) {
	discovery := &service{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/extensions?cursor_token=opaque", nil)
	response := httptest.NewRecorder()
	discovery.handleCollection(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(url.Values{"cursor_token": {"opaque"}}); apiErr == nil || apiErr.Code != "invalid_pagination_request" {
		t.Fatalf("singleton query error = %#v", apiErr)
	}
}

func TestExtensionDiscoveryRequiresAuthentication_Unit(t *testing.T) {
	discovery := &service{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/extensions", nil)
	response := httptest.NewRecorder()
	discovery.handleCollection(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestPublicRouteInventoryExtensionDiscovery_Unit(t *testing.T) {
	routes := extensionroutetest.PublicDiscovery()
	if len(routes) != 1 {
		t.Fatalf("extension discovery route inventory = %#v", routes)
	}
	route := routes[0]
	if route.Method != http.MethodGet || route.Template != "/api/v1/extensions" ||
		route.SuccessStatus != http.StatusOK || !route.SuccessEnvelope {
		t.Fatalf("extension discovery route inventory = %#v", route)
	}
}

func TestOpenAPIExtensionDiscoveryExposesClosedContract_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	operation := openAPIMapAt(t, document, "paths", "/api/v1/extensions", "get")
	if operation["operationId"] != "listDeploymentExtensions" {
		t.Fatalf("extension discovery operationId = %v", operation["operationId"])
	}
	for status, schemaName := range map[string]string{
		"200": "ExtensionDiscoveryEnvelope",
		"400": "ErrorEnvelope",
		"401": "ErrorEnvelope",
	} {
		schema := openAPIMapAt(t, operation, "responses", status, "content", "application/json", "schema")
		if schema["$ref"] != "#/components/schemas/"+schemaName {
			t.Fatalf("extension discovery response %s schema = %#v", status, schema)
		}
	}

	schemas := openAPIMapAt(t, document, "components", "schemas")
	envelope := openAPIMapAt(t, schemas, "ExtensionDiscoveryEnvelope")
	requireClosedObject(t, envelope, []string{"data", "meta"})
	if dataRef := openAPIMapAt(t, envelope, "properties", "data"); dataRef["$ref"] != "#/components/schemas/ExtensionDiscoveryData" {
		t.Fatalf("extension discovery envelope data = %#v", dataRef)
	}

	data := openAPIMapAt(t, schemas, "ExtensionDiscoveryData")
	requireClosedObject(t, data, []string{"extensions"})
	if item := openAPIMapAt(t, data, "properties", "extensions", "items"); item["$ref"] != "#/components/schemas/ExtensionProfileResource" {
		t.Fatalf("extension discovery item schema = %#v", item)
	}

	resource := openAPIMapAt(t, schemas, "ExtensionProfileResource")
	required := []string{"profile_id", "claimable", "claimed", "contract_major", "route_families", "workspace_keys", "capabilities"}
	requireClosedObject(t, resource, required)
	properties := openAPIMapAt(t, resource, "properties")
	if len(properties) != 7 {
		t.Fatalf("extension profile property count = %d, want 7", len(properties))
	}
	if profileID := openAPIMapAt(t, schemas, "ExtensionProfileID"); profileID["type"] != "string" || profileID["pattern"] != "^[a-z][a-z0-9_]{0,63}$" {
		t.Fatalf("extension profile id schema = %#v", profileID)
	}
	if routeFamily := openAPIMapAt(t, schemas, "ExtensionRouteFamily"); routeFamily["type"] != "string" || routeFamily["pattern"] == nil {
		t.Fatalf("extension route family schema = %#v", routeFamily)
	}
}

func requireClosedObject(t *testing.T, object map[string]any, required []string) {
	t.Helper()
	if object["type"] != "object" || object["additionalProperties"] != false {
		t.Fatalf("schema must be a closed object: %#v", object)
	}
	raw, ok := object["required"].([]any)
	if !ok {
		t.Fatalf("required fields = %T", object["required"])
	}
	actual := make([]string, 0, len(raw))
	for _, item := range raw {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("required field = %T", item)
		}
		actual = append(actual, value)
	}
	if !reflect.DeepEqual(actual, required) {
		t.Fatalf("required fields = %v, want %v", actual, required)
	}
}

func openAPIMapAt(t *testing.T, root any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI path %v reaches %T before %q", path, current, segment)
		}
		current, ok = object[segment]
		if !ok {
			t.Fatalf("OpenAPI path %v is missing %q", path, segment)
		}
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("OpenAPI path %v resolves to %T", path, current)
	}
	return object
}
