package config

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestOpenTelemetryConfigDefaultsAndClosedNamespace(t *testing.T) {
	cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), nil)

	if !cfg.Telemetry.Enabled {
		t.Fatal("telemetry.enabled should default to true")
	}
	if cfg.Telemetry.Exporter.Kind != "none" {
		t.Fatalf("unexpected exporter kind default: got %q", cfg.Telemetry.Exporter.Kind)
	}
	if cfg.Telemetry.Exporter.Endpoint != "" {
		t.Fatalf("telemetry exporter endpoint must default to omitted, got %q", cfg.Telemetry.Exporter.Endpoint)
	}
	if cfg.Telemetry.Resource.ServiceName != "cartulary.app" || cfg.Telemetry.Resource.ServiceNamespace != "cartulary" {
		t.Fatalf("unexpected resource defaults: service=%q namespace=%q", cfg.Telemetry.Resource.ServiceName, cfg.Telemetry.Resource.ServiceNamespace)
	}
	if cfg.Telemetry.Resource.ServiceVersion != "0.0.0+unknown" {
		t.Fatalf("unexpected service version default: %q", cfg.Telemetry.Resource.ServiceVersion)
	}
	if cfg.Telemetry.Resource.ServiceInstanceID == "" {
		t.Fatal("service_instance_id default must be generated")
	}
	requireTelemetryUUIDV4(t, cfg.Telemetry.Resource.ServiceInstanceID)

	err := loadInvalidConfig(t, string(fixturesConfigValid(t))+"\n[telemetry.unregistered]\nvalue = true\n", nil)
	requireDiagnostic(t, err, "telemetry.unregistered.value", "unknown_key")
}

