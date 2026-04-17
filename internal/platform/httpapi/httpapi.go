package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/platform/jobs"
)

const RequestIDHeader = "X-Request-Id"

type DependencySet struct {
	Config      config.Config
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
	Jobs        *jobs.Manager
}

type RouteRegistrar func(*http.ServeMux, DependencySet)

type Options struct {
	Dependencies      DependencySet
	AdditionalRoutes  []RouteRegistrar
	RequestIDSequence func() string
}

type SuccessEnvelope struct {
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

type ErrorEnvelope struct {
	RequestID string       `json:"request_id"`
	Error     ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string         `json:"code"`
	Status  int            `json:"status"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

type requestIDContextKey struct{}

var requestIDSequence uint64

func NewHandler(options ...Options) (http.Handler, error) {
	var option Options
	if len(options) > 0 {
		option = options[0]
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "cartulary bootstrap server")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ready")
	})

	for _, registrar := range option.AdditionalRoutes {
		if registrar != nil {
			registrar(mux, option.Dependencies)
		}
	}

	return withRequestID(mux, option.RequestIDSequence), nil
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func WriteSuccess(w http.ResponseWriter, r *http.Request, status int, data any) error {
	return writeJSON(w, status, SuccessEnvelope{
		RequestID: RequestIDFromContext(r.Context()),
		Data:      data,
	})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any) error {
	return writeJSON(w, status, ErrorEnvelope{
		RequestID: RequestIDFromContext(r.Context()),
		Error: ErrorPayload{
			Code:    code,
			Status:  status,
			Message: message,
			Details: details,
		},
	})
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
