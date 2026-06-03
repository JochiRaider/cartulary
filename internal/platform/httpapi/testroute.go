package httpapi

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	TestRoutesEnabledEnv       = "CARTULARY_ENABLE_TEST_ROUTES"
	TestRuntimeMarkerEnv       = "CARTULARY_TEST_RUNTIME_MARKER"
	TestRuntimeMarkerValue     = "harness-owned"
	TestRouteTokenEnv          = "CARTULARY_TEST_ROUTE_TOKEN"
	TestRuntimeAPIOriginEnv    = "CARTULARY_WEB_E2E_API_ORIGIN"
	TestRuntimePublicOriginEnv = "CARTULARY_WEB_E2E_PUBLIC_ORIGIN"
	TestRouteTokenHeader       = "X-Cartulary-Test-Route-Token"
)

var weakTestRouteTokens = map[string]struct{}{
	"test":     {},
	"token":    {},
	"secret":   {},
	"password": {},
	"changeme": {},
}

type TestRouteGuard struct {
	Token          string
	ExpectedHost   string
	AllowedOrigins map[string]struct{}
}

func TestRoutesEnabled(env map[string]string) bool {
	return lookupTestRouteEnv(env, TestRoutesEnabledEnv) == "1"
}

func NewTestRouteGuard(env map[string]string) (TestRouteGuard, error) {
	var guard TestRouteGuard
	if !TestRoutesEnabled(env) {
		return guard, fmt.Errorf("test routes require %s=1", TestRoutesEnabledEnv)
	}
	if lookupTestRouteEnv(env, TestRuntimeMarkerEnv) != TestRuntimeMarkerValue {
		return guard, fmt.Errorf("test routes require %s must be %q", TestRuntimeMarkerEnv, TestRuntimeMarkerValue)
	}
	token := lookupTestRouteEnv(env, TestRouteTokenEnv)
	if !ValidTestRouteToken(token) {
		return guard, fmt.Errorf("test routes require %s must be a visible ASCII token of length 43..512 and not a weak token", TestRouteTokenEnv)
	}
	boundary, err := resolveTestRouteBoundary(env)
	if err != nil {
		return guard, err
	}
	return TestRouteGuard{
		Token:          token,
		ExpectedHost:   boundary.expectedHost,
		AllowedOrigins: boundary.allowedOrigins,
	}, nil
}

func (g TestRouteGuard) Protect(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !g.Authorize(w, r) {
			return
		}
		handler(w, r)
	}
}

func (g TestRouteGuard) Authorize(w http.ResponseWriter, r *http.Request) bool {
	if !g.AllowedRequestBoundary(r) {
		_ = WriteError(w, r, http.StatusForbidden, "test_route_forbidden", "test route origin is forbidden", map[string]any{})
		return false
	}
	if !g.Authorized(r) {
		_ = WriteError(w, r, http.StatusForbidden, "test_route_forbidden", "test route token is required", map[string]any{})
		return false
	}
	return true
}

func (g TestRouteGuard) Authorized(r *http.Request) bool {
	got := r.Header.Get(TestRouteTokenHeader)
	if !ValidTestRouteToken(got) || len(got) != len(g.Token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(g.Token)) == 1
}

func (g TestRouteGuard) AllowedRequestBoundary(r *http.Request) bool {
	if g.ExpectedHost != "" && r.Host != g.ExpectedHost {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return len(g.AllowedOrigins) == 0
	}
	normalized, err := NormalizeTestRouteOrigin(origin)
	if err != nil {
		return false
	}
	_, ok := g.AllowedOrigins[normalized]
	return ok
}

func ValidTestRouteToken(token string) bool {
	if len(token) < 43 || len(token) > 512 {
		return false
	}
	repeated := true
	first := rune(0)
	for index, r := range token {
		if r <= ' ' || r > '~' {
			return false
		}
		if index == 0 {
			first = r
		} else if r != first {
			repeated = false
		}
	}
	if repeated {
		return false
	}
	if _, weak := weakTestRouteTokens[strings.ToLower(token)]; weak {
		return false
	}
	return true
}

func NormalizeTestRouteOrigin(value string) (string, error) {
	normalized, _, err := parseTestRouteOrigin(value)
	return normalized, err
}

func lookupTestRouteEnv(env map[string]string, key string) string {
	if env != nil {
		return env[key]
	}
	return os.Getenv(key)
}

type testRouteBoundary struct {
	expectedHost   string
	allowedOrigins map[string]struct{}
}

func resolveTestRouteBoundary(env map[string]string) (testRouteBoundary, error) {
	boundary := testRouteBoundary{allowedOrigins: map[string]struct{}{}}
	apiOrigin := lookupTestRouteEnv(env, TestRuntimeAPIOriginEnv)
	if strings.TrimSpace(apiOrigin) != "" {
		normalized, parsed, err := normalizeTestRouteConfiguredOrigin(TestRuntimeAPIOriginEnv, apiOrigin)
		if err != nil {
			return boundary, err
		}
		boundary.expectedHost = parsed.Host
		boundary.allowedOrigins[normalized] = struct{}{}
	}
	publicOrigin := lookupTestRouteEnv(env, TestRuntimePublicOriginEnv)
	if strings.TrimSpace(publicOrigin) != "" {
		normalized, _, err := normalizeTestRouteConfiguredOrigin(TestRuntimePublicOriginEnv, publicOrigin)
		if err != nil {
			return boundary, err
		}
		boundary.allowedOrigins[normalized] = struct{}{}
	}
	return boundary, nil
}

func normalizeTestRouteConfiguredOrigin(name string, value string) (string, *url.URL, error) {
	normalized, parsed, err := parseTestRouteOrigin(value)
	if err != nil {
		return "", nil, fmt.Errorf("test routes require %s must be an http(s) origin with scheme and host: %w", name, err)
	}
	return normalized, parsed, nil
}

func parseTestRouteOrigin(value string) (string, *url.URL, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", nil, errors.New("unsupported scheme")
	}
	if parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", nil, errors.New("not an origin")
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), parsed, nil
}
