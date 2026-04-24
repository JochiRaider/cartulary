package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const unauthorizedCode = "session_required"

type Service struct {
	store      authStore
	hub        sessionHub
	keys       authn.MasterKeys
	pagination *pagination.Registry
	now        func() time.Time
}

type authStore interface {
	GetUserByNormalizedEmail(context.Context, string) (authn.UserRecord, error)
	GetUserByID(context.Context, uuid.UUID) (authn.UserRecord, error)
	ListIncidentMembershipSummaries(context.Context, uuid.UUID) ([]authn.IncidentMembershipSummary, error)
	GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error)
	CreateSessionWithConcurrency(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string) (authn.SessionRecord, *authn.SessionRecord, error)
	SlideSession(context.Context, uuid.UUID, authn.SessionTiming) error
	RevokeSession(context.Context, uuid.UUID, string, time.Time) error
	IssueBootstrapToken(context.Context, uuid.UUID, []byte, time.Time) (authn.BootstrapTokenRecord, error)
	GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
	GetPendingTOTPEnrollmentForUser(context.Context, uuid.UUID, time.Time) (*authn.PendingTOTPEnrollmentRecord, error)
	GetPendingTOTPEnrollmentByID(context.Context, uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error)
	BeginTOTPEnrollment(context.Context, uuid.UUID, string, *uuid.UUID, *uuid.UUID, string, []byte, []byte, bool, time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error)
	ActivateTOTPEnrollment(context.Context, authn.UserRecord, uuid.UUID, string, *uuid.UUID, *uuid.UUID, time.Time) (authn.TOTPCompleteResult, error)
	GetRouteIdempotency(context.Context, string, string, string) (authn.RouteIdempotencyRecord, error)
	ChangePassword(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error)
	ListUsers(context.Context) ([]authn.UserRecord, error)
	CreateUser(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error)
	UpdateUser(context.Context, authn.UserRecord, uuid.UUID, int64, *string, *string, *bool, *bool, *bool, string, time.Time) (authn.UserRecord, []uuid.UUID, error)
	AdminResetPassword(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, []byte, string, time.Time) (authn.AdminPasswordResetResult, error)
	AdminResetTOTP(context.Context, authn.UserRecord, uuid.UUID, int64, string, []byte, string, time.Time) (authn.AdminTOTPResetResult, error)
	AdminRevokeAllSessions(context.Context, authn.UserRecord, uuid.UUID, string, []byte, string, time.Time) (authn.AdminRevokeAllResult, error)
}

type sessionHub interface {
	RegisterSession(uuid.UUID) (<-chan string, func())
	RevokeSession(uuid.UUID, string)
}

type SessionPrincipal struct {
	AuthSource   AuthSource
	SessionToken string
	Session      authn.SessionRecord
	User         authn.UserRecord
}

