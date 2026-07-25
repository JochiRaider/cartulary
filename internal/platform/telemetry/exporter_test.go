package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBuildOTLPHTTPURLs(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     OTLPHTTPURLs
	}{
		{
			name:     "empty prefix",
			endpoint: "HTTPS://COLLECTOR.example.test:4318",
			want: OTLPHTTPURLs{
				Traces:  "https://collector.example.test:4318/v1/traces",
				Metrics: "https://collector.example.test:4318/v1/metrics",
				Logs:    "https://collector.example.test:4318/v1/logs",
			},
		},
		{
			name:     "root prefix",
			endpoint: "http://collector.example.test:4318/",
			want: OTLPHTTPURLs{
				Traces:  "http://collector.example.test:4318/v1/traces",
				Metrics: "http://collector.example.test:4318/v1/metrics",
				Logs:    "http://collector.example.test:4318/v1/logs",
			},
		},
		{
			name:     "path prefix",
			endpoint: "https://collector.example.test:4318/otel/prefix/",
			want: OTLPHTTPURLs{
				Traces:  "https://collector.example.test:4318/otel/prefix/v1/traces",
				Metrics: "https://collector.example.test:4318/otel/prefix/v1/metrics",
				Logs:    "https://collector.example.test:4318/otel/prefix/v1/logs",
			},
		},
		{
			name:     "ipv6 host",
			endpoint: "https://[2001:db8::1]:4318/otel",
			want: OTLPHTTPURLs{
				Traces:  "https://[2001:db8::1]:4318/otel/v1/traces",
				Metrics: "https://[2001:db8::1]:4318/otel/v1/metrics",
				Logs:    "https://[2001:db8::1]:4318/otel/v1/logs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildOTLPHTTPURLs(tt.endpoint)
			if err != nil {
				t.Fatalf("build OTLP/HTTP URLs: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected URLs:\n got: %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestBuildOTLPHTTPURLsRejectsUnsupportedEndpointShapes(t *testing.T) {
	for _, endpoint := range []string{
		"https://collector.example.test",
		"grpc://collector.example.test:4318",
		"https://user:pass@collector.example.test:4318/otel",
		"https://collector.example.test:4318/otel?signal=traces",
		"https://collector.example.test:4318/otel#fragment",
		"https://collector.example.test:4318/otel%2Fprefix",
		"https://collector.example.test:4318/otel//prefix",
		"https://collector.example.test:4318/otel/../prefix",
		"https://collector.example.test:4318/otel/./prefix",
		"https://caf\u00e9.example.test:4318/otel",
		"https://xn--caf-dma.example.test:4318/otel",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := BuildOTLPHTTPURLs(endpoint); err == nil {
				t.Fatalf("expected endpoint %q to be rejected", endpoint)
			}
		})
	}
}

func TestBuildOTLPGRPCTarget(t *testing.T) {
	tests := []struct {
		endpoint string
		want     OTLPGRPCTarget
	}{
		{
			endpoint: "https://COLLECTOR.example.test:4317",
			want:     OTLPGRPCTarget{Target: "collector.example.test:4317", Secure: true},
		},
		{
			endpoint: "http://collector.example.test:4317/",
			want:     OTLPGRPCTarget{Target: "collector.example.test:4317", Secure: false},
		},
		{
			endpoint: "https://[2001:db8::2]:4317",
			want:     OTLPGRPCTarget{Target: "[2001:db8::2]:4317", Secure: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			got, err := BuildOTLPGRPCTarget(tt.endpoint)
			if err != nil {
				t.Fatalf("build OTLP/gRPC target: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected target: got %#v want %#v", got, tt.want)
			}
		})
	}
}

func TestBuildOTLPGRPCTargetRejectsPerSignalOrPathDivergence(t *testing.T) {
	for _, endpoint := range []string{
		"https://collector.example.test:4317/v1/traces",
		"https://collector.example.test:4317/otel",
		"collector.example.test:4317",
		"https://collector.example.test",
		"https://collector.example.test:4317?signal=traces",
		"https://xn--caf-dma.example.test:4317",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := BuildOTLPGRPCTarget(endpoint); err == nil {
				t.Fatalf("expected endpoint %q to be rejected", endpoint)
			}
		})
	}
}

