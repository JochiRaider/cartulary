package httpapi

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type PublicAuthentication string

const (
	PublicAuthenticationPublic             PublicAuthentication = "public"
	PublicAuthenticationSession            PublicAuthentication = "session"
	PublicAuthenticationSessionOrBootstrap PublicAuthentication = "session_or_bootstrap"
)

type CanonicalPublicRouteExclusion string

const (
	CanonicalPublicRouteExclusionDetailedViewFamily CanonicalPublicRouteExclusion = "detailed_view_family"
	CanonicalPublicRouteExclusionNetworkFlow        CanonicalPublicRouteExclusion = "network_flow_contract"
)

type PublicOperation struct {
	OwnerID        string
	Method         string
	PathTemplate   string
	OperationID    string
	Authentication PublicAuthentication
	StateChanging  bool
	SuccessStatus  int
}

func NewPublicOperation(
	ownerID string,
	method string,
	pathTemplate string,
	operationID string,
	authentication PublicAuthentication,
	stateChanging bool,
	successStatus int,
) PublicOperation {
	return PublicOperation{
		OwnerID:        ownerID,
		Method:         method,
		PathTemplate:   pathTemplate,
		OperationID:    operationID,
		Authentication: authentication,
		StateChanging:  stateChanging,
		SuccessStatus:  successStatus,
	}
}

type PublicOperationRegistry struct {
	mu         sync.Mutex
	operations map[string]PublicOperation
}

func NewPublicOperationRegistry() *PublicOperationRegistry {
	return &PublicOperationRegistry{operations: make(map[string]PublicOperation)}
}

func (registry *PublicOperationRegistry) Declare(operations ...PublicOperation) error {
	if registry == nil {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.operations == nil {
		registry.operations = make(map[string]PublicOperation)
	}
	declaredOperationIDs := make(map[string]string, len(registry.operations)+len(operations))
	for key, operation := range registry.operations {
		declaredOperationIDs[operation.OperationID] = key
	}
	pending := make(map[string]PublicOperation, len(operations))
	for _, operation := range operations {
		if err := ValidatePublicOperation(operation); err != nil {
			return err
		}
		key := operation.Method + " " + operation.PathTemplate
		if existing, ok := registry.operations[key]; ok {
			return fmt.Errorf(
				"duplicate public operation %s: %s and %s",
				key,
				existing.OperationID,
				operation.OperationID,
			)
		}
		if existing, ok := pending[key]; ok {
			return fmt.Errorf(
				"duplicate public operation %s: %s and %s",
				key,
				existing.OperationID,
				operation.OperationID,
			)
		}
		if existingKey, ok := declaredOperationIDs[operation.OperationID]; ok {
			return fmt.Errorf(
				"duplicate public operation ID %q: %s and %s",
				operation.OperationID,
				existingKey,
				key,
			)
		}
		pending[key] = operation
		declaredOperationIDs[operation.OperationID] = key
	}
	for key, operation := range pending {
		registry.operations[key] = operation
	}
	return nil
}

func (registry *PublicOperationRegistry) Snapshot() []PublicOperation {
	if registry == nil {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	operations := make([]PublicOperation, 0, len(registry.operations))
	for _, operation := range registry.operations {
		operations = append(operations, operation)
	}
	sort.Slice(operations, func(left, right int) bool {
		if operations[left].PathTemplate != operations[right].PathTemplate {
			return operations[left].PathTemplate < operations[right].PathTemplate
		}
		if operations[left].Method != operations[right].Method {
			return operations[left].Method < operations[right].Method
		}
		return operations[left].OperationID < operations[right].OperationID
	})
	return operations
}

func DeclarePublicOperations(deps DependencySet, operations ...PublicOperation) error {
	return deps.PublicOperations.Declare(operations...)
}

func HandlePublicRoute(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	mux.HandleFunc(pattern, handler)
}

func HandleExcludedPublicRoute(
	mux *http.ServeMux,
	pattern string,
	handler http.HandlerFunc,
	exclusion CanonicalPublicRouteExclusion,
) {
	switch exclusion {
	case CanonicalPublicRouteExclusionDetailedViewFamily, CanonicalPublicRouteExclusionNetworkFlow:
		mux.HandleFunc(pattern, handler)
	default:
		panic(fmt.Sprintf("invalid canonical public route exclusion %q", exclusion))
	}
}

func ValidatePublicOperation(operation PublicOperation) error {
	switch {
	case operation.OwnerID == "":
		return fmt.Errorf("public operation %q has no owner", operation.OperationID)
	case operation.Method == "":
		return fmt.Errorf("public operation %q has no method", operation.OperationID)
	case operation.Method != strings.ToUpper(operation.Method):
		return fmt.Errorf("public operation %q method must be uppercase", operation.OperationID)
	case operation.PathTemplate == "" || !strings.HasPrefix(operation.PathTemplate, "/api/v1/"):
		return fmt.Errorf("public operation %q has invalid API path %q", operation.OperationID, operation.PathTemplate)
	case operation.OperationID == "":
		return fmt.Errorf("%s %s has no operation ID", operation.Method, operation.PathTemplate)
	case operation.SuccessStatus < http.StatusOK || operation.SuccessStatus >= http.StatusBadRequest:
		return fmt.Errorf("public operation %q has invalid success status %d", operation.OperationID, operation.SuccessStatus)
	}
	switch operation.Authentication {
	case PublicAuthenticationPublic, PublicAuthenticationSession, PublicAuthenticationSessionOrBootstrap:
		return nil
	default:
		return fmt.Errorf("public operation %q has invalid authentication class %q", operation.OperationID, operation.Authentication)
	}
}
