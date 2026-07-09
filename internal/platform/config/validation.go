package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

type filesystemRoot struct {
	Path       string
	ConfigPath string
}

const limitRegistryMaxInt64 = int64(^uint64(0) >> 1)

func Validate(cfg Config) (Config, error) {
	normalized := cfg
	diagnostics := validateConfigStructure(&normalized, configPresence{})
	if len(diagnostics) > 0 {
		return Config{}, newDiagnosticsError(diagnostics)
	}
	return normalized, nil
}

func ValidateForStartup(cfg Config) (Config, error) {
	normalized, err := Validate(cfg)
	if err != nil {
		return Config{}, err
	}

	diagnostics := validateStartupFilesystemRoots(&normalized)
	if len(diagnostics) > 0 {
		return Config{}, newDiagnosticsError(diagnostics)
	}

	return normalized, nil
}

func ResolveTelemetrySecretReferences(cfg Config, env map[string]string) error {
	_, err := ResolveTelemetryExporterHeaders(cfg, env)
	if err != nil {
		return err
	}
	if !cfg.Telemetry.Attribute.HMACSecretRef.Empty() {
		diagnostics := make([]Diagnostic, 0)
		validateTelemetrySecretRef(cfg.Telemetry.Attribute.HMACSecretRef, "telemetry.attribute.hmac_secret_ref", &diagnostics)
		if len(diagnostics) == 0 {
			_, _ = validateResolvedTelemetrySecret(cfg.Telemetry.Attribute.HMACSecretRef, "telemetry.attribute.hmac_secret_ref", env, &diagnostics)
		}
		if len(diagnostics) > 0 {
			return newDiagnosticsError(diagnostics)
		}
	}
	return nil
}

func ResolveTelemetryExporterHeaders(cfg Config, env map[string]string) (map[string]string, error) {
	diagnostics := make([]Diagnostic, 0)
	headerBlockBytes := 0
	resolved := make(map[string]string, len(cfg.Telemetry.Exporter.Headers))
	for name, ref := range cfg.Telemetry.Exporter.Headers {
		validateTelemetrySecretRef(ref, "telemetry.exporter.headers."+name, &diagnostics)
		if len(diagnostics) == 0 {
			if value, ok := validateResolvedTelemetrySecret(ref, "telemetry.exporter.headers."+name, env, &diagnostics); ok {
				headerBlockBytes += len(strings.ToLower(name)) + len(": ") + len(value)
				resolved[strings.ToLower(name)] = value
			}
		}
	}
	if len(diagnostics) == 0 && headerBlockBytes > 8192 {
		appendTelemetryDiagnostic(&diagnostics, "telemetry.exporter.headers", "configured telemetry exporter header block must be at most 8192 bytes")
	}
	if len(diagnostics) > 0 {
		return nil, newDiagnosticsError(diagnostics)
	}
	return resolved, nil
}

func validateConfigStructure(cfg *Config, presence configPresence) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)

	applyDefaultLimitValues(cfg, presence)

	switch {
	case cfg.ConfigSchemaID == "":
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "missing_required_key",
			Message:    "config_schema_id is required",
		})
	case cfg.ConfigSchemaID != expectedConfigSchemaID:
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "unsupported_config_schema_id",
			Message:    fmt.Sprintf("unsupported config_schema_id %q", cfg.ConfigSchemaID),
		})
	}

	if cfg.DeploymentProfile == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "deployment_profile",
			ReasonCode: "missing_required_key",
			Message:    "deployment_profile is required",
		})
	} else if !isValidDeploymentProfile(cfg.DeploymentProfile) {
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "deployment_profile",
			ReasonCode: "invalid_enum",
			Message:    fmt.Sprintf("deployment_profile %q is not supported", cfg.DeploymentProfile),
		})
	}

	validatePublicOrigin(&cfg.Application, &diagnostics)
	validateRootBinding(&cfg.Roots.DatabaseStorage, "roots.database_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ObjectStorage, "roots.object_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.BackupStorage, "roots.backup_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ReferencePackStorage, "roots.reference_pack_storage", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.TemporaryWork, "roots.temporary_work", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.ExportOutputs, "roots.export_outputs", cfg.DeploymentProfile, false, false, &diagnostics)
	validateBootstrapManifestPath(&cfg.Bootstrap, presence, &diagnostics)
	validateEnterpriseAuthenticationConfig(&cfg.EnterpriseAuthentication, presence, &diagnostics)
	validateLimitRegistry(cfg.Limits, &diagnostics)
	validateTelemetryConfig(&cfg.Telemetry, presence, &diagnostics)

	if len(diagnostics) > 0 {
		return diagnostics
	}

	roots := collectFilesystemRoots(*cfg)
	diagnostics = append(diagnostics, detectBackupStorageRootSubstitution(roots)...)
	diagnostics = append(diagnostics, detectFilesystemRootOverlap(roots)...)
	return diagnostics
}

func validatePublicOrigin(application *ApplicationConfig, diagnostics *[]Diagnostic) {
	if strings.TrimSpace(application.PublicOrigin) == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "missing_required_key",
			Message:    "application public origin is required",
		})
		return
	}

	parsed, err := url.Parse(application.PublicOrigin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin must be an absolute http or https origin",
		})
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin scheme must be http or https",
		})
		return
	}
	if parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin must not include userinfo, path, query, or fragment",
		})
		return
	}
	application.PublicOrigin = parsed.Scheme + "://" + parsed.Host
}