func TestExporterUserAgentGrammar(t *testing.T) {
	userAgent, err := ExporterUserAgent("1.2.3", "v1.38.0")
	if err != nil {
		t.Fatalf("build User-Agent: %v", err)
	}
	if userAgent != "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0" {
		t.Fatalf("unexpected User-Agent %q", userAgent)
	}

	for _, input := range []struct {
		serviceVersion  string
		exporterVersion string
	}{
		{serviceVersion: "", exporterVersion: "v1.38.0"},
		{serviceVersion: " 1.2.3 ", exporterVersion: "v1.38.0"},
		{serviceVersion: "1.2.3", exporterVersion: ""},
		{serviceVersion: "1.2.3", exporterVersion: "1.38.0"},
		{serviceVersion: "1.2.3", exporterVersion: "v1.38.0 (collector)"},
		{serviceVersion: "1.2.3", exporterVersion: "vv"},
	} {
		t.Run(input.serviceVersion+" "+input.exporterVersion, func(t *testing.T) {
			if got, err := ExporterUserAgent(input.serviceVersion, input.exporterVersion); err == nil {
				t.Fatalf("expected User-Agent input to fail, got %q", got)
			}
		})
	}

	if strings.Contains(userAgent, "(") || strings.Contains(userAgent, "collector.example") || strings.Count(userAgent, " ") != 1 {
		t.Fatalf("User-Agent contains non-adopted metadata: %q", userAgent)
	}
}

func TestOTLPStatusRetryClassification(t *testing.T) {
	for _, statusCode := range []int{429, 502, 503, 504} {
		if got := ClassifyOTLPHTTPStatus(statusCode); got != RetryTransient {
			t.Fatalf("HTTP %d should be transient, got %q", statusCode, got)
		}
		if got := DropReasonForRetryClassification(ClassifyOTLPHTTPStatus(statusCode)); got != "" {
			t.Fatalf("HTTP %d should not record permanent discard drop reason, got %q", statusCode, got)
		}
	}
	for _, statusCode := range []int{400, 401, 403, 404, 422, 500, 501} {
		if got := ClassifyOTLPHTTPStatus(statusCode); got != RetryPermanent {
			t.Fatalf("HTTP %d should be permanent, got %q", statusCode, got)
		}
		if got := DropReasonForRetryClassification(ClassifyOTLPHTTPStatus(statusCode)); got != ExporterPermanentDiscardDropReason {
			t.Fatalf("HTTP %d should record permanent discard drop reason, got %q", statusCode, got)
		}
	}

	for _, code := range []string{"CANCELLED", "DEADLINE_EXCEEDED", "ABORTED", "OUT_OF_RANGE", "UNAVAILABLE", "DATA_LOSS"} {
		if got := ClassifyOTLPGRPCStatus(code, false); got != RetryTransient {
			t.Fatalf("gRPC %s should be transient, got %q", code, got)
		}
		if got := DropReasonForRetryClassification(ClassifyOTLPGRPCStatus(code, false)); got != "" {
			t.Fatalf("gRPC %s should not record permanent discard drop reason, got %q", code, got)
		}
	}
	if got := ClassifyOTLPGRPCStatus("RESOURCE_EXHAUSTED", true); got != RetryTransient {
		t.Fatalf("RESOURCE_EXHAUSTED with retry-info should be transient, got %q", got)
	}
	if got := DropReasonForRetryClassification(ClassifyOTLPGRPCStatus("RESOURCE_EXHAUSTED", true)); got != "" {
		t.Fatalf("RESOURCE_EXHAUSTED with retry-info should not record permanent discard drop reason, got %q", got)
	}
	for _, code := range []string{"RESOURCE_EXHAUSTED", "INVALID_ARGUMENT", "UNAUTHENTICATED", "PERMISSION_DENIED", "NOT_FOUND", "UNIMPLEMENTED"} {
		if got := ClassifyOTLPGRPCStatus(code, false); got != RetryPermanent {
			t.Fatalf("gRPC %s should be permanent without retry-info, got %q", code, got)
		}
		if got := DropReasonForRetryClassification(ClassifyOTLPGRPCStatus(code, false)); got != ExporterPermanentDiscardDropReason {
			t.Fatalf("gRPC %s should record permanent discard drop reason, got %q", code, got)
		}
	}
}

