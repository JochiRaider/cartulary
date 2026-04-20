package phase2test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type LoginResult struct {
	SessionCookie *http.Cookie
	CSRFCookie    *http.Cookie
}

type AuditEventRecord struct {
	EventKind   string
	ActorUserID string
	EventSource string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
	Before      map[string]any
	After       map[string]any
}

func DoJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
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

func ProvisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (LoginResult, string) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	login := LoginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, WithCookies(login.SessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func CreateIncident(t testing.TB, server *httptestx.Server, actor LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		WithCookies(actor.SessionCookie, actor.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateTimelineRow(t testing.TB, server *httptestx.Server, actor LoginResult, incidentID string, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		WithCookies(actor.SessionCookie, actor.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func SeedLocalUserFlags(
	t testing.TB,
	db *sql.DB,
	email string,
	displayName string,
	password string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	isActive bool,
) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return userID
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

func QueryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func LookupAuditEvents(t testing.TB, db *sql.DB, incidentID string) []AuditEventRecord {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT event_kind,
       actor_user_id::text,
       event_source,
       COALESCE(client_txn_id, ''),
       COALESCE(request_id, ''),
       created_at,
       before_json,
       after_json
  FROM deployment_admin_audit_events
 WHERE incident_id::text = $1
 ORDER BY created_at ASC
`, incidentID)
	if err != nil {
		t.Fatalf("query audit events: %v", err)
	}
	defer rows.Close()

	events := make([]AuditEventRecord, 0, 4)
	for rows.Next() {
		var (
			record        AuditEventRecord
			beforePayload []byte
			afterPayload  []byte
		)
		if err := rows.Scan(&record.EventKind, &record.ActorUserID, &record.EventSource, &record.ClientTxnID, &record.RequestID, &record.CreatedAt, &beforePayload, &afterPayload); err != nil {
			t.Fatalf("scan audit event: %v", err)
		}
		record.Before = decodeJSONMap(t, beforePayload)
		record.After = decodeJSONMap(t, afterPayload)
		events = append(events, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate audit events: %v", err)
	}
	return events
}

func RequireTimelineSocketRejected(
	t testing.TB,
	serverURL string,
	incidentID string,
	sessionToken string,
	wantStatus int,
	wantCode string,
) {
	t.Helper()

	target, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	target.Scheme = strings.Replace(target.Scheme, "http", "ws", 1)
	target.Path = "/ws/v1/incidents/" + incidentID + "/views/cartulary.view.timeline.v1/changes"

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, target.String(), &websocket.DialOptions{HTTPHeader: headers})
	if err == nil {
		t.Fatal("expected websocket dial to fail")
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for websocket dial, err=%v", err)
	}
	httptestx.RequireErrorEnvelope(t, resp, wantStatus, wantCode)
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, WithHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, WithHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
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

func decodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	if len(payload) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode JSON payload: %v", err)
	}
	return decoded
}

func ReadBodyString(t testing.TB, body io.ReadCloser) string {
	t.Helper()
	defer body.Close()
	payload, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(payload)
}
