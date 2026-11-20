package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/networkflowroutes"
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
	// ErrorCodes are the declared error codes by status (Core 01 REQ-01-702).
	ErrorCodes map[int][]string
	Security   [][]string
	// PathParameters are the path template's parameter declarations in
	// template order (Core 01 REQ-01-707).
	PathParameters []PathParameter
	// OwnerGate marks an operation whose owner defines its own admission
	// order (Core 01 REQ-01-707 family clause; Network Flow's NF NLSpec Table
	// 21-C): central admission runs the owner's gate after the path
	// parameters before it and before those marked AfterOwnerGate. Such an
	// operation binds only through BindAdmittedOwnerOrderRoutes.
	OwnerGate bool
	// ErrorCodeRegistry, when set, names the owner registry that governs the
	// operation's error codes instead of Core 01 REQ-01-702 declarations. Such
	// an operation is not an OpenAPI operation, declares no ErrorCodes, and is
	// outside the undeclared-code guard and its coverage records.
	ErrorCodeRegistry string
}

// PathParameter is one path parameter's declared format or pattern, its
// optional length bounds (a zero MaxLength is unbounded), and the not-found
// response a malformed value receives (Core 01 REQ-01-707). The details are
// JSON values. AfterOwnerGate places the parameter after the operation's owner
// gate.
type PathParameter struct {
	Name            string
	Format          string
	Pattern         string
	MinLength       int
	MaxLength       int
	NotFoundCode    string
	NotFoundStatus  int
	NotFoundDetails map[string]any
	AfterOwnerGate  bool
}

type OperationMetadata struct {
	OwnerID            string
	OperationID        string
	Availability       string
	CookieCSRFRequired bool
	ErrorCodes         map[int][]string
	// ErrorCodeRegistry is the operation's ContractOperation.ErrorCodeRegistry.
	ErrorCodeRegistry string
}

// GovernedByDeclaredErrorCodes reports whether Core 01 REQ-01-702 governs the
// operation's error codes, so the undeclared-code guard applies to it.
func (metadata OperationMetadata) GovernedByDeclaredErrorCodes() bool {
	return metadata.ErrorCodeRegistry == ""
}

// DeclaresErrorCode reports whether the operation declares code at status
// (Core 01 REQ-01-702). Every operation declares 500 internal_error and 503
// service_unavailable, so they need no exemption; only the pre-resolution
// codes, which no resolved operation writes, belong to no operation.
func (metadata OperationMetadata) DeclaresErrorCode(status int, code string) bool {
	for _, declared := range metadata.ErrorCodes[status] {
		if declared == code {
			return true
		}
	}
	return false
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
	modes           map[string]AdmissionMode
	families        []boundRouteFamily
	supplementary   []SupplementaryRoute
}

// SupplementaryRouteKind classifies a route the registry records although it
// is no public operation.
type SupplementaryRouteKind string

const (
	// SupplementarySocketRoute is a WebSocket upgrade route under /ws/v1/
	// (Core 01 REQ-01-708), bound with BindSocketRoute.
	SupplementarySocketRoute SupplementaryRouteKind = "socket"
	// SupplementaryTestRoute is a test-only harness route (Testing Harness
	// NLSpec TH-HARNESS-REQ-450), bound with BindTestRoute.
	SupplementaryTestRoute SupplementaryRouteKind = "test"
)

// SupplementaryRoute is one recorded socket or test-only harness route: the
// one method it serves and its path pattern.
type SupplementaryRoute struct {
	Kind   SupplementaryRouteKind
	Method string
	Path   string
}

// recordSupplementary records a socket or test route, refusing a second
// registration of the same path.
func (registry *RouteRegistry) recordSupplementary(route SupplementaryRoute) error {
	if registry == nil {
		return errors.New("public route registry is required")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	for _, prior := range registry.supplementary {
		if prior.Path == route.Path {
			return fmt.Errorf("%s route %q is already bound", route.Kind, route.Path)
		}
	}
	registry.supplementary = append(registry.supplementary, route)
	return nil
}

// SupplementaryRoutes returns every recorded socket and test-only harness
// route, by path then method, for the catalog test that no route is
// registered outside the registry.
func (registry *RouteRegistry) SupplementaryRoutes() []SupplementaryRoute {
	if registry == nil {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	routes := append([]SupplementaryRoute(nil), registry.supplementary...)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Method < routes[j].Method
	})
	return routes
}

