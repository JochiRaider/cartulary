package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	csrfSecuredOperation = ContractOperation{OwnerID: "module.example", OperationID: "mutateExample", Availability: BaseOperationAvailability,
		Security: [][]string{{"bearerSession"}, {"csrfCookie", "csrfHeader", "sessionCookie"}}}
	sessionSecuredOperation = ContractOperation{OwnerID: "module.example", OperationID: "readExample", Availability: BaseOperationAvailability,
		Security: [][]string{{"bearerSession"}, {"sessionCookie"}}}
	publicOperation = ContractOperation{OwnerID: "module.example", OperationID: "publicExample", Availability: BaseOperationAvailability,
		Security: [][]string{}}
	bootstrapSecuredOperation = ContractOperation{OwnerID: "module.example", OperationID: "bootstrapExample", Availability: BaseOperationAvailability,
		Security: [][]string{{"bearerSession"}, {"csrfCookie", "csrfHeader", "sessionCookie"}, {"credentialBootstrapBearer"}}}
)

type guardedResponse struct {
	status   int
	code     string
	body     string
	header   http.Header
	reported []SecurityMismatch
}

// serveGuarded serves one request to handler resolved as operation behind a
// guard. before runs ahead of resolution, as an owner family policy does.
func serveGuarded(t *testing.T, enforce bool, operation ContractOperation, before func(*http.Request), handler http.HandlerFunc) guardedResponse {
	t.Helper()
	var reported []SecurityMismatch
	guard := &SecurityGuard{Enforce: enforce, Report: func(mismatch SecurityMismatch) { reported = append(reported, mismatch) }}
	resolved := withOperationMetadata(operation, handler)
	policy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if before != nil {
			before(r)
		}
		resolved(w, r)
	})
	recorder := httptest.NewRecorder()
	recorder.Header().Set(RequestIDHeader, "req-security")
	withSecurityGuard(policy, guard).ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/example", nil))
	response := guardedResponse{status: recorder.Code, body: recorder.Body.String(), header: recorder.Header(), reported: reported}
	var envelope ErrorEnvelope
	if json.Unmarshal(recorder.Body.Bytes(), &envelope) == nil {
		response.code = envelope.Error.Code
	}
	return response
}

func cookiePrincipal(csrf bool) func(*http.Request) {
	return func(r *http.Request) {
		ObserveAuthenticationAttempt(r.Context())
		if csrf {
			ObserveCSRFVerification(r.Context())
		}
		ObservePrincipal(r.Context(), AuthenticatedBySessionCookie)
	}
}

func attempt(r *http.Request) { ObserveAuthenticationAttempt(r.Context()) }

func verifyCSRF(r *http.Request) { ObserveCSRFVerification(r.Context()) }

func principalFrom(source AuthenticationSource) func(*http.Request) {
	return func(r *http.Request) {
		ObserveAuthenticationAttempt(r.Context())
		ObservePrincipal(r.Context(), source)
	}
}

func succeedAfter(observe func(*http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if observe != nil {
			observe(r)
		}
		_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"ok": true})
	}
}

func failAfter(observe func(*http.Request), status int, code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if observe != nil {
			observe(r)
		}
		_ = WriteError(w, r, status, code, code, nil)
	}
}

