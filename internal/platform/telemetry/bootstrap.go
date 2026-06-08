package telemetry

import (
	"context"
	"crypto/tls"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	logglobal "go.opentelemetry.io/otel/log/global"
	lognoop "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

const otlpExporterGoVersion = "v1.41.0"

type Runtime struct {
	enabled           bool
	networkExport     bool
	exporterKind      string
	exporterEndpoint  string
	samplerProfile    string
	sampleRatio       float64
	resource          ResourceIdentity
	diagnosticHeaders map[string]string
	shutdowns         []func(context.Context) error
	shutdownOnce      sync.Once
}

type BootstrapOptions struct {
	ClaimedExtensionProfiles []string
}

type BootstrapOption func(*BootstrapOptions)

func WithClaimedExtensionProfiles(profileIDs []string) BootstrapOption {
	return func(options *BootstrapOptions) {
		options.ClaimedExtensionProfiles = append([]string(nil), profileIDs...)
	}
}

func Bootstrap(_ context.Context, cfg config.Config, env map[string]string, opts ...BootstrapOption) (*Runtime, error) {
	var options BootstrapOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	if err := config.ResolveTelemetrySecretReferences(cfg, env); err != nil {
		return nil, err
	}
	resolvedHeaders, err := config.ResolveTelemetryExporterHeaders(cfg, env)
	if err != nil {
		return nil, err
	}
	resource, err := BuildResourceIdentity(cfg, options.ClaimedExtensionProfiles)
	if err != nil {
		return nil, err
	}
	samplerProfile := ResolveSamplerProfile(cfg.Telemetry.Traces.SampleRatio, cfg.Telemetry.Traces.SamplerProfile)
	if !cfg.Telemetry.Enabled {
		installNoopProviders()
		return &Runtime{exporterKind: "none", samplerProfile: samplerProfile, sampleRatio: cfg.Telemetry.Traces.SampleRatio, resource: resource}, nil
	}
	runtime := &Runtime{
		enabled:        true,
		exporterKind:   "none",
		samplerProfile: samplerProfile,
		sampleRatio:    cfg.Telemetry.Traces.SampleRatio,
		resource:       resource,
	}
	if cfg.Telemetry.Exporter.Kind == "none" {
		installNoopProviders()
		return runtime, nil
	}
	if err := withContainedOTelEnvironment(func() error {
		return runtime.activateOTLP(context.Background(), cfg, resource, resolvedHeaders)
	}); err != nil {
		runtime.shutdownBuiltProviders(context.Background())
		installNoopProviders()
		return nil, err
	}
	return runtime, nil
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.shutdownOnce.Do(func() {
		_ = r.shutdownBuiltProviders(ctx)
		installNoopProviders()
	})
	return nil
}

func (r *Runtime) Enabled() bool {
	return r != nil && r.enabled
}

func (r *Runtime) NetworkExportEnabled() bool {
	return r != nil && r.networkExport
}

func (r *Runtime) ExporterKind() string {
	if r == nil || r.exporterKind == "" {
		return "none"
	}
	return r.exporterKind
}

func (r *Runtime) ExporterEndpoint() string {
	if r == nil {
		return ""
	}
	return r.exporterEndpoint
}

func (r *Runtime) ExporterDiagnosticHeaders() map[string]string {
	if r == nil {
		return nil
	}
	return copyStringMap(r.diagnosticHeaders)
}

func (r *Runtime) SamplerProfile() string {
	if r == nil || r.samplerProfile == "" {
		return SamplerProfileAlwaysOff
	}
	return r.samplerProfile
}

func (r *Runtime) SampleRatio() float64 {
	if r == nil {
		return 0
	}
	return r.sampleRatio
}

func (r *Runtime) ResourceIdentity() ResourceIdentity {
	if r == nil {
		return ResourceIdentity{}
	}
	return ResourceIdentity{
		SchemaURL:  r.resource.SchemaURL,
		Attributes: append([]attribute.KeyValue(nil), r.resource.Attributes...),
	}
}

func (r *Runtime) activateOTLP(ctx context.Context, cfg config.Config, identity ResourceIdentity, resolvedHeaders map[string]string) error {
	userAgent, err := ExporterUserAgent(cfg.Telemetry.Resource.ServiceVersion, otlpExporterGoVersion)
	if err != nil {
		return config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.exporter.user_agent",
			ReasonCode: "invalid_telemetry_config",
			Message:    err.Error(),
		})
	}
	headerPlan, err := BuildExporterRequestHeaders(resolvedHeaders, userAgent)
	if err != nil {
		return config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.exporter.headers",
			ReasonCode: "invalid_telemetry_config",
			Message:    err.Error(),
		})
	}
	res := sdkresource.NewSchemaless(identity.Attributes...)
	r.exporterKind = cfg.Telemetry.Exporter.Kind
	r.exporterEndpoint = cfg.Telemetry.Exporter.Endpoint
	r.diagnosticHeaders = copyStringMap(headerPlan.DiagnosticHeaders)

	if cfg.Telemetry.Traces.Enabled {
		traceExporter, err := newTraceExporter(ctx, cfg, headerPlan, userAgent)
		if err != nil {
			return err
		}
		traceExporter = &retryingSpanExporter{exporter: traceExporter, controller: newRetryController(cfg)}
		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdkSampler(r.samplerProfile, r.sampleRatio)),
			sdktrace.WithBatcher(traceExporter,
				sdktrace.WithMaxQueueSize(int(cfg.Telemetry.Processor.MaxQueueSize)),
				sdktrace.WithMaxExportBatchSize(int(cfg.Telemetry.Processor.MaxExportBatchSize)),
				sdktrace.WithBatchTimeout(durationMS(cfg.Telemetry.Processor.Traces.ScheduleDelayMS)),
				sdktrace.WithExportTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			),
		)
		r.shutdowns = append(r.shutdowns, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
		r.networkExport = true
	} else {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
	}

	if cfg.Telemetry.Metrics.Enabled {
		metricExporter, err := newMetricExporter(ctx, cfg, headerPlan, userAgent)
		if err != nil {
			return err
		}
		metricExporter = &retryingMetricExporter{exporter: metricExporter, controller: newRetryController(cfg)}
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithExemplarFilter(exemplar.AlwaysOffFilter),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(durationMS(cfg.Telemetry.Processor.Metrics.ScheduleDelayMS)),
				sdkmetric.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			)),
		)
		r.shutdowns = append(r.shutdowns, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
		r.networkExport = true
	} else {
		otel.SetMeterProvider(noop.NewMeterProvider())
	}

	selfMetrics := newTelemetrySelfMetrics(cfg)
	if cfg.Telemetry.Logs.BridgeEnabled {
		logExporter, err := newLogExporter(ctx, cfg, headerPlan, userAgent)
		if err != nil {
			return err
		}
		logExporter = &retryingLogExporter{exporter: logExporter, controller: newRetryController(cfg)}
		logProcessor := newDropNewLogProcessor(logExporter, dropNewLogProcessorConfig{
			MaxQueueSize:       int(cfg.Telemetry.Processor.MaxQueueSize),
			MaxExportBatchSize: int(cfg.Telemetry.Processor.MaxExportBatchSize),
			ExportInterval:     durationMS(cfg.Telemetry.Processor.Logs.ScheduleDelayMS),
			ExportTimeout:      durationMS(cfg.Telemetry.Processor.ExportTimeoutMS),
			SignalKind:         "logs",
			OnDrop:             selfMetrics.recordDrop,
		})
		loggerProvider := sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(logProcessor),
		)
		r.shutdowns = append(r.shutdowns, loggerProvider.Shutdown)
		logglobal.SetLoggerProvider(loggerProvider)
		r.networkExport = true
	} else {
		logglobal.SetLoggerProvider(lognoop.NewLoggerProvider())
	}

	return nil
}