// NewRouteRegistry returns the registry of every public operation: the
// OpenAPI catalog and the Network Flow route contract, which is a reserved
// route family outside the OpenAPI document.
func NewRouteRegistry(claims []ExtensionClaim) (*RouteRegistry, error) {
	return newRouteRegistry(append(generatedContractOperations(), networkFlowContractOperations()...), claims)
}

func newRouteRegistry(operations []ContractOperation, claims []ExtensionClaim) (*RouteRegistry, error) {
	registry := &RouteRegistry{
		operationsByID:  make(map[string]ContractOperation, len(operations)),
		operationsByKey: make(map[string]ContractOperation, len(operations)),
		claimedProfiles: make(map[string]struct{}),
		active:          make(map[string]ContractOperation),
		modes:           make(map[string]AdmissionMode),
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

// Bind binds an owner-admitted operation: route resolution (Core 01
// REQ-01-707 step 1) rejects any method but the operation's own, and the
// handler authenticates and reads its path itself.
func (registry *RouteRegistry) Bind(
	mux *http.ServeMux,
	ownerID string,
	operationID string,
	handler http.HandlerFunc,
) error {
	return registry.bind(mux, ownerID, operationID, handler, AdmissionOwner, nil, nil, false)
}

// BindAdmitted binds a centrally admitted operation (REQ-01-707): after route
// resolution, the request is authenticated by the operation's declared
// security through authenticator, then each path parameter is admitted by its
// declared format, before the handler runs. Handlers read the principal from
// the authenticator's accessor and typed parameters through PathUUID and
// PathString. An operation that admits a credential bootstrap bearer passes
// that alternative to the authenticator through SecurityRequirement.Bootstrap.
// Binding fails for a secured operation without an authenticator.
func (registry *RouteRegistry) BindAdmitted(
	mux *http.ServeMux,
	ownerID string,
	operationID string,
	handler http.HandlerFunc,
	authenticator Authenticator,
) error {
	return registry.bind(mux, ownerID, operationID, handler, AdmissionCentral, authenticator, nil, false)
}

// bind binds one operation. With noStore, every response the resolved
// operation writes carries Cache-Control: no-store, from its authentication
// and path-parameter failures to its handler's own responses. Route
// resolution's 405, which belongs to no operation (REQ-01-707), does not.
func (registry *RouteRegistry) bind(
	mux *http.ServeMux,
	ownerID string,
	operationID string,
	handler http.HandlerFunc,
	mode AdmissionMode,
	authenticator Authenticator,
	gate OwnerGate,
	noStore bool,
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
	if mode == AdmissionCentral {
		requirement := securityRequirementOf(operation)
		if !requirement.Public && authenticator == nil {
			return fmt.Errorf("public operation %q requires an authenticator for central admission", operationID)
		}
		switch {
		case operation.OwnerGate && gate == nil:
			return fmt.Errorf("public operation %q requires its owner gate (BindAdmittedOwnerOrderRoutes)", operationID)
		case !operation.OwnerGate && gate != nil:
			return fmt.Errorf("public operation %q declares no owner gate", operationID)
		}
		parameters, err := compilePathParameters(operation)
		if err != nil {
			return err
		}
		handler = admitOperation(requirement, authenticator, operationID, parameters, gate, handler)
	}
	resolved := withOperationMetadata(operation, handler)
	if noStore {
		resolved = withNoStore(resolved)
	}
	mux.HandleFunc(operation.Pattern, registry.methodNotAllowedFor(operation, resolved))
	registry.active[operationID] = cloneContractOperation(operation)
	registry.modes[operationID] = mode
	return nil
}

// withNoStore sets Cache-Control: no-store before any admission stage or the
// handler writes.
func withNoStore(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next(w, r)
	}
}

// BindOption adjusts how BindAdmittedOwnerRoutes binds an owner's operations.
type BindOption func(*bindOptions)

type bindOptions struct {
	noStore map[string]struct{}
}

// NoStore marks operations whose every resolved response carries
// Cache-Control: no-store, failures included, as Core 01 REQ-01-144B and
// REQ-01-151.1 require of the saved-view reads and workbook startup. Binding
// fails for an operation the owner does not bind.
func NoStore(operationIDs ...string) BindOption {
	return func(options *bindOptions) {
		if options.noStore == nil {
			options.noStore = make(map[string]struct{}, len(operationIDs))
		}
		for _, operationID := range operationIDs {
			options.noStore[operationID] = struct{}{}
		}
	}
}

func withOperationMetadata(operation ContractOperation, handler http.HandlerFunc) http.HandlerFunc {
	metadata := OperationMetadata{
		OwnerID: operation.OwnerID, OperationID: operation.OperationID, Availability: operation.Availability,
		ErrorCodes: cloneErrorCodes(operation.ErrorCodes), ErrorCodeRegistry: operation.ErrorCodeRegistry,
	}
	for _, alternative := range operation.Security {
		cookie, csrf := false, false
		for _, scheme := range alternative {
			cookie = cookie || scheme == "sessionCookie"
			csrf = csrf || scheme == "csrfHeader" || scheme == "csrfCookie"
		}
		metadata.CookieCSRFRequired = metadata.CookieCSRFRequired || cookie && csrf
	}
	guarded := guardOperationSecurity(operation, handler)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), operationMetadataContextKey{}, metadata)
		guarded(w, r.WithContext(ctx))
	}
}

