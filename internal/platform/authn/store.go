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

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrNotFound = errors.New("authn: not found")
var ErrClientTxnConflict = errors.New("authn: client transaction conflict")
var ErrSubjectMismatch = errors.New("authn: subject mismatch")
var ErrPendingExpired = errors.New("authn: pending enrollment expired")
var ErrPendingConsumed = errors.New("authn: pending enrollment consumed")
var ErrLastDeploymentAdmin = errors.New("authn: last deployment admin")
var ErrUserVersionConflict = errors.New("authn: user version conflict")
var ErrPreferencesVersionConflict = errors.New("authn: preferences version conflict")
var ErrAuthProviderNotFound = errors.New("authn: auth provider not found")
var ErrAuthProviderDisabled = errors.New("authn: auth provider disabled")
var ErrAuthProviderTypeChangeNotSupported = errors.New("authn: auth provider type change not supported")
var ErrEnterpriseTransactionNotFound = errors.New("authn: enterprise auth transaction not found")
var ErrEnterpriseTransactionExpired = errors.New("authn: enterprise auth transaction expired")
var ErrEnterpriseTransactionUsed = errors.New("authn: enterprise auth transaction already used")
var ErrEnterpriseTransactionProviderMismatch = errors.New("authn: enterprise auth transaction provider mismatch")
var ErrEnterpriseTransactionStateMismatch = errors.New("authn: enterprise auth transaction state mismatch")
var ErrEnterpriseTransactionBrowserMismatch = errors.New("authn: enterprise auth transaction browser mismatch")
var ErrEnterpriseTransactionCompletionMismatch = errors.New("authn: enterprise auth transaction completion mismatch")
var ErrEnterpriseIdentityNoLinkedUser = errors.New("authn: enterprise identity has no linked user")
var ErrEnterpriseIdentityInactiveUser = errors.New("authn: enterprise identity user inactive")
var ErrAuthBindingNotFound = errors.New("authn: auth binding not found")
var ErrAuthBindingNotActive = errors.New("authn: auth binding not active")
var ErrAuthBindingProviderSubjectInUse = errors.New("authn: auth binding provider subject in use")
var ErrAuthBindingProviderAlreadyLinkedForUser = errors.New("authn: auth binding provider already linked for user")

