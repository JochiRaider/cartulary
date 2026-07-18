package auth_test

import (
	"context"
	"database/sql"
	"encoding/base32"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/routetest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/securityassert"
)

func TestUserListContinuationUsesLiveRows(t *testing.T) {
	runtime := flowtest.StartRuntime(t)

	server, db := startServer(t, runtime, "authentication-support-pagination")
	defer db.Close()

	seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000701", "pagination-admin@example.test", "Pagination Admin", "PaginationAdmin1!", true)
	_ = seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000702", "pagination-one@example.test", "Pagination One", "PaginationOne1!", false)
	userTwoID := seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000703", "pagination-two@example.test", "Pagination Two", "PaginationTwo1!", false)

	adminSession, adminCSRF := loginLocalUser(t, server, "pagination-admin@example.test", "PaginationAdmin1!", nil)

	initialList := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users",
		nil,
		withCookies(adminSession),
	)
	initialBody := httptestx.RequireSuccessEnvelope(t, initialList, http.StatusOK)
	initialUsers := initialBody["data"].(map[string]any)["users"].([]any)
	userTwoIndex := indexOfUser(initialUsers, userTwoID)
	if userTwoIndex < 1 {
		t.Fatalf("expected user two to appear after at least one user, got index %d in %#v", userTwoIndex, initialUsers)
	}

	firstPage := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?limit="+url.QueryEscape(strconv.Itoa(userTwoIndex)),
		nil,
		withCookies(adminSession),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageUsers := firstPageBody["data"].(map[string]any)["users"].([]any)
	if len(firstPageUsers) != userTwoIndex {
		t.Fatalf("expected %d users on first page, got %#v", userTwoIndex, firstPageUsers)
	}
	nextCursor := requirePagingCursor(t, firstPageBody)

	patchResp := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/users/"+userTwoID,
		map[string]any{
			"base_user_version": 1,
			"display_name":      "Pagination Two Patched",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		withCookies(adminSession),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedUsers := continuedBody["data"].(map[string]any)["users"].([]any)
	liveContinuedUser := findUserResource(t, continuedUsers, userTwoID)
	if liveContinuedUser["display_name"] != "Pagination Two Patched" {
		t.Fatalf("expected live user payload, got %#v", liveContinuedUser)
	}

	fresh := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users",
		nil,
		withCookies(adminSession),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	liveUser := findUserResource(t, freshBody["data"].(map[string]any)["users"].([]any), userTwoID)
	if liveUser["display_name"] != "Pagination Two Patched" {
		t.Fatalf("expected fresh user list to reflect live mutation, got %#v", liveUser)
	}
}

func TestUserListSearchTokenizesIdentifiersAndBindsCursorScope(t *testing.T) {
	runtime := flowtest.StartRuntime(t)

	server, db := startServer(t, runtime, "authentication-support-user-list-search")
	defer db.Close()

	nonAdminID := seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000711", "search-viewer@example.test", "Search Viewer", "SearchViewer1!", false)
	adminID := seedFixedLocalUser(t, db, "d56d8685-f36e-448c-8f44-bd2978aa26d8", "uuid-admin@example.test", "UUID Admin", "UUIDAdmin12!", true)

	adminSession, _ := loginLocalUser(t, server, "uuid-admin@example.test", "UUIDAdmin12!", nil)

	searchResp := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?limit=1&search="+url.QueryEscape(adminID)+"&is_active=true&is_deployment_admin=true",
		nil,
		withCookies(adminSession),
	)
	searchBody := httptestx.RequireSuccessEnvelope(t, searchResp, http.StatusOK)
	searchRows := searchBody["data"].(map[string]any)["users"].([]any)
	if len(searchRows) != 1 || searchRows[0].(map[string]any)["user_id"] != adminID {
		t.Fatalf("full UUID search must find the deployment admin before pagination, got %#v", searchRows)
	}

	emailResp := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?search="+url.QueryEscape("uuid-admin@example.test"),
		nil,
		withCookies(adminSession),
	)
	emailBody := httptestx.RequireSuccessEnvelope(t, emailResp, http.StatusOK)
	findUserResource(t, emailBody["data"].(map[string]any)["users"].([]any), adminID)

	firstPage := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?limit=1",
		nil,
		withCookies(adminSession),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := firstPageBody["data"].(map[string]any)["users"].([]any)
	if len(firstPageRows) != 1 || firstPageRows[0].(map[string]any)["user_id"] != nonAdminID {
		t.Fatalf("expected unsearched first page to prove later-page search fixture, got %#v", firstPageRows)
	}
	nextCursor := requirePagingCursor(t, firstPageBody)
	mismatchedCursor := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?cursor_token="+url.QueryEscape(nextCursor)+"&search="+url.QueryEscape(adminID),
		nil,
		withCookies(adminSession),
	)
	mismatchBody := httptestx.RequireErrorEnvelope(t, mismatchedCursor, http.StatusBadRequest, "invalid_pagination_request")
	details := mismatchBody["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected cursor_query_mismatch for changed user-list search scope, got %#v", details)
	}

	nonAdminSession, _ := loginLocalUser(t, server, "search-viewer@example.test", "SearchViewer1!", nil)
	denied := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/users?search="+url.QueryEscape(adminID),
		nil,
		withCookies(nonAdminSession),
	)
	httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
}

func TestSurfaceEnvelope(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-envelope")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessSurfaceEnvelope) {
		t.Run(string(route.ID), func(t *testing.T) {
			req := ctx.buildSuccessRequest(t, route)
			resp := ctx.do(t, req)
			httptestx.RequireSuccessEnvelope(t, resp, route.SuccessStatus)
		})
	}
}

func TestBootstrapBoundaries(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-bootstrap")
	defer ctx.db.Close()

	bootstrapUserID, bootstrapEmail, bootstrapPassword := ctx.newLocalUser(t, "bootstrap-boundary", true, false, true)
	bootstrapToken := requireBootstrapLogin(t, ctx.server, bootstrapEmail, bootstrapPassword)
	bootstrapIncidentID := incidentstoretest.SeedIncidentMembershipSQL(t, ctx.db, bootstrapUserID, "authentication-support-bootstrap-ws")
	targetUserID, _, _ := ctx.newLocalUser(t, "bootstrap-target", false, false, true)
	totpTargetID, _, _, _, _ := ctx.newActiveTOTPLoggedInUser(t, "bootstrap-totp-target", false)
	revokeTargetID, _, _, _, _ := ctx.newLoggedInLocalUser(t, "bootstrap-revoke-target", false, false, true)

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessBootstrapBoundary) {
		t.Run(string(route.ID), func(t *testing.T) {
			if route.Transport == routetest.RouteTransportWebSocket {
				flowtest.RequireBootstrapWebsocketRejected(t, ctx.server.HTTP.URL, bootstrapIncidentID, bootstrapToken)
				return
			}

			req := ctx.buildBootstrapBoundaryRequest(t, route, bootstrapToken, targetUserID, totpTargetID, revokeTargetID)
			resp := ctx.do(t, req)
			body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
			details := body["error"].(map[string]any)["details"].(map[string]any)
			if got := details["reason_code"]; got != "not_allowed_for_route" {
				t.Fatalf("unexpected bootstrap rejection reason_code for %s: got %v want not_allowed_for_route", route.ID, got)
			}
		})
	}
}