// BindAdmittedOwnerRoutes binds an owner's operations for central admission
// (REQ-01-707) with deps.Authenticator, in operation ID order.
func BindAdmittedOwnerRoutes(
	mux *http.ServeMux,
	deps DependencySet,
	ownerID string,
	handlers map[string]http.HandlerFunc,
	options ...BindOption,
) error {
	registry, err := dependencyRouteRegistry(deps)
	if err != nil {
		return err
	}
	var settings bindOptions
	for _, option := range options {
		if option != nil {
			option(&settings)
		}
	}
	for operationID := range settings.noStore {
		if _, bound := handlers[operationID]; !bound {
			return fmt.Errorf("owner %q marks operation %q no-store without binding it", ownerID, operationID)
		}
	}
	operationIDs := make([]string, 0, len(handlers))
	for operationID := range handlers {
		operationIDs = append(operationIDs, operationID)
	}
	sort.Strings(operationIDs)
	for _, operationID := range operationIDs {
		_, noStore := settings.noStore[operationID]
		if err := registry.bind(mux, ownerID, operationID, handlers[operationID], AdmissionCentral, deps.Authenticator, nil, noStore); err != nil {
			return err
		}
	}
	return nil
}

// OwnerOrder binds an owner whose own document defines its admission order
// (Core 01 REQ-01-707 family clause): every operation declares an owner gate,
// which central admission runs after authentication and the path parameters
// before it, and before the parameters marked AfterOwnerGate. Network Flow's
// gate decides incident visibility, role and lifecycle (NF NLSpec Table 21-C).
type OwnerOrder struct {
	OwnerID  string
	Gate     OwnerGate
	Handlers map[string]http.HandlerFunc
}

// BindAdmittedOwnerOrderRoutes binds an owner-ordered owner's operations for
// central admission with deps.Authenticator, in operation ID order. Binding
// fails for an operation that declares no owner gate.
func BindAdmittedOwnerOrderRoutes(mux *http.ServeMux, deps DependencySet, order OwnerOrder) error {
	if order.Gate == nil {
		return fmt.Errorf("owner %q binds an owner order without its gate", order.OwnerID)
	}
	registry, err := dependencyRouteRegistry(deps)
	if err != nil {
		return err
	}
	operationIDs := make([]string, 0, len(order.Handlers))
	for operationID := range order.Handlers {
		operationIDs = append(operationIDs, operationID)
	}
	sort.Strings(operationIDs)
	for _, operationID := range operationIDs {
		if err := registry.bind(mux, order.OwnerID, operationID, order.Handlers[operationID], AdmissionCentral, deps.Authenticator, order.Gate, false); err != nil {
			return err
		}
	}
	return nil
}

func dependencyRouteRegistry(deps DependencySet) (*RouteRegistry, error) {
	if deps.PublicRoutes != nil {
		return deps.PublicRoutes, nil
	}
	return NewRouteRegistry(ExtensionClaimsFromDependencies(deps))
}

