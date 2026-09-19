package harnesscontrol

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const testNetworkFlowAuthTransitionSchemaID = "cartulary.test.network_flow_auth_transition_control.v2"

const (
	NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization         = "network_flow.route.before_authorization"
	NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck = "network_flow.cursor.before_authorization_recheck"
)

const (
	NetworkFlowAuthTransitionKindIncidentMembershipRevoked   = "incident_membership_revoked"
	NetworkFlowAuthTransitionKindIncidentMembershipRestored  = "incident_membership_restored"
	NetworkFlowAuthTransitionKindIncidentDeleted             = "incident_deleted"
	NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted = "network_flow_table_soft_deleted"
	NetworkFlowAuthTransitionKindNetworkFlowTableRenamed     = "network_flow_table_renamed"
	NetworkFlowAuthTransitionKindSessionRevoked              = "session_revoked"
)

const (
	NetworkFlowAuthResourceIncident                = "incident"
	NetworkFlowAuthResourceNetworkFlowTable        = "network_flow_table"
	NetworkFlowAuthResourceNetworkFlowCursor       = "network_flow_cursor"
	NetworkFlowAuthResourceNetworkFlowGraph        = "network_flow_graph"
	NetworkFlowAuthResourceNetworkFlowContributors = "network_flow_contributors"
	NetworkFlowAuthResourceNetworkFlowWorkspace    = "network_flow_workspace"
)

var (
	networkFlowAuthTransitionBoundaries = map[string]struct{}{
		NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization:         {},
		NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck: {},
	}

	networkFlowAuthTransitionKinds = map[string]struct{}{
		NetworkFlowAuthTransitionKindIncidentMembershipRevoked:   {},
		NetworkFlowAuthTransitionKindIncidentMembershipRestored:  {},
		NetworkFlowAuthTransitionKindIncidentDeleted:             {},
		NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted: {},
		NetworkFlowAuthTransitionKindNetworkFlowTableRenamed:     {},
		NetworkFlowAuthTransitionKindSessionRevoked:              {},
	}

	networkFlowAuthResourceKinds = map[string]struct{}{
		NetworkFlowAuthResourceIncident:                {},
		NetworkFlowAuthResourceNetworkFlowTable:        {},
		NetworkFlowAuthResourceNetworkFlowCursor:       {},
		NetworkFlowAuthResourceNetworkFlowGraph:        {},
		NetworkFlowAuthResourceNetworkFlowContributors: {},
		NetworkFlowAuthResourceNetworkFlowWorkspace:    {},
	}
)

type NetworkFlowAuthTransitionRegistry struct {
	mu          sync.Mutex
	transitions map[string]NetworkFlowAuthTransition
	fixtures    map[string]string
}

type NetworkFlowAuthTransition struct {
	ID             string
	Boundary       string
	TransitionKind string
	ActorRef       string
	IncidentRef    string
	ResourceKind   string
	ResourceRef    string
	CorrelationKey string
}

type networkFlowAuthTransitionService struct {
	guard       httpapi.TestRouteGuard
	transitions *NetworkFlowAuthTransitionRegistry
}

type networkFlowAuthTransitionRequest struct {
	Boundary       string  `json:"boundary"`
	TransitionKind string  `json:"transition_kind"`
	ActorRef       string  `json:"actor_ref"`
	IncidentRef    string  `json:"incident_ref"`
	ResourceKind   string  `json:"resource_kind"`
	ResourceRef    string  `json:"resource_ref"`
	CorrelationKey *string `json:"correlation_key"`
	ConsumeOnce    bool    `json:"consume_once"`
}

type networkFlowAuthTransitionResult struct {
	SchemaID       string `json:"schema_id"`
	ControlID      string `json:"control_id"`
	Boundary       string `json:"boundary"`
	TransitionKind string `json:"transition_kind"`
	ActorRef       string `json:"actor_ref"`
	IncidentRef    string `json:"incident_ref"`
	ResourceKind   string `json:"resource_kind"`
	ResourceRef    string `json:"resource_ref"`
	CorrelationKey string `json:"correlation_key,omitempty"`
	ConsumeOnce    bool   `json:"consume_once"`
}

