package auth

import (
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestAuthOpenAPIResponsesAndSecurityMatchRuntime_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := authOpenAPIObjectAt(t, document, "paths")

	expectedStatuses := map[string][]string{
		"beginEnterpriseAuth":             {"200", "400", "404", "500"},
		"beginTOTPEnrollment":             {"200", "400", "401", "403", "409", "500"},
		"changeCurrentPassword":           {"200", "400", "401", "403", "409", "500"},
		"completeEnterpriseOIDC":          {"303", "404", "409", "500"},
		"completeEnterpriseSAML":          {"303", "404", "409", "500"},
		"completeTOTPEnrollment":          {"200", "400", "401", "403", "409", "500"},
		"createDeploymentUser":            {"201", "400", "401", "403", "409", "500"},
		"createEnterpriseAuthBinding":     {"201", "400", "401", "403", "404", "409", "500"},
		"getCredentialState":              {"200", "400", "401", "409", "500"},
		"getCurrentAccountPreferences":    {"200", "400", "401", "409", "500"},
		"getCurrentAccountProfile":        {"200", "400", "401", "409", "500"},
		"getCurrentSession":               {"200", "400", "401", "409", "500"},
		"getDeploymentUser":               {"200", "400", "401", "403", "404", "409", "500"},
		"listAdministrativeAuditEvents":   {"200", "400", "401", "403", "409", "500"},
		"listDeploymentUsers":             {"200", "400", "401", "403", "409", "500"},
		"listEnterpriseAuthProviders":     {"200", "400", "404", "500"},
		"loginLocalUser":                  {"200", "400", "401", "500"},
		"logoutCurrentSession":            {"200", "401", "403", "409", "500"},
		"patchCurrentAccountProfile":      {"200", "400", "401", "403", "409", "500"},
		"patchDeploymentUser":             {"200", "400", "401", "403", "409", "500"},
		"putCurrentAccountPreferences":    {"200", "400", "401", "403", "409", "500"},
		"resetDeploymentUserPassword":     {"200", "400", "401", "403", "409", "500"},
		"resetDeploymentUserTOTP":         {"200", "400", "401", "403", "409", "500"},
		"retireEnterpriseAuthBinding":     {"200", "400", "401", "403", "404", "409", "500"},
		"revokeAllDeploymentUserSessions": {"200", "400", "401", "403", "409", "500"},
		"rotateEnterpriseAuthBinding":     {"200", "400", "401", "403", "404", "409", "500"},
	}

	descriptors := append(PublicOperations(), EnterprisePublicOperations()...)
	if len(descriptors) != len(expectedStatuses) {
		t.Fatalf("auth descriptor inventory changed: got %d want %d", len(descriptors), len(expectedStatuses))
	}
	seen := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		operation := authOpenAPIObjectAt(
			t,
			authOpenAPIObjectAt(t, paths, descriptor.PathTemplate),
			authOpenAPIMethodKey(descriptor.Method),
		)
		if operation["operationId"] != descriptor.OperationID {
			t.Fatalf("%s operationId mismatch: %#v", descriptor.OperationID, operation["operationId"])
		}

		responses := authOpenAPIObjectAt(t, operation, "responses")
		statuses := authOpenAPISortedKeys(responses)
		if want := expectedStatuses[descriptor.OperationID]; !slices.Equal(statuses, want) {
			t.Fatalf("%s response statuses changed: got %v want %v", descriptor.OperationID, statuses, want)
		}
		for status, rawResponse := range responses {
			response := authOpenAPIObject(t, rawResponse)
			if ref, ok := response["$ref"].(string); ok {
				if ref == "" {
					t.Fatalf("%s %s has an empty response reference", descriptor.OperationID, status)
				}
				continue
			}
			if response["description"] == nil {
				t.Fatalf("%s %s has neither a response reference nor description", descriptor.OperationID, status)
			}
		}

		if want := authOpenAPISecurityForDescriptor(descriptor); !reflect.DeepEqual(operation["security"], want) {
			t.Fatalf("%s security mismatch: got %#v want %#v", descriptor.OperationID, operation["security"], want)
		}
		seen[descriptor.OperationID] = struct{}{}
	}
	for operationID := range expectedStatuses {
		if _, ok := seen[operationID]; !ok {
			t.Fatalf("expected auth operation %s has no runtime descriptor", operationID)
		}
	}

	assertAuthResponseSchemaChain(t, document, "AuthSessionSuccessResponse", "SessionEnvelope", "SessionResource")
	assertAuthResponseSchemaChain(t, document, "AuthCredentialStateSuccessResponse", "CredentialStateEnvelope", "CredentialStateResource")
	assertAuthResponseSchemaChain(t, document, "AuthSafeUserSuccessResponse", "SafeUserEnvelope", "SafeUserResource")
	assertAuthResponseSchemaChain(t, document, "AuthTOTPBeginSuccessResponse", "TOTPBeginEnvelope", "TOTPBeginResponse")
}

func authOpenAPISecurityForDescriptor(descriptor httpapi.PublicOperation) []any {
	if descriptor.Authentication == httpapi.PublicAuthenticationPublic {
		return []any{}
	}
	if descriptor.Authentication == httpapi.PublicAuthenticationSessionOrBootstrap {
		return []any{
			map[string]any{"sessionCookie": []any{}, "csrfCookie": []any{}, "csrfHeader": []any{}},
			map[string]any{"bearerSession": []any{}},
			map[string]any{"credentialBootstrapBearer": []any{}},
		}
	}
	if descriptor.StateChanging {
		return []any{
			map[string]any{"sessionCookie": []any{}, "csrfCookie": []any{}, "csrfHeader": []any{}},
			map[string]any{"bearerSession": []any{}},
		}
	}
	return []any{
		map[string]any{"sessionCookie": []any{}},
		map[string]any{"bearerSession": []any{}},
	}
}

func assertAuthResponseSchemaChain(t testing.TB, document map[string]any, responseName, envelopeName, resourceName string) {
	t.Helper()
	components := authOpenAPIObjectAt(t, document, "components")
	response := authOpenAPIObjectAt(t, authOpenAPIObjectAt(t, components, "responses"), responseName)
	schema := authOpenAPIObjectAt(
		t,
		authOpenAPIObjectAt(t, authOpenAPIObjectAt(t, response, "content"), "application/json"),
		"schema",
	)
	if want := "#/components/schemas/" + envelopeName; schema["$ref"] != want {
		t.Fatalf("%s schema reference mismatch: got %#v want %q", responseName, schema["$ref"], want)
	}
	envelope := authOpenAPIObjectAt(t, authOpenAPIObjectAt(t, components, "schemas"), envelopeName)
	data := authOpenAPIObjectAt(t, authOpenAPIObjectAt(t, envelope, "properties"), "data")
	if want := "#/components/schemas/" + resourceName; data["$ref"] != want {
		t.Fatalf("%s data reference mismatch: got %#v want %q", envelopeName, data["$ref"], want)
	}
}

func authOpenAPIObjectAt(t testing.TB, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, part := range path {
		current = authOpenAPIObject(t, current[part])
	}
	return current
}

func authOpenAPIObject(t testing.TB, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI object, got %T", value)
	}
	return object
}

func authOpenAPISortedKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func authOpenAPIMethodKey(method string) string {
	switch method {
	case "DELETE":
		return "delete"
	case "GET":
		return "get"
	case "PATCH":
		return "patch"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	default:
		return method
	}
}
