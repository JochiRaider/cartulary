package routetest

import (
	"net/http"
	"strings"
	"testing"
)

type RouteTransport string

const (
	RouteTransportHTTP      RouteTransport = "http"
	RouteTransportWebSocket RouteTransport = "websocket"
)

type RouteHarnessClass string

const (
	RouteHarnessSurfaceEnvelope     RouteHarnessClass = "surface_envelope"
	RouteHarnessBootstrapBoundary   RouteHarnessClass = "bootstrap_boundary"
	RouteHarnessCSRF                RouteHarnessClass = "csrf"
	RouteHarnessReplayStoredPayload RouteHarnessClass = "replay_stored_payload"
	RouteHarnessMutationAudit       RouteHarnessClass = "mutation_audit"
	RouteHarnessSessionRevocation   RouteHarnessClass = "session_revocation"
	RouteHarnessAuthorization       RouteHarnessClass = "authorization_rederivation"
	RouteHarnessRequestContracts    RouteHarnessClass = "request_contracts"
)

var routeHarnessClasses = []RouteHarnessClass{
	RouteHarnessSurfaceEnvelope,
	RouteHarnessBootstrapBoundary,
	RouteHarnessCSRF,
	RouteHarnessReplayStoredPayload,
	RouteHarnessMutationAudit,
	RouteHarnessSessionRevocation,
	RouteHarnessAuthorization,
	RouteHarnessRequestContracts,
}

type RouteHarnessRequirement string

const (
	RouteHarnessRequired      RouteHarnessRequirement = "required"
	RouteHarnessNotApplicable RouteHarnessRequirement = "n/a"
)

type RouteID string

const (
	RouteLogin               RouteID = "login"
	RouteSession             RouteID = "session"
	RouteLogout              RouteID = "logout"
	RouteCredentialState     RouteID = "credential_state"
	RoutePasswordChange      RouteID = "password_change"
	RouteTOTPBegin           RouteID = "totp_begin"
	RouteTOTPComplete        RouteID = "totp_complete"
	RouteAdministrativeAudit RouteID = "administrative_audit"
	RouteUsersList           RouteID = "users_list"
	RouteUsersCreate         RouteID = "users_create"
	RouteUsersGet            RouteID = "users_get"
	RouteUsersPatch          RouteID = "users_patch"
	RouteUsersPasswordReset  RouteID = "users_password_reset"
	RouteUsersTOTPReset      RouteID = "users_totp_reset"
	RouteUsersRevokeAll      RouteID = "users_revoke_all"
	RouteSessionLifecycleWS  RouteID = "session_lifecycle_ws"
)

type RouteAuthorizationChange string

const (
	RouteAuthorizationNotApplicable         RouteAuthorizationChange = "n/a"
	RouteAuthorizationDemoteDeploymentAdmin RouteAuthorizationChange = "demote_deployment_admin"
)

type RouteInventoryFixture struct {
	UserID     string
	IncidentID string
}

type RouteClosedVocabularyContract struct {
	Field      string
	ReasonCode string
}

type RouteWritableStringContract struct {
	Field string
}

type RouteRequestContracts struct {
	ClosedVocabulary []RouteClosedVocabularyContract
	WritableStrings  []RouteWritableStringContract
}

func (contracts RouteRequestContracts) Empty() bool {
	return len(contracts.ClosedVocabulary) == 0 && len(contracts.WritableStrings) == 0
}

type RouteInventoryEntry struct {
	ID            RouteID
	Transport     RouteTransport
	Method        string
	Template      string
	SuccessStatus int
	RequiresCSRF  bool

	AuthorizationChange RouteAuthorizationChange
	AuthorizationStatus int
	AuthorizationCode   string

	RequestContracts    RouteRequestContracts
	HarnessRequirements map[RouteHarnessClass]RouteHarnessRequirement
}

func BuildRoutePath(template string, fixture RouteInventoryFixture) string {
	replacer := strings.NewReplacer(
		"{user_id}", fixture.UserID,
		"{incident_id}", fixture.IncidentID,
	)
	return replacer.Replace(template)
}