func NewNetworkFlowAuthTransitionRegistry() *NetworkFlowAuthTransitionRegistry {
	return &NetworkFlowAuthTransitionRegistry{transitions: map[string]NetworkFlowAuthTransition{}, fixtures: map[string]string{}}
}

func RegisterNetworkFlowAuthTransitionRoutes(transitions *NetworkFlowAuthTransitionRegistry) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		if transitions == nil {
			return fmt.Errorf("register network flow auth-transition route: transition registry is required")
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register network flow auth-transition route: %w", err)
		}
		service := &networkFlowAuthTransitionService{
			guard:       guard,
			transitions: transitions,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/network-flow-auth-transitions", service.handleArm)
		return nil
	}
}

func (r *NetworkFlowAuthTransitionRegistry) ConsumeNetworkFlowAuthTransition(boundary string, actorRef string, incidentRef string, resourceKind string, resourceRef string) (NetworkFlowAuthTransition, bool) {
	return r.ConsumeNetworkFlowAuthTransitionFor(boundary, actorRef, incidentRef, resourceKind, resourceRef, "")
}

func (r *NetworkFlowAuthTransitionRegistry) ConsumeNetworkFlowAuthTransitionFor(boundary string, actorRef string, incidentRef string, resourceKind string, resourceRef string, correlationKey string) (NetworkFlowAuthTransition, bool) {
	if r == nil {
		return NetworkFlowAuthTransition{}, false
	}
	key := networkFlowAuthTransitionKey(boundary, actorRef, incidentRef, resourceKind, resourceRef)
	correlationKey = strings.TrimSpace(correlationKey)
	r.mu.Lock()
	defer r.mu.Unlock()
	transition, ok := r.transitions[key]
	if !ok {
		return NetworkFlowAuthTransition{}, false
	}
	if transition.CorrelationKey != "" && transition.CorrelationKey != correlationKey {
		return NetworkFlowAuthTransition{}, false
	}
	delete(r.transitions, key)
	return transition, true
}

func (r *NetworkFlowAuthTransitionRegistry) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transitions = map[string]NetworkFlowAuthTransition{}
	r.fixtures = map[string]string{}
}

func (r *NetworkFlowAuthTransitionRegistry) arm(transition NetworkFlowAuthTransition) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.transitions == nil {
		r.transitions = map[string]NetworkFlowAuthTransition{}
	}
	key := networkFlowAuthTransitionKey(transition.Boundary, transition.ActorRef, transition.IncidentRef, transition.ResourceKind, transition.ResourceRef)
	if _, exists := r.transitions[key]; exists {
		return false
	}
	r.transitions[key] = transition
	return true
}

func networkFlowAuthTransitionKey(boundary string, actorRef string, incidentRef string, resourceKind string, resourceRef string) string {
	return strings.TrimSpace(boundary) + "\x00" + strings.TrimSpace(actorRef) + "\x00" + strings.TrimSpace(incidentRef) + "\x00" + strings.TrimSpace(resourceKind) + "\x00" + strings.TrimSpace(resourceRef)
}

func (s *networkFlowAuthTransitionService) handleArm(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	request, err := decodeNetworkFlowAuthTransitionRequest(r)
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_auth_transition_request", "invalid Network Flow auth-transition request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	transition, err := request.networkFlowAuthTransition()
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_auth_transition_request", "invalid Network Flow auth-transition request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.transitions.fixtureBound(transition) {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_auth_transition_request", "unresolved fixture references", map[string]any{"reason": "unresolved_fixture_reference"})
		return
	}
	if !s.transitions.arm(transition) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_network_flow_auth_transition_already_armed", "Network Flow auth transition is already armed", map[string]any{})
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, networkFlowAuthTransitionResult{
		SchemaID:       testNetworkFlowAuthTransitionSchemaID,
		ControlID:      transition.ID,
		Boundary:       transition.Boundary,
		TransitionKind: transition.TransitionKind,
		ActorRef:       transition.ActorRef,
		IncidentRef:    transition.IncidentRef,
		ResourceKind:   transition.ResourceKind,
		ResourceRef:    transition.ResourceRef,
		CorrelationKey: transition.CorrelationKey,
		ConsumeOnce:    true,
	})
}

