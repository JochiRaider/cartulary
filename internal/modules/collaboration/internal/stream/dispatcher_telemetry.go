package stream

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

func (dispatcher *Dispatcher) recordDispatcherRun(ctx context.Context, processed int, err error) {
	result := "success"
	errorCode := ""
	if err != nil {
		result = "failed"
		errorCode = "database_or_delivery_failure"
	}
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.operation", "dispatch"),
		attribute.String("cartulary.result", result),
		attribute.String("cartulary.error_code", errorCode),
	)
	runs, _ := telemetry.Meter(telemetry.ScopeCollaboration, telemetry.VersionUnknown).Int64Counter(
		"cartulary.collaboration.dispatcher.runs",
		metric.WithUnit("{run}"),
		metric.WithDescription("Durable Collaboration dispatcher runs."),
	)
	events, _ := telemetry.Meter(telemetry.ScopeCollaboration, telemetry.VersionUnknown).Int64Counter(
		"cartulary.collaboration.dispatcher.events",
		metric.WithUnit("{event}"),
		metric.WithDescription("Durable Collaboration events sequenced or delivered."),
	)
	runs.Add(ctx, 1, metric.WithAttributes(attrs...))
	if processed > 0 {
		events.Add(ctx, int64(processed), metric.WithAttributes(attrs...))
	}
}
