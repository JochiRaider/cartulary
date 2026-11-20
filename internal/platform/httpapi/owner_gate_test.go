package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	gateScopeID = "00000000-0000-4000-8000-0000000000d1"
	gateItemID  = "it_0a1b"
	gateOwner   = "module.gatefixture"
)

type gateMarkerKey struct{}

func gateOperation(id, method, path string, security [][]string, parameters ...PathParameter) ContractOperation {
	return ContractOperation{
		OwnerID: gateOwner, Method: method, PathTemplate: path, Pattern: method + " " + path, OperationID: id,
		Availability: BaseOperationAvailability, SuccessStatuses: []int{http.StatusOK}, Security: security,
		PathParameters: parameters, OwnerGate: true, ErrorCodeRegistry: "contracts/gatefixture/errors.json",
	}
}

var (
	gateScopeParam = PathParameter{Name: "scope_id", Format: "uuid", NotFoundCode: "scope_not_found", NotFoundStatus: http.StatusNotFound}
	gateItemParam  = PathParameter{Name: "item_id", Pattern: "^it_[a-f0-9]{4}$", NotFoundCode: "item_not_found", NotFoundStatus: http.StatusNotFound, AfterOwnerGate: true,
		NotFoundDetails: map[string]any{"allowed_states": []any{"active"}, "item_id": nil, "reason_code": "not_found"}}
)

type gateFixture struct {
	handler http.Handler
	events  []string
}

// newGateFixture binds an owner-ordered owner behind NewHandler with the
// declared-security and undeclared-code guards enforcing. Its gate denies a
// scope by X-Fixture-Scope (hidden: 404, closed: 409) and otherwise marks the
// request; it records whether the item parameter was already admitted.
func newGateFixture(t *testing.T, mismatches *[]SecurityMismatch, undeclared *[]string) *gateFixture {
	t.Helper()
	operations := []ContractOperation{
		gateOperation("getItem", http.MethodGet, "/api/v1/scopes/{scope_id}/items/{item_id}", admissionReadSecurity, gateScopeParam, gateItemParam),
		gateOperation("patchItem", http.MethodPatch, "/api/v1/scopes/{scope_id}/items/{item_id}", admissionWriteSecurity, gateScopeParam, gateItemParam),
		gateOperation("listItems", http.MethodGet, "/api/v1/scopes/{scope_id}/items", admissionReadSecurity, gateScopeParam),
	}
	registry, err := newRouteRegistry(operations, nil)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &gateFixture{}
	authenticator := &fakeAuthenticator{observe: true, events: &fixture.events}
	gate := func(r *http.Request, operationID string) (*http.Request, *APIError) {
		_, itemAdmitted := PathString(r, "item_id")
		scope, scopeAdmitted := PathUUID(r, "scope_id")
		fixture.events = append(fixture.events, "gate:"+operationID)
		if r.Context().Value(admissionPrincipalKey{}) == nil || !scopeAdmitted || scope != uuid.MustParse(gateScopeID) || itemAdmitted {
			t.Errorf("gate saw principal=%v scope=%v/%t item admitted=%t", r.Context().Value(admissionPrincipalKey{}), scope, scopeAdmitted, itemAdmitted)
		}
		switch r.Header.Get("X-Fixture-Scope") {
		case "hidden":
			return r, &APIError{Status: http.StatusNotFound, Code: "scope_not_found", Details: map[string]any{}}
		case "closed":
			return r, &APIError{Status: http.StatusConflict, Code: "scope_closed", Details: map[string]any{}}
		}
		return r.WithContext(context.WithValue(r.Context(), gateMarkerKey{}, operationID)), nil
	}
	record := func(id string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fixture.events = append(fixture.events, "handler:"+id)
			if r.Context().Value(gateMarkerKey{}) != id {
				t.Errorf("%s ran without its gate", id)
			}
			if scope, ok := PathUUID(r, "scope_id"); !ok || scope != uuid.MustParse(gateScopeID) {
				t.Errorf("%s scope = %v %t", id, scope, ok)
			}
			if id != "listItems" {
				if item, ok := PathString(r, "item_id"); !ok || item != gateItemID {
					t.Errorf("%s item = %q %t", id, item, ok)
				}
			}
			// An owner-registry code is not a REQ-01-702 declaration.
			if r.Header.Get("X-Fixture-Owner-Error") == "ok" {
				_ = WriteError(w, r, http.StatusConflict, "item_not_active", "item_not_active", map[string]any{})
				return
			}
			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"operation": id})
		}
	}
	options := Options{
		SecurityGuard: &SecurityGuard{Enforce: true, Report: func(mismatch SecurityMismatch) { *mismatches = append(*mismatches, mismatch) }},
		ErrorCodeGuard: &ErrorCodeGuard{Enforce: true, Report: func(operationID string, status int, code string) {
			*undeclared = append(*undeclared, operationID+" "+code)
		}, Observe: func(operationID string, status int, code string) {
			*undeclared = append(*undeclared, "observed "+operationID+" "+code)
		}},
	}
	options.Dependencies = testExtensionDependenciesWith(t, DependencySet{PublicRoutes: registry, Authenticator: authenticator}, nil)
	options.AdditionalRoutes = []RouteRegistrar{func(mux *http.ServeMux, deps DependencySet) error {
		return BindAdmittedOwnerOrderRoutes(mux, deps, OwnerOrder{OwnerID: gateOwner, Gate: gate, Handlers: map[string]http.HandlerFunc{
			"getItem": record("getItem"), "patchItem": record("patchItem"), "listItems": record("listItems"),
		}})
	}}
	fixture.handler, err = NewHandler(options)
	if err != nil {
		t.Fatal(err)
	}
	if modes := registry.AdmissionModes(); !reflect.DeepEqual(modes, map[string]AdmissionMode{"getItem": AdmissionCentral, "patchItem": AdmissionCentral, "listItems": AdmissionCentral}) {
		t.Fatalf("owner-order admission modes = %v", modes)
	}
	return fixture
}

