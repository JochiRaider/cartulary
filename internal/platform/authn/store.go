package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("authn: not found")
var ErrClientTxnConflict = errors.New("authn: client transaction conflict")
var ErrSubjectMismatch = errors.New("authn: subject mismatch")
var ErrPendingExpired = errors.New("authn: pending enrollment expired")
var ErrPendingConsumed = errors.New("authn: pending enrollment consumed")
var ErrLastDeploymentAdmin = errors.New("authn: last deployment admin")
var ErrUserVersionConflict = errors.New("authn: user version conflict")

type Store struct {
	pool *pgxpool.Pool
}

type UserRecord struct {
	ID                   uuid.UUID
	Email                string
	DisplayName          string
	PasswordHash         string
	PasswordChangedAt    time.Time
	MFARequired          bool
	IsActive             bool
	IsDeploymentAdmin    bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	UpdatedByUserID      *uuid.UUID
	LastLoginAt          *time.Time
	UserVersion          int64
	TOTPEnrolledAt       *time.Time
	TOTPSecretCiphertext []byte
	TOTPSecretNonce      []byte
}

type SessionRecord struct {
	ID                       uuid.UUID
	UserID                   uuid.UUID
	AuthenticatedAt          time.Time
	LastQualifyingActivityAt time.Time
	IdleExpiresAt            time.Time
	AbsoluteExpiresAt        time.Time
	SessionExpiresAt         time.Time
	RevokedAt                *time.Time
	RevokeReasonCode         *string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type BootstrapTokenRecord struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	IssuedAt     time.Time
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	SupersededAt *time.Time
}

type PendingTOTPEnrollmentRecord struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	AuthScopeKind             string
	AuthScopeSessionID        *uuid.UUID
	AuthScopeBootstrapTokenID *uuid.UUID
	ClientTxnID               string
	SecretCiphertext          []byte
	SecretNonce               []byte
	ReplacesActive            bool
	CreatedAt                 time.Time
	ExpiresAt                 time.Time
	ConsumedAt                *time.Time
}

type RouteIdempotencyRecord struct {
	RouteKey     string
	ScopeKey     string
	ClientTxnID  string
	RequestHash  []byte
	StatusCode   int
	ResponseJSON []byte
}

type IncidentMembershipSummary struct {
	IncidentID uuid.UUID
	Role       string
}

type PasswordChangeResult struct {
	PasswordChangedAt time.Time
	RevokedSessionIDs []uuid.UUID
	ResponseJSON      []byte
	Replayed          bool
}

type TOTPCompleteResult struct {
	EnrolledAt        time.Time
	SessionsRevoked   bool
	RevokedSessionIDs []uuid.UUID
}

type UserCreateResult struct {
	User         UserRecord
	ResponseJSON []byte
	Replayed     bool
	StatusCode   int
}

type AdminPasswordResetResult struct {
	User              UserRecord
	RevokedSessionIDs []uuid.UUID
	ResponseJSON      []byte
	Replayed          bool
}

type AdminTOTPResetResult struct {
	User              UserRecord
	RevokedSessionIDs []uuid.UUID
	ResponseJSON      []byte
	Replayed          bool
}

type AdminRevokeAllResult struct {
	RevokedAt         time.Time
	RevokedSessionIDs []uuid.UUID
	ResponseJSON      []byte
	Replayed          bool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetUserByNormalizedEmail(ctx context.Context, email string) (UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
       created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
  FROM users
 WHERE email = $1
`, email)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return user, err
}

func (s *Store) GetUserByID(ctx context.Context, userID uuid.UUID) (UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
       created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
  FROM users
 WHERE id = $1
`, userID)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return user, err
}

func (s *Store) ListIncidentMembershipSummaries(ctx context.Context, userID uuid.UUID) ([]IncidentMembershipSummary, error) {
	rows, err := s.pool.Query(ctx, `
SELECT incident_id, role
  FROM incident_memberships
 WHERE user_id = $1
 ORDER BY incident_id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list incident membership summaries: %w", err)
	}
	defer rows.Close()

	summaries := make([]IncidentMembershipSummary, 0)
	for rows.Next() {
		var summary IncidentMembershipSummary
		if err := rows.Scan(&summary.IncidentID, &summary.Role); err != nil {
			return nil, fmt.Errorf("scan incident membership summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate incident membership summaries: %w", err)
	}
	return summaries, nil
}

func (s *Store) GetSessionByFingerprint(ctx context.Context, fingerprint []byte) (SessionRecord, UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT s.id, s.user_id, s.authenticated_at, s.last_qualifying_activity_at, s.idle_expires_at, s.absolute_expires_at,
       s.session_expires_at, s.revoked_at, s.revoke_reason_code, s.created_at, s.updated_at,
       u.id, u.email::text, u.display_name, u.password_hash, u.password_changed_at, u.mfa_required, u.is_active,
       u.is_deployment_admin, u.created_at, u.updated_at, u.updated_by_user_id, u.last_login_at, u.user_version,
       u.totp_enrolled_at, u.totp_secret_ciphertext, u.totp_secret_nonce
  FROM user_sessions s
  JOIN users u ON u.id = s.user_id
 WHERE s.token_fingerprint = $1
`, fingerprint)

	var session SessionRecord
	var user UserRecord
	err := row.Scan(
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
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.PasswordChangedAt,
		&user.MFARequired,
		&user.IsActive,
		&user.IsDeploymentAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UpdatedByUserID,
		&user.LastLoginAt,
		&user.UserVersion,
		&user.TOTPEnrolledAt,
		&user.TOTPSecretCiphertext,
		&user.TOTPSecretNonce,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionRecord{}, UserRecord{}, ErrNotFound
	}
	return session, user, err
}