func TestPlanRetryEnvelope(t *testing.T) {
	retry := telemetryconfiguration.ExporterRetryConfig{
		Enabled:           true,
		MaxElapsedMS:      30_000,
		InitialIntervalMS: 100,
		MaxIntervalMS:     5_000,
		Multiplier:        2,
	}

	tests := []struct {
		name    string
		retry   telemetryconfiguration.ExporterRetryConfig
		index   int64
		elapsed int64
		sample  int64
		stop    bool
		want    RetryPlan
	}{
		{name: "first retry within full jitter bound", retry: retry, index: 1, sample: 99, want: RetryPlan{BaseIntervalMS: 100, StartRetry: true}},
		{name: "second retry uses multiplier", retry: retry, index: 2, sample: 199, want: RetryPlan{BaseIntervalMS: 200, StartRetry: true}},
		{name: "base interval caps at max interval", retry: retry, index: 10, sample: 4999, want: RetryPlan{BaseIntervalMS: 5000, StartRetry: true}},
		{name: "sample outside bound does not start", retry: retry, index: 1, sample: 101, want: RetryPlan{BaseIntervalMS: 100}},
		{name: "max elapsed cutoff does not start", retry: retry, index: 1, elapsed: 29_950, sample: 100, want: RetryPlan{BaseIntervalMS: 100}},
		{name: "disabled retry", retry: withRetryEnabled(retry, false), index: 1, sample: 1, want: RetryPlan{}},
		{name: "zero max elapsed disables retry", retry: withRetryMaxElapsed(retry, 0), index: 1, sample: 1, want: RetryPlan{}},
		{name: "shutdown prevents new retry loop", retry: retry, index: 1, sample: 1, stop: true, want: RetryPlan{}},
		{name: "initial export attempt is not a retry", retry: retry, index: 0, sample: 1, want: RetryPlan{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanRetry(tt.retry, tt.index, tt.elapsed, tt.sample, tt.stop)
			if got != tt.want {
				t.Fatalf("unexpected retry plan: got %#v want %#v", got, tt.want)
			}
		})
	}
}

func TestRetryingSpanExporterUsesConfiguredMultiplier(t *testing.T) {
	exporter := &scriptedSpanExporter{
		errs: []error{
			status.Error(codes.Unavailable, "collector unavailable"),
			errors.New("retry-able request failure: body: busy"),
			nil,
		},
	}
	var sampledBases []time.Duration
	controller := retryController{
		settings: retrySettings{
			Enabled:         true,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     time.Second,
			MaxElapsedTime:  10 * time.Second,
			Multiplier:      3,
		},
		sampleDelay: func(max time.Duration) time.Duration {
			sampledBases = append(sampledBases, max)
			return 0
		},
		sleep: func(context.Context, time.Duration) error { return nil },
		now:   time.Now,
	}
	retrying := &retryingSpanExporter{exporter: exporter, controller: controller}

	if err := retrying.ExportSpans(context.Background(), nil); err != nil {
		t.Fatalf("retrying exporter should recover from transient errors: %v", err)
	}
	if exporter.calls != 3 {
		t.Fatalf("expected two retries after transient failures, got %d calls", exporter.calls)
	}
	want := []time.Duration{100 * time.Millisecond, 300 * time.Millisecond}
	if len(sampledBases) != len(want) {
		t.Fatalf("unexpected sampled bases: got %#v want %#v", sampledBases, want)
	}
	for i := range want {
		if sampledBases[i] != want[i] {
			t.Fatalf("retry base %d mismatch: got %s want %s", i, sampledBases[i], want[i])
		}
	}
}

func TestRetryingSpanExporterDoesNotRetryPermanentError(t *testing.T) {
	exporter := &scriptedSpanExporter{errs: []error{status.Error(codes.PermissionDenied, "denied")}}
	retrying := &retryingSpanExporter{
		exporter: exporter,
		controller: retryController{
			settings: retrySettings{
				Enabled:         true,
				InitialInterval: 100 * time.Millisecond,
				MaxInterval:     time.Second,
				MaxElapsedTime:  10 * time.Second,
				Multiplier:      2,
			},
			sampleDelay: func(max time.Duration) time.Duration {
				t.Fatalf("permanent failures must not sample a retry delay, got max %s", max)
				return 0
			},
			sleep: func(context.Context, time.Duration) error {
				t.Fatal("permanent failures must not sleep for retry")
				return nil
			},
			now: time.Now,
		},
	}

	if err := retrying.ExportSpans(context.Background(), nil); err == nil {
		t.Fatal("expected permanent exporter error")
	}
	if exporter.calls != 1 {
		t.Fatalf("permanent failure should be attempted once, got %d calls", exporter.calls)
	}
}

