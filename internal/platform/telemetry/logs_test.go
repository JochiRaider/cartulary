package telemetry

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestLogCorrelationSafeFields(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("10000000000000000000000000000001")
	spanID, _ := trace.SpanIDFromHex("2000000000000001")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	}))

	fields := LogCorrelation(ctx, LogCorrelationFields{
		Module:      "workbook",
		RouteFamily: "incidents",
		Operation:   "query",
		Result:      "success",
		ErrorCode:   "invalid_view_query",
		ErrorClass:  "request_invalid",
	})

	if fields["trace_id"] != traceID.String() || fields["span_id"] != spanID.String() || fields["trace_flags"] != "sampled" {
		t.Fatalf("missing trace fields: %#v", fields)
	}
	for _, key := range []string{"cartulary.module", "cartulary.route_family", "cartulary.operation", "cartulary.result", "cartulary.error_code", "cartulary.error_class"} {
		if fields[key] == "" {
			t.Fatalf("missing safe field %s in %#v", key, fields)
		}
	}
}

func TestLogCorrelationDropsUnsafeFields(t *testing.T) {
	fields := LogCorrelation(context.Background(), LogCorrelationFields{
		Module:      "workbook/raw",
		RouteFamily: "/api/v1/incidents/10000000",
		Operation:   "query text",
		Result:      "raw",
		ErrorCode:   "invalid sql",
		ErrorClass:  "path:/tmp/secret",
	})

	if len(fields) != 0 {
		t.Fatalf("unsafe fields must be omitted: %#v", fields)
	}
}

func TestLogBridgeDisabledDoesNotExport(t *testing.T) {
	record, ok := MapLogRecord(context.Background(), LogBridgeConfig{
		Enabled:      false,
		BodyMaxChars: 2048,
		Scope:        ScopeWorkbook,
	}, LocalLogEntry{Body: "safe local diagnostic", Severity: "info"})
	if ok {
		t.Fatalf("disabled log bridge must not export a LogRecord: %#v", record)
	}
}

func TestLogBridgeEnabledMapping(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("10000000000000000000000000000001")
	spanID, _ := trace.SpanIDFromHex("2000000000000001")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	}))
	resource, err := BuildResourceIdentity(validTelemetryBootstrapConfig(t), []string{"import"})
	if err != nil {
		t.Fatalf("build resource identity: %v", err)
	}

	observed := time.Date(2026, 6, 8, 4, 0, 0, 0, time.UTC)
	record, ok := MapLogRecord(ctx, LogBridgeConfig{
		Enabled:        true,
		BodyMaxChars:   5,
		Scope:          ScopeWorkbook,
		ServiceVersion: "1.2.3",
		Resource:       resource,
	}, LocalLogEntry{
		ObservedTimestamp: observed,
		Severity:          "warn",
		Body:              "abcdefg",
		Fields: LogCorrelationFields{
			Module:      "workbook",
			RouteFamily: "incidents",
			Operation:   "query",
			Result:      "success",
			ErrorCode:   "invalid_view_query",
			ErrorClass:  "request_invalid",
		},
	})
	if !ok {
		t.Fatal("enabled log bridge should map a safe LogRecord")
	}
	if record.ObservedTimestamp != observed || record.TraceID != traceID.String() || record.SpanID != spanID.String() || record.TraceFlags != "sampled" {
		t.Fatalf("missing top-level trace/timestamp fields: %#v", record)
	}
	if record.SeverityNumber != 13 || record.SeverityText != "WARN" || record.Severity != "WARN" {
		t.Fatalf("unexpected severity mapping: %#v", record)
	}
	if record.Body != "abcde" {
		t.Fatalf("body should be truncated after redaction by Unicode scalar count, got %q", record.Body)
	}
	if record.InstrumentationScopeName != string(ScopeWorkbook) ||
		record.InstrumentationVersion != "1.2.3" ||
		record.InstrumentationSchemaURL != "" ||
		len(record.ScopeAttributes) != 0 {
		t.Fatalf("unexpected instrumentation scope mapping: %#v", record)
	}
	if record.Resource.SchemaURL != "" || len(record.Resource.Attributes) != len(resource.Attributes) {
		t.Fatalf("unexpected resource mapping: %#v", record.Resource)
	}
	if _, ok := record.Attributes["trace_id"]; ok {
		t.Fatalf("trace id must be a top-level field, not a LogRecord attribute: %#v", record.Attributes)
	}
	for _, key := range []string{"cartulary.module", "cartulary.route_family", "cartulary.operation", "cartulary.result", "cartulary.error_code", "cartulary.error_class"} {
		if record.Attributes[key] == "" {
			t.Fatalf("missing safe LogRecord attribute %s in %#v", key, record.Attributes)
		}
	}
}

func TestLogBridgeBodyBoundsAndExceptionReduction(t *testing.T) {
	record, ok := MapLogRecord(context.Background(), LogBridgeConfig{
		Enabled:      true,
		BodyMaxChars: 0,
		Scope:        ScopeTelemetry,
	}, LocalLogEntry{Severity: "custom", Body: "safe body"})
	if !ok {
		t.Fatal("enabled log bridge should export empty body at zero bound")
	}
	if record.Body != "" || record.SeverityNumber != 9 || record.SeverityText != "INFO" {
		t.Fatalf("unexpected zero-bound body or unknown severity mapping: %#v", record)
	}

	record, ok = MapLogRecord(context.Background(), LogBridgeConfig{
		Enabled:      true,
		BodyMaxChars: 2048,
		Scope:        ScopeTelemetry,
	}, LocalLogEntry{
		Severity: "error",
		Body:     "panic exception.stacktrace select * from incidents where id='10000000-0000-4000-8000-000000000001' /tmp/secret",
		Fields: LogCorrelationFields{
			ErrorClass: "internal_error",
		},
	})
	if !ok {
		t.Fatal("unsafe exception body should be redacted, not exported raw")
	}
	if record.Body != "" {
		t.Fatalf("forbidden exception detail should be removed from LogRecord body, got %q", record.Body)
	}
	if record.Attributes["exception.message"] != "" || record.Attributes["exception.stacktrace"] != "" {
		t.Fatalf("forbidden exception fields must not be emitted: %#v", record.Attributes)
	}
	if record.Attributes["cartulary.error_class"] != "internal_error" {
		t.Fatalf("safe error class should remain available: %#v", record.Attributes)
	}
}

func TestLogBridgeOmitsEventNameExceptionParameterAndSpanEvents(t *testing.T) {
	recordType := reflect.TypeOf(LogRecord{})
	for _, forbiddenField := range []string{"EventName", "Exception"} {
		if _, ok := recordType.FieldByName(forbiddenField); ok {
			t.Fatalf("LogRecord must not expose %s", forbiddenField)
		}
	}
	if LogBridgeUsesOptionalExceptionParameter() {
		t.Fatal("log bridge must not use the optional OTel Logs API Exception parameter")
	}
	if LogBridgeCreatesSpanEvents() {
		t.Fatal("log bridge must not create span events")
	}
}