// TestSecurityGuardHonoursDeclarations covers Core 01 REQ-01-703 and
// TH-HARNESS-REQ-859 (TH-HARNESS-AC-156): each mismatch kind becomes 500
// internal_error with exactly one finding, and correctly declared operations
// of each kind are unchanged.
func TestSecurityGuardHonoursDeclarations(t *testing.T) {
	type finding struct {
		status             int
		declared, observed string
	}
	for _, tc := range []struct {
		name      string
		operation ContractOperation
		before    func(*http.Request)
		handler   http.HandlerFunc
		// want is the mismatch; nil means the original response is unchanged.
		want         *finding
		wantStatus   int
		wantOriginal string
	}{
		{name: "undeclared operation authenticates a cookie session", operation: publicOperation,
			handler: succeedAfter(cookiePrincipal(false)), want: &finding{200, "public", "session"}},
		{name: "undeclared operation authenticates a bearer session", operation: publicOperation,
			handler: succeedAfter(principalFrom(AuthenticatedByBearerSession)), want: &finding{200, "public", "session"}},
		{name: "undeclared operation verifies a cookie session with CSRF", operation: publicOperation,
			handler: succeedAfter(cookiePrincipal(true)), want: &finding{200, "public", "session_csrf"}},
		{name: "undeclared operation demands a session it does not get", operation: publicOperation,
			handler: failAfter(attempt, http.StatusUnauthorized, "session_required"), want: &finding{401, "public", "session"}},
		{name: "declared operation succeeds unauthenticated", operation: csrfSecuredOperation,
			handler: succeedAfter(nil), want: &finding{200, "session_csrf", "public"}},
		{name: "declared operation succeeds after a failed attempt", operation: sessionSecuredOperation,
			handler: succeedAfter(attempt), want: &finding{200, "session", "public"}},
		{name: "declared operation writes nothing unauthenticated", operation: sessionSecuredOperation,
			handler: func(http.ResponseWriter, *http.Request) {}, want: &finding{200, "session", "public"}},
		{name: "CSRF verified where the cookie alternative omits it", operation: sessionSecuredOperation,
			handler: succeedAfter(cookiePrincipal(true)), want: &finding{200, "session", "session_csrf"}},
		{name: "CSRF rejection where the cookie alternative omits it", operation: sessionSecuredOperation,
			handler: failAfter(verifyCSRF, http.StatusForbidden, "csrf_verification_failed"), want: &finding{403, "session", "session_csrf"}},
		{name: "CSRF skipped where the cookie alternative declares it", operation: csrfSecuredOperation,
			handler: succeedAfter(cookiePrincipal(false)), want: &finding{200, "session_csrf", "session"}},
		{name: "family policy skipped CSRF before resolution", operation: csrfSecuredOperation,
			before: cookiePrincipal(false), handler: succeedAfter(nil), want: &finding{200, "session_csrf", "session"}},
		{name: "credential bootstrap where undeclared", operation: csrfSecuredOperation,
			handler: succeedAfter(principalFrom(AuthenticatedByCredentialBootstrap)), want: &finding{200, "session_csrf", "credential_bootstrap"}},

		{name: "cookie session with CSRF where declared", operation: csrfSecuredOperation,
			handler: succeedAfter(cookiePrincipal(true)), wantStatus: 200},
		{name: "bearer session without CSRF where CSRF pairs the cookie", operation: csrfSecuredOperation,
			handler: succeedAfter(principalFrom(AuthenticatedByBearerSession)), wantStatus: 200},
		{name: "family policy authenticates and the operation verifies CSRF", operation: csrfSecuredOperation,
			before: cookiePrincipal(false), handler: succeedAfter(verifyCSRF), wantStatus: 200},
		{name: "cookie request refused for CSRF where declared", operation: csrfSecuredOperation,
			handler:    failAfter(verifyCSRF, http.StatusForbidden, "csrf_verification_failed"),
			wantStatus: 403, wantOriginal: "csrf_verification_failed"},
		{name: "unauthenticated request refused where declared", operation: csrfSecuredOperation,
			handler:    failAfter(attempt, http.StatusUnauthorized, "session_required"),
			wantStatus: 401, wantOriginal: "session_required"},
		{name: "cookie session without CSRF where declared without it", operation: sessionSecuredOperation,
			handler: succeedAfter(cookiePrincipal(false)), wantStatus: 200},
		{name: "bearer session where declared", operation: sessionSecuredOperation,
			handler: succeedAfter(principalFrom(AuthenticatedByBearerSession)), wantStatus: 200},
		{name: "credential bootstrap where declared", operation: bootstrapSecuredOperation,
			handler: succeedAfter(principalFrom(AuthenticatedByCredentialBootstrap)), wantStatus: 200},
		{name: "undeclared operation relies on its own capability", operation: publicOperation,
			handler: succeedAfter(nil), wantStatus: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, enforce := range []bool{true, false} {
				response := serveGuarded(t, enforce, tc.operation, tc.before, tc.handler)
				if tc.want == nil {
					if len(response.reported) != 0 || response.status != tc.wantStatus || response.code != tc.wantOriginal {
						t.Fatalf("enforce=%v: correct operation changed: %d %q reported=%v", enforce, response.status, response.code, response.reported)
					}
					continue
				}
				want := []SecurityMismatch{{OperationID: tc.operation.OperationID, Status: tc.want.status, Declared: tc.want.declared, Observed: tc.want.observed}}
				if !reflect.DeepEqual(response.reported, want) {
					t.Fatalf("enforce=%v: reported %v, want %v", enforce, response.reported, want)
				}
				if enforce && (response.status != http.StatusInternalServerError || response.code != "internal_error") {
					t.Fatalf("enforced mismatch kept its response: %d %q", response.status, response.code)
				}
				if !enforce && response.status != tc.want.status {
					t.Fatalf("observed mismatch changed the response: %d", response.status)
				}
			}
		})
	}
}