func validateStartupFilesystemRoots(cfg *Config) []Diagnostic {
	roots := collectFilesystemRoots(*cfg)
	for i := range roots {
		canonicalPath, diagnostic := canonicalizeFilesystemRoot(roots[i].Path, roots[i].ConfigPath)
		if diagnostic != nil {
			return []Diagnostic{*diagnostic}
		}
		roots[i].Path = canonicalPath
	}

	diagnostics := detectBackupStorageRootSubstitution(roots)
	diagnostics = append(diagnostics, detectFilesystemRootOverlap(roots)...)
	if len(diagnostics) > 0 {
		return diagnostics
	}

	return nil
}

func validateRootBinding(binding *RootBinding, path string, deploymentProfile string, allowManagedService bool, managedServiceDependsOnProfile bool, diagnostics *[]Diagnostic) {
	if binding.BindingKind == "" && binding.Path == "" && binding.ServiceRef == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "missing_required_key",
			Message:    "required runtime root is missing",
		})
		return
	}

	if binding.BindingKind == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path + ".binding_kind",
			ReasonCode: "missing_required_key",
			Message:    "binding_kind is required",
		})
		return
	}

	switch binding.BindingKind {
	case "filesystem_root":
		if binding.Path == "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".path",
				ReasonCode: "missing_required_key",
				Message:    "path is required for filesystem_root bindings",
			})
		} else if normalized, diagnostic := validateConfiguredPath(binding.Path, path+".path"); diagnostic != nil {
			*diagnostics = append(*diagnostics, *diagnostic)
		} else {
			binding.Path = normalized
		}

		if binding.ServiceRef != "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".service_ref",
				ReasonCode: "type_mismatch",
				Message:    "service_ref is not allowed for filesystem_root bindings",
			})
		}

	case "managed_service":
		if !allowManagedService || (managedServiceDependsOnProfile && deploymentProfile == "disconnected") {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".binding_kind",
				ReasonCode: "profile_incompatible_binding",
				Message:    fmt.Sprintf("binding_kind %q is not allowed for deployment_profile %q", binding.BindingKind, deploymentProfile),
			})
		}

		if binding.ServiceRef == "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".service_ref",
				ReasonCode: "missing_required_key",
				Message:    "service_ref is required for managed_service bindings",
			})
		}

		if binding.Path != "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".path",
				ReasonCode: "type_mismatch",
				Message:    "path is not allowed for managed_service bindings",
			})
		}

	default:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path + ".binding_kind",
			ReasonCode: "invalid_enum",
			Message:    fmt.Sprintf("binding_kind %q is not supported", binding.BindingKind),
		})
	}
}

func validateConfiguredPath(raw string, path string) (string, *Diagnostic) {
	return validateConfiguredAbsolutePOSIXPath(raw, path, "filesystem root paths")
}

func validateConfiguredManifestPath(raw string, path string) (string, *Diagnostic) {
	return validateConfiguredAbsolutePOSIXPath(raw, path, "bootstrap manifest path")
}

func validateConfiguredProviderManifestPath(raw string, path string) (string, *Diagnostic) {
	return validateConfiguredAbsolutePOSIXPath(raw, path, "enterprise provider manifest path")
}

func validateConfiguredAbsolutePOSIXPath(raw string, path string, subject string) (string, *Diagnostic) {
	if !isPOSIXAbsolutePath(raw) {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_not_absolute",
			Message:    subject + " must be an absolute POSIX path",
		}
	}

	if strings.ContainsRune(raw, '\x00') {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    subject + " must not contain NUL",
		}
	}

	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    subject + " must not use shell expansion segments",
		}
	}

	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &Diagnostic{
				Path:       path,
				ReasonCode: "path_forbidden_segment",
				Message:    subject + " must not contain . or .. segments",
			}
		}
	}

	return cleanPOSIXPath(raw), nil
}

func validateBootstrapManifestPath(bootstrap *BootstrapConfig, presence configPresence, diagnostics *[]Diagnostic) {
	if bootstrap.FirstAdminManifestPath == "" {
		if presence.isDefined("bootstrap", "first_admin_manifest_path") {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       "bootstrap.first_admin_manifest_path",
				ReasonCode: "path_not_absolute",
				Message:    "bootstrap manifest path must be an absolute POSIX path when configured",
			})
		}
		return
	}

	if normalized, diagnostic := validateConfiguredManifestPath(bootstrap.FirstAdminManifestPath, "bootstrap.first_admin_manifest_path"); diagnostic != nil {
		*diagnostics = append(*diagnostics, *diagnostic)
	} else {
		bootstrap.FirstAdminManifestPath = normalized
	}
}

func validateEnterpriseAuthenticationConfig(enterprise *EnterpriseAuthenticationConfig, presence configPresence, diagnostics *[]Diagnostic) {
	manifestPathDefined := presence.isDefined("enterprise_authentication", "provider_manifest_path")
	if !enterprise.Claimed {
		if enterprise.ProviderManifestPath != "" || manifestPathDefined {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       "enterprise_authentication.provider_manifest_path",
				ReasonCode: "profile_incompatible_binding",
				Message:    "enterprise provider manifest path is valid only when enterprise authentication is claimed",
			})
		}
		return
	}

	if enterprise.ProviderManifestPath == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "enterprise_authentication.provider_manifest_path",
			ReasonCode: "provider_manifest_path_missing",
			Message:    "enterprise provider manifest path is required when enterprise authentication is claimed",
		})
		return
	}
	if normalized, diagnostic := validateConfiguredProviderManifestPath(enterprise.ProviderManifestPath, "enterprise_authentication.provider_manifest_path"); diagnostic != nil {
		*diagnostics = append(*diagnostics, *diagnostic)
	} else {
		enterprise.ProviderManifestPath = normalized
	}
}