func newTraceExporter(ctx context.Context, cfg config.Config, headerPlan ExporterHeaderPlan, userAgent string) (sdktrace.SpanExporter, error) {
	switch cfg.Telemetry.Exporter.Kind {
	case "otlp_http":
		urls, err := BuildOTLPHTTPURLs(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(urls.Traces),
			otlptracehttp.WithHeaders(sdkHTTPHeaders(headerPlan)),
			otlptracehttp.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlptracehttp.WithRetry(otlptracehttp.RetryConfig(exporterRetryConfig(cfg))),
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
		}
		return otlptracehttp.New(ctx, opts...)
	case "otlp_grpc":
		target, err := BuildOTLPGRPCTarget(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(target.Target),
			otlptracegrpc.WithHeaders(sdkGRPCHeaders(headerPlan)),
			otlptracegrpc.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig(exporterRetryConfig(cfg))),
			otlptracegrpc.WithDialOption(grpc.WithUserAgent(userAgent)),
		}
		if target.Secure {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
		} else {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlptracegrpc.WithCompressor("gzip"))
		}
		return otlptracegrpc.New(ctx, opts...)
	default:
		return nil, errors.New("unsupported telemetry exporter kind")
	}
}

func newMetricExporter(ctx context.Context, cfg config.Config, headerPlan ExporterHeaderPlan, userAgent string) (sdkmetric.Exporter, error) {
	switch cfg.Telemetry.Exporter.Kind {
	case "otlp_http":
		urls, err := BuildOTLPHTTPURLs(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpointURL(urls.Metrics),
			otlpmetrichttp.WithHeaders(sdkHTTPHeaders(headerPlan)),
			otlpmetrichttp.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig(exporterRetryConfig(cfg))),
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
		}
		return otlpmetrichttp.New(ctx, opts...)
	case "otlp_grpc":
		target, err := BuildOTLPGRPCTarget(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(target.Target),
			otlpmetricgrpc.WithHeaders(sdkGRPCHeaders(headerPlan)),
			otlpmetricgrpc.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig(exporterRetryConfig(cfg))),
			otlpmetricgrpc.WithDialOption(grpc.WithUserAgent(userAgent)),
		}
		if target.Secure {
			opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
		} else {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlpmetricgrpc.WithCompressor("gzip"))
		}
		return otlpmetricgrpc.New(ctx, opts...)
	default:
		return nil, errors.New("unsupported telemetry exporter kind")
	}
}

