package auth_test

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// authAdmissionBodies are request bodies that pass each auth operation's
// decoding, so a later outcome is decided by the addressed resource.
var authAdmissionBodies = map[string]any{
	"patchDeploymentUser":             map[string]any{"base_user_version": 1, "display_name": "Admission Patched"},
	"resetDeploymentUserPassword":     map[string]any{"base_user_version": 1, "client_txn_id": "txn-admission-reset", "new_password": "AdmissionReset1!"},
	"resetDeploymentUserTOTP":         map[string]any{"base_user_version": 1, "client_txn_id": "txn-admission-totp-reset"},
	"revokeAllDeploymentUserSessions": map[string]any{"client_txn_id": "txn-admission-revoke", "reason": "admission"},
	"createEnterpriseAuthBinding":     map[string]any{"base_user_version": 1, "client_txn_id": "txn-admission-bind", "provider_key": "admission", "provider_subject": "subject"},
	"rotateEnterpriseAuthBinding":     map[string]any{"base_user_version": 1, "client_txn_id": "txn-admission-rotate", "new_provider_subject": "rotated"},
	"retireEnterpriseAuthBinding":     map[string]any{"base_user_version": 1, "client_txn_id": "txn-admission-retire"},
}

// authOperationPath fills an auth operation's path template.
func authOperationPath(operation httpapi.ContractOperation, values map[string]string) string {
	path := operation.PathTemplate
	for _, parameter := range operation.PathParameters {
		path = strings.ReplaceAll(path, "{"+parameter.Name+"}", values[parameter.Name])
	}
	return path
}

func authOperation(t testing.TB, operationID string) httpapi.ContractOperation {
	t.Helper()
	for _, operation := range httpapi.ContractOperationsForOwner("module.auth") {
		if operation.OperationID == operationID {
			return operation
		}
	}
	t.Fatalf("no auth operation %q", operationID)
	return httpapi.ContractOperation{}
}

// requireAuthResponsesAlike requires two error responses to be identical in
// everything but their request identity (Core 01 REQ-01-707 step 3).
func requireAuthResponsesAlike(t testing.TB, got, want *http.Response, status int, code string) {
	t.Helper()
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{got, want} {
		if cookies := response.Header.Values("Set-Cookie"); len(cookies) != 0 {
			t.Fatalf("%s response set cookies %v", code, cookies)
		}
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("response %v differs from the missing-resource response %v", bodies[0], bodies[1])
	}
}