func TestOpenTelemetryEnvironmentBindingParser(t *testing.T) {
	t.Run("empty env values are omitted for telemetry keys", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND":                "",
			"CARTULARY__TELEMETRY__TRACES__SAMPLE_RATIO":          "",
			"CARTULARY__TELEMETRY__RESOURCE__SERVICE_INSTANCE_ID": "",
		})
		if cfg.Telemetry.Exporter.Kind != "none" || cfg.Telemetry.Traces.SampleRatio != 0.10 {
			t.Fatalf("empty telemetry env overlay should use defaults, got kind=%q ratio=%f", cfg.Telemetry.Exporter.Kind, cfg.Telemetry.Traces.SampleRatio)
		}
		if cfg.Telemetry.Resource.ServiceInstanceID == "" {
			t.Fatal("empty service_instance_id env overlay should allow generated default")
		}
	})

	t.Run("rejects string null overlays", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": "null",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")
	})

	t.Run("rejects signed integers and exponent decimals", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__PROCESSOR__MAX_QUEUE_SIZE": "-1",
			"CARTULARY__TELEMETRY__TRACES__SAMPLE_RATIO":      "1e-1",
		})
		requireDiagnostic(t, err, "telemetry.processor.max_queue_size", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.traces.sample_ratio", "invalid_telemetry_config")
	})

	t.Run("parses header map overlays as secret refs", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "Authorization=otel-token,tenant_id=tenant.ref",
		})
		if cfg.Telemetry.Exporter.Headers["authorization"].Name != "otel-token" {
			t.Fatalf("authorization header secret ref not parsed: %#v", cfg.Telemetry.Exporter.Headers)
		}
		if cfg.Telemetry.Exporter.Headers["tenant_id"].Kind != "env" {
			t.Fatalf("tenant header secret ref kind not parsed: %#v", cfg.Telemetry.Exporter.Headers)
		}
	})

	t.Run("rejects hostile header overlays before secret resolution", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "Authorization=one,authorization=two",
		})
		requireDiagnostic(t, err, "telemetry.exporter.headers", "invalid_telemetry_config")

		err = loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "content-type=otel-token",
		})
		requireDiagnostic(t, err, "telemetry.exporter.headers.content-type", "invalid_telemetry_config")

		err = loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__ATTRIBUTE__HMAC_SECRET_REF": "bad/name",
		})
		requireDiagnostic(t, err, "telemetry.attribute.hmac_secret_ref", "invalid_telemetry_config")
	})

	t.Run("ignores upstream OTel environment names", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"OTEL_ATTRIBUTE_COUNT_LIMIT":                             "1",
			"OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT":                      "1",
			"OTEL_BLRP_EXPORT_TIMEOUT":                               "1",
			"OTEL_BLRP_MAX_EXPORT_BATCH_SIZE":                        "1",
			"OTEL_BSP_MAX_EXPORT_BATCH_SIZE":                         "1",
			"OTEL_BSP_SCHEDULE_DELAY":                                "1",
			"OTEL_CONFIG_FILE":                                       "/tmp/otel.yaml",
			"OTEL_ENTITIES":                                          "host.id=hostile",
			"OTEL_EXPERIMENTAL_CONFIG_FILE":                          "/tmp/otel-experimental.yaml",
			"OTEL_EXPORTER_OTLP_CERTIFICATE":                         "/tmp/cert.pem",
			"OTEL_EXPORTER_OTLP_ENDPOINT":                            "https://collector.example.test:4318/otel",
			"OTEL_EXPORTER_OTLP_HEADERS":                             "authorization=secret",
			"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":                       "https://logs.example.test:4318",
			"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT":                    "https://metrics.example.test:4318",
			"OTEL_EXPORTER_OTLP_PROTOCOL":                            "grpc",
			"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT":                     "https://traces.example.test:4318",
			"OTEL_EXPORTER_PROMETHEUS_HOST":                          "0.0.0.0",
			"OTEL_EXPORTER_ZIPKIN_ENDPOINT":                          "https://zipkin.example.test/api/v2/spans",
			"OTEL_GO_XRAY_CONTEXT_MISSING":                           "LOG_ERROR",
			"OTEL_LOG_LEVEL":                                         "debug",
			"OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT":                   "1",
			"OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT":            "1",
			"OTEL_LOGS_EXPORTER":                                     "otlp",
			"OTEL_METRIC_EXPORT_INTERVAL":                            "1",
			"OTEL_METRIC_EXPORT_TIMEOUT":                             "1",
			"OTEL_METRICS_EXEMPLAR_FILTER":                           "always_on",
			"OTEL_METRICS_EXPORTER":                                  "otlp",
			"OTEL_PROPAGATORS":                                       "baggage,tracecontext",
			"OTEL_RESOURCE_ATTRIBUTES":                               "service.name=hostile,host.id=hostile",
			"OTEL_SDK_DISABLED":                                      "true",
			"OTEL_SEMCONV_STABILITY_OPT_IN":                          "http/dup",
			"OTEL_SERVICE_NAME":                                      "hostile",
			"OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT":                        "1",
			"OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT":                 "1",
			"OTEL_SPAN_EVENT_COUNT_LIMIT":                            "1",
			"OTEL_SPAN_LINK_COUNT_LIMIT":                             "1",
			"OTEL_TRACES_EXPORTER":                                   "otlp",
			"OTEL_TRACES_SAMPLER":                                    "always_on",
			"OTEL_TRACES_SAMPLER_ARG":                                "1.0",
			"OTEL_PYTHON_DISABLED_INSTRUMENTATIONS":                  "all",
			"OTEL_EXPORTER_OTLP_TRACES_HEADERS":                      "authorization=trace",
			"OTEL_EXPORTER_OTLP_METRICS_HEADERS":                     "authorization=metric",
			"OTEL_EXPORTER_OTLP_LOGS_HEADERS":                        "authorization=log",
			"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL":                     "http/protobuf",
			"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL":                    "grpc",
			"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL":                       "grpc",
			"OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE":           "/tmp/trace-cert.pem",
			"OTEL_EXPORTER_OTLP_METRICS_CLIENT_CERTIFICATE":          "/tmp/metric-cert.pem",
			"OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE":             "/tmp/log-cert.pem",
			"OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY":                   "/tmp/trace-key.pem",
			"OTEL_EXPORTER_OTLP_METRICS_CLIENT_KEY":                  "/tmp/metric-key.pem",
			"OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY":                     "/tmp/log-key.pem",
			"OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE":                  "/tmp/trace-ca.pem",
			"OTEL_EXPORTER_OTLP_METRICS_CERTIFICATE":                 "/tmp/metric-ca.pem",
			"OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE":                    "/tmp/log-ca.pem",
			"OTEL_EXPORTER_OTLP_TRACES_COMPRESSION":                  "gzip",
			"OTEL_EXPORTER_OTLP_METRICS_COMPRESSION":                 "gzip",
			"OTEL_EXPORTER_OTLP_LOGS_COMPRESSION":                    "gzip",
			"OTEL_EXPORTER_OTLP_TRACES_TIMEOUT":                      "1",
			"OTEL_EXPORTER_OTLP_METRICS_TIMEOUT":                     "1",
			"OTEL_EXPORTER_OTLP_LOGS_TIMEOUT":                        "1",
			"OTEL_EXPORTER_OTLP_TRACES_INSECURE":                     "true",
			"OTEL_EXPORTER_OTLP_METRICS_INSECURE":                    "true",
			"OTEL_EXPORTER_OTLP_LOGS_INSECURE":                       "true",
			"OTEL_EXPORTER_OTLP_TRACES_TEMPORALITY_PREFERENCE":       "delta",
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE":      "delta",
			"OTEL_EXPORTER_OTLP_LOGS_TEMPORALITY_PREFERENCE":         "delta",
			"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL_UNSUPPORTED":         "custom",
			"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL_UNSUPPORTED":        "custom",
			"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL_UNSUPPORTED":           "custom",
			"OTEL_EXPORTER_OTLP_TRACES_HEADERS_UNSUPPORTED":          "x=trace",
			"OTEL_EXPORTER_OTLP_METRICS_HEADERS_UNSUPPORTED":         "x=metric",
			"OTEL_EXPORTER_OTLP_LOGS_HEADERS_UNSUPPORTED":            "x=log",
			"OTEL_EXPORTER_PROMETHEUS_PORT":                          "9464",
			"OTEL_EXPORTER_PROMETHEUS_WITH_RESOURCE_CONSTANT_LABELS": "service.name",
		})
		if cfg.Telemetry.Exporter.Kind != "none" || cfg.Telemetry.Exporter.Endpoint != "" {
			t.Fatalf("upstream OTel env must not configure exporter: %#v", cfg.Telemetry.Exporter)
		}
		if cfg.Telemetry.Resource.ServiceName != "cartulary.app" {
			t.Fatalf("upstream OTel env must not configure resource service name: %q", cfg.Telemetry.Resource.ServiceName)
		}
		if cfg.Telemetry.Traces.SampleRatio != 0.10 || cfg.Telemetry.Traces.SamplerProfile != "auto" {
			t.Fatalf("upstream OTel env must not configure sampler: %#v", cfg.Telemetry.Traces)
		}
		if cfg.Telemetry.Traces.AcceptRemoteContext {
			t.Fatal("upstream OTel env must not enable remote context")
		}
		if cfg.Telemetry.Metrics.Exemplars.Enabled {
			t.Fatal("upstream OTel env must not enable exemplars")
		}
		if cfg.Telemetry.Logs.BridgeEnabled {
			t.Fatal("upstream OTel env must not enable log bridge")
		}
		if cfg.Telemetry.Processor.MaxQueueSize != 2048 || cfg.Telemetry.Processor.MaxExportBatchSize != 512 {
			t.Fatalf("upstream OTel env must not configure processor bounds: %#v", cfg.Telemetry.Processor)
		}
	})
}

