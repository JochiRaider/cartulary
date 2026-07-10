package testruntime

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

const testNetworkFlowAuditAssertionSchemaID = "cartulary.test.network_flow_audit_assertion_control.v1"

const (
	NetworkFlowAuditAssertionExactCount      = "exact_count"
	NetworkFlowAuditAssertionZeroOccurrences = "zero_occurrences"
	NetworkFlowAuditAssertionNoAuditReplay   = "no_audit_replay"
)

const (
	NetworkFlowAuditEventTableCreated            = "network_flow_table_created"
	NetworkFlowAuditEventTableRenamed            = "network_flow_table_renamed"
	NetworkFlowAuditEventTableSoftDeleted        = "network_flow_table_soft_deleted"
	NetworkFlowAuditEventGraphQueryExecuted      = "network_flow_graph_query_executed"
	NetworkFlowAuditEventIndicatorBindingCreated = "network_flow_indicator_binding_created"
	NetworkFlowAuditEventIndicatorBindingReused  = "network_flow_indicator_binding_reused"
)

const (
	NetworkFlowAuditResourceTable            = "network_flow_table"
	NetworkFlowAuditResourceGraph            = "network_flow_graph"
	NetworkFlowAuditResourceIndicatorBinding = "network_flow_indicator_binding"
	NetworkFlowAuditResourceImport           = "network_flow_import"
)

var (
	networkFlowAuditAssertionKinds = map[string]struct{}{
		NetworkFlowAuditAssertionExactCount:      {},
		NetworkFlowAuditAssertionZeroOccurrences: {},
		NetworkFlowAuditAssertionNoAuditReplay:   {},
	}

	networkFlowAuditEventCodes = map[string]struct{}{
		NetworkFlowAuditEventTableCreated:            {},
		NetworkFlowAuditEventTableRenamed:            {},
		NetworkFlowAuditEventTableSoftDeleted:        {},
		NetworkFlowAuditEventGraphQueryExecuted:      {},
		NetworkFlowAuditEventIndicatorBindingCreated: {},
		NetworkFlowAuditEventIndicatorBindingReused:  {},
	}

	networkFlowAuditResourceKinds = map[string]struct{}{
		NetworkFlowAuditResourceTable:            {},
		NetworkFlowAuditResourceGraph:            {},
		NetworkFlowAuditResourceIndicatorBinding: {},
		NetworkFlowAuditResourceImport:           {},
	}
)

type NetworkFlowAuditAssertionRegistry struct {
	mu         sync.Mutex
	assertions map[string]NetworkFlowAuditAssertion
}

type NetworkFlowAuditAssertion struct {
	ID                      string
	AssertionKind           string
	EventCode               string
	OperationRef            string
	ActorRef                string
	IncidentRef             string
	ResourceKind            string
	ResourceRef             string
	BaselineCount           int
	ExpectedFinalCount      int
	ExpectedReplayIncrement int
	CorrelationKey          string
}

type networkFlowAuditAssertionService struct {
	guard      httpapi.TestRouteGuard
	assertions *NetworkFlowAuditAssertionRegistry
}

type networkFlowAuditAssertionRequest struct {
	AssertionKind           string  `json:"assertion_kind"`
	EventCode               string  `json:"event_code"`
	OperationRef            string  `json:"operation_ref"`
	ActorRef                string  `json:"actor_ref"`
	IncidentRef             string  `json:"incident_ref"`
	ResourceKind            string  `json:"resource_kind"`
	ResourceRef             string  `json:"resource_ref"`
	BaselineCount           int     `json:"baseline_count"`
	ExpectedFinalCount      int     `json:"expected_final_count"`
	ExpectedReplayIncrement int     `json:"expected_replay_increment"`
	CorrelationKey          *string `json:"correlation_key"`
	ConsumeOnce             bool    `json:"consume_once"`
}

type networkFlowAuditAssertionResult struct {
	SchemaID                string `json:"schema_id"`
	AssertionID             string `json:"assertion_id"`
	AssertionKind           string `json:"assertion_kind"`
	EventCode               string `json:"event_code"`
	OperationRef            string `json:"operation_ref"`
	ActorRef                string `json:"actor_ref"`
	IncidentRef             string `json:"incident_ref"`
	ResourceKind            string `json:"resource_kind"`
	ResourceRef             string `json:"resource_ref"`
	BaselineCount           int    `json:"baseline_count"`
	ExpectedFinalCount      int    `json:"expected_final_count"`
	ExpectedReplayIncrement int    `json:"expected_replay_increment"`
	CorrelationKey          string `json:"correlation_key,omitempty"`
	ConsumeOnce             bool   `json:"consume_once"`
}