func (s *Store) CreateSessionWithConcurrency(ctx context.Context, user UserRecord, fingerprint []byte, timing SessionTiming, requestID string) (SessionRecord, *SessionRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SessionRecord{}, nil, fmt.Errorf("begin session transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rows, err := tx.Query(ctx, `
SELECT id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
       session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at
  FROM user_sessions
 WHERE user_id = $1
   AND revoked_at IS NULL
   AND session_expires_at > $2
 ORDER BY last_qualifying_activity_at ASC, authenticated_at ASC, id ASC
 FOR UPDATE
`, user.ID, timing.AuthenticatedAt)
	if err != nil {
		return SessionRecord{}, nil, fmt.Errorf("query active sessions: %w", err)
	}
	defer rows.Close()

	active := make([]SessionRecord, 0, 6)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return SessionRecord{}, nil, err
		}
		active = append(active, session)
	}
	if err := rows.Err(); err != nil {
		return SessionRecord{}, nil, fmt.Errorf("iterate active sessions: %w", err)
	}

	var revoked *SessionRecord
	if len(active) >= 5 {
		session := active[0]
		revokedAt := timing.AuthenticatedAt
		if _, err := tx.Exec(ctx, `
UPDATE user_sessions
   SET revoked_at = $2,
       revoke_reason_code = $3,
       updated_at = $2
 WHERE id = $1
`, session.ID, revokedAt, ConcurrencyLimitReasonCode); err != nil {
			return SessionRecord{}, nil, fmt.Errorf("revoke concurrency victim: %w", err)
		}
		session.RevokedAt = &revokedAt
		reason := ConcurrencyLimitReasonCode
		session.RevokeReasonCode = &reason
		revoked = &session

		if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, request_id, before_json, after_json)
VALUES ($1, $2, 'auth.login', 'session_revoked', $3::text, $4::text, jsonb_build_object('session_id', $5::text), jsonb_build_object('reason_code', $3::text))
`, user.ID, user.ID, ConcurrencyLimitReasonCode, requestID, session.ID); err != nil {
			return SessionRecord{}, nil, fmt.Errorf("insert concurrency audit event: %w", err)
		}
	}

	var created SessionRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO user_sessions (
    user_id, token_fingerprint, authenticated_at, last_qualifying_activity_at,
    idle_expires_at, absolute_expires_at, session_expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
          session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at
`,
		user.ID,
		fingerprint,
		timing.AuthenticatedAt,
		timing.LastQualifyingActivityAt,
		timing.IdleExpiresAt,
		timing.AbsoluteExpiresAt,
		timing.SessionExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.AuthenticatedAt,
		&created.LastQualifyingActivityAt,
		&created.IdleExpiresAt,
		&created.AbsoluteExpiresAt,
		&created.SessionExpiresAt,
		&created.RevokedAt,
		&created.RevokeReasonCode,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return SessionRecord{}, nil, fmt.Errorf("insert session: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`, user.ID, timing.AuthenticatedAt); err != nil {
		return SessionRecord{}, nil, fmt.Errorf("update user last_login_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return SessionRecord{}, nil, fmt.Errorf("commit session transaction: %w", err)
	}
	return created, revoked, nil
}

func (s *Store) SlideSession(ctx context.Context, sessionID uuid.UUID, timing SessionTiming) error {
	_, err := s.pool.Exec(ctx, `
UPDATE user_sessions
   SET last_qualifying_activity_at = $2,
       idle_expires_at = $3,
       session_expires_at = $4,
       updated_at = $2
 WHERE id = $1
`, sessionID, timing.LastQualifyingActivityAt, timing.IdleExpiresAt, timing.SessionExpiresAt)
	if err != nil {
		return fmt.Errorf("slide session expiry: %w", err)
	}
	return nil
}

func (s *Store) ExpireSessionForTest(ctx context.Context, sessionID uuid.UUID, now time.Time) error {
	expiredAt := now.UTC().Add(-time.Second)
	_, err := s.pool.Exec(ctx, `
UPDATE user_sessions
   SET last_qualifying_activity_at = $2,
       idle_expires_at = $2,
       session_expires_at = $2,
       updated_at = $3
 WHERE id = $1
`, sessionID, expiredAt, now.UTC())
	if err != nil {
		return fmt.Errorf("expire session for test: %w", err)
	}
	return nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionID uuid.UUID, reasonCode string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, $3),
       updated_at = $2
 WHERE id = $1
