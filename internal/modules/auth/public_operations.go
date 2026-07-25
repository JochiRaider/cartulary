package auth

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		publicOperation(http.MethodGet, "/api/v1/account/preferences", "getCurrentAccountPreferences", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodPut, "/api/v1/account/preferences", "putCurrentAccountPreferences", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodGet, "/api/v1/account/profile", "getCurrentAccountProfile", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodPatch, "/api/v1/account/profile", "patchCurrentAccountProfile", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodGet, "/api/v1/administrative-audit-events", "listAdministrativeAuditEvents", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodGet, "/api/v1/auth/credential-state", "getCredentialState", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/login", "loginLocalUser", httpapi.PublicAuthenticationPublic, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/logout", "logoutCurrentSession", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/mfa/totp/begin", "beginTOTPEnrollment", httpapi.PublicAuthenticationSessionOrBootstrap, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/mfa/totp/complete", "completeTOTPEnrollment", httpapi.PublicAuthenticationSessionOrBootstrap, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/password/change", "changeCurrentPassword", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodGet, "/api/v1/auth/session", "getCurrentSession", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodGet, "/api/v1/users", "listDeploymentUsers", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/users", "createDeploymentUser", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		publicOperation(http.MethodGet, "/api/v1/users/{user_id}", "getDeploymentUser", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		publicOperation(http.MethodPatch, "/api/v1/users/{user_id}", "patchDeploymentUser", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/users/{user_id}/mfa/totp/reset", "resetDeploymentUserTOTP", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/users/{user_id}/password/reset", "resetDeploymentUserPassword", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/users/{user_id}/sessions/revoke-all", "revokeAllDeploymentUserSessions", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}

func EnterprisePublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		publicOperation(http.MethodGet, "/api/v1/auth/oidc/{provider_key}/callback", "completeEnterpriseOIDC", httpapi.PublicAuthenticationPublic, false, http.StatusSeeOther),
		publicOperation(http.MethodGet, "/api/v1/auth/providers", "listEnterpriseAuthProviders", httpapi.PublicAuthenticationPublic, false, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/providers/{provider_key}/begin", "beginEnterpriseAuth", httpapi.PublicAuthenticationPublic, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/auth/saml/{provider_key}/acs", "completeEnterpriseSAML", httpapi.PublicAuthenticationPublic, true, http.StatusSeeOther),
		publicOperation(http.MethodPost, "/api/v1/users/{user_id}/auth-bindings", "createEnterpriseAuthBinding", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		publicOperation(http.MethodDelete, "/api/v1/users/{user_id}/auth-bindings/{auth_binding_id}", "retireEnterpriseAuthBinding", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		publicOperation(http.MethodPost, "/api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate", "rotateEnterpriseAuthBinding", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}

func publicOperation(
	method string,
	pathTemplate string,
	operationID string,
	authentication httpapi.PublicAuthentication,
	stateChanging bool,
	successStatus int,
) httpapi.PublicOperation {
	return httpapi.PublicOperation{
		OwnerID:        "module.auth",
		Method:         method,
		PathTemplate:   pathTemplate,
		OperationID:    operationID,
		Authentication: authentication,
		StateChanging:  stateChanging,
		SuccessStatus:  successStatus,
	}
}
