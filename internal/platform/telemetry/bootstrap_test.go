package telemetry

import (
	"context"
	"testing"

	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

type telemetryBootstrapSettings struct {
	DeploymentProfile string
	Telemetry         telemetryconfiguration.Config
}

type absentTelemetryPresence struct{}

func (absentTelemetryPresence) Defined(...string) bool { return false }

func bootstrapFromConfig(ctx context.Context, cfg telemetryBootstrapSettings, env map[string]string, options ...BootstrapOption) (*Runtime, error) {
	return Bootstrap(ctx, cfg.Telemetry, cfg.DeploymentProfile, env, options...)
}

func buildResourceIdentityFromConfig(cfg telemetryBootstrapSettings, claims ResolvedClaimIdentity) (ResourceIdentity, error) {
	return BuildResourceIdentity(cfg.Telemetry, cfg.DeploymentProfile, claims)
}

func TestBootstrapNoSDKExportDisabled(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	runtime, err := bootstrapFromConfig(context.Background(), cfg, map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318",
		"OTEL_TRACES_EXPORTER":        "otlp",
	}, WithResolvedClaimIdentity(resolvedClaimIdentity(t)))
	if err != nil {
		t.Fatalf("bootstrap export-disabled telemetry: %v", err)
	}
	if !runtime.Enabled() {
		t.Fatal("telemetry should be enabled with no SDK/exporter when exporter.kind is none")
	}
	if runtime.SamplerProfile() != SamplerProfileTraceIDRatioCompat || runtime.SampleRatio() != 0.10 {
		t.Fatalf("unexpected default sampler profile: profile=%q ratio=%f", runtime.SamplerProfile(), runtime.SampleRatio())
	}
	if runtime.NetworkExportEnabled() || runtime.ExporterKind() != "none" || runtime.ExporterEndpoint() != "" {
		t.Fatalf("export-disabled bootstrap must not create network export or retain default endpoint: kind=%q endpoint=%q network=%t",
			runtime.ExporterKind(), runtime.ExporterEndpoint(), runtime.NetworkExportEnabled())
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown no-SDK telemetry: %v", err)
	}
}

func TestBootstrapActivatesExplicitOTLPHTTPExport(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	cfg.Telemetry.Exporter.Kind = "otlp_http"
	cfg.Telemetry.Exporter.Endpoint = "https://collector.example.test:4318/otel"
	cfg.Telemetry.Exporter.Protocol = "http/protobuf"
	cfg.Telemetry.Exporter.Headers = map[string]telemetryconfiguration.SecretRef{
		"authorization": {Kind: "env", Name: "otel-token"},
	}

	runtime, err := bootstrapFromConfig(context.Background(), cfg, map[string]string{
		"CARTULARY_SECRET_OTEL_TOKEN": "safe-secret-value",
	}, WithResolvedClaimIdentity(resolvedClaimIdentity(t)))
	if err != nil {
		t.Fatalf("bootstrap explicit OTLP/HTTP export: %v", err)
	}
	if !runtime.Enabled() || !runtime.NetworkExportEnabled() {
		t.Fatalf("explicit exporter should enable network export: enabled=%t network=%t", runtime.Enabled(), runtime.NetworkExportEnabled())
	}
	if runtime.ExporterKind() != "otlp_http" || runtime.ExporterEndpoint() != cfg.Telemetry.Exporter.Endpoint {
		t.Fatalf("unexpected exporter identity: kind=%q endpoint=%q", runtime.ExporterKind(), runtime.ExporterEndpoint())
	}
	diagnosticHeaders := runtime.ExporterDiagnosticHeaders()
	if diagnosticHeaders["authorization"] != "[redacted]" {
		t.Fatalf("exporter diagnostic headers must redact secrets: %#v", diagnosticHeaders)
	}
	if ExporterDiagnosticsContainSecret(ExporterHeaderPlan{DiagnosticHeaders: diagnosticHeaders}, "safe-secret-value") {
		t.Fatal("diagnostic headers leaked a resolved exporter secret")
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown explicit OTLP/HTTP telemetry: %v", err)
	}
}

func TestBootstrapIgnoresHostileOTelEnvironment(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	cfg.Telemetry.Resource.ServiceName = "cartulary.app"
	cfg.Telemetry.Exporter.Kind = "none"
	cfg.Telemetry.Exporter.Endpoint = ""

	runtime, err := bootstrapFromConfig(context.Background(), cfg, map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT":   "https://collector.example.test:4318/otel",
		"OTEL_EXPERIMENTAL_CONFIG_FILE": "/tmp/otel.yaml",
		"OTEL_LOGS_EXPORTER":            "otlp",
		"OTEL_METRICS_EXPORTER":         "otlp",
		"OTEL_RESOURCE_ATTRIBUTES":      "service.name=hostile",
		"OTEL_SERVICE_NAME":             "hostile",
		"OTEL_TRACES_EXPORTER":          "otlp",
		"OTEL_TRACES_SAMPLER":           "always_on",
		"OTEL_TRACES_SAMPLER_ARG":       "1.0",
	}, WithResolvedClaimIdentity(resolvedClaimIdentity(t)))
	if err != nil {
		t.Fatalf("bootstrap should ignore upstream OTel env while export is disabled: %v", err)
	}
	if !runtime.Enabled() {
		t.Fatal("hostile OTel env must not disable no-SDK telemetry accessors")
	}
	if runtime.NetworkExportEnabled() || runtime.ExporterKind() != "none" || runtime.ExporterEndpoint() != "" {
		t.Fatalf("hostile OTel env must not create exporter state: kind=%q endpoint=%q network=%t",
			runtime.ExporterKind(), runtime.ExporterEndpoint(), runtime.NetworkExportEnabled())
	}
	resource := runtime.ResourceIdentity()
	if got := resourceAttrValue(resource, "service.name"); got != "cartulary.app" {
		t.Fatalf("hostile OTel env changed service.name: %q", got)
	}
}

func validTelemetryBootstrapConfig(t testing.TB) telemetryBootstrapSettings {
	t.Helper()
	telemetryConfig, findings := telemetryconfiguration.NormalizeAndValidate(
		telemetryconfiguration.Config{},
		absentTelemetryPresence{},
	)
	if len(findings) != 0 {
		t.Fatalf("construct default telemetry settings: %#v", findings)
	}
	return telemetryBootstrapSettings{
		DeploymentProfile: "on_prem",
		Telemetry:         telemetryConfig,
	}
}

func resourceAttrValue(resource ResourceIdentity, key string) string {
	for _, attr := range resource.Attributes {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}
