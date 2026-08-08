package jobs

import (
	"context"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

func (m *Manager) registerActiveGauge() {
	if m == nil || m.pool == nil || strings.TrimSpace(m.serviceVersion) == "" || m.activeGaugeRegistered {
		return
	}
	m.activeGaugeRegistered = true
	_, _ = telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion()).Int64ObservableGauge(
		"cartulary.jobs.active",
		metric.WithUnit("{job}"),
		metric.WithDescription("Active background jobs by kind."),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			return m.observeActiveJobs(ctx, observer)
		}),
	)
}

func (m *Manager) observeActiveJobs(ctx context.Context, observer metric.Int64Observer) error {
	if m == nil || m.pool == nil {
		return nil
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
	}
	counts := map[string]int64{"unknown": 0}
	definitions := map[string]ExtensionJobContract(nil)
	if m.definitions != nil {
		definitions = m.definitions.byKind
	}
	kinds := make([]string, 0, len(definitions)+1)
	for kind := range definitions {
		counts[kind] = 0
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	kinds = append(kinds, "unknown")
	rows, err := m.pool.Query(ctx, `
SELECT job_kind, COUNT(*)::bigint
  FROM jobs
 WHERE status = 'running'
 GROUP BY scope_kind
`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var storedKind *string
		var count int64
		if err := rows.Scan(&storedKind, &count); err != nil {
			return err
		}
		kind := "unknown"
		if storedKind != nil {
			kind = m.catalogJobKind(*storedKind)
		}
		if count < 0 {
			count = 0
		}
		counts[kind] += count
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, kind := range kinds {
		observer.Observe(counts[kind], metric.WithAttributes(telemetry.SafeAttributes(attribute.String("cartulary.job_kind", kind))...))
	}
	return nil
}

func (m *Manager) startJobSpan(ctx context.Context, name string, jobKind string, operation string) (context.Context, trace.Span) {
	ctx, span := telemetry.Tracer(telemetry.ScopeJobs, m.telemetryServiceVersion()).Start(
		ctx,
		name,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.job_kind", jobKind),
			attribute.String("cartulary.operation", operation),
		)...),
	)
	return ctx, span
}

func (m *Manager) finishJobSpan(span trace.Span, operation string, jobKind string, terminalStatus string, result string, err error) {
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.job_kind", jobKind),
		attribute.String("cartulary.operation", operation),
		attribute.String("cartulary.result", result),
	)
	if terminalStatus != "" {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.job_terminal_status", terminalStatus))...)
	}
	if err != nil {
		span.SetStatus(codes.Error, "")
	}
	span.SetAttributes(attrs...)
	span.End()
}

func (m *Manager) recordJobDuration(ctx context.Context, resource Resource, jobKind string, result string) {
	if resource.StartedAt == nil || resource.FinishedAt == nil || resource.FinishedAt.Before(*resource.StartedAt) {
		return
	}
	meter := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion())
	duration, _ := meter.Float64Histogram(
		"cartulary.jobs.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Background job runtime duration."),
	)
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.job_kind", jobKind),
		attribute.String("cartulary.job_terminal_status", resource.Status),
		attribute.String("cartulary.result", result),
	)
	if resource.ErrorSummary != nil && safeJobTelemetryToken(resource.ErrorSummary.Code) {
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_code", resource.ErrorSummary.Code))...)
	}
	duration.Record(ctx, resource.FinishedAt.Sub(*resource.StartedAt).Seconds(), metric.WithAttributes(attrs...))
}

func (m *Manager) telemetryServiceVersion() string {
	if m != nil && strings.TrimSpace(m.serviceVersion) != "" {
		return m.serviceVersion
	}
	return telemetry.VersionUnknown
}

func (m *Manager) catalogJobKind(jobKind string) string {
	if m == nil || !safeJobTelemetryToken(jobKind) {
		return "unknown"
	}
	if m.definitions == nil {
		return "unknown"
	}
	if _, present := m.definitions.byKind[jobKind]; !present {
		return "unknown"
	}
	return jobKind
}

func resultForJobError(err error) string {
	if err == nil {
		return "success"
	}
	return "failed"
}

func resultForTerminalStatus(status string) string {
	switch status {
	case StatusSucceeded:
		return "success"
	case StatusCanceled:
		return "canceled"
	case StatusFailed:
		return "failed"
	default:
		return "failed"
	}
}

func safeJobTelemetryToken(value string) bool {
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