type Store struct {
	pool postgres.DB
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

type UserListFilter struct {
	SearchTokens      []string
	IsActive          *bool
	IsDeploymentAdmin *bool
}

type AccountPreferencesRecord struct {
	UserID             uuid.UUID
	DensityMode        *string
	PreferencesVersion int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SessionRecord struct {
	ID                       uuid.UUID
	UserID                   uuid.UUID
	ProviderType             string
	AuthBindingID            *uuid.UUID
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
	ActorUserID  uuid.UUID
	RequestHash  []byte
	StatusCode   int
	ResponseJSON []byte
}

type RouteIdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
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

type AccountProfilePatchResult struct {
	User         UserRecord
	ResponseJSON []byte
	Replayed     bool
	StatusCode   int
}

type AccountPreferencesPutResult struct {
	Preferences  AccountPreferencesRecord
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

type EnterpriseAuthProviderRecord struct {
	ID                        uuid.UUID
	ProviderKey               string
	ProviderType              string
	DisplayName               string
	IsEnabled                 bool
	IsInteractive             bool
	AuthorizationEndpoint     *string
	Issuer                    *string
	Audience                  *string
	TokenEndpoint             *string
	JWKSURI                   *string
	ClientID                  *string
	ClientSecretRefKind       *string
	ClientSecretRefName       *string
	AdditionalScopes          []string
	SAMLIDPEntityID           *string
	SAMLSSOURL                *string
	SAMLIDPSigningCertificate []string
	SAMLSPHostEntityID        *string
	SAMLSubjectSource         *EnterpriseAuthSAMLSubjectSource
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type EnterpriseAuthSAMLSubjectSource struct {
	Kind          string `json:"kind"`
	AttributeName string `json:"attribute_name,omitempty"`
}

type EnterpriseAuthTransactionRecord struct {
	ID                     uuid.UUID
	ProviderID             uuid.UUID
	ProviderKey            string
	ProviderType           string
	ReturnTo               string
	State                  *string
	Nonce                  *string
	RelayState             *string
	BrowserBindingHash     []byte
	PKCEVerifierCiphertext []byte
	PKCEVerifierNonce      []byte
	SAMLRequestID          *string
	SAMLCompletionHash     []byte
	SAMLSubject            *string
	SAMLStagedAt           *time.Time
	CreatedAt              time.Time
	ExpiresAt              time.Time
	ConsumedAt             *time.Time
}

type EnterpriseAuthProviderDefinition struct {
	ProviderKey               string
	ProviderType              string
	DisplayName               string
	Enabled                   bool
	AuthorizationEndpoint     *string
	Issuer                    *string
	Audience                  *string
	TokenEndpoint             *string
	JWKSURI                   *string
	ClientID                  *string
	ClientSecretRefKind       *string
	ClientSecretRefName       *string
	AdditionalScopes          []string
	SAMLIDPEntityID           *string
	SAMLSSOURL                *string
	SAMLIDPSigningCertificate []string
	SAMLSPHostEntityID        *string
	SAMLSubjectSource         *EnterpriseAuthSAMLSubjectSource
}

type EnterpriseAuthBindingSummary struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProviderKey     string
	ProviderType    string
	ProviderSubject string
	CreatedAt       time.Time
	LastAuthAt      *time.Time
}

type EnterpriseAuthCompletionResult struct {
	User          UserRecord
	Binding       EnterpriseAuthBindingSummary
	ReturnTo      string
	TransactionID uuid.UUID
}

type EnterpriseAuthBindingResult struct {
	User              UserRecord
	ResponseJSON      []byte
	RevokedSessionIDs []uuid.UUID
	Replayed          bool
	StatusCode        int
}

func NewStore(pool postgres.DB) *Store {
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

// GetUserByIDForUpdateTx rederives and locks the authoritative actor row in a
// caller-owned mutation transaction.
func (s *Store) GetUserByIDForUpdateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (UserRecord, error) {
	if s == nil || tx == nil {
		return UserRecord{}, ErrNotFound
	}
	return fetchUserForUpdate(ctx, tx, userID)
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
       s.session_expires_at, s.revoked_at, s.revoke_reason_code, s.created_at, s.updated_at, s.provider_type, s.auth_binding_id,
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
		&session.ProviderType,
		&session.AuthBindingID,
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
	return s.createSessionWithConcurrency(ctx, user, fingerprint, timing, requestID, "local", nil)
}

func (s *Store) CreateSessionWithProviderConcurrency(ctx context.Context, user UserRecord, fingerprint []byte, timing SessionTiming, requestID string, providerType string, authBindingID *uuid.UUID) (SessionRecord, *SessionRecord, error) {
	return s.createSessionWithConcurrency(ctx, user, fingerprint, timing, requestID, providerType, authBindingID)
}

func (s *Store) createSessionWithConcurrency(ctx context.Context, user UserRecord, fingerprint []byte, timing SessionTiming, requestID string, providerType string, authBindingID *uuid.UUID) (SessionRecord, *SessionRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SessionRecord{}, nil, fmt.Errorf("begin session transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rows, err := tx.Query(ctx, `
SELECT id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
       session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at, provider_type, auth_binding_id
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
	activeSummaries := make([]SessionSummary, 0, 6)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return SessionRecord{}, nil, err
		}
		active = append(active, session)
		activeSummaries = append(activeSummaries, SessionSummary{
			SessionID:                session.ID,
			LastQualifyingActivityAt: session.LastQualifyingActivityAt,
			AuthenticatedAt:          session.AuthenticatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return SessionRecord{}, nil, fmt.Errorf("iterate active sessions: %w", err)
	}

	var revoked *SessionRecord
	if len(active) >= 5 {
		victim, ok := SelectSessionForConcurrencyLimit(activeSummaries, uuid.Nil)
		if !ok {
			return SessionRecord{}, nil, fmt.Errorf("select concurrency victim: no eligible active session")
		}
		var session SessionRecord
		for _, activeSession := range active {
			if activeSession.ID == victim.SessionID {
				session = activeSession
				break
			}
		}
		if session.ID == uuid.Nil {
			return SessionRecord{}, nil, fmt.Errorf("select concurrency victim: selected session %s was not loaded", victim.SessionID)
		}
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

		if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
			ActorUserID:  &user.ID,
			TargetUserID: &user.ID,
			RawSource:    "auth.login",
			RawKind:      "session_revoked",
			ReasonCode:   &reason,
			RequestID:    &requestID,
			Before:       map[string]any{"session_id": session.ID},
			After:        map[string]any{"reason_code": ConcurrencyLimitReasonCode},
			OccurredAt:   timing.AuthenticatedAt.UTC(),
			Source:       administrativeaudit.SourceSystem,
			ActionCode:   administrativeaudit.ActionSessionsRevoked,
			TargetKind:   administrativeaudit.TargetUser,
			TargetID:     user.ID.String(),
			Changes: []administrativeaudit.Change{
				administrativeaudit.Visible("reason_code", nil, ConcurrencyLimitReasonCode),
				administrativeaudit.Visible("sessions_revoked", 0, 1),
			},
		}); err != nil {
			return SessionRecord{}, nil, fmt.Errorf("insert concurrency audit event: %w", err)
		}
	}

	var created SessionRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO user_sessions (
    user_id, token_fingerprint, authenticated_at, last_qualifying_activity_at,
    idle_expires_at, absolute_expires_at, session_expires_at, provider_type, auth_binding_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
          session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at, provider_type, auth_binding_id
`,
		user.ID,
		fingerprint,
		timing.AuthenticatedAt,
		timing.LastQualifyingActivityAt,
		timing.IdleExpiresAt,
		timing.AbsoluteExpiresAt,
		timing.SessionExpiresAt,
		providerType,
		authBindingID,
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
		&created.ProviderType,
		&created.AuthBindingID,
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

func (s *Store) SlideSession(ctx context.Context, sessionID uuid.UUID, timing SessionTiming) (SessionTiming, error) {
	var persisted SessionTiming
	if err := s.pool.QueryRow(ctx, `
WITH next_timing AS (
    SELECT id,
           authenticated_at,
           absolute_expires_at,
           GREATEST(last_qualifying_activity_at, $2::timestamptz) AS last_qualifying_activity_at
      FROM user_sessions
     WHERE id = $1
)
UPDATE user_sessions AS sessions
   SET last_qualifying_activity_at = next_timing.last_qualifying_activity_at,
       idle_expires_at = next_timing.last_qualifying_activity_at + ($3::bigint * interval '1 microsecond'),
       session_expires_at = LEAST(next_timing.last_qualifying_activity_at + ($3::bigint * interval '1 microsecond'), next_timing.absolute_expires_at),
       updated_at = next_timing.last_qualifying_activity_at
  FROM next_timing
 WHERE sessions.id = next_timing.id
RETURNING sessions.authenticated_at,
          sessions.last_qualifying_activity_at,
          sessions.idle_expires_at,
          sessions.absolute_expires_at,
          sessions.session_expires_at
`, sessionID, timing.LastQualifyingActivityAt, int64(SessionIdleTTL/time.Microsecond)).Scan(
		&persisted.AuthenticatedAt,
		&persisted.LastQualifyingActivityAt,
		&persisted.IdleExpiresAt,
		&persisted.AbsoluteExpiresAt,
		&persisted.SessionExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionTiming{}, ErrNotFound
		}
		return SessionTiming{}, fmt.Errorf("slide session expiry: %w", err)
	}
	return persisted, nil
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

func (s *Store) ListActiveSessionsForUser(ctx context.Context, userID uuid.UUID) ([]SessionRecord, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, user_id, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
       session_expires_at, revoked_at, revoke_reason_code, created_at, updated_at, provider_type, auth_binding_id
  FROM user_sessions
 WHERE user_id = $1
   AND revoked_at IS NULL
   AND session_expires_at > now()
 ORDER BY created_at ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list active sessions for user: %w", err)
	}
	defer rows.Close()

	records := make([]SessionRecord, 0)
	for rows.Next() {
		record, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active sessions for user: %w", err)
	}
	return records, nil
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

	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &userID,
		TargetUserID: &userID,
		RawSource:    "auth.totp.begin",
		RawKind:      "totp_enrollment_begun",
		ClientTxnID:  &clientTxnID,
		After: map[string]any{
			"enrollment_id":   created.ID,
			"replaces_active": replacesActive,
		},
		OccurredAt: created.CreatedAt.UTC(),
		Source:     administrativeaudit.SourceAPI,
		ActionCode: administrativeaudit.ActionTOTPEnrollmentBegun,
		TargetKind: administrativeaudit.TargetUser,
		TargetID:   userID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("replaces_active", nil, replacesActive),
			administrativeaudit.Redacted("totp_secret"),
		},
	}); err != nil {
		return PendingTOTPEnrollmentRecord{}, false, fmt.Errorf("append totp enrollment begun audit: %w", err)
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

	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &user.ID,
		TargetUserID: &user.ID,
		RawSource:    "auth.totp.complete",
		RawKind:      "totp_enrollment_completed",
		ClientTxnID:  &pending.ClientTxnID,
		After: map[string]any{
			"enrolled_at":      enrolledAt,
			"sessions_revoked": len(result.RevokedSessionIDs),
		},
		OccurredAt: enrolledAt,
		Source:     administrativeaudit.SourceAPI,
		ActionCode: administrativeaudit.ActionTOTPEnrollmentCompleted,
		TargetKind: administrativeaudit.TargetUser,
		TargetID:   user.ID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("enrolled_at", user.TOTPEnrolledAt, enrolledAt),
			administrativeaudit.Visible("sessions_revoked", 0, len(result.RevokedSessionIDs)),
			administrativeaudit.Redacted("totp_secret"),
		},
	}); err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("append totp enrollment completed audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return TOTPCompleteResult{}, fmt.Errorf("commit totp complete transaction: %w", err)
	}
	return result, nil
}

func (s *Store) GetRouteIdempotency(ctx context.Context, key RouteIdempotencyKey) (RouteIdempotencyRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID)
	var record RouteIdempotencyRecord
	if err := row.Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RouteIdempotencyRecord{}, ErrNotFound
		}
		return RouteIdempotencyRecord{}, err
	}
	return record, nil
}

// InsertRouteIdempotency stores one committed route result under a route-local
// target scope. ScopeKey must not include ActorUserID; actor scoping is enforced
// by the key and the database uniqueness constraint.
func InsertRouteIdempotency(ctx context.Context, tx pgx.Tx, key RouteIdempotencyKey, targetUserID *uuid.UUID, requestHash []byte, statusCode int, responseJSON []byte) error {
	_, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, key.RouteKey, key.ScopeKey, key.ClientTxnID, key.ActorUserID, targetUserID, requestHash, statusCode, responseJSON)
	return err
}

func InsertRouteIdempotencyPayload(ctx context.Context, tx pgx.Tx, key RouteIdempotencyKey, targetUserID *uuid.UUID, requestHash []byte, statusCode int, payload any) error {
	responseJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal idempotency payload: %w", err)
	}
	return InsertRouteIdempotency(ctx, tx, key, targetUserID, requestHash, statusCode, responseJSON)
}

func ActorOnlyRouteIdempotencyKey(routeKey string, actorUserID uuid.UUID, clientTxnID string) RouteIdempotencyKey {
	return RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actorUserID,
		ScopeKey:    "actor",
		ClientTxnID: clientTxnID,
	}
}

func (s *Store) PatchAccountProfile(
	ctx context.Context,
	actor UserRecord,
	baseUserVersion int64,
	displayName string,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (AccountProfilePatchResult, error) {
	key := ActorOnlyRouteIdempotencyKey("account.profile.patch", actor.ID, clientTxnID)
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AccountProfilePatchResult{}, ErrClientTxnConflict
		}
		return AccountProfilePatchResult{ResponseJSON: existing.ResponseJSON, Replayed: true, StatusCode: existing.StatusCode}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return AccountProfilePatchResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AccountProfilePatchResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := fetchUserForUpdate(ctx, tx, actor.ID)
	if err != nil {
		return AccountProfilePatchResult{}, err
	}
	if current.UserVersion != baseUserVersion {
		return AccountProfilePatchResult{}, ErrUserVersionConflict
	}

	updated := current
	if current.DisplayName != displayName {
		if err := tx.QueryRow(ctx, `
UPDATE users
   SET display_name = $2,
       updated_at = $3,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $1
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, actor.ID, displayName, now.UTC()).Scan(
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
			return AccountProfilePatchResult{}, err
		}

		reason := "account_profile_updated"
		if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
			ActorUserID:  &actor.ID,
			TargetUserID: &actor.ID,
			RawSource:    "account.profile.patch",
			RawKind:      "account_profile_updated",
			ReasonCode:   &reason,
			ClientTxnID:  &clientTxnID,
			RequestID:    &requestID,
			Before: map[string]any{
				"display_name": current.DisplayName,
				"user_version": current.UserVersion,
			},
			After: map[string]any{
				"display_name": updated.DisplayName,
				"user_version": updated.UserVersion,
			},
			OccurredAt: now.UTC(),
			Source:     administrativeaudit.SourceAPI,
			ActionCode: administrativeaudit.ActionUserProfileUpdated,
			TargetKind: administrativeaudit.TargetUser,
			TargetID:   actor.ID.String(),
			Changes: []administrativeaudit.Change{
				administrativeaudit.Visible("display_name", current.DisplayName, updated.DisplayName),
				administrativeaudit.Visible("user_version", current.UserVersion, updated.UserVersion),
			},
		}); err != nil {
			return AccountProfilePatchResult{}, err
		}
	}

	responseJSON, err := json.Marshal(accountProfileResponse(updated))
	if err != nil {
		return AccountProfilePatchResult{}, err
	}
	if err := InsertRouteIdempotency(ctx, tx, key, &actor.ID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AccountProfilePatchResult{}, ErrClientTxnConflict
		}
		return AccountProfilePatchResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AccountProfilePatchResult{}, err
	}
	return AccountProfilePatchResult{User: updated, ResponseJSON: responseJSON, StatusCode: http.StatusOK}, nil
}