type CredentialAuthContext struct {
	Principal          *SessionPrincipal
	BootstrapToken     *authn.BootstrapTokenRecord
	BootstrapTokenText string
	User               authn.UserRecord
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/auth/login", service.handleLogin)
		mux.HandleFunc("/api/v1/auth/session", service.handleSession)
		mux.HandleFunc("/api/v1/auth/logout", service.handleLogout)
		mux.HandleFunc("/api/v1/auth/credential-state", service.handleCredentialState)
		mux.HandleFunc("/api/v1/auth/password/change", service.handlePasswordChange)
		mux.HandleFunc("/api/v1/auth/mfa/totp/begin", service.handleTOTPBegin)
		mux.HandleFunc("/api/v1/auth/mfa/totp/complete", service.handleTOTPComplete)
		mux.HandleFunc("/api/v1/users", service.handleUsersCollection)
		mux.HandleFunc("/api/v1/users/", service.handleUsersMember)
		return nil
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/test/auth/touch", service.handleTouch)
		mux.HandleFunc("/ws/v1/test/session-lifecycle", service.handleTestSocket)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	paginator := deps.Pagination
	if paginator == nil {
		paginator = pagination.NewRegistry()
	}

	return &Service{
		store:      authn.NewStore(deps.Postgres),
		hub:        deps.WSHub,
		keys:       keys,
		pagination: paginator,
		now:        now,
	}, nil
}

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, apiErr := DecodeLoginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	user, err := s.store.GetUserByNormalizedEmail(r.Context(), request.Username)
	if err != nil || !user.IsActive {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Details: map[string]any{}})
		return
	}

	ok, err := authn.VerifyPasswordHash(user.PasswordHash, request.Password)
	if err != nil || !ok {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Details: map[string]any{}})
		return
	}

	if user.MFARequired {
		hasActiveTOTP := user.TOTPEnrolledAt != nil && len(user.TOTPSecretCiphertext) > 0 && len(user.TOTPSecretNonce) > 0
		if !hasActiveTOTP {
			bootstrapToken, bootstrapExpiresAt, err := s.issueBootstrapToken(r.Context(), user.ID)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			writeAPIError(w, r, &APIError{
				Status: http.StatusUnauthorized,
				Code:   "mfa_setup_required",
				Details: map[string]any{
					"required_setup_kinds": []string{"totp"},
					"bootstrap_token":      bootstrapToken,
					"bootstrap_expires_at": bootstrapExpiresAt,
				},
			})
			return
		}

		if request.SecondFactor == nil {
			writeAPIError(w, r, &APIError{
				Status: http.StatusUnauthorized,
				Code:   "mfa_required",
				Details: map[string]any{
					"required_second_factor_kinds": []string{"totp"},
				},
			})
			return
		}

		secretBytes, err := authn.DecryptSecret(s.keys, user.TOTPSecretCiphertext, user.TOTPSecretNonce)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		secretBase32 := strings.ToUpper(authn.EncodeSecretBase32(secretBytes))
		if !authn.ValidateTOTPCode(secretBase32, request.SecondFactor.Code, s.now()) {
			writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}})
			return
		}
	}

	sessionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	session, revoked, err := s.store.CreateSessionWithConcurrency(
		r.Context(),
		user,
		authn.FingerprintToken(s.keys, sessionToken),
		authn.NewSessionTiming(s.now()),
		httpapi.RequestIDFromContext(r.Context()),
	)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if revoked != nil {
		s.hub.RevokeSession(revoked.ID, authn.ConcurrencyLimitReasonCode)
	}

	s.setAuthCookies(w, sessionToken)
	resource, err := s.buildSessionResource(r.Context(), user, session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if apiErr := ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	resource, err := s.buildSessionResource(r.Context(), principal.User, principal.Session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	if err := s.store.RevokeSession(r.Context(), principal.Session.ID, "session_revoked", s.now()); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.hub.RevokeSession(principal.Session.ID, "session_revoked")

	s.clearAuthCookies(w)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"user_id":          principal.User.ID,
		"sessions_revoked": false,
		"logged_out":       true,
	})
}

func (s *Service) handleCredentialState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	pending, err := s.store.GetPendingTOTPEnrollmentForUser(r.Context(), principal.User.ID, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var pendingExpiresAt *time.Time
	if pending != nil {
		pendingExpiresAt = &pending.ExpiresAt
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildCredentialStateResource(principal.User, pendingExpiresAt))
}