`, sessionID, now.UTC(), reasonCode)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *Store) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reasonCode string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, $3),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
`, userID, now.UTC(), reasonCode)
	if err != nil {
		return fmt.Errorf("revoke all user sessions: %w", err)
	}
	return nil
}

func (s *Store) IssueBootstrapToken(ctx context.Context, userID uuid.UUID, fingerprint []byte, now time.Time) (BootstrapTokenRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BootstrapTokenRecord{}, fmt.Errorf("begin bootstrap token transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `
UPDATE bootstrap_tokens
   SET superseded_at = $2
 WHERE user_id = $1
   AND consumed_at IS NULL
   AND superseded_at IS NULL
   AND expires_at > $2
`, userID, now.UTC()); err != nil {
		return BootstrapTokenRecord{}, fmt.Errorf("supersede bootstrap tokens: %w", err)
	}

	var token BootstrapTokenRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO bootstrap_tokens (user_id, token_fingerprint, issued_at, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, issued_at, expires_at, consumed_at, superseded_at
`, userID, fingerprint, now.UTC(), now.UTC().Add(BootstrapTokenTTL)).Scan(
		&token.ID,
		&token.UserID,
		&token.IssuedAt,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.SupersededAt,
	); err != nil {
		return BootstrapTokenRecord{}, fmt.Errorf("insert bootstrap token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return BootstrapTokenRecord{}, fmt.Errorf("commit bootstrap token transaction: %w", err)
	}
	return token, nil
}

func (s *Store) GetBootstrapTokenByFingerprint(ctx context.Context, fingerprint []byte) (BootstrapTokenRecord, UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT b.id, b.user_id, b.issued_at, b.expires_at, b.consumed_at, b.superseded_at,
       u.id, u.email::text, u.display_name, u.password_hash, u.password_changed_at, u.mfa_required, u.is_active,
       u.is_deployment_admin, u.created_at, u.updated_at, u.updated_by_user_id, u.last_login_at, u.user_version,
       u.totp_enrolled_at, u.totp_secret_ciphertext, u.totp_secret_nonce
  FROM bootstrap_tokens b
  JOIN users u ON u.id = b.user_id
 WHERE b.token_fingerprint = $1
`, fingerprint)

	var token BootstrapTokenRecord
	var user UserRecord
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.IssuedAt,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.SupersededAt,
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.PasswordChangedAt,
		&user.MFARequired,
		&user.IsActive,
		&user.IsDeploymentAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UpdatedByUserID,
		&user.LastLoginAt,
		&user.UserVersion,
		&user.TOTPEnrolledAt,
		&user.TOTPSecretCiphertext,
		&user.TOTPSecretNonce,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return BootstrapTokenRecord{}, UserRecord{}, ErrNotFound
	}
	return token, user, err
}

func (s *Store) GetPendingTOTPEnrollmentForUser(ctx context.Context, userID uuid.UUID, now time.Time) (*PendingTOTPEnrollmentRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
       secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
  FROM pending_totp_enrollments
 WHERE user_id = $1
   AND consumed_at IS NULL
   AND expires_at > $2
 ORDER BY created_at DESC, id DESC
 LIMIT 1
`, userID, now.UTC())
	record, err := scanPendingTOTPEnrollment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *Store) GetPendingTOTPEnrollmentByID(ctx context.Context, enrollmentID uuid.UUID) (*PendingTOTPEnrollmentRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
       secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
  FROM pending_totp_enrollments
 WHERE id = $1
`, enrollmentID)
	record, err := scanPendingTOTPEnrollment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *Store) BeginTOTPEnrollment(
	ctx context.Context,
	userID uuid.UUID,
	authScopeKind string,
	sessionID *uuid.UUID,
	bootstrapTokenID *uuid.UUID,
	clientTxnID string,
	secretCiphertext []byte,
	secretNonce []byte,
	replacesActive bool,
	now time.Time,
) (PendingTOTPEnrollmentRecord, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("begin totp enrollment transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
       secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
  FROM pending_totp_enrollments
 WHERE user_id = $1
   AND consumed_at IS NULL
 ORDER BY created_at DESC, id DESC
 LIMIT 1
 FOR UPDATE
`, userID)
	current, err := scanPendingTOTPEnrollment(row)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("query current pending totp enrollment: %w", err)
	default:
		if !current.ExpiresAt.After(now.UTC()) {
			if _, err := tx.Exec(ctx, `DELETE FROM pending_totp_enrollments WHERE id = $1`, current.ID); err != nil {
				return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("clear expired pending totp enrollment: %w", err)
			}
		} else if current.AuthScopeKind == authScopeKind &&
			current.ClientTxnID == clientTxnID &&
			uuidPointersEqual(current.AuthScopeSessionID, sessionID) &&
			uuidPointersEqual(current.AuthScopeBootstrapTokenID, bootstrapTokenID) {
			if err := tx.Commit(ctx); err != nil {
				return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("commit replayed totp enrollment transaction: %w", err)
			}
			return current, true, nil
		} else {
			return PendingTOTPEnrollmentRecord{}, false, ErrClientTxnConflict
		}
	}

	var created PendingTOTPEnrollmentRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO pending_totp_enrollments (
    user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
    secret_ciphertext, secret_nonce, replaces_active, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
          secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
`, userID, authScopeKind, sessionID, bootstrapTokenID, clientTxnID, secretCiphertext, secretNonce, replacesActive, now.UTC().Add(PendingTOTPEnrollmentTTL)).Scan(
		&created.ID,
		&created.UserID,
		&created.AuthScopeKind,
		&created.AuthScopeSessionID,
		&created.AuthScopeBootstrapTokenID,
		&created.ClientTxnID,
		&created.SecretCiphertext,
		&created.SecretNonce,
		&created.ReplacesActive,
		&created.CreatedAt,
		&created.ExpiresAt,
		&created.ConsumedAt,
	); err != nil {
		if IsUniqueViolation(err) {
			return PendingTOTPEnrollmentRecord{}, false, ErrClientTxnConflict
		}
		return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("insert pending totp enrollment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("commit totp enrollment transaction: %w", err)
	}
	return created, false, nil
}

func (s *Store) ActivateTOTPEnrollment(
	ctx context.Context,
	user UserRecord,
	enrollmentID uuid.UUID,
	authScopeKind string,
	sessionID *uuid.UUID,
	bootstrapTokenID *uuid.UUID,
	now time.Time,
) (TOTPCompleteResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("begin totp complete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT id, user_id, auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id,
       secret_ciphertext, secret_nonce, replaces_active, created_at, expires_at, consumed_at
  FROM pending_totp_enrollments
 WHERE id = $1
 FOR UPDATE
`, enrollmentID)
	pending, err := scanPendingTOTPEnrollment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return TOTPCompleteResult{}, ErrNotFound
	}
	if err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("query pending totp enrollment: %w", err)
	}
	if pending.UserID != user.ID {
		return TOTPCompleteResult{}, ErrSubjectMismatch
	}
	if status := PendingEnrollmentStatusAt(&pending, now); status == PendingEnrollmentExpired {
		return TOTPCompleteResult{}, ErrPendingExpired
	} else if status == PendingEnrollmentConsumed {
		return TOTPCompleteResult{}, ErrPendingConsumed
	}
	if pending.AuthScopeKind != authScopeKind ||
		!uuidPointersEqual(pending.AuthScopeSessionID, sessionID) ||
		!uuidPointersEqual(pending.AuthScopeBootstrapTokenID, bootstrapTokenID) {
		return TOTPCompleteResult{}, ErrSubjectMismatch
	}

	enrolledAt := now.UTC()
	if _, err := tx.Exec(ctx, `
UPDATE users
   SET totp_enrolled_at = $2,
       totp_secret_ciphertext = $3,
       totp_secret_nonce = $4,
       updated_at = $2,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $1
`, user.ID, enrolledAt, pending.SecretCiphertext, pending.SecretNonce); err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("activate totp secret: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE pending_totp_enrollments
   SET consumed_at = $2
 WHERE id = $1
`, pending.ID, enrolledAt); err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("consume pending totp enrollment: %w", err)
	}

	if bootstrapTokenID != nil {
		if _, err := tx.Exec(ctx, `
UPDATE bootstrap_tokens
   SET consumed_at = $2
 WHERE id = $1
`, *bootstrapTokenID, enrolledAt); err != nil {
			return TOTPCompleteResult{}, fmt.Errorf("consume bootstrap token: %w", err)
		}
	}

	result := TOTPCompleteResult{
		EnrolledAt:      enrolledAt,
		SessionsRevoked: pending.ReplacesActive,
	}
	if pending.ReplacesActive {
		rows, err := tx.Query(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, 'session_revoked'),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
RETURNING id
`, user.ID, enrolledAt)
		if err != nil {
			return TOTPCompleteResult{}, fmt.Errorf("revoke sessions for totp replacement: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var sessionID uuid.UUID
			if err := rows.Scan(&sessionID); err != nil {
				return TOTPCompleteResult{}, fmt.Errorf("scan revoked replacement session: %w", err)
			}
			result.RevokedSessionIDs = append(result.RevokedSessionIDs, sessionID)
		}
		if err := rows.Err(); err != nil {
			return TOTPCompleteResult{}, fmt.Errorf("iterate revoked replacement sessions: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("commit totp complete transaction: %w", err)
	}
	return result, nil
}

func (s *Store) GetRouteIdempotency(ctx context.Context, routeKey string, scopeKey string, clientTxnID string) (RouteIdempotencyRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, routeKey, scopeKey, clientTxnID)
	var record RouteIdempotencyRecord
	if err := row.Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RouteIdempotencyRecord{}, ErrNotFound
		}
		return RouteIdempotencyRecord{}, err
	}
	return record, nil
}