func decodeNetworkFlowAuthTransitionRequest(r *http.Request) (networkFlowAuthTransitionRequest, error) {
	var request networkFlowAuthTransitionRequest
	err := decodeControlRequest(r, &request, "boundary", "transition_kind", "actor_ref", "incident_ref", "resource_kind", "resource_ref", "consume_once")
	return request, err
}

func (r networkFlowAuthTransitionRequest) networkFlowAuthTransition() (NetworkFlowAuthTransition, error) {
	boundary := strings.TrimSpace(r.Boundary)
	transitionKind := strings.TrimSpace(r.TransitionKind)
	actorRef := strings.TrimSpace(r.ActorRef)
	incidentRef := strings.TrimSpace(r.IncidentRef)
	resourceKind := strings.TrimSpace(r.ResourceKind)
	resourceRef := strings.TrimSpace(r.ResourceRef)
	if _, ok := networkFlowAuthTransitionBoundaries[boundary]; !ok {
		return NetworkFlowAuthTransition{}, errors.New("boundary is not a supported Network Flow auth-transition boundary")
	}
	if _, ok := networkFlowAuthTransitionKinds[transitionKind]; !ok {
		return NetworkFlowAuthTransition{}, errors.New("transition_kind is not supported")
	}
	if !isNetworkFlowAuthTransitionRef(actorRef) {
		return NetworkFlowAuthTransition{}, errors.New("actor_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if !isNetworkFlowAuthTransitionRef(incidentRef) {
		return NetworkFlowAuthTransition{}, errors.New("incident_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if _, ok := networkFlowAuthResourceKinds[resourceKind]; !ok {
		return NetworkFlowAuthTransition{}, errors.New("resource_kind is not supported")
	}
	if !isNetworkFlowAuthTransitionRef(resourceRef) {
		return NetworkFlowAuthTransition{}, errors.New("resource_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if !r.ConsumeOnce {
		return NetworkFlowAuthTransition{}, errors.New("consume_once must be true")
	}
	correlationKey := ""
	if r.CorrelationKey != nil {
		correlationKey = strings.TrimSpace(*r.CorrelationKey)
		if !isNetworkFlowAuthTransitionRef(correlationKey) {
			return NetworkFlowAuthTransition{}, errors.New("correlation_key must be an ASCII fixture reference no longer than 128 characters")
		}
	}
	return NetworkFlowAuthTransition{
		ID:             uuid.NewString(),
		Boundary:       boundary,
		TransitionKind: transitionKind,
		ActorRef:       actorRef,
		IncidentRef:    incidentRef,
		ResourceKind:   resourceKind,
		ResourceRef:    resourceRef,
		CorrelationKey: correlationKey,
	}, nil
}

func isNetworkFlowAuthTransitionRef(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII {
			return false
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '.', '_', ':', '-':
			continue
		default:
			return false
		}
	}
	return true
}

// RequireConsumed is called by a fixture before reporting success. Pending
// controls are missing evidence, including controls for unresolved references.
func (r *NetworkFlowAuthTransitionRegistry) RequireConsumed() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.transitions) != 0 {
		return errors.New("required Network Flow authorization transitions were not consumed")
	}
	return nil
}

// BindFixture reserves verified fixture references within this registry instance.
// A different set of rows cannot take over already bound symbolic references.
func (r *NetworkFlowAuthTransitionRegistry) BindFixture(actor, incident, kind, resource, identity string) error {
	if !isNetworkFlowAuthTransitionRef(actor) || !isNetworkFlowAuthTransitionRef(incident) || !isNetworkFlowAuthTransitionRef(resource) || identity == "" {
		return errors.New("invalid fixture binding")
	}
	if _, ok := networkFlowAuthResourceKinds[kind]; !ok {
		return errors.New("unknown fixture resource kind")
	}
	key := networkFlowAuthTransitionKey("", actor, incident, kind, resource)
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.fixtures[key]; ok && old != identity {
		return errors.New("fixture references already bound to different rows")
	}
	if r.fixtures == nil {
		r.fixtures = map[string]string{}
	}
	r.fixtures[key] = identity
	return nil
}
func (r *NetworkFlowAuthTransitionRegistry) fixtureBound(t NetworkFlowAuthTransition) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.fixtures[networkFlowAuthTransitionKey("", t.ActorRef, t.IncidentRef, t.ResourceKind, t.ResourceRef)]
	return ok
}