func (s *Service) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, apiErr := DecodePasswordChangeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	requestHash := hashRequestPayload(map[string]any{
		"current_password": requestSecretFingerprint(s.keys, request.CurrentPassword),
		"new_password":     requestSecretFingerprint(s.keys, request.NewPassword),
		"second_factor":    requestSecondFactorHashPayload(s.keys, request.SecondFactor),
	})
	if existing, err := s.store.GetRouteIdempotency(r.Context(), "auth.password.change", principal.User.ID.String(), request.ClientTxnID); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		s.clearAuthCookies(w)
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
		return
	} else if !errors.Is(err, authn.ErrNotFound) {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	ok, err := authn.VerifyPasswordHash(principal.User.PasswordHash, request.CurrentPassword)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !ok {
		writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_current_password", Details: map[string]any{}})
		return
	}
	if apiErr := s.validateActiveTOTP(principal.User, request.SecondFactor); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	newPasswordHash, err := authn.HashPassword(request.NewPassword)
	if err != nil {
		writeAPIError(w, r, invalidAuthRequest("new_password", err.Error()))
		return
	}

	result, err := s.store.ChangePassword(
		r.Context(),
		principal.User,
		request.ClientTxnID,
		requestHash,
		newPasswordHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	for _, sessionID := range result.RevokedSessionIDs {
		s.hub.RevokeSession(sessionID, "session_revoked")
	}

	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.clearAuthCookies(w)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

func (s *Service) handleTOTPBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, apiErr := DecodeTOTPBeginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	authContext, apiErr := s.authenticateAuthRequest(r, true, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	replacesActive := authContext.User.TOTPEnrolledAt != nil && len(authContext.User.TOTPSecretCiphertext) > 0 && len(authContext.User.TOTPSecretNonce) > 0
	if authContext.Principal != nil && replacesActive {
		if request.CurrentPassword == nil {
			writeAPIError(w, r, invalidAuthRequest("current_password", "current_password is required for totp replacement"))
			return
		}
		ok, err := authn.VerifyPasswordHash(authContext.User.PasswordHash, *request.CurrentPassword)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if !ok {
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_current_password", Details: map[string]any{}})
			return
		}
		if apiErr := s.validateActiveTOTP(authContext.User, request.SecondFactor); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
	}

	clearSecret, secretBase32, err := authn.GenerateTOTPSecret()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	ciphertext, nonce, err := authn.EncryptSecret(s.keys, clearSecret)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var sessionID *uuid.UUID
	var bootstrapTokenID *uuid.UUID
	authScopeKind := "bootstrap_token"
	if authContext.Principal != nil {
		sessionID = &authContext.Principal.Session.ID
		authScopeKind = "session"
	} else if authContext.BootstrapToken != nil {
		bootstrapTokenID = &authContext.BootstrapToken.ID
	}

	pending, replayed, err := s.store.BeginTOTPEnrollment(
		r.Context(),
		authContext.User.ID,
		authScopeKind,
		sessionID,
		bootstrapTokenID,
		request.ClientTxnID,
		ciphertext,
		nonce,
		replacesActive,
		s.now(),
	)
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if replayed {
		clearSecret, err = authn.DecryptSecret(s.keys, pending.SecretCiphertext, pending.SecretNonce)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		secretBase32 = authn.EncodeSecretBase32(clearSecret)
	}

	if authContext.Principal != nil {
		if err := s.slideSessionIfNeeded(r.Context(), authContext.Principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"enrollment_id": pending.ID,
		"expires_at":    pending.ExpiresAt,
		"totp_setup":    BuildTOTPSetup(secretBase32, authContext.User.Email),
	})
}

func (s *Service) handleTOTPComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, apiErr := DecodeTOTPCompleteRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	authContext, apiErr := s.authenticateAuthRequest(r, true, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	enrollmentID, err := uuid.Parse(request.EnrollmentID)
	if err != nil {
		writeAPIError(w, r, invalidAuthRequest("enrollment_id", "enrollment_id must be a uuid"))
		return
	}

	pending, err := s.store.GetPendingTOTPEnrollmentByID(r.Context(), enrollmentID)
	if errors.Is(err, authn.ErrNotFound) {
		writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if pending.UserID != authContext.User.ID {
		if authContext.BootstrapToken != nil {
			writeAPIError(w, r, BootstrapRejectedError("subject_mismatch"))
		} else {
			writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		}
		return
	}
	if status := authn.PendingEnrollmentStatusAt(pending, s.now()); status == authn.PendingEnrollmentExpired {
		writeAPIError(w, r, TOTPSetupNotPendingError("expired"))
		return
	} else if status == authn.PendingEnrollmentConsumed {
		writeAPIError(w, r, TOTPSetupNotPendingError("consumed"))
		return
	}

	clearSecret, err := authn.DecryptSecret(s.keys, pending.SecretCiphertext, pending.SecretNonce)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !authn.ValidateTOTPCode(authn.EncodeSecretBase32(clearSecret), request.Code, s.now()) {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}})
		return
	}

	var sessionID *uuid.UUID
	var bootstrapTokenID *uuid.UUID
	authScopeKind := "bootstrap_token"
	if authContext.Principal != nil {
		sessionID = &authContext.Principal.Session.ID
		authScopeKind = "session"
	} else if authContext.BootstrapToken != nil {
		bootstrapTokenID = &authContext.BootstrapToken.ID
	}

	result, err := s.store.ActivateTOTPEnrollment(
		r.Context(),
		authContext.User,
		enrollmentID,
		authScopeKind,
		sessionID,
		bootstrapTokenID,
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrNotFound):
		writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		return
	case errors.Is(err, authn.ErrPendingExpired):
		writeAPIError(w, r, TOTPSetupNotPendingError("expired"))
		return
	case errors.Is(err, authn.ErrPendingConsumed):
		writeAPIError(w, r, TOTPSetupNotPendingError("consumed"))
		return
	case errors.Is(err, authn.ErrSubjectMismatch):
		if authContext.BootstrapToken != nil {
			writeAPIError(w, r, BootstrapRejectedError("subject_mismatch"))
		} else {
			writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		}
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	for _, sessionID := range result.RevokedSessionIDs {
		s.hub.RevokeSession(sessionID, "session_revoked")
	}
	if result.SessionsRevoked {
		s.clearAuthCookies(w)
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"user_id":          authContext.User.ID,
		"totp":             map[string]any{"enrolled_at": result.EnrolledAt},
		"sessions_revoked": result.SessionsRevoked,
	})
}

