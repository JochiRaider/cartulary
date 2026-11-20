package httpapi

import (
	"context"
	"net/http"
	"sync"
)

// SecurityGuard enforces Core 01 REQ-01-703 for resolved operations
// (TH-HARNESS-REQ-859): each operation authenticates exactly as its declared
// OpenAPI security states. Only harness builds and test servers set it; without
// it no request is observed and responses are unchanged. Report receives each
// mismatch once; Enforce replaces a response not yet started with 500
// internal_error.
type SecurityGuard struct {
	Enforce bool
	Report  func(SecurityMismatch)
}

// SecurityMismatch is one declared-versus-observed authentication finding.
// Declared and Observed are security classes, never credentials. Status is
// the status the operation wrote or attempted to write.
type SecurityMismatch struct {
	OperationID string
	Status      int
	Declared    string
	Observed    string
}

// Security classes of a mismatch (cartulary.harness_contract_evidence.v1).
const (
	SecurityClassPublic              = "public"
	SecurityClassSession             = "session"
	SecurityClassSessionCSRF         = "session_csrf"
	SecurityClassCredentialBootstrap = "credential_bootstrap"
)

// AuthenticationSource names the credential that established a principal.
type AuthenticationSource string

const (
	AuthenticatedBySessionCookie       AuthenticationSource = "cookie"
	AuthenticatedByBearerSession       AuthenticationSource = "bearer"
	AuthenticatedByCredentialBootstrap AuthenticationSource = "credential_bootstrap"
)

// ObserveAuthenticationAttempt records that request handling asked for a
// session or bearer session, whatever the outcome.
func ObserveAuthenticationAttempt(ctx context.Context) {
	observeSecurity(ctx, func(facts *securityFacts) { facts.attempted = true })
}

// ObservePrincipal records that request handling established a principal from
// source.
func ObservePrincipal(ctx context.Context, source AuthenticationSource) {
	observeSecurity(ctx, func(facts *securityFacts) {
		facts.attempted = true
		facts.source = source
	})
}

// ObserveCSRFVerification records that request handling checked the CSRF
// double-submit proof of a cookie request, whatever the outcome.
func ObserveCSRFVerification(ctx context.Context) {
	observeSecurity(ctx, func(facts *securityFacts) { facts.csrf = true })
}

type securityObservationContextKey struct{}

// securityFacts are the authentication facts one request establishes.
type securityFacts struct {
	attempted bool
	source    AuthenticationSource
	csrf      bool
}

func (facts securityFacts) class() string {
	switch {
	case facts.source == AuthenticatedByCredentialBootstrap:
		return SecurityClassCredentialBootstrap
	case facts.csrf:
		return SecurityClassSessionCSRF
	case facts.attempted:
		return SecurityClassSession
	default:
		return SecurityClassPublic
	}
}

// declaredSecurity is the projection of one operation's OpenAPI security
// alternatives that the guard compares against.
type declaredSecurity struct {
	operationID string
	// none: no alternative names a scheme. anonymous: an alternative admits an
	// unauthenticated request.
	none, anonymous                       bool
	cookie, cookieCSRF, bearer, bootstrap bool
}

func newDeclaredSecurity(operation ContractOperation) declaredSecurity {
	declared := declaredSecurity{operationID: operation.OperationID, none: true, anonymous: len(operation.Security) == 0}
	for _, alternative := range operation.Security {
		cookie, csrf := false, false
		for _, scheme := range alternative {
			switch scheme {
			case "sessionCookie":
				cookie = true
			case "csrfCookie", "csrfHeader":
				csrf = true
			case "bearerSession":
				declared.bearer = true
			case "credentialBootstrapBearer":
				declared.bootstrap = true
			}
		}
		declared.none = declared.none && len(alternative) == 0
		declared.anonymous = declared.anonymous || len(alternative) == 0
		declared.cookieCSRF = declared.cookieCSRF || cookie && csrf
		declared.cookie = declared.cookie || cookie && !csrf
	}
	return declared
}

// class names the declared alternative relevant to the observed source.
func (declared declaredSecurity) class(source AuthenticationSource) string {
	switch {
	case declared.none:
		return SecurityClassPublic
	case source == AuthenticatedBySessionCookie && declared.cookie && !declared.cookieCSRF:
		return SecurityClassSession
	case declared.cookieCSRF:
		return SecurityClassSessionCSRF
	case declared.cookie || declared.bearer:
		return SecurityClassSession
	case declared.bootstrap:
		return SecurityClassCredentialBootstrap
	default:
		return SecurityClassPublic
	}
}

// mismatch applies REQ-01-703 to the facts. The authentication rules apply
// to every response; success adds the rules for a status below 400.
func (declared declaredSecurity) mismatch(facts securityFacts, success bool) (string, string, bool) {
	declaredClass := declared.class(facts.source)
	switch {
	case declared.none:
		// An operation without requirements must not authenticate a session.
		return declaredClass, facts.class(), facts.attempted || facts.csrf
	case facts.csrf && !declared.cookieCSRF:
		// CSRF verified for an alternative that does not declare it.
		return declaredClass, SecurityClassSessionCSRF, true
	case facts.source == AuthenticatedBySessionCookie && !declared.cookie && !declared.cookieCSRF,
		facts.source == AuthenticatedByBearerSession && !declared.bearer,
		facts.source == AuthenticatedByCredentialBootstrap && !declared.bootstrap:
		// A principal from an undeclared scheme.
		return declaredClass, facts.class(), true
	case !success:
		return "", "", false
	case facts.source == AuthenticatedBySessionCookie && !facts.csrf && !declared.cookie:
		// The declared cookie alternative pairs CSRF, which was not verified.
		return declaredClass, SecurityClassSession, true
	case facts.source == "" && !declared.anonymous:
		// Success without a principal although requirements are declared.
		return declaredClass, SecurityClassPublic, true
	default:
		return "", "", false
	}
}

