package extensiondiscovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
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

// serveDiscovery binds the discovery singleton for central admission behind
// the registry's route resolution, as production does (Core 01 REQ-01-707),
// with an authenticator over sessions.
func serveDiscovery(t *testing.T, sessions httpauth.SessionStore) http.Handler {
	t.Helper()
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	deps := httpapi.DependencySet{PublicRoutes: registry, Authenticator: httpauth.NewAuthenticator(httpauth.Options{Store: sessions, Now: time.Now})}
	mux := http.NewServeMux()
	if err := RegisterRoutes()(mux, deps); err != nil {
		t.Fatal(err)
	}
	if mode := registry.AdmissionModes()["listDeploymentExtensions"]; mode != httpapi.AdmissionCentral {
		t.Fatalf("listDeploymentExtensions admission mode = %q, want central", mode)
	}
	return registry.Handler(mux)
}

// activeSessions authenticates any bearer token as one active session.
type activeSessions struct{}

func (activeSessions) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	now := time.Now().UTC()
	user := authn.UserRecord{ID: uuid.MustParse("7e1a2b3c-4d5e-4f60-8a71-b2c3d4e5f601"), IsActive: true}
	return authn.SessionRecord{
		ID: uuid.MustParse("7e1a2b3c-4d5e-4f60-8a71-b2c3d4e5f602"), UserID: user.ID,
		AuthenticatedAt: now, LastQualifyingActivityAt: now,
		IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(time.Hour), SessionExpiresAt: now.Add(time.Hour),
	}, user, nil
}

func (activeSessions) GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}

func (activeSessions) RevokeSession(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}

// TestExtensionDiscoveryAuthenticatesBeforeSingletonQuery_Unit: central
// admission authenticates first (Core 01 REQ-01-707 step 2), so an
// unauthenticated singleton query is session_required, and only an
// authenticated one meets the singleton-query check (step 5).
func TestExtensionDiscoveryAuthenticatesBeforeSingletonQuery_Unit(t *testing.T) {
	handler := serveDiscovery(t, activeSessions{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/extensions?cursor_token=opaque", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d; body=%s", response.Code, http.StatusUnauthorized, response.Body)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/extensions?cursor_token=opaque", nil)
	request.Header.Set("Authorization", "Bearer extension-discovery-query")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("authenticated status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body)
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(url.Values{"cursor_token": {"opaque"}}); apiErr == nil || apiErr.Code != "invalid_pagination_request" {
		t.Fatalf("singleton query error = %#v", apiErr)
	}
}

func TestExtensionDiscoveryRequiresAuthentication_Unit(t *testing.T) {
	response := httptest.NewRecorder()
	serveDiscovery(t, nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/extensions", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
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
