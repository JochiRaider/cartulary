package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase1_UserListUsesSnapshotStablePagination_I_1_07(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-07-users-pagination")
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
	snapshotUser := findUserResource(t, continuedUsers, userTwoID)
	if snapshotUser["display_name"] != "Pagination Two" {
		t.Fatalf("expected snapshot-stable user payload, got %#v", snapshotUser)
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
