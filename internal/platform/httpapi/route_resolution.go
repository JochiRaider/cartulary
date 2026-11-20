package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Route-resolution rejections (Core 01 REQ-01-707 step 1). They precede
// resolution and belong to no operation, so they are written without
// operation metadata and outside the undeclared-code rule (REQ-01-702).
const (
	routeNotFoundCode    = "route_not_found"
	methodNotAllowedCode = "method_not_allowed"
)

// Handler returns the public HTTP surface of mux for REQ-01-707 route
// resolution. It registers the /api catch-alls, which answer a path that
// matches no bound operation with 404 route_not_found and a path that matches
// one only under other methods with 405 method_not_allowed. The /ws/v1
// catch-alls answer a path that matches no WebSocket route bound with
// BindSocketRoute with 404 route_not_found (REQ-01-708 step 1). A
// non-canonical /api or /ws/v1 path, one with an empty, "." or ".." segment,
// matches nothing and is never redirected. Other paths keep the mux's
// behaviour. Call it once per mux, after every operation is bound.
func (registry *RouteRegistry) Handler(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/api/", registry.serveUnmatched)
	mux.HandleFunc("/api", registry.serveUnmatched)
	mux.HandleFunc(socketRouteRoot+"/", writeRouteNotFound)
	mux.HandleFunc(socketRouteRoot, writeRouteNotFound)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (isAPIRequest(r) || isSocketRouteRequest(r)) && cleanFamilyRequestPath(r) != r.URL.EscapedPath() {
			writeRouteNotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func isAPIRequest(r *http.Request) bool {
	escaped := r.URL.EscapedPath()
	return escaped == "/api" || strings.HasPrefix(escaped, "/api/")
}

// socketRouteRoot is the WebSocket root whose route resolution Core 01
// REQ-01-708 step 1 defines.
const socketRouteRoot = "/ws/v1"

func isSocketRouteRequest(r *http.Request) bool {
	escaped := r.URL.EscapedPath()
	return escaped == socketRouteRoot || strings.HasPrefix(escaped, socketRouteRoot+"/")
}

// BindSocketRoute binds a WebSocket upgrade route under /ws/v1/ with Core 01
// REQ-01-708 step 1: only GET reaches handler. Any other method on the path,
// HEAD included, receives 405 method_not_allowed with Allow: GET before the
// handler runs, so before its Origin check and authentication. A path no
// bound route matches receives 404 route_not_found from Handler. The route is
// recorded in deps' route registry, so the catalog test sees it.
func BindSocketRoute(mux *http.ServeMux, deps DependencySet, path string, handler http.HandlerFunc) error {
	if mux == nil || handler == nil || !strings.HasPrefix(path, socketRouteRoot+"/") || strings.ContainsAny(path, " \t") {
		return fmt.Errorf("invalid WebSocket route %q", path)
	}
	registry, err := dependencyRouteRegistry(deps)
	if err != nil {
		return err
	}
	if err := registry.recordSupplementary(SupplementaryRoute{Kind: SupplementarySocketRoute, Method: http.MethodGet, Path: path}); err != nil {
		return err
	}
	allowed := []string{http.MethodGet}
	// ServeMux routes HEAD to a GET pattern; REQ-01-708 treats it as another
	// method.
	mux.HandleFunc(http.MethodGet+" "+path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, r, allowed)
			return
		}
		handler(w, r)
	})
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		writeMethodNotAllowed(w, r, allowed)
	})
	return nil
}

// serveUnmatched answers an /api request no bound operation pattern selected.
func (registry *RouteRegistry) serveUnmatched(w http.ResponseWriter, r *http.Request) {
	if allowed := registry.allowedMethods(r.URL.EscapedPath()); len(allowed) > 0 {
		writeMethodNotAllowed(w, r, allowed)
		return
	}
	writeRouteNotFound(w, r)
}

// allowedMethods returns, in ascending order, the methods of every active
// operation whose path template matches the escaped path. Like ServeMux, it
// splits at literal slashes and decodes each segment once, so an encoded
// slash stays inside its segment; a template parameter matches any non-empty
// segment.
func (registry *RouteRegistry) allowedMethods(escapedPath string) []string {
	segments, ok := decodedPathSegments(escapedPath)
	if !ok {
		return nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	methods := make(map[string]struct{})
	for _, operation := range registry.active {
		if templateMatchesSegments(operation.PathTemplate, segments) {
			methods[operation.Method] = struct{}{}
		}
	}
	allowed := make([]string, 0, len(methods))
	for method := range methods {
		allowed = append(allowed, method)
	}
	sort.Strings(allowed)
	return allowed
}

func decodedPathSegments(escapedPath string) ([]string, bool) {
	if !strings.HasPrefix(escapedPath, "/") {
		return nil, false
	}
	segments := strings.Split(escapedPath[1:], "/")
	for index, segment := range segments {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return nil, false
		}
		segments[index] = decoded
	}
	return segments, true
}

func templateMatchesSegments(template string, segments []string) bool {
	templateSegments := strings.Split(strings.TrimPrefix(template, "/"), "/")
	if len(templateSegments) != len(segments) {
		return false
	}
	for index, templateSegment := range templateSegments {
		if isRouteTemplateSegment(templateSegment) {
			if segments[index] == "" {
				return false
			}
			continue
		}
		if templateSegment != segments[index] {
			return false
		}
	}
	return true
}

// methodNotAllowedFor rejects a request whose method is not the bound
// operation's own; ServeMux routes HEAD to a GET pattern, which REQ-01-707
// treats as an undeclared method.
func (registry *RouteRegistry) methodNotAllowedFor(operation ContractOperation, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == operation.Method {
			next(w, r)
			return
		}
		allowed := registry.allowedMethods(r.URL.EscapedPath())
		if len(allowed) == 0 {
			allowed = []string{operation.Method}
		}
		writeMethodNotAllowed(w, r, allowed)
	}
}

func writeRouteNotFound(w http.ResponseWriter, r *http.Request) {
	writeResolutionRejection(w, r, http.StatusNotFound, routeNotFoundCode, map[string]any{})
}

// writeMethodNotAllowed names the declared methods in the Allow header and,
// in the same order, in error.details.allowed_methods.
func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request, allowed []string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	methods := make([]any, 0, len(allowed))
	for _, method := range allowed {
		methods = append(methods, method)
	}
	writeResolutionRejection(w, r, http.StatusMethodNotAllowed, methodNotAllowedCode, map[string]any{"allowed_methods": methods})
}

// writeResolutionRejection writes the common envelope; a HEAD response carries
// its headers only.
func writeResolutionRejection(w http.ResponseWriter, r *http.Request, status int, code string, details map[string]any) {
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		return
	}
	_ = WriteErrorWithOptions(w, r, status, code, code, details, ErrorOptions{})
}