func TestBuildExporterRequestHeadersRedactsConfiguredSecrets(t *testing.T) {
	plan, err := BuildExporterRequestHeaders(map[string]string{
		"X-CARTULARY-API-Key": "secret token",
		"x-tenant":            "tenant secret",
	}, "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0")
	if err != nil {
		t.Fatalf("build exporter headers: %v", err)
	}
	if plan.RequestHeaders["content-type"] != "application/x-protobuf" || plan.RequestHeaders["user-agent"] == "" {
		t.Fatalf("protocol-required headers missing from request: %#v", plan.RequestHeaders)
	}
	if plan.RequestHeaders["x-cartulary-api-key"] != "secret token" || plan.RequestHeaders["x-tenant"] != "tenant secret" {
		t.Fatalf("configured headers were not canonicalized into request headers: %#v", plan.RequestHeaders)
	}
	if plan.DiagnosticHeaders["x-cartulary-api-key"] != "[redacted]" || plan.DiagnosticHeaders["x-tenant"] != "[redacted]" {
		t.Fatalf("configured secret headers were not redacted in diagnostics: %#v", plan.DiagnosticHeaders)
	}
	if ExporterDiagnosticsContainSecret(plan, "secret token", "tenant secret") {
		t.Fatalf("diagnostic headers leaked configured secret values: %#v", plan.DiagnosticHeaders)
	}
}

func TestBuildExporterRequestHeadersRejectsUnsafeShapes(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string]string
		userAgent string
	}{
		{name: "protocol-owned", headers: map[string]string{"User-Agent": "secret"}, userAgent: "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0"},
		{name: "duplicate canonical", headers: map[string]string{"x-tenant": "one", "X-Tenant": "two"}, userAgent: "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0"},
		{name: "invalid header name", headers: map[string]string{"x tenant": "secret"}, userAgent: "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0"},
		{name: "invalid header value", headers: map[string]string{"x-tenant": " secret"}, userAgent: "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.38.0"},
		{name: "unsafe user agent", headers: map[string]string{"x-tenant": "secret"}, userAgent: "Cartulary/1.2.3 (node) OTel-OTLP-Exporter-go/v1.38.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := BuildExporterRequestHeaders(tt.headers, tt.userAgent); err == nil {
				t.Fatal("expected unsafe exporter header plan to fail")
			}
		})
	}
}

func TestExportFailureMetricUsesSafeRegistryAttributes(t *testing.T) {
	name, attrs, ok := ExportFailureMetric(ExportFailure{
		SignalKind:   "traces",
		ExporterKind: "otlp_http",
		ErrorClass:   "exporter_transient",
	})
	if !ok || name != TelemetryExportFailureMetricName {
		t.Fatalf("expected export failure metric, got name=%q ok=%t", name, ok)
	}
	got := attributesByName(attrs)
	if len(got) != 3 ||
		got["cartulary.signal_kind"] != "traces" ||
		got["cartulary.telemetry.exporter_kind"] != "otlp_http" ||
		got["cartulary.error_class"] != "exporter_transient" {
		t.Fatalf("unexpected export failure attributes: %#v", got)
	}
	if _, _, ok := ExportFailureMetric(ExportFailure{
		SignalKind:   "trace-10000000-0000-4000-8000-000000000001",
		ExporterKind: "http://collector.example.test:4318",
		ErrorClass:   "select",
	}); ok {
		t.Fatal("unsafe export failure attributes must not produce a metric")
	}
	if _, _, ok := ExportFailureMetric(ExportFailure{
		SignalKind:   "traces",
		ExporterKind: "otlp_http",
		ErrorClass:   "exporter_transient",
		Recursive:    true,
	}); ok {
		t.Fatal("recursive export failure must not produce a metric")
	}
}

