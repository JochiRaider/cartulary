package harnessruntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const testPublicErrorFaultSchemaID = "cartulary.test.public_error_fault.v1"

type PublicErrorFaultRegistry struct {
	mu     sync.Mutex
	faults map[string]publicErrorFault
}

type publicErrorFault struct {
	ID        string
	Method    string
	Path      string
	Status    int
	Code      string
	Message   string
	Retryable bool
	Details   map[string]any
}

type publicErrorFaultService struct {
	guard  httpapi.TestRouteGuard
	faults *PublicErrorFaultRegistry
}

type publicErrorFaultRequest struct {
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Status      int            `json:"status"`
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	Retryable   *bool          `json:"retryable"`
	Details     map[string]any `json:"details"`
	ConsumeOnce bool           `json:"consume_once"`
}

type publicErrorFaultResult struct {
	SchemaID    string `json:"schema_id"`
	FaultID     string `json:"fault_id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Status      int    `json:"status"`
	Code        string `json:"code"`
	Retryable   bool   `json:"retryable"`
	ConsumeOnce bool   `json:"consume_once"`
}

func NewPublicErrorFaultRegistry() *PublicErrorFaultRegistry {
	return &PublicErrorFaultRegistry{faults: map[string]publicErrorFault{}}
}

func RegisterPublicErrorFaultRoutes(faults *PublicErrorFaultRegistry) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		if faults == nil {
			return fmt.Errorf("register public error fault route: fault registry is required")
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register public error fault route: %w", err)
		}
		service := &publicErrorFaultService{
			guard:  guard,
			faults: faults,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/public-error-faults", service.handleArm)
		return nil
	}
}

func (r *PublicErrorFaultRegistry) ConsumePublicErrorFault(method string, path string) (httpapi.PublicErrorFault, bool) {
	if r == nil {
		return httpapi.PublicErrorFault{}, false
	}
	key := publicErrorFaultKey(method, path)
	r.mu.Lock()
	defer r.mu.Unlock()
	fault, ok := r.faults[key]
	if !ok {
		return httpapi.PublicErrorFault{}, false
	}
	delete(r.faults, key)
	details := map[string]any{}
	for key, value := range fault.Details {
		details[key] = value
	}
	return httpapi.PublicErrorFault{
		Status:    fault.Status,
		Code:      fault.Code,
		Message:   fault.Message,
		Retryable: fault.Retryable,
		Details:   details,
	}, true
}

func (r *PublicErrorFaultRegistry) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.faults = map[string]publicErrorFault{}
}

func (r *PublicErrorFaultRegistry) arm(fault publicErrorFault) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.faults == nil {
		r.faults = map[string]publicErrorFault{}
	}
	if len(r.faults) > 0 {
		return false
	}
	r.faults[publicErrorFaultKey(fault.Method, fault.Path)] = fault
	return true
}

func publicErrorFaultKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "\x00" + strings.TrimSpace(path)
}

func (s *publicErrorFaultService) handleArm(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	request, err := decodePublicErrorFaultRequest(r)
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_public_error_fault_request", "invalid public error fault request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	fault, err := request.publicErrorFault()
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_public_error_fault_request", "invalid public error fault request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.faults.arm(fault) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_public_error_fault_already_armed", "public error fault is already armed", map[string]any{})
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, publicErrorFaultResult{
		SchemaID:    testPublicErrorFaultSchemaID,
		FaultID:     fault.ID,
		Method:      fault.Method,
		Path:        fault.Path,
		Status:      fault.Status,
		Code:        fault.Code,
		Retryable:   fault.Retryable,
		ConsumeOnce: true,
	})
}

func decodePublicErrorFaultRequest(r *http.Request) (publicErrorFaultRequest, error) {
	var request publicErrorFaultRequest
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

func (r publicErrorFaultRequest) publicErrorFault() (publicErrorFault, error) {
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	path := strings.TrimSpace(r.Path)
	code := strings.TrimSpace(r.Code)
	message := strings.TrimSpace(r.Message)
	if method == "" {
		return publicErrorFault{}, errors.New("method is required")
	}
	if path == "" {
		return publicErrorFault{}, errors.New("path is required")
	}
	if !strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/api/v1/test/") {
		return publicErrorFault{}, errors.New("path must be an ordinary /api/v1/ route and must not target /api/v1/test/")
	}
	if strings.ContainsAny(path, "?#") {
		return publicErrorFault{}, errors.New("path must be an exact path without query or fragment")
	}
	if r.Status < 400 || r.Status > 599 {
		return publicErrorFault{}, errors.New("status must be between 400 and 599")
	}
	if code == "" {
		return publicErrorFault{}, errors.New("code is required")
	}
	if !r.ConsumeOnce {
		return publicErrorFault{}, errors.New("consume_once must be true")
	}
	retryable := false
	if r.Retryable != nil {
		retryable = *r.Retryable
	}
	details := map[string]any{}
	for key, value := range r.Details {
		if strings.TrimSpace(key) == "" {
			return publicErrorFault{}, errors.New("details keys must be non-empty")
		}
		details[key] = value
	}
	return publicErrorFault{
		ID:        uuid.NewString(),
		Method:    method,
		Path:      path,
		Status:    r.Status,
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Details:   details,
	}, nil
}
