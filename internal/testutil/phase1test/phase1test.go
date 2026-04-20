package phase1test

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

type LoginResult struct {
	SessionCookie *http.Cookie
	CSRFCookie    *http.Cookie
}

type SessionRow struct {
	AuthenticatedAt          time.Time
	SessionID                string
	LastQualifyingActivityAt time.Time
	IdleExpiresAt            time.Time
	AbsoluteExpiresAt        time.Time
	SessionExpiresAt         time.Time
	RevokedAt                sql.NullTime
	RevokeReasonCode         sql.NullString
}

type AuditEventRecord struct {
	EventKind    string
	ActorUserID  string
	TargetUserID string
	EventSource  string
	ReasonCode   string
	ClientTxnID  string
	RequestID    string
	CreatedAt    time.Time
	Before       map[string]any
	After        map[string]any
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

func SeedLocalUser(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool) string {
	t.Helper()
	return SeedLocalUserFlags(t, db, email, displayName, password, mfaRequired, false, true)
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

func SeedLocalUserWithActiveTOTP(
	t testing.TB,
	db *sql.DB,
	email string,
	displayName string,
	password string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	secretBase32 string,
) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load auth master keys: %v", err)
	}
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretBase32)
	if err != nil {
		t.Fatalf("decode base32 totp secret: %v", err)
	}
	ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
	if err != nil {
		t.Fatalf("encrypt totp secret: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce)
VALUES ($1, $2, $3, $4, true, $5, now(), $6, $7)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isDeploymentAdmin, ciphertext, nonce).Scan(&userID); err != nil {
		t.Fatalf("seed local user with totp: %v", err)
	}
	return userID
}

func LoginLocalUser(t testing.TB, baseURL string, username string, password string, headers func(*http.Request)) (*http.Cookie, *http.Cookie) {
	t.Helper()

	req := httptestx.NewJSONRequest(t, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if headers != nil {
		headers(req)
	}
	resp := httptestx.Do(t, http.DefaultClient, req)
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

func LoginLocalUserWithSecondFactor(t testing.TB, baseURL string, username string, password string, code string) LoginResult {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
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

func RequireBootstrapLogin(t testing.TB, baseURL string, username string, password string) string {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	token, _ := details["bootstrap_token"].(string)
	if token == "" {
		t.Fatalf("expected bootstrap_token on mfa_setup_required response, got %#v", details)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			t.Fatal("mfa_setup_required must not set a session cookie")
		}
	}
	return token
}

func BeginTOTPEnrollment(t testing.TB, baseURL string, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/auth/mfa/totp/begin",
		body,
		WithHeader("Authorization", "Bearer "+bootstrapToken),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func CompleteInitialEnrollment(t testing.TB, baseURL string, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/auth/mfa/totp/complete",
		map[string]any{
			"client_txn_id": clientTxnID,
			"enrollment_id": enrollmentID,
			"code":          GenerateTOTPCode(t, secretBase32),
		},
		WithHeader("Authorization", "Bearer "+bootstrapToken),
	)
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

func ConnectSessionSocket(t testing.TB, serverURL string, sessionToken string) *wstest.Client {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	client := wstest.ConnectWithHeaders(t, serverURL, "/ws/v1/test/session-lifecycle", headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connected, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("read connected websocket message: %v", err)
	}
	wstest.RequireMessageType(t, connected, "connected")
	return client
}

func ExpectSessionRevoked(t testing.TB, client *wstest.Client, wantReasonCode string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	revoked, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("read session_revoked message: %v", err)
	}
	wstest.RequireSessionRevoked(t, revoked)

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		t.Fatalf("decode session_revoked payload: %v", err)
	}
	if payload["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected session_revoked payload: %#v", payload)
	}

	_, err = client.Receive(ctx)
	wstest.RequireClose(t, err, websocket.StatusPolicyViolation, "")
}

func RequireBootstrapWebsocketRejected(t testing.TB, serverURL string, bootstrapToken string) {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+bootstrapToken)

	_, resp, err := wstest.TryConnect(serverURL, "/ws/v1/test/session-lifecycle", headers)
	if err == nil {
		t.Fatal("expected bootstrap-token websocket dial to fail")
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for bootstrap-token websocket dial, err=%v", err)
	}
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected websocket bootstrap rejection: %#v", details)
	}
}

func QuerySessionRow(t testing.TB, db *sql.DB, userID string) SessionRow {
	t.Helper()
	return QuerySingleSession(
		t,
		db,
		`SELECT authenticated_at, id::text, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID,
	)
}

func QuerySessionByID(t testing.TB, db *sql.DB, sessionID string) SessionRow {
	t.Helper()
	return QuerySingleSession(
		t,
		db,
		`SELECT authenticated_at, id::text, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE id::text = $1`,
		sessionID,
	)
}

func QuerySingleSession(t testing.TB, db *sql.DB, query string, args ...any) SessionRow {
	t.Helper()

	var row SessionRow
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(
		&row.AuthenticatedAt,
		&row.SessionID,
		&row.LastQualifyingActivityAt,
		&row.IdleExpiresAt,
		&row.AbsoluteExpiresAt,
		&row.SessionExpiresAt,
		&row.RevokedAt,
		&row.RevokeReasonCode,
	); err != nil {
		t.Fatalf("query session row: %v", err)
	}
	return row
}

func QueryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func LookupUserAuditEvents(t testing.TB, db *sql.DB, targetUserID string) []AuditEventRecord {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT event_kind,
       actor_user_id::text,
       target_user_id::text,
       event_source,
       reason_code,
       COALESCE(client_txn_id, ''),
       COALESCE(request_id, ''),
       created_at,
       before_json,
       after_json
  FROM deployment_admin_audit_events
 WHERE target_user_id::text = $1
 ORDER BY created_at ASC
`, targetUserID)
	if err != nil {
		t.Fatalf("query user audit events: %v", err)
	}
	defer rows.Close()

	events := make([]AuditEventRecord, 0, 4)
	for rows.Next() {
		var (
			record     AuditEventRecord
			beforeJSON []byte
			afterJSON  []byte
		)
		if err := rows.Scan(
			&record.EventKind,
			&record.ActorUserID,
			&record.TargetUserID,
			&record.EventSource,
			&record.ReasonCode,
			&record.ClientTxnID,
			&record.RequestID,
			&record.CreatedAt,
			&beforeJSON,
			&afterJSON,
		); err != nil {
			t.Fatalf("scan user audit event: %v", err)
		}
		record.Before = DecodeJSONMap(t, beforeJSON)
		record.After = DecodeJSONMap(t, afterJSON)
		events = append(events, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate user audit events: %v", err)
	}
	return events
}

func DecodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	if len(payload) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode json map: %v", err)
	}
	return decoded
}

func RequireAuditEventBySource(t testing.TB, events []AuditEventRecord, source string) AuditEventRecord {
	t.Helper()
	for _, event := range events {
		if event.EventSource == source {
			return event
		}
	}
	t.Fatalf("expected audit event source %q in %#v", source, events)
	return AuditEventRecord{}
}

func CreateIncidentResource(t testing.TB, baseURL string, sessionCookie *http.Cookie, csrfCookie *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/incidents",
		body,
		WithCookies(sessionCookie, csrfCookie),
		WithHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateIncidentMembership(t testing.TB, baseURL string, incidentID string, sessionCookie *http.Cookie, csrfCookie *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/incidents/"+incidentID+"/memberships",
		body,
		WithCookies(sessionCookie, csrfCookie),
		WithHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}
