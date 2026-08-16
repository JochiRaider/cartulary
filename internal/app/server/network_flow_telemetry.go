package server

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

type networkFlowTelemetryObserver struct {
	serviceVersion string
	phaseDuration  metric.Float64Histogram
	rows           metric.Int64Histogram
	objects        metric.Int64Histogram
	cleanupOps     metric.Int64Counter
	cleanupTime    metric.Float64Histogram
	cleanupDeleted metric.Int64Counter
	logger         telemetry.LoggerHandle

	mu                    sync.RWMutex
	cleanupHealthObserved bool
	eligibleBacklog       int64
	oldestAgeObserved     bool
	oldestEligibleAge     time.Duration
}

func newNetworkFlowTelemetryObserver(serviceVersion string) *networkFlowTelemetryObserver {
	meter := telemetry.Meter(telemetry.ScopeNetworkFlow, serviceVersion)
	observer := &networkFlowTelemetryObserver{
		serviceVersion: serviceVersion,
		logger:         telemetry.Logger(telemetry.ScopeNetworkFlow, serviceVersion),
	}
	observer.phaseDuration, _ = meter.Float64Histogram(
		telemetry.NetworkFlowGraphPhaseDurationMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Network Flow graph materialization phase duration."),
	)
	observer.rows, _ = meter.Int64Histogram(
		telemetry.NetworkFlowGraphRowsMetricName,
		metric.WithUnit("{row}"),
		metric.WithDescription("Accepted source rows contributing to one graph result."),
	)
	observer.objects, _ = meter.Int64Histogram(
		telemetry.NetworkFlowGraphObjectsMetricName,
		metric.WithUnit("{object}"),
		metric.WithDescription("Network Flow graph result objects by closed kind."),
	)
	observer.cleanupOps, _ = meter.Int64Counter(
		telemetry.NetworkFlowCleanupOperationsMetricName,
		metric.WithUnit("{operation}"),
		metric.WithDescription("Network Flow cleanup sweep operations by closed result."),
	)
	observer.cleanupTime, _ = meter.Float64Histogram(
		telemetry.NetworkFlowCleanupDurationMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Network Flow cleanup sweep duration."),
	)
	observer.cleanupDeleted, _ = meter.Int64Counter(
		telemetry.NetworkFlowCleanupDeletedMetricName,
		metric.WithUnit("{object}"),
		metric.WithDescription("Expired leases and eligible projection results deleted by Network Flow cleanup."),
	)
	_, _ = meter.Int64ObservableGauge(
		telemetry.NetworkFlowCleanupEligibleMetricName,
		metric.WithUnit("{result}"),
		metric.WithDescription("Current eligible projection-result backlog."),
		metric.WithInt64Callback(func(_ context.Context, metricObserver metric.Int64Observer) error {
			observer.mu.RLock()
			defer observer.mu.RUnlock()
			if observer.cleanupHealthObserved {
				metricObserver.Observe(observer.eligibleBacklog)
			}
			return nil
		}),
	)
	_, _ = meter.Float64ObservableGauge(
		telemetry.NetworkFlowCleanupOldestAgeMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Age from published_at of the oldest currently eligible projection result."),
		metric.WithFloat64Callback(func(_ context.Context, metricObserver metric.Float64Observer) error {
			observer.mu.RLock()
			defer observer.mu.RUnlock()
			if observer.cleanupHealthObserved && observer.oldestAgeObserved {
				metricObserver.Observe(observer.oldestEligibleAge.Seconds())
			}
			return nil
		}),
	)
	return observer
}

func (observer *networkFlowTelemetryObserver) ObserveGraphPhase(
	ctx context.Context,
	observation networkflow.GraphPhaseTelemetryObservation,
) {
	if observer == nil || !networkFlowGraphPhase(observation.Phase) || !networkFlowGraphMode(observation.GraphMode) {
		return
	}
	result, errorClass := closedNetworkFlowTelemetryOutcome(observation.Result, observation.ErrorClass)
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.phase", observation.Phase),
		attribute.String("cartulary.graph_mode", observation.GraphMode),
		attribute.String("cartulary.result", result),
	)
	if errorClass != "" {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_class", errorClass))...)
	}
	duration := max(observation.Duration, 0)
	if observer.phaseDuration != nil {
		observer.phaseDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}
	spanAttrs := telemetry.SafeAttributes(append([]attribute.KeyValue{
		attribute.String("cartulary.operation", "graph_materialization"),
	}, attrs...)...)
	observer.recordSpan(ctx, "cartulary.network_flow.graph.phase", duration, result, spanAttrs)
	observer.emitStaticLog(ctx, "Network Flow graph phase completed.", result, map[string]string{
		"cartulary.module": "network_flow", "cartulary.operation": "graph_materialization",
		"cartulary.phase": observation.Phase, "cartulary.graph_mode": observation.GraphMode,
		"cartulary.result": result, "cartulary.error_class": errorClass,
	})
}

func (observer *networkFlowTelemetryObserver) ObserveGraphResult(
	ctx context.Context,
	observation networkflow.GraphResultTelemetryObservation,
) {
	if observer == nil || !networkFlowGraphMode(observation.GraphMode) ||
		observation.ContributingRows < 0 || observation.Vertices < 0 || observation.Edges < 0 || observation.TimeBuckets < 0 {
		return
	}
	result, _ := closedNetworkFlowTelemetryOutcome(observation.Result, "")
	base := telemetry.SafeAttributes(
		attribute.String("cartulary.graph_mode", observation.GraphMode),
		attribute.String("cartulary.result", result),
	)
	if observer.rows != nil {
		observer.rows.Record(ctx, int64(observation.ContributingRows), metric.WithAttributes(base...))
	}
	observer.recordGraphObjects(ctx, base, "vertex", observation.Vertices)
	observer.recordGraphObjects(ctx, base, "edge", observation.Edges)
	if observation.GraphMode == "time_bucket_v1" {
		observer.recordGraphObjects(ctx, base, "time_bucket", observation.TimeBuckets)
	}
}

