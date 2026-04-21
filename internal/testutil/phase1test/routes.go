package phase1test

import (
	"net/http"
	"strings"
)

type RouteTransport string

const (
	RouteTransportHTTP      RouteTransport = "http"
	RouteTransportWebSocket RouteTransport = "websocket"
)

type RouteCheck string

const (
	RouteCheckBootstrapBoundary RouteCheck = "bootstrap_boundary"
	RouteCheckCSRF              RouteCheck = "csrf"
	RouteCheckEnvelope          RouteCheck = "envelope"
	RouteCheckMutationAudit     RouteCheck = "mutation_audit"
	RouteCheckReplay            RouteCheck = "replay"
	RouteCheckSecretSafePayload RouteCheck = "secret_safe_payload"
	RouteCheckSessionRevocation RouteCheck = "session_revocation"
)

type RouteID string

const (
	RouteLogin              RouteID = "login"
	RouteSession            RouteID = "session"
	RouteLogout             RouteID = "logout"
	RouteCredentialState    RouteID = "credential_state"
	RoutePasswordChange     RouteID = "password_change"
	RouteTOTPBegin          RouteID = "totp_begin"
	RouteTOTPComplete       RouteID = "totp_complete"
	RouteUsersList          RouteID = "users_list"
	RouteUsersCreate        RouteID = "users_create"
	RouteUsersGet           RouteID = "users_get"
	RouteUsersPatch         RouteID = "users_patch"
	RouteUsersPasswordReset RouteID = "users_password_reset"
	RouteUsersTOTPReset     RouteID = "users_totp_reset"
	RouteUsersRevokeAll     RouteID = "users_revoke_all"
	RouteSessionLifecycleWS RouteID = "session_lifecycle_ws"
)

type RouteInventoryFixture struct {
	UserID string
}

type RouteInventoryEntry struct {
	ID            RouteID
	Transport     RouteTransport
	Method        string
	Template      string
	SuccessStatus int
	RequiresCSRF  bool
	Checks        []RouteCheck
}

func BuildRoutePath(template string, fixture RouteInventoryFixture) string {
	replacer := strings.NewReplacer(
		"{user_id}", fixture.UserID,
	)
	return replacer.Replace(template)
}

func PublicRouteInventory() []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			ID:            RouteLogin,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/auth/login",
			SuccessStatus: http.StatusOK,
			Checks:        []RouteCheck{RouteCheckEnvelope},
		},
		{
			ID:            RouteSession,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodGet,
			Template:      "/api/v1/auth/session",
			SuccessStatus: http.StatusOK,
			Checks:        []RouteCheck{RouteCheckBootstrapBoundary, RouteCheckEnvelope},
		},
		{
			ID:            RouteLogout,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/auth/logout",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckSessionRevocation,
			},
		},
		{
			ID:            RouteCredentialState,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodGet,
			Template:      "/api/v1/auth/credential-state",
			SuccessStatus: http.StatusOK,
			Checks:        []RouteCheck{RouteCheckBootstrapBoundary, RouteCheckEnvelope},
		},
		{
			ID:            RoutePasswordChange,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/auth/password/change",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
				RouteCheckSessionRevocation,
			},
		},
		{
			ID:            RouteTOTPBegin,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/auth/mfa/totp/begin",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
			},
		},
		{
			ID:            RouteTOTPComplete,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/auth/mfa/totp/complete",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
			},
		},
		{
			ID:            RouteUsersList,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodGet,
			Template:      "/api/v1/users",
			SuccessStatus: http.StatusOK,
			Checks:        []RouteCheck{RouteCheckBootstrapBoundary, RouteCheckEnvelope},
		},
		{
			ID:            RouteUsersCreate,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/users",
			SuccessStatus: http.StatusCreated,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckMutationAudit,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
			},
		},
		{
			ID:            RouteUsersGet,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodGet,
			Template:      "/api/v1/users/{user_id}",
			SuccessStatus: http.StatusOK,
			Checks:        []RouteCheck{RouteCheckBootstrapBoundary, RouteCheckEnvelope},
		},
		{
			ID:            RouteUsersPatch,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPatch,
			Template:      "/api/v1/users/{user_id}",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckMutationAudit,
			},
		},
		{
			ID:            RouteUsersPasswordReset,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/users/{user_id}/password/reset",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckMutationAudit,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
				RouteCheckSessionRevocation,
			},
		},
		{
			ID:            RouteUsersTOTPReset,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/users/{user_id}/mfa/totp/reset",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckMutationAudit,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
				RouteCheckSessionRevocation,
			},
		},
		{
			ID:            RouteUsersRevokeAll,
			Transport:     RouteTransportHTTP,
			Method:        http.MethodPost,
			Template:      "/api/v1/users/{user_id}/sessions/revoke-all",
			SuccessStatus: http.StatusOK,
			RequiresCSRF:  true,
			Checks: []RouteCheck{
				RouteCheckBootstrapBoundary,
				RouteCheckCSRF,
				RouteCheckEnvelope,
				RouteCheckMutationAudit,
				RouteCheckReplay,
				RouteCheckSecretSafePayload,
				RouteCheckSessionRevocation,
			},
		},
		{
			ID:        RouteSessionLifecycleWS,
			Transport: RouteTransportWebSocket,
			Method:    http.MethodGet,
			Template:  "/ws/v1/test/session-lifecycle",
			Checks:    []RouteCheck{RouteCheckBootstrapBoundary, RouteCheckSessionRevocation},
		},
	}
}
