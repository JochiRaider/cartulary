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

const testNetworkFlowFaultSchemaID = "cartulary.test.network_flow_fault_control.v1"

const (
	NetworkFlowFaultBoundaryImportBeforeOwnerPrepare                = "network_flow.import.before_owner_prepare"
	NetworkFlowFaultBoundaryImportAfterOwnerPrepare                 = "network_flow.import.after_owner_prepare"
	NetworkFlowFaultBoundaryImportAfterIndicatorPrepare             = "network_flow.import.after_indicator_prepare"
	NetworkFlowFaultBoundaryImportAfterAuditPrepare                 = "network_flow.import.after_audit_prepare"
	NetworkFlowFaultBoundaryImportAfterIdempotencyPrepare           = "network_flow.import.after_idempotency_prepare"
	NetworkFlowFaultBoundaryImportAfterTerminalPublicationPrepare   = "network_flow.import.after_terminal_publication_prepare"
	NetworkFlowFaultBoundaryImportBeforeTransactionCommit           = "network_flow.import.before_transaction_commit"
	NetworkFlowFaultBoundaryImportAfterTransactionCommitBeforeReply = "network_flow.import.after_transaction_commit_before_reply"
	NetworkFlowFaultBoundaryWorkerBeforeHandlerStart                = "network_flow.worker.before_handler_start"
	NetworkFlowFaultBoundaryWorkerBeforeApplyStart                  = "network_flow.worker.before_apply_start"
	NetworkFlowFaultBoundaryWorkerBeforeCancellationCheck           = "network_flow.worker.before_cancellation_check"
	NetworkFlowFaultBoundaryWorkerBeforeFinalCommit                 = "network_flow.worker.before_final_commit"
	NetworkFlowFaultBoundaryWorkerAfterFinalCommitBeforePublication = "network_flow.worker.after_final_commit_before_terminal_publication"
	NetworkFlowFaultBoundaryWorkerAfterPublicationBeforeAck         = "network_flow.worker.after_terminal_publication_before_ack"
	NetworkFlowFaultBoundaryWorkerBeforeReplayReconciliation        = "network_flow.worker.before_replay_reconciliation"
)

const (
	NetworkFlowFaultKindReturnError   = "return_error"
	NetworkFlowFaultKindPanic         = "panic"
	NetworkFlowFaultKindCancelContext = "cancel_context"
	NetworkFlowFaultKindWorkerCrash   = "worker_crash"
	NetworkFlowFaultKindWorkerCancel  = "worker_cancel"
)

var (
	networkFlowFaultBoundaries = map[string]struct{}{
		NetworkFlowFaultBoundaryImportBeforeOwnerPrepare:                {},
		NetworkFlowFaultBoundaryImportAfterOwnerPrepare:                 {},
		NetworkFlowFaultBoundaryImportAfterIndicatorPrepare:             {},
		NetworkFlowFaultBoundaryImportAfterAuditPrepare:                 {},
		NetworkFlowFaultBoundaryImportAfterIdempotencyPrepare:           {},
		NetworkFlowFaultBoundaryImportAfterTerminalPublicationPrepare:   {},
		NetworkFlowFaultBoundaryImportBeforeTransactionCommit:           {},
		NetworkFlowFaultBoundaryImportAfterTransactionCommitBeforeReply: {},
		NetworkFlowFaultBoundaryWorkerBeforeHandlerStart:                {},
		NetworkFlowFaultBoundaryWorkerBeforeApplyStart:                  {},
		NetworkFlowFaultBoundaryWorkerBeforeCancellationCheck:           {},
		NetworkFlowFaultBoundaryWorkerBeforeFinalCommit:                 {},
		NetworkFlowFaultBoundaryWorkerAfterFinalCommitBeforePublication: {},
		NetworkFlowFaultBoundaryWorkerAfterPublicationBeforeAck:         {},
		NetworkFlowFaultBoundaryWorkerBeforeReplayReconciliation:        {},
	}

	networkFlowFaultKinds = map[string]struct{}{
		NetworkFlowFaultKindReturnError:   {},
		NetworkFlowFaultKindPanic:         {},
		NetworkFlowFaultKindCancelContext: {},
		NetworkFlowFaultKindWorkerCrash:   {},
		NetworkFlowFaultKindWorkerCancel:  {},
	}
)

type NetworkFlowFaultRegistry struct {
	mu    sync.Mutex
	fault *NetworkFlowFault
}

type NetworkFlowFault struct {
	ID             string
	Boundary       string
	FaultKind      string
	ErrorCode      string
	CorrelationKey string
}

type networkFlowFaultService struct {
	guard  httpapi.TestRouteGuard
	faults *NetworkFlowFaultRegistry
}

type networkFlowFaultRequest struct {
	Boundary       string  `json:"boundary"`
	FaultKind      string  `json:"fault_kind"`
	ErrorCode      string  `json:"error_code"`
	CorrelationKey *string `json:"correlation_key"`
	ConsumeOnce    bool    `json:"consume_once"`
}

type networkFlowFaultResult struct {
	SchemaID       string `json:"schema_id"`
	FaultID        string `json:"fault_id"`
	Boundary       string `json:"boundary"`
	FaultKind      string `json:"fault_kind"`
	ErrorCode      string `json:"error_code,omitempty"`
	CorrelationKey string `json:"correlation_key,omitempty"`
	ConsumeOnce    bool   `json:"consume_once"`
}

