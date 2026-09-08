package httpapi

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestMembershipManagementOpenAPIProjectsExactVersionedBodies_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObjectAt(t, document, "paths")
	list := openAPIObjectAt(t, openAPIObjectAt(t, paths, "/api/v1/incidents/{incident_id}/memberships"), "get")
	parameters, ok := list["parameters"].([]any)
	if !ok || len(parameters) != 2 {
		t.Fatalf("list must project only limit and cursor_token: %#v", list["parameters"])
	}
	for index, name := range []string{"limit", "cursor_token"} {
		parameter, ok := parameters[index].(map[string]any)
		if !ok || parameter["name"] != name || parameter["in"] != "query" {
			t.Fatalf("unexpected list parameter: %#v", parameters[index])
		}
	}
	route := openAPIObjectAt(t, paths, "/api/v1/incidents/{incident_id}/memberships/{user_id}")
	schemas := openAPIObjectAt(t, openAPIObjectAt(t, document, "components"), "schemas")
	for _, tc := range []struct {
		method, schema string
		required       []any
	}{
		{"patch", "IncidentMembershipPatchRequest", []any{"base_membership_version", "role"}},
		{"delete", "IncidentMembershipDeleteRequest", []any{"base_membership_version"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			body := openAPIObjectAt(t, openAPIObjectAt(t, route, tc.method), "requestBody")
			if body["required"] != true {
				t.Fatal("membership body must be required")
			}
			ref := openAPIObjectAt(t, openAPIObjectAt(t, openAPIObjectAt(t, body, "content"), "application/json"), "schema")
			if ref["$ref"] != "#/components/schemas/"+tc.schema {
				t.Fatalf("unexpected request schema: %#v", ref)
			}
			schema := openAPIObjectAt(t, schemas, tc.schema)
			if schema["additionalProperties"] != false || !reflect.DeepEqual(schema["required"], tc.required) {
				t.Fatalf("request must contain exactly its required versioned fields: %#v", schema)
			}
			properties := openAPIObjectAt(t, schema, "properties")
			if len(properties) != len(tc.required) {
				t.Fatal("unexpected request fields")
			}
			version := openAPIObjectAt(t, properties, "base_membership_version")
			if version["type"] != "integer" || version["minimum"] != float64(1) {
				t.Fatalf("expected positive version: %#v", version)
			}
			if tc.method == "patch" && !reflect.DeepEqual(openAPIObjectAt(t, properties, "role")["enum"], []any{"viewer", "editor", "reviewer", "admin"}) {
				t.Fatal("unexpected roles")
			}
		})
	}
}