func ValidateRouteInventory(t testing.TB, routes []RouteInventoryEntry) {
	t.Helper()

	for _, route := range routes {
		if route.Method == "" {
			t.Fatalf("authentication route %s missing method", route.ID)
		}
		if route.Template == "" {
			t.Fatalf("authentication route %s missing template", route.ID)
		}
		if route.SuccessStatus == 0 {
			t.Fatalf("authentication route %s missing success status", route.ID)
		}
		if route.HarnessRequirements == nil {
			t.Fatalf("authentication route %s missing harness requirements", route.ID)
		}
		for _, harness := range routeHarnessClasses {
			requirement, ok := route.HarnessRequirements[harness]
			if !ok {
				t.Fatalf("authentication route %s missing harness requirement for %s", route.ID, harness)
			}
			if requirement != RouteHarnessRequired && requirement != RouteHarnessNotApplicable {
				t.Fatalf("authentication route %s has invalid harness requirement %q for %s", route.ID, requirement, harness)
			}
		}
		if route.Transport == RouteTransportWebSocket {
			if route.HarnessRequirements[RouteHarnessSurfaceEnvelope] != RouteHarnessNotApplicable {
				t.Fatalf("authentication websocket route %s must not require the surface envelope harness", route.ID)
			}
		}
		if route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessRequired {
			t.Fatalf("authentication route %s requires csrf but matrix marks %s", route.ID, route.HarnessRequirements[RouteHarnessCSRF])
		}
		if !route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessNotApplicable {
			t.Fatalf("authentication route %s does not require csrf but matrix marks %s", route.ID, route.HarnessRequirements[RouteHarnessCSRF])
		}
		switch route.AuthorizationChange {
		case RouteAuthorizationNotApplicable:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessNotApplicable {
				t.Fatalf("authentication route %s marks authorization n/a but matrix requires it", route.ID)
			}
		case RouteAuthorizationDemoteDeploymentAdmin:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessRequired {
				t.Fatalf("authentication route %s must require authorization re-derivation", route.ID)
			}
			if route.AuthorizationStatus == 0 || route.AuthorizationCode == "" {
				t.Fatalf("authentication route %s missing authorization expectation", route.ID)
			}
		default:
			t.Fatalf("authentication route %s has invalid authorization change %q", route.ID, route.AuthorizationChange)
		}
		if route.RequestContracts.Empty() {
			if route.HarnessRequirements[RouteHarnessRequestContracts] != RouteHarnessNotApplicable {
				t.Fatalf("authentication route %s has no request contracts but matrix requires them", route.ID)
			}
		} else {
			if route.HarnessRequirements[RouteHarnessRequestContracts] != RouteHarnessRequired {
				t.Fatalf("authentication route %s has request contracts but matrix does not require them", route.ID)
			}
			for _, contract := range route.RequestContracts.ClosedVocabulary {
				if contract.Field == "" {
					t.Fatalf("authentication route %s has a closed-vocabulary contract with no field", route.ID)
				}
			}
			for _, contract := range route.RequestContracts.WritableStrings {
				if contract.Field == "" {
					t.Fatalf("authentication route %s has a writable-string contract with no field", route.ID)
				}
			}
		}
	}
}