func TestCSRFProtection(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-csrf")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessCSRF) {
		t.Run(string(route.ID), func(t *testing.T) {
			req := ctx.buildCSRFFailureRequest(t, route)
			resp := ctx.do(t, req)
			httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "csrf_verification_failed")
		})
	}
}

func TestReplayAndStoredPayloadSafety(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-replay")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessReplayStoredPayload) {
		t.Run(string(route.ID), func(t *testing.T) {
			req := ctx.buildSuccessRequest(t, route)
			firstResp := ctx.do(t, req)
			firstData := httptestx.RequireSuccessEnvelope(t, firstResp, route.SuccessStatus)["data"].(map[string]any)

			switch route.ID {
			case routetest.RoutePasswordChange:
				idempotency := lookupRouteIdempotency(t, ctx.db, req.routeKey, req.actorUserID, req.scopeKey, req.clientTxnID)
				if idempotency.StatusCode != route.SuccessStatus {
					t.Fatalf("unexpected idempotency status for %s: got %d want %d", route.ID, idempotency.StatusCode, route.SuccessStatus)
				}
				securityassert.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

				replayLogin := loginLocalUserWithSecondFactor(t, ctx.server, req.replayEmail, req.replayPassword, generateTOTPCode(t, req.replaySecretBase32))
				req.cookies = []*http.Cookie{replayLogin.sessionCookie, replayLogin.csrfCookie}
				req.headers = map[string]string{authn.CSRFHeaderName: replayLogin.csrfCookie.Value}
				replayedResp := ctx.do(t, req)
				replayedData := httptestx.RequireSuccessEnvelope(t, replayedResp, http.StatusOK)["data"].(map[string]any)
				requireJSONEquivalent(t, replayedData, firstData)

				if got := queryCount(t, ctx.db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND actor_user_id::text = $2 AND scope_key = $3 AND client_txn_id = $4`, req.routeKey, req.actorUserID, req.scopeKey, req.clientTxnID); got != 1 {
					t.Fatalf("expected one route_idempotency row for %s, got %d", route.ID, got)
				}
			case routetest.RouteTOTPBegin:
				replayedResp := ctx.do(t, req)
				replayedData := httptestx.RequireSuccessEnvelope(t, replayedResp, route.SuccessStatus)["data"].(map[string]any)
				if firstData["enrollment_id"] != replayedData["enrollment_id"] {
					t.Fatalf("expected begin replay to reuse enrollment_id: first=%v replay=%v", firstData["enrollment_id"], replayedData["enrollment_id"])
				}
				firstSetup := firstData["totp_setup"].(map[string]any)
				replayedSetup := replayedData["totp_setup"].(map[string]any)
				if firstSetup["secret_base32"] != replayedSetup["secret_base32"] {
					t.Fatalf("expected begin replay to reuse setup secret: first=%v replay=%v", firstSetup["secret_base32"], replayedSetup["secret_base32"])
				}
				ciphertext, nonce := queryPendingTOTPEnrollmentSecretMaterial(t, ctx.db, req.actorUserID, req.clientTxnID)
				requireEncryptedSecretMaterial(t, ciphertext, nonce, firstSetup["secret_base32"].(string))
				if got := queryCount(t, ctx.db, `SELECT COUNT(*) FROM pending_totp_enrollments WHERE user_id::text = $1 AND client_txn_id = $2`, req.actorUserID, req.clientTxnID); got != 1 {
					t.Fatalf("expected one pending enrollment row for %s, got %d", route.ID, got)
				}
			case routetest.RouteTOTPComplete:
				securityassert.RequireSecretSafePayload(t, firstData, forbiddenSecretKeys())
				if _, ok := firstData["totp_setup"]; ok {
					t.Fatalf("totp complete must not return setup material, got %#v", firstData)
				}
				ciphertext, nonce := queryUserTOTPSecretMaterial(t, ctx.db, req.actorUserID)
				requireEncryptedSecretMaterial(t, ciphertext, nonce, req.secretBase32)
				replayedResp := ctx.do(t, req)
				replayedBody := httptestx.RequireErrorEnvelope(t, replayedResp, http.StatusConflict, "credential_bootstrap_rejected")
				details := replayedBody["error"].(map[string]any)["details"].(map[string]any)
				if got := details["reason_code"]; got != "consumed" {
					t.Fatalf("unexpected complete replay rejection for %s: got %v want consumed", route.ID, got)
				}
			default:
				idempotency := lookupRouteIdempotency(t, ctx.db, req.routeKey, req.actorUserID, req.scopeKey, req.clientTxnID)
				if idempotency.StatusCode != route.SuccessStatus {
					t.Fatalf("unexpected idempotency status for %s: got %d want %d", route.ID, idempotency.StatusCode, route.SuccessStatus)
				}
				securityassert.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

				replayedResp := ctx.do(t, req)
				replayedData := httptestx.RequireSuccessEnvelope(t, replayedResp, http.StatusOK)["data"].(map[string]any)
				requireJSONEquivalent(t, replayedData, firstData)

				if got := queryCount(t, ctx.db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND actor_user_id::text = $2 AND scope_key = $3 AND client_txn_id = $4`, req.routeKey, req.actorUserID, req.scopeKey, req.clientTxnID); got != 1 {
					t.Fatalf("expected one route_idempotency row for %s, got %d", route.ID, got)
				}
			}
		})
	}
}

func TestAuditAttribution(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-audit")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessMutationAudit) {
		t.Run(string(route.ID), func(t *testing.T) {
			req := ctx.buildSuccessRequest(t, route)
			resp := ctx.do(t, req)
			body := httptestx.RequireSuccessEnvelope(t, resp, route.SuccessStatus)
			if req.targetUserID == "" {
				req.targetUserID = body["data"].(map[string]any)["user_id"].(string)
			}

			events := lookupUserAuditEvents(t, ctx.db, req.targetUserID)
			event := requireAuditEventBySource(t, events, req.auditSource)
			auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
				ActorUserID: event.ActorUserID,
				Source:      event.EventSource,
				ClientTxnID: event.ClientTxnID,
				RequestID:   event.RequestID,
				CreatedAt:   event.CreatedAt,
			}, req.actorUserID, req.auditSource, req.clientTxnID)
			securityassert.RequireSecretSafePayload(t, event.Before, forbiddenSecretKeys())
			securityassert.RequireSecretSafePayload(t, event.After, forbiddenSecretKeys())
		})
	}
}

