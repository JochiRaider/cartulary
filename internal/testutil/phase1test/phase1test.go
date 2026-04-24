package phase1test

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
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

const sessionSocketWaitTimeout = 5 * time.Second

type SessionSocketClient struct {
	raw    *wstest.Client
	events chan sessionSocketEvent
}

type sessionSocketEvent struct {
	message *platformws.Message
	err     error
}

type unexpectedSessionSocketMessageError struct {
	messageType string
}

func (e unexpectedSessionSocketMessageError) Error() string {
	return fmt.Sprintf("unexpected websocket message before close: got %q", e.messageType)
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
		options...,
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

func ConnectSessionSocket(t testing.TB, serverURL string, sessionToken string) *SessionSocketClient {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	rawClient := wstest.ConnectWithHeaders(t, serverURL, "/ws/v1/test/session-lifecycle", headers)

	return newSessionSocketClient(t, rawClient)
}

func newSessionSocketClient(t testing.TB, rawClient *wstest.Client) *SessionSocketClient {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), sessionSocketWaitTimeout)
	defer cancel()

	connected, err := rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("read connected websocket message: %v", err)
	}
	wstest.RequireMessageType(t, connected, "connected")

	client := &SessionSocketClient{
		raw:    rawClient,
		events: make(chan sessionSocketEvent, 8),
	}
	go client.readLoop()
	return client
}

func (c *SessionSocketClient) readLoop() {
	defer close(c.events)

	for {
		message, err := c.raw.Receive(context.Background())
		if err != nil {
			c.events <- sessionSocketEvent{err: err}
			return
		}

		current := message
		c.events <- sessionSocketEvent{message: &current}
	}
}

func (c *SessionSocketClient) awaitEvent(timeout time.Duration) (sessionSocketEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case event, ok := <-c.events:
		if !ok {
			return sessionSocketEvent{}, fmt.Errorf("session socket event stream closed")
		}
		return event, nil
	case <-ctx.Done():
		return sessionSocketEvent{}, ctx.Err()
	}
}

func (c *SessionSocketClient) AwaitNextMessage(timeout time.Duration) (platformws.Message, error) {
	event, err := c.awaitEvent(timeout)
	if err != nil {
		return platformws.Message{}, err
	}
	if event.err != nil {
		return platformws.Message{}, event.err
	}
	return *event.message, nil
}

func (c *SessionSocketClient) AwaitClose(timeout time.Duration) error {
	event, err := c.awaitEvent(timeout)
	if err != nil {
		return err
	}
	if event.err == nil {
		return unexpectedSessionSocketMessageError{messageType: event.message.Type}
	}
	return event.err
}

func (c *SessionSocketClient) Close(code websocket.StatusCode, reason string) {
	if c == nil || c.raw == nil {
		return
	}
	c.raw.Close(code, reason)
}

func AwaitSessionRevoked(client *SessionSocketClient, wantReasonCode string) error {
	revoked, err := client.AwaitNextMessage(sessionSocketWaitTimeout)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("timed out waiting for session_revoked message")
		case websocket.CloseStatus(err) >= 0:
			return fmt.Errorf("websocket closed before session_revoked message: %w", err)
		default:
			return fmt.Errorf("read session_revoked message: %w", err)
		}
	}
	if revoked.Type != "session_revoked" {
		return fmt.Errorf("unexpected websocket message before close: got %q want %q", revoked.Type, "session_revoked")
	}

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		return fmt.Errorf("decode session_revoked payload: %w", err)
	}
	if payload["reason_code"] != wantReasonCode {
		return fmt.Errorf("unexpected session_revoked payload reason_code: got %v want %q", payload["reason_code"], wantReasonCode)
	}

	closeErr := client.AwaitClose(sessionSocketWaitTimeout)
	switch {
	case closeErr == nil:
		return nil
	case errors.Is(closeErr, context.DeadlineExceeded):
		return fmt.Errorf("timed out waiting for websocket close after session_revoked")
	default:
		var unexpectedMessageErr unexpectedSessionSocketMessageError
		if errors.As(closeErr, &unexpectedMessageErr) {
			return closeErr
		}

		closeStatus := websocket.CloseStatus(closeErr)
		if closeStatus != websocket.StatusPolicyViolation {
			return fmt.Errorf("unexpected websocket close status: got %d want %d: %w", closeStatus, websocket.StatusPolicyViolation, closeErr)
		}
		if !strings.Contains(closeErr.Error(), "session_revoked") {
			return fmt.Errorf("unexpected websocket close error: %v", closeErr)
		}
		return nil
	}
}

func ExpectSessionRevoked(t testing.TB, client *SessionSocketClient, wantReasonCode string) {
	t.Helper()

	if err := AwaitSessionRevoked(client, wantReasonCode); err != nil {
		t.Fatal(err)
	}
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
