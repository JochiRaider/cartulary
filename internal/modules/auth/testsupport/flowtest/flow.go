package flowtest

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
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

func ProvisionBootstrapAdmin(t testing.TB, baseURL string) (LoginResult, string) {
	t.Helper()

	bootstrapToken := RequireBootstrapLogin(t, baseURL, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := BeginTOTPEnrollment(t, baseURL, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteInitialEnrollment(t, baseURL, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	login := LoginLocalUserWithSecondFactor(t, baseURL, "bootstrap-admin@example.test", "BootstrapPass1!", GenerateTOTPCode(t, secretBase32))

	sessionResp := DoJSON(t, http.MethodGet, baseURL+"/api/v1/auth/session", nil, WithCookies(login.SessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func ProvisionBootstrapAdminUUID(t testing.TB, baseURL string) (LoginResult, uuid.UUID) {
	t.Helper()

	login, userID := ProvisionBootstrapAdmin(t, baseURL)
	parsed, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("parse bootstrap admin id: %v", err)
	}
	return login, parsed
}

type SessionSocketClient = incidentwstest.Client

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

func WithTestRouteToken() func(*http.Request) {
	return WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)
}

func SetClockOffset(
	t testing.TB,
	baseURL string,
	offsetSeconds int64,
	options ...func(*http.Request),
) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/test/clock/set",
		map[string]any{
			"offset_seconds": offsetSeconds,
		},
		append([]func(*http.Request){WithTestRouteToken()}, options...)...,
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func ResetClockOffset(
	t testing.TB,
	baseURL string,
	options ...func(*http.Request),
) {
	t.Helper()
	SetClockOffset(t, baseURL, 0, options...)
}

func WithClockOffset(
	t testing.TB,
	baseURL string,
	offsetSeconds int64,
	options ...func(*http.Request),
) {
	t.Helper()
	SetClockOffset(t, baseURL, offsetSeconds, options...)
	t.Cleanup(func() {
		ResetClockOffset(t, baseURL, options...)
	})
}

func SeedLocalUser(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool) string {
	t.Helper()
	return SeedLocalUserFlags(t, db, email, displayName, password, mfaRequired, false, true)
}

func SeedLocalUserRecord(
	t testing.TB,
	db *sql.DB,
	email string,
	displayName string,
	password string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	isActive bool,
) authn.UserRecord {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var record authn.UserRecord
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(
		&record.ID,
		&record.Email,
		&record.DisplayName,
		&record.PasswordHash,
		&record.PasswordChangedAt,
		&record.MFARequired,
		&record.IsActive,
		&record.IsDeploymentAdmin,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UpdatedByUserID,
		&record.LastLoginAt,
		&record.UserVersion,
		&record.TOTPEnrolledAt,
		&record.TOTPSecretCiphertext,
		&record.TOTPSecretNonce,
	); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return record
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
	return SeedLocalUserRecord(t, db, email, displayName, password, mfaRequired, isDeploymentAdmin, isActive).ID.String()
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

	return SeedLocalUserWithActiveTOTPRecord(t, db, email, displayName, password, mfaRequired, isDeploymentAdmin, secretBase32).ID.String()
}

func SeedLocalUserWithActiveTOTPRecord(
	t testing.TB,
	db *sql.DB,
	email string,
	displayName string,
	password string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	secretBase32 string,
) authn.UserRecord {
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

	var record authn.UserRecord
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce)
VALUES ($1, $2, $3, $4, true, $5, now(), $6, $7)
RETURNING id, email, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, email, displayName, hash, mfaRequired, isDeploymentAdmin, ciphertext, nonce).Scan(
		&record.ID,
		&record.Email,
		&record.DisplayName,
		&record.PasswordHash,
		&record.PasswordChangedAt,
		&record.MFARequired,
		&record.IsActive,
		&record.IsDeploymentAdmin,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UpdatedByUserID,
		&record.LastLoginAt,
		&record.UserVersion,
		&record.TOTPEnrolledAt,
		&record.TOTPSecretCiphertext,
		&record.TOTPSecretNonce,
	); err != nil {
		t.Fatalf("seed local user with totp: %v", err)
	}
	return record
}

func SeedSession(
	t testing.TB,
	db *sql.DB,
	keys authn.MasterKeys,
	userID uuid.UUID,
	sessionToken string,
	authenticatedAt time.Time,
	lastQualifyingActivityAt time.Time,
) authn.SessionRecord {
	t.Helper()

	idleExpiresAt := lastQualifyingActivityAt.UTC().Add(authn.SessionIdleTTL)
	absoluteExpiresAt := authenticatedAt.UTC().Add(authn.SessionAbsoluteTTL)
	sessionExpiresAt := idleExpiresAt
	if absoluteExpiresAt.Before(sessionExpiresAt) {
		sessionExpiresAt = absoluteExpiresAt
	}

	var session authn.SessionRecord
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO user_sessions (
    user_id,
    token_fingerprint,
    authenticated_at,
    last_qualifying_activity_at,
    idle_expires_at,
    absolute_expires_at,
    session_expires_at,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $3, $3)
RETURNING id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
          session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at
`,
		userID,
		authn.FingerprintToken(keys, sessionToken),
		authenticatedAt.UTC(),
		lastQualifyingActivityAt.UTC(),
		idleExpiresAt,
		absoluteExpiresAt,
		sessionExpiresAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.AuthenticatedAt,
		&session.LastQualifyingActivityAt,
		&session.IdleExpiresAt,
		&session.AbsoluteExpiresAt,
		&session.SessionExpiresAt,
		&session.RevokedAt,
		&session.RevokeReasonCode,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	return session
}

func SeedPendingTOTPEnrollment(
	t testing.TB,
	db *sql.DB,
	keys authn.MasterKeys,
	userID uuid.UUID,
	sessionID *uuid.UUID,
	bootstrapTokenID *uuid.UUID,
	clientTxnID string,
	secretBytes []byte,
	replacesActive bool,
	now time.Time,
) authn.PendingTOTPEnrollmentRecord {
	t.Helper()

	ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
	if err != nil {
		t.Fatalf("encrypt pending totp secret: %v", err)
	}

	authScopeKind := "bootstrap_token"
	if sessionID != nil {
		authScopeKind = "session"
	}

	var record authn.PendingTOTPEnrollmentRecord
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO pending_totp_enrollments (
    user_id,
    auth_scope_kind,
    auth_scope_session_id,
    auth_scope_bootstrap_token_id,
    client_txn_id,
    secret_ciphertext,
    secret_nonce,
    replaces_active,
    created_at,
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
          secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
`,
		userID,
		authScopeKind,
		sessionID,
		bootstrapTokenID,
		clientTxnID,
		ciphertext,
		nonce,
		replacesActive,
		now.UTC(),
		now.UTC().Add(authn.PendingTOTPEnrollmentTTL),
	).Scan(
		&record.ID,
		&record.UserID,
		&record.AuthScopeKind,
		&record.AuthScopeSessionID,
		&record.AuthScopeBootstrapTokenID,
		&record.ClientTxnID,
		&record.SecretCiphertext,
		&record.SecretNonce,
		&record.ReplacesActive,
		&record.CreatedAt,
		&record.ExpiresAt,
		&record.ConsumedAt,
	); err != nil {
		t.Fatalf("seed pending totp enrollment: %v", err)
	}
	return record
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
	authCookies := httptestx.RequireAuthCookies(t, resp.Cookies())
	return authCookies.Session, authCookies.CSRF
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
	authCookies := httptestx.RequireAuthCookies(t, resp.Cookies())
	return LoginResult{SessionCookie: authCookies.Session, CSRFCookie: authCookies.CSRF}
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

func ConnectSessionSocket(t testing.TB, serverURL string, incidentID string, sessionToken string) *SessionSocketClient {
	t.Helper()

	return incidentwstest.ConnectAndHello(t, serverURL, incidentID, incidentwstest.ConnectOptions{
		SessionToken: sessionToken,
	})
}

func AwaitSessionRevoked(client *SessionSocketClient, wantReasonCode string) error {
	return incidentwstest.AwaitSessionRevoked(client, wantReasonCode)
}

func ExpectSessionRevoked(t testing.TB, client *SessionSocketClient, wantReasonCode string) {
	t.Helper()

	incidentwstest.ExpectSessionRevoked(t, client, wantReasonCode)
}

func RequireBootstrapWebsocketRejected(t testing.TB, serverURL string, incidentID string, bootstrapToken string) {
	t.Helper()

	incidentwstest.RequireBootstrapTokenRejected(t, serverURL, incidentID, bootstrapToken)
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

func QueryLeastRecentlyUsedSessionRow(t testing.TB, db *sql.DB, userID string) SessionRow {
	t.Helper()
	return QuerySingleSession(
		t,
		db,
		`SELECT authenticated_at, id::text, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE user_id = $1 ORDER BY last_qualifying_activity_at ASC, authenticated_at ASC, id ASC LIMIT 1`,
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

func QueryUserIDByEmail(t testing.TB, db *sql.DB, email string) string {
	t.Helper()

	var userID string
	if err := db.QueryRowContext(context.Background(), `SELECT id::text FROM users WHERE email = $1`, email).Scan(&userID); err != nil {
		t.Fatalf("query user id by email: %v", err)
	}
	return userID
}

func FormatUserSessions(t testing.TB, db *sql.DB, userID string) string {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT authenticated_at,
       id::text,
       last_qualifying_activity_at,
       revoked_at,
       revoke_reason_code
 FROM user_sessions
 WHERE user_id::text = $1
 ORDER BY last_qualifying_activity_at ASC, authenticated_at ASC, id ASC
`, userID)
	if err != nil {
		t.Fatalf("query user sessions: %v", err)
	}
	defer rows.Close()

	summary := make([]string, 0, 6)
	for rows.Next() {
		var (
			authenticatedAt time.Time
			sessionID       string
			lastActivityAt  time.Time
			revokedAt       sql.NullTime
			reasonCode      sql.NullString
		)
		if err := rows.Scan(&authenticatedAt, &sessionID, &lastActivityAt, &revokedAt, &reasonCode); err != nil {
			t.Fatalf("scan user session: %v", err)
		}
		summary = append(summary, fmt.Sprintf(
			"{id=%s auth=%s last=%s revoked=%t reason=%q}",
			sessionID,
			authenticatedAt.UTC().Format(time.RFC3339Nano),
			lastActivityAt.UTC().Format(time.RFC3339Nano),
			revokedAt.Valid,
			reasonCode.String,
		))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate user sessions: %v", err)
	}
	return strings.Join(summary, ", ")
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