func (s *Service) handleUsersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		binding, cursor, reasonCode := pagination.ResolveRequest(
			r.URL.Query(),
			"users.list",
			principal.User.ID.String(),
			nil,
		)
		if reasonCode != "" {
			writeAPIError(w, r, userPaginationError(reasonCode))
			return
		}

		var (
			rows       []json.RawMessage
			nextCursor *pagination.Cursor
			err        error
		)
		if cursor == nil {
			users, listErr := s.store.ListUsers(r.Context())
			if listErr != nil {
				writeAPIError(w, r, internalAPIError(listErr))
				return
			}
			resources := make([]map[string]any, 0, len(users))
			for _, user := range users {
				resources = append(resources, BuildSafeUserResource(user))
			}
			rows, err = pagination.MarshalResources(resources)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			rows, nextCursor = s.pagination.Start(binding, rows)
		} else {
			rows, nextCursor, err = s.pagination.Continue(binding, *cursor)
			switch {
			case errors.Is(err, pagination.ErrCursorSnapshotExpired):
				writeAPIError(w, r, userPaginationError(pagination.ReasonCursorSnapshotUnavailable))
				return
			case errors.Is(err, pagination.ErrInvalidCursorToken):
				writeAPIError(w, r, userPaginationError(pagination.ReasonInvalidCursorToken))
				return
			case err != nil:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}

		var nextCursorToken *string
		if nextCursor != nil {
			token, err := pagination.EncodeCursor(*nextCursor)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			nextCursorToken = &token
		}
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
			"users": rows,
		}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextCursorToken != nil,
			NextCursor: nextCursorToken,
		})
	case http.MethodPost:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeUserCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"auth_kind":           request.AuthKind,
			"email":               request.Email,
			"display_name":        request.DisplayName,
			"initial_password":    requestSecretFingerprint(s.keys, request.InitialPassword),
			"mfa_required":        request.MFARequired,
			"is_deployment_admin": request.IsDeploymentAdmin,
		})
		passwordHash, err := authn.HashPassword(request.InitialPassword)
		if err != nil {
			writeAPIError(w, r, invalidMutationPayload("initial_password", "invalid_password"))
			return
		}
		result, err := s.store.CreateUser(
			r.Context(),
			principal.User,
			request.Email,
			request.DisplayName,
			passwordHash,
			request.MFARequired,
			request.IsDeploymentAdmin,
			request.ClientTxnID,
			requestHash,
			httpapi.RequestIDFromContext(r.Context()),
			s.now(),
		)
		if errors.Is(err, authn.ErrClientTxnConflict) {
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		}
		if authn.IsUniqueViolation(err) {
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_mutation_payload", Details: map[string]any{"field": "email"}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, result.StatusCode, payload)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleUsersMember(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if trimmed == "" || trimmed == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	segments := strings.Split(trimmed, "/")
	userID, err := uuid.Parse(segments[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if match, ok := httpapi.MatchReservedExtensionFamily(r.URL.Path); ok && !match.Claimed {
		writeAPIError(w, r, &APIError{
			Status: http.StatusNotFound,
			Code:   "extension_profile_not_claimed",
			Details: map[string]any{
				"profile_id":   match.ProfileID,
				"route_family": match.RouteFamily,
			},
		})
		return
	}

	if len(segments) == 1 {
		switch r.Method {
		case http.MethodGet:
			if apiErr := ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			principal, apiErr := s.authenticateSessionRequest(r, false)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			user, err := s.store.GetUserByID(r.Context(), userID)
			if errors.Is(err, authn.ErrNotFound) {
				writeAPIError(w, r, &APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}})
				return
			}
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildSafeUserResource(user))
		case http.MethodPatch:
			principal, apiErr := s.authenticateSessionRequest(r, true)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			request, apiErr := DecodeUserPatchRequest(r.Body)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			user, revokedSessions, err := s.store.UpdateUser(
				r.Context(),
				principal.User,
				userID,
				request.BaseUserVersion,
				request.Email,
				request.DisplayName,
				request.IsActive,
				request.MFARequired,
				request.IsDeploymentAdmin,
				httpapi.RequestIDFromContext(r.Context()),
				s.now(),
			)
			switch {
			case errors.Is(err, authn.ErrUserVersionConflict):
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
				return
			case errors.Is(err, authn.ErrLastDeploymentAdmin):
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "last_deployment_admin", Details: map[string]any{}})
				return
			case authn.IsUniqueViolation(err):
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_mutation_payload", Details: map[string]any{"field": "email"}})
				return
			case err != nil:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			for _, sessionID := range revokedSessions {
				s.hub.RevokeSession(sessionID, "session_revoked")
				if sessionID == principal.Session.ID {
					s.clearAuthCookies(w)
				}
			}
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildSafeUserResource(user))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(segments) < 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	switch strings.Join(segments[1:], "/") {
	case "password/reset":
		request, apiErr := DecodeAdminPasswordResetRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version": request.BaseUserVersion,
			"new_password":      requestSecretFingerprint(s.keys, request.NewPassword),
			"reason":            request.Reason,
		})
		passwordHash, err := authn.HashPassword(request.NewPassword)
		if err != nil {
			writeAPIError(w, r, invalidMutationPayload("new_password", "invalid_password"))
			return
		}
		result, err := s.store.AdminResetPassword(
			r.Context(),
			principal.User,
			userID,
			request.BaseUserVersion,
			passwordHash,
			request.ClientTxnID,
			requestHash,
			httpapi.RequestIDFromContext(r.Context()),
			s.now(),
		)
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, authn.ErrUserVersionConflict):
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.hub.RevokeSession(sessionID, "session_revoked")
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	case "mfa/totp/reset":
		request, apiErr := DecodeAdminTOTPResetRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version": request.BaseUserVersion,
			"reason":            request.Reason,
		})
		result, err := s.store.AdminResetTOTP(
			r.Context(),
			principal.User,
			userID,
			request.BaseUserVersion,
			request.ClientTxnID,
			requestHash,
			httpapi.RequestIDFromContext(r.Context()),
			s.now(),
		)
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, authn.ErrUserVersionConflict):
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.hub.RevokeSession(sessionID, "session_revoked")
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	case "sessions/revoke-all":
		request, apiErr := DecodeAdminRevokeAllRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"reason": request.Reason,
		})
		result, err := s.store.AdminRevokeAllSessions(
			r.Context(),
			principal.User,
			userID,
			request.ClientTxnID,
			requestHash,
			httpapi.RequestIDFromContext(r.Context()),
			s.now(),
		)
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.hub.RevokeSession(sessionID, "session_revoked")
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	default:
		http.NotFound(w, r)
	}
}

