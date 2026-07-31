package evidence

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// routeAdmission owns authentication, CSRF/session admission, incident
// membership/role checks, and session sliding for Evidence transport.
type routeAdmission struct {
	incidents incidents.Access
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

func (admission routeAdmission) requireMembership(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, admission.incidents, incidentID, userID)
}

func (admission routeAdmission) requireRole(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	roles ...string,
) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, admission.incidents, incidentID, userID, roles...)
}

func (admission routeAdmission) slide(
	ctx context.Context,
	principal *httpauth.Principal,
	method string,
	path string,
) error {
	return httpauth.SlideSessionIfNeeded(ctx, admission.auth, principal, method, path, admission.now)
}
