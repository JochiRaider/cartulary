package telemetry

import (
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func HTTPMiddleware(next http.Handler, serviceVersion string) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}

	tracer := Tracer(ScopeHTTPAPI, serviceVersion)
	meter := Meter(ScopeHTTPAPI, serviceVersion)
	duration, _ := meter.Float64Histogram(
		"cartulary.http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("HTTP server request lifecycle duration."),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		method := safeHTTPMethod(r.Method)
		routeTemplate, routeFamily := safeHTTPRoute(r.URL.Path)
		ctx, span := tracer.Start(
			r.Context(),
			method+" "+routeTemplate,
			trace.WithNewRoot(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(SafeAttributes(
				attribute.String("http.request.method", method),
				attribute.String("http.route", routeTemplate),
				attribute.String("cartulary.route_family", routeFamily),
			)...),
		)
		recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		result := resultForStatus(recorder.status)
		attrs := SafeAttributes(
			attribute.String("http.request.method", method),
			attribute.Int("http.response.status_code", recorder.status),
			attribute.String("http.route", routeTemplate),
			attribute.String("cartulary.route_family", routeFamily),
			attribute.String("cartulary.result", result),
		)
		span.SetAttributes(attrs...)
		if recorder.status >= 400 {
			span.SetStatus(codes.Error, "")
		}
		span.End()
		duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	})
}

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseStatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseStatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func safeHTTPMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return "UNKNOWN"
	}
	for _, r := range method {
		if r < 'A' || r > 'Z' {
			return "OTHER"
		}
	}
	return method
}

func safeHTTPRoute(path string) (string, string) {
	switch {
	case path == "/":
		return "/", "web.root"
	case path == "/healthz":
		return "/healthz", "health"
	case path == "/readyz":
		return "/readyz", "readiness"
	case strings.HasPrefix(path, "/assets/"):
		return "/assets/{asset}", "web.asset"
	case strings.HasPrefix(path, "/api/v1/auth/"):
		return "/api/v1/auth/{operation}", "auth"
	case strings.HasPrefix(path, "/api/v1/incidents/"):
		return "/api/v1/incidents/{incident_route}", "incidents"
	case strings.HasPrefix(path, "/api/v1/records/"):
		return "/api/v1/records/{record_route}", "records"
	case strings.HasPrefix(path, "/api/v1/jobs/"):
		return "/api/v1/jobs/{job_route}", "jobs"
	case strings.HasPrefix(path, "/api/v1/view-schemas"):
		return "/api/v1/view-schemas/{view_schema_route}", "view_schemas"
	case strings.HasPrefix(path, "/api/v1/"):
		return "/api/v1/{route_family}", "api"
	case strings.HasPrefix(path, "/ws/v1/"):
		return "/ws/v1/{route_family}", "websocket"
	default:
		return "/{unmatched}", "unmatched"
	}
}

func resultForStatus(status int) string {
	switch {
	case status >= 200 && status < 400:
		return "success"
	case status == http.StatusConflict:
		return "conflict"
	case status >= 400 && status < 500:
		return "rejected"
	default:
		return "failed"
	}
}