func TestSessionRevocation(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-revocation")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessSessionRevocation) {
		t.Run(string(route.ID), func(t *testing.T) {
			switch route.ID {
			case routetest.RouteSessionLifecycleWS:
				socketUserID, _, _, socketSession, _ := ctx.newLoggedInLocalUser(t, "support-socket-bootstrap", false, false, true)
				if socketUserID == "" {
					t.Fatal("expected socket test user")
				}
				incidentID := incidentstoretest.SeedIncidentMembershipSQL(t, ctx.db, socketUserID, "authentication-support-socket-bootstrap")
				socket := connectSessionSocket(t, ctx.server, incidentID, socketSession.Value)
				socket.Close(websocket.StatusNormalClosure, "support_cleanup")
			case routetest.RouteLogout:
				_, _, _, sessionCookie, csrfCookie, socket := ctx.newSocketSession(t, "support-logout", false)
				defer socket.Close(websocket.StatusNormalClosure, "support_cleanup")

				resp := doJSON(
					t,
					http.MethodPost,
					ctx.server.HTTP.URL+"/api/v1/auth/logout",
					nil,
					withCookies(sessionCookie, csrfCookie),
					withHeader(authn.CSRFHeaderName, csrfCookie.Value),
				)
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
				expectSessionRevoked(t, socket, "session_revoked")
			case routetest.RoutePasswordChange:
				userID, _, password, secretBase32, login := ctx.newActiveTOTPLoggedInUser(t, "support-password-change", false)
				incidentID := incidentstoretest.SeedIncidentMembershipSQL(t, ctx.db, userID, "authentication-support-password-change")
				socket := connectSessionSocket(t, ctx.server, incidentID, login.sessionCookie.Value)
				defer socket.Close(websocket.StatusNormalClosure, "support_cleanup")

				resp := doJSON(
					t,
					http.MethodPost,
					ctx.server.HTTP.URL+"/api/v1/auth/password/change",
					map[string]any{
						"client_txn_id":    ctx.nextClientTxn("support-password-change"),
						"current_password": password,
						"new_password":     "SupportPasswordChanged1!",
						"second_factor": map[string]any{
							"kind": "totp",
							"assertion": map[string]any{
								"code": generateTOTPCode(t, secretBase32),
							},
						},
					},
					withCookies(login.sessionCookie, login.csrfCookie),
					withHeader(authn.CSRFHeaderName, login.csrfCookie.Value),
				)
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
				expectSessionRevoked(t, socket, "session_revoked")
				if got := queryCount(t, ctx.db, `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL`, userID); got != 0 {
					t.Fatalf("expected password change to revoke all sessions, got %d active rows", got)
				}
			case routetest.RouteUsersPasswordReset:
				targetID, _, _, targetSession, _, socket := ctx.newSocketSession(t, "support-password-reset-target", false)
				defer socket.Close(websocket.StatusNormalClosure, "support_cleanup")

				resp := doJSON(
					t,
					http.MethodPost,
					ctx.server.HTTP.URL+"/api/v1/users/"+targetID+"/password/reset",
					map[string]any{
						"base_user_version": 1,
						"client_txn_id":     ctx.nextClientTxn("support-password-reset"),
						"new_password":      "SupportResetPassword1!",
						"reason":            "support reset",
					},
					withCookies(ctx.adminSession, ctx.adminCSRF),
					withHeader(authn.CSRFHeaderName, ctx.adminCSRF.Value),
				)
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
				expectSessionRevoked(t, socket, "session_revoked")
				if replay := doJSON(t, http.MethodGet, ctx.server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(targetSession)); replay.StatusCode != http.StatusUnauthorized {
					t.Fatalf("expected reset target session to be unauthorized after revoke, got %d", replay.StatusCode)
				}
			case routetest.RouteUsersTOTPReset:
				targetID, _, _, _, targetLogin := ctx.newActiveTOTPLoggedInUser(t, "support-totp-reset-target", false)
				incidentID := incidentstoretest.SeedIncidentMembershipSQL(t, ctx.db, targetID, "authentication-support-totp-reset")
				socket := connectSessionSocket(t, ctx.server, incidentID, targetLogin.sessionCookie.Value)
				defer socket.Close(websocket.StatusNormalClosure, "support_cleanup")

				resp := doJSON(
					t,
					http.MethodPost,
					ctx.server.HTTP.URL+"/api/v1/users/"+targetID+"/mfa/totp/reset",
					map[string]any{
						"base_user_version": 1,
						"client_txn_id":     ctx.nextClientTxn("support-totp-reset"),
						"reason":            "support totp reset",
					},
					withCookies(ctx.adminSession, ctx.adminCSRF),
					withHeader(authn.CSRFHeaderName, ctx.adminCSRF.Value),
				)
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
				expectSessionRevoked(t, socket, "session_revoked")
			case routetest.RouteUsersRevokeAll:
				targetID, _, _, _, _, socket := ctx.newSocketSession(t, "support-revoke-all-target", false)
				defer socket.Close(websocket.StatusNormalClosure, "support_cleanup")

				resp := doJSON(
					t,
					http.MethodPost,
					ctx.server.HTTP.URL+"/api/v1/users/"+targetID+"/sessions/revoke-all",
					map[string]any{
						"client_txn_id": ctx.nextClientTxn("support-revoke-all"),
						"reason":        "support revoke all",
					},
					withCookies(ctx.adminSession, ctx.adminCSRF),
					withHeader(authn.CSRFHeaderName, ctx.adminCSRF.Value),
				)
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
				expectSessionRevoked(t, socket, "session_revoked")
			default:
				t.Fatalf("unexpected session-revocation route %s", route.ID)
			}
		})
	}
}

