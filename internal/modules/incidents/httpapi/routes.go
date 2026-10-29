package httpapi

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type Application interface {
	ListVisibleIncidents(context.Context, uuid.UUID, incidents.IncidentListPageRequest) ([]incidents.IncidentRecord, error)
	GetVisibleIncident(context.Context, uuid.UUID, uuid.UUID) (incidents.IncidentRecord, error)
	CreateIncident(context.Context, authn.UserRecord, incidents.IncidentCreateAdmission, string) (incidents.CreateIncidentResult, error)
	UpdateIncident(context.Context, authn.UserRecord, uuid.UUID, incidents.IncidentPatchAdmission, string) (incidents.IncidentRecord, bool, error)
	ListMemberships(context.Context, uuid.UUID, incidents.MembershipListPageRequest) ([]incidents.MembershipRecord, error)
	CreateMembership(context.Context, authn.UserRecord, uuid.UUID, incidents.MembershipCreateAdmission, string) (incidents.MembershipCreateResult, error)
	UpdateMembership(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, incidents.MembershipPatchAdmission, string) (incidents.MembershipRecord, bool, error)
}

type AdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

type TerminalMutationCoordinator interface {
	CoordinateIncidentLifecycle(context.Context, authn.UserRecord, uuid.UUID, incidents.IncidentLifecycleAdmission, string) (incidents.IncidentLifecycleResult, error)
	CoordinateMembershipDeletion(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, incidents.MembershipDeleteAdmission, string) (incidents.MembershipDeleteResult, error)
}

type Dependencies struct {
	Application                 Application
	AdmissionChecker            AdmissionChecker
	TerminalMutationCoordinator TerminalMutationCoordinator
}

type service struct {
	application       Application
	membershipAudit   *membershipAuditReader
	terminalMutations TerminalMutationCoordinator
	incidentAdmission AdmissionChecker
	authStore         *authn.Store
	keys              authn.MasterKeys
	cursorCodec       *pagination.Codec
	now               func() time.Time
}

func RegisterRoutes(dependencies Dependencies) platformhttpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, dependencies)
		if err != nil {
			return err
		}
		return platformhttpapi.BindOwnerRoutes(mux, deps, "module.incidents", map[string]http.HandlerFunc{
			"closeIncident":                     service.handleCloseIncident,
			"createIncident":                    service.handleCreateIncident,
			"createIncidentMembership":          service.handleCreateIncidentMembership,
			"deleteIncidentMembership":          service.handleDeleteIncidentMembership,
			"getIncident":                       service.handleGetIncident,
			"listIncidentMembershipAuditEvents": service.handleMembershipAuditEvents,
			"listIncidentMemberships":           service.handleListIncidentMemberships,
			"listVisibleIncidents":              service.handleListVisibleIncidents,
			"patchIncident":                     service.handlePatchIncident,
			"patchIncidentMembership":           service.handlePatchIncidentMembership,
			"reopenIncident":                    service.handleReopenIncident,
		})
	}
}

func newService(deps platformhttpapi.DependencySet, dependencies Dependencies) (*service, error) {
	if isNilRouteDependency(dependencies.Application) {
		return nil, errors.New("incidents: application is required for route registration")
	}
	if isNilRouteDependency(dependencies.AdmissionChecker) {
		return nil, errors.New("incidents: admission checker is required for route registration")
	}
	if isNilRouteDependency(dependencies.TerminalMutationCoordinator) {
		return nil, errors.New("incidents: terminal mutation coordinator is required for route registration")
	}
	membershipAudit, err := newMembershipAuditReader(deps.PostgresHandle())
	if err != nil {
		return nil, err
	}
	if deps.Now == nil {
		return nil, errors.New("incidents: clock is required for route registration")
	}
	if deps.CursorCodec == nil {
		return nil, errors.New("incidents: cursor codec is required for route registration")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	return &service{
		application:       dependencies.Application,
		membershipAudit:   membershipAudit,
		terminalMutations: dependencies.TerminalMutationCoordinator,
		incidentAdmission: dependencies.AdmissionChecker,
		authStore:         authn.NewStore(deps.PostgresHandle()),
		keys:              keys,
		cursorCodec:       deps.CursorCodec,
		now:               deps.Now,
	}, nil
}

func isNilRouteDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// routeUUID validates the identifier selected by the registered operation.
// Decoded separators must never be reinterpreted as another operation's path.
func routeUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		http.NotFound(w, r)
		return uuid.Nil, false
	}
	return id, true
}
