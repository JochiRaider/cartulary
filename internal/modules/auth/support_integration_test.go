package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestUserListContinuationUsesLiveRows(t *testing.T) {
	runtime := appsupport.StartRuntime(t)

	server, db := startServer(t, runtime, "authentication-support-pagination")
	defer db.Close()

	seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000701", "pagination-admin@example.test", "Pagination Admin", "PaginationAdmin1!", true)
	_ = flowtest.SeedActor(t, db, flowtest.ActorAttributes{Email: "pagination-one@example.test", DisplayName: "Pagination One", ID: uuid.MustParse("10000000-0000-0000-0000-000000000702")}).ID.String()
	userTwoID := flowtest.SeedActor(t, db, flowtest.ActorAttributes{Email: "pagination-two@example.test", DisplayName: "Pagination Two", ID: uuid.MustParse("10000000-0000-0000-0000-000000000703")}).ID.String()

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
	runtime := appsupport.StartRuntime(t)

	server, db := startServer(t, runtime, "authentication-support-user-list-search")
	defer db.Close()

	seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000711", "search-viewer@example.test", "Search Viewer", "SearchViewer1!", false)
	// An unrelated row may sort before the named viewer, including the bootstrap
	// admin. Only the searched user's position beyond the first page matters.
	flowtest.SeedActor(t, db, flowtest.ActorAttributes{Email: "search-earlier@example.test", DisplayName: "Earlier User", ID: uuid.MustParse("00000000-0000-4000-8000-000000000710")})
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
	if len(firstPageRows) != 1 || firstPageRows[0].(map[string]any)["user_id"] == adminID {
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

	hash, err := authn.HashPassword(context.Background(), password)
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
