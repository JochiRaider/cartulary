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
	assertRoutePattern(t, baseMux, http.MethodPost, "/api/v1/auth/login", "POST /api/v1/auth/login")
	assertRoutePattern(t, baseMux, http.MethodGet, "/api/v1/auth/providers", "")
	assertRoutePattern(t, baseMux, http.MethodPost, "/api/v1/auth/providers/provider/begin", "")
	assertRoutePattern(t, baseMux, http.MethodGet, "/api/v1/auth/oidc/provider/callback", "")
	assertRoutePattern(t, baseMux, http.MethodPost, "/api/v1/auth/saml/provider/acs", "")

	enterpriseMux := http.NewServeMux()
	if err := RegisterEnterpriseRoutes()(enterpriseMux, httpapi.DependencySet{}); err != nil {
		t.Fatal(err)
	}
	assertRoutePattern(t, enterpriseMux, http.MethodGet, "/api/v1/auth/providers", "GET /api/v1/auth/providers")
	assertRoutePattern(t, enterpriseMux, http.MethodPost, "/api/v1/auth/providers/provider/begin", "POST /api/v1/auth/providers/{provider_key}/begin")
	assertRoutePattern(t, enterpriseMux, http.MethodGet, "/api/v1/auth/oidc/provider/callback", "GET /api/v1/auth/oidc/{provider_key}/callback")
	assertRoutePattern(t, enterpriseMux, http.MethodPost, "/api/v1/auth/saml/provider/acs", "POST /api/v1/auth/saml/{provider_key}/acs")

}

func TestEnterpriseAuthBindingsUseNarrowRouteAdmission_Unit(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodPost, "/api/v1/users/10000000-0000-0000-0000-000000000001/auth-bindings", nil)
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

func assertRoutePattern(t testing.TB, mux *http.ServeMux, method string, path string, want string) {
	t.Helper()
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, got := mux.Handler(request)
	if got != want {
		t.Fatalf("route %q pattern = %q; want %q", path, got, want)
	}
}
