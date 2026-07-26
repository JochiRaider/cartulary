package ws

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

func (h *Hub) ConfigureTelemetry(serviceVersion string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.serviceVersion = serviceVersion
	if h.activeGaugeRegistered {
		h.mu.Unlock()
		return
	}
	h.activeGaugeRegistered = true
	h.mu.Unlock()

	_, _ = telemetry.Meter(telemetry.ScopeCollaboration, h.telemetryServiceVersion()).Int64ObservableGauge(
		"cartulary.collaboration.connections.active",
		metric.WithUnit("{connection}"),
		metric.WithDescription("Active accepted WebSocket connections."),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(h.ActiveConnections())
			return nil
		}),
	)
}

func (h *Hub) startEventSend(eventType string) func(result string, dropReason string) {
	ctx, span := telemetry.Tracer(telemetry.ScopeCollaboration, h.telemetryServiceVersion()).Start(
		context.Background(),
		"cartulary.collaboration.event_send",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(attribute.String("cartulary.websocket.event_type", safeWebSocketEventType(eventType)))...),
	)
	return func(result string, dropReason string) {
		h.finishEventSend(ctx, span, eventType, result, dropReason)
	}
}

func (h *Hub) finishEventSend(ctx context.Context, span trace.Span, eventType string, result string, dropReason string) {
	eventType = safeWebSocketEventType(eventType)
	result = safeWebSocketResult(result)
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.websocket.event_type", eventType),
		attribute.String("cartulary.result", result),
	)
	if dropReason != "" {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.drop_reason", safeDropReason(dropReason)))...)
	}
	if result == "failed" || result == "rejected" || result == "dropped" {
		span.SetStatus(codes.Error, "")
	}
	span.SetAttributes(attrs...)
	span.End()

	counter, _ := telemetry.Meter(telemetry.ScopeCollaboration, h.telemetryServiceVersion()).Int64Counter(
		"cartulary.collaboration.events.sent",
		metric.WithUnit("{event}"),
		metric.WithDescription("WebSocket events sent."),
	)
	counter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (h *Hub) telemetryServiceVersion() string {
	if h != nil && h.serviceVersion != "" {
		return h.serviceVersion
	}
	return telemetry.VersionUnknown
}

func safeWebSocketEventType(eventType string) string {
	switch eventType {
	case "record_changed", "job_progress", "presence_delta", "presence_snapshot", "hello_ack", "resume_result", "ping", "session_revoked", "error":
		return eventType
	default:
		return "other"
	}
}

func safeWebSocketResult(result string) string {
	switch result {
	case "success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped":
		return result
	default:
		return "failed"
	}
}

func safeDropReason(reason string) string {
	switch reason {
	case "queue_full", "redaction_rejected", "exporter_permanent_discard", "shutdown_timeout", "recursion_guard", "metric_overflow":
		return reason
	default:
		return "redaction_rejected"
	}
}
