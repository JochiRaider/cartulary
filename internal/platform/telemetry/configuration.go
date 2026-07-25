package telemetry

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

var ConfigurationKey = mustConfigurationKey()

func mustConfigurationKey() config.Key[telemetryconfiguration.Config] {
	key, err := config.NewKey[telemetryconfiguration.Config]("platform.telemetry")
	if err != nil {
		panic(err)
	}
	return key
}

// RegisterConfigurationContribution installs the telemetry-owned namespace in
// the application catalog. The generic loader has already supplied presence,
// overlay, normalization, and structural findings through the owner policy.
func RegisterConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[telemetryconfiguration.Config]{
		Key:       ConfigurationKey,
		Namespace: "telemetry",
		Paths: []string{
			"telemetry.attribute.hmac_secret_ref",
			"telemetry.attribute.incident_correlation",
			"telemetry.enabled",
			"telemetry.exporter.compression",
			"telemetry.exporter.endpoint",
			"telemetry.exporter.headers",
			"telemetry.exporter.kind",
			"telemetry.exporter.protocol",
			"telemetry.exporter.retry.enabled",
			"telemetry.exporter.retry.initial_interval_ms",
			"telemetry.exporter.retry.max_elapsed_ms",
			"telemetry.exporter.retry.max_interval_ms",
			"telemetry.exporter.retry.multiplier",
			"telemetry.logs.body_max_chars",
			"telemetry.logs.bridge_enabled",
			"telemetry.metrics.enabled",
			"telemetry.metrics.exemplars.enabled",
			"telemetry.metrics.temporality_profile",
			"telemetry.otel_env_passthrough",
			"telemetry.processor.export_timeout_ms",
			"telemetry.processor.logs.schedule_delay_ms",
			"telemetry.processor.max_export_batch_size",
			"telemetry.processor.max_queue_size",
			"telemetry.processor.metrics.schedule_delay_ms",
			"telemetry.processor.overflow_policy",
			"telemetry.processor.traces.schedule_delay_ms",
			"telemetry.resource.deployment_environment_name",
			"telemetry.resource.service_instance_id",
			"telemetry.resource.service_name",
			"telemetry.resource.service_namespace",
			"telemetry.resource.service_version",
			"telemetry.self_diagnostics.enabled",
			"telemetry.self_diagnostics.recursion_guard",
			"telemetry.shutdown.flush_timeout_ms",
			"telemetry.traces.accept_remote_context",
			"telemetry.traces.enabled",
			"telemetry.traces.sample_ratio",
			"telemetry.traces.sampler_profile",
		},
		ApplyOverlay: func(settings telemetryconfiguration.Config, segments []string, raw string) (telemetryconfiguration.Config, *config.Diagnostic) {
			finding := telemetryconfiguration.ApplyOverlay(&settings, segments, raw)
			if finding == nil {
				return settings, nil
			}
			return settings, &config.Diagnostic{
				Path:       finding.Path,
				ReasonCode: finding.ReasonCode,
				Message:    finding.Message,
			}
		},
		Project: func(source config.Source) (telemetryconfiguration.Config, []config.Diagnostic) {
			var settings telemetryconfiguration.Config
			if err := source.Decode("telemetry", &settings); err != nil {
				return telemetryconfiguration.Config{}, []config.Diagnostic{{
					Path:       "telemetry",
					ReasonCode: "invalid_telemetry_config",
					Message:    err.Error(),
				}}
			}
			normalized, findings := telemetryconfiguration.NormalizeAndValidate(settings, source)
			diagnostics := make([]config.Diagnostic, len(findings))
			for index, finding := range findings {
				diagnostics[index] = config.Diagnostic{
					Path:       finding.Path,
					ReasonCode: finding.ReasonCode,
					Message:    finding.Message,
				}
			}
			return normalized, diagnostics
		},
		Clone: telemetryconfiguration.Clone,
	})
}

// RegisterSecretPurposes resolves telemetry references during explicit owner
// preflight and binds every value to its exact purpose.
func RegisterSecretPurposes(settings telemetryconfiguration.Config, env map[string]string, registry *secretpurpose.Registry) error {
	resolved, findings := telemetryconfiguration.ResolveSecrets(settings, env)
	if len(findings) > 0 {
		return configurationDiagnosticsError(findings)
	}
	for _, secret := range resolved.Secrets {
		if err := registry.Register(secret.Name, secret.Purpose, secret.Value); err != nil {
			return config.NewDiagnosticsError(config.Diagnostic{
				Path:       secret.Purpose,
				ReasonCode: "invalid_telemetry_config",
				Message:    "telemetry secret purpose is reused",
			})
		}
	}
	return nil
}

func configurationDiagnosticsError(findings []telemetryconfiguration.Finding) error {
	if len(findings) == 0 {
		return nil
	}
	diagnostics := make([]config.Diagnostic, len(findings))
	for index, finding := range findings {
		diagnostics[index] = config.Diagnostic{
			Path:       finding.Path,
			ReasonCode: finding.ReasonCode,
			Message:    finding.Message,
		}
	}
	return config.NewDiagnosticsError(diagnostics...)
}

func ConfigurationValue(snapshot config.Snapshot) (telemetryconfiguration.Config, error) {
	settings, err := config.Value(snapshot, ConfigurationKey)
	if err != nil {
		return telemetryconfiguration.Config{}, fmt.Errorf("project telemetry configuration: %w", err)
	}
	return settings, nil
}