func applyDefaultLimitValues(cfg *Config, presence configPresence) {
	applyDefaultInt64(&cfg.Limits.ObjectBlobs.MaxDeclaredByteSize, DefaultObjectBlobMaxDeclaredByteSize, presence, "limits", "object_blobs", "max_declared_byte_size")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCSVSourceBytes, DefaultImportMaxCSVSourceBytes, presence, "limits", "imports", "max_csv_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxXLSXSourceBytes, DefaultImportMaxXLSXSourceBytes, presence, "limits", "imports", "max_xlsx_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxRows, DefaultImportMaxRows, presence, "limits", "imports", "max_rows")
	applyDefaultInt64(&cfg.Limits.Imports.MaxColumns, DefaultImportMaxColumns, presence, "limits", "imports", "max_columns")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCells, DefaultImportMaxCells, presence, "limits", "imports", "max_cells")
	applyDefaultInt64(&cfg.Limits.Archives.DefaultMaxExtractedBytes, DefaultArchiveMaxExtractedBytes, presence, "limits", "archives", "default_max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Archives.MaxCompressionRatio, DefaultArchiveMaxCompressionRatio, presence, "limits", "archives", "max_compression_ratio")
	applyDefaultInt64(&cfg.Limits.Archives.MaxMembers, DefaultArchiveMaxMembers, presence, "limits", "archives", "max_members")
	applyDefaultInt64(&cfg.Limits.ReferencePacks.MaxExtractedBytes, DefaultReferencePackMaxExtractedBytes, presence, "limits", "reference_packs", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.IncidentBundles.MaxExtractedBytes, DefaultIncidentBundleMaxExtractedBytes, presence, "limits", "incident_bundles", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxPreviewablePayloadBytes, DefaultPreviewMaxPreviewablePayloadBytes, presence, "limits", "previews", "max_previewable_payload_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxTextInlineBytes, DefaultPreviewMaxTextInlineBytes, presence, "limits", "previews", "max_text_inline_bytes")
}

func applyDefaultTelemetryValues(cfg *TelemetryConfig, presence configPresence) {
	applyDefaultBool(&cfg.Enabled, true, presence, "telemetry", "enabled")
	applyDefaultString(&cfg.Exporter.Kind, "none", presence, "telemetry", "exporter", "kind")
	applyDefaultString(&cfg.Exporter.Compression, "none", presence, "telemetry", "exporter", "compression")
	applyDefaultBool(&cfg.Exporter.Retry.Enabled, true, presence, "telemetry", "exporter", "retry", "enabled")
	applyDefaultInt64(&cfg.Exporter.Retry.MaxElapsedMS, 30000, presence, "telemetry", "exporter", "retry", "max_elapsed_ms")
	applyDefaultInt64(&cfg.Exporter.Retry.InitialIntervalMS, 100, presence, "telemetry", "exporter", "retry", "initial_interval_ms")
	applyDefaultInt64(&cfg.Exporter.Retry.MaxIntervalMS, 5000, presence, "telemetry", "exporter", "retry", "max_interval_ms")
	applyDefaultFloat64(&cfg.Exporter.Retry.Multiplier, 2.0, presence, "telemetry", "exporter", "retry", "multiplier")
	applyDefaultBool(&cfg.Traces.Enabled, true, presence, "telemetry", "traces", "enabled")
	applyDefaultFloat64(&cfg.Traces.SampleRatio, 0.10, presence, "telemetry", "traces", "sample_ratio")
	applyDefaultString(&cfg.Traces.SamplerProfile, "auto", presence, "telemetry", "traces", "sampler_profile")
	applyDefaultBool(&cfg.Metrics.Enabled, true, presence, "telemetry", "metrics", "enabled")
	applyDefaultString(&cfg.Metrics.TemporalityProfile, "cartulary.metrics.temporality.cumulative.v1", presence, "telemetry", "metrics", "temporality_profile")
	applyDefaultInt64(&cfg.Logs.BodyMaxChars, 2048, presence, "telemetry", "logs", "body_max_chars")
	applyDefaultInt64(&cfg.Processor.MaxQueueSize, 2048, presence, "telemetry", "processor", "max_queue_size")
	applyDefaultInt64(&cfg.Processor.MaxExportBatchSize, 512, presence, "telemetry", "processor", "max_export_batch_size")
	applyDefaultInt64(&cfg.Processor.Traces.ScheduleDelayMS, 5000, presence, "telemetry", "processor", "traces", "schedule_delay_ms")
	applyDefaultInt64(&cfg.Processor.Metrics.ScheduleDelayMS, 60000, presence, "telemetry", "processor", "metrics", "schedule_delay_ms")
	applyDefaultInt64(&cfg.Processor.Logs.ScheduleDelayMS, 1000, presence, "telemetry", "processor", "logs", "schedule_delay_ms")
	applyDefaultInt64(&cfg.Processor.ExportTimeoutMS, 2000, presence, "telemetry", "processor", "export_timeout_ms")
	applyDefaultString(&cfg.Processor.OverflowPolicy, "drop_new", presence, "telemetry", "processor", "overflow_policy")
	applyDefaultInt64(&cfg.Shutdown.FlushTimeoutMS, 5000, presence, "telemetry", "shutdown", "flush_timeout_ms")
	applyDefaultBool(&cfg.SelfDiagnostics.Enabled, true, presence, "telemetry", "self_diagnostics", "enabled")
	applyDefaultString(&cfg.SelfDiagnostics.RecursionGuard, "drop_telemetry_about_telemetry", presence, "telemetry", "self_diagnostics", "recursion_guard")
	applyDefaultString(&cfg.Resource.ServiceName, "cartulary.app", presence, "telemetry", "resource", "service_name")
	applyDefaultString(&cfg.Resource.ServiceNamespace, "cartulary", presence, "telemetry", "resource", "service_namespace")
	applyDefaultString(&cfg.Resource.ServiceVersion, "0.0.0+unknown", presence, "telemetry", "resource", "service_version")
	if cfg.Resource.ServiceInstanceID == "" && !presence.isDefined("telemetry", "resource", "service_instance_id") {
		cfg.Resource.ServiceInstanceID = uuid.NewString()
	}
	applyDefaultString(&cfg.Attribute.IncidentCorrelation, "none", presence, "telemetry", "attribute", "incident_correlation")

	switch cfg.Exporter.Kind {
	case "otlp_http":
		applyDefaultString(&cfg.Exporter.Protocol, "http/protobuf", presence, "telemetry", "exporter", "protocol")
	case "otlp_grpc":
		applyDefaultString(&cfg.Exporter.Protocol, "grpc", presence, "telemetry", "exporter", "protocol")
	}
	if cfg.Exporter.Headers == nil {
		cfg.Exporter.Headers = map[string]SecretRef{}
	}
}

func applyDefaultInt64(target *int64, defaultValue int64, presence configPresence, path ...string) {
	if *target != 0 || presence.isDefined(path...) {
		return
	}
	*target = defaultValue
}

func applyDefaultBool(target *bool, defaultValue bool, presence configPresence, path ...string) {
	if *target || presence.isDefined(path...) {
		return
	}
	*target = defaultValue
}

func applyDefaultString(target *string, defaultValue string, presence configPresence, path ...string) {
	if *target != "" || presence.isDefined(path...) {
		return
	}
	*target = defaultValue
}

func applyDefaultFloat64(target *float64, defaultValue float64, presence configPresence, path ...string) {
	if *target != 0 || presence.isDefined(path...) {
		return
	}
	*target = defaultValue
}

func validateLimitRegistry(limits LimitConfig, diagnostics *[]Diagnostic) {
	validateLimitValue(limits.ObjectBlobs.MaxDeclaredByteSize, "limits.object_blobs.max_declared_byte_size", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxCSVSourceBytes, "limits.imports.max_csv_source_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxXLSXSourceBytes, "limits.imports.max_xlsx_source_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxRows, "limits.imports.max_rows", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxColumns, "limits.imports.max_columns", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxCells, "limits.imports.max_cells", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Archives.DefaultMaxExtractedBytes, "limits.archives.default_max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Archives.MaxCompressionRatio, "limits.archives.max_compression_ratio", 1, 1000, diagnostics)
	validateLimitValue(limits.Archives.MaxMembers, "limits.archives.max_members", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.ReferencePacks.MaxExtractedBytes, "limits.reference_packs.max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.IncidentBundles.MaxExtractedBytes, "limits.incident_bundles.max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Previews.MaxPreviewablePayloadBytes, "limits.previews.max_previewable_payload_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Previews.MaxTextInlineBytes, "limits.previews.max_text_inline_bytes", 1, limitRegistryMaxInt64, diagnostics)
}

func validateLimitValue(value int64, path string, min int64, max int64, diagnostics *[]Diagnostic) {
	switch {
	case value < min:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "value_below_minimum",
			Message:    fmt.Sprintf("value must be at least %d", min),
		})
	case value > max:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "value_above_maximum",
			Message:    fmt.Sprintf("value must be at most %d", max),
		})
	}
}

