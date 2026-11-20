package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const networkFlowFixtureRoutes = `{
  "schema_id": "cartulary.network_flow_route_contracts.v1",
  "profile_id": "fixture_profile",
  "path_parameters": {
    "incident_id": {"format": "uuid", "admission_stage": "incident_visibility_role_lifecycle",
      "not_found": {"code": "incident_not_found", "http_status": 404, "details": {}}},
    "table_id": {"schema_def": "TableID", "admission_stage": "resource_path",
      "not_found": {"code": "fixture_table_not_found", "http_status": 404,
        "details": {"allowed_states": ["active"], "table_id": null, "reason_code": "not_found"}}}
  },
  "routes": [
    {"route_id": "fx.tables.patch", "method": "PATCH", "path": "/api/v1/incidents/{incident_id}/fixture/tables/{table_id}",
      "security": [["bearerSession"], ["csrfCookie", "csrfHeader", "sessionCookie"]], "success_http_statuses": [200]},
    {"route_id": "fx.tables.list", "method": "GET", "path": "/api/v1/incidents/{incident_id}/fixture/tables",
      "security": [["bearerSession"], ["sessionCookie"]], "success_http_statuses": [200]}
  ]
}`

func networkFlowFixtureSources(t *testing.T, edit func(routes, policy, schemas, nfErrors map[string]any)) networkFlowSources {
	t.Helper()
	decode := func(text string) map[string]any {
		var value map[string]any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	routes := decode(networkFlowFixtureRoutes)
	policy := decode(`{"admission_stages": ["session_and_csrf", "incident_visibility_role_lifecycle", "resource_path", "query_body_framing"],
	  "operations": [{"route_id": "fx.tables.list", "csrf_required": false}, {"route_id": "fx.tables.patch", "csrf_required": true}]}`)
	schemas := decode(`{"$defs": {"TableID": {"type": "string", "pattern": "^fxt_[a-f0-9]{32}$"}, "Loose": {"type": "string", "pattern": "fxt_"}}}`)
	nfErrors := decode(`{"errors": [{"code": "fixture_table_not_found", "scope": "route", "http_status": 404}, {"code": "fixture_row_failure", "scope": "row_diagnostic", "http_status": null}]}`)
	if edit != nil {
		edit(routes, policy, schemas, nfErrors)
	}
	var sources networkFlowSources
	sources.index = networkFlowContractIndex{OwnerID: "module.fixture", ContractFiles: map[string]string{"errors": "contracts/fixture/errors.json"}}
	for _, item := range []struct {
		value  map[string]any
		target any
	}{{routes, &sources.routes}, {policy, &sources.policy}, {schemas, &sources.schemas}, {nfErrors, &sources.nfErrors}} {
		encoded, err := json.Marshal(item.value)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, item.target); err != nil {
			t.Fatal(err)
		}
	}
	if err := json.Unmarshal([]byte(`{"errors": [{"code": "incident_not_found", "http_status": 404}]}`), &sources.core); err != nil {
		t.Fatal(err)
	}
	return sources
}

