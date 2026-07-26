package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/openapioperations"
)

const BaseOperationAvailability = "base"

type ContractOperation struct {
	OwnerID         string
	Method          string
	PathTemplate    string
	Pattern         string
	OperationID     string
	Availability    string
	StateChanging   bool
	SuccessStatuses []int
	Security        [][]string
}

type OperationMetadata struct {
	OwnerID      string
	OperationID  string
	Availability string
}

type RouteDiagnostics struct {
	CanonicalSHA256         string
	DocumentVersion         string
	SupportedOperationCount int
	ActiveOperationCount    int
	ClaimedProfiles         []string
}

type operationMetadataContextKey struct{}

type RouteRegistry struct {
	mu              sync.Mutex
	operationsByID  map[string]ContractOperation
	operationsByKey map[string]ContractOperation
	claimedProfiles map[string]struct{}
	active          map[string]ContractOperation
}

func NewRouteRegistry(claims []ExtensionClaim) (*RouteRegistry, error) {
	return newRouteRegistry(generatedContractOperations(), claims)
}

func newRouteRegistry(operations []ContractOperation, claims []ExtensionClaim) (*RouteRegistry, error) {
	registry := &RouteRegistry{
		operationsByID:  make(map[string]ContractOperation, len(operations)),
		operationsByKey: make(map[string]ContractOperation, len(operations)),
		claimedProfiles: make(map[string]struct{}),
		active:          make(map[string]ContractOperation),
	}
	for _, claim := range claims {
		if claim.Claimed {
			registry.claimedProfiles[claim.ProfileID] = struct{}{}
		}
	}
	for _, operation := range operations {
		if err := validateContractOperation(operation); err != nil {
			return nil, err
		}
		if previous, duplicate := registry.operationsByID[operation.OperationID]; duplicate {
			return nil, fmt.Errorf(
				"duplicate generated operation ID %q at %s and %s",
				operation.OperationID,
				previous.Pattern,
				operation.Pattern,
			)
		}
		if previous, duplicate := registry.operationsByKey[operation.Pattern]; duplicate {
			return nil, fmt.Errorf(
				"duplicate generated operation pattern %q owned by %s and %s",
				operation.Pattern,
				previous.OwnerID,
				operation.OwnerID,
			)
		}
		registry.operationsByID[operation.OperationID] = cloneContractOperation(operation)
		registry.operationsByKey[operation.Pattern] = cloneContractOperation(operation)
	}
	return registry, nil
}