func validateTelemetryConfig(cfg *TelemetryConfig, presence configPresence, diagnostics *[]Diagnostic) {
	applyDefaultTelemetryValues(cfg, presence)

	validateTelemetryEnum(cfg.Exporter.Kind, "telemetry.exporter.kind", []string{"none", "otlp_http", "otlp_grpc"}, diagnostics)
	validateTelemetryEnum(cfg.Exporter.Compression, "telemetry.exporter.compression", []string{"none", "gzip"}, diagnostics)
	validateTelemetryBoolFalse(cfg.OTelEnvPassthrough, "telemetry.otel_env_passthrough", diagnostics)
	validateTelemetryBoolFalse(cfg.Traces.AcceptRemoteContext, "telemetry.traces.accept_remote_context", diagnostics)
	validateTelemetryBoolFalse(cfg.Metrics.Exemplars.Enabled, "telemetry.metrics.exemplars.enabled", diagnostics)
	validateTelemetryEnum(cfg.Metrics.TemporalityProfile, "telemetry.metrics.temporality_profile", []string{"cartulary.metrics.temporality.cumulative.v1"}, diagnostics)
	validateTelemetryEnum(cfg.Processor.OverflowPolicy, "telemetry.processor.overflow_policy", []string{"drop_new"}, diagnostics)
	validateTelemetryEnum(cfg.SelfDiagnostics.RecursionGuard, "telemetry.self_diagnostics.recursion_guard", []string{"drop_telemetry_about_telemetry"}, diagnostics)
	validateTelemetryEnum(cfg.Attribute.IncidentCorrelation, "telemetry.attribute.incident_correlation", []string{"none", "hmac_64bit"}, diagnostics)
	validateTelemetryEnum(cfg.Traces.SamplerProfile, "telemetry.traces.sampler_profile", []string{"auto", "cartulary.sampler.always_on.v1", "cartulary.sampler.always_off.v1", "cartulary.sampler.traceidratio_compat.v1"}, diagnostics)
	validateTelemetryFloat(cfg.Traces.SampleRatio, "telemetry.traces.sample_ratio", 0, 1, diagnostics)
	validateTelemetryFloat(cfg.Exporter.Retry.Multiplier, "telemetry.exporter.retry.multiplier", 1, 5, diagnostics)
	validateTelemetryInt(cfg.Exporter.Retry.MaxElapsedMS, "telemetry.exporter.retry.max_elapsed_ms", 0, 300000, diagnostics)
	validateTelemetryInt(cfg.Exporter.Retry.InitialIntervalMS, "telemetry.exporter.retry.initial_interval_ms", 50, 30000, diagnostics)
	validateTelemetryInt(cfg.Exporter.Retry.MaxIntervalMS, "telemetry.exporter.retry.max_interval_ms", 100, 60000, diagnostics)
	validateTelemetryInt(cfg.Logs.BodyMaxChars, "telemetry.logs.body_max_chars", 0, 8192, diagnostics)
	validateTelemetryInt(cfg.Processor.MaxQueueSize, "telemetry.processor.max_queue_size", 1, 65536, diagnostics)
	validateTelemetryInt(cfg.Processor.MaxExportBatchSize, "telemetry.processor.max_export_batch_size", 1, cfg.Processor.MaxQueueSize, diagnostics)
	validateTelemetryInt(cfg.Processor.Traces.ScheduleDelayMS, "telemetry.processor.traces.schedule_delay_ms", 100, 300000, diagnostics)
	validateTelemetryInt(cfg.Processor.Metrics.ScheduleDelayMS, "telemetry.processor.metrics.schedule_delay_ms", 100, 300000, diagnostics)
	validateTelemetryInt(cfg.Processor.Logs.ScheduleDelayMS, "telemetry.processor.logs.schedule_delay_ms", 100, 300000, diagnostics)
	validateTelemetryInt(cfg.Processor.ExportTimeoutMS, "telemetry.processor.export_timeout_ms", 100, 10000, diagnostics)
	validateTelemetryInt(cfg.Shutdown.FlushTimeoutMS, "telemetry.shutdown.flush_timeout_ms", 100, 30000, diagnostics)
	validateTelemetryToken(cfg.Resource.ServiceName, "telemetry.resource.service_name", diagnostics)
	validateTelemetryToken(cfg.Resource.ServiceNamespace, "telemetry.resource.service_namespace", diagnostics)
	validateTelemetryVersion(cfg.Resource.ServiceVersion, diagnostics)
	validateTelemetryOptionalToken(cfg.Resource.DeploymentEnvironmentName, "telemetry.resource.deployment_environment_name", diagnostics)
	validateTelemetryInstanceID(cfg.Resource.ServiceInstanceID, presence, diagnostics)
	validateTelemetryHeaders(cfg.Exporter.Headers, diagnostics)
	if !cfg.Attribute.HMACSecretRef.Empty() {
		validateTelemetrySecretRef(cfg.Attribute.HMACSecretRef, "telemetry.attribute.hmac_secret_ref", diagnostics)
	}

	endpointConfigured := cfg.Exporter.Endpoint != ""
	endpointDefined := presence.isDefined("telemetry", "exporter", "endpoint")
	switch cfg.Exporter.Kind {
	case "none":
		if endpointConfigured {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.endpoint", "endpoint must be omitted when telemetry.exporter.kind is none")
		}
		if cfg.Exporter.Protocol != "" || presence.isDefined("telemetry", "exporter", "protocol") {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.protocol", "protocol must be derived only when telemetry export is enabled")
		}
	case "otlp_http":
		if !endpointConfigured {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.endpoint", "endpoint is required when telemetry.exporter.kind is otlp_http")
		} else {
			validateOTLPHTTPEndpoint(cfg.Exporter.Endpoint, "telemetry.exporter.endpoint", diagnostics)
		}
		if cfg.Exporter.Protocol != "http/protobuf" {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.protocol", "otlp_http requires protocol http/protobuf")
		}
	case "otlp_grpc":
		if !endpointConfigured {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.endpoint", "endpoint is required when telemetry.exporter.kind is otlp_grpc")
		} else {
			validateOTLPGRPCEndpoint(cfg.Exporter.Endpoint, "telemetry.exporter.endpoint", diagnostics)
		}
		if cfg.Exporter.Protocol != "grpc" {
			appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.protocol", "otlp_grpc requires protocol grpc")
		}
	}
	if endpointDefined && cfg.Exporter.Endpoint == "" {
		appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.endpoint", "endpoint must not be empty")
	}
	if cfg.Exporter.Retry.MaxIntervalMS < cfg.Exporter.Retry.InitialIntervalMS {
		appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.retry.max_interval_ms", "max_interval_ms must be greater than or equal to initial_interval_ms")
	}
	if cfg.Processor.MaxExportBatchSize > cfg.Processor.MaxQueueSize {
		appendTelemetryDiagnostic(diagnostics, "telemetry.processor.max_export_batch_size", "max_export_batch_size must be less than or equal to max_queue_size")
	}
	validateSamplerConsistency(cfg.Traces.SamplerProfile, cfg.Traces.SampleRatio, diagnostics)
	if cfg.Attribute.IncidentCorrelation == "hmac_64bit" && cfg.Attribute.HMACSecretRef.Empty() {
		appendTelemetryDiagnostic(diagnostics, "telemetry.attribute.hmac_secret_ref", "hmac_secret_ref is required when incident correlation is hmac_64bit")
	}
}

