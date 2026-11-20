package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

type boundRouteFamily struct {
	profileID string
	root      string
	handler   http.Handler
}

// OwnerFamily is a claimed extension family whose owner section defines its
// own admission order (Core 01 REQ-01-707), as the reference-pack family does
// in §17.4.
type OwnerFamily struct {
	OwnerID   string
	ProfileID string
	// Root is the claimed family root; every operation under it belongs to
	// the family.
	Root string
	// Authorize is the owner's family policy. It runs after the gate's
	// session authentication, reads the principal through the authenticator's
	// accessor, and its error ends the request before any routing.
	Authorize func(*http.Request) *APIError
	// AdmitQuery, when set, is the family's route-specific query admission.
	// It runs for the selected operation after its cookie CSRF and before its
	// path parameters.
	AdmitQuery func(r *http.Request, operationID string) *APIError
	// Handlers holds one handler for every operation under Root.
	Handlers map[string]http.HandlerFunc
}

// familyGateRequirement is the gate's session authentication: no cookie CSRF,
// which each selected operation verifies by its own declaration, and no
// credential bootstrap alternative, which no family operation may declare.
var familyGateRequirement = SecurityRequirement{}

// BindAdmittedOwnerFamilyRoutes binds one complete catalog family for central
// admission in the family's own order (Core 01 REQ-01-707, §17.4). After
// reserved-unclaimed dispatch, the family gate authenticates a session through
// deps.Authenticator and applies the owner's policy, both before path
// canonicality, method selection and unknown descendants; a non-canonical path
// or an unknown descendant then receives route_not_found and an undeclared
// method method_not_allowed, which belong to no operation. The selected
// operation verifies its declared cookie CSRF on the gate's principal, runs the
// owner's query admission, then admits its path parameters before its handler.
// No owner policy is inferred from route spelling or implemented here.
func BindAdmittedOwnerFamilyRoutes(deps DependencySet, family OwnerFamily) error {
	registry, authenticator, root := deps.PublicRoutes, deps.Authenticator, family.Root
	if registry == nil || family.Authorize == nil || root == "" {
		return errors.New("owner family requires a route registry, root and policy")
	}
	if authenticator == nil {
		return fmt.Errorf("owner family %q requires an authenticator for central admission", root)
	}
	matched := false
	for _, route := range extensionRoutesFromDependencies(deps) {
		if route.ProfileID == family.ProfileID && route.RouteFamily == root && route.Claimed {
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("owner family %q is not an admitted claimed route", root)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, claimed := registry.claimedProfiles[family.ProfileID]; !claimed {
		return fmt.Errorf("owner family profile %q is not claimed", family.ProfileID)
	}
	for _, prior := range registry.families {
		if routeFamilyMatchesPath(root, prior.root) || routeFamilyMatchesPath(prior.root, root) {
			return fmt.Errorf("owner family %q overlaps a bound family", root)
		}
	}
	operations := make(map[string]ContractOperation)
	for id, operation := range registry.operationsByID {
		if !routeFamilyMatchesPath(root, operation.PathTemplate) {
			continue
		}
		if operation.OwnerID != family.OwnerID || operation.Availability != family.ProfileID {
			return fmt.Errorf("owner family %q overlaps another owner or availability", root)
		}
		if _, bound := registry.active[id]; bound {
			return fmt.Errorf("owner family operation %q is already bound", id)
		}
		if operation.OwnerGate {
			return fmt.Errorf("owner family operation %q declares an owner gate", id)
		}
		// The gate admits only a session, so every operation must require one.
		if declared := newDeclaredSecurity(operation); declared.anonymous || declared.bootstrap {
			return fmt.Errorf("owner family operation %q must require a session and no credential bootstrap", id)
		}
		operations[id] = operation
	}
	if len(operations) == 0 || len(operations) != len(family.Handlers) {
		return fmt.Errorf("owner family %q requires its complete operation catalog", root)
	}
	byPath := make(map[string]map[string]http.HandlerFunc)
	for id, handler := range family.Handlers {
		operation, known := operations[id]
		if !known || handler == nil {
			return fmt.Errorf("owner family %q has an unknown or nil operation %q", root, id)
		}
		parameters, err := compilePathParameters(operation)
		if err != nil {
			return err
		}
		if byPath[operation.PathTemplate] == nil {
			byPath[operation.PathTemplate] = make(map[string]http.HandlerFunc)
		}
		admitted := admitFamilyOperation(securityRequirementOf(operation), authenticator, family.AdmitQuery, id, parameters, handler)
		byPath[operation.PathTemplate][operation.Method] = withOperationMetadata(operation, admitted)
	}
	mux := http.NewServeMux()
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		methods := byPath[path]
		allowed := make([]string, 0, len(methods))
		for method := range methods {
			allowed = append(allowed, method)
		}
		sort.Strings(allowed)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if handler := methods[r.Method]; handler != nil {
				handler(w, r)
				return
			}
			writeMethodNotAllowed(w, r, allowed)
		})
	}
	// An unknown descendant of the family root matches no family operation.
	mux.HandleFunc("/", writeRouteNotFound)
	gate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated, apiErr := authenticator.Authenticate(r, familyGateRequirement)
		if apiErr != nil {
			WriteAPIError(w, r, apiErr)
			return
		}
		r = authenticated
		if apiErr := family.Authorize(r); apiErr != nil {
			WriteAPIError(w, r, apiErr)
			return
		}
		// Classify normalized spellings for admission, but never rewrite or
		// redirect them into an operation with potentially different identity.
		if cleanFamilyRequestPath(r) != r.URL.EscapedPath() {
			writeRouteNotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
	for id, operation := range operations {
		registry.active[id] = cloneContractOperation(operation)
		registry.modes[id] = AdmissionCentral
	}
	registry.families = append(registry.families, boundRouteFamily{profileID: family.ProfileID, root: root, handler: gate})
	return nil
}