// TestAuthRouteAdmissionOrder_Integration covers Core 01 REQ-01-707, REQ-01-703
// and AC-617 for the 27 auth operations admitted centrally: authentication and
// cookie CSRF precede path parameters and query validation; a malformed or
// encoded-slash user_id, auth_binding_id or provider_key receives the declared
// not-found response, identical to a missing resource, before the
// deployment-administration gate; the credentialBootstrapBearer alternative is
// admitted only where declared and only as the sole auth mode (REQ-01-522,
// Table 3.3.2.2-D); route resolution rejections are enveloped; session
// inspection never slides idle expiry while other operations do (Core 04
// REQ-04-010); and without the enterprise_authentication claim the unclaimed
// family gate answers every enterprise operation (REQ-01-702).
func TestAuthRouteAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	t.Run("claimed enterprise profile", func(t *testing.T) {
		server, db := startEnterpriseServer(t, runtime, "auth-admission-order")
		defer db.Close()
		base := server.HTTP.URL
		start := httptestx.SetClockFixed(t, server, time.Date(2026, time.November, 19, 9, 0, 0, 0, time.UTC))
		seedLocalUserFlags(t, db, "admission-admin@example.test", "Admission Admin", "AdmissionAdmin1!", false, true, true)
		memberID := seedLocalUserFlags(t, db, "admission-member@example.test", "Admission Member", "AdmissionMember1!", false, false, true)
		seedLocalUser(t, db, "admission-setup@example.test", "Admission Setup", "AdmissionSetup1!", true)
		admin := flowtest.LoginLocalUser(t, http.DefaultClient, base, "admission-admin@example.test", "AdmissionAdmin1!", nil)
		member := flowtest.LoginLocalUser(t, http.DefaultClient, base, "admission-member@example.test", "AdmissionMember1!", nil)
		missingID := uuid.NewString()
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		// Each helper takes the calling (sub)test's t.
		send := func(t testing.TB, method, path string, body any, login *authcookietest.AuthCookies, options ...func(*http.Request)) *http.Response {
			t.Helper()
			if login != nil {
				options = append(options, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
			}
			if method == http.MethodGet || method == http.MethodHead {
				body = nil
			}
			return httptestx.DoJSON(t, client, method, base+path, body, options...)
		}
		secured := make([]httpapi.ContractOperation, 0, 21)
		for _, operation := range httpapi.ContractOperationsForOwner("module.auth") {
			if len(operation.Security) != 0 {
				secured = append(secured, operation)
			}
		}
		if len(secured) != 21 {
			t.Fatalf("secured auth operations = %d, want 21", len(secured))
		}
		malformedValues := map[string]string{"user_id": "not-a-uuid", "auth_binding_id": "not-a-uuid"}

		t.Run("authentication and csrf precede path parameters and query validation", func(t *testing.T) {
			for _, operation := range secured {
				path := authOperationPath(operation, malformedValues) + "?limit=1"
				body := authAdmissionBodies[operation.OperationID]
				httptestx.RequireErrorEnvelope(t, send(t, operation.Method, path, body, nil), http.StatusUnauthorized, "session_required")
				if operation.StateChanging {
					withoutCSRF := send(t, operation.Method, path, body, nil, httptestx.WithCookies(admin.Session))
					httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
				}
			}
			// The singleton query check follows authentication, and on
			// getDeploymentUser the deployment-administration gate as well.
			httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/auth/session?limit=1", nil, &member), http.StatusBadRequest, "invalid_pagination_request")
			httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/users/"+memberID+"?limit=1", nil, &member), http.StatusForbidden, "authorization_denied")
			httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/users/"+memberID+"?limit=1", nil, &admin), http.StatusBadRequest, "invalid_pagination_request")
		})

		t.Run("a malformed identifier is the declared not-found before the admin gate", func(t *testing.T) {
			for _, operationID := range []string{"getDeploymentUser", "patchDeploymentUser", "resetDeploymentUserPassword", "resetDeploymentUserTOTP", "revokeAllDeploymentUserSessions", "createEnterpriseAuthBinding", "rotateEnterpriseAuthBinding", "retireEnterpriseAuthBinding"} {
				operation := authOperation(t, operationID)
				code := operation.PathParameters[0].NotFoundCode
				body := authAdmissionBodies[operationID]
				missing := map[string]string{"user_id": missingID, "auth_binding_id": uuid.NewString()}
				for _, malformedUser := range []string{"not-a-uuid", strings.ReplaceAll(memberID, "-", ""), "%7B" + memberID + "%7D", memberID + "%2Fx", strings.ToUpper(memberID) + "0"} {
					malformed := authOperationPath(operation, map[string]string{"user_id": malformedUser, "auth_binding_id": missing["auth_binding_id"]})
					requireAuthResponsesAlike(t, send(t, operation.Method, malformed, body, &admin), send(t, operation.Method, authOperationPath(operation, missing), body, &admin), http.StatusNotFound, code)
					// Syntax is decided before the gate (decision 12): a non-admin
					// receives the same response, and only a well-formed
					// identifier reaches the gate.
					requireAuthResponsesAlike(t, send(t, operation.Method, malformed, body, &member), send(t, operation.Method, authOperationPath(operation, missing), body, &admin), http.StatusNotFound, code)
				}
				httptestx.RequireErrorEnvelope(t, send(t, operation.Method, authOperationPath(operation, missing), body, &member), http.StatusForbidden, "authorization_denied")
				if len(operation.PathParameters) == 2 {
					malformedBinding := authOperationPath(operation, map[string]string{"user_id": missingID, "auth_binding_id": "not-a-uuid"})
					requireAuthResponsesAlike(t, send(t, operation.Method, malformedBinding, body, &admin), send(t, operation.Method, authOperationPath(operation, missing), body, &admin), http.StatusNotFound, operation.PathParameters[1].NotFoundCode)
				}
			}
		})

		t.Run("an encoded slash stays inside user_id", func(t *testing.T) {
			before := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, memberID)
			missingUser := send(t, http.MethodGet, "/api/v1/users/"+missingID, nil, &admin)
			requireAuthResponsesAlike(t, send(t, http.MethodGet, "/api/v1/users/"+memberID+"%2Fpassword%2Freset", nil, &admin), missingUser, http.StatusNotFound, "user_not_found")
			revokeBody := authAdmissionBodies["revokeAllDeploymentUserSessions"]
			requireAuthResponsesAlike(t,
				send(t, http.MethodPost, "/api/v1/users/"+memberID+"%2Fsessions%2Frevoke-all/password/reset", authAdmissionBodies["resetDeploymentUserPassword"], &admin),
				send(t, http.MethodPost, "/api/v1/users/"+missingID+"/password/reset", authAdmissionBodies["resetDeploymentUserPassword"], &admin),
				http.StatusNotFound, "user_not_found")
			requireAuthResponsesAlike(t,
				send(t, http.MethodPatch, "/api/v1/users/"+memberID+"%2Fsessions%2Frevoke-all", revokeBody, &admin),
				send(t, http.MethodPatch, "/api/v1/users/"+missingID, revokeBody, &admin),
				http.StatusNotFound, "user_not_found")
			if after := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, memberID); after != before {
				t.Fatalf("encoded-slash requests revoked sessions: before=%d after=%d", before, after)
			}
		})

		t.Run("a malformed provider_key is the declared not-found", func(t *testing.T) {
			for _, operationID := range []string{"beginEnterpriseAuth", "completeEnterpriseOIDC", "completeEnterpriseSAML", "finishEnterpriseSAML"} {
				operation := authOperation(t, operationID)
				for _, malformedKey := range []string{"UPPER", "a%2Fb", strings.Repeat("a", 65), "-leading"} {
					response := send(t, operation.Method, authOperationPath(operation, map[string]string{"provider_key": malformedKey}), map[string]any{}, nil)
					if cookies := response.Header.Values("Set-Cookie"); len(cookies) != 0 {
						t.Fatalf("%s set cookies %v", operationID, cookies)
					}
					if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "auth_provider_not_found")); len(details) != 0 {
						t.Fatalf("%s not-found details = %v", operationID, details)
					}
				}
			}
			// Every enterprise provider route looks the provider up before its
			// body, RelayState, completion or browser-binding checks, so a
			// well-formed missing key receives the malformed key's response
			// (Core 01 REQ-01-707 step 3).
			for _, operationID := range []string{"beginEnterpriseAuth", "completeEnterpriseOIDC", "completeEnterpriseSAML", "finishEnterpriseSAML"} {
				operation := authOperation(t, operationID)
				for _, body := range []any{map[string]any{}, []any{}} {
					requireAuthResponsesAlike(t,
						send(t, operation.Method, authOperationPath(operation, map[string]string{"provider_key": "a%2Fb"}), body, nil),
						send(t, operation.Method, authOperationPath(operation, map[string]string{"provider_key": "missing-provider"}), body, nil),
						http.StatusNotFound, "auth_provider_not_found")
				}
			}
		})

		t.Run("an existing provider is classified before its protocol checks", func(t *testing.T) {
			seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "admission-saml", Type: "saml", Name: "Admission SAML", Enabled: true, Interactive: true, Issuer: "https://idp.example.test", Audience: "cartulary-sp"})
			seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "admission-saml-off", Type: "saml", Name: "Admission SAML Off", Enabled: false, Interactive: true, Issuer: "https://idp.example.test", Audience: "cartulary-sp"})
			requireReason := func(t testing.TB, response *http.Response, status int, code, reason string) {
				t.Helper()
				if cookies := response.Header.Values("Set-Cookie"); len(cookies) != 0 {
					t.Fatalf("%s response set cookies %v", code, cookies)
				}
				details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, status, code))
				if reason == "" && len(details) != 0 || reason != "" && !reflect.DeepEqual(details, map[string]any{"reason_code": reason}) {
					t.Fatalf("%s details = %v, want reason %q", code, details, reason)
				}
			}
			acs := authOperation(t, "completeEnterpriseSAML")
			finish := authOperation(t, "finishEnterpriseSAML")
			begin := authOperation(t, "beginEnterpriseAuth")
			// An enabled provider keeps its transaction checks.
			requireReason(t, send(t, acs.Method, authOperationPath(acs, map[string]string{"provider_key": "admission-saml"}), map[string]any{}, nil), http.StatusConflict, "enterprise_auth_transaction_rejected", "not_found")
			requireReason(t, send(t, finish.Method, authOperationPath(finish, map[string]string{"provider_key": "admission-saml"}), nil, nil), http.StatusConflict, "enterprise_auth_transaction_rejected", "completion_mismatch")
			// A disabled provider is rejected before them, as the OIDC callback
			// rejects it before its state and browser binding.
			for _, operation := range []httpapi.ContractOperation{acs, finish} {
				requireReason(t, send(t, operation.Method, authOperationPath(operation, map[string]string{"provider_key": "admission-saml-off"}), map[string]any{}, nil), http.StatusNotFound, "auth_provider_disabled", "")
			}
			requireReason(t, send(t, begin.Method, authOperationPath(begin, map[string]string{"provider_key": "admission-saml-off"}), []any{}, nil), http.StatusNotFound, "auth_provider_disabled", "")
		})

		t.Run("the credential bootstrap alternative", func(t *testing.T) {
			bootstrapToken := flowtest.RequireBootstrapLogin(t, http.DefaultClient, base, "admission-setup@example.test", "AdmissionSetup1!")
			bearer := httptestx.WithHeader("Authorization", "Bearer "+bootstrapToken)
			beginBody := map[string]any{"client_txn_id": "txn-admission-bootstrap-begin"}
			// A second auth mode is rejected before any setup state exists.
			mixed := send(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", beginBody, &admin, bearer)
			if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, mixed, http.StatusBadRequest, "invalid_auth_request")); details["field"] != "authorization" {
				t.Fatalf("mixed auth mode details = %v", details)
			}
			for _, operation := range []struct{ method, path string }{{http.MethodGet, "/api/v1/auth/credential-state"}, {http.MethodGet, "/api/v1/users"}, {http.MethodPost, "/api/v1/auth/logout"}} {
				rejected := send(t, operation.method, operation.path, map[string]any{}, nil, bearer)
				if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, rejected, http.StatusConflict, "credential_bootstrap_rejected")); details["reason_code"] != "not_allowed_for_route" {
					t.Fatalf("%s bootstrap rejection details = %v", operation.path, details)
				}
			}
			if count := queryCount(t, db, `SELECT count(*) FROM pending_totp_enrollments`); count != 0 {
				t.Fatalf("rejected bootstrap requests created %d enrollments", count)
			}
			// The bootstrap bearer needs no CSRF proof; a cookie session does.
			begin := flowtest.BeginTOTPEnrollment(t, client, base, bootstrapToken, beginBody)
			if begin["enrollment_id"] == nil {
				t.Fatalf("bootstrap begin = %v", begin)
			}
			httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, "/api/v1/auth/mfa/totp/begin", beginBody, nil, httptestx.WithCookies(member.Session)), http.StatusForbidden, "csrf_verification_failed")
			// An expired token is rejected with its reason before the route.
			expiredToken := flowtest.RequireBootstrapLogin(t, http.DefaultClient, base, "admission-setup@example.test", "AdmissionSetup1!")
			httptestx.SetClockFixed(t, server, start.Add(11*time.Minute))
			defer httptestx.SetClockFixed(t, server, start)
			for _, path := range []string{"/api/v1/auth/mfa/totp/begin", "/api/v1/auth/credential-state"} {
				method := http.MethodPost
				if strings.HasSuffix(path, "credential-state") {
					method = http.MethodGet
				}
				expired := send(t, method, path, beginBody, nil, httptestx.WithHeader("Authorization", "Bearer "+expiredToken))
				if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, expired, http.StatusConflict, "credential_bootstrap_rejected")); details["reason_code"] != "expired" {
					t.Fatalf("%s expired bootstrap details = %v", path, details)
				}
			}
		})

		t.Run("route resolution rejections are enveloped", func(t *testing.T) {
			requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow ...string) {
				t.Helper()
				if got := response.Header.Get("Allow"); got != strings.Join(allow, ", ") {
					t.Fatalf("Allow = %q, want %v", got, allow)
				}
				allowed := make([]any, 0, len(allow))
				for _, method := range allow {
					allowed = append(allowed, method)
				}
				if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")); !reflect.DeepEqual(details, map[string]any{"allowed_methods": allowed}) {
					t.Fatalf("405 details = %v", details)
				}
			}
			head := send(t, http.MethodHead, "/api/v1/account/profile", nil, &member)
			headBody, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != "GET, PATCH" || len(headBody) != 0 {
				t.Fatalf("HEAD = %d Allow=%q body=%q err=%v", head.StatusCode, head.Header.Get("Allow"), headBody, err)
			}
			requireMethodNotAllowed(t, send(t, http.MethodDelete, "/api/v1/auth/session", nil, &member), "GET")
			requireMethodNotAllowed(t, send(t, http.MethodPut, "/api/v1/users/"+memberID, map[string]any{}, &admin), "GET", "PATCH")
			for _, unknown := range []string{"/api/v1/users/" + memberID + "/unknown", "/api/v1/auth/unknown", "/api/v1/users//" + memberID} {
				response := send(t, http.MethodGet, unknown, nil, &admin)
				if location := response.Header.Get("Location"); location != "" {
					t.Fatalf("%s redirected to %s", unknown, location)
				}
				if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "route_not_found")); len(details) != 0 {
					t.Fatalf("route_not_found details = %v", details)
				}
			}
		})

		t.Run("only session inspection is not qualifying activity", func(t *testing.T) {
			login := querySessionRow(t, db, memberID)
			httptestx.SetClockFixed(t, server, start.Add(5*time.Minute))
			defer httptestx.SetClockFixed(t, server, start)
			httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, "/api/v1/auth/session", nil, &member), http.StatusOK)
			if after := querySessionRow(t, db, memberID); !after.lastQualifyingActivityAt.Equal(login.lastQualifyingActivityAt) || !after.idleExpiresAt.Equal(login.idleExpiresAt) {
				t.Fatalf("session inspection slid idle expiry: before=%v after=%v", login.idleExpiresAt, after.idleExpiresAt)
			}
			httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, "/api/v1/auth/credential-state", nil, &member), http.StatusOK)
			if after := querySessionRow(t, db, memberID); !after.lastQualifyingActivityAt.Equal(start.Add(5 * time.Minute)) {
				t.Fatalf("credential state did not slide idle expiry: last activity %v", after.lastQualifyingActivityAt)
			}
		})
	})

	t.Run("unclaimed enterprise profile", func(t *testing.T) {
		server, db := startServer(t, runtime, "auth-admission-unclaimed")
		defer db.Close()
		base := server.HTTP.URL
		adminID := seedLocalUserFlags(t, db, "unclaimed-admin@example.test", "Unclaimed Admin", "UnclaimedAdmin1!", false, true, true)
		admin := flowtest.LoginLocalUser(t, http.DefaultClient, base, "unclaimed-admin@example.test", "UnclaimedAdmin1!", nil)
		var families []string
		for _, profile := range contracttest.CurrentProfileExtensions(t) {
			if profile.ProfileID == "enterprise_authentication" {
				families = profile.RouteFamilies
			}
		}
		values := map[string]string{"user_id": adminID, "auth_binding_id": uuid.NewString(), "provider_key": "provider"}
		gated := 0
		for _, operation := range httpapi.ContractOperationsForOwner("module.auth") {
			if operation.Availability != "enterprise_authentication" {
				continue
			}
			gated++
			path := authOperationPath(operation, values)
			wantFamily := ""
			for _, family := range families {
				if strings.HasPrefix(path, strings.ReplaceAll(family, "{user_id}", adminID)) {
					wantFamily = family
				}
			}
			// The reserved-family gate precedes route resolution and
			// authentication, so it answers with and without a session.
			for _, options := range [][]func(*http.Request){nil, {httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value)}} {
				var body any
				if operation.Method != http.MethodGet {
					body = map[string]any{}
				}
				response := httptestx.DoJSON(t, http.DefaultClient, operation.Method, base+path, body, options...)
				details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "extension_profile_not_claimed"))
				if !reflect.DeepEqual(details, map[string]any{"profile_id": "enterprise_authentication", "route_family": wantFamily}) {
					t.Fatalf("%s gate details = %v, want family %q", operation.OperationID, details, wantFamily)
				}
			}
		}
		if gated != 8 {
			t.Fatalf("enterprise operations = %d, want 8", gated)
		}
	})
}
