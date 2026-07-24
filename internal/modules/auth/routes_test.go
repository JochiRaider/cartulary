package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestEnterpriseRouteRegistrationRequiresExactAdmission_Unit(t *testing.T) {
	t.Parallel()

	baseMux := http.NewServeMux()
	if err := RegisterRoutes()(baseMux, httpapi.DependencySet{}); err != nil {
		t.Fatal(err)
	}
	assertRoutePattern(t, baseMux, "/api/v1/auth/login", "/api/v1/auth/login")
	assertRoutePattern(t, baseMux, "/api/v1/auth/providers", "")
	assertRoutePattern(t, baseMux, "/api/v1/auth/oidc/provider/begin", "")
	assertRoutePattern(t, baseMux, "/api/v1/auth/saml/provider/begin", "")

	enterpriseMux := http.NewServeMux()
	if err := RegisterEnterpriseRoutes()(enterpriseMux, httpapi.DependencySet{}); err != nil {
		t.Fatal(err)
	}
	assertRoutePattern(t, enterpriseMux, "/api/v1/auth/providers", "/api/v1/auth/providers")
	assertRoutePattern(t, enterpriseMux, "/api/v1/auth/oidc/provider/begin", "/api/v1/auth/oidc/")
	assertRoutePattern(t, enterpriseMux, "/api/v1/auth/saml/provider/begin", "/api/v1/auth/saml/")
}

func TestEnterpriseAuthBindingsUseNarrowRouteAdmission_Unit(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodGet, "/api/v1/users/10000000-0000-0000-0000-000000000001/auth-bindings", nil)
	if err != nil {
		t.Fatal(err)
	}
	baseMux := http.NewServeMux()
	if err := RegisterRoutes()(baseMux, httpapi.DependencySet{}); err != nil {
		t.Fatal(err)
	}
	baseResponse := httptest.NewRecorder()
	baseMux.ServeHTTP(baseResponse, request)
	if baseResponse.Code != http.StatusNotFound {
		t.Fatalf("unadmitted auth bindings status = %d; want 404", baseResponse.Code)
	}

	admittedMux := http.NewServeMux()
	if err := RegisterRoutes(WithEnterpriseAuthBindings())(admittedMux, httpapi.DependencySet{}); err != nil {
		t.Fatal(err)
	}
	admittedResponse := httptest.NewRecorder()
	admittedMux.ServeHTTP(admittedResponse, request.Clone(request.Context()))
	if admittedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("admitted auth bindings status = %d; want 401", admittedResponse.Code)
	}
}

func assertRoutePattern(t testing.TB, mux *http.ServeMux, path string, want string) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, got := mux.Handler(request)
	if got != want {
		t.Fatalf("route %q pattern = %q; want %q", path, got, want)
	}
}
