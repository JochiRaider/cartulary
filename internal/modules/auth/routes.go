package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

const (
	unauthorizedCode         = "session_required"
	sessionExpiredReasonCode = "session_expired"
	sessionRevokedReasonCode = "session_revoked"
)

const (
	EnterpriseOIDCVerifierOverrideKey = "auth.enterprise_oidc_verifier"
	EnterpriseSAMLVerifierOverrideKey = "auth.enterprise_saml_verifier"
)

type Service struct {
	store        authStore
	revocations  sessionRevocationPublisher
	keys         authn.MasterKeys
	cursorCodec  *pagination.Codec
	now          func() time.Time
	profiles     []httpapi.ExtensionProfile
	oidcVerifier enterpriseOIDCVerifier
	samlVerifier enterpriseSAMLVerifier
}

type authStore interface {
	localLoginStore
	sessionStore
	sessionMembershipSummaryReader
	credentialLifecycleStore
	accountStore
	userAdminStore
	administrativeAuditStore
	enterpriseAuthStore
}

type localLoginStore interface {
	GetUserByNormalizedEmail(context.Context, string) (authn.UserRecord, error)
}

type sessionStore interface {
	GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error)
	CreateSessionWithConcurrency(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string) (authn.SessionRecord, *authn.SessionRecord, error)
	CreateSessionWithProviderConcurrency(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string, string, *uuid.UUID) (authn.SessionRecord, *authn.SessionRecord, error)
	SlideSession(context.Context, uuid.UUID, authn.SessionTiming) (authn.SessionTiming, error)
	RevokeSession(context.Context, uuid.UUID, string, time.Time) error
	IssueBootstrapToken(context.Context, uuid.UUID, []byte, time.Time) (authn.BootstrapTokenRecord, error)
	GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
}

// sessionMembershipSummaryReader feeds informational session bootstrap state;
// route authorization must use route-specific checks instead of this resource.
type sessionMembershipSummaryReader interface {
	ListIncidentMembershipSummaries(context.Context, uuid.UUID) ([]authn.IncidentMembershipSummary, error)
}

type credentialLifecycleStore interface {
	GetPendingTOTPEnrollmentForUser(context.Context, uuid.UUID, time.Time) (*authn.PendingTOTPEnrollmentRecord, error)
	GetPendingTOTPEnrollmentByID(context.Context, uuid.UUID) (*authn.PendingTOTPEnrollmentRecord, error)
	BeginTOTPEnrollment(context.Context, uuid.UUID, string, *uuid.UUID, *uuid.UUID, string, []byte, []byte, bool, time.Time) (authn.PendingTOTPEnrollmentRecord, bool, error)
	ActivateTOTPEnrollment(context.Context, authn.UserRecord, uuid.UUID, string, *uuid.UUID, *uuid.UUID, time.Time) (authn.TOTPCompleteResult, error)
	GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	ChangePassword(context.Context, authn.UserRecord, string, []byte, string, string, time.Time) (authn.PasswordChangeResult, error)
}

type accountStore interface {
	PatchAccountProfile(context.Context, authn.UserRecord, int64, string, string, []byte, string, time.Time) (authn.AccountProfilePatchResult, error)
	GetOrCreateAccountPreferences(context.Context, uuid.UUID, time.Time) (authn.AccountPreferencesRecord, error)
	PutAccountPreferences(context.Context, authn.UserRecord, int64, string, *string, []byte, string, time.Time) (authn.AccountPreferencesPutResult, error)
}

type userAdminStore interface {
	GetUserByID(context.Context, uuid.UUID) (authn.UserRecord, error)
	ListUsers(context.Context, authn.UserListFilter) ([]authn.UserRecord, error)
	CreateUser(context.Context, authn.UserRecord, string, string, string, bool, bool, string, []byte, string, time.Time) (authn.UserCreateResult, error)
	UpdateUser(context.Context, authn.UserRecord, uuid.UUID, int64, *string, *string, *bool, *bool, *bool, string, time.Time) (authn.UserRecord, []uuid.UUID, error)
	AdminResetPassword(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, []byte, string, time.Time) (authn.AdminPasswordResetResult, error)
	AdminResetTOTP(context.Context, authn.UserRecord, uuid.UUID, int64, string, []byte, string, time.Time) (authn.AdminTOTPResetResult, error)
	AdminRevokeAllSessions(context.Context, authn.UserRecord, uuid.UUID, string, []byte, string, time.Time) (authn.AdminRevokeAllResult, error)
	ListEnterpriseAuthBindingSummaries(context.Context, uuid.UUID) ([]authn.EnterpriseAuthBindingSummary, error)
}

