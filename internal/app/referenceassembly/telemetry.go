package referenceassembly

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type operationObserver struct{ serviceVersion string }

func NewOperationObserver(serviceVersion string) reference_data.OperationObserver {
	return operationObserver{serviceVersion: serviceVersion}
}

func (o operationObserver) BeginReferenceOperation(ctx context.Context, operation string) (context.Context, func(string)) {
	if !slices.Contains(reference_data.ReferenceOperationNames(), operation) {
		return ctx, func(string) {}
	}
	start := time.Now()
	ctx, span := telemetry.Tracer(telemetry.ScopeReferencePack, o.serviceVersion).Start(ctx, operation, trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(telemetry.SafeAttributes(attribute.String("cartulary.operation", operation))...))
	duration, instrumentErr := telemetry.Meter(telemetry.ScopeReferencePack, o.serviceVersion).Float64Histogram("cartulary.reference_pack.operation.duration", metric.WithUnit("s"), metric.WithDescription("Reference Pack owner operation duration."))
	var once sync.Once
	return ctx, func(result string) {
		once.Do(func() {
			if !slices.Contains([]string{"success", "rejected", "conflict", "canceled", "failed", "timeout"}, result) {
				result = "failed"
			}
			attrs := telemetry.SafeAttributes(attribute.String("cartulary.operation", operation), attribute.String("cartulary.result", result))
			span.SetAttributes(attrs...)
			if result != "success" && result != "canceled" {
				span.SetStatus(codes.Error, "")
			}
			span.End()
			if instrumentErr == nil {
				duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attrs...))
			}
		})
	}
}