func (observer *networkFlowTelemetryObserver) ObserveGraphCleanup(
	ctx context.Context,
	observation networkflow.GraphCleanupTelemetryObservation,
) {
	if observer == nil || observation.DeletedLeases < 0 || observation.DeletedResults < 0 {
		return
	}
	result, errorClass := closedNetworkFlowTelemetryOutcome(observation.Result, observation.ErrorClass)
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.operation", "cleanup_sweep"),
		attribute.String("cartulary.result", result),
	)
	if errorClass != "" {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_class", errorClass))...)
	}
	duration := max(observation.Duration, 0)
	if observer.cleanupOps != nil {
		observer.cleanupOps.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
	if observer.cleanupTime != nil {
		observer.cleanupTime.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}
	observer.recordCleanupDeleted(ctx, "lease", observation.DeletedLeases, result)
	observer.recordCleanupDeleted(ctx, "projection_result", observation.DeletedResults, result)
	observer.updateCleanupHealth(observation)
	spanAttrs := telemetry.SafeAttributes(append(attrs,
		attribute.String("cartulary.phase", "cleanup_sweep"),
	)...)
	observer.recordSpan(ctx, "cartulary.network_flow.cleanup", duration, result, spanAttrs)
	observer.emitStaticLog(ctx, "Network Flow cleanup sweep completed.", result, map[string]string{
		"cartulary.module": "network_flow", "cartulary.operation": "cleanup_sweep",
		"cartulary.phase": "cleanup_sweep", "cartulary.result": result,
		"cartulary.error_class": errorClass,
	})
}

func (observer *networkFlowTelemetryObserver) recordGraphObjects(
	ctx context.Context,
	base []attribute.KeyValue,
	kind string,
	count int,
) {
	if observer.objects == nil {
		return
	}
	attrs := telemetry.SafeAttributes(append(base, attribute.String("cartulary.graph_object_kind", kind))...)
	observer.objects.Record(ctx, int64(count), metric.WithAttributes(attrs...))
}

func (observer *networkFlowTelemetryObserver) recordCleanupDeleted(ctx context.Context, kind string, count int, result string) {
	if observer.cleanupDeleted == nil || count <= 0 {
		return
	}
	observer.cleanupDeleted.Add(ctx, int64(count), metric.WithAttributes(telemetry.SafeAttributes(
		attribute.String("cartulary.graph_object_kind", kind),
		attribute.String("cartulary.result", result),
	)...))
}

func (observer *networkFlowTelemetryObserver) updateCleanupHealth(observation networkflow.GraphCleanupTelemetryObservation) {
	if !observation.HealthSnapshotValid || observation.EligibleResultBacklog < 0 ||
		(observation.EligibleResultBacklog == 0 && observation.OldestEligibleResultAge != nil) ||
		(observation.OldestEligibleResultAge != nil && *observation.OldestEligibleResultAge < 0) {
		return
	}
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.cleanupHealthObserved = true
	observer.eligibleBacklog = observation.EligibleResultBacklog
	observer.oldestAgeObserved = observation.OldestEligibleResultAge != nil
	observer.oldestEligibleAge = 0
	if observation.OldestEligibleResultAge != nil {
		observer.oldestEligibleAge = *observation.OldestEligibleResultAge
	}
}

func (observer *networkFlowTelemetryObserver) recordSpan(
	ctx context.Context,
	name string,
	duration time.Duration,
	result string,
	attrs []attribute.KeyValue,
) {
	finishedAt := time.Now().UTC()
	_, span := telemetry.Tracer(telemetry.ScopeNetworkFlow, observer.serviceVersion).Start(
		ctx, name,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithTimestamp(finishedAt.Add(-duration)),
		trace.WithAttributes(attrs...),
	)
	if result != "success" {
		span.SetStatus(codes.Error, "")
	}
	span.End(trace.WithTimestamp(finishedAt))
}

func (observer *networkFlowTelemetryObserver) emitStaticLog(
	ctx context.Context,
	body string,
	result string,
	attrs map[string]string,
) {
	for key, value := range attrs {
		if value == "" {
			delete(attrs, key)
		}
	}
	severity, number := "warn", 13
	if result == "success" {
		severity, number = "info", 9
	}
	observer.logger.Emit(ctx, telemetry.LogRecord{
		Timestamp: time.Now().UTC(), Severity: severity, SeverityNumber: number,
		SeverityText: map[bool]string{true: "INFO", false: "WARN"}[result == "success"],
		Body:         body, Attributes: attrs,
	})
}

func closedNetworkFlowTelemetryOutcome(result string, errorClass string) (string, string) {
	switch result {
	case "success", "rejected", "conflict", "canceled", "failed", "timeout":
	default:
		result = "failed"
	}
	if result == "success" || result == "canceled" {
		return result, ""
	}
	switch errorClass {
	case "request_invalid", "concurrency_conflict", "lifecycle_conflict", "not_found", "policy_rejected",
		"dependency_unavailable", "invariant_violation", "timeout", "serialization_conflict", "constraint_violation", "internal_error":
	default:
		errorClass = "internal_error"
	}
	return result, errorClass
}

func networkFlowGraphPhase(value string) bool {
	switch value {
	case "source_validation", "source_scan", "projection", "publication":
		return true
	default:
		return false
	}
}

func networkFlowGraphMode(value string) bool {
	return value == "default_flow_edge_v1" || value == "time_bucket_v1"
}

var _ networkflow.GraphTelemetryObserver = (*networkFlowTelemetryObserver)(nil)