type administrativeAuditStore interface {
	ListAdministrativeAuditEvents(context.Context, authn.AdministrativeAuditFilter) ([]authn.AdministrativeAuditRecord, error)
}

type enterpriseAuthStore interface {
	ListEnterpriseAuthProviders(context.Context) ([]authn.EnterpriseAuthProviderRecord, error)
	GetEnterpriseAuthProviderByKey(context.Context, string) (authn.EnterpriseAuthProviderRecord, error)
	CreateEnterpriseAuthTransaction(context.Context, authn.EnterpriseAuthProviderRecord, string, *string, *string, []byte, *string, []byte, time.Time) (authn.EnterpriseAuthTransactionRecord, error)
	CompleteOIDCEnterpriseAuthTransaction(context.Context, string, string, []byte, *string, string, time.Time) (authn.EnterpriseAuthCompletionResult, error)
	StageSAMLEnterpriseAuthTransaction(context.Context, string, string, string, []byte, time.Time) (authn.EnterpriseAuthTransactionRecord, error)
	CompleteSAMLEnterpriseAuthTransaction(context.Context, string, []byte, []byte, time.Time) (authn.EnterpriseAuthCompletionResult, error)
	CreateEnterpriseAuthBinding(context.Context, authn.UserRecord, uuid.UUID, int64, string, string, string, *string, []byte, string, time.Time) (authn.EnterpriseAuthBindingResult, error)
	RotateEnterpriseAuthBinding(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, int64, string, string, *string, []byte, string, time.Time) (authn.EnterpriseAuthBindingResult, error)
	RetireEnterpriseAuthBinding(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, int64, string, *string, []byte, string, time.Time) (authn.EnterpriseAuthBindingResult, error)
}

type sessionRevocationPublisher interface {
	RevokeSession(uuid.UUID, string)
}

type SessionPrincipal = httpauth.Principal

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
		mux.HandleFunc("/api/v1/auth/providers", service.handleEnterpriseProviders)
		mux.HandleFunc("/api/v1/auth/providers/", service.handleEnterpriseProviders)
		mux.HandleFunc("/api/v1/auth/oidc/", service.handleEnterpriseOIDC)
		mux.HandleFunc("/api/v1/auth/saml/", service.handleEnterpriseSAML)
		mux.HandleFunc("/api/v1/auth/credential-state", service.handleCredentialState)
		mux.HandleFunc("/api/v1/auth/password/change", service.handlePasswordChange)
		mux.HandleFunc("/api/v1/auth/mfa/totp/begin", service.handleTOTPBegin)
		mux.HandleFunc("/api/v1/auth/mfa/totp/complete", service.handleTOTPComplete)
		mux.HandleFunc("/api/v1/account/profile", service.handleAccountProfile)
		mux.HandleFunc("/api/v1/account/preferences", service.handleAccountPreferences)
		mux.HandleFunc("/api/v1/administrative-audit-events", service.handleAdministrativeAuditEvents)
		mux.HandleFunc("/api/v1/users", service.handleUsersCollection)
		mux.HandleFunc("/api/v1/users/", service.handleUsersMember)
		return nil
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/test/auth/touch", guard.Protect(service.handleTouch))
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
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}

	oidcVerifier := enterpriseOIDCVerifier(enterpriseauth.UnconfiguredOIDCVerifier{})
	if override, ok := deps.ModuleOverrides[EnterpriseOIDCVerifierOverrideKey]; ok {
		if verifier, ok := override.(enterpriseOIDCVerifier); ok {
			oidcVerifier = verifier
		} else {
			return nil, fmt.Errorf("auth enterprise OIDC verifier override has type %T", override)
		}
	}
	samlVerifier := enterpriseSAMLVerifier(enterpriseauth.UnconfiguredSAMLVerifier{})
	if override, ok := deps.ModuleOverrides[EnterpriseSAMLVerifierOverrideKey]; ok {
		if verifier, ok := override.(enterpriseSAMLVerifier); ok {
			samlVerifier = verifier
		} else {
			return nil, fmt.Errorf("auth enterprise SAML verifier override has type %T", override)
		}
	}

	return &Service{
		store:        authn.NewStore(deps.PostgresHandle()),
		revocations:  deps.WSHub,
		keys:         keys,
		cursorCodec:  cursorCodec,
		now:          now,
		profiles:     httpapi.ResolveExtensionProfiles(deps.ExtensionProfiles),
		oidcVerifier: oidcVerifier,
		samlVerifier: samlVerifier,
	}, nil
}