func TestOpenTelemetryConfigValidation(t *testing.T) {
	t.Run("requires endpoint only for enabled exporter kinds", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND": "otlp_http",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")

		err = loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": "https://collector.example.test:4318/otel",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")
	})

	t.Run("accepts and rejects endpoint grammars", func(t *testing.T) {
		validHTTP := string(fixturesConfigValid(t)) + `
[telemetry.exporter]
kind = "otlp_http"
endpoint = "https://collector.example.test:4318/otel"
`
		if _, err := LoadWithOptions(LoadOptions{Path: writeTempConfig(t, validHTTP)}); err != nil {
			t.Fatalf("load valid OTLP/HTTP endpoint: %v", err)
		}

		validGRPC := string(fixturesConfigValid(t)) + `
[telemetry.exporter]
kind = "otlp_grpc"
endpoint = "http://collector.example.test:4317/"
`
		if _, err := LoadWithOptions(LoadOptions{Path: writeTempConfig(t, validGRPC)}); err != nil {
			t.Fatalf("load valid OTLP/gRPC endpoint: %v", err)
		}

		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND":     "otlp_grpc",
			"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": "https://collector.example.test:4317/v1/traces",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")

		err = loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND":     "otlp_http",
			"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": "https://caf\u00e9.example.test:4318/otel",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")

		err = loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND":     "otlp_http",
			"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": "https://xn--caf-dma.example.test:4318/otel",
		})
		requireDiagnostic(t, err, "telemetry.exporter.endpoint", "invalid_telemetry_config")
	})

	t.Run("enforces cross-key bounds", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__PROCESSOR__MAX_QUEUE_SIZE":            "8",
			"CARTULARY__TELEMETRY__PROCESSOR__MAX_EXPORT_BATCH_SIZE":     "9",
			"CARTULARY__TELEMETRY__EXPORTER__RETRY__INITIAL_INTERVAL_MS": "1000",
			"CARTULARY__TELEMETRY__EXPORTER__RETRY__MAX_INTERVAL_MS":     "999",
			"CARTULARY__TELEMETRY__TRACES__ACCEPT_REMOTE_CONTEXT":        "true",
			"CARTULARY__TELEMETRY__METRICS__EXEMPLARS__ENABLED":          "true",
			"CARTULARY__TELEMETRY__OTEL_ENV_PASSTHROUGH":                 "true",
		})
		requireDiagnostic(t, err, "telemetry.processor.max_export_batch_size", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.exporter.retry.max_interval_ms", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.traces.accept_remote_context", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.metrics.exemplars.enabled", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.otel_env_passthrough", "invalid_telemetry_config")
	})

	t.Run("materializes cross-key rule fixtures", func(t *testing.T) {
		base := string(fixturesConfigValid(t))
		validHTTPEndpoint := "https://collector.example.test:4318/otel"
		validGRPCEndpoint := "https://collector.example.test:4317/"
		cases := []struct {
			id         string
			content    string
			env        map[string]string
			wantPath   string
			wantReason string
		}{
			{
				id: "OTEL-CFG-001",
				env: map[string]string{
					"CARTULARY__TELEMETRY__EXPORTER__KIND": "otlp_http",
				},
				wantPath:   "telemetry.exporter.endpoint",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-002",
				env: map[string]string{
					"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": validHTTPEndpoint,
				},
				wantPath:   "telemetry.exporter.endpoint",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-003",
				env: map[string]string{
					"CARTULARY__TELEMETRY__EXPORTER__KIND":     "otlp_http",
					"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": validHTTPEndpoint,
					"CARTULARY__TELEMETRY__EXPORTER__PROTOCOL": "grpc",
				},
				wantPath:   "telemetry.exporter.protocol",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-004",
				env: map[string]string{
					"CARTULARY__TELEMETRY__PROCESSOR__MAX_QUEUE_SIZE":        "8",
					"CARTULARY__TELEMETRY__PROCESSOR__MAX_EXPORT_BATCH_SIZE": "9",
				},
				wantPath:   "telemetry.processor.max_export_batch_size",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-005",
				env: map[string]string{
					"CARTULARY__TELEMETRY__EXPORTER__RETRY__INITIAL_INTERVAL_MS": "1000",
					"CARTULARY__TELEMETRY__EXPORTER__RETRY__MAX_INTERVAL_MS":     "999",
				},
				wantPath:   "telemetry.exporter.retry.max_interval_ms",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-006",
				env: map[string]string{
					"CARTULARY__TELEMETRY__ATTRIBUTE__INCIDENT_CORRELATION": "hmac_64bit",
				},
				wantPath:   "telemetry.attribute.hmac_secret_ref",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-007",
				env: map[string]string{
					"CARTULARY__TELEMETRY__TRACES__ACCEPT_REMOTE_CONTEXT": "true",
				},
				wantPath:   "telemetry.traces.accept_remote_context",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-008",
				env: map[string]string{
					"CARTULARY__TELEMETRY__TRACES__SAMPLER_PROFILE": "ProbabilitySampler",
				},
				wantPath:   "telemetry.traces.sampler_profile",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-008A",
				env: map[string]string{
					"CARTULARY__TELEMETRY__TRACES__SAMPLE_RATIO":    "0.5",
					"CARTULARY__TELEMETRY__TRACES__SAMPLER_PROFILE": "cartulary.sampler.always_on.v1",
				},
				wantPath:   "telemetry.traces.sampler_profile",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-009",
				env: map[string]string{
					"CARTULARY__TELEMETRY__EXPORTER__KIND":     "otlp_grpc",
					"CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": validGRPCEndpoint,
					"CARTULARY__TELEMETRY__EXPORTER__PROTOCOL": "http/json",
				},
				wantPath:   "telemetry.exporter.protocol",
				wantReason: "invalid_telemetry_config",
			},
			{
				id:       "OTEL-CFG-010",
				content:  base + "\n[telemetry.exporter.traces]\nendpoint = \"https://traces.example.test:4318\"\n",
				wantPath: "telemetry.exporter.traces.endpoint", wantReason: "unknown_key",
			},
			{
				id:       "OTEL-CFG-011",
				content:  base + "\n[telemetry.exporter.metrics.headers]\nauthorization = { kind = \"env\", name = \"metrics-token\" }\n",
				wantPath: "telemetry.exporter.metrics.headers.authorization", wantReason: "unknown_key",
			},
			{
				id: "OTEL-CFG-012",
				env: map[string]string{
					"CARTULARY__TELEMETRY__METRICS__EXEMPLARS__ENABLED": "true",
				},
				wantPath:   "telemetry.metrics.exemplars.enabled",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-013",
				env: map[string]string{
					"CARTULARY__TELEMETRY__LOGS__BODY_MAX_CHARS": "8193",
				},
				wantPath:   "telemetry.logs.body_max_chars",
				wantReason: "invalid_telemetry_config",
			},
			{
				id: "OTEL-CFG-014",
				env: map[string]string{
					"CARTULARY__TELEMETRY__OTEL_ENV_PASSTHROUGH": "true",
				},
				wantPath:   "telemetry.otel_env_passthrough",
				wantReason: "invalid_telemetry_config",
			},
		}

		for _, tc := range cases {
			t.Run(tc.id, func(t *testing.T) {
				content := tc.content
				if content == "" {
					content = base
				}
				err := loadInvalidConfig(t, content, tc.env)
				requireDiagnostic(t, err, tc.wantPath, tc.wantReason)
			})
		}
	})

	t.Run("enforces sampler consistency", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__TRACES__SAMPLE_RATIO":    "0.5",
			"CARTULARY__TELEMETRY__TRACES__SAMPLER_PROFILE": "cartulary.sampler.always_on.v1",
		})
		requireDiagnostic(t, err, "telemetry.traces.sampler_profile", "invalid_telemetry_config")
	})

	t.Run("enforces closed enums and bounds", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__KIND":                    "stdout",
			"CARTULARY__TELEMETRY__EXPORTER__COMPRESSION":             "brotli",
			"CARTULARY__TELEMETRY__EXPORTER__RETRY__MULTIPLIER":       "5.1",
			"CARTULARY__TELEMETRY__EXPORTER__RETRY__MAX_ELAPSED_MS":   "300001",
			"CARTULARY__TELEMETRY__LOGS__BODY_MAX_CHARS":              "8193",
			"CARTULARY__TELEMETRY__METRICS__TEMPORALITY_PROFILE":      "delta",
			"CARTULARY__TELEMETRY__PROCESSOR__OVERFLOW_POLICY":        "block",
			"CARTULARY__TELEMETRY__PROCESSOR__EXPORT_TIMEOUT_MS":      "99",
			"CARTULARY__TELEMETRY__SELF_DIAGNOSTICS__RECURSION_GUARD": "record",
			"CARTULARY__TELEMETRY__SHUTDOWN__FLUSH_TIMEOUT_MS":        "99",
		})
		requireDiagnostic(t, err, "telemetry.exporter.kind", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.exporter.compression", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.exporter.retry.multiplier", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.exporter.retry.max_elapsed_ms", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.logs.body_max_chars", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.metrics.temporality_profile", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.processor.overflow_policy", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.processor.export_timeout_ms", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.self_diagnostics.recursion_guard", "invalid_telemetry_config")
		requireDiagnostic(t, err, "telemetry.shutdown.flush_timeout_ms", "invalid_telemetry_config")
	})

	t.Run("enforces service instance id opacity predicate", func(t *testing.T) {
		valid := "10000000-0000-4000-8000-000000000001"
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__RESOURCE__SERVICE_INSTANCE_ID": valid,
		})
		if cfg.Telemetry.Resource.ServiceInstanceID != valid {
			t.Fatalf("configured service_instance_id was not preserved: %q", cfg.Telemetry.Resource.ServiceInstanceID)
		}
		requireTelemetryUUIDV4(t, cfg.Telemetry.Resource.ServiceInstanceID)

		for _, tc := range []struct {
			name  string
			value string
		}{
			{name: "arbitrary string", value: "host-prod-1"},
			{name: "nil uuid", value: "00000000-0000-0000-0000-000000000000"},
			{name: "non-v4 uuid", value: "10000000-0000-1000-8000-000000000001"},
			{name: "uppercase uuid", value: "10000000-0000-4000-8000-0000000000AA"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
					"CARTULARY__TELEMETRY__RESOURCE__SERVICE_INSTANCE_ID": tc.value,
				})
				requireDiagnostic(t, err, "telemetry.resource.service_instance_id", "invalid_telemetry_config")
			})
		}
	})
}