func (s *Store) GetOrCreateAccountPreferences(ctx context.Context, userID uuid.UUID, now time.Time) (AccountPreferencesRecord, error) {
	row := s.pool.QueryRow(ctx, `
INSERT INTO account_preferences (user_id, density_mode, preferences_version, created_at, updated_at)
VALUES ($1, NULL, 1, $2, $2)
ON CONFLICT (user_id) DO NOTHING
RETURNING user_id, density_mode, preferences_version, created_at, updated_at
`, userID, now.UTC())
	record, err := scanAccountPreferences(row)
	if err == nil {
		return record, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return AccountPreferencesRecord{}, err
	}
	row = s.pool.QueryRow(ctx, `
SELECT user_id, density_mode, preferences_version, created_at, updated_at
  FROM account_preferences
 WHERE user_id = $1
`, userID)
	return scanAccountPreferences(row)
}

func (s *Store) PutAccountPreferences(
	ctx context.Context,
	actor UserRecord,
	basePreferencesVersion int64,
	clientTxnID string,
	densityMode *string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (AccountPreferencesPutResult, error) {
	key := ActorOnlyRouteIdempotencyKey("account.preferences.put", actor.ID, clientTxnID)
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AccountPreferencesPutResult{}, ErrClientTxnConflict
		}
		return AccountPreferencesPutResult{ResponseJSON: existing.ResponseJSON, Replayed: true, StatusCode: existing.StatusCode}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return AccountPreferencesPutResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AccountPreferencesPutResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `
INSERT INTO account_preferences (user_id, density_mode, preferences_version, created_at, updated_at)
VALUES ($1, NULL, 1, $2, $2)
ON CONFLICT (user_id) DO NOTHING
`, actor.ID, now.UTC()); err != nil {
		return AccountPreferencesPutResult{}, err
	}

	row := tx.QueryRow(ctx, `
SELECT user_id, density_mode, preferences_version, created_at, updated_at
  FROM account_preferences
 WHERE user_id = $1
 FOR UPDATE
`, actor.ID)
	current, err := scanAccountPreferences(row)
	if err != nil {
		return AccountPreferencesPutResult{}, err
	}
	if current.PreferencesVersion != basePreferencesVersion {
		return AccountPreferencesPutResult{}, ErrPreferencesVersionConflict
	}

	updated := current
	if !nullableStringEqual(current.DensityMode, densityMode) {
		row = tx.QueryRow(ctx, `
UPDATE account_preferences
   SET density_mode = $2,
       preferences_version = preferences_version + 1,
       updated_at = $3
 WHERE user_id = $1
RETURNING user_id, density_mode, preferences_version, created_at, updated_at
`, actor.ID, densityMode, now.UTC())
		updated, err = scanAccountPreferences(row)
		if err != nil {
			return AccountPreferencesPutResult{}, err
		}
		reason := "account_preferences_updated"
		if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
			ActorUserID:  &actor.ID,
			TargetUserID: &actor.ID,
			RawSource:    "account.preferences.put",
			RawKind:      "account_preferences_updated",
			ReasonCode:   &reason,
			ClientTxnID:  &clientTxnID,
			RequestID:    &requestID,
			Before: map[string]any{
				"density_mode":        current.DensityMode,
				"preferences_version": current.PreferencesVersion,
			},
			After: map[string]any{
				"density_mode":        updated.DensityMode,
				"preferences_version": updated.PreferencesVersion,
			},
			OccurredAt: now.UTC(),
			Source:     administrativeaudit.SourceAPI,
			ActionCode: administrativeaudit.ActionAccountPreferencesUpdated,
			TargetKind: administrativeaudit.TargetAccountPreferences,
			TargetID:   actor.ID.String(),
			Changes: []administrativeaudit.Change{
				administrativeaudit.Visible("density_mode", current.DensityMode, updated.DensityMode),
				administrativeaudit.Visible("preferences_version", current.PreferencesVersion, updated.PreferencesVersion),
			},
		}); err != nil {
			return AccountPreferencesPutResult{}, err
		}
	}

	responseJSON, err := json.Marshal(accountPreferencesResponse(updated))
	if err != nil {
		return AccountPreferencesPutResult{}, err
	}
	if err := InsertRouteIdempotency(ctx, tx, key, &actor.ID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AccountPreferencesPutResult{}, ErrClientTxnConflict
		}
		return AccountPreferencesPutResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AccountPreferencesPutResult{}, err
	}
	return AccountPreferencesPutResult{Preferences: updated, ResponseJSON: responseJSON, StatusCode: http.StatusOK}, nil
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
	key := ActorOnlyRouteIdempotencyKey("auth.password.change", user.ID, clientTxnID)
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
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

	if err := InsertRouteIdempotency(ctx, tx, key, &user.ID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return PasswordChangeResult{}, ErrClientTxnConflict
		}
		return PasswordChangeResult{}, fmt.Errorf("insert password-change idempotency: %w", err)
	}

	reason := "password_changed"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &user.ID,
		TargetUserID: &user.ID,
		RawSource:    "auth.password.change",
		RawKind:      "password_changed",
		ReasonCode:   &reason,
		ClientTxnID:  &clientTxnID,
		RequestID:    &requestID,
		After: map[string]any{
			"sessions_revoked": len(result.RevokedSessionIDs),
		},
		OccurredAt: changedAt,
		Source:     administrativeaudit.SourceAPI,
		ActionCode: administrativeaudit.ActionPasswordChanged,
		TargetKind: administrativeaudit.TargetUser,
		TargetID:   user.ID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Redacted("password"),
			administrativeaudit.Visible("sessions_revoked", 0, len(result.RevokedSessionIDs)),
		},
	}); err != nil {
		return PasswordChangeResult{}, fmt.Errorf("append password-change audit: %w", err)
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

func (s *Store) ListUsers(ctx context.Context, filter UserListFilter) ([]UserRecord, error) {
	var isActive any
	if filter.IsActive != nil {
		isActive = *filter.IsActive
	}
	var isDeploymentAdmin any
	if filter.IsDeploymentAdmin != nil {
		isDeploymentAdmin = *filter.IsDeploymentAdmin
	}
	rows, err := s.pool.Query(ctx, `
SELECT id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
       created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
  FROM users
 WHERE ($1::boolean IS NULL OR is_active = $1)
   AND ($2::boolean IS NULL OR is_deployment_admin = $2)
 ORDER BY id ASC
`, isActive, isDeploymentAdmin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]UserRecord, 0, 32)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		if !listquery.MatchSearchTokens(filter.SearchTokens, user.ID.String(), user.Email, user.DisplayName) {
			continue
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) ListAdministrativeAuditEvents(ctx context.Context, filter administrativeaudit.ListFilter) ([]administrativeaudit.Record, error) {
	return administrativeaudit.List(ctx, s.pool, filter)
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
	key := ActorOnlyRouteIdempotencyKey("users.create", actor.ID, clientTxnID)
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
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

	reason := "user_created"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &actor.ID,
		TargetUserID: &created.ID,
		RawSource:    "users.create",
		RawKind:      "user_created",
		ReasonCode:   &reason,
		ClientTxnID:  &clientTxnID,
		RequestID:    &requestID,
		After:        map[string]any{"user_id": created.ID.String()},
		OccurredAt:   now.UTC(),
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionUserCreated,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     created.ID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("display_name", nil, created.DisplayName),
			administrativeaudit.Visible("email", nil, created.Email),
			administrativeaudit.Visible("is_active", nil, created.IsActive),
			administrativeaudit.Visible("is_deployment_admin", nil, created.IsDeploymentAdmin),
			administrativeaudit.Visible("mfa_required", nil, created.MFARequired),
		},
	}); err != nil {
		return UserCreateResult{}, err
	}

	responseJSON, err := json.Marshal(safeUserResponse(created))
	if err != nil {
		return UserCreateResult{}, err
	}

	if err := InsertRouteIdempotency(ctx, tx, key, &created.ID, requestHash, http.StatusCreated, responseJSON); err != nil {
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

	actionCode, changes := userUpdateAdministrativeChanges(target, updated)
	reason := "user_updated"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &actor.ID,
		TargetUserID: &updated.ID,
		RawSource:    "users.patch",
		RawKind:      "user_updated",
		ReasonCode:   &reason,
		RequestID:    &requestID,
		Before:       map[string]any{"user_version": target.UserVersion},
		After:        map[string]any{"user_version": updated.UserVersion},
		OccurredAt:   now.UTC(),
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   actionCode,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     updated.ID.String(),
		Changes:      changes,
	}); err != nil {
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
	key := RouteIdempotencyKey{
		RouteKey:    "users.password.reset",
		ActorUserID: actor.ID,
		ScopeKey:    targetUserID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
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

	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminPasswordResetResult{}, ErrClientTxnConflict
		}
		return AdminPasswordResetResult{}, err
	}

	reason := "password_reset"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &actor.ID,
		TargetUserID: &targetUserID,
		RawSource:    "users.password.reset",
		RawKind:      "password_reset",
		ReasonCode:   &reason,
		ClientTxnID:  &clientTxnID,
		RequestID:    &requestID,
		After:        map[string]any{"user_version": updated.UserVersion},
		OccurredAt:   changedAt,
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionPasswordReset,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetUserID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Redacted("password"),
			administrativeaudit.Visible("sessions_revoked", 0, len(result.RevokedSessionIDs)),
			administrativeaudit.Visible("user_version", target.UserVersion, updated.UserVersion),
		},
	}); err != nil {
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
	key := RouteIdempotencyKey{
		RouteKey:    "users.totp.reset",
		ActorUserID: actor.ID,
		ScopeKey:    targetUserID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
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

	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminTOTPResetResult{}, ErrClientTxnConflict
		}
		return AdminTOTPResetResult{}, err
	}

	reason := "totp_reset"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &actor.ID,
		TargetUserID: &targetUserID,
		RawSource:    "users.totp.reset",
		RawKind:      "totp_reset",
		ReasonCode:   &reason,
		ClientTxnID:  &clientTxnID,
		RequestID:    &requestID,
		After:        map[string]any{"user_version": updated.UserVersion},
		OccurredAt:   changedAt,
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionTOTPReset,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetUserID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("sessions_revoked", 0, len(revoked)),
			administrativeaudit.Redacted("totp_secret"),
			administrativeaudit.Visible("user_version", target.UserVersion, updated.UserVersion),
		},
	}); err != nil {
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
	key := RouteIdempotencyKey{
		RouteKey:    "users.sessions.revoke_all",
		ActorUserID: actor.ID,
		ScopeKey:    targetUserID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
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

	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return AdminRevokeAllResult{}, ErrClientTxnConflict
		}
		return AdminRevokeAllResult{}, err
	}

	reason := "sessions_revoke_all"
	if err := appendAdministrativeProjectionTx(ctx, tx, administrativeProjection{
		ActorUserID:  &actor.ID,
		TargetUserID: &targetUserID,
		RawSource:    "users.sessions.revoke_all",
		RawKind:      "sessions_revoke_all",
		ReasonCode:   &reason,
		ClientTxnID:  &clientTxnID,
		RequestID:    &requestID,
		After:        map[string]any{"revoked_at": revokedAt},
		OccurredAt:   revokedAt,
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionSessionsRevoked,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetUserID.String(),
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("sessions_revoked", 0, len(revoked)),
		},
	}); err != nil {
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
		&session.ProviderType,
		&session.AuthBindingID,
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

func accountProfileResponse(user UserRecord) map[string]any {
	return map[string]any{
		"user_id":      user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"user_version": user.UserVersion,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}
}

func accountPreferencesResponse(record AccountPreferencesRecord) map[string]any {
	return map[string]any{
		"user_id":             record.UserID,
		"density_mode":        record.DensityMode,
		"preferences_version": record.PreferencesVersion,
		"created_at":          record.CreatedAt,
		"updated_at":          record.UpdatedAt,
	}
}

func scanAccountPreferences(row interface{ Scan(...any) error }) (AccountPreferencesRecord, error) {
	var record AccountPreferencesRecord
	err := row.Scan(
		&record.UserID,
		&record.DensityMode,
		&record.PreferencesVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func nullableStringEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