func (fixture *gateFixture) serve(method, target string, headers map[string]string) *httptest.ResponseRecorder {
	fixture.events = nil
	request := httptest.NewRequest(method, target, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	return recorder
}

// TestCentralAdmissionOwnerGateOrder covers Core 01 REQ-01-707's family clause
// as Network Flow's Table 21-C uses it: route resolution, then authentication,
// then the path parameters before the owner gate, then the gate, then the
// parameters after it, then the handler. A malformed resource identifier is
// therefore not revealed to a caller the gate denies; the not-found details
// are JSON values from the projection; the guards accept an owner-registry
// operation without REQ-01-702 declarations; binding refuses a gate-less owner
// order, a gated operation bound without its gate, and a gate for an operation
// that declares none; and the public registry carries the Network Flow route
// contract outside the OpenAPI catalog.
func TestCentralAdmissionOwnerGateOrder(t *testing.T) {
	var mismatches []SecurityMismatch
	var guarded []string
	fixture := newGateFixture(t, &mismatches, &guarded)
	authenticated := map[string]string{"X-Fixture-Auth": "ok"}
	with := func(extra map[string]string) map[string]string {
		headers := map[string]string{"X-Fixture-Auth": "ok"}
		for key, value := range extra {
			headers[key] = value
		}
		return headers
	}
	item := "/api/v1/scopes/" + gateScopeID + "/items/"

	t.Run("route resolution precedes authentication and the gate", func(t *testing.T) {
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodDelete, item+gateItemID, nil), "GET", "PATCH")
		if recorder := fixture.serve(http.MethodHead, item+gateItemID, nil); recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET, PATCH" || recorder.Body.Len() != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q", recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
		}
		requireAdmissionError(t, fixture.serve(http.MethodGet, item+gateItemID+"/unknown", nil), http.StatusNotFound, "route_not_found", map[string]any{})
		if len(fixture.events) != 0 {
			t.Fatalf("a route-resolution rejection ran admission: %v", fixture.events)
		}
	})
	t.Run("authentication precedes every parameter and the gate", func(t *testing.T) {
		requireAdmissionError(t, fixture.serve(http.MethodGet, "/api/v1/scopes/malformed/items/malformed", nil), http.StatusUnauthorized, "session_required", map[string]any{})
		if !reflect.DeepEqual(fixture.events, []string{"authenticate"}) {
			t.Fatalf("events = %v", fixture.events)
		}
	})
	t.Run("a pre-gate parameter precedes the gate", func(t *testing.T) {
		requireAdmissionError(t, fixture.serve(http.MethodGet, "/api/v1/scopes/malformed/items/malformed", authenticated), http.StatusNotFound, "scope_not_found", map[string]any{})
		requireAdmissionError(t, fixture.serve(http.MethodGet, "/api/v1/scopes/"+strings.ReplaceAll(gateScopeID, "-", "")+"/items", authenticated), http.StatusNotFound, "scope_not_found", map[string]any{})
		if !reflect.DeepEqual(fixture.events, []string{"authenticate"}) {
			t.Fatalf("events = %v", fixture.events)
		}
	})
	t.Run("the gate precedes a post-gate parameter", func(t *testing.T) {
		for _, value := range []string{"malformed", "it_0A1B", gateItemID + "%2Fx"} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, item+value, with(map[string]string{"X-Fixture-Scope": "hidden"})), http.StatusNotFound, "scope_not_found", map[string]any{})
			requireAdmissionError(t, fixture.serve(http.MethodPatch, item+value, with(map[string]string{"X-Fixture-Scope": "closed"})), http.StatusConflict, "scope_closed", map[string]any{})
			if !reflect.DeepEqual(fixture.events, []string{"authenticate", "gate:patchItem"}) {
				t.Fatalf("events = %v", fixture.events)
			}
		}
	})
	t.Run("a post-gate parameter receives its declared JSON details", func(t *testing.T) {
		for _, value := range []string{"malformed", "it_0A1B", gateItemID + "%2Fx", gateItemID + "0"} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, item+value, authenticated), http.StatusNotFound, "item_not_found",
				map[string]any{"allowed_states": []any{"active"}, "item_id": nil, "reason_code": "not_found"})
			if !reflect.DeepEqual(fixture.events, []string{"authenticate", "gate:getItem"}) {
				t.Fatalf("events = %v", fixture.events)
			}
		}
	})
	t.Run("an admitted request reaches its handler after its gate", func(t *testing.T) {
		for _, tc := range []struct {
			method, path, operationID string
		}{
			{http.MethodGet, item + gateItemID, "getItem"},
			{http.MethodPatch, item + gateItemID, "patchItem"},
			{http.MethodGet, "/api/v1/scopes/" + strings.ToUpper(gateScopeID) + "/items", "listItems"},
		} {
			if recorder := fixture.serve(tc.method, tc.path, authenticated); recorder.Code != http.StatusOK {
				t.Fatalf("%s %s = %d %s", tc.method, tc.path, recorder.Code, recorder.Body)
			}
			if want := []string{"authenticate", "gate:" + tc.operationID, "handler:" + tc.operationID}; !reflect.DeepEqual(fixture.events, want) {
				t.Fatalf("events = %v, want %v", fixture.events, want)
			}
		}
	})
	t.Run("the guards accept an owner-registry operation", func(t *testing.T) {
		if recorder := fixture.serve(http.MethodGet, item+gateItemID, with(map[string]string{"X-Fixture-Owner-Error": "ok"})); recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "item_not_active") {
			t.Fatalf("owner-registry code = %d %s, want it unchanged", recorder.Code, recorder.Body)
		}
		if len(mismatches) != 0 || len(guarded) != 0 {
			t.Fatalf("guards reported mismatches=%v undeclared or observed=%v", mismatches, guarded)
		}
	})
	t.Run("binding enforces the declared owner order", func(t *testing.T) {
		registry, err := newRouteRegistry(append(admissionOperations(),
			gateOperation("getItem", http.MethodGet, "/api/v1/scopes/{scope_id}/items/{item_id}", admissionReadSecurity, gateScopeParam, gateItemParam)), nil)
		if err != nil {
			t.Fatal(err)
		}
		deps := DependencySet{PublicRoutes: registry, Authenticator: &fakeAuthenticator{}}
		handler := func(http.ResponseWriter, *http.Request) {}
		gate := func(r *http.Request, _ string) (*http.Request, *APIError) { return r, nil }
		for _, tc := range []struct {
			name string
			bind func() error
			want string
		}{
			{"a gated operation bound without its gate", func() error {
				return BindAdmittedOwnerRoutes(http.NewServeMux(), deps, gateOwner, map[string]http.HandlerFunc{"getItem": handler})
			}, `"getItem" requires its owner gate`},
			{"a gate for an operation that declares none", func() error {
				return BindAdmittedOwnerOrderRoutes(http.NewServeMux(), deps, OwnerOrder{OwnerID: admissionOwner, Gate: gate, Handlers: map[string]http.HandlerFunc{"getThing": handler}})
			}, `"getThing" declares no owner gate`},
			{"an owner order without its gate", func() error {
				return BindAdmittedOwnerOrderRoutes(http.NewServeMux(), deps, OwnerOrder{OwnerID: gateOwner, Handlers: map[string]http.HandlerFunc{"getItem": handler}})
			}, "without its gate"},
		} {
			if err := tc.bind(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s: bind error = %v, want %q", tc.name, err, tc.want)
			}
		}
		if modes := registry.AdmissionModes(); len(modes) != 0 {
			t.Fatalf("refused bindings were recorded: %v", modes)
		}
		if _, err := newRouteRegistry([]ContractOperation{{
			OwnerID: gateOwner, Method: http.MethodGet, PathTemplate: "/api/v1/scopes/{scope_id}", Pattern: "GET /api/v1/scopes/{scope_id}", OperationID: "inconsistent",
			Availability: BaseOperationAvailability, SuccessStatuses: []int{http.StatusOK}, PathParameters: []PathParameter{{Name: "scope_id", Format: "uuid", AfterOwnerGate: true}},
		}}, nil); err == nil || !strings.Contains(err.Error(), "after an owner gate it does not declare") {
			t.Fatalf("inconsistent catalog error = %v", err)
		}
	})
	t.Run("the public registry carries the Network Flow route contract", func(t *testing.T) {
		registry, err := NewRouteRegistry(nil)
		if err != nil {
			t.Fatal(err)
		}
		networkFlow := 0
		for _, operation := range registry.operationsByID {
			if operation.OwnerID != "module.networkflow" {
				if operation.OwnerGate || operation.ErrorCodeRegistry != "" {
					t.Fatalf("OpenAPI operation %s declares an owner order or registry", operation.OperationID)
				}
				continue
			}
			networkFlow++
			if !operation.OwnerGate || operation.ErrorCodeRegistry == "" || operation.Availability != "network_flow_activity" || len(operation.ErrorCodes) != 0 {
				t.Fatalf("Network Flow operation %s = %#v", operation.OperationID, operation)
			}
			if first := operation.PathParameters[0]; first.Name != "incident_id" || first.Format != "uuid" || first.AfterOwnerGate || first.NotFoundCode != "incident_not_found" {
				t.Fatalf("%s incident_id = %#v", operation.OperationID, first)
			}
			for _, parameter := range operation.PathParameters[1:] {
				if !parameter.AfterOwnerGate || parameter.Pattern == "" || !strings.HasPrefix(parameter.NotFoundCode, "network_flow_") {
					t.Fatalf("%s resource parameter = %#v", operation.OperationID, parameter)
				}
			}
		}
		if networkFlow != 19 {
			t.Fatalf("registry Network Flow operations = %d, want 19", networkFlow)
		}
		for _, operation := range ContractOperations() {
			if operation.OwnerID == "module.networkflow" {
				t.Fatalf("the OpenAPI catalog lists Network Flow operation %s", operation.OperationID)
			}
		}
	})
}