func TestPlanExporterAttemptTimeoutClassification(t *testing.T) {
	transient := PlanExporterAttemptTimeout(2_000, 2_000, RetryTransient)
	if !transient.TimedOut || transient.Classification != RetryTransient || transient.ProductHotPathBlocked {
		t.Fatalf("transient timeout plan mismatch: %#v", transient)
	}
	permanent := PlanExporterAttemptTimeout(2_000, 2_001, RetryPermanent)
	if !permanent.TimedOut || permanent.Classification != RetryPermanent || permanent.ProductHotPathBlocked {
		t.Fatalf("permanent timeout plan mismatch: %#v", permanent)
	}
	if got := PlanExporterAttemptTimeout(2_000, 1_999, RetryTransient); got.TimedOut || got.Classification != "" || got.ProductHotPathBlocked {
		t.Fatalf("non-timeout attempt should be empty and non-blocking: %#v", got)
	}
}

func TestOfferProcessorQueueDropNewOverflow(t *testing.T) {
	accepted := OfferProcessorQueue(ProcessorQueue{SignalKind: "metrics", MaxSize: 2, Depth: 1})
	if !accepted.Accepted || accepted.Depth != 2 || accepted.RetainedQueued != 2 || accepted.DroppedNewItem || accepted.ProductBlocked {
		t.Fatalf("accepted queue offer mismatch: %#v", accepted)
	}
	if accepted.QueueDepthMetric != TelemetryQueueDepthMetricName || accepted.QueueDepthValue != 2 || !accepted.QueueDepthCurrent {
		t.Fatalf("queue depth observation mismatch: %#v", accepted)
	}

	full := OfferProcessorQueue(ProcessorQueue{SignalKind: "logs", MaxSize: 2, Depth: 2})
	if full.Accepted || !full.DroppedNewItem || full.Depth != 2 || full.RetainedQueued != 2 || full.ProductBlocked {
		t.Fatalf("full queue must retain queued items, drop new item, and avoid product blocking: %#v", full)
	}
	if full.OverflowPolicy != "drop_new" || full.DropReason != QueueFullDropReason || full.MetricName != TelemetryItemDroppedMetricName {
		t.Fatalf("full queue drop policy mismatch: %#v", full)
	}
	got := attributesByName(full.Attributes)
	if len(got) != 2 || got["cartulary.signal_kind"] != "logs" || got["cartulary.drop_reason"] != QueueFullDropReason {
		t.Fatalf("full queue metric attributes mismatch: %#v", got)
	}

	recursive := OfferProcessorQueue(ProcessorQueue{SignalKind: "logs", MaxSize: 2, Depth: 2, Recursive: true})
	if !recursive.DroppedNewItem || recursive.MetricName != "" || len(recursive.Attributes) != 0 {
		t.Fatalf("recursive queue overflow should drop without recursive metric: %#v", recursive)
	}
}

func TestPlanShutdownTimeoutAndIdempotence(t *testing.T) {
	timedOut := PlanShutdown(ShutdownPlan{
		SignalKind:     "traces",
		FlushTimeoutMS: 5_000,
		ElapsedMS:      5_001,
		ActiveProvider: true,
	})
	if !timedOut.ContinueShutdown || !timedOut.CallShutdown || timedOut.ShutdownCallCount != 1 || !timedOut.TimedOut || timedOut.ProductBlocked {
		t.Fatalf("shutdown timeout plan mismatch: %#v", timedOut)
	}
	if timedOut.DropReason != ShutdownTimeoutDropReason || timedOut.MetricName != TelemetryItemDroppedMetricName {
		t.Fatalf("shutdown timeout drop metric mismatch: %#v", timedOut)
	}
	got := attributesByName(timedOut.Attributes)
	if got["cartulary.signal_kind"] != "traces" || got["cartulary.drop_reason"] != ShutdownTimeoutDropReason {
		t.Fatalf("shutdown timeout attributes mismatch: %#v", got)
	}

	already := PlanShutdown(ShutdownPlan{
		SignalKind:      "traces",
		FlushTimeoutMS:  5_000,
		ElapsedMS:       6_000,
		ActiveProvider:  true,
		AlreadyShutdown: true,
	})
	if !already.ContinueShutdown || already.CallShutdown || already.ShutdownCallCount != 0 || already.TimedOut {
		t.Fatalf("repeated shutdown should be idempotent: %#v", already)
	}

	recursive := PlanShutdown(ShutdownPlan{
		SignalKind:     "traces",
		FlushTimeoutMS: 5_000,
		ElapsedMS:      5_001,
		ActiveProvider: true,
		Recursive:      true,
	})
	if !recursive.TimedOut || recursive.MetricName != "" || len(recursive.Attributes) != 0 {
		t.Fatalf("recursive shutdown timeout should not record recursive metric: %#v", recursive)
	}
}