func validateTelemetryEnum(value string, path string, allowed []string, diagnostics *[]Diagnostic) {
	for _, candidate := range allowed {
		if value == candidate {
			return
		}
	}
	appendTelemetryDiagnostic(diagnostics, path, "value is outside the adopted telemetry enum")
}

func validateTelemetryBoolFalse(value bool, path string, diagnostics *[]Diagnostic) {
	if value {
		appendTelemetryDiagnostic(diagnostics, path, "value must be false in the adopted telemetry profile")
	}
}

func validateTelemetryInt(value int64, path string, min int64, max int64, diagnostics *[]Diagnostic) {
	if value < min {
		appendTelemetryDiagnostic(diagnostics, path, fmt.Sprintf("value must be at least %d", min))
		return
	}
	if value > max {
		appendTelemetryDiagnostic(diagnostics, path, fmt.Sprintf("value must be at most %d", max))
	}
}

func validateTelemetryFloat(value float64, path string, min float64, max float64, diagnostics *[]Diagnostic) {
	if value < min || value > max {
		appendTelemetryDiagnostic(diagnostics, path, fmt.Sprintf("value must be between %.2f and %.2f", min, max))
	}
}

func validateTelemetryToken(value string, path string, diagnostics *[]Diagnostic) {
	if !isValidTelemetryToken(value) {
		appendTelemetryDiagnostic(diagnostics, path, "value must be 1..128 ASCII letters, digits, '.', '_', or '-'")
	}
}