func NewNetworkFlowFaultRegistry() *NetworkFlowFaultRegistry {
	return &NetworkFlowFaultRegistry{}
}

func RegisterNetworkFlowFaultRoutes(faults *NetworkFlowFaultRegistry) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		if faults == nil {
			return fmt.Errorf("register network flow fault route: fault registry is required")
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register network flow fault route: %w", err)
		}
		service := &networkFlowFaultService{
			guard:  guard,
			faults: faults,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/network-flow-faults", service.handleArm)
		return nil
	}
}

func (r *NetworkFlowFaultRegistry) ConsumeNetworkFlowFault(boundary string) (NetworkFlowFault, bool) {
	return r.consume(boundary, "")
}

func (r *NetworkFlowFaultRegistry) ConsumeNetworkFlowFaultFor(boundary string, correlationKey string) (NetworkFlowFault, bool) {
	return r.consume(boundary, strings.TrimSpace(correlationKey))
}

func (r *NetworkFlowFaultRegistry) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fault = nil
}

func (r *NetworkFlowFaultRegistry) arm(fault NetworkFlowFault) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fault != nil {
		return false
	}
	copy := fault
	r.fault = &copy
	return true
}

func (r *NetworkFlowFaultRegistry) consume(boundary string, correlationKey string) (NetworkFlowFault, bool) {
	if r == nil {
		return NetworkFlowFault{}, false
	}
	boundary = strings.TrimSpace(boundary)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fault == nil || r.fault.Boundary != boundary {
		return NetworkFlowFault{}, false
	}
	if r.fault.CorrelationKey != "" && r.fault.CorrelationKey != correlationKey {
		return NetworkFlowFault{}, false
	}
	fault := *r.fault
	r.fault = nil
	return fault, true
}

func (s *networkFlowFaultService) handleArm(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	request, err := decodeNetworkFlowFaultRequest(r)
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_fault_request", "invalid Network Flow fault request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	fault, err := request.networkFlowFault()
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_fault_request", "invalid Network Flow fault request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.faults.arm(fault) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_network_flow_fault_already_armed", "Network Flow fault is already armed", map[string]any{})
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, networkFlowFaultResult{
		SchemaID:       testNetworkFlowFaultSchemaID,
		FaultID:        fault.ID,
		Boundary:       fault.Boundary,
		FaultKind:      fault.FaultKind,
		ErrorCode:      fault.ErrorCode,
		CorrelationKey: fault.CorrelationKey,
		ConsumeOnce:    true,
	})
}

func decodeNetworkFlowFaultRequest(r *http.Request) (networkFlowFaultRequest, error) {
	var request networkFlowFaultRequest
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

func (r networkFlowFaultRequest) networkFlowFault() (NetworkFlowFault, error) {
	boundary := strings.TrimSpace(r.Boundary)
	faultKind := strings.TrimSpace(r.FaultKind)
	errorCode := strings.TrimSpace(r.ErrorCode)
	if _, ok := networkFlowFaultBoundaries[boundary]; !ok {
		return NetworkFlowFault{}, errors.New("boundary is not a supported Network Flow fault boundary")
	}
	if _, ok := networkFlowFaultKinds[faultKind]; !ok {
		return NetworkFlowFault{}, errors.New("fault_kind is not supported")
	}
	if !r.ConsumeOnce {
		return NetworkFlowFault{}, errors.New("consume_once must be true")
	}
	if isWorkerFaultKind(faultKind) && !strings.HasPrefix(boundary, "network_flow.worker.") {
		return NetworkFlowFault{}, errors.New("worker fault kinds require a worker boundary")
	}
	if faultKind == NetworkFlowFaultKindReturnError {
		if !isSafeNetworkFlowFaultErrorCode(errorCode) {
			return NetworkFlowFault{}, errors.New("return_error faults require a safe error_code")
		}
	} else if errorCode != "" {
		return NetworkFlowFault{}, errors.New("error_code is accepted only for return_error faults")
	}
	correlationKey := ""
	if r.CorrelationKey != nil {
		correlationKey = strings.TrimSpace(*r.CorrelationKey)
		if !isNetworkFlowFaultCorrelationKey(correlationKey) {
			return NetworkFlowFault{}, errors.New("correlation_key must be an ASCII token no longer than 128 characters")
		}
	}
	return NetworkFlowFault{
		ID:             uuid.NewString(),
		Boundary:       boundary,
		FaultKind:      faultKind,
		ErrorCode:      errorCode,
		CorrelationKey: correlationKey,
	}, nil
}

func isWorkerFaultKind(faultKind string) bool {
	return faultKind == NetworkFlowFaultKindWorkerCrash || faultKind == NetworkFlowFaultKindWorkerCancel
}

func isSafeNetworkFlowFaultErrorCode(value string) bool {
	if len(value) < 2 || len(value) > 128 {
		return false
	}
	for i, r := range value {
		if r > unicode.MaxASCII {
			return false
		}
		if i == 0 {
			if r < 'a' || r > 'z' {
				return false
			}
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func isNetworkFlowFaultCorrelationKey(value string) bool {
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