func TestOpenTelemetrySecretReferences(t *testing.T) {
	t.Run("validates secret ref shape and resolution", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "Authorization=otel-token",
		})
		if err := ResolveTelemetrySecretReferences(cfg, map[string]string{
			"CARTULARY_SECRET_OTEL_TOKEN": "visible secret value",
		}); err != nil {
			t.Fatalf("resolve safe telemetry secret: %v", err)
		}
	})

	t.Run("rejects missing and unsafe resolved secret values", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "Authorization=otel-token",
		})
		err := ResolveTelemetrySecretReferences(cfg, map[string]string{})
		requireDiagnostic(t, err, "telemetry.exporter.headers.authorization", "invalid_telemetry_config")

		err = ResolveTelemetrySecretReferences(cfg, map[string]string{
			"CARTULARY_SECRET_OTEL_TOKEN": " bad\nvalue ",
		})
		requireDiagnostic(t, err, "telemetry.exporter.headers.authorization", "invalid_telemetry_config")
		if strings.Contains(err.Error(), "bad\nvalue") {
			t.Fatal("secret diagnostics must not contain raw secret material")
		}
	})

	t.Run("enforces resolved header byte bounds", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "A=one",
		})
		if err := ResolveTelemetrySecretReferences(cfg, map[string]string{
			"CARTULARY_SECRET_ONE": strings.Repeat("A", 4096),
		}); err != nil {
			t.Fatalf("resolve exact-size telemetry header value: %v", err)
		}
		err := ResolveTelemetrySecretReferences(cfg, map[string]string{
			"CARTULARY_SECRET_ONE": strings.Repeat("A", 4097),
		})
		requireDiagnostic(t, err, "telemetry.exporter.headers.a", "invalid_telemetry_config")

		cfg = mustLoadConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__EXPORTER__HEADERS": "A=one,B=two,C=three",
		})
		err = ResolveTelemetrySecretReferences(cfg, map[string]string{
			"CARTULARY_SECRET_ONE":   strings.Repeat("A", 3000),
			"CARTULARY_SECRET_TWO":   strings.Repeat("B", 3000),
			"CARTULARY_SECRET_THREE": strings.Repeat("C", 3000),
		})
		requireDiagnostic(t, err, "telemetry.exporter.headers", "invalid_telemetry_config")
		if strings.Contains(err.Error(), strings.Repeat("A", 32)) {
			t.Fatal("header block diagnostics must not contain raw secret material")
		}
	})

	t.Run("requires HMAC secret ref when incident correlation is enabled", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixturesConfigValid(t)), map[string]string{
			"CARTULARY__TELEMETRY__ATTRIBUTE__INCIDENT_CORRELATION": "hmac_64bit",
		})
		requireDiagnostic(t, err, "telemetry.attribute.hmac_secret_ref", "invalid_telemetry_config")
	})
}

func fixturesConfigValid(t testing.TB) []byte {
	t.Helper()
	return fixtures.MustRead("config", "valid.toml")
}

func requireTelemetryUUIDV4(t testing.TB, value string) {
	t.Helper()
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.Version() != 4 || parsed == uuid.Nil || value != strings.ToLower(parsed.String()) {
		t.Fatalf("value must be canonical lowercase non-nil UUID v4, got %q", value)
	}
}
