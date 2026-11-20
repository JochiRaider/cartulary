package collaboration

import (
	"errors"
	"net/http"
	"time"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type routeService struct {
	incidentAccess *admission.Checker
	authenticator  httpapi.Authenticator
	hub            *hub
	replay         *privatestream.PostgresStream
	acceptSocket   protocol.AcceptSocket
	checkOrigin    protocol.CheckOrigin
	serviceVersion string
	workspaces     claimedWorkspaces
	now            func() time.Time
}

func RegisterRoutes(runtime *Runtime) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newRouteService(deps, runtime)
		if err != nil {
			return err
		}
		// Route resolution, REQ-01-708 step 1, belongs to the platform: any
		// method but GET receives 405 before the handler runs. The registry
		// records the route for the catalog test.
		return httpapi.BindSocketRoute(mux, deps, "/ws/v1/incidents/{incident_id}", service.handleIncidentSocket)
	}
}

func newRouteService(deps httpapi.DependencySet, runtime *Runtime) (*routeService, error) {
	if runtime == nil || runtime.hub == nil {
		return nil, errors.New("collaboration runtime dependency is required")
	}
	// The socket authenticates through the central authenticator, which
	// publishes a session it finds expired to that session's sockets (Core 01
	// REQ-01-708, REQ-01-277). The socket re-reads its session through it at
	// the session's expiry point, too, and revokes nothing itself.
	if deps.Authenticator == nil {
		return nil, errors.New("collaboration socket requires the request authenticator")
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &routeService{
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authenticator:  deps.Authenticator,
		hub:            runtime.hub,
		replay:         runtime.store,
		acceptSocket:   runtime.routes.acceptSocket,
		checkOrigin:    runtime.routes.checkOrigin,
		serviceVersion: runtime.routes.serviceVersion,
		workspaces:     runtime.routes.workspaces,
		now:            now,
	}, nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}
