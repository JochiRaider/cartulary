package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/metric"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

type telemetrySelfMetrics struct {
	enabled     bool
	dropCounter metric.Int64Counter
}

func newTelemetrySelfMetrics(cfg config.Config) telemetrySelfMetrics {
	if !cfg.Telemetry.SelfDiagnostics.Enabled || !cfg.Telemetry.Metrics.Enabled {
		return telemetrySelfMetrics{}
	}
	counter, err := Meter(ScopeTelemetry, cfg.Telemetry.Resource.ServiceVersion).Int64Counter(
		TelemetryItemDroppedMetricName,
		metric.WithUnit("{item}"),
		metric.WithDescription("Telemetry items dropped before or during export."),
	)
	if err != nil {
		return telemetrySelfMetrics{}
	}
	return telemetrySelfMetrics{enabled: true, dropCounter: counter}
}

func (m telemetrySelfMetrics) recordDrop(ctx context.Context, signalKind string, reason string) {
	if !m.enabled {
		return
	}
	name, attrs, ok := dropMetric(signalKind, reason, signalKind == string(ScopeTelemetry))
	if !ok || name != TelemetryItemDroppedMetricName {
		return
	}
	m.dropCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