func RoutesForHarness(t testing.TB, routes []RouteInventoryEntry, harness RouteHarnessClass) []RouteInventoryEntry {
	t.Helper()

	ValidateRouteInventory(t, routes)

	filtered := make([]RouteInventoryEntry, 0, len(routes))
	for _, route := range routes {
		if route.HarnessRequirements[harness] == RouteHarnessRequired {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func PublicRouteInventory() []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			ID:                  RouteLogin,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/auth/login",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			RequestContracts: RouteRequestContracts{
				ClosedVocabulary: []RouteClosedVocabularyContract{
					{Field: "second_factor.kind"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteSession,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodGet,
			Template:            "/api/v1/auth/session",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
			),
		},
		{
			ID:                  RouteLogout,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/auth/logout",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessSessionRevocation,
			),
		},
		{
			ID:                  RouteCredentialState,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodGet,
			Template:            "/api/v1/auth/credential-state",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
			),
		},
		{
			ID:                  RoutePasswordChange,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/auth/password/change",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			RequestContracts: RouteRequestContracts{
				ClosedVocabulary: []RouteClosedVocabularyContract{
					{Field: "second_factor.kind"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessSessionRevocation,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteTOTPBegin,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/auth/mfa/totp/begin",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			RequestContracts: RouteRequestContracts{
				ClosedVocabulary: []RouteClosedVocabularyContract{
					{Field: "second_factor.kind"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteTOTPComplete,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/auth/mfa/totp/complete",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
			),
		},
		{
			ID:                  RouteAdministrativeAudit,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodGet,
			Template:            "/api/v1/administrative-audit-events",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessAuthorization,
			),
		},
		{
			ID:                  RouteUsersList,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodGet,
			Template:            "/api/v1/users",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessAuthorization,
			),
		},
		{
			ID:                  RouteUsersCreate,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/users",
			SuccessStatus:       http.StatusCreated,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			RequestContracts: RouteRequestContracts{
				ClosedVocabulary: []RouteClosedVocabularyContract{
					{Field: "auth_kind", ReasonCode: "invalid_auth_kind"},
				},
				WritableStrings: []RouteWritableStringContract{
					{Field: "email"},
					{Field: "display_name"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessMutationAudit,
				RouteHarnessAuthorization,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteUsersGet,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodGet,
			Template:            "/api/v1/users/{user_id}",
			SuccessStatus:       http.StatusOK,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessAuthorization,
			),
		},
		{
			ID:                  RouteUsersPatch,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPatch,
			Template:            "/api/v1/users/{user_id}",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			RequestContracts: RouteRequestContracts{
				WritableStrings: []RouteWritableStringContract{
					{Field: "email"},
					{Field: "display_name"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessMutationAudit,
				RouteHarnessAuthorization,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteUsersPasswordReset,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/users/{user_id}/password/reset",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			RequestContracts: RouteRequestContracts{
				WritableStrings: []RouteWritableStringContract{
					{Field: "reason"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessMutationAudit,
				RouteHarnessSessionRevocation,
				RouteHarnessAuthorization,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteUsersTOTPReset,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/users/{user_id}/mfa/totp/reset",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			RequestContracts: RouteRequestContracts{
				WritableStrings: []RouteWritableStringContract{
					{Field: "reason"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessMutationAudit,
				RouteHarnessSessionRevocation,
				RouteHarnessAuthorization,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteUsersRevokeAll,
			Transport:           RouteTransportHTTP,
			Method:              http.MethodPost,
			Template:            "/api/v1/users/{user_id}/sessions/revoke-all",
			SuccessStatus:       http.StatusOK,
			RequiresCSRF:        true,
			AuthorizationChange: RouteAuthorizationDemoteDeploymentAdmin,
			AuthorizationStatus: http.StatusForbidden,
			AuthorizationCode:   "authorization_denied",
			RequestContracts: RouteRequestContracts{
				WritableStrings: []RouteWritableStringContract{
					{Field: "reason"},
				},
			},
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessBootstrapBoundary,
				RouteHarnessCSRF,
				RouteHarnessReplayStoredPayload,
				RouteHarnessMutationAudit,
				RouteHarnessSessionRevocation,
				RouteHarnessAuthorization,
				RouteHarnessRequestContracts,
			),
		},
		{
			ID:                  RouteSessionLifecycleWS,
			Transport:           RouteTransportWebSocket,
			Method:              http.MethodGet,
			Template:            "/ws/v1/incidents/{incident_id}",
			SuccessStatus:       http.StatusSwitchingProtocols,
			AuthorizationChange: RouteAuthorizationNotApplicable,
			HarnessRequirements: routeHarnessRequirements(
				RouteHarnessBootstrapBoundary,
				RouteHarnessSessionRevocation,
			),
		},
	}
}

func routeHarnessRequirements(required ...RouteHarnessClass) map[RouteHarnessClass]RouteHarnessRequirement {
	requirements := make(map[RouteHarnessClass]RouteHarnessRequirement, len(routeHarnessClasses))
	for _, harness := range routeHarnessClasses {
		requirements[harness] = RouteHarnessNotApplicable
	}
	for _, harness := range required {
		requirements[harness] = RouteHarnessRequired
	}
	return requirements
}