func TestPlanSelfDiagnosticRecursionGuard(t *testing.T) {
	normal := PlanSelfDiagnostic(SelfDiagnosticPlan{SignalKind: "logs"})
	if !normal.Record || normal.DropReason != "" || !normal.RecursionBound {
		t.Fatalf("non-recursive self-diagnostic should record once: %#v", normal)
	}

	guarded := PlanSelfDiagnostic(SelfDiagnosticPlan{SignalKind: "logs", Recursive: true, MetricAllowed: true})
	if guarded.Record || guarded.DropReason != RecursionGuardDropReason || guarded.MetricName != TelemetryItemDroppedMetricName || !guarded.RecursionBound {
		t.Fatalf("recursive self-diagnostic should be bounded and record safe drop metric: %#v", guarded)
	}
	got := attributesByName(guarded.Attributes)
	if got["cartulary.signal_kind"] != "logs" || got["cartulary.drop_reason"] != RecursionGuardDropReason {
		t.Fatalf("recursion guard metric attributes mismatch: %#v", got)
	}

	suppressed := PlanSelfDiagnostic(SelfDiagnosticPlan{SignalKind: "logs", Recursive: true})
	if suppressed.Record || suppressed.MetricName != "" || len(suppressed.Attributes) != 0 || suppressed.DropReason != RecursionGuardDropReason {
		t.Fatalf("recursive self-diagnostic with unsafe metric path should suppress metric: %#v", suppressed)
	}
}

func TestRuntimeInvarianceMatrixMatchesNoExportBaseline(t *testing.T) {
	matrix := RuntimeInvarianceMatrix()
	wantCount := 6 * 4
	if len(matrix) != wantCount {
		t.Fatalf("runtime invariance matrix count mismatch: got %d want %d", len(matrix), wantCount)
	}
	seen := map[string]bool{}
	for _, row := range matrix {
		key := string(row.Surface) + "/" + string(row.FailureMode)
		if seen[key] {
			t.Fatalf("duplicate runtime invariance row %s", key)
		}
		seen[key] = true
		if row.ProductResponse != "match_no_export_baseline" || row.CommittedState != "match_no_export_baseline" || row.ProductBlocked {
			t.Fatalf("runtime invariance row does not match baseline: %#v", row)
		}
	}
	for _, surface := range []RuntimeSurface{
		RuntimeSurfaceHTTPRequest,
		RuntimeSurfaceWorkbookQuery,
		RuntimeSurfaceWorkbookMutation,
		RuntimeSurfaceWebSocketSend,
		RuntimeSurfaceEvidenceAccess,
		RuntimeSurfaceBackgroundJobTransition,
	} {
		for _, failure := range []RuntimeFailureMode{
			RuntimeFailureExporterFailure,
			RuntimeFailureExporterTimeout,
			RuntimeFailureQueueOverflow,
			RuntimeFailureRedactionRejection,
		} {
			key := string(surface) + "/" + string(failure)
			if !seen[key] {
				t.Fatalf("missing runtime invariance row %s", key)
			}
		}
	}
}

type scriptedSpanExporter struct {
	errs  []error
	calls int
}

func (e *scriptedSpanExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error {
	if e.calls >= len(e.errs) {
		e.calls++
		return nil
	}
	err := e.errs[e.calls]
	e.calls++
	return err
}

func (e *scriptedSpanExporter) Shutdown(context.Context) error {
	return nil
}

func withRetryEnabled(retry telemetryconfiguration.ExporterRetryConfig, enabled bool) telemetryconfiguration.ExporterRetryConfig {
	retry.Enabled = enabled
	return retry
}

func withRetryMaxElapsed(retry telemetryconfiguration.ExporterRetryConfig, maxElapsedMS int64) telemetryconfiguration.ExporterRetryConfig {
	retry.MaxElapsedMS = maxElapsedMS
	return retry
}