func validateTelemetryOptionalToken(value string, path string, diagnostics *[]Diagnostic) {
	if value == "" {
		return
	}
	validateTelemetryToken(value, path, diagnostics)
}

func validateTelemetryVersion(value string, diagnostics *[]Diagnostic) {
	if value == "0.0.0+unknown" {
		return
	}
	if len(value) > 128 || !semverPattern.MatchString(value) {
		appendTelemetryDiagnostic(diagnostics, "telemetry.resource.service_version", "service_version must be SemVer 2.0.0 or 0.0.0+unknown")
	}
}

func validateTelemetryInstanceID(value string, presence configPresence, diagnostics *[]Diagnostic) {
	if value == "" {
		if presence.isDefined("telemetry", "resource", "service_instance_id") {
			appendTelemetryDiagnostic(diagnostics, "telemetry.resource.service_instance_id", "service_instance_id must be a canonical lowercase UUID v4 when configured")
		}
		return
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.Version() != 4 || value != strings.ToLower(parsed.String()) || parsed == uuid.Nil {
		appendTelemetryDiagnostic(diagnostics, "telemetry.resource.service_instance_id", "service_instance_id must be a canonical lowercase non-nil UUID v4")
	}
}

func validateTelemetryHeaders(headers map[string]SecretRef, diagnostics *[]Diagnostic) {
	if len(headers) > 16 {
		appendTelemetryDiagnostic(diagnostics, "telemetry.exporter.headers", "at most 16 telemetry exporter headers are allowed")
	}
	seen := make(map[string]string)
	for name, ref := range headers {
		canonicalName := strings.ToLower(name)
		path := "telemetry.exporter.headers." + name
		if !isValidTelemetryHeaderName(name) {
			appendTelemetryDiagnostic(diagnostics, path, "telemetry exporter header name is invalid")
		}
		if _, forbidden := protocolOwnedTelemetryHeaders[canonicalName]; forbidden {
			appendTelemetryDiagnostic(diagnostics, path, "telemetry exporter header must not override a protocol-owned header")
		}
		if previous, exists := seen[canonicalName]; exists && previous != name {
			appendTelemetryDiagnostic(diagnostics, path, "telemetry exporter header duplicates another header after lowercase canonicalization")
		}
		seen[canonicalName] = name
		validateTelemetrySecretRef(ref, path, diagnostics)
	}
}

func validateTelemetrySecretRef(ref SecretRef, path string, diagnostics *[]Diagnostic) {
	if ref.Kind != "env" || !isValidSecretRefName(ref.Name) {
		appendTelemetryDiagnostic(diagnostics, path, "telemetry secret references must use secret_ref_v1 kind env and a safe name")
	}
}

func validateResolvedTelemetrySecret(ref SecretRef, path string, env map[string]string, diagnostics *[]Diagnostic) (string, bool) {
	value, ok := lookupEnv(env, secretRefEnvName(ref.Name))
	if !ok || value == "" || !isValidResolvedTelemetrySecret(value) {
		appendTelemetryDiagnostic(diagnostics, path, "telemetry secret reference could not be resolved to a safe value")
		return "", false
	}
	return value, true
}

func validateSamplerConsistency(profile string, ratio float64, diagnostics *[]Diagnostic) {
	switch profile {
	case "cartulary.sampler.always_off.v1":
		if ratio != 0 {
			appendTelemetryDiagnostic(diagnostics, "telemetry.traces.sampler_profile", "always_off sampler requires sample_ratio 0.0")
		}
	case "cartulary.sampler.always_on.v1":
		if ratio != 1 {
			appendTelemetryDiagnostic(diagnostics, "telemetry.traces.sampler_profile", "always_on sampler requires sample_ratio 1.0")
		}
	case "cartulary.sampler.traceidratio_compat.v1":
		if ratio <= 0 || ratio >= 1 {
			appendTelemetryDiagnostic(diagnostics, "telemetry.traces.sampler_profile", "traceidratio sampler requires 0.0 < sample_ratio < 1.0")
		}
	}
}

func validateOTLPHTTPEndpoint(raw string, path string, diagnostics *[]Diagnostic) {
	parsed, err := url.Parse(raw)
	if err != nil || !validTelemetryEndpointBase(parsed) || parsed.Path == "" && strings.HasSuffix(raw, "//") {
		appendTelemetryDiagnostic(diagnostics, path, "OTLP/HTTP endpoint must be an absolute http(s) URL with explicit port and no userinfo, query, or fragment")
		return
	}
	if parsed.Path != "" && parsed.Path != "/" {
		if strings.Contains(parsed.EscapedPath(), "%") || strings.Contains(parsed.Path, "//") {
			appendTelemetryDiagnostic(diagnostics, path, "OTLP/HTTP endpoint path must not contain encoded or duplicate slash segments")
			return
		}
		for _, segment := range strings.Split(strings.Trim(parsed.Path, "/"), "/") {
			if segment == "" || segment == "." || segment == ".." || !telemetryPathSegmentPattern.MatchString(segment) {
				appendTelemetryDiagnostic(diagnostics, path, "OTLP/HTTP endpoint path contains an unsupported segment")
				return
			}
		}
	}
}

func validateOTLPGRPCEndpoint(raw string, path string, diagnostics *[]Diagnostic) {
	parsed, err := url.Parse(raw)
	if err != nil || !validTelemetryEndpointBase(parsed) {
		appendTelemetryDiagnostic(diagnostics, path, "OTLP/gRPC endpoint must be an absolute http(s) URL with explicit port and no userinfo, query, fragment, or path")
		return
	}
	if parsed.Path != "" && parsed.Path != "/" {
		appendTelemetryDiagnostic(diagnostics, path, "OTLP/gRPC endpoint must not include a non-root path")
	}
}

func validTelemetryEndpointBase(parsed *url.URL) bool {
	if parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if host == "" || port == "" {
		return false
	}
	if strings.ContainsAny(host, "\t\n\r ") || strings.Contains(host, "%") {
		return false
	}
	if strings.Contains(host, ":") {
		return net.ParseIP(host) != nil
	}
	if strings.Contains(strings.ToLower(host), "xn--") {
		return false
	}
	for _, r := range host {
		if r > 127 {
			return false
		}
	}
	return telemetryEndpointHostPattern.MatchString(host)
}

func appendTelemetryDiagnostic(diagnostics *[]Diagnostic, path string, message string) {
	*diagnostics = append(*diagnostics, Diagnostic{
		Path:       path,
		ReasonCode: "invalid_telemetry_config",
		Message:    message,
	})
}

func detectFilesystemRootOverlap(roots []filesystemRoot) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	for i := 0; i < len(roots); i++ {
		for j := i + 1; j < len(roots); j++ {
			switch {
			case roots[i].Path == roots[j].Path:
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[j].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[j].Path, roots[i].ConfigPath),
				})
			case pathWithinRoot(roots[i].Path, roots[j].Path):
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[j].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[j].Path, roots[i].ConfigPath),
				})
			case pathWithinRoot(roots[j].Path, roots[i].Path):
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[i].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[i].Path, roots[j].ConfigPath),
				})
			}
		}
	}

	sortDiagnostics(diagnostics)
	return diagnostics
}