// TestSecurityGuardReportsOnceAndReplacesWholeResponse covers TH-HARNESS-REQ-859:
// a request yields at most one finding, a mismatch found after the response
// started is reported without rewriting written bytes, and an enforced
// replacement carries none of the handler's headers or body.
func TestSecurityGuardReportsOnceAndReplacesWholeResponse(t *testing.T) {
	t.Run("several mismatching facts and writes report once", func(t *testing.T) {
		response := serveGuarded(t, true, publicOperation, cookiePrincipal(true), func(w http.ResponseWriter, r *http.Request) {
			ObservePrincipal(r.Context(), AuthenticatedByBearerSession)
			w.WriteHeader(http.StatusOK)
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("first"))
			ObserveCSRFVerification(r.Context())
			_, _ = w.Write([]byte("second"))
		})
		if len(response.reported) != 1 || response.status != http.StatusInternalServerError || response.code != "internal_error" {
			t.Fatalf("got %d %q reported=%v", response.status, response.code, response.reported)
		}
	})
	t.Run("a mismatch after the response started is reported unchanged", func(t *testing.T) {
		response := serveGuarded(t, true, publicOperation, nil, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			ObservePrincipal(r.Context(), AuthenticatedBySessionCookie)
			_, _ = w.Write([]byte("written"))
			ObserveCSRFVerification(r.Context())
		})
		want := []SecurityMismatch{{OperationID: "publicExample", Status: http.StatusCreated, Declared: "public", Observed: "session"}}
		if !reflect.DeepEqual(response.reported, want) || response.status != http.StatusCreated || response.body != "written" {
			t.Fatalf("got %d %q reported=%v", response.status, response.body, response.reported)
		}
	})
	t.Run("the replacement restores the pre-handler header", func(t *testing.T) {
		response := serveGuarded(t, true, csrfSecuredOperation, nil, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Disposition", "attachment")
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("handler body"))
		})
		if response.code != "internal_error" || response.header.Get("Content-Disposition") != "" ||
			response.header.Get("Content-Type") != "application/json" || response.header.Get(RequestIDHeader) != "req-security" {
			t.Fatalf("replacement = %q %v", response.body, response.header)
		}
	})
	t.Run("an unresolved request is outside the rule", func(t *testing.T) {
		var reported []SecurityMismatch
		guard := &SecurityGuard{Enforce: true, Report: func(mismatch SecurityMismatch) { reported = append(reported, mismatch) }}
		recorder := httptest.NewRecorder()
		withSecurityGuard(succeedAfter(cookiePrincipal(true)), guard).ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws/v1/example", nil))
		if recorder.Code != http.StatusOK || len(reported) != 0 {
			t.Fatalf("got %d reported=%v", recorder.Code, reported)
		}
	})
}

// TestSecurityGuardAbsentLeavesRequestsUnchanged covers the production path of
// TH-HARNESS-REQ-859: without a guard the handler receives the original
// writer and a mismatching operation keeps its response.
func TestSecurityGuardAbsentLeavesRequestsUnchanged(t *testing.T) {
	recorder := httptest.NewRecorder()
	withOperationMetadata(publicOperation, func(w http.ResponseWriter, r *http.Request) {
		if w != recorder {
			t.Error("unguarded handler received a wrapped writer")
		}
		cookiePrincipal(true)(r)
		_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"ok": true})
	})(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/example", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("unguarded response = %d %s", recorder.Code, recorder.Body.String())
	}
}

// TestSecurityGuardOptionGuardsBoundOperations covers the Options wiring of
// TH-HARNESS-REQ-859 for a catalog operation bound through the route
// registry, with and without the guard.
func TestSecurityGuardOptionGuardsBoundOperations(t *testing.T) {
	serve := func(t *testing.T, guard *SecurityGuard) *httptest.ResponseRecorder {
		t.Helper()
		handler, err := NewHandler(Options{
			Dependencies:  testExtensionDependencies(t, nil),
			SecurityGuard: guard,
			AdditionalRoutes: []RouteRegistrar{func(mux *http.ServeMux, deps DependencySet) error {
				// redeemEvidenceHandle declares the bearer session and the session
				// cookie without CSRF; a handler verifying CSRF mismatches.
				return deps.PublicRoutes.Bind(mux, "module.evidence", "redeemEvidenceHandle", succeedAfter(cookiePrincipal(true)))
			}},
		})
		if err != nil {
			t.Fatalf("NewHandler(): %v", err)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/evidence-handles/handle", nil))
		return recorder
	}
	var reported []SecurityMismatch
	guarded := serve(t, &SecurityGuard{Enforce: true, Report: func(mismatch SecurityMismatch) { reported = append(reported, mismatch) }})
	want := []SecurityMismatch{{OperationID: "redeemEvidenceHandle", Status: http.StatusOK, Declared: "session", Observed: "session_csrf"}}
	if guarded.Code != http.StatusInternalServerError || !reflect.DeepEqual(reported, want) {
		t.Fatalf("guarded = %d reported=%v", guarded.Code, reported)
	}
	if unguarded := serve(t, nil); unguarded.Code != http.StatusOK {
		t.Fatalf("unguarded = %d", unguarded.Code)
	}
}