func (s *Service) handleTouch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}.Slide(s.now())
	if err := s.store.SlideSession(r.Context(), principal.Session.ID, sliding); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	principal.Session.LastQualifyingActivityAt = sliding.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = sliding.IdleExpiresAt
	principal.Session.SessionExpiresAt = sliding.SessionExpiresAt
	resource, err := s.buildSessionResource(r.Context(), principal.User, principal.Session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleTestSocket(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	conn, err := platformws.Accept(w, r)
	if err != nil {
		return
	}
	closed := false
	defer func() {
		if !closed {
			conn.CloseNow()
		}
	}()

	revocations, unregister := s.hub.RegisterSession(principal.Session.ID)
	defer unregister()

	ctx := context.Background()
	if err := platformws.WriteJSON(ctx, conn, platformws.Message{
		Type:    "connected",
		Payload: platformws.RawPayload(map[string]any{"session_id": principal.Session.ID.String()}),
	}); err != nil {
		return
	}

	type readResult struct {
		closeRequested bool
	}
	readResults := make(chan readResult, 1)
	go func() {
		for {
			var message platformws.Message
			if err := platformws.ReadJSON(context.Background(), conn, &message); err != nil {
				readResults <- readResult{}
				return
			}
			if message.Type == "close_me" {
				readResults <- readResult{closeRequested: true}
				return
			}
		}
	}()

	for {
		select {
		case reasonCode := <-revocations:
			writeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := platformws.WriteJSON(writeCtx, conn, platformws.Message{
				Type:    "session_revoked",
				Payload: platformws.RawPayload(map[string]any{"reason_code": reasonCode}),
			})
			cancel()
			if err != nil {
				return
			}
			if err := conn.Close(websocket.StatusPolicyViolation, "session_revoked"); err == nil {
				closed = true
			}
			return
		case result := <-readResults:
			if result.closeRequested {
				if err := conn.Close(websocket.StatusNormalClosure, "client_complete"); err == nil {
					closed = true
				}
			}
			return
		}
	}
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (SessionPrincipal, *APIError) {
	context, apiErr := s.authenticateAuthRequest(r, stateChanging, false)
	if apiErr != nil {
		return SessionPrincipal{}, apiErr
	}
	if context.Principal == nil {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	return *context.Principal, nil
}

func (s *Service) authenticateAuthRequest(r *http.Request, stateChanging bool, allowBootstrap bool) (CredentialAuthContext, *APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		if principal, apiErr := s.authenticateSessionToken(r, AuthSourceBearer, token, false); apiErr == nil {
			return CredentialAuthContext{Principal: &principal, User: principal.User}, nil
		} else if apiErr.Code != unauthorizedCode {
			return CredentialAuthContext{}, apiErr
		}

		bootstrapToken, user, err := s.store.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(s.keys, token))
		if err == nil {
			if reason := authn.BootstrapReasonCode(bootstrapToken, s.now()); reason != "" {
				return CredentialAuthContext{}, BootstrapRejectedError(reason)
			}
			if !allowBootstrap || !AllowsBootstrapTokenRoute(r.URL.Path) {
				return CredentialAuthContext{}, BootstrapRejectedError("not_allowed_for_route")
			}
			if cookie, _ := r.Cookie(authn.SessionCookieName); cookie != nil {
				return CredentialAuthContext{}, invalidAuthRequest("authorization", "exactly one auth mode is allowed")
			}
			return CredentialAuthContext{
				BootstrapToken:     &bootstrapToken,
				BootstrapTokenText: token,
				User:               user,
			}, nil
		}
		if err != nil && !errors.Is(err, authn.ErrNotFound) {
			return CredentialAuthContext{}, internalAPIError(err)
		}
		return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if errors.Is(err, authn.ErrNotFound) {
		return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		return CredentialAuthContext{}, internalAPIError(err)
	}
	principal, apiErr := s.authenticateSessionToken(r, AuthSourceCookie, cookie.Value, stateChanging)
	if apiErr != nil {
		return CredentialAuthContext{}, apiErr
	}
	return CredentialAuthContext{Principal: &principal, User: principal.User}, nil
}

func (s *Service) authenticateSessionToken(r *http.Request, authSource AuthSource, sessionToken string, stateChanging bool) (SessionPrincipal, *APIError) {
	if stateChanging && authSource == AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(s.keys, sessionToken) {
			return SessionPrincipal{}, &APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return SessionPrincipal{}, apiErr
		}
	}

	session, user, err := s.store.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(s.keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		return SessionPrincipal{}, internalAPIError(err)
	}
	if user.IsActive == false {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if session.RevokedAt != nil {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.store.RevokeSession(context.Background(), session.ID, "session_expired", now)
		s.hub.RevokeSession(session.ID, "session_expired")
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	return SessionPrincipal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}

func (s *Service) issueBootstrapToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := authn.GenerateOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	record, err := s.store.IssueBootstrapToken(ctx, userID, authn.FingerprintToken(s.keys, token), s.now())
	if err != nil {
		return "", time.Time{}, err
	}
	return token, record.ExpiresAt, nil
}

func (s *Service) setAuthCookies(w http.ResponseWriter, sessionToken string) {
	csrfToken := authn.CSRFTokenForSessionToken(s.keys, sessionToken)
	http.SetCookie(w, &http.Cookie{
		Name:     authn.SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     authn.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
}

func (s *Service) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authn.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     authn.CSRFCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
}

func (s *Service) buildSessionResource(ctx context.Context, user authn.UserRecord, session authn.SessionRecord) (map[string]any, error) {
	mfaState := "not_required"
	if user.MFARequired {
		mfaState = "satisfied"
	}

	memberships, err := s.store.ListIncidentMembershipSummaries(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	summaries := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		summaries = append(summaries, map[string]any{
			"incident_id": membership.IncidentID,
			"role":        membership.Role,
		})
	}

	return map[string]any{
		"user_id":             user.ID,
		"display_name":        user.DisplayName,
		"provider_type":       "local",
		"mfa_state":           mfaState,
		"is_deployment_admin": user.IsDeploymentAdmin,
		"authenticated_at":    session.AuthenticatedAt,
		"idle_expires_at":     session.IdleExpiresAt,
		"absolute_expires_at": session.AbsoluteExpiresAt,
		"session_expires_at":  session.SessionExpiresAt,
		"memberships":         summaries,
	}, nil
}

func (s *Service) validateActiveTOTP(user authn.UserRecord, secondFactor *SecondFactorAssertion) *APIError {
	if user.TOTPEnrolledAt == nil || len(user.TOTPSecretCiphertext) == 0 || len(user.TOTPSecretNonce) == 0 {
		return nil
	}
	if secondFactor == nil {
		return &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}}
	}
	secretBytes, err := authn.DecryptSecret(s.keys, user.TOTPSecretCiphertext, user.TOTPSecretNonce)
	if err != nil {
		return internalAPIError(err)
	}
	if !authn.ValidateTOTPCode(authn.EncodeSecretBase32(secretBytes), secondFactor.Code, s.now()) {
		return &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}}
	}
	return nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *SessionPrincipal, method string, path string) error {
	if principal == nil || !ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}.Slide(s.now())
	if err := s.store.SlideSession(ctx, principal.Session.ID, sliding); err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = sliding.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = sliding.IdleExpiresAt
	principal.Session.SessionExpiresAt = sliding.SessionExpiresAt
	return nil
}

func requestSecretFingerprint(keys authn.MasterKeys, value string) []byte {
	return authn.FingerprintRequestValue(keys, value)
}

func requestSecondFactorHashPayload(keys authn.MasterKeys, secondFactor *SecondFactorAssertion) any {
	if secondFactor == nil {
		return nil
	}
	return map[string]any{
		"kind": secondFactor.Kind,
		"code": requestSecretFingerprint(keys, secondFactor.Code),
	}
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func hashesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