// admitFamilyOperation admits a selected family operation after the gate:
// its declared cookie CSRF on the gate's principal, the owner's query
// admission, then its path parameters (Core 01 §17.4).
func admitFamilyOperation(requirement SecurityRequirement, authenticator Authenticator, admitQuery func(*http.Request, string) *APIError, operationID string, parameters []admittedPathParameter, handler http.HandlerFunc) http.HandlerFunc {
	admitPath := admitPathParameters(parameters, handler)
	return func(w http.ResponseWriter, r *http.Request) {
		if requirement.CookieCSRF {
			if apiErr := authenticator.VerifyCSRF(r); apiErr != nil {
				WriteAPIError(w, r, apiErr)
				return
			}
		}
		if admitQuery != nil {
			if apiErr := admitQuery(r, operationID); apiErr != nil {
				WriteAPIError(w, r, apiErr)
				return
			}
		}
		admitPath(w, r)
	}
}

func (registry *RouteRegistry) withOwnerFamilies(next http.Handler, routes []ExtensionRoute) (http.Handler, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	families := append([]boundRouteFamily(nil), registry.families...)
	for _, family := range families {
		matched := false
		for _, route := range routes {
			matched = matched || route.Claimed && route.ProfileID == family.profileID && route.RouteFamily == family.root
		}
		if !matched {
			return nil, errors.New("bound owner family is absent from the claimed route projection")
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, family := range families {
			if routeFamilyMatchesRequest(family.root, r) {
				family.handler.ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	}), nil
}

// Family admission considers both the original spelling and its structural
// cleanup. Split before unescaping so encoded slashes/dots remain segment data.
// Neither the request URL nor router-owned parameter values are modified.
func routeFamilyMatchesRequest(root string, r *http.Request) bool {
	return routeFamilyMatchesEscapedPath(root, r.URL.EscapedPath()) ||
		routeFamilyMatchesEscapedPath(root, cleanFamilyRequestPath(r))
}

func routeFamilyMatchesEscapedPath(root, escaped string) bool {
	segments := splitRouteSegments(escaped)
	for index, segment := range segments {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return false
		}
		segments[index] = decoded
	}
	return routeFamilyMatchesSegments(root, segments)
}

func cleanFamilyRequestPath(r *http.Request) string {
	escaped := r.URL.EscapedPath()
	cleaned := path.Clean("/" + strings.TrimPrefix(escaped, "/"))
	// A trailing slash is a distinct route spelling, not structural cleanup.
	if cleaned != "/" && strings.HasSuffix(escaped, "/") {
		cleaned += "/"
	}
	return cleaned
}
