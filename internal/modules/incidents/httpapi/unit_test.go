package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestIncidentAccessDecisionUsesNotFoundForMissingMembershipAndDeniedForInsufficientRole_Unit(t *testing.T) {
	requireErrorContract(t, "incident_not_found", http.StatusNotFound)
	_, readDenied := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialNotVisible}, "")
	if readDenied == nil || readDenied.Status != http.StatusNotFound || readDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for read access: %#v", readDenied)
	}

	_, writeDenied := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialNotVisible}, "reviewer|admin")
	if writeDenied == nil || writeDenied.Status != http.StatusNotFound || writeDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for write access: %#v", writeDenied)
	}

	_, guard := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialInsufficientRole}, "reviewer|admin")
	if guard == nil || guard.Status != http.StatusForbidden || guard.Code != "authorization_denied" {
		t.Fatalf("insufficient incident role must keep authorization_denied: %#v", guard)
	}
}

func TestRegisterRoutesRequiresExactDependencies(t *testing.T) {
	t.Run("registered route matrix", testRegisteredIncidentRoutes)
	t.Run("supplied clock and codec", testIncidentSuppliedDependencies)
	t.Run("missing supplied dependencies", func(t *testing.T) {
		for _, missing := range []string{"clock", "cursor codec"} {
			t.Run(missing, func(t *testing.T) {
				deps, owner := incidentRouteTestDependencies()
				registry, err := httpapi.NewRouteRegistry(nil)
				if err != nil {
					t.Fatal(err)
				}
				deps.PublicRoutes = registry
				if missing == "clock" {
					deps.Now = nil
				} else {
					deps.CursorCodec = nil
				}
				err = RegisterRoutes(owner)(http.NewServeMux(), deps)
				if err == nil || !strings.Contains(err.Error(), missing+" is required") {
					t.Fatalf("registration error = %v", err)
				}
				if len(registry.Snapshot()) != 0 {
					t.Fatal("missing dependency partially registered routes")
				}
			})
		}
	})
	validApplication := &incidents.Application{}
	validAdmission := &admission.Checker{}
	var typedNilApplication *incidents.Application
	var typedNilAdmission *admission.Checker
	var typedNilCoordinator *terminalMutationCoordinatorStub

	tests := []struct {
		name         string
		dependencies Dependencies
		wantError    string
	}{
		{name: "missing application", dependencies: Dependencies{}, wantError: "application is required"},
		{name: "typed nil application", dependencies: Dependencies{Application: typedNilApplication}, wantError: "application is required"},
		{name: "missing admission", dependencies: Dependencies{Application: validApplication}, wantError: "admission checker is required"},
		{name: "typed nil admission", dependencies: Dependencies{Application: validApplication, AdmissionChecker: typedNilAdmission}, wantError: "admission checker is required"},
		{name: "missing coordinator", dependencies: Dependencies{Application: validApplication, AdmissionChecker: validAdmission}, wantError: "terminal mutation coordinator is required"},
		{name: "typed nil coordinator", dependencies: Dependencies{Application: validApplication, AdmissionChecker: validAdmission, TerminalMutationCoordinator: typedNilCoordinator}, wantError: "terminal mutation coordinator is required"},
		{name: "missing PostgreSQL", dependencies: Dependencies{Application: validApplication, AdmissionChecker: validAdmission, TerminalMutationCoordinator: &terminalMutationCoordinatorStub{}}, wantError: "PostgreSQL dependency is required for membership audit reading"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := RegisterRoutes(test.dependencies)(http.NewServeMux(), httpapi.DependencySet{})
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("route registration error = %v, want %q", err, test.wantError)
			}
		})
	}
}

type terminalMutationCoordinatorStub struct{}

func (*terminalMutationCoordinatorStub) CoordinateIncidentLifecycle(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	incidents.IncidentLifecycleAdmission,
	string,
) (incidents.IncidentLifecycleResult, error) {
	return incidents.IncidentLifecycleResult{}, nil
}

func (*terminalMutationCoordinatorStub) CoordinateMembershipDeletion(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	uuid.UUID,
	incidents.MembershipDeleteAdmission,
	string,
) (incidents.MembershipDeleteResult, error) {
	return incidents.MembershipDeleteResult{}, nil
}

// Embedding an unimplemented port makes any unexpected database access fail.
// Unauthenticated, malformed-path and unsupported-method cases need no storage.
type routeRegistrationDB struct{ postgres.DB }