func detectBackupStorageRootSubstitution(roots []filesystemRoot) []Diagnostic {
	var backupRoot *filesystemRoot
	for i := range roots {
		if roots[i].ConfigPath == "roots.backup_storage.path" {
			backupRoot = &roots[i]
			break
		}
	}
	if backupRoot == nil {
		return nil
	}

	diagnostics := make([]Diagnostic, 0, 2)
	for _, root := range roots {
		switch root.ConfigPath {
		case "roots.export_outputs.path", "roots.temporary_work.path":
			if filesystemRootsOverlap(backupRoot.Path, root.Path) {
				diagnostics = append(diagnostics, Diagnostic{
					Path:       backupRoot.ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("backup storage root %q must remain distinct from %s", backupRoot.Path, strings.TrimSuffix(root.ConfigPath, ".path")),
				})
			}
		}
	}

	sortDiagnostics(diagnostics)
	return diagnostics
}

func filesystemRootsOverlap(left string, right string) bool {
	return left == right || pathWithinRoot(left, right) || pathWithinRoot(right, left)
}

func collectFilesystemRoots(cfg Config) []filesystemRoot {
	roots := make([]filesystemRoot, 0, 6)
	if cfg.Roots.DatabaseStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.DatabaseStorage.Path, ConfigPath: "roots.database_storage.path"})
	}
	if cfg.Roots.ObjectStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ObjectStorage.Path, ConfigPath: "roots.object_storage.path"})
	}
	if cfg.Roots.BackupStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.BackupStorage.Path, ConfigPath: "roots.backup_storage.path"})
	}
	if cfg.Roots.ReferencePackStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ReferencePackStorage.Path, ConfigPath: "roots.reference_pack_storage.path"})
	}
	if cfg.Roots.TemporaryWork.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.TemporaryWork.Path, ConfigPath: "roots.temporary_work.path"})
	}
	if cfg.Roots.ExportOutputs.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ExportOutputs.Path, ConfigPath: "roots.export_outputs.path"})
	}

	return roots
}