func (s *Store) ChangePassword(
	ctx context.Context,
	user UserRecord,
	clientTxnID string,
	requestHash []byte,
	newPasswordHash string,
	requestID string,
	now time.Time,
) (PasswordChangeResult, error) {
	scopeKey := user.ID.String()
	if existing, err := s.GetRouteIdempotency(ctx, "auth.password.change", scopeKey, clientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return PasswordChangeResult{}, ErrClientTxnConflict
		}
		return PasswordChangeResult{
			ResponseJSON: existing.ResponseJSON,
			Replayed:     true,
		}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return PasswordChangeResult{}, fmt.Errorf("query password-change idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PasswordChangeResult{}, fmt.Errorf("begin password-change transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	changedAt := now.UTC()
	var effectiveChangedAt time.Time
	if err := tx.QueryRow(ctx, `
UPDATE users
   SET password_hash = $2,
       password_changed_at = $3,
       updated_at = $3,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $1
RETURNING password_changed_at
`, user.ID, newPasswordHash, changedAt).Scan(&effectiveChangedAt); err != nil {
		return PasswordChangeResult{}, fmt.Errorf("update password hash: %w", err)
	}

	rows, err := tx.Query(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, 'session_revoked'),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
RETURNING id
`, user.ID, effectiveChangedAt)
	if err != nil {
		return PasswordChangeResult{}, fmt.Errorf("revoke sessions for password change: %w", err)
	}
	defer rows.Close()

	result := PasswordChangeResult{
		PasswordChangedAt: effectiveChangedAt,
	}
	for rows.Next() {
		var sessionID uuid.UUID
		if err := rows.Scan(&sessionID); err != nil {
			return PasswordChangeResult{}, fmt.Errorf("scan revoked password-change session: %w", err)
		}
		result.RevokedSessionIDs = append(result.RevokedSessionIDs, sessionID)
	}
	if err := rows.Err(); err != nil {
		return PasswordChangeResult{}, fmt.Errorf("iterate revoked password-change sessions: %w", err)
	}

	responseJSON, err := json.Marshal(map[string]any{
		"user_id":          user.ID,
		"password":         map[string]any{"changed_at": effectiveChangedAt},
		"sessions_revoked": true,
	})
	if err != nil {
		return PasswordChangeResult{}, fmt.Errorf("marshal password-change idempotency response: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "auth.password.change", scopeKey, clientTxnID, user.ID, user.ID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return PasswordChangeResult{}, ErrClientTxnConflict
		}
		return PasswordChangeResult{}, fmt.Errorf("insert password-change idempotency: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return PasswordChangeResult{}, fmt.Errorf("commit password-change transaction: %w", err)
	}
	result.ResponseJSON = responseJSON
	return result, nil
}

func (s *Store) CountActiveDeploymentAdmins(ctx context.Context) (int, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
SELECT COUNT(*)
  FROM users
 WHERE is_deployment_admin = true
   AND is_active = true
`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) ListUsers(ctx context.Context, limit int, cursor *uuid.UUID) ([]UserRecord, *string, error) {
	query := `
SELECT id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
       created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
  FROM users
`
	args := []any{}
	if cursor != nil {
		query += " WHERE id > $1"
		args = append(args, *cursor)
	}
	orderArg := len(args) + 1
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", orderArg)
	args = append(args, limit+1)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	users := make([]UserRecord, 0, limit+1)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var nextCursor *string
	if len(users) > limit {
		cursorValue := users[limit-1].ID.String()
		nextCursor = &cursorValue
		users = users[:limit]
	}
	return users, nextCursor, nil
}

func (s *Store) CreateUser(
	ctx context.Context,
	actor UserRecord,
	email string,
	displayName string,
	passwordHash string,
	mfaRequired bool,
	isDeploymentAdmin bool,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (UserCreateResult, error) {
	scopeKey := actor.ID.String()
	if existing, err := s.GetRouteIdempotency(ctx, "users.create", scopeKey, clientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return UserCreateResult{}, ErrClientTxnConflict
		}
		return UserCreateResult{
			ResponseJSON: existing.ResponseJSON,
			Replayed:     true,
			StatusCode:   http.StatusOK,
		}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return UserCreateResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UserCreateResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var created UserRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO users (
    email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, updated_by_user_id, updated_at
)
VALUES ($1, $2, $3, $4, true, $5, $6, $7)
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, email, displayName, passwordHash, mfaRequired, isDeploymentAdmin, actor.ID, now.UTC()).Scan(
		&created.ID,
		&created.Email,
		&created.DisplayName,
		&created.PasswordHash,
		&created.PasswordChangedAt,
		&created.MFARequired,
		&created.IsActive,
		&created.IsDeploymentAdmin,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.UpdatedByUserID,
		&created.LastLoginAt,
		&created.UserVersion,
		&created.TOTPEnrolledAt,
		&created.TOTPSecretCiphertext,
		&created.TOTPSecretNonce,
	); err != nil {
		return UserCreateResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, after_json)
VALUES ($1, $2, 'users.create', 'user_created', 'user_created', $3, $4, jsonb_build_object('user_id', $5::text))
`, actor.ID, created.ID, clientTxnID, requestID, created.ID.String()); err != nil {
		return UserCreateResult{}, err
	}

	responseJSON, err := json.Marshal(safeUserResponse(created))
	if err != nil {
		return UserCreateResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "users.create", scopeKey, clientTxnID, actor.ID, created.ID, requestHash, http.StatusCreated, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return UserCreateResult{}, ErrClientTxnConflict
		}
		return UserCreateResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return UserCreateResult{}, err
	}
	return UserCreateResult{
		User:         created,
		ResponseJSON: responseJSON,
		StatusCode:   http.StatusCreated,
	}, nil
}

func (s *Store) UpdateUser(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	baseUserVersion int64,
	email *string,
	displayName *string,
	isActive *bool,
	mfaRequired *bool,
	isDeploymentAdmin *bool,
	requestID string,
	now time.Time,
) (UserRecord, []uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UserRecord{}, nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return UserRecord{}, nil, err
	}
	if target.UserVersion != baseUserVersion {
		return UserRecord{}, nil, ErrUserVersionConflict
	}

	nextEmail := target.Email
	if email != nil {
		nextEmail = *email
	}
	nextDisplayName := target.DisplayName
	if displayName != nil {
		nextDisplayName = *displayName
	}
	nextIsActive := target.IsActive
	if isActive != nil {
		nextIsActive = *isActive
	}
	nextMFARequired := target.MFARequired
	if mfaRequired != nil {
		nextMFARequired = *mfaRequired
	}
	nextIsDeploymentAdmin := target.IsDeploymentAdmin
	if isDeploymentAdmin != nil {
		nextIsDeploymentAdmin = *isDeploymentAdmin
	}

	if target.IsDeploymentAdmin && target.IsActive && !(nextIsDeploymentAdmin && nextIsActive) && countActiveAdminsTx(ctx, tx) <= 1 {
		return UserRecord{}, nil, ErrLastDeploymentAdmin
	}

	var updated UserRecord
	if err := tx.QueryRow(ctx, `
UPDATE users
   SET email = $2,
       display_name = $3,
       is_active = $4,
       mfa_required = $5,
       is_deployment_admin = $6,
       updated_at = $7,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $8
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, actor.ID, nextEmail, nextDisplayName, nextIsActive, nextMFARequired, nextIsDeploymentAdmin, now.UTC(), targetUserID).Scan(
		&updated.ID,
		&updated.Email,
		&updated.DisplayName,
		&updated.PasswordHash,
		&updated.PasswordChangedAt,
		&updated.MFARequired,
		&updated.IsActive,
		&updated.IsDeploymentAdmin,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.LastLoginAt,
		&updated.UserVersion,
		&updated.TOTPEnrolledAt,
		&updated.TOTPSecretCiphertext,
		&updated.TOTPSecretNonce,
	); err != nil {
		return UserRecord{}, nil, err
	}

	var revoked []uuid.UUID
	if target.IsActive && !nextIsActive {
		rows, err := tx.Query(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, 'session_revoked'),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
RETURNING id
`, targetUserID, now.UTC())
		if err != nil {
			return UserRecord{}, nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var sessionID uuid.UUID
			if err := rows.Scan(&sessionID); err != nil {
				return UserRecord{}, nil, err
			}
			revoked = append(revoked, sessionID)
		}
		if err := rows.Err(); err != nil {
			return UserRecord{}, nil, err
		}
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, request_id, before_json, after_json)
VALUES ($1, $2, 'users.patch', 'user_updated', 'user_updated', $3, jsonb_build_object('user_version', $4::bigint), jsonb_build_object('user_version', $5::bigint))
`, actor.ID, updated.ID, requestID, target.UserVersion, updated.UserVersion); err != nil {
		return UserRecord{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return UserRecord{}, nil, err
	}
	return updated, revoked, nil
}

func (s *Store) AdminResetPassword(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	baseUserVersion int64,
	newPasswordHash string,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (AdminPasswordResetResult, error) {
	scopeKey := actor.ID.String() + ":" + targetUserID.String()
	if existing, err := s.GetRouteIdempotency(ctx, "users.password.reset", scopeKey, clientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AdminPasswordResetResult{}, ErrClientTxnConflict
		}
		return AdminPasswordResetResult{
			ResponseJSON: existing.ResponseJSON,
			Replayed:     true,
		}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return AdminPasswordResetResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AdminPasswordResetResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return AdminPasswordResetResult{}, err
	}
	if target.UserVersion != baseUserVersion {
		return AdminPasswordResetResult{}, ErrUserVersionConflict
	}

	changedAt := now.UTC()
	var updated UserRecord
	if err := tx.QueryRow(ctx, `
UPDATE users
   SET password_hash = $2,
       password_changed_at = $3,
       updated_at = $3,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $4
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, actor.ID, newPasswordHash, changedAt, targetUserID).Scan(
		&updated.ID,
		&updated.Email,
		&updated.DisplayName,
		&updated.PasswordHash,
		&updated.PasswordChangedAt,
		&updated.MFARequired,
		&updated.IsActive,
		&updated.IsDeploymentAdmin,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.LastLoginAt,
		&updated.UserVersion,
		&updated.TOTPEnrolledAt,
		&updated.TOTPSecretCiphertext,
		&updated.TOTPSecretNonce,
	); err != nil {
		return AdminPasswordResetResult{}, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE bootstrap_tokens
   SET superseded_at = $2
 WHERE user_id = $1
   AND consumed_at IS NULL
   AND superseded_at IS NULL
`, targetUserID, changedAt); err != nil {
		return AdminPasswordResetResult{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM pending_totp_enrollments WHERE user_id = $1`, targetUserID); err != nil {
		return AdminPasswordResetResult{}, err
	}

	result := AdminPasswordResetResult{User: updated}
	rows, err := tx.Query(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, 'session_revoked'),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
RETURNING id
`, targetUserID, changedAt)
	if err != nil {
		return AdminPasswordResetResult{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var sessionID uuid.UUID
		if err := rows.Scan(&sessionID); err != nil {
			return AdminPasswordResetResult{}, err
		}
		result.RevokedSessionIDs = append(result.RevokedSessionIDs, sessionID)
	}
	if err := rows.Err(); err != nil {
		return AdminPasswordResetResult{}, err
	}

	responseJSON, err := json.Marshal(safeUserResponse(updated))
	if err != nil {
		return AdminPasswordResetResult{}, err
	}
	result.ResponseJSON = responseJSON

	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "users.password.reset", scopeKey, clientTxnID, actor.ID, targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminPasswordResetResult{}, ErrClientTxnConflict
		}
		return AdminPasswordResetResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, after_json)
VALUES ($1, $2, 'users.password.reset', 'password_reset', 'password_reset', $3, $4, jsonb_build_object('user_version', $5::bigint))
`, actor.ID, targetUserID, clientTxnID, requestID, updated.UserVersion); err != nil {
		return AdminPasswordResetResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminPasswordResetResult{}, err
	}
	return result, nil
}

func (s *Store) AdminResetTOTP(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	baseUserVersion int64,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (AdminTOTPResetResult, error) {
	scopeKey := actor.ID.String() + ":" + targetUserID.String()
	if existing, err := s.GetRouteIdempotency(ctx, "users.totp.reset", scopeKey, clientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AdminTOTPResetResult{}, ErrClientTxnConflict
		}
		return AdminTOTPResetResult{
			ResponseJSON: existing.ResponseJSON,
			Replayed:     true,
		}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return AdminTOTPResetResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AdminTOTPResetResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return AdminTOTPResetResult{}, err
	}
	if target.UserVersion != baseUserVersion {
		return AdminTOTPResetResult{}, ErrUserVersionConflict
	}

	changedAt := now.UTC()
	if _, err := tx.Exec(ctx, `
DELETE FROM pending_totp_enrollments
 WHERE user_id = $1
`, targetUserID); err != nil {
		return AdminTOTPResetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE bootstrap_tokens
   SET superseded_at = $2
 WHERE user_id = $1
   AND consumed_at IS NULL
   AND superseded_at IS NULL
`, targetUserID, changedAt); err != nil {
		return AdminTOTPResetResult{}, err
	}

	var updated UserRecord
	if err := tx.QueryRow(ctx, `
UPDATE users
   SET totp_enrolled_at = NULL,
       totp_secret_ciphertext = NULL,
       totp_secret_nonce = NULL,
       updated_at = $2,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $3
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, actor.ID, changedAt, targetUserID).Scan(
		&updated.ID,
		&updated.Email,
		&updated.DisplayName,
		&updated.PasswordHash,
		&updated.PasswordChangedAt,
		&updated.MFARequired,
		&updated.IsActive,
		&updated.IsDeploymentAdmin,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.LastLoginAt,
		&updated.UserVersion,
		&updated.TOTPEnrolledAt,
		&updated.TOTPSecretCiphertext,
		&updated.TOTPSecretNonce,
	); err != nil {
		return AdminTOTPResetResult{}, err
	}

	revoked, err := revokeAllSessionsTx(ctx, tx, targetUserID, changedAt)
	if err != nil {
		return AdminTOTPResetResult{}, err
	}

	responseJSON, err := json.Marshal(safeUserResponse(updated))
	if err != nil {
		return AdminTOTPResetResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "users.totp.reset", scopeKey, clientTxnID, actor.ID, targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminTOTPResetResult{}, ErrClientTxnConflict
		}
		return AdminTOTPResetResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, after_json)
VALUES ($1, $2, 'users.totp.reset', 'totp_reset', 'totp_reset', $3, $4, jsonb_build_object('user_version', $5::bigint))
`, actor.ID, targetUserID, clientTxnID, requestID, updated.UserVersion); err != nil {
		return AdminTOTPResetResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminTOTPResetResult{}, err
	}
	return AdminTOTPResetResult{
		User:              updated,
		RevokedSessionIDs: revoked,
		ResponseJSON:      responseJSON,
	}, nil
}

