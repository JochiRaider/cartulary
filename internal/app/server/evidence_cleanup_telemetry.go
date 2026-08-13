package server

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

type evidenceCleanupTelemetryObserver struct {
	operations metric.Int64Counter
	duration   metric.Float64Histogram
	logger     telemetry.LoggerHandle

	mu             sync.RWMutex
	healthObserved bool
	overdue        int64
	oldestAge      time.Duration
}

func newEvidenceCleanupTelemetryObserver(serviceVersion string) (*evidenceCleanupTelemetryObserver, error) {
	meter := telemetry.Meter(telemetry.ScopeEvidence, serviceVersion)
	observer := &evidenceCleanupTelemetryObserver{
		logger: telemetry.Logger(telemetry.ScopeEvidence, serviceVersion),
	}
	operations, err := meter.Int64Counter(
		telemetry.EvidenceCleanupOperationsMetricName,
		metric.WithUnit("{operation}"),
		metric.WithDescription("Evidence cleanup sweep operations by closed result."),
	)
	if err != nil {
		return nil, err
	}
	duration, err := meter.Float64Histogram(
		telemetry.EvidenceCleanupSweepDurationMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Evidence cleanup sweep duration."),
	)
	if err != nil {
		return nil, err
	}
	if _, err := meter.Int64ObservableGauge(
		telemetry.EvidenceCleanupOverdueMetricName,
		metric.WithUnit("{blob}"),
		metric.WithDescription("Current failed unattached Evidence blobs beyond the deletion deadline."),
		metric.WithInt64Callback(func(_ context.Context, metricObserver metric.Int64Observer) error {
			observer.mu.RLock()
			defer observer.mu.RUnlock()
			if observer.healthObserved {
				metricObserver.Observe(observer.overdue)
			}
			return nil
		}),
	); err != nil {
		return nil, err
	}
	if _, err := meter.Float64ObservableGauge(
		telemetry.EvidenceCleanupOldestAgeMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Age of the oldest currently eligible failed unattached Evidence blob."),
		metric.WithFloat64Callback(func(_ context.Context, metricObserver metric.Float64Observer) error {
			observer.mu.RLock()
			defer observer.mu.RUnlock()
			if observer.healthObserved {
				metricObserver.Observe(observer.oldestAge.Seconds())
			}
			return nil
		}),
	); err != nil {
		return nil, err
	}
	observer.operations = operations
	observer.duration = duration
	return observer, nil
}

func (observer *evidenceCleanupTelemetryObserver) ObserveCleanupSweep(
	ctx context.Context,
	observation evidence.CleanupSweepObservation,
) {
	if observer == nil {
		return
	}
	operation, result, errorClass := closedEvidenceCleanupTelemetry(observation)
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.operation", operation),
		attribute.String("cartulary.result", result),
	)
	if errorClass != "" {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_class", errorClass))...)
	}
	observer.operations.Add(ctx, 1, metric.WithAttributes(attrs...))
	observer.duration.Record(ctx, max(observation.Duration.Seconds(), 0), metric.WithAttributes(attrs...))
	if observation.HealthSnapshotValid {
		overdue := max(observation.OverdueBlobCount, 0)
		oldestAge := max(observation.OldestEligibleBlobAge, 0)
		observer.mu.Lock()
		observer.healthObserved = true
		observer.overdue = overdue
		observer.oldestAge = oldestAge
		observer.mu.Unlock()
	}
	observer.logger.Emit(ctx, telemetry.LogRecord{
		Timestamp:      time.Now().UTC(),
		Severity:       cleanupLogSeverity(result),
		SeverityNumber: cleanupLogSeverityNumber(result),
		SeverityText:   cleanupLogSeverityText(result),
		Body:           "Evidence cleanup sweep completed.",
		Attributes:     cleanupLogAttributes(operation, result, errorClass),
	})
}

func closedEvidenceCleanupTelemetry(observation evidence.CleanupSweepObservation) (string, string, string) {
	operation := "cleanup_sweep"
	result := observation.Result
	switch result {
	case "success", "canceled", "failed", "timeout":
	default:
		result = "failed"
	}
	errorClass := observation.ErrorClass
	switch errorClass {
	case "", "dependency_unavailable", "timeout", "internal_error":
	default:
		errorClass = "internal_error"
	}
	if result == "success" {
		errorClass = ""
	} else if errorClass == "" {
		errorClass = "internal_error"
	}
	return operation, result, errorClass
}

func cleanupLogAttributes(operation string, result string, errorClass string) map[string]string {
	attrs := map[string]string{
		"cartulary.module":    "evidence",
		"cartulary.operation": operation,
		"cartulary.result":    result,
	}
	if errorClass != "" {
		attrs["cartulary.error_class"] = errorClass
	}
	return attrs
}

func cleanupLogSeverity(result string) string {
	if result == "success" {
		return "info"
	}
	return "warn"
}

func cleanupLogSeverityNumber(result string) int {
	if result == "success" {
		return 9
	}
	return 13
}

func cleanupLogSeverityText(result string) string {
	if result == "success" {
		return "INFO"
	}
	return "WARN"
}

var _ evidence.CleanupObserver = (*evidenceCleanupTelemetryObserver)(nil)
