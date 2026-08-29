package collaboration

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type routeService struct {
	incidentAccess *admission.Checker
	authStore      *authn.Store
	hub            *hub
	replay         *privatestream.PostgresStream
	keys           authn.MasterKeys
	acceptSocket   protocol.AcceptSocket
	checkOrigin    protocol.CheckBrowserOrigin
	serviceVersion string
	now            func() time.Time
}

func RegisterRoutes(runtime *Runtime) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newRouteService(deps, runtime)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /ws/v1/incidents/{incident_id}", service.handleIncidentSocket)
		return nil
	}
}

func newRouteService(deps httpapi.DependencySet, runtime *Runtime) (*routeService, error) {
	if runtime == nil || runtime.hub == nil {
		return nil, errors.New("collaboration runtime dependency is required")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &routeService{
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		hub:            runtime.hub,
		replay:         runtime.store,
		keys:           keys,
		acceptSocket:   runtime.routes.acceptSocket,
		checkOrigin:    runtime.routes.checkOrigin,
		serviceVersion: runtime.routes.serviceVersion,
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

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	value, err := uuid.Parse(raw)
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}
