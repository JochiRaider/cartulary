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

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const RequestIDHeader = "X-Request-Id"

type DependencySet struct {
	Config      config.Config
	Env         map[string]string
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
	Jobs        *jobs.Manager
	WSHub       *platformws.Hub
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
			if match, ok := MatchReservedExtensionFamily(r.URL.Path); ok && !match.Claimed {
				_ = WriteError(w, r, http.StatusNotFound, "extension_profile_not_claimed", "extension profile not claimed", map[string]any{
					"profile_id":   match.ProfileID,
					"route_family": match.RouteFamily,
				})
				return
			}
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
			if err := registrar(mux, option.Dependencies); err != nil {
				return nil, err
			}
		}
	}

	return withRequestID(mux, option.RequestIDSequence), nil
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
	return writeJSON(w, status, ErrorEnvelope{
		Error: ErrorPayload{
			Code:      code,
			Status:    status,
			RequestID: RequestIDFromContext(r.Context()),
			Retryable: status == http.StatusConflict && code == "record_locked",
			Message:   message,
			Details:   details,
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