func NewNetworkFlowAuditAssertionRegistry() *NetworkFlowAuditAssertionRegistry {
	return &NetworkFlowAuditAssertionRegistry{assertions: map[string]NetworkFlowAuditAssertion{}}
}

func RegisterNetworkFlowAuditAssertionRoutes(assertions *NetworkFlowAuditAssertionRegistry) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		if assertions == nil {
			return fmt.Errorf("register network flow audit assertion route: assertion registry is required")
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register network flow audit assertion route: %w", err)
		}
		service := &networkFlowAuditAssertionService{
			guard:      guard,
			assertions: assertions,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/network-flow-audit-assertions", service.handleArm)
		return nil
	}
}

func (r *NetworkFlowAuditAssertionRegistry) ConsumeNetworkFlowAuditAssertion(eventCode string, operationRef string, resourceKind string, resourceRef string) (NetworkFlowAuditAssertion, bool) {
	return r.ConsumeNetworkFlowAuditAssertionFor(eventCode, operationRef, resourceKind, resourceRef, "")
}

func (r *NetworkFlowAuditAssertionRegistry) ConsumeNetworkFlowAuditAssertionFor(eventCode string, operationRef string, resourceKind string, resourceRef string, correlationKey string) (NetworkFlowAuditAssertion, bool) {
	if r == nil {
		return NetworkFlowAuditAssertion{}, false
	}
	key := networkFlowAuditAssertionKey(eventCode, operationRef, resourceKind, resourceRef)
	correlationKey = strings.TrimSpace(correlationKey)
	r.mu.Lock()
	defer r.mu.Unlock()
	assertion, ok := r.assertions[key]
	if !ok {
		return NetworkFlowAuditAssertion{}, false
	}
	if assertion.CorrelationKey != "" && assertion.CorrelationKey != correlationKey {
		return NetworkFlowAuditAssertion{}, false
	}
	delete(r.assertions, key)
	return assertion, true
}

func (r *NetworkFlowAuditAssertionRegistry) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.assertions = map[string]NetworkFlowAuditAssertion{}
}

func (r *NetworkFlowAuditAssertionRegistry) arm(assertion NetworkFlowAuditAssertion) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.assertions == nil {
		r.assertions = map[string]NetworkFlowAuditAssertion{}
	}
	key := networkFlowAuditAssertionKey(assertion.EventCode, assertion.OperationRef, assertion.ResourceKind, assertion.ResourceRef)
	if _, exists := r.assertions[key]; exists {
		return false
	}
	r.assertions[key] = assertion
	return true
}

func networkFlowAuditAssertionKey(eventCode string, operationRef string, resourceKind string, resourceRef string) string {
	return strings.TrimSpace(eventCode) + "\x00" + strings.TrimSpace(operationRef) + "\x00" + strings.TrimSpace(resourceKind) + "\x00" + strings.TrimSpace(resourceRef)
}

func (s *networkFlowAuditAssertionService) handleArm(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	request, err := decodeNetworkFlowAuditAssertionRequest(r)
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_audit_assertion_request", "invalid Network Flow audit assertion request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	assertion, err := request.networkFlowAuditAssertion()
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_audit_assertion_request", "invalid Network Flow audit assertion request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.assertions.arm(assertion) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_network_flow_audit_assertion_already_armed", "Network Flow audit assertion is already armed", map[string]any{})
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, networkFlowAuditAssertionResult{
		SchemaID:                testNetworkFlowAuditAssertionSchemaID,
		AssertionID:             assertion.ID,
		AssertionKind:           assertion.AssertionKind,
		EventCode:               assertion.EventCode,
		OperationRef:            assertion.OperationRef,
		ActorRef:                assertion.ActorRef,
		IncidentRef:             assertion.IncidentRef,
		ResourceKind:            assertion.ResourceKind,
		ResourceRef:             assertion.ResourceRef,
		BaselineCount:           assertion.BaselineCount,
		ExpectedFinalCount:      assertion.ExpectedFinalCount,
		ExpectedReplayIncrement: assertion.ExpectedReplayIncrement,
		CorrelationKey:          assertion.CorrelationKey,
		ConsumeOnce:             true,
	})
}

