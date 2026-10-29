package httpapi

import (
	"context"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *platformhttpapi.APIError) {
	return s.requireIncidentRole(ctx, incidentID, userID, admission.RolesMember, "")
}

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
	grant, err := s.incidentAdmission.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	return incidentAdmissionResult(grant, err, requiredRole)
}

func incidentAdmissionResult(grant admission.Grant, err error, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, incidentNotFoundError()
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, authorizationDeniedError(requiredRole)
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *platformhttpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = platformhttpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
