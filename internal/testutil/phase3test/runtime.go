package phase3test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	workbookstartupbootstrap "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

type LoginResult struct {
	SessionCookie *http.Cookie
	CSRFCookie    *http.Cookie
}

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func DoJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func DoRawJSON(t testing.TB, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create raw JSON request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func WithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func WithHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func ProvisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (LoginResult, uuid.UUID) {
	t.Helper()

	bootstrapToken := RequireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-phase3-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-phase3-bootstrap-admin-complete")
	login := LoginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", GenerateTOTPCode(t, secretBase32))

	sessionResp := DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, WithCookies(login.SessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, MustUUID(t, sessionData["user_id"].(string))
}

func CreateIncident(t testing.TB, server *httptestx.Server, admin LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		WithCookies(admin.SessionCookie, admin.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateIncidentInStore(t testing.TB, pool postgres.DB, actor authn.UserRecord, clientTxnID string, incidentKey string, title string) incidents.IncidentRecord {
	t.Helper()

	store := incidents.NewStoreWithOptions(pool, incidents.StoreOptions{
		WorkbookBootstrap: workbookstartupbootstrap.NewIncidentCreatePreferencesPort(),
	})
	result, err := store.CreateIncident(context.Background(), actor, incidents.CreateIncidentRequest{
		ClientTxnID: clientTxnID,
		IncidentKey: incidentKey,
		Title:       title,
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Now().UTC())
	if err != nil {
		t.Fatalf("create incident in store: %v", err)
	}
	return result.Incident
}

func CreateTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, actor LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timelineViewSchemaID+"/rows",
		body,
		WithCookies(actor.SessionCookie, actor.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, email string, role string, admin LoginResult) {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-phase3-membership-create-" + userID,
			"email":         email,
			"role":          role,
		},
		WithCookies(admin.SessionCookie, admin.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	responseBody := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	if responseBody["user_id"] != userID {
		t.Fatalf("unexpected membership create payload: %#v", responseBody)
	}
}

func UpdateMembershipRole(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int, role string, admin LoginResult) {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": baseVersion,
			"role":                    role,
		},
		WithCookies(admin.SessionCookie, admin.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func DeleteMembership(t testing.TB, server *httptestx.Server, incidentID string, userID string, baseVersion int64, admin LoginResult) {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodDelete,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": baseVersion,
		},
		WithCookies(admin.SessionCookie, admin.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	httptestx.RequireStatus(t, resp, http.StatusNoContent)
}

func SeedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) authn.UserRecord {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var record authn.UserRecord
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, created_at, updated_at, user_version
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(
		&record.ID,
		&record.Email,
		&record.DisplayName,
		&record.PasswordHash,
		&record.MFARequired,
		&record.IsActive,
		&record.IsDeploymentAdmin,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UserVersion,
	); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return record
}

func LoginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return sessionCookie, csrfCookie
}

func LoginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) LoginResult {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func MustUUID(t testing.TB, raw string) uuid.UUID {
	t.Helper()

	value, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return value
}

func QueryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func QueryTimelineEnvelope(t testing.TB, server *httptestx.Server, incidentID string, login LoginResult, body map[string]any) map[string]any {
	t.Helper()

	queryResp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		body,
		WithCookies(login.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)
}

func QueryTimelineRows(t testing.TB, server *httptestx.Server, incidentID string, login LoginResult) []any {
	t.Helper()

	return QueryTimelineEnvelope(t, server, incidentID, login, map[string]any{})["data"].(map[string]any)["rows"].([]any)
}

func FindRow(t testing.TB, rows []any, recordID string) map[string]any {
	t.Helper()

	for _, candidate := range rows {
		row := candidate.(map[string]any)
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("record_id %s not found in rows %#v", recordID, rows)
	return nil
}

func RequireNoTimelineCollaborationEmission(t testing.TB, client *TimelineSocketClient, changes <-chan platformws.RecordChange) {
	t.Helper()

	ExpectNoTimelineSocketMessage(t, client)
	timelinetest.RequireNoRecordChange(t, changes, 300*time.Millisecond)
}

func RequireMutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	changeSet := timelinetest.LookupChangeSet(t, db, changeSetID)
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, wantActorUserID, wantSource, wantClientTxnID)
	if got := timelinetest.CountChangeSetMutations(t, db, changeSetID); got != wantMutationRows {
		t.Fatalf("unexpected mutation row count for %s: got %d want %d", changeSetID, got, wantMutationRows)
	}
	if got := timelinetest.CountRecordRevisions(t, db, recordID); got != wantRevisions {
		t.Fatalf("unexpected record revision count for %s: got %d want %d", recordID, got, wantRevisions)
	}
}

func RequireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func BeginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, WithHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func CompleteInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          GenerateTOTPCode(t, secretBase32),
	}, WithHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func GenerateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}
