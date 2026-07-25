package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const RequestIDHeader = "X-Request-Id"

type DependencySet struct {
	Config              config.Config
	Env                 map[string]string
	Postgres            *pgxpool.Pool
	PostgresDB          postgres.DB
	ObjectStore         objectstore.Store
	Jobs                *jobs.Manager
	JobRunner           *jobs.Runner
	WSHub               *platformws.Hub
	CursorCodec         *pagination.Codec
	ExtensionDiscovery  ExtensionDiscoveryProvider
	ExtensionClaims     ExtensionClaimProvider
	ExtensionRoutes     ExtensionRouteProvider
	ExtensionWorkspaces ExtensionWorkspaceProvider
	Readiness           ReadinessChecker
	Admission           AdmissionGate
	PublicErrorFaults   PublicErrorFaultStore
	ModuleOverrides     map[string]any
	Now                 func() time.Time
}

func (deps DependencySet) PostgresHandle() postgres.DB {
	if deps.PostgresDB != nil {
		return deps.PostgresDB
	}
	return deps.Postgres
}

type RouteRegistrar func(*http.ServeMux, DependencySet) error

type Options struct {
	Dependencies      DependencySet
	AdditionalRoutes  []RouteRegistrar
	RequestIDSequence func() string
}

type EnvelopeMeta struct {
	RequestID string      `json:"request_id"`
	Paging    *PagingMeta `json:"paging,omitempty"`
	Query     any         `json:"query,omitempty"`
}

type PagingMeta struct {
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

type SuccessEnvelope struct {
	Data any          `json:"data"`
	Meta EnvelopeMeta `json:"meta"`
}

type ErrorEnvelope struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code      string         `json:"code"`
	Status    int            `json:"status"`
	RequestID string         `json:"request_id"`
	Retryable bool           `json:"retryable"`
	Message   string         `json:"message,omitempty"`
	Details   map[string]any `json:"details"`
	Conflict  any            `json:"conflict,omitempty"`
}

type PublicErrorFault struct {
	Status    int
	Code      string
	Message   string
	Retryable bool
	Details   map[string]any
}

type PublicErrorFaultStore interface {
	ConsumePublicErrorFault(method string, path string) (PublicErrorFault, bool)
}

type AdmissionGate interface {
	AdmissionOpen() bool
	FatalReason() string
}

type requestIDContextKey struct{}

var requestIDSequence uint64

func NewHandler(options ...Options) (http.Handler, error) {
	var option Options
	if len(options) > 0 {
		option = options[0]
	}
	if err := requireExtensionProjections(option.Dependencies); err != nil {
		return nil, err
	}
	rootHTML, err := webassets.ReadBrowserRootHTML()
	if err != nil {
		return nil, fmt.Errorf("load embedded web root: %w", err)
	}
	staticFS, err := webassets.StaticFS()
	if err != nil {
		return nil, fmt.Errorf("load embedded web assets: %w", err)
	}
	readiness := option.Dependencies.Readiness
	if readiness == nil {
		readiness = NewDependencyReadinessChecker(option.Dependencies.Postgres, option.Dependencies.ObjectStore)
	}

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/", http.FileServerFS(staticFS)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(rootHTML))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if admission := option.Dependencies.Admission; admission != nil && !admission.AdmissionOpen() {
			status := ReadinessStatusStartingDependencyProbe
			if admission.FatalReason() != "" {
				status = ReadinessStatusFatalIntegrityFailure
			}
			_ = writeJSON(w, http.StatusServiceUnavailable, ReadinessState{SchemaID: ReadinessSchemaID, Status: status})
			return
		}
		state := readiness.CheckReadiness(r.Context())
		status := http.StatusServiceUnavailable
		if state.Status == ReadinessStatusReady {
			status = http.StatusOK
		}
		_ = writeJSON(w, status, state)
	})

	for _, registrar := range option.AdditionalRoutes {
		if registrar != nil {
			if err := registrar(mux, option.Dependencies); err != nil {
				return nil, err
			}
		}
	}

	handler := http.Handler(mux)
	handler = withUnclaimedReservedExtensionFamilies(handler, extensionRoutesFromDependencies(option.Dependencies))
	if option.Dependencies.PublicErrorFaults != nil {
		handler = withPublicErrorFaults(handler, option.Dependencies.PublicErrorFaults)
	}
	if option.Dependencies.Config.Telemetry.Enabled {
		handler = telemetry.HTTPMiddleware(handler, option.Dependencies.Config.Telemetry.Resource.ServiceVersion)
	}
	handler = withAdmissionGate(handler, option.Dependencies.Admission)

	return withRequestID(handler, option.RequestIDSequence), nil
}