func TestAuthorizationReDerivation(t *testing.T) {
	runtime := flowtest.StartRuntime(t)

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessAuthorization) {
		t.Run(string(route.ID), func(t *testing.T) {
			ctx := newSupportRouteContext(t, runtime, "authentication-support-authorization-"+string(route.ID))
			defer ctx.db.Close()

			beforeReq := ctx.buildSuccessRequest(t, route)
			beforeResp := ctx.do(t, beforeReq)
			httptestx.RequireSuccessEnvelope(t, beforeResp, route.SuccessStatus)

			ctx.demotePrimaryAdmin(t)

			afterReq := ctx.buildSuccessRequest(t, route)
			afterResp := ctx.do(t, afterReq)
			body := httptestx.RequireErrorEnvelope(t, afterResp, route.AuthorizationStatus, route.AuthorizationCode)
			errorValue := body["error"].(map[string]any)
			contractassert.RequireAuthorizationReDerived(
				t,
				contractassert.AuthorizationOutcome{Status: route.SuccessStatus, Code: "success"},
				contractassert.AuthorizationOutcome{Status: afterResp.StatusCode, Code: errorValue["code"].(string)},
			)
		})
	}
}

func TestRequestContracts(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	ctx := newSupportRouteContext(t, runtime, "authentication-support-request-contracts")
	defer ctx.db.Close()

	for _, route := range routetest.RoutesForHarness(t, routetest.PublicRouteInventory(), routetest.RouteHarnessRequestContracts) {
		for _, contract := range route.RequestContracts.ClosedVocabulary {
			t.Run(string(route.ID)+"/closed-vocabulary/"+contract.Field, func(t *testing.T) {
				resp := ctx.do(t, ctx.buildClosedVocabularyRequest(t, route, contract.Field))
				body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, ctx.closedVocabularyCode(route, contract.Field))
				errorValue := body["error"].(map[string]any)
				contractassert.RequireClosedVocabularyRejected(
					t,
					errorValue["code"].(string),
					errorValue["details"].(map[string]any),
					contract.Field,
					contract.ReasonCode,
				)
			})
		}

		for _, contract := range route.RequestContracts.WritableStrings {
			t.Run(string(route.ID)+"/writable-string/"+contract.Field, func(t *testing.T) {
				ctx.requireWritableStringNormalization(t, route, contract.Field)
			})
		}
	}
}

type SupportRouteContext struct {
	server           *httptestx.Server
	db               *sql.DB
	adminID          string
	adminSession     *http.Cookie
	adminCSRF        *http.Cookie
	peerAdminID      string
	peerAdminSession *http.Cookie
	peerAdminCSRF    *http.Cookie
	sequence         int
}

type SupportRequest struct {
	route              routetest.RouteInventoryEntry
	path               string
	body               any
	rawBody            *string
	cookies            []*http.Cookie
	headers            map[string]string
	clientTxnID        string
	routeKey           string
	scopeKey           string
	actorUserID        string
	targetUserID       string
	auditSource        string
	secretBase32       string
	replayEmail        string
	replayPassword     string
	replaySecretBase32 string
}

func newSupportRouteContext(
	t testing.TB,
	runtime *flowtest.RuntimeHarness,
	prefix string,
) *SupportRouteContext {
	t.Helper()

	server, db := startServer(t, runtime, prefix)
	adminID := seedLocalUserFlags(t, db, prefix+"-admin@example.test", "Support Admin", "SupportAdminPass1!", false, true, true)
	adminSession, adminCSRF := loginLocalUser(t, server, prefix+"-admin@example.test", "SupportAdminPass1!", nil)
	peerAdminID := seedLocalUserFlags(t, db, prefix+"-peer-admin@example.test", "Support Peer Admin", "SupportPeerAdminPass1!", false, true, true)
	peerAdminSession, peerAdminCSRF := loginLocalUser(t, server, prefix+"-peer-admin@example.test", "SupportPeerAdminPass1!", nil)

	return &SupportRouteContext{
		server:           server,
		db:               db,
		adminID:          adminID,
		adminSession:     adminSession,
		adminCSRF:        adminCSRF,
		peerAdminID:      peerAdminID,
		peerAdminSession: peerAdminSession,
		peerAdminCSRF:    peerAdminCSRF,
	}
}