func (registry *RouteRegistry) Bind(
	mux *http.ServeMux,
	ownerID string,
	operationID string,
	handler http.HandlerFunc,
) error {
	if registry == nil {
		return errors.New("public route registry is required")
	}
	if mux == nil || handler == nil {
		return fmt.Errorf("bind public operation %q requires a mux and handler", operationID)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	operation, ok := registry.operationsByID[operationID]
	if !ok {
		return fmt.Errorf("unknown generated public operation %q", operationID)
	}
	if operation.OwnerID != ownerID {
		return fmt.Errorf(
			"public operation %q is owned by %q, not %q",
			operationID,
			operation.OwnerID,
			ownerID,
		)
	}
	if previous, duplicate := registry.active[operationID]; duplicate {
		return fmt.Errorf("public operation %q is already bound at %s", operationID, previous.Pattern)
	}
	metadata := OperationMetadata{
		OwnerID:      operation.OwnerID,
		OperationID:  operation.OperationID,
		Availability: operation.Availability,
	}
	mux.HandleFunc(operation.Pattern, func(writer http.ResponseWriter, request *http.Request) {
		contextWithOperation := context.WithValue(request.Context(), operationMetadataContextKey{}, metadata)
		handler(writer, request.WithContext(contextWithOperation))
	})
	registry.active[operationID] = cloneContractOperation(operation)
	return nil
}

func (registry *RouteRegistry) BindOwner(
	mux *http.ServeMux,
	ownerID string,
	handlers map[string]http.HandlerFunc,
) error {
	operationIDs := make([]string, 0, len(handlers))
	for operationID := range handlers {
		operationIDs = append(operationIDs, operationID)
	}
	sort.Strings(operationIDs)
	for _, operationID := range operationIDs {
		if err := registry.Bind(mux, ownerID, operationID, handlers[operationID]); err != nil {
			return err
		}
	}
	return nil
}

func BindOwnerRoutes(
	mux *http.ServeMux,
	deps DependencySet,
	ownerID string,
	handlers map[string]http.HandlerFunc,
) error {
	registry := deps.PublicRoutes
	if registry == nil {
		var err error
		registry, err = NewRouteRegistry(ExtensionClaimsFromDependencies(deps))
		if err != nil {
			return err
		}
	}
	return registry.BindOwner(mux, ownerID, handlers)
}

func (registry *RouteRegistry) ValidateActive() error {
	if registry == nil {
		return errors.New("public route registry is required")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	var missing []string
	var unexpected []string
	for operationID, operation := range registry.operationsByID {
		_, active := registry.active[operationID]
		expected := operation.Availability == BaseOperationAvailability
		if !expected {
			_, expected = registry.claimedProfiles[operation.Availability]
		}
		switch {
		case expected && !active:
			missing = append(missing, operation.Pattern+" ["+operationID+"]")
		case !expected && active:
			unexpected = append(unexpected, operation.Pattern+" ["+operationID+"]")
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)
	if len(missing) > 0 || len(unexpected) > 0 {
		return fmt.Errorf(
			"public route parity failed: missing=%s unexpected=%s",
			strings.Join(missing, ", "),
			strings.Join(unexpected, ", "),
		)
	}
	return nil
}

func (registry *RouteRegistry) Snapshot() []ContractOperation {
	if registry == nil {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	operations := make([]ContractOperation, 0, len(registry.active))
	for _, operation := range registry.active {
		operations = append(operations, cloneContractOperation(operation))
	}
	sortContractOperations(operations)
	return operations
}

func (registry *RouteRegistry) Diagnostics() RouteDiagnostics {
	diagnostics := RouteDiagnostics{
		CanonicalSHA256: openapioperations.CanonicalSHA256,
		DocumentVersion: openapioperations.DocumentVersion,
	}
	if registry == nil {
		return diagnostics
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	diagnostics.SupportedOperationCount = len(registry.operationsByID)
	diagnostics.ActiveOperationCount = len(registry.active)
	diagnostics.ClaimedProfiles = make([]string, 0, len(registry.claimedProfiles))
	for profileID := range registry.claimedProfiles {
		diagnostics.ClaimedProfiles = append(diagnostics.ClaimedProfiles, profileID)
	}
	sort.Strings(diagnostics.ClaimedProfiles)
	return diagnostics
}

func ContractOperations() []ContractOperation {
	operations := generatedContractOperations()
	sortContractOperations(operations)
	return operations
}

func ContractOperationsForOwner(ownerID string) []ContractOperation {
	all := generatedContractOperations()
	operations := make([]ContractOperation, 0)
	for _, operation := range all {
		if operation.OwnerID == ownerID {
			operations = append(operations, cloneContractOperation(operation))
		}
	}
	sortContractOperations(operations)
	return operations
}

func OperationMetadataFromContext(ctx context.Context) (OperationMetadata, bool) {
	metadata, ok := ctx.Value(operationMetadataContextKey{}).(OperationMetadata)
	return metadata, ok
}

func generatedContractOperations() []ContractOperation {
	generated := openapioperations.All()
	operations := make([]ContractOperation, 0, len(generated))
	for _, operation := range generated {
		operations = append(operations, ContractOperation{
			OwnerID:         operation.OwnerID,
			Method:          operation.Method,
			PathTemplate:    operation.PathTemplate,
			Pattern:         operation.Pattern,
			OperationID:     operation.OperationID,
			Availability:    operation.Availability,
			StateChanging:   operation.StateChanging,
			SuccessStatuses: append([]int(nil), operation.SuccessStatuses...),
			Security:        cloneSecurity(operation.Security),
		})
	}
	return operations
}

func validateContractOperation(operation ContractOperation) error {
	switch {
	case operation.OwnerID == "":
		return fmt.Errorf("generated operation %q has no owner", operation.OperationID)
	case operation.Method == "" || operation.Method != strings.ToUpper(operation.Method):
		return fmt.Errorf("generated operation %q has invalid method %q", operation.OperationID, operation.Method)
	case operation.PathTemplate == "" || !strings.HasPrefix(operation.PathTemplate, "/api/v1/"):
		return fmt.Errorf("generated operation %q has invalid path %q", operation.OperationID, operation.PathTemplate)
	case operation.Pattern != operation.Method+" "+operation.PathTemplate:
		return fmt.Errorf("generated operation %q has invalid ServeMux pattern %q", operation.OperationID, operation.Pattern)
	case operation.OperationID == "":
		return fmt.Errorf("%s has no operation ID", operation.Pattern)
	case operation.Availability == "":
		return fmt.Errorf("generated operation %q has no availability", operation.OperationID)
	case len(operation.SuccessStatuses) == 0:
		return fmt.Errorf("generated operation %q has no success status", operation.OperationID)
	}
	return nil
}

func cloneContractOperation(operation ContractOperation) ContractOperation {
	operation.SuccessStatuses = append([]int(nil), operation.SuccessStatuses...)
	operation.Security = cloneSecurity(operation.Security)
	return operation
}

func cloneSecurity(security [][]string) [][]string {
	result := make([][]string, len(security))
	for index := range security {
		result[index] = append([]string(nil), security[index]...)
	}
	return result
}

func sortContractOperations(operations []ContractOperation) {
	sort.Slice(operations, func(left, right int) bool {
		if operations[left].PathTemplate != operations[right].PathTemplate {
			return operations[left].PathTemplate < operations[right].PathTemplate
		}
		if operations[left].Method != operations[right].Method {
			return operations[left].Method < operations[right].Method
		}
		return operations[left].OperationID < operations[right].OperationID
	})
}