func withAdmissionGate(next http.Handler, admission AdmissionGate) http.Handler {
	if admission == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if admission.AdmissionOpen() || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		reasonCode := "extension_publication_pending"
		if admission.FatalReason() != "" {
			reasonCode = "extension_integrity_failure"
		}
		_ = WriteErrorWithOptions(w, r, http.StatusServiceUnavailable, "service_unavailable", "service unavailable", map[string]any{
			"reason_code": reasonCode,
		}, ErrorOptions{Retryable: true})
	})
}

func withUnclaimedReservedExtensionFamilies(next http.Handler, routes []ExtensionRoute) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			if match, ok := MatchReservedExtensionRouteIn(routes, r.URL.Path); ok && !match.Claimed {
				_ = WriteError(w, r, http.StatusNotFound, "extension_profile_not_claimed", "extension profile not claimed", map[string]any{
					"profile_id":   match.ProfileID,
					"route_family": match.RouteFamily,
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func ExtensionDiscoveryFromDependencies(deps DependencySet) []ExtensionProfile {
	if deps.ExtensionDiscovery == nil {
		return nil
	}
	return cloneExtensionProfiles(deps.ExtensionDiscovery.ExtensionDiscoveryProfiles())
}

func ExtensionClaimsFromDependencies(deps DependencySet) []ExtensionClaim {
	if deps.ExtensionClaims == nil {
		return nil
	}
	return cloneExtensionClaims(deps.ExtensionClaims.ExtensionClaims())
}

func extensionRoutesFromDependencies(deps DependencySet) []ExtensionRoute {
	if deps.ExtensionRoutes == nil {
		return nil
	}
	return cloneExtensionRoutes(deps.ExtensionRoutes.ExtensionRoutes())
}

func ExtensionWorkspacesFromDependencies(deps DependencySet) []ExtensionWorkspacePublication {
	if deps.ExtensionWorkspaces == nil {
		return nil
	}
	return cloneExtensionWorkspaces(deps.ExtensionWorkspaces.ExtensionWorkspaces())
}

func requireExtensionProjections(deps DependencySet) error {
	missing := make([]string, 0, 4)
	if deps.ExtensionDiscovery == nil {
		missing = append(missing, "discovery")
	}
	if deps.ExtensionClaims == nil {
		missing = append(missing, "claims")
	}
	if deps.ExtensionRoutes == nil {
		missing = append(missing, "routes")
	}
	if deps.ExtensionWorkspaces == nil {
		missing = append(missing, "workspaces")
	}
	if len(missing) != 0 {
		return fmt.Errorf("http extension projections must be explicit: missing %s", strings.Join(missing, ", "))
	}
	return nil
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func WriteSuccess(w http.ResponseWriter, r *http.Request, status int, data any) error {
	return WriteSuccessWithMeta(w, r, status, data, EnvelopeMeta{
		RequestID: RequestIDFromContext(r.Context()),
	})
}

func WriteSuccessWithPaging(w http.ResponseWriter, r *http.Request, status int, data any, paging PagingMeta) error {
	return WriteSuccessWithMeta(w, r, status, data, EnvelopeMeta{
		RequestID: RequestIDFromContext(r.Context()),
		Paging:    &paging,
	})
}

func WriteSuccessWithMeta(w http.ResponseWriter, r *http.Request, status int, data any, meta EnvelopeMeta) error {
	return writeJSON(w, status, SuccessEnvelope{
		Data: data,
		Meta: meta,
	})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any) error {
	return WriteErrorWithConflict(w, r, status, code, message, details, nil)
}

func WriteErrorWithConflict(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any, conflict any) error {
	return WriteErrorWithOptions(w, r, status, code, message, details, ErrorOptions{
		Conflict:  conflict,
		Retryable: status == http.StatusConflict && code == "record_locked",
	})
}

type ErrorOptions struct {
	Conflict  any
	Retryable bool
}

func WriteErrorWithOptions(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any, options ErrorOptions) error {
	if details == nil {
		details = map[string]any{}
	}
	return writeJSON(w, status, ErrorEnvelope{
		Error: ErrorPayload{
			Code:      code,
			Status:    status,
			RequestID: RequestIDFromContext(r.Context()),
			Retryable: options.Retryable,
			Message:   message,
			Details:   details,
			Conflict:  options.Conflict,
		},
	})
}

func withPublicErrorFaults(next http.Handler, faults PublicErrorFaultStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isOrdinaryPublicAPIRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		fault, ok := faults.ConsumePublicErrorFault(r.Method, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		_ = WriteErrorWithOptions(w, r, fault.Status, fault.Code, fault.Message, fault.Details, ErrorOptions{
			Retryable: fault.Retryable,
		})
	})
}

func isOrdinaryPublicAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/") && !strings.HasPrefix(path, "/api/v1/test/")
}

func withRequestID(next http.Handler, generator func() string) http.Handler {
	if generator == nil {
		generator = nextRequestID
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generator()
		}

		w.Header().Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func nextRequestID() string {
	value := atomic.AddUint64(&requestIDSequence, 1)
	return "req-" + strconv.FormatUint(value, 10)
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(payload)
}