func (c *SupportRouteContext) buildSuccessRequest(t testing.TB, route routetest.RouteInventoryEntry) SupportRequest {
	t.Helper()

	switch route.ID {
	case routetest.RouteLogin:
		_, email, password := c.newLocalUser(t, "login", false, false, true)
		return SupportRequest{
			route:   route,
			path:    route.Template,
			body:    map[string]any{"username": email, "password": password},
			headers: map[string]string{},
		}
	case routetest.RouteSession:
		_, _, _, sessionCookie, _ := c.newLoggedInLocalUser(t, "session", false, false, true)
		return SupportRequest{
			route:   route,
			path:    route.Template,
			cookies: []*http.Cookie{sessionCookie},
			headers: map[string]string{},
		}
	case routetest.RouteLogout:
		_, _, _, sessionCookie, csrfCookie := c.newLoggedInLocalUser(t, "logout", false, false, true)
		return SupportRequest{
			route:   route,
			path:    route.Template,
			cookies: []*http.Cookie{sessionCookie, csrfCookie},
			headers: map[string]string{authn.CSRFHeaderName: csrfCookie.Value},
		}
	case routetest.RouteCredentialState:
		_, _, _, sessionCookie, _ := c.newLoggedInLocalUser(t, "credential-state", false, false, true)
		return SupportRequest{
			route:   route,
			path:    route.Template,
			cookies: []*http.Cookie{sessionCookie},
			headers: map[string]string{},
		}
	case routetest.RoutePasswordChange:
		userID, email, password, secretBase32, login := c.newActiveTOTPLoggedInUser(t, "password-change", false)
		clientTxnID := c.nextClientTxn("password-change")
		newPassword := "SupportPasswordChange1!"
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    clientTxnID,
				"current_password": password,
				"new_password":     newPassword,
				"second_factor": map[string]any{
					"kind": "totp",
					"assertion": map[string]any{
						"code": generateTOTPCode(t, secretBase32),
					},
				},
			},
			cookies:            []*http.Cookie{login.sessionCookie, login.csrfCookie},
			headers:            map[string]string{authn.CSRFHeaderName: login.csrfCookie.Value},
			clientTxnID:        clientTxnID,
			routeKey:           "auth.password.change",
			scopeKey:           "actor",
			actorUserID:        userID,
			targetUserID:       userID,
			replayEmail:        email,
			replayPassword:     newPassword,
			replaySecretBase32: secretBase32,
		}
	case routetest.RouteTOTPBegin:
		userID, email, password := c.newLocalUser(t, "totp-begin-bootstrap", true, false, true)
		bootstrapToken := requireBootstrapLogin(t, c.server, email, password)
		clientTxnID := c.nextClientTxn("totp-begin")
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id": clientTxnID,
			},
			headers:      map[string]string{"Authorization": "Bearer " + bootstrapToken},
			clientTxnID:  clientTxnID,
			actorUserID:  userID,
			targetUserID: userID,
		}
	case routetest.RouteTOTPComplete:
		userID, email, password := c.newLocalUser(t, "totp-complete-bootstrap", true, false, true)
		bootstrapToken := requireBootstrapLogin(t, c.server, email, password)
		beginTxnID := c.nextClientTxn("totp-complete-begin")
		beginResp := doJSON(
			t,
			http.MethodPost,
			c.server.HTTP.URL+"/api/v1/auth/mfa/totp/begin",
			map[string]any{"client_txn_id": beginTxnID},
			withHeader("Authorization", "Bearer "+bootstrapToken),
		)
		beginData := httptestx.RequireSuccessEnvelope(t, beginResp, http.StatusOK)["data"].(map[string]any)
		enrollmentID := beginData["enrollment_id"].(string)
		secretBase32 := beginData["totp_setup"].(map[string]any)["secret_base32"].(string)
		completeTxnID := c.nextClientTxn("totp-complete")
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id": completeTxnID,
				"enrollment_id": enrollmentID,
				"code":          generateTOTPCode(t, secretBase32),
			},
			headers:      map[string]string{"Authorization": "Bearer " + bootstrapToken},
			clientTxnID:  completeTxnID,
			actorUserID:  userID,
			targetUserID: userID,
			secretBase32: secretBase32,
		}
	case routetest.RouteUsersList:
		return SupportRequest{
			route:   route,
			path:    route.Template,
			cookies: []*http.Cookie{c.adminSession},
			headers: map[string]string{},
		}
	case routetest.RouteUsersCreate:
		email, password := c.nextIdentity("users-create-target")
		clientTxnID := c.nextClientTxn("users-create")
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    clientTxnID,
				"auth_kind":        "local",
				"email":            email,
				"display_name":     "Support Created User",
				"initial_password": password,
			},
			cookies:     []*http.Cookie{c.adminSession, c.adminCSRF},
			headers:     map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
			clientTxnID: clientTxnID,
			routeKey:    "users.create",
			scopeKey:    "actor",
			actorUserID: c.adminID,
			auditSource: "users.create",
		}
	case routetest.RouteUsersGet:
		targetUserID, _, _ := c.newLocalUser(t, "users-get-target", false, false, true)
		return SupportRequest{
			route:   route,
			path:    routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			cookies: []*http.Cookie{c.adminSession},
			headers: map[string]string{},
		}
	case routetest.RouteUsersPatch:
		targetUserID, _, _ := c.newLocalUser(t, "users-patch-target", false, false, true)
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"base_user_version": 1,
				"display_name":      "Support Patched User",
			},
			cookies:      []*http.Cookie{c.adminSession, c.adminCSRF},
			headers:      map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
			actorUserID:  c.adminID,
			targetUserID: targetUserID,
			auditSource:  "users.patch",
		}
	case routetest.RouteUsersPasswordReset:
		targetUserID, _, _ := c.newLocalUser(t, "users-password-reset-target", false, false, true)
		clientTxnID := c.nextClientTxn("users-password-reset")
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"base_user_version": 1,
				"client_txn_id":     clientTxnID,
				"new_password":      "SupportResetPassword1!",
				"reason":            "support password reset",
			},
			cookies:      []*http.Cookie{c.adminSession, c.adminCSRF},
			headers:      map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
			clientTxnID:  clientTxnID,
			routeKey:     "users.password.reset",
			scopeKey:     targetUserID,
			actorUserID:  c.adminID,
			targetUserID: targetUserID,
			auditSource:  "users.password.reset",
		}
	case routetest.RouteUsersTOTPReset:
		targetUserID, _, _, _, _ := c.newActiveTOTPLoggedInUser(t, "users-totp-reset-target", false)
		clientTxnID := c.nextClientTxn("users-totp-reset")
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"base_user_version": 1,
				"client_txn_id":     clientTxnID,
				"reason":            "support totp reset",
			},
			cookies:      []*http.Cookie{c.adminSession, c.adminCSRF},
			headers:      map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
			clientTxnID:  clientTxnID,
			routeKey:     "users.totp.reset",
			scopeKey:     targetUserID,
			actorUserID:  c.adminID,
			targetUserID: targetUserID,
			auditSource:  "users.totp.reset",
		}
	case routetest.RouteUsersRevokeAll:
		targetUserID, _, _, _, _ := c.newLoggedInLocalUser(t, "users-revoke-all-target", false, false, true)
		clientTxnID := c.nextClientTxn("users-revoke-all")
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"client_txn_id": clientTxnID,
				"reason":        "support revoke all",
			},
			cookies:      []*http.Cookie{c.adminSession, c.adminCSRF},
			headers:      map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
			clientTxnID:  clientTxnID,
			routeKey:     "users.sessions.revoke_all",
			scopeKey:     targetUserID,
			actorUserID:  c.adminID,
			targetUserID: targetUserID,
			auditSource:  "users.sessions.revoke_all",
		}
	default:
		t.Fatalf("unsupported success route %s", route.ID)
		return SupportRequest{}
	}
}

func (c *SupportRouteContext) buildCSRFFailureRequest(t testing.TB, route routetest.RouteInventoryEntry) SupportRequest {
	t.Helper()

	switch route.ID {
	case routetest.RouteTOTPBegin:
		userID, email, password, secretBase32, login := c.newActiveTOTPLoggedInUser(t, "totp-begin-csrf", false)
		_ = email
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("totp-begin-csrf"),
				"current_password": password,
				"second_factor": map[string]any{
					"kind": "totp",
					"assertion": map[string]any{
						"code": generateTOTPCode(t, secretBase32),
					},
				},
			},
			rawBody:      malformedJSONBody(),
			cookies:      []*http.Cookie{login.sessionCookie, login.csrfCookie},
			headers:      map[string]string{},
			actorUserID:  userID,
			targetUserID: userID,
		}
	case routetest.RouteTOTPComplete:
		userID, _, password, secretBase32, login := c.newActiveTOTPLoggedInUser(t, "totp-complete-csrf", false)
		beginResp := doJSON(
			t,
			http.MethodPost,
			c.server.HTTP.URL+"/api/v1/auth/mfa/totp/begin",
			map[string]any{
				"client_txn_id":    c.nextClientTxn("totp-complete-csrf-begin"),
				"current_password": password,
				"second_factor": map[string]any{
					"kind": "totp",
					"assertion": map[string]any{
						"code": generateTOTPCode(t, secretBase32),
					},
				},
			},
			withCookies(login.sessionCookie, login.csrfCookie),
			withHeader(authn.CSRFHeaderName, login.csrfCookie.Value),
		)
		beginData := httptestx.RequireSuccessEnvelope(t, beginResp, http.StatusOK)["data"].(map[string]any)
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id": c.nextClientTxn("totp-complete-csrf"),
				"enrollment_id": beginData["enrollment_id"].(string),
				"code":          generateTOTPCode(t, beginData["totp_setup"].(map[string]any)["secret_base32"].(string)),
			},
			rawBody:      malformedJSONBody(),
			cookies:      []*http.Cookie{login.sessionCookie, login.csrfCookie},
			headers:      map[string]string{},
			actorUserID:  userID,
			targetUserID: userID,
		}
	default:
		req := c.buildSuccessRequest(t, route)
		delete(req.headers, authn.CSRFHeaderName)
		if route.ID == routetest.RoutePasswordChange {
			req.rawBody = malformedJSONBody()
		}
		return req
	}
}

