package telemetry

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const VersionUnknown = "0.0.0+unknown"

type Scope string

const (
	ScopeHTTPAPI       Scope = "cartulary.httpapi"
	ScopeWorkbook      Scope = "cartulary.workbook"
	ScopeCollaboration Scope = "cartulary.collaboration"
	ScopeEvidence      Scope = "cartulary.evidence"
	ScopeJobs          Scope = "cartulary.jobs"
	ScopeNetworkFlow   Scope = "cartulary.network_flow"
	ScopePortability   Scope = "cartulary.incident_portability"
	ScopePostgres      Scope = "cartulary.postgres"
	ScopeObjectStore   Scope = "cartulary.objectstore"
	ScopeTelemetry     Scope = "cartulary.telemetry"
)

var registeredScopes = map[Scope]struct{}{
	ScopeHTTPAPI:       {},
	ScopeWorkbook:      {},
	ScopeCollaboration: {},
	ScopeEvidence:      {},
	ScopeJobs:          {},
	ScopeNetworkFlow:   {},
	ScopePortability:   {},
	ScopePostgres:      {},
	ScopeObjectStore:   {},
	ScopeTelemetry:     {},
}

type LoggerHandle struct {
	scopeName string
	version   string
	logger    otellog.Logger
}

type LogRecord struct {
	Timestamp                time.Time
	ObservedTimestamp        time.Time
	TraceID                  string
	SpanID                   string
	TraceFlags               string
	Severity                 string
	SeverityNumber           int
	SeverityText             string
	Body                     string
	Resource                 ResourceIdentity
	InstrumentationScopeName string
	InstrumentationVersion   string
	InstrumentationSchemaURL string
	ScopeAttributes          []attribute.KeyValue
	Attributes               map[string]string
}

func RegisteredScope(scope Scope) bool {
	_, ok := registeredScopes[scope]
	return ok
}

func Tracer(scope Scope, serviceVersion string) trace.Tracer {
	if !RegisteredScope(scope) {
		return tracenoop.NewTracerProvider().Tracer(string(ScopeTelemetry), trace.WithInstrumentationVersion(resolveServiceVersion(serviceVersion)))
	}
	return otel.Tracer(string(scope), trace.WithInstrumentationVersion(resolveServiceVersion(serviceVersion)))
}

func Meter(scope Scope, serviceVersion string) metric.Meter {
	if !RegisteredScope(scope) {
		return noop.NewMeterProvider().Meter(string(ScopeTelemetry), metric.WithInstrumentationVersion(resolveServiceVersion(serviceVersion)))
	}
	return otel.Meter(string(scope), metric.WithInstrumentationVersion(resolveServiceVersion(serviceVersion)))
}

func Logger(scope Scope, serviceVersion string) LoggerHandle {
	if !RegisteredScope(scope) {
		return newLoggerHandle(ScopeTelemetry, serviceVersion)
	}
	return newLoggerHandle(scope, serviceVersion)
}

func newLoggerHandle(scope Scope, serviceVersion string) LoggerHandle {
	version := resolveServiceVersion(serviceVersion)
	return LoggerHandle{
		scopeName: string(scope),
		version:   version,
		logger:    logglobal.Logger(string(scope), otellog.WithInstrumentationVersion(version)),
	}
}

func (l LoggerHandle) ScopeName() string {
	return l.scopeName
}

func (l LoggerHandle) InstrumentationVersion() string {
	return l.version
}

func (l LoggerHandle) SchemaURL() string {
	return ""
}

func (l LoggerHandle) ScopeAttributes() []attribute.KeyValue {
	return nil
}

func (l LoggerHandle) Enabled() bool {
	if l.logger == nil {
		return false
	}
	return l.logger.Enabled(context.Background(), otellog.EnabledParameters{})
}

func (l LoggerHandle) Emit(ctx context.Context, record LogRecord) {
	if l.logger == nil {
		return
	}
	var otelRecord otellog.Record
	if !record.Timestamp.IsZero() {
		otelRecord.SetTimestamp(record.Timestamp)
	}
	if !record.ObservedTimestamp.IsZero() {
		otelRecord.SetObservedTimestamp(record.ObservedTimestamp)
	}
	if record.SeverityNumber > 0 {
		otelRecord.SetSeverity(otellog.Severity(record.SeverityNumber))
	}
	if record.SeverityText != "" {
		otelRecord.SetSeverityText(record.SeverityText)
	}
	otelRecord.SetBody(otellog.StringValue(record.Body))
	if len(record.Attributes) > 0 {
		attrs := make([]otellog.KeyValue, 0, len(record.Attributes))
		for key, value := range record.Attributes {
			if key == "" || value == "" {
				continue
			}
			attrs = append(attrs, otellog.String(key, value))
		}
		otelRecord.AddAttributes(attrs...)
	}
	l.logger.Emit(ctx, otelRecord)
}

func resolveServiceVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return VersionUnknown
	}
	return version
}
