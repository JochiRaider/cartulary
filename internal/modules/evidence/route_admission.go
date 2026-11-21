package evidence

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	admissionpkg "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// routeAdmission owns incident membership/role checks and session sliding for
// Evidence transport. Central admission authenticates by each operation's
// declared security and admits its path parameters before a handler runs
// (Core 01 REQ-01-707 steps 1 to 3).
type routeAdmission struct {
	incidents *admissionpkg.Checker
	auth      routeSessionStore
	now       func() time.Time
}

// routeSessionStore is the session surface Evidence transport slides against;
// the production value is the shared authn store.
type routeSessionStore interface {
	httpauth.SessionSlider
}

// admittedPrincipal returns the principal central admission established
// (REQ-01-707 step 2); its absence is a binding error.
func admittedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, ok := httpauth.AdmittedPrincipal(r)
	if !ok {
		return httpauth.Principal{}, httpapi.InternalAPIError(errors.New("evidence route was not admitted centrally"))
	}
	return principal, nil
}

// admittedPathUUID returns a uuid path parameter central admission admitted
// (REQ-01-707 step 3); its absence is a binding error.
func admittedPathUUID(r *http.Request, name string) (uuid.UUID, *httpapi.APIError) {
	value, ok := httpapi.PathUUID(r, name)
	if !ok {
		return uuid.UUID{}, httpapi.InternalAPIError(errors.New("evidence route did not admit " + name))
	}
	return value, nil
}

// admittedPathString returns a capability token central admission admitted.
// Only its syntax is checked there; whether it names a live capability stays
// route handling (REQ-01-707).
func admittedPathString(r *http.Request, name string) (string, *httpapi.APIError) {
	value, ok := httpapi.PathString(r, name)
	if !ok {
		return "", httpapi.InternalAPIError(errors.New("evidence route did not admit " + name))
	}
	return value, nil
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
	return evidenceVisibilityResult(grant, err)
}

// evidenceVisibilityResult classifies a member-visibility read. Every member
// role is admitted, so the read has no role refusal: an incident the caller
// cannot see is not found, and any other failure is internal.
func evidenceVisibilityResult(grant admissionpkg.Grant, err error) (admissionpkg.Grant, *httpapi.APIError) {
	switch {
	case admissionpkg.IsDenied(err, admissionpkg.DenialNotVisible):
		return admissionpkg.Grant{}, httpapi.NotFound("incident_not_found")
	case err != nil:
		return admissionpkg.Grant{}, httpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}

// concealedIncident decides visibility like visibleIncident, except that an
// incident the caller cannot see receives the addressed resource's own
// not-found response (REQ-01-704). Every other failure keeps its
// classification, so an admission read failure stays internal_error
// (REQ-01-706).
func (admission routeAdmission) concealedIncident(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	notFound func() *httpapi.APIError,
) (admissionpkg.Grant, *httpapi.APIError) {
	return admission.concealedRole(ctx, incidentID, userID, admissionpkg.RolesMember, "member", notFound)
}

// concealedRole decides the caller's visibility of the incident and then the
// route's role from one admission read, so a membership removed while the
// request runs can never surface between the two decisions. An incident the
// caller cannot see receives the addressed resource's own not-found response
// (REQ-01-704); a visible caller without one of roles receives
// authorization_denied; any other failure stays internal_error (REQ-01-706).
func (admission routeAdmission) concealedRole(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	roles admissionpkg.RoleSet,
	requiredRole string,
	notFound func() *httpapi.APIError,
) (admissionpkg.Grant, *httpapi.APIError) {
	grant, err := admission.incidents.Check(ctx, incidentID, userID, admissionpkg.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admissionpkg.LifecycleAny,
	})
	if admissionpkg.IsDenied(err, admissionpkg.DenialNotVisible) {
		return admissionpkg.Grant{}, notFound()
	}
	return evidenceAdmissionResult(grant, err, requiredRole)
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
		return admissionpkg.Grant{}, httpapi.NotFound("incident_not_found")
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
) error {
	return httpauth.SlideSessionIfNeeded(ctx, admission.auth, principal, admission.now)
}