// securityObservation is the per-request authentication record. The route
// wrapper creates it only when a guard is configured. The rules are evaluated
// over all facts recorded when the response starts, the first point at which
// a response can still be replaced; facts recorded after the start are
// evaluated as they arrive and only reported. A request yields at most one
// finding.
type securityObservation struct {
	mu       sync.Mutex
	guard    *SecurityGuard
	declared *declaredSecurity
	facts    securityFacts
	reported bool
	started  bool
	// status is the response status once the response has started.
	status int
}

func observeSecurity(ctx context.Context, update func(*securityFacts)) {
	observation, _ := ctx.Value(securityObservationContextKey{}).(*securityObservation)
	if observation == nil {
		return
	}
	observation.mu.Lock()
	update(&observation.facts)
	finding := observation.evaluateLocked(false)
	observation.mu.Unlock()
	observation.deliver(finding)
}

// evaluateLocked returns the request's one finding, once, after the response
// of a resolved operation has started.
func (observation *securityObservation) evaluateLocked(success bool) *SecurityMismatch {
	if observation.declared == nil || observation.reported || !observation.started {
		return nil
	}
	declared, observed, mismatched := observation.declared.mismatch(observation.facts, success)
	if !mismatched {
		return nil
	}
	observation.reported = true
	return &SecurityMismatch{OperationID: observation.declared.operationID, Status: observation.status, Declared: declared, Observed: observed}
}

func (observation *securityObservation) deliver(finding *SecurityMismatch) {
	if finding != nil && observation.guard.Report != nil {
		observation.guard.Report(*finding)
	}
}

func withSecurityGuard(next http.Handler, guard *SecurityGuard) http.Handler {
	if guard == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observation := &securityObservation{guard: guard}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), securityObservationContextKey{}, observation)))
	})
}

// guardOperationSecurity binds a resolved operation to the request's
// observation and checks the response start. Without an observation the
// handler runs unchanged.
func guardOperationSecurity(operation ContractOperation, handler http.HandlerFunc) http.HandlerFunc {
	declared := newDeclaredSecurity(operation)
	return func(w http.ResponseWriter, r *http.Request) {
		observation, _ := r.Context().Value(securityObservationContextKey{}).(*securityObservation)
		if observation == nil {
			handler(w, r)
			return
		}
		observation.mu.Lock()
		if observation.declared != nil {
			// A nested resolution keeps the outer operation's observation.
			observation.mu.Unlock()
			handler(w, r)
			return
		}
		observation.declared = &declared
		observation.mu.Unlock()
		guarded := &securityResponseWriter{ResponseWriter: w, request: r, observation: observation, header: w.Header().Clone()}
		handler(guarded, r)
		guarded.start(http.StatusOK)
	}
}

// securityResponseWriter checks the response start against the declared
// security. An enforced mismatch replaces the response with 500
// internal_error and discards the handler's remaining output.
type securityResponseWriter struct {
	http.ResponseWriter
	request     *http.Request
	observation *securityObservation
	// header is the header set before the handler ran; an enforced
	// replacement restores it.
	header    http.Header
	started   bool
	discarded bool
}

func (w *securityResponseWriter) start(status int) {
	if w.started {
		return
	}
	w.started = true
	observation := w.observation
	observation.mu.Lock()
	observation.started, observation.status = true, status
	finding := observation.evaluateLocked(status < http.StatusBadRequest)
	observation.mu.Unlock()
	observation.deliver(finding)
	if finding == nil || !observation.guard.Enforce {
		return
	}
	w.discarded = true
	header := w.ResponseWriter.Header()
	for name := range header {
		delete(header, name)
	}
	for name, values := range w.header {
		header[name] = values
	}
	_ = WriteErrorWithOptions(w.ResponseWriter, w.request, http.StatusInternalServerError, "internal_error", "internal_error", nil, ErrorOptions{})
}

func (w *securityResponseWriter) WriteHeader(status int) {
	// An informational header precedes the response it does not start.
	if status >= http.StatusOK || status == http.StatusSwitchingProtocols {
		w.start(status)
	}
	if !w.discarded {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *securityResponseWriter) Write(data []byte) (int, error) {
	if !w.started {
		w.start(http.StatusOK)
	}
	if w.discarded {
		return len(data), nil
	}
	return w.ResponseWriter.Write(data)
}

// Flush starts the response like a write, so flushing cannot bypass the check.
func (w *securityResponseWriter) Flush() {
	if !w.started {
		w.start(http.StatusOK)
	}
	if !w.discarded {
		_ = http.NewResponseController(w.ResponseWriter).Flush()
	}
}

func (w *securityResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
