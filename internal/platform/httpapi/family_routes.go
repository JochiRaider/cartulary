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

// BindOwnerFamilyRoutes binds one complete catalog family behind its owner's
// policy. The family router selects only explicitly declared methods; policy
// runs before path canonicality, method selection and unknown descendants. No owner
// policy is inferred from route spelling or implemented by this substrate.
func BindOwnerFamilyRoutes(deps DependencySet, ownerID, profileID, root string, policy func(http.Handler) http.Handler, handlers map[string]http.HandlerFunc) error {
	registry := deps.PublicRoutes
	if registry == nil || policy == nil || root == "" {
		return errors.New("owner family requires a route registry, root and policy")
	}
	matched := false
	for _, route := range extensionRoutesFromDependencies(deps) {
		if route.ProfileID == profileID && route.RouteFamily == root && route.Claimed {
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("owner family %q is not an admitted claimed route", root)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, claimed := registry.claimedProfiles[profileID]; !claimed {
		return fmt.Errorf("owner family profile %q is not claimed", profileID)
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
		if operation.OwnerID != ownerID || operation.Availability != profileID {
			return fmt.Errorf("owner family %q overlaps another owner or availability", root)
		}
		if _, bound := registry.active[id]; bound {
			return fmt.Errorf("owner family operation %q is already bound", id)
		}
		operations[id] = operation
	}
	if len(operations) == 0 || len(operations) != len(handlers) {
		return fmt.Errorf("owner family %q requires its complete operation catalog", root)
	}
	byPath := make(map[string]map[string]http.HandlerFunc)
	for id, handler := range handlers {
		operation, known := operations[id]
		if !known || handler == nil {
			return fmt.Errorf("owner family %q has an unknown or nil operation %q", root, id)
		}
		if byPath[operation.PathTemplate] == nil {
			byPath[operation.PathTemplate] = make(map[string]http.HandlerFunc)
		}
		byPath[operation.PathTemplate][operation.Method] = withOperationMetadata(operation, handler)
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
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
	}
	guarded := policy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Classify normalized spellings for admission, but never rewrite or
		// redirect them into an operation with potentially different identity.
		if cleanFamilyRequestPath(r) != r.URL.EscapedPath() {
			http.NotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	if guarded == nil {
		return errors.New("owner family policy returned no handler")
	}
	for id, operation := range operations {
		registry.active[id] = cloneContractOperation(operation)
	}
	registry.families = append(registry.families, boundRouteFamily{profileID: profileID, root: root, handler: guarded})
	return nil
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