// TestNetworkFlowRouteProjection covers the S-1100 projection of the Network
// Flow route contract into the public route registry (Core 01 REQ-01-707 and
// its family clause, NF NLSpec Table 21-C): each route gets its owner,
// availability, declared security, state-changing flag from the operation
// policy, and its path parameters in template order with their pattern or
// format, owner-gate position and registered not-found response; an
// incomplete or inconsistent contract fails generation; and the repository's
// own contract projects all nineteen routes behind the owner gate.
func TestNetworkFlowRouteProjection(t *testing.T) {
	routes, err := projectNetworkFlowRoutes(networkFlowFixtureSources(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	incident := networkFlowPathParameter{Name: "incident_id", Format: "uuid", NotFoundCode: "incident_not_found", NotFoundStatus: 404, NotFoundDetails: map[string]any{}}
	table := networkFlowPathParameter{Name: "table_id", Pattern: "^fxt_[a-f0-9]{32}$", AfterOwnerGate: true, NotFoundCode: "fixture_table_not_found", NotFoundStatus: 404,
		NotFoundDetails: map[string]any{"allowed_states": []any{"active"}, "table_id": nil, "reason_code": "not_found"}}
	want := []networkFlowRoute{
		{RouteID: "fx.tables.list", OwnerID: "module.fixture", Availability: "fixture_profile", Method: "GET", Path: "/api/v1/incidents/{incident_id}/fixture/tables",
			Pattern: "GET /api/v1/incidents/{incident_id}/fixture/tables", SuccessStatuses: []int{200}, Security: [][]string{{"bearerSession"}, {"sessionCookie"}},
			OwnerGate: true, ErrorRegistry: "contracts/fixture/errors.json", PathParameters: []networkFlowPathParameter{incident}},
		{RouteID: "fx.tables.patch", OwnerID: "module.fixture", Availability: "fixture_profile", Method: "PATCH", Path: "/api/v1/incidents/{incident_id}/fixture/tables/{table_id}",
			Pattern: "PATCH /api/v1/incidents/{incident_id}/fixture/tables/{table_id}", StateChanging: true, SuccessStatuses: []int{200},
			Security: [][]string{{"bearerSession"}, {"csrfCookie", "csrfHeader", "sessionCookie"}}, OwnerGate: true, ErrorRegistry: "contracts/fixture/errors.json",
			PathParameters: []networkFlowPathParameter{incident, table}},
	}
	if !reflect.DeepEqual(routes, want) {
		t.Fatalf("projection = %#v\nwant %#v", routes, want)
	}
	generated, err := renderNetworkFlow(routes)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		`{RouteID: "fx.tables.list", OwnerID: "module.fixture", Availability: "fixture_profile", Method: "GET"`,
		`StateChanging: true, SuccessStatuses: []int{200}, Security: [][]string{[]string{"bearerSession"}, []string{"csrfCookie", "csrfHeader", "sessionCookie"}}, OwnerGate: true, ErrorRegistry: "contracts/fixture/errors.json"`,
		`{Name: "incident_id", Format: "uuid", Pattern: "", AfterOwnerGate: false, NotFoundCode: "incident_not_found", NotFoundStatus: 404, NotFoundDetails: map[string]any{}}`,
		`{Name: "table_id", Format: "", Pattern: "^fxt_[a-f0-9]{32}$", AfterOwnerGate: true, NotFoundCode: "fixture_table_not_found", NotFoundStatus: 404, NotFoundDetails: map[string]any{"allowed_states": []any{"active"}, "reason_code": "not_found", "table_id": nil}}`,
		"func cloneObject(object map[string]any) map[string]any",
	} {
		if !bytes.Contains(generated, []byte(fragment)) {
			t.Fatalf("generated catalog lacks %q:\n%s", fragment, generated)
		}
	}

	parameter := func(routes map[string]any, name string) map[string]any {
		return routes["path_parameters"].(map[string]any)[name].(map[string]any)
	}
	route := func(routes map[string]any, index int) map[string]any {
		return routes["routes"].([]any)[index].(map[string]any)
	}
	for _, tc := range []struct {
		name      string
		edit      func(routes, policy, schemas, nfErrors map[string]any)
		wantError string
	}{
		{"undeclared placeholder", func(routes, _, _, _ map[string]any) {
			delete(routes["path_parameters"].(map[string]any), "table_id")
		}, `path parameter "table_id" has no declaration`},
		{"unused declaration", func(routes, _, _, _ map[string]any) {
			routes["path_parameters"].(map[string]any)["ghost_id"] = map[string]any{"format": "uuid", "admission_stage": "resource_path", "not_found": map[string]any{"code": "incident_not_found", "http_status": 404, "details": map[string]any{}}}
		}, `"ghost_id" is declared but no route uses it`},
		{"unknown schema definition", func(routes, _, _, _ map[string]any) { parameter(routes, "table_id")["schema_def"] = "Missing" }, "is not a string pattern definition"},
		{"unanchored pattern", func(routes, _, _, _ map[string]any) { parameter(routes, "table_id")["schema_def"] = "Loose" }, "must be anchored"},
		{"format and schema", func(routes, _, _, _ map[string]any) { parameter(routes, "table_id")["format"] = "uuid" }, "exactly one of format and schema_def"},
		{"unsupported format", func(routes, _, _, _ map[string]any) { parameter(routes, "incident_id")["format"] = "date" }, `format "date" is not supported`},
		{"unregistered code", func(routes, _, _, _ map[string]any) {
			parameter(routes, "table_id")["not_found"].(map[string]any)["code"] = "fixture_row_failure"
		}, "is not registered"},
		{"status mismatch", func(routes, _, _, _ map[string]any) {
			parameter(routes, "table_id")["not_found"].(map[string]any)["http_status"] = 409
		}, "is registered at 404, not 409"},
		{"missing details", func(routes, _, _, _ map[string]any) {
			delete(parameter(routes, "table_id")["not_found"].(map[string]any), "details")
		}, "details must be an object"},
		{"unknown stage", func(routes, _, _, _ map[string]any) { parameter(routes, "table_id")["admission_stage"] = "nowhere" }, "is not an operation-policy stage"},
		{"stage before the gate", func(routes, _, _, _ map[string]any) {
			parameter(routes, "table_id")["admission_stage"] = "session_and_csrf"
		}, "precedes the incident_visibility_role_lifecycle stage"},
		{"stages out of template order", func(routes, _, _, _ map[string]any) {
			parameter(routes, "incident_id")["admission_stage"] = "resource_path"
			parameter(routes, "table_id")["admission_stage"] = "incident_visibility_role_lifecycle"
		}, "precedes an earlier admission stage"},
		{"CSRF on a read", func(routes, _, _, _ map[string]any) {
			route(routes, 1)["security"] = []any{[]any{"bearerSession"}, []any{"csrfCookie", "csrfHeader", "sessionCookie"}}
		}, "cookie CSRF is declared=true but the operation policy requires false"},
		{"no CSRF on a mutation", func(routes, _, _, _ map[string]any) {
			route(routes, 0)["security"] = []any{[]any{"bearerSession"}, []any{"sessionCookie"}}
		}, "cookie CSRF is declared=false but the operation policy requires true"},
		{"anonymous alternative", func(routes, _, _, _ map[string]any) {
			route(routes, 1)["security"] = []any{[]any{}, []any{"sessionCookie"}}
		}, "anonymous alternative"},
		{"unsorted alternative", func(routes, _, _, _ map[string]any) {
			route(routes, 0)["security"] = []any{[]any{"bearerSession"}, []any{"sessionCookie", "csrfCookie", "csrfHeader"}}
		}, "must be sorted"},
		{"unknown scheme", func(routes, _, _, _ map[string]any) {
			route(routes, 1)["security"] = []any{[]any{"apiKey"}}
		}, `unknown scheme "apiKey"`},
		{"route without policy", func(_, policy, _, _ map[string]any) {
			policy["operations"] = policy["operations"].([]any)[:1]
		}, `route "fx.tables.patch" has no operation policy`},
		{"policy outside the contract", func(_, policy, _, _ map[string]any) {
			policy["operations"] = append(policy["operations"].([]any), map[string]any{"route_id": "fx.ghost", "csrf_required": false})
		}, "operation policy names routes outside the route contract"},
		{"wrong schema", func(routes, _, _, _ map[string]any) {
			routes["schema_id"] = "cartulary.network_flow_route_contracts.v2"
		}, "unexpected schema_id"},
	} {
		if _, err := projectNetworkFlowRoutes(networkFlowFixtureSources(t, tc.edit)); err == nil || !strings.Contains(err.Error(), tc.wantError) {
			t.Errorf("%s: error = %v, want substring %q", tc.name, err, tc.wantError)
		}
	}

	t.Run("the repository contract", func(t *testing.T) {
		routes, err := loadNetworkFlowRoutes("../..")
		if err != nil {
			t.Fatal(err)
		}
		if len(routes) != 19 {
			t.Fatalf("route count = %d, want 19", len(routes))
		}
		for _, route := range routes {
			if route.OwnerID != "module.networkflow" || route.Availability != "network_flow_activity" || !route.OwnerGate || route.ErrorRegistry != "contracts/network-flow/errors.v1.json" {
				t.Fatalf("%s projection = %#v", route.RouteID, route)
			}
			if route.PathParameters[0].Name != "incident_id" || route.PathParameters[0].AfterOwnerGate {
				t.Fatalf("%s does not admit incident_id before its owner gate: %#v", route.RouteID, route.PathParameters)
			}
			for _, parameter := range route.PathParameters[1:] {
				if !parameter.AfterOwnerGate || parameter.Pattern == "" {
					t.Fatalf("%s resource parameter %#v is not after the owner gate", route.RouteID, parameter)
				}
			}
		}
	})
}
