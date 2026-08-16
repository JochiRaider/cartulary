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

func (m *Manager) registerJobGauges() {
	if m == nil || m.pool == nil || strings.TrimSpace(m.serviceVersion) == "" || m.jobGaugesRegistered {
		return
	}
	m.jobGaugesRegistered = true
	meter := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion())
	_, _ = meter.Int64ObservableGauge(
		"cartulary.jobs.active",
		metric.WithUnit("{job}"),
		metric.WithDescription("Active background jobs by kind."),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			return m.observeActiveJobs(ctx, observer)
		}),
	)
	_, _ = meter.Int64ObservableGauge(
		telemetry.JobsQueuedMetricName,
		metric.WithUnit("{job}"),
		metric.WithDescription("Queued background jobs, including retry-delayed jobs, by kind."),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			return m.observeQueuedJobs(ctx, observer)
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
	definitions := map[string]Definition(nil)
	if m.catalog != nil {
		definitions = m.catalog.byKind
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
 GROUP BY job_kind
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

func (m *Manager) observeQueuedJobs(ctx context.Context, observer metric.Int64Observer) error {
	if m == nil || m.pool == nil {
		return nil
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
	}
	counts, kinds := m.emptyCatalogJobCounts()
	rows, err := m.pool.Query(ctx, `
SELECT job_kind, COUNT(*)::bigint
  FROM jobs
 WHERE status IN ('queued', 'running')
   AND handler_attempt_id IS NULL
   AND handler_failure_count < $1
 GROUP BY job_kind
`, m.policy.MaximumFailures)
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
			return ErrInvalidTransition
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

func (m *Manager) emptyCatalogJobCounts() (map[string]int64, []string) {
	counts := map[string]int64{"unknown": 0}
	definitions := map[string]Definition(nil)
	if m != nil && m.catalog != nil {
		definitions = m.catalog.byKind
	}
	kinds := make([]string, 0, len(definitions)+1)
	for kind := range definitions {
		counts[kind] = 0
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return counts, append(kinds, "unknown")
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

func (m *Manager) recordQueueWait(ctx context.Context, jobKind string, duration time.Duration) {
	if duration < 0 {
		return
	}
	histogram, _ := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion()).Float64Histogram(
		telemetry.JobsQueueWaitDurationMetricName,
		metric.WithUnit("s"),
		metric.WithDescription("Eligibility-to-successful-claim duration."),
	)
	histogram.Record(ctx, duration.Seconds(), metric.WithAttributes(telemetry.SafeAttributes(
		attribute.String("cartulary.job_kind", m.catalogJobKind(jobKind)),
	)...))
}

func (m *Manager) recordAttempt(ctx context.Context, jobKind string, result string) {
	counter, _ := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion()).Int64Counter(
		"cartulary.jobs.attempts",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Completed handler attempts by kind and closed outcome."),
	)
	counter.Add(ctx, 1, metric.WithAttributes(telemetry.SafeAttributes(
		attribute.String("cartulary.job_kind", m.catalogJobKind(jobKind)),
		attribute.String("cartulary.result", result),
	)...))
}

func (m *Manager) recordLeaseRenewalFailure(ctx context.Context, jobKind string, result string) {
	counter, _ := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion()).Int64Counter(
		"cartulary.jobs.lease_renewal.failures",
		metric.WithUnit("{failure}"),
		metric.WithDescription("Failed lease renewals by kind and closed outcome."),
	)
	counter.Add(ctx, 1, metric.WithAttributes(telemetry.SafeAttributes(
		attribute.String("cartulary.job_kind", m.catalogJobKind(jobKind)),
		attribute.String("cartulary.result", result),
	)...))
}

func (m *Manager) recordExpiredJobs(ctx context.Context, counts map[string]int64) {
	if len(counts) == 0 {
		return
	}
	counter, _ := telemetry.Meter(telemetry.ScopeJobs, m.telemetryServiceVersion()).Int64Counter(
		"cartulary.jobs.expired",
		metric.WithUnit("{job}"),
		metric.WithDescription("Job resources compacted after logical expiry."),
	)
	for jobKind, count := range counts {
		if count <= 0 {
			continue
		}
		counter.Add(ctx, count, metric.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.job_kind", m.catalogJobKind(jobKind)),
		)...))
	}
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
	if m.catalog == nil {
		return "unknown"
	}
	if _, present := m.catalog.byKind[jobKind]; !present {
		return "unknown"
	}
	return jobKind
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