func incidentRouteTestDependencies() (httpapi.DependencySet, Dependencies) {
	env := map[string]string{}
	configtest.EnsureCryptographicTestEnvironment(env)
	return httpapi.DependencySet{
		PostgresDB: &routeRegistrationDB{}, Env: env,
		Now:         func() time.Time { return time.Date(2031, 2, 3, 4, 5, 6, 0, time.UTC) },
		CursorCodec: pagination.NewCodec([32]byte{17}),
	}, Dependencies{Application: &incidents.Application{}, AdmissionChecker: &admission.Checker{}, TerminalMutationCoordinator: &terminalMutationCoordinatorStub{}}
}

func testIncidentSuppliedDependencies(t *testing.T) {
	deps, owner := incidentRouteTestDependencies()
	service, err := newService(deps, owner)
	if err != nil {
		t.Fatal(err)
	}
	if service.cursorCodec != deps.CursorCodec || !service.now().Equal(deps.Now()) {
		t.Fatal("route construction replaced the supplied codec or clock")
	}
}

func testRegisteredIncidentRoutes(t *testing.T) {
	deps, owner := incidentRouteTestDependencies()
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	deps.PublicRoutes = registry
	mux := http.NewServeMux()
	if err := RegisterRoutes(owner)(mux, deps); err != nil {
		t.Fatal(err)
	}
	if got := len(registry.Snapshot()); got != 11 {
		t.Fatalf("registered operations = %d, want 11", got)
	}
	const id = "00000000-0000-0000-0000-000000000701"
	const user = "00000000-0000-0000-0000-000000000702"
	const root = "/api/v1/incidents"
	const member = root + "/" + id
	operations := []struct{ id, method, path string }{
		{"listVisibleIncidents", "GET", root}, {"createIncident", "POST", root},
		{"getIncident", "GET", member}, {"patchIncident", "PATCH", member},
		{"closeIncident", "POST", member + "/close"}, {"reopenIncident", "POST", member + "/reopen"},
		{"listIncidentMemberships", "GET", member + "/memberships"}, {"createIncidentMembership", "POST", member + "/memberships"},
		{"patchIncidentMembership", "PATCH", member + "/memberships/" + user}, {"deleteIncidentMembership", "DELETE", member + "/memberships/" + user},
		{"listIncidentMembershipAuditEvents", "GET", member + "/membership-audit-events"},
	}
	check := func(t *testing.T, method, path string, status int, code string) {
		t.Helper()
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader("{")))
		if response.Code != status {
			t.Fatalf("%s %s: status=%d body=%s, want %d", method, path, response.Code, response.Body, status)
		}
		if code != "" {
			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != code {
				t.Fatalf("code=%s, want %s", body.Error.Code, code)
			}
		}
	}
	for _, operation := range operations {
		t.Run(operation.id, func(t *testing.T) {
			check(t, operation.method, operation.path, http.StatusUnauthorized, "session_required")
			check(t, "PUT", operation.path, http.StatusMethodNotAllowed, "")
			if operation.method == "GET" {
				check(t, "HEAD", operation.path, http.StatusMethodNotAllowed, "")
			}
			if strings.Contains(operation.path, id) {
				check(t, operation.method, strings.Replace(operation.path, id, "invalid", 1), http.StatusNotFound, "")
				check(t, operation.method, strings.Replace(operation.path, id, id+"%2Fextra", 1), http.StatusNotFound, "")
			}
			if strings.Contains(operation.path, user) {
				check(t, operation.method, strings.Replace(operation.path, user, "invalid", 1), http.StatusNotFound, "")
			}
			check(t, operation.method, operation.path+"/", http.StatusNotFound, "")
			extraStatus := http.StatusNotFound
			if operation.id == "createIncident" || operation.id == "listIncidentMemberships" || operation.id == "createIncidentMembership" {
				extraStatus = http.StatusMethodNotAllowed
			}
			check(t, operation.method, operation.path+"/extra", extraStatus, "")
		})
	}
	t.Run("query precedence", func(t *testing.T) {
		check(t, "GET", root+"?limit=bad", http.StatusUnauthorized, "session_required")
		check(t, "GET", member+"/memberships?limit=bad", http.StatusUnauthorized, "session_required")
		check(t, "GET", member+"/membership-audit-events?limit=bad", http.StatusUnauthorized, "session_required")
		check(t, "GET", member+"?limit=1", http.StatusBadRequest, "invalid_pagination_request")
	})
	t.Run("encoded separators cannot select another operation", func(t *testing.T) {
		check(t, "GET", member+"%2Fmemberships", http.StatusNotFound, "")
		check(t, "GET", member+"%2Fmembership-audit-events", http.StatusNotFound, "")
		check(t, "PATCH", member+"%2Fmemberships%2F"+user, http.StatusNotFound, "")
	})
}
