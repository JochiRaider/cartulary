// Package logcapture supports tests of the real application log export boundary.
// Vendor log APIs and wire decoding remain owned by telemetry.
package logcapture

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	logglobal "go.opentelemetry.io/otel/log/global"
	collectlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Flush waits for the configured application log exporter without exposing its
// provider to a source-owner test.
func Flush(ctx context.Context) error {
	provider, ok := logglobal.GetLoggerProvider().(interface{ ForceFlush(context.Context) error })
	if !ok {
		return fmt.Errorf("configured log exporter is not active")
	}
	return provider.ForceFlush(ctx)
}

// DecodeExport preserves the complete exported envelope and all record fields.
// The caller owns the encoded-body bound and the total retained export bound.
func DecodeExport(body []byte, maxRecords int) (json.RawMessage, []map[string]any, error) {
	request := &collectlogs.ExportLogsServiceRequest{}
	if err := proto.Unmarshal(body, request); err != nil {
		return nil, nil, err
	}
	raw, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(request)
	if err != nil {
		return nil, nil, err
	}
	records := []map[string]any{}
	for _, resourceLogs := range request.ResourceLogs {
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			for _, record := range scopeLogs.LogRecords {
				if len(records) >= maxRecords {
					return nil, nil, fmt.Errorf("collector record overflow")
				}
				records = append(records, map[string]any{
					"scope": scopeLogs.Scope.Name, "scope_version": scopeLogs.Scope.Version, "scope_schema_url": scopeLogs.SchemaUrl, "scope_attributes": logAttributes(scopeLogs.Scope.Attributes), "scope_dropped_attributes": scopeLogs.Scope.DroppedAttributesCount,
					"resource_schema_url": resourceLogs.SchemaUrl, "resource_attributes": logAttributes(resourceLogs.Resource.Attributes), "resource_dropped_attributes": resourceLogs.Resource.DroppedAttributesCount,
					"event_name": record.EventName, "body": logValue(record.Body), "attributes": logAttributes(record.Attributes), "severity_number": int(record.SeverityNumber), "severity_text": record.SeverityText, "dropped_attributes": record.DroppedAttributesCount, "trace_id": traceID(record.TraceId, 16), "span_id": traceID(record.SpanId, 8), "trace_flags": record.Flags,
					"timestamp": time.Unix(0, int64(record.TimeUnixNano)).UTC().Format(time.RFC3339Nano), "observed_timestamp": time.Unix(0, int64(record.ObservedTimeUnixNano)).UTC().Format(time.RFC3339Nano),
				})
			}
		}
	}
	return json.RawMessage(raw), records, nil
}

func logValue(value *common.AnyValue) any {
	if value == nil {
		return nil
	}
	switch v := value.Value.(type) {
	case *common.AnyValue_StringValue:
		return v.StringValue
	case *common.AnyValue_IntValue:
		return v.IntValue
	case *common.AnyValue_DoubleValue:
		return v.DoubleValue
	case *common.AnyValue_BoolValue:
		return v.BoolValue
	case *common.AnyValue_BytesValue:
		return map[string]any{"base64": base64.StdEncoding.EncodeToString(v.BytesValue)}
	case *common.AnyValue_ArrayValue:
		out := []any{}
		for _, item := range v.ArrayValue.Values {
			out = append(out, logValue(item))
		}
		return out
	case *common.AnyValue_KvlistValue:
		return logAttributes(v.KvlistValue.Values)
	default:
		return nil
	}
}
func logAttributes(values []*common.KeyValue) map[string]any {
	out := map[string]any{}
	for _, v := range values {
		out[v.Key] = logValue(v.Value)
	}
	return out
}
func traceID(raw []byte, width int) string {
	if len(raw) == 0 {
		return strings.Repeat("0", width*2)
	}
	return hex.EncodeToString(raw)
}