func (c *SupportRouteContext) buildBootstrapBoundaryRequest(
	t testing.TB,
	route routetest.RouteInventoryEntry,
	bootstrapToken string,
	targetUserID string,
	totpTargetID string,
	revokeTargetID string,
) SupportRequest {
	t.Helper()

	headers := map[string]string{"Authorization": "Bearer " + bootstrapToken}
	switch route.ID {
	case routetest.RouteSession, routetest.RouteLogout, routetest.RouteCredentialState:
		return SupportRequest{route: route, path: route.Template, headers: headers}
	case routetest.RoutePasswordChange:
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("bootstrap-password-change"),
				"current_password": "bootstrap-current",
				"new_password":     "bootstrap-next",
				"second_factor": map[string]any{
					"kind": "totp",
					"assertion": map[string]any{
						"code": "000000",
					},
				},
			},
			headers: headers,
		}
	case routetest.RouteUsersList:
		return SupportRequest{route: route, path: route.Template, headers: headers}
	case routetest.RouteUsersCreate:
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("bootstrap-users-create"),
				"auth_kind":        "local",
				"email":            "bootstrap-denied@example.test",
				"display_name":     "Bootstrap Denied",
				"initial_password": "BootstrapDenied1!",
			},
			headers: headers,
		}
	case routetest.RouteUsersGet:
		return SupportRequest{route: route, path: routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}), headers: headers}
	case routetest.RouteUsersPatch:
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"base_user_version": 1,
				"display_name":      "Bootstrap Denied Patch",
			},
			headers: headers,
		}
	case routetest.RouteUsersPasswordReset:
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body: map[string]any{
				"base_user_version": 1,
				"client_txn_id":     c.nextClientTxn("bootstrap-users-password-reset"),
				"new_password":      "BootstrapDeniedReset1!",
				"reason":            "bootstrap denied",
			},
			headers: headers,
		}
	case routetest.RouteUsersTOTPReset:
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: totpTargetID}),
			body: map[string]any{
				"base_user_version": 1,
				"client_txn_id":     c.nextClientTxn("bootstrap-users-totp-reset"),
				"reason":            "bootstrap denied",
			},
			headers: headers,
		}
	case routetest.RouteUsersRevokeAll:
		return SupportRequest{
			route: route,
			path:  routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: revokeTargetID}),
			body: map[string]any{
				"client_txn_id": c.nextClientTxn("bootstrap-users-revoke-all"),
				"reason":        "bootstrap denied",
			},
			headers: headers,
		}
	default:
		t.Fatalf("unsupported bootstrap-boundary route %s", route.ID)
		return SupportRequest{}
	}
}

func (c *SupportRouteContext) demotePrimaryAdmin(t testing.TB) {
	t.Helper()

	resp := doJSON(
		t,
		http.MethodPatch,
		c.server.HTTP.URL+"/api/v1/users/"+c.adminID,
		map[string]any{
			"base_user_version":   c.userVersion(t, c.adminID),
			"is_deployment_admin": false,
		},
		withCookies(c.peerAdminSession, c.peerAdminCSRF),
		withHeader(authn.CSRFHeaderName, c.peerAdminCSRF.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if got := data["is_deployment_admin"]; got != false {
		t.Fatalf("expected primary admin to be demoted, got %#v", data)
	}
}

func (c *SupportRouteContext) buildClosedVocabularyRequest(
	t testing.TB,
	route routetest.RouteInventoryEntry,
	field string,
) SupportRequest {
	t.Helper()

	switch route.ID {
	case routetest.RouteLogin:
		_, email, password := c.newLocalUser(t, "closed-vocabulary-login", false, false, true)
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"username": email,
				"password": password,
				"second_factor": map[string]any{
					"kind": "webauthn",
					"assertion": map[string]any{
						"code": "123456",
					},
				},
			},
			headers: map[string]string{},
		}
	case routetest.RoutePasswordChange:
		userID, _, password, _, login := c.newActiveTOTPLoggedInUser(t, "closed-vocabulary-password-change", false)
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("closed-vocabulary-password-change"),
				"current_password": password,
				"new_password":     "SupportPasswordChanged2!",
				"second_factor": map[string]any{
					"kind": "webauthn",
					"assertion": map[string]any{
						"code": "123456",
					},
				},
			},
			cookies:      []*http.Cookie{login.sessionCookie, login.csrfCookie},
			headers:      map[string]string{authn.CSRFHeaderName: login.csrfCookie.Value},
			actorUserID:  userID,
			targetUserID: userID,
		}
	case routetest.RouteTOTPBegin:
		userID, _, password, secretBase32, login := c.newActiveTOTPLoggedInUser(t, "closed-vocabulary-totp-begin", false)
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("closed-vocabulary-totp-begin"),
				"current_password": password,
				"second_factor": map[string]any{
					"kind": "webauthn",
					"assertion": map[string]any{
						"code": generateTOTPCode(t, secretBase32),
					},
				},
			},
			cookies:      []*http.Cookie{login.sessionCookie, login.csrfCookie},
			headers:      map[string]string{authn.CSRFHeaderName: login.csrfCookie.Value},
			actorUserID:  userID,
			targetUserID: userID,
		}
	case routetest.RouteUsersCreate:
		if field != "auth_kind" {
			t.Fatalf("unsupported closed-vocabulary field %s for route %s", field, route.ID)
		}
		email, password := c.nextIdentity("closed-vocabulary-users-create")
		return SupportRequest{
			route: route,
			path:  route.Template,
			body: map[string]any{
				"client_txn_id":    c.nextClientTxn("closed-vocabulary-users-create"),
				"auth_kind":        "ldap",
				"email":            email,
				"display_name":     "Closed Vocabulary Create",
				"initial_password": password,
			},
			cookies: []*http.Cookie{c.adminSession, c.adminCSRF},
			headers: map[string]string{authn.CSRFHeaderName: c.adminCSRF.Value},
		}
	default:
		t.Fatalf("unsupported closed-vocabulary route %s", route.ID)
		return SupportRequest{}
	}
}

