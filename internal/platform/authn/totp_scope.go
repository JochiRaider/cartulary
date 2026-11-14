package authn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrSessionAuthorityChanged = errors.New("authn: session authority changed")
var ErrPasswordAuthorityChanged = errors.New("authn: password authority changed")
var ErrTOTPAuthorityChanged = errors.New("authn: TOTP authority changed")

type TOTPEnrollmentCreation struct {
	ID             uuid.UUID
	SecretEnvelope []byte
}

type TOTPBeginInput struct {
	User             UserRecord
	AuthScopeKind    string
	SessionID        *uuid.UUID
	BootstrapTokenID *uuid.UUID
	ClientTxnID      string
	// Nil is replay-only admission; it must never create a new enrollment.
	Creation *TOTPEnrollmentCreation
}

type BootstrapScopeError struct {
	ReasonCode string
}

func (e *BootstrapScopeError) Error() string {
	return "authn: bootstrap scope rejected: " + e.ReasonCode
}

// Callers hold the user lock before taking the auth-scope and enrollment locks.
func validateTOTPAuthScopeTx(ctx context.Context, tx pgx.Tx, user UserRecord, kind string, sessionID, bootstrapID *uuid.UUID, now time.Time) error {
	if !user.IsActive {
		return ErrSessionAuthorityChanged
	}
	switch kind {
	case "session":
		if sessionID == nil || bootstrapID != nil {
			return ErrSubjectMismatch
		}
		var owner uuid.UUID
		var revokedAt *time.Time
		var expiresAt time.Time
		err := tx.QueryRow(ctx, `SELECT user_id, revoked_at, session_expires_at FROM user_sessions WHERE id=$1 FOR UPDATE`, *sessionID).Scan(&owner, &revokedAt, &expiresAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionAuthorityChanged
		}
		if err != nil {
			return fmt.Errorf("lock TOTP session scope: %w", err)
		}
		if owner != user.ID {
			return ErrSubjectMismatch
		}
		if revokedAt != nil || !expiresAt.After(now) {
			return ErrSessionAuthorityChanged
		}
	case "bootstrap_token":
		if bootstrapID == nil || sessionID != nil {
			return ErrSubjectMismatch
		}
		var token BootstrapTokenRecord
		err := tx.QueryRow(ctx, `SELECT id, user_id, issued_at, expires_at, consumed_at, superseded_at FROM bootstrap_tokens WHERE id=$1 FOR UPDATE`, *bootstrapID).Scan(
			&token.ID, &token.UserID, &token.IssuedAt, &token.ExpiresAt, &token.ConsumedAt, &token.SupersededAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionAuthorityChanged
		}
		if err != nil {
			return fmt.Errorf("lock TOTP bootstrap scope: %w", err)
		}
		if token.UserID != user.ID {
			return ErrSubjectMismatch
		}
		if reason := BootstrapReasonCode(token, now); reason != "" {
			return &BootstrapScopeError{ReasonCode: reason}
		}
	default:
		return ErrSubjectMismatch
	}
	return nil
}
