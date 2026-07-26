package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
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

const (
	ProfileID                                     = "enterprise_authentication"
	EnterpriseOIDCRouteContributionID             = "enterprise_authentication.auth_oidc_route"
	EnterpriseProvidersRouteContributionID        = "enterprise_authentication.auth_providers_route"
	EnterpriseSAMLRouteContributionID             = "enterprise_authentication.auth_saml_route"
	EnterpriseUserAuthBindingsRouteContributionID = "enterprise_authentication.user_auth_bindings_route"
)

var enterpriseRouteContributionIDs = []string{
	EnterpriseOIDCRouteContributionID,
	EnterpriseProvidersRouteContributionID,
	EnterpriseSAMLRouteContributionID,
	EnterpriseUserAuthBindingsRouteContributionID,
}

func EnterpriseRouteContributionIDs() []string {
	return append([]string(nil), enterpriseRouteContributionIDs...)
}

type Service struct {
	loginStore              localLoginStore
	sessionStore            sessionStore
	sessionMembershipReader sessionMembershipSummaryReader
	credentialStore         credentialLifecycleStore
	accountStore            accountStore
	userAdminStore          userAdminStore
	deploymentAuditReader   deploymentAuditReader
	enterpriseStore         enterpriseAuthStore
	revocations             sessionRevocationPublisher
	keys                    authn.MasterKeys
	cursorCodec             *pagination.Codec
	env                     map[string]string
	publicOrigin            string
	now                     func() time.Time
	enterpriseAdmitted      bool
	oidcVerifier            enterpriseOIDCVerifier
	samlVerifier            enterpriseSAMLVerifier
	beginRedirect           enterpriseBeginRedirectBuilder
}

type authStore interface {
	localLoginStore
	sessionStore
	sessionMembershipSummaryReader
	credentialLifecycleStore
	accountStore
	userAdminStore
	deploymentAuditReader
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

type deploymentAuditReader interface {
	ListAdministrativeAuditEvents(context.Context, administrativeaudit.ListFilter) ([]administrativeaudit.Record, error)
}

type enterpriseAuthStore interface {
	ListEnterpriseAuthProviders(context.Context) ([]authn.EnterpriseAuthProviderRecord, error)
	GetEnterpriseAuthProviderByKey(context.Context, string) (authn.EnterpriseAuthProviderRecord, error)
	CreateEnterpriseAuthTransaction(context.Context, authn.EnterpriseAuthProviderRecord, string, *string, *string, []byte, []byte, []byte, *string, *string, []byte, time.Time) (authn.EnterpriseAuthTransactionRecord, error)
	GetOIDCEnterpriseAuthTransactionForCallback(context.Context, string, string, []byte, time.Time) (authn.EnterpriseAuthTransactionRecord, error)
	GetSAMLEnterpriseAuthTransactionForACS(context.Context, string, string, time.Time) (authn.EnterpriseAuthTransactionRecord, error)
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

type RouteOption func(*routeOptions)

type routeOptions struct {
	enterpriseAuthBindings bool
	publicOrigin           string
}

func WithEnterpriseAuthBindings() RouteOption {
	return func(options *routeOptions) {
		options.enterpriseAuthBindings = true
	}
}

func WithPublicOrigin(publicOrigin string) RouteOption {
	return func(options *routeOptions) {
		options.publicOrigin = publicOrigin
	}
}

func RegisterRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		settings := routeOptions{}
		for _, option := range options {
			if option != nil {
				option(&settings)
			}
		}
		service, err := newService(deps, settings.enterpriseAuthBindings, settings.publicOrigin)
		if err != nil {
			return err
		}
		handlers := map[string]http.HandlerFunc{
			"beginTOTPEnrollment":             service.handleTOTPBegin,
			"changeCurrentPassword":           service.handlePasswordChange,
			"completeTOTPEnrollment":          service.handleTOTPComplete,
			"createDeploymentUser":            service.handleUsersCollection,
			"getCredentialState":              service.handleCredentialState,
			"getCurrentAccountPreferences":    service.handleAccountPreferences,
			"getCurrentAccountProfile":        service.handleAccountProfile,
			"getCurrentSession":               service.handleSession,
			"getDeploymentUser":               service.handleUsersMember,
			"listAdministrativeAuditEvents":   service.handleAdministrativeAuditEvents,
			"listDeploymentUsers":             service.handleUsersCollection,
			"loginLocalUser":                  service.handleLogin,
			"logoutCurrentSession":            service.handleLogout,
			"patchCurrentAccountProfile":      service.handleAccountProfile,
			"patchDeploymentUser":             service.handleUsersMember,
			"putCurrentAccountPreferences":    service.handleAccountPreferences,
			"resetDeploymentUserPassword":     service.handleUsersMember,
			"resetDeploymentUserTOTP":         service.handleUsersMember,
			"revokeAllDeploymentUserSessions": service.handleUsersMember,
		}
		if settings.enterpriseAuthBindings {
			handlers["createEnterpriseAuthBinding"] = service.handleUsersMember
			handlers["retireEnterpriseAuthBinding"] = service.handleUsersMember
			handlers["rotateEnterpriseAuthBinding"] = service.handleUsersMember
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.auth", handlers)
	}
}

func RegisterEnterpriseRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		settings := routeOptions{}
		for _, option := range options {
			if option != nil {
				option(&settings)
			}
		}
		service, err := newService(deps, true, settings.publicOrigin)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.auth", map[string]http.HandlerFunc{
			"beginEnterpriseAuth":         service.handleEnterpriseProviders,
			"completeEnterpriseOIDC":      service.handleEnterpriseOIDC,
			"completeEnterpriseSAML":      service.handleEnterpriseSAML,
			"finishEnterpriseSAML":        service.handleEnterpriseSAML,
			"listEnterpriseAuthProviders": service.handleEnterpriseProviders,
		})
	}
}

func RegisterTestRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		settings := routeOptions{}
		for _, option := range options {
			if option != nil {
				option(&settings)
			}
		}
		service, err := newService(deps, false, settings.publicOrigin)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/test/auth/touch", guard.Protect(service.handleTouch))
		return nil
	}
}

func newService(deps httpapi.DependencySet, enterpriseAdmitted bool, publicOrigin string) (*Service, error) {
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
	if enterpriseAdmitted {
		oidcVerifier = enterpriseOIDCVerifier(enterpriseauth.ProductionOIDCVerifier{})
	}
	beginRedirect := enterpriseBeginRedirectBuilder(enterpriseauth.BuildBeginRedirect)
	if override, ok := deps.ModuleOverrides[EnterpriseOIDCVerifierOverrideKey]; ok && enterpriseAdmitted {
		if verifier, ok := override.(enterpriseOIDCVerifier); ok {
			oidcVerifier = verifier
			beginRedirect = deterministicEnterpriseBeginRedirect
		} else {
			return nil, fmt.Errorf("auth enterprise OIDC verifier override has type %T", override)
		}
	}
	samlVerifier := enterpriseSAMLVerifier(enterpriseauth.UnconfiguredSAMLVerifier{})
	if enterpriseAdmitted {
		samlVerifier = enterpriseSAMLVerifier(enterpriseauth.ProductionSAMLVerifier{})
	}
	if override, ok := deps.ModuleOverrides[EnterpriseSAMLVerifierOverrideKey]; ok && enterpriseAdmitted {
		if verifier, ok := override.(enterpriseSAMLVerifier); ok {
			samlVerifier = verifier
			beginRedirect = deterministicEnterpriseBeginRedirect
		} else {
			return nil, fmt.Errorf("auth enterprise SAML verifier override has type %T", override)
		}
	}

	backingStore := authn.NewStore(deps.PostgresHandle())
	return &Service{
		loginStore:              backingStore,
		sessionStore:            backingStore,
		sessionMembershipReader: backingStore,
		credentialStore:         backingStore,
		accountStore:            backingStore,
		userAdminStore:          backingStore,
		deploymentAuditReader:   backingStore,
		enterpriseStore:         backingStore,
		revocations:             deps.WSHub,
		keys:                    keys,
		cursorCodec:             cursorCodec,
		env:                     deps.Env,
		publicOrigin:            publicOrigin,
		now:                     now,
		enterpriseAdmitted:      enterpriseAdmitted,
		oidcVerifier:            oidcVerifier,
		samlVerifier:            samlVerifier,
		beginRedirect:           beginRedirect,
	}, nil
}
