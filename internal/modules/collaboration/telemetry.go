package collaboration

import (
	"context"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

func (s *routeService) startWebSocketLifecycle(ctx context.Context, operation string) (context.Context, func(result string, errorCode string)) {
	operation = safeWebSocketLifecycleOperation(operation)
	ctx, span := telemetry.Tracer(telemetry.ScopeCollaboration, s.telemetryServiceVersion()).Start(
		ctx,
		"cartulary.collaboration.websocket",
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(attribute.String("cartulary.operation", operation))...),
	)
	return ctx, func(result string, errorCode string) {
		result = safeWebSocketLifecycleResult(result)
		attrs := telemetry.SafeAttributes(
			attribute.String("cartulary.operation", operation),
			attribute.String("cartulary.result", result),
		)
		if safeWebSocketLifecycleToken(errorCode) {
			attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_code", errorCode))...)
		}
		if result != "success" {
			span.SetStatus(codes.Error, "")
		}
		span.SetAttributes(attrs...)
		span.End()
	}
}

func (s *routeService) telemetryServiceVersion() string {
	if s != nil && strings.TrimSpace(s.serviceVersion) != "" {
		return s.serviceVersion
	}
	return telemetry.VersionUnknown
}

func webSocketLifecycleResultForAPIError(apiErr *httpapi.APIError) (string, string) {
	if apiErr == nil {
		return "success", ""
	}
	switch {
	case apiErr.Status == http.StatusConflict:
		return "conflict", apiErr.Code
	case apiErr.Status >= 400 && apiErr.Status < 500:
		return "rejected", apiErr.Code
	case apiErr.Status >= 500:
		return "failed", apiErr.Code
	default:
		return "failed", apiErr.Code
	}
}

func safeWebSocketLifecycleOperation(operation string) string {
	switch operation {
	case "connect":
		return operation
	default:
		return "unknown"
	}
}

func safeWebSocketLifecycleResult(result string) string {
	switch result {
	case "success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped":
		return result
	default:
		return "failed"
	}
}

func safeWebSocketLifecycleToken(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