// AdmissionModes returns each bound operation's admission mode by operation
// ID, for the catalog test that every bound operation is admitted centrally.
func (registry *RouteRegistry) AdmissionModes() map[string]AdmissionMode {
	if registry == nil {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	modes := make(map[string]AdmissionMode, len(registry.modes))
	for operationID, mode := range registry.modes {
		modes[operationID] = mode
	}
	return modes
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
			ErrorCodes:      cloneErrorCodes(operation.ErrorCodes),
			Security:        cloneSecurity(operation.Security),
			PathParameters:  generatedPathParameters(operation.PathParameters),
		})
	}
	return operations
}

func generatedPathParameters(generated []openapioperations.PathParameter) []PathParameter {
	if len(generated) == 0 {
		return nil
	}
	parameters := make([]PathParameter, 0, len(generated))
	for _, parameter := range generated {
		var details map[string]any
		if parameter.NotFoundDetails != nil {
			details = make(map[string]any, len(parameter.NotFoundDetails))
			for key, value := range parameter.NotFoundDetails {
				details[key] = value
			}
		}
		parameters = append(parameters, PathParameter{
			Name:            parameter.Name,
			Format:          parameter.Format,
			Pattern:         parameter.Pattern,
			MinLength:       parameter.MinLength,
			MaxLength:       parameter.MaxLength,
			NotFoundCode:    parameter.NotFoundCode,
			NotFoundStatus:  parameter.NotFoundStatus,
			NotFoundDetails: details,
		})
	}
	return parameters
}

// networkFlowContractOperations projects the generated Network Flow route
// contract: a reserved route family outside the OpenAPI document whose owner
// keeps its own admission order and error registry (Core 01 REQ-01-707 family
// clause, NF NLSpec Table 21-C).
func networkFlowContractOperations() []ContractOperation {
	routes := networkflowroutes.All()
	operations := make([]ContractOperation, 0, len(routes))
	for _, route := range routes {
		var parameters []PathParameter
		for _, parameter := range route.PathParameters {
			parameters = append(parameters, PathParameter{
				Name:            parameter.Name,
				Format:          parameter.Format,
				Pattern:         parameter.Pattern,
				NotFoundCode:    parameter.NotFoundCode,
				NotFoundStatus:  parameter.NotFoundStatus,
				NotFoundDetails: parameter.NotFoundDetails,
				AfterOwnerGate:  parameter.AfterOwnerGate,
			})
		}
		operations = append(operations, ContractOperation{
			OwnerID:           route.OwnerID,
			Method:            route.Method,
			PathTemplate:      route.Path,
			Pattern:           route.Pattern,
			OperationID:       route.RouteID,
			Availability:      route.Availability,
			StateChanging:     route.StateChanging,
			SuccessStatuses:   route.SuccessStatuses,
			Security:          route.Security,
			PathParameters:    parameters,
			OwnerGate:         route.OwnerGate,
			ErrorCodeRegistry: route.ErrorRegistry,
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
	for _, parameter := range operation.PathParameters {
		if parameter.AfterOwnerGate && !operation.OwnerGate {
			return fmt.Errorf("generated operation %q admits %q after an owner gate it does not declare", operation.OperationID, parameter.Name)
		}
	}
	return nil
}

func cloneContractOperation(operation ContractOperation) ContractOperation {
	operation.SuccessStatuses = append([]int(nil), operation.SuccessStatuses...)
	operation.ErrorCodes = cloneErrorCodes(operation.ErrorCodes)
	operation.Security = cloneSecurity(operation.Security)
	operation.PathParameters = clonePathParameters(operation.PathParameters)
	return operation
}

func clonePathParameters(parameters []PathParameter) []PathParameter {
	if parameters == nil {
		return nil
	}
	result := append([]PathParameter(nil), parameters...)
	for index := range result {
		result[index].NotFoundDetails = cloneNotFoundDetails(result[index].NotFoundDetails)
	}
	return result
}

func cloneNotFoundDetails(details map[string]any) map[string]any {
	if details == nil {
		return nil
	}
	result := make(map[string]any, len(details))
	for key, value := range details {
		result[key] = cloneDetailValue(value)
	}
	return result
}

func cloneDetailValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneNotFoundDetails(typed)
	case []any:
		result := make([]any, len(typed))
		for index := range typed {
			result[index] = cloneDetailValue(typed[index])
		}
		return result
	default:
		return typed
	}
}

func cloneErrorCodes(declared map[int][]string) map[int][]string {
	result := make(map[int][]string, len(declared))
	for status, codes := range declared {
		result[status] = append([]string(nil), codes...)
	}
	return result
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
