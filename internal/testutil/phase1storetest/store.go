package phase1storetest

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

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