func (s *Store) AdminRevokeAllSessions(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (AdminRevokeAllResult, error) {
	scopeKey := actor.ID.String() + ":" + targetUserID.String()
	if existing, err := s.GetRouteIdempotency(ctx, "users.sessions.revoke_all", scopeKey, clientTxnID); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AdminRevokeAllResult{}, ErrClientTxnConflict
		}
		return AdminRevokeAllResult{
			ResponseJSON: existing.ResponseJSON,
			Replayed:     true,
		}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return AdminRevokeAllResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AdminRevokeAllResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	revokedAt := now.UTC()
	revoked, err := revokeAllSessionsTx(ctx, tx, targetUserID, revokedAt)
	if err != nil {
		return AdminRevokeAllResult{}, err
	}
	responseJSON, err := json.Marshal(map[string]any{
		"user_id":          targetUserID,
		"sessions_revoked": true,
		"revoked_at":       revokedAt,
	})
	if err != nil {
		return AdminRevokeAllResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "users.sessions.revoke_all", scopeKey, clientTxnID, actor.ID, targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminRevokeAllResult{}, ErrClientTxnConflict
		}
		return AdminRevokeAllResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, after_json)
VALUES ($1, $2, 'users.sessions.revoke_all', 'sessions_revoke_all', 'sessions_revoke_all', $3, $4, jsonb_build_object('revoked_at', $5::timestamptz))
`, actor.ID, targetUserID, clientTxnID, requestID, revokedAt); err != nil {
		return AdminRevokeAllResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminRevokeAllResult{}, err
	}
	return AdminRevokeAllResult{
		RevokedAt:         revokedAt,
		RevokedSessionIDs: revoked,
		ResponseJSON:      responseJSON,
	}, nil
}

func scanUser(scanner interface{ Scan(...any) error }) (UserRecord, error) {
	var user UserRecord
	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.PasswordChangedAt,
		&user.MFARequired,
		&user.IsActive,
		&user.IsDeploymentAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UpdatedByUserID,
		&user.LastLoginAt,
		&user.UserVersion,
		&user.TOTPEnrolledAt,
		&user.TOTPSecretCiphertext,
		&user.TOTPSecretNonce,
	)
	return user, err
}

func scanSession(scanner interface{ Scan(...any) error }) (SessionRecord, error) {
	var session SessionRecord
	err := scanner.Scan(
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
	)
	return session, err
}

func scanPendingTOTPEnrollment(scanner interface{ Scan(...any) error }) (PendingTOTPEnrollmentRecord, error) {
	var record PendingTOTPEnrollmentRecord
	err := scanner.Scan(
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
	)
	return record, err
}

func uuidPointersEqual(left *uuid.UUID, right *uuid.UUID) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}

func fetchUserForUpdate(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (UserRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
       created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
  FROM users
 WHERE id = $1
 FOR UPDATE
`, userID)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return user, err
}