func (c *SupportRouteContext) closedVocabularyCode(route routetest.RouteInventoryEntry, field string) string {
	switch route.ID {
	case routetest.RouteLogin, routetest.RoutePasswordChange, routetest.RouteTOTPBegin:
		return "invalid_auth_request"
	case routetest.RouteUsersCreate:
		return "invalid_mutation_payload"
	default:
		panic("unsupported closed-vocabulary route " + string(route.ID) + " field " + field)
	}
}

func (c *SupportRouteContext) requireWritableStringNormalization(
	t testing.TB,
	route routetest.RouteInventoryEntry,
	field string,
) {
	t.Helper()

	switch route.ID {
	case routetest.RouteUsersCreate:
		email, password := c.nextIdentity("writable-users-create")
		body := map[string]any{
			"client_txn_id":    c.nextClientTxn("writable-users-create"),
			"auth_kind":        "local",
			"email":            email,
			"display_name":     "Writable Create",
			"initial_password": password,
		}
		var want string
		switch field {
		case "email":
			body["email"] = "  " + email + "  "
			want = email
		case "display_name":
			body["display_name"] = "  Writable Create  "
			want = "Writable Create"
		default:
			t.Fatalf("unsupported writable-string field %s for route %s", field, route.ID)
		}
		resp := doJSON(
			t,
			route.Method,
			c.server.HTTP.URL+route.Template,
			body,
			withCookies(c.adminSession, c.adminCSRF),
			withHeader(authn.CSRFHeaderName, c.adminCSRF.Value),
		)
		data := httptestx.RequireSuccessEnvelope(t, resp, route.SuccessStatus)["data"].(map[string]any)
		got, ok := data[field].(string)
		if !ok {
			t.Fatalf("expected %s response field for route %s, got %#v", field, route.ID, data)
		}
		contractassert.RequireWritableStringNormalization(t, got, want)
	case routetest.RouteUsersPatch:
		targetUserID, _, _ := c.newLocalUser(t, "writable-users-patch-target", false, false, true)
		body := map[string]any{
			"base_user_version": 1,
		}
		var want string
		switch field {
		case "email":
			body["email"] = "  WritablePatch@Example.Test  "
			want = "WritablePatch@Example.Test"
		case "display_name":
			body["display_name"] = "  Writable Patch  "
			want = "Writable Patch"
		default:
			t.Fatalf("unsupported writable-string field %s for route %s", field, route.ID)
		}
		resp := doJSON(
			t,
			route.Method,
			c.server.HTTP.URL+routetest.BuildRoutePath(route.Template, routetest.RouteInventoryFixture{UserID: targetUserID}),
			body,
			withCookies(c.adminSession, c.adminCSRF),
			withHeader(authn.CSRFHeaderName, c.adminCSRF.Value),
		)
		data := httptestx.RequireSuccessEnvelope(t, resp, route.SuccessStatus)["data"].(map[string]any)
		got, ok := data[field].(string)
		if !ok {
			t.Fatalf("expected %s response field for route %s, got %#v", field, route.ID, data)
		}
		contractassert.RequireWritableStringNormalization(t, got, want)
	case routetest.RouteUsersPasswordReset, routetest.RouteUsersTOTPReset, routetest.RouteUsersRevokeAll:
		if field != "reason" {
			t.Fatalf("unsupported writable-string field %s for route %s", field, route.ID)
		}
		req := c.buildSuccessRequest(t, route)
		body := cloneJSONMap(t, req.body)
		body["reason"] = "  Support reason normalization  "
		req.body = body

		rawReason := body["reason"].(string)
		normalizedReason := authn.NormalizeReasonNote(&rawReason)
		if normalizedReason == nil {
			t.Fatalf("expected normalized reason note for %q", rawReason)
			return
		}
		contractassert.RequireWritableStringNormalization(t, *normalizedReason, "Support reason normalization")

		firstResp := c.do(t, req)
		firstData := httptestx.RequireSuccessEnvelope(t, firstResp, route.SuccessStatus)["data"].(map[string]any)

		replayReq := req
		replayBody := cloneJSONMap(t, req.body)
		replayBody["reason"] = *normalizedReason
		replayReq.body = replayBody
		replayedResp := c.do(t, replayReq)
		replayedData := httptestx.RequireSuccessEnvelope(t, replayedResp, route.SuccessStatus)["data"].(map[string]any)
		requireJSONEquivalent(t, replayedData, firstData)
		if got := queryCount(t, c.db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND actor_user_id::text = $2 AND scope_key = $3 AND client_txn_id = $4`, req.routeKey, req.actorUserID, req.scopeKey, req.clientTxnID); got != 1 {
			t.Fatalf("expected one route_idempotency row for %s, got %d", route.ID, got)
		}
	default:
		t.Fatalf("unsupported writable-string route %s", route.ID)
	}
}

func (c *SupportRouteContext) do(t testing.TB, req SupportRequest) *http.Response {
	t.Helper()

	options := make([]func(*http.Request), 0, len(req.headers)+1)
	if len(req.cookies) > 0 {
		options = append(options, withCookies(req.cookies...))
	}
	for key, value := range req.headers {
		options = append(options, withHeader(key, value))
	}
	if req.rawBody != nil {
		httpReq, err := http.NewRequest(req.route.Method, c.server.HTTP.URL+req.path, strings.NewReader(*req.rawBody))
		if err != nil {
			t.Fatalf("create raw json request: %v", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		for _, option := range options {
			option(httpReq)
		}
		return httptestx.Do(t, http.DefaultClient, httpReq)
	}
	return doJSON(t, req.route.Method, c.server.HTTP.URL+req.path, req.body, options...)
}

func malformedJSONBody() *string {
	body := "{"
	return &body
}

func (c *SupportRouteContext) userVersion(t testing.TB, userID string) int64 {
	t.Helper()

	var version int64
	if err := c.db.QueryRowContext(context.Background(), `SELECT user_version FROM users WHERE id::text = $1`, userID).Scan(&version); err != nil {
		t.Fatalf("query user version for %s: %v", userID, err)
	}
	return version
}

func (c *SupportRouteContext) newLocalUser(t testing.TB, tag string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) (string, string, string) {
	t.Helper()

	email, password := c.nextIdentity(tag)
	displayName := fmt.Sprintf("Support %s %02d", strings.ReplaceAll(tag, "-", " "), c.sequence)
	userID := seedLocalUserFlags(t, c.db, email, displayName, password, mfaRequired, isDeploymentAdmin, isActive)
	return userID, email, password
}

func (c *SupportRouteContext) newLoggedInLocalUser(
	t testing.TB,
	tag string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	isActive bool,
) (string, string, string, *http.Cookie, *http.Cookie) {
	t.Helper()

	userID, email, password := c.newLocalUser(t, tag, mfaRequired, isDeploymentAdmin, isActive)
	sessionCookie, csrfCookie := loginLocalUser(t, c.server, email, password, nil)
	return userID, email, password, sessionCookie, csrfCookie
}

func (c *SupportRouteContext) newActiveTOTPLoggedInUser(
	t testing.TB,
	tag string,
	isDeploymentAdmin bool,
) (string, string, string, string, loginResult) {
	t.Helper()

	c.sequence++
	email := fmt.Sprintf("%s-totp-%02d@example.test", tag, c.sequence)
	password := fmt.Sprintf("SupportTotpPass%02d!Aa", c.sequence)
	displayName := fmt.Sprintf("Support TOTP %02d", c.sequence)
	secretBase32 := supportSecretBase32(c.sequence)
	userID := seedLocalUserWithActiveTOTP(t, c.db, email, displayName, password, true, isDeploymentAdmin, secretBase32)
	login := loginLocalUserWithSecondFactor(t, c.server, email, password, generateTOTPCode(t, secretBase32))
	return userID, email, password, secretBase32, login
}

func (c *SupportRouteContext) newSocketSession(
	t testing.TB,
	tag string,
	activeTOTP bool,
) (string, string, string, *http.Cookie, *http.Cookie, *flowtest.SessionSocketClient) {
	t.Helper()

	if activeTOTP {
		userID, _, _, _, login := c.newActiveTOTPLoggedInUser(t, tag, false)
		incidentID := incidentstoretest.SeedIncidentMembershipSQL(t, c.db, userID, tag+"-socket")
		socket := connectSessionSocket(t, c.server, incidentID, login.sessionCookie.Value)
		return userID, "", "", login.sessionCookie, login.csrfCookie, socket
	}

	userID, _, _, sessionCookie, csrfCookie := c.newLoggedInLocalUser(t, tag, false, false, true)
	incidentID := incidentstoretest.SeedIncidentMembershipSQL(t, c.db, userID, tag+"-socket")
	socket := connectSessionSocket(t, c.server, incidentID, sessionCookie.Value)
	return userID, "", "", sessionCookie, csrfCookie, socket
}

func (c *SupportRouteContext) nextClientTxn(tag string) string {
	c.sequence++
	return fmt.Sprintf("txn-%s-%02d", tag, c.sequence)
}

func (c *SupportRouteContext) nextIdentity(tag string) (string, string) {
	c.sequence++
	email := fmt.Sprintf("%s-%02d@example.test", tag, c.sequence)
	password := fmt.Sprintf("SupportPass%02d!Aa", c.sequence)
	return email, password
}

func cloneJSONMap(t testing.TB, body any) map[string]any {
	t.Helper()

	typed, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("expected json object body, got %T", body)
	}
	cloned := make(map[string]any, len(typed))
	for key, value := range typed {
		cloned[key] = value
	}
	return cloned
}

func supportSecretBase32(seed int) string {
	secrets := []string{
		"JBSWY3DPEHPK3PXP",
		"JBSWY3DPEHPK3QAA",
		"JBSWY3DPEHPK3QAB",
		"JBSWY3DPEHPK3QAC",
		"JBSWY3DPEHPK3QAD",
		"JBSWY3DPEHPK3QAE",
		"JBSWY3DPEHPK3QAF",
		"JBSWY3DPEHPK3QAG",
	}
	return secrets[seed%len(secrets)]
}

func queryPendingTOTPEnrollmentSecretMaterial(t testing.TB, db *sql.DB, userID string, clientTxnID string) ([]byte, []byte) {
	t.Helper()

	var ciphertext []byte
	var nonce []byte
	if err := db.QueryRowContext(context.Background(), `
SELECT secret_ciphertext, secret_nonce
  FROM pending_totp_enrollments
 WHERE user_id::text = $1
   AND client_txn_id = $2
 ORDER BY created_at DESC
 LIMIT 1
`, userID, clientTxnID).Scan(&ciphertext, &nonce); err != nil {
		t.Fatalf("query pending totp material: %v", err)
	}
	return ciphertext, nonce
}

func queryUserTOTPSecretMaterial(t testing.TB, db *sql.DB, userID string) ([]byte, []byte) {
	t.Helper()

	var ciphertext []byte
	var nonce []byte
	if err := db.QueryRowContext(context.Background(), `
SELECT totp_secret_ciphertext, totp_secret_nonce
  FROM users
 WHERE id::text = $1
`, userID).Scan(&ciphertext, &nonce); err != nil {
		t.Fatalf("query user totp material: %v", err)
	}
	return ciphertext, nonce
}

func requireEncryptedSecretMaterial(t testing.TB, ciphertext []byte, nonce []byte, clearSecretBase32 string) {
	t.Helper()

	clearSecret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(clearSecretBase32)
	if err != nil {
		t.Fatalf("decode clear totp secret: %v", err)
	}
	if len(ciphertext) == 0 || len(nonce) == 0 {
		t.Fatalf("expected encrypted secret material, got ciphertext=%d nonce=%d", len(ciphertext), len(nonce))
	}
	if string(ciphertext) == string(clearSecret) {
		t.Fatal("expected ciphertext to differ from clear secret material")
	}
}

func indexOfUser(rows []any, userID string) int {
	for index, row := range rows {
		candidate := row.(map[string]any)
		if candidate["user_id"] == userID {
			return index
		}
	}
	return -1
}

func seedFixedLocalUser(t testing.TB, db *sql.DB, userID string, email string, displayName string, password string, isDeploymentAdmin bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	parsedID := uuid.MustParse(userID)
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, false, true, $5)
`, parsedID, email, displayName, hash, isDeploymentAdmin); err != nil {
		t.Fatalf("seed fixed user: %v", err)
	}
	return userID
}

func requirePagingCursor(t testing.TB, envelope map[string]any) string {
	t.Helper()

	meta := envelope["meta"].(map[string]any)
	paging := meta["paging"].(map[string]any)
	token, ok := paging["next_cursor"].(string)
	if !ok || token == "" {
		t.Fatalf("expected next cursor, got %#v", paging["next_cursor"])
	}
	return token
}

func findUserResource(t testing.TB, rows []any, userID string) map[string]any {
	t.Helper()

	for _, row := range rows {
		candidate := row.(map[string]any)
		if candidate["user_id"] == userID {
			return candidate
		}
	}
	t.Fatalf("expected user %s in %#v", userID, rows)
	return nil
}