func newLogExporter(ctx context.Context, cfg config.Config, headerPlan ExporterHeaderPlan, userAgent string) (sdklog.Exporter, error) {
	switch cfg.Telemetry.Exporter.Kind {
	case "otlp_http":
		urls, err := BuildOTLPHTTPURLs(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpointURL(urls.Logs),
			otlploghttp.WithHeaders(sdkHTTPHeaders(headerPlan)),
			otlploghttp.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlploghttp.WithRetry(otlploghttp.RetryConfig(exporterRetryConfig(cfg))),
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlploghttp.WithCompression(otlploghttp.GzipCompression))
		}
		return otlploghttp.New(ctx, opts...)
	case "otlp_grpc":
		target, err := BuildOTLPGRPCTarget(cfg.Telemetry.Exporter.Endpoint)
		if err != nil {
			return nil, err
		}
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(target.Target),
			otlploggrpc.WithHeaders(sdkGRPCHeaders(headerPlan)),
			otlploggrpc.WithTimeout(durationMS(cfg.Telemetry.Processor.ExportTimeoutMS)),
			otlploggrpc.WithRetry(otlploggrpc.RetryConfig(exporterRetryConfig(cfg))),
			otlploggrpc.WithDialOption(grpc.WithUserAgent(userAgent)),
		}
		if target.Secure {
			opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
		} else {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		if cfg.Telemetry.Exporter.Compression == "gzip" {
			opts = append(opts, otlploggrpc.WithCompressor("gzip"))
		}
		return otlploggrpc.New(ctx, opts...)
	default:
		return nil, errors.New("unsupported telemetry exporter kind")
	}
}

func (r *Runtime) shutdownBuiltProviders(ctx context.Context) error {
	var err error
	for i := len(r.shutdowns) - 1; i >= 0; i-- {
		if r.shutdowns[i] != nil {
			err = errors.Join(err, r.shutdowns[i](ctx))
		}
	}
	r.shutdowns = nil
	return err
}

func installNoopProviders() {
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	otel.SetMeterProvider(noop.NewMeterProvider())
	logglobal.SetLoggerProvider(lognoop.NewLoggerProvider())
}

func sdkSampler(profile string, ratio float64) sdktrace.Sampler {
	switch profile {
	case SamplerProfileAlwaysOn:
		return sdktrace.AlwaysSample()
	case SamplerProfileTraceIDRatioCompat:
		return sdktrace.TraceIDRatioBased(ratio)
	default:
		return sdktrace.NeverSample()
	}
}

type otlpRetryConfig struct {
	Enabled         bool
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

func exporterRetryConfig(cfg config.Config) otlpRetryConfig {
	return otlpRetryConfig{
		Enabled:         false,
		InitialInterval: durationMS(cfg.Telemetry.Exporter.Retry.InitialIntervalMS),
		MaxInterval:     durationMS(cfg.Telemetry.Exporter.Retry.MaxIntervalMS),
		MaxElapsedTime:  durationMS(cfg.Telemetry.Exporter.Retry.MaxElapsedMS),
	}
}

func sdkHTTPHeaders(plan ExporterHeaderPlan) map[string]string {
	headers := make(map[string]string, len(plan.RequestHeaders))
	for name, value := range plan.RequestHeaders {
		switch strings.ToLower(name) {
		case "content-type":
			continue
		default:
			headers[name] = value
		}
	}
	return headers
}

func sdkGRPCHeaders(plan ExporterHeaderPlan) map[string]string {
	headers := make(map[string]string, len(plan.RequestHeaders))
	for name, value := range plan.RequestHeaders {
		switch strings.ToLower(name) {
		case "content-type", "user-agent":
			continue
		default:
			headers[name] = value
		}
	}
	return headers
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func durationMS(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

var otelEnvironmentMu sync.Mutex

func withContainedOTelEnvironment(fn func() error) error {
	otelEnvironmentMu.Lock()
	defer otelEnvironmentMu.Unlock()

	previous := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || !strings.HasPrefix(key, "OTEL_") {
			continue
		}
		previous[key] = value
		_ = os.Unsetenv(key)
	}
	defer func() {
		for _, entry := range os.Environ() {
			key, _, ok := strings.Cut(entry, "=")
			if ok && strings.HasPrefix(key, "OTEL_") {
				_ = os.Unsetenv(key)
			}
		}
		for key, value := range previous {
			_ = os.Setenv(key, value)
		}
	}()
	return fn()
}