func countActiveAdminsTx(ctx context.Context, tx pgx.Tx) int {
	var count int
	_ = tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM users
 WHERE is_deployment_admin = true
   AND is_active = true
`).Scan(&count)
	return count
}

func revokeAllSessionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, revokedAt time.Time) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
UPDATE user_sessions
   SET revoked_at = COALESCE(revoked_at, $2),
       revoke_reason_code = COALESCE(revoke_reason_code, 'session_revoked'),
       updated_at = $2
 WHERE user_id = $1
   AND revoked_at IS NULL
RETURNING id
`, userID, revokedAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revoked []uuid.UUID
	for rows.Next() {
		var sessionID uuid.UUID
		if err := rows.Scan(&sessionID); err != nil {
			return nil, err
		}
		revoked = append(revoked, sessionID)
	}
	return revoked, rows.Err()
}

func safeUserResponse(user UserRecord) map[string]any {
	return map[string]any{
		"user_id":             user.ID,
		"email":               user.Email,
		"display_name":        user.DisplayName,
		"is_active":           user.IsActive,
		"mfa_required":        user.MFARequired,
		"is_deployment_admin": user.IsDeploymentAdmin,
		"created_at":          user.CreatedAt,
		"updated_at":          user.UpdatedAt,
		"updated_by_user_id":  user.UpdatedByUserID,
		"last_login_at":       user.LastLoginAt,
		"user_version":        user.UserVersion,
		"auth_bindings": []map[string]any{
			{
				"provider_type": "local",
				"provider_key":  "local",
				"username":      user.Email,
				"created_at":    user.CreatedAt,
			},
		},
	}
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