func canonicalizeFilesystemRoot(root string, configPath string) (string, *Diagnostic) {
	cleaned := cleanPOSIXPath(root)
	existingPrefix, exists, err := nearestExistingPath(cleaned)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("inspect filesystem root: %v", err),
		}
	}

	if !exists {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    "filesystem root must resolve under an existing writable parent",
		}
	}

	info, err := os.Stat(existingPrefix)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("stat filesystem root: %v", err),
		}
	}
	if !info.IsDir() {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "type_mismatch",
			Message:    "filesystem root must resolve to a directory",
		}
	}

	resolvedPrefix, err := filepath.EvalSymlinks(existingPrefix)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("resolve filesystem root symlinks: %v", err),
		}
	}

	resolvedPrefix = cleanPOSIXPath(resolvedPrefix)
	if !isWritablePath(resolvedPrefix) {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    "filesystem root is not writable at startup",
		}
	}

	if existingPrefix == cleaned {
		return resolvedPrefix, nil
	}

	suffix := strings.TrimPrefix(cleaned, existingPrefix)
	return cleanPOSIXPath(resolvedPrefix + suffix), nil
}

func nearestExistingPath(root string) (string, bool, error) {
	current := root
	for {
		_, err := os.Stat(current)
		if err == nil {
			return current, true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, err
		}

		parent := pathpkg.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func resolvePathWithinFilesystemRoot(root string, relativePath string) (string, error) {
	cleanedRoot := filepath.Clean(root)
	if !filepath.IsAbs(cleanedRoot) {
		return "", fmt.Errorf("filesystem root %q must be absolute", root)
	}
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("path %q must be relative to the configured filesystem root", relativePath)
	}

	resolvedPath := filepath.Clean(filepath.Join(cleanedRoot, relativePath))
	relativeResolvedPath, err := filepath.Rel(cleanedRoot, resolvedPath)
	if err != nil {
		return "", fmt.Errorf("resolve path within filesystem root: %w", err)
	}
	if relativeResolvedPath == ".." || strings.HasPrefix(relativeResolvedPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes configured filesystem root %q", relativePath, cleanedRoot)
	}

	return resolvedPath, nil
}

func writeFileWithinFilesystemRoot(root string, relativePath string, data []byte, perm os.FileMode) error {
	resolvedPath, err := resolvePathWithinFilesystemRoot(root, relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o700); err != nil {
		return fmt.Errorf("create parent directories within filesystem root: %w", err)
	}
	if err := os.WriteFile(resolvedPath, data, perm); err != nil {
		return fmt.Errorf("write file within filesystem root: %w", err)
	}
	return nil
}

func pathWithinRoot(root string, target string) bool {
	if root == target {
		return true
	}
	if root == "/" {
		return strings.HasPrefix(target, "/")
	}
	return strings.HasPrefix(target, root+"/")
}

func cleanPOSIXPath(value string) string {
	cleaned := pathpkg.Clean(value)
	if cleaned == "." {
		return value
	}
	return cleaned
}

func isPOSIXAbsolutePath(value string) bool {
	return strings.HasPrefix(value, "/")
}

func isWritablePath(path string) bool {
	return unix.Access(path, unix.W_OK|unix.X_OK) == nil
}

func isValidDeploymentProfile(profile string) bool {
	switch profile {
	case "disconnected", "on_prem", "cloud":
		return true
	default:
		return false
	}
}

var (
	telemetryTokenPattern        = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	telemetryHeaderNamePattern   = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
	telemetryEndpointHostPattern = regexp.MustCompile(`^[A-Za-z0-9.-]+$`)
	telemetryPathSegmentPattern  = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,64}$`)
	secretRefNamePattern         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	semverPattern                = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
)

var protocolOwnedTelemetryHeaders = map[string]struct{}{
	"host":              {},
	"content-type":      {},
	"content-length":    {},
	"transfer-encoding": {},
	"connection":        {},
	"te":                {},
	"user-agent":        {},
	"traceparent":       {},
	"tracestate":        {},
	"baggage":           {},
}

func isValidTelemetryToken(value string) bool {
	return telemetryTokenPattern.MatchString(value)
}

func isValidTelemetryHeaderName(value string) bool {
	return telemetryHeaderNamePattern.MatchString(value)
}

func isValidSecretRefName(value string) bool {
	if !secretRefNamePattern.MatchString(value) {
		return false
	}
	return normalizedSecretRefSuffix(value) != ""
}

func secretRefEnvName(name string) string {
	return "CARTULARY_SECRET_" + normalizedSecretRefSuffix(name)
}

func normalizedSecretRefSuffix(name string) string {
	var builder strings.Builder
	previousUnderscore := false
	for _, r := range name {
		var next rune
		switch {
		case r >= 'a' && r <= 'z':
			next = r - ('a' - 'A')
		case r >= 'A' && r <= 'Z':
			next = r
		case r >= '0' && r <= '9':
			next = r
		default:
			next = '_'
		}
		if next == '_' {
			if builder.Len() == 0 || previousUnderscore {
				previousUnderscore = true
				continue
			}
			previousUnderscore = true
			builder.WriteRune(next)
			continue
		}
		previousUnderscore = false
		builder.WriteRune(next)
	}
	return strings.Trim(builder.String(), "_")
}

func isValidResolvedTelemetrySecret(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || !utf8.ValidString(value) || len(value) > 4096 {
		return false
	}
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\x00' || r == 0x7f || r < 0x20 || r > 0x7e {
			return false
		}
	}
	return true
}