func decodeNetworkFlowAuditAssertionRequest(r *http.Request) (networkFlowAuditAssertionRequest, error) {
	var request networkFlowAuditAssertionRequest
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

func (r networkFlowAuditAssertionRequest) networkFlowAuditAssertion() (NetworkFlowAuditAssertion, error) {
	assertionKind := strings.TrimSpace(r.AssertionKind)
	eventCode := strings.TrimSpace(r.EventCode)
	operationRef := strings.TrimSpace(r.OperationRef)
	actorRef := strings.TrimSpace(r.ActorRef)
	incidentRef := strings.TrimSpace(r.IncidentRef)
	resourceKind := strings.TrimSpace(r.ResourceKind)
	resourceRef := strings.TrimSpace(r.ResourceRef)
	if _, ok := networkFlowAuditAssertionKinds[assertionKind]; !ok {
		return NetworkFlowAuditAssertion{}, errors.New("assertion_kind is not supported")
	}
	if _, ok := networkFlowAuditEventCodes[eventCode]; !ok {
		return NetworkFlowAuditAssertion{}, errors.New("event_code is not supported")
	}
	if !isNetworkFlowAuditAssertionRef(operationRef) {
		return NetworkFlowAuditAssertion{}, errors.New("operation_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if !isNetworkFlowAuditAssertionRef(actorRef) {
		return NetworkFlowAuditAssertion{}, errors.New("actor_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if !isNetworkFlowAuditAssertionRef(incidentRef) {
		return NetworkFlowAuditAssertion{}, errors.New("incident_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if _, ok := networkFlowAuditResourceKinds[resourceKind]; !ok {
		return NetworkFlowAuditAssertion{}, errors.New("resource_kind is not supported")
	}
	if !isNetworkFlowAuditAssertionRef(resourceRef) {
		return NetworkFlowAuditAssertion{}, errors.New("resource_ref must be an ASCII fixture reference no longer than 128 characters")
	}
	if err := validateNetworkFlowAuditAssertionCounts(assertionKind, r.BaselineCount, r.ExpectedFinalCount, r.ExpectedReplayIncrement); err != nil {
		return NetworkFlowAuditAssertion{}, err
	}
	if !r.ConsumeOnce {
		return NetworkFlowAuditAssertion{}, errors.New("consume_once must be true")
	}
	correlationKey := ""
	if r.CorrelationKey != nil {
		correlationKey = strings.TrimSpace(*r.CorrelationKey)
		if !isNetworkFlowAuditAssertionRef(correlationKey) {
			return NetworkFlowAuditAssertion{}, errors.New("correlation_key must be an ASCII fixture reference no longer than 128 characters")
		}
	}
	return NetworkFlowAuditAssertion{
		ID:                      uuid.NewString(),
		AssertionKind:           assertionKind,
		EventCode:               eventCode,
		OperationRef:            operationRef,
		ActorRef:                actorRef,
		IncidentRef:             incidentRef,
		ResourceKind:            resourceKind,
		ResourceRef:             resourceRef,
		BaselineCount:           r.BaselineCount,
		ExpectedFinalCount:      r.ExpectedFinalCount,
		ExpectedReplayIncrement: r.ExpectedReplayIncrement,
		CorrelationKey:          correlationKey,
	}, nil
}

func validateNetworkFlowAuditAssertionCounts(assertionKind string, baselineCount int, expectedFinalCount int, expectedReplayIncrement int) error {
	for name, value := range map[string]int{
		"baseline_count":            baselineCount,
		"expected_final_count":      expectedFinalCount,
		"expected_replay_increment": expectedReplayIncrement,
	} {
		if value < 0 || value > 1_000_000 {
			return fmt.Errorf("%s must be between 0 and 1000000", name)
		}
	}
	if expectedFinalCount < baselineCount {
		return errors.New("expected_final_count must be greater than or equal to baseline_count")
	}
	switch assertionKind {
	case NetworkFlowAuditAssertionZeroOccurrences:
		if baselineCount != 0 || expectedFinalCount != 0 || expectedReplayIncrement != 0 {
			return errors.New("zero_occurrences requires all count fields to be zero")
		}
	case NetworkFlowAuditAssertionNoAuditReplay:
		if expectedReplayIncrement != 0 {
			return errors.New("no_audit_replay requires expected_replay_increment=0")
		}
	}
	return nil
}

func isNetworkFlowAuditAssertionRef(value string) bool {
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
