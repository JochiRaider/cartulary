package evidence

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	admissionpkg "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// routeAdmission owns authentication, CSRF/session admission, incident
// membership/role checks, and session sliding for Evidence transport.
type routeAdmission struct {
	incidents *admissionpkg.Checker
	auth      *authn.Store
	keys      authn.MasterKeys
	now       func() time.Time
}

func (admission routeAdmission) authenticate(
	request *http.Request,
	stateChanging bool,
) (httpauth.Principal, *httpapi.APIError) {
	return httpauth.AuthenticateRequest(request, httpauth.Options{
		Store:         admission.auth,
		Keys:          admission.keys,
		Now:           admission.now,
		StateChanging: stateChanging,
	})
}

func (admission routeAdmission) visibleIncident(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (admissionpkg.Grant, *httpapi.APIError) {
	grant, err := admission.incidents.Check(ctx, incidentID, userID, admissionpkg.Requirement{
		AllowedRoles: admissionpkg.RolesMember,
		Lifecycle:    admissionpkg.LifecycleAny,
	})
	return evidenceAdmissionResult(grant, err, "member")
}

func (admission routeAdmission) requireRole(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	roles admissionpkg.RoleSet,
	requiredRole string,
) (admissionpkg.Grant, *httpapi.APIError) {
	grant, err := admission.incidents.Check(ctx, incidentID, userID, admissionpkg.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admissionpkg.LifecycleAny,
	})
	return evidenceAdmissionResult(grant, err, requiredRole)
}

func evidenceAdmissionResult(grant admissionpkg.Grant, err error, requiredRole string) (admissionpkg.Grant, *httpapi.APIError) {
	switch {
	case admissionpkg.IsDenied(err, admissionpkg.DenialNotVisible):
		return admissionpkg.Grant{}, incidentNotFoundError()
	case admissionpkg.IsDenied(err, admissionpkg.DenialInsufficientRole):
		return admissionpkg.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admissionpkg.Grant{}, httpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}

func (admission routeAdmission) slide(
	ctx context.Context,
	principal *httpauth.Principal,
	method string,
	path string,
) error {
	return httpauth.SlideSessionIfNeeded(ctx, admission.auth, principal, method, path, admission.now)
}
