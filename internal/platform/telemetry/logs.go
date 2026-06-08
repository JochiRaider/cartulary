package telemetry

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type LogBridgeConfig struct {
	Enabled        bool
	BodyMaxChars   int
	Scope          Scope
	ServiceVersion string
	Resource       ResourceIdentity
}

type LocalLogEntry struct {
	Timestamp         time.Time
	ObservedTimestamp time.Time
	Severity          string
	Body              string
	Fields            LogCorrelationFields
}

type LogCorrelationFields struct {
	Module      string
	RouteFamily string
	Operation   string
	Result      string
	ErrorCode   string
	ErrorClass  string
}

func MapLogRecord(ctx context.Context, cfg LogBridgeConfig, entry LocalLogEntry) (LogRecord, bool) {
	if !cfg.Enabled {
		return LogRecord{}, false
	}

	observed := entry.ObservedTimestamp
	if observed.IsZero() {
		observed = time.Now().UTC()
	}
	severityNumber, severityText := logSeverity(entry.Severity)
	logger := Logger(cfg.Scope, cfg.ServiceVersion)
	correlation := LogCorrelation(ctx, entry.Fields)
	record := LogRecord{
		Timestamp:                entry.Timestamp,
		ObservedTimestamp:        observed,
		TraceID:                  correlation["trace_id"],
		SpanID:                   correlation["span_id"],
		TraceFlags:               correlation["trace_flags"],
		Severity:                 severityText,
		SeverityNumber:           severityNumber,
		SeverityText:             severityText,
		Body:                     safeLogBody(entry.Body, cfg.BodyMaxChars),
		Resource:                 copyResourceIdentity(cfg.Resource),
		InstrumentationScopeName: logger.ScopeName(),
		InstrumentationVersion:   logger.InstrumentationVersion(),
		InstrumentationSchemaURL: logger.SchemaURL(),
		ScopeAttributes:          append([]attribute.KeyValue(nil), logger.ScopeAttributes()...),
		Attributes:               logRecordAttributes(correlation),
	}
	return record, true
}

func LogBridgeUsesOptionalExceptionParameter() bool {
	return false
}

func LogBridgeCreatesSpanEvents() bool {
	return false
}

func LogCorrelation(ctx context.Context, fields LogCorrelationFields) map[string]string {
	correlation := make(map[string]string)
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		correlation["trace_id"] = spanContext.TraceID().String()
		correlation["span_id"] = spanContext.SpanID().String()
		if spanContext.IsSampled() {
			correlation["trace_flags"] = "sampled"
		}
	}
	if knownModule(fields.Module) {
		correlation["cartulary.module"] = fields.Module
	}
	if safeToken(fields.RouteFamily) {
		correlation["cartulary.route_family"] = fields.RouteFamily
	}
	if safeToken(fields.Operation) {
		correlation["cartulary.operation"] = fields.Operation
	}
	if knownResult(fields.Result) {
		correlation["cartulary.result"] = fields.Result
	}
	if safeToken(fields.ErrorCode) {
		correlation["cartulary.error_code"] = fields.ErrorCode
	}
	if knownErrorClass(fields.ErrorClass) {
		correlation["cartulary.error_class"] = fields.ErrorClass
	}
	return correlation
}

func logSeverity(severity string) (int, string) {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "trace":
		return 1, "TRACE"
	case "debug":
		return 5, "DEBUG"
	case "warn":
		return 13, "WARN"
	case "error":
		return 17, "ERROR"
	case "fatal":
		return 21, "FATAL"
	case "info":
		fallthrough
	default:
		return 9, "INFO"
	}
}

func safeLogBody(body string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	if unsafeLogBody(body) {
		return ""
	}
	runes := []rune(body)
	if len(runes) > maxChars {
		return string(runes[:maxChars])
	}
	return body
}

func unsafeLogBody(body string) bool {
	lower := strings.ToLower(body)
	return containsForbiddenValueShape(body) ||
		uuidLikePattern.MatchString(body) ||
		strings.Contains(lower, "select ") ||
		strings.Contains(lower, "insert ") ||
		strings.Contains(lower, "update ") ||
		strings.Contains(lower, "delete ") ||
		strings.Contains(lower, "exception.") ||
		strings.Contains(lower, "stacktrace") ||
		strings.Contains(lower, "panic")
}

func logRecordAttributes(correlation map[string]string) map[string]string {
	attrs := make(map[string]string)
	for key, value := range correlation {
		switch key {
		case "trace_id", "span_id", "trace_flags":
			continue
		default:
			attrs[key] = value
		}
	}
	return attrs
}

func copyResourceIdentity(resource ResourceIdentity) ResourceIdentity {
	return ResourceIdentity{
		SchemaURL:  resource.SchemaURL,
		Attributes: append([]attribute.KeyValue(nil), resource.Attributes...),
	}
}

func knownModule(module string) bool {
	switch module {
	case "httpapi", "workbook", "collaboration", "jobs", "postgres", "objectstore", "telemetry":
		return true
	default:
		return false
	}
}

func knownResult(result string) bool {
	switch result {
	case "success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped":
		return true
	default:
		return false
	}
}

func knownErrorClass(errorClass string) bool {
	switch errorClass {
	case "request_invalid", "authentication", "capability_unavailable", "concurrency_conflict", "lifecycle_conflict", "not_found",
		"expired_or_consumed", "policy_rejected", "dependency_unavailable", "invariant_violation", "timeout", "serialization_conflict",
		"constraint_violation", "exporter_transient", "exporter_permanent", "redaction_rejected", "queue_full", "shutdown_timeout",
		"recursion_guard", "internal_error":
		return true
	default:
		return false
	}
}

func safeToken(value string) bool {
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
