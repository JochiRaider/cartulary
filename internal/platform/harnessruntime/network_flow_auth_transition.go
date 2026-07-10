package harnessruntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const testNetworkFlowAuthTransitionSchemaID = "cartulary.test.network_flow_auth_transition_control.v1"

const (
	NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization            = "network_flow.route.before_authorization"
	NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup = "network_flow.route.after_authorization_before_lookup"
	NetworkFlowAuthTransitionBoundaryRouteAfterLookupBeforeResponse      = "network_flow.route.after_lookup_before_response"
	NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck    = "network_flow.cursor.before_authorization_recheck"
	NetworkFlowAuthTransitionBoundaryWebSocketBeforeInvalidationPublish  = "network_flow.websocket.before_invalidation_publish"
	NetworkFlowAuthTransitionBoundaryFixtureAfterTransition              = "network_flow.fixture.after_transition"
)

const (
	NetworkFlowAuthTransitionKindIncidentMembershipRevoked   = "incident_membership_revoked"
	NetworkFlowAuthTransitionKindIncidentMembershipRestored  = "incident_membership_restored"
	NetworkFlowAuthTransitionKindIncidentSoftDeleted         = "incident_soft_deleted"
	NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted = "network_flow_table_soft_deleted"
	NetworkFlowAuthTransitionKindNetworkFlowTableRenamed     = "network_flow_table_renamed"
	NetworkFlowAuthTransitionKindSessionRevoked              = "session_revoked"
	NetworkFlowAuthTransitionKindExtensionClaimRemoved       = "extension_claim_removed"
)

const (
	NetworkFlowAuthResourceIncident                = "incident"
	NetworkFlowAuthResourceNetworkFlowTable        = "network_flow_table"
	NetworkFlowAuthResourceNetworkFlowCursor       = "network_flow_cursor"
	NetworkFlowAuthResourceNetworkFlowGraph        = "network_flow_graph"
	NetworkFlowAuthResourceNetworkFlowContributors = "network_flow_contributors"
	NetworkFlowAuthResourceNetworkFlowWorkspace    = "network_flow_workspace"
)

const (
	NetworkFlowHiddenResponseNotFound                   = "not_found"
	NetworkFlowHiddenResponseForbiddenWithoutResource   = "forbidden_without_resource"
	NetworkFlowHiddenResponseEmptyCollection            = "empty_collection"
	NetworkFlowHiddenResponseCursorRejected             = "cursor_rejected"
	NetworkFlowHiddenResponseExtensionProfileNotClaimed = "extension_profile_not_claimed"
	NetworkFlowHiddenResponseInvalidationEvent          = "invalidation_event"
)

var (
	networkFlowAuthTransitionBoundaries = map[string]struct{}{
		NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization:            {},
		NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup: {},
		NetworkFlowAuthTransitionBoundaryRouteAfterLookupBeforeResponse:      {},
		NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck:    {},
		NetworkFlowAuthTransitionBoundaryWebSocketBeforeInvalidationPublish:  {},
		NetworkFlowAuthTransitionBoundaryFixtureAfterTransition:              {},
	}

	networkFlowAuthTransitionKinds = map[string]struct{}{
		NetworkFlowAuthTransitionKindIncidentMembershipRevoked:   {},
		NetworkFlowAuthTransitionKindIncidentMembershipRestored:  {},
		NetworkFlowAuthTransitionKindIncidentSoftDeleted:         {},
		NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted: {},
		NetworkFlowAuthTransitionKindNetworkFlowTableRenamed:     {},
		NetworkFlowAuthTransitionKindSessionRevoked:              {},
		NetworkFlowAuthTransitionKindExtensionClaimRemoved:       {},
	}

	networkFlowAuthResourceKinds = map[string]struct{}{
		NetworkFlowAuthResourceIncident:                {},
		NetworkFlowAuthResourceNetworkFlowTable:        {},
		NetworkFlowAuthResourceNetworkFlowCursor:       {},
		NetworkFlowAuthResourceNetworkFlowGraph:        {},
		NetworkFlowAuthResourceNetworkFlowContributors: {},
		NetworkFlowAuthResourceNetworkFlowWorkspace:    {},
	}

	networkFlowHiddenResponseKinds = map[string]struct{}{
		NetworkFlowHiddenResponseNotFound:                   {},
		NetworkFlowHiddenResponseForbiddenWithoutResource:   {},
		NetworkFlowHiddenResponseEmptyCollection:            {},
		NetworkFlowHiddenResponseCursorRejected:             {},
		NetworkFlowHiddenResponseExtensionProfileNotClaimed: {},
		NetworkFlowHiddenResponseInvalidationEvent:          {},
	}
)

type NetworkFlowAuthTransitionRegistry struct {
	mu          sync.Mutex
	transitions map[string]NetworkFlowAuthTransition
}

type NetworkFlowAuthTransition struct {
	ID                      string
	Boundary                string
	TransitionKind          string
	ActorRef                string
	IncidentRef             string
	ResourceKind            string
	ResourceRef             string
	HiddenResponseKind      string
	MustNotDiscloseResource bool
	CorrelationKey          string
}

type networkFlowAuthTransitionService struct {
	guard       httpapi.TestRouteGuard
	transitions *NetworkFlowAuthTransitionRegistry
}

type networkFlowAuthTransitionRequest struct {
	Boundary                string  `json:"boundary"`
	TransitionKind          string  `json:"transition_kind"`
	ActorRef                string  `json:"actor_ref"`
	IncidentRef             string  `json:"incident_ref"`
	ResourceKind            string  `json:"resource_kind"`
	ResourceRef             string  `json:"resource_ref"`
	HiddenResponseKind      string  `json:"hidden_response_kind"`
	MustNotDiscloseResource bool    `json:"must_not_disclose_resource"`
	CorrelationKey          *string `json:"correlation_key"`
	ConsumeOnce             bool    `json:"consume_once"`
}

type networkFlowAuthTransitionResult struct {
	SchemaID                string `json:"schema_id"`
	ControlID               string `json:"control_id"`
	Boundary                string `json:"boundary"`
	TransitionKind          string `json:"transition_kind"`
	ActorRef                string `json:"actor_ref"`
	IncidentRef             string `json:"incident_ref"`
	ResourceKind            string `json:"resource_kind"`
	ResourceRef             string `json:"resource_ref"`
	HiddenResponseKind      string `json:"hidden_response_kind"`
	MustNotDiscloseResource bool   `json:"must_not_disclose_resource"`
	CorrelationKey          string `json:"correlation_key,omitempty"`
	ConsumeOnce             bool   `json:"consume_once"`
}

func NewNetworkFlowAuthTransitionRegistry() *NetworkFlowAuthTransitionRegistry {
	return &NetworkFlowAuthTransitionRegistry{transitions: map[string]NetworkFlowAuthTransition{}}
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
	if !s.transitions.arm(transition) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_network_flow_auth_transition_already_armed", "Network Flow auth transition is already armed", map[string]any{})
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, networkFlowAuthTransitionResult{
		SchemaID:                testNetworkFlowAuthTransitionSchemaID,
		ControlID:               transition.ID,
		Boundary:                transition.Boundary,
		TransitionKind:          transition.TransitionKind,
		ActorRef:                transition.ActorRef,
		IncidentRef:             transition.IncidentRef,
		ResourceKind:            transition.ResourceKind,
		ResourceRef:             transition.ResourceRef,
		HiddenResponseKind:      transition.HiddenResponseKind,
		MustNotDiscloseResource: true,
		CorrelationKey:          transition.CorrelationKey,
		ConsumeOnce:             true,
	})
}

func decodeNetworkFlowAuthTransitionRequest(r *http.Request) (networkFlowAuthTransitionRequest, error) {
	var request networkFlowAuthTransitionRequest
	if r.Body == nil {
		return request, errors.New("body is required")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, fmt.Errorf("decode body: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return request, errors.New("body must contain a single JSON object")
	}
	return request, nil
}

func (r networkFlowAuthTransitionRequest) networkFlowAuthTransition() (NetworkFlowAuthTransition, error) {
	boundary := strings.TrimSpace(r.Boundary)
	transitionKind := strings.TrimSpace(r.TransitionKind)
	actorRef := strings.TrimSpace(r.ActorRef)
	incidentRef := strings.TrimSpace(r.IncidentRef)
	resourceKind := strings.TrimSpace(r.ResourceKind)
	resourceRef := strings.TrimSpace(r.ResourceRef)
	hiddenResponseKind := strings.TrimSpace(r.HiddenResponseKind)
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
	if _, ok := networkFlowHiddenResponseKinds[hiddenResponseKind]; !ok {
		return NetworkFlowAuthTransition{}, errors.New("hidden_response_kind is not supported")
	}
	if !r.MustNotDiscloseResource {
		return NetworkFlowAuthTransition{}, errors.New("must_not_disclose_resource must be true")
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
		ID:                      uuid.NewString(),
		Boundary:                boundary,
		TransitionKind:          transitionKind,
		ActorRef:                actorRef,
		IncidentRef:             incidentRef,
		ResourceKind:            resourceKind,
		ResourceRef:             resourceRef,
		HiddenResponseKind:      hiddenResponseKind,
		MustNotDiscloseResource: true,
		CorrelationKey:          correlationKey,
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
