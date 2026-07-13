package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	DefaultConfigPath           = "/etc/cartulary/config.toml"
	ConfigFileEnv               = "CARTULARY_CONFIG_FILE"
	InvalidDeploymentConfigCode = "invalid_deployment_config"
	overlayPrefix               = "CARTULARY__"
	expectedConfigSchemaID      = "cartulary.deployment_config.v1"

	DefaultObjectBlobMaxDeclaredByteSize     int64 = 536870912
	DefaultImportMaxCSVSourceBytes           int64 = 33554432
	DefaultImportMaxXLSXSourceBytes          int64 = 67108864
	DefaultImportMaxRows                     int64 = 100000
	DefaultImportMaxColumns                  int64 = 256
	DefaultImportMaxCells                    int64 = 5000000
	DefaultArchiveMaxExtractedBytes          int64 = 2147483648
	DefaultArchiveMaxCompressionRatio        int64 = 100
	DefaultArchiveMaxMembers                 int64 = 10000
	DefaultReferencePackMaxExtractedBytes    int64 = 536870912
	DefaultIncidentBundleMaxExtractedBytes   int64 = 68719476736
	DefaultPreviewMaxPreviewablePayloadBytes int64 = 33554432
	DefaultPreviewMaxTextInlineBytes         int64 = 1048576

	PublicSortLimit             = 8
	PublicFilterLimit           = 16
	PublicChangeLimit           = 32
	PublicCollectionActionLimit = 64
)

type LoadOptions struct {
	Path string
	Env  map[string]string
}

type Diagnostic struct {
	Path       string `json:"path,omitempty"`
	ReasonCode string `json:"reason_code"`
	Message    string `json:"message"`
}

type DiagnosticsError struct {
	Code        string       `json:"code"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Config struct {
	ConfigSchemaID           string                         `toml:"config_schema_id"`
	DeploymentProfile        string                         `toml:"deployment_profile"`
	Application              ApplicationConfig              `toml:"application"`
	Roots                    RootBindings                   `toml:"roots"`
	Bootstrap                BootstrapConfig                `toml:"bootstrap"`
	EnterpriseAuthentication EnterpriseAuthenticationConfig `toml:"enterprise_authentication"`
	NetworkFlowActivity      NetworkFlowActivityConfig      `toml:"network_flow_activity"`
	Limits                   LimitConfig                    `toml:"limits"`
	Telemetry                TelemetryConfig                `toml:"telemetry"`
}

type ApplicationConfig struct {
	PublicOrigin string `toml:"public_origin"`
}

type RootBindings struct {
	DatabaseStorage      RootBinding `toml:"database_storage"`
	ObjectStorage        RootBinding `toml:"object_storage"`
	BackupStorage        RootBinding `toml:"backup_storage"`
	ReferencePackStorage RootBinding `toml:"reference_pack_storage"`
	TemporaryWork        RootBinding `toml:"temporary_work"`
	ExportOutputs        RootBinding `toml:"export_outputs"`
}

type RootBinding struct {
	BindingKind string `toml:"binding_kind"`
	Path        string `toml:"path"`
	ServiceRef  string `toml:"service_ref"`
}

type BootstrapConfig struct {
	FirstAdminManifestPath string `toml:"first_admin_manifest_path"`
}

type EnterpriseAuthenticationConfig struct {
	Claimed              bool   `toml:"claimed"`
	ProviderManifestPath string `toml:"provider_manifest_path"`
}

type NetworkFlowActivityConfig struct {
	Claimed             bool   `toml:"claimed"`
	KeyRingManifestPath string `toml:"key_ring_manifest_path"`
}

type LimitConfig struct {
	ObjectBlobs     ObjectBlobLimits     `toml:"object_blobs"`
	Imports         ImportLimits         `toml:"imports"`
	Archives        ArchiveLimits        `toml:"archives"`
	ReferencePacks  ReferencePackLimits  `toml:"reference_packs"`
	IncidentBundles IncidentBundleLimits `toml:"incident_bundles"`
	Previews        PreviewLimits        `toml:"previews"`
}

type ObjectBlobLimits struct {
	MaxDeclaredByteSize int64 `toml:"max_declared_byte_size"`
}

type ImportLimits struct {
	MaxCSVSourceBytes  int64 `toml:"max_csv_source_bytes"`
	MaxXLSXSourceBytes int64 `toml:"max_xlsx_source_bytes"`
	MaxRows            int64 `toml:"max_rows"`
	MaxColumns         int64 `toml:"max_columns"`
	MaxCells           int64 `toml:"max_cells"`
}

type ArchiveLimits struct {
	DefaultMaxExtractedBytes int64 `toml:"default_max_extracted_bytes"`
	MaxCompressionRatio      int64 `toml:"max_compression_ratio"`
	MaxMembers               int64 `toml:"max_members"`
}

type ReferencePackLimits struct {
	MaxExtractedBytes int64 `toml:"max_extracted_bytes"`
}

type IncidentBundleLimits struct {
	MaxExtractedBytes int64 `toml:"max_extracted_bytes"`
}

type PreviewLimits struct {
	MaxPreviewablePayloadBytes int64 `toml:"max_previewable_payload_bytes"`
	MaxTextInlineBytes         int64 `toml:"max_text_inline_bytes"`
}

type TelemetryConfig struct {
	Enabled            bool                           `toml:"enabled"`
	OTelEnvPassthrough bool                           `toml:"otel_env_passthrough"`
	Exporter           TelemetryExporterConfig        `toml:"exporter"`
	Traces             TelemetryTracesConfig          `toml:"traces"`
	Metrics            TelemetryMetricsConfig         `toml:"metrics"`
	Logs               TelemetryLogsConfig            `toml:"logs"`
	Processor          TelemetryProcessorConfig       `toml:"processor"`
	Shutdown           TelemetryShutdownConfig        `toml:"shutdown"`
	SelfDiagnostics    TelemetrySelfDiagnosticsConfig `toml:"self_diagnostics"`
	Resource           TelemetryResourceConfig        `toml:"resource"`
	Attribute          TelemetryAttributeConfig       `toml:"attribute"`
}

type TelemetryExporterConfig struct {
	Kind        string                       `toml:"kind"`
	Endpoint    string                       `toml:"endpoint"`
	Protocol    string                       `toml:"protocol"`
	Compression string                       `toml:"compression"`
	Headers     map[string]SecretRef         `toml:"headers"`
	Retry       TelemetryExporterRetryConfig `toml:"retry"`
}

type TelemetryExporterRetryConfig struct {
	Enabled           bool    `toml:"enabled"`
	MaxElapsedMS      int64   `toml:"max_elapsed_ms"`
	InitialIntervalMS int64   `toml:"initial_interval_ms"`
	MaxIntervalMS     int64   `toml:"max_interval_ms"`
	Multiplier        float64 `toml:"multiplier"`
}

type TelemetryTracesConfig struct {
	Enabled             bool    `toml:"enabled"`
	SampleRatio         float64 `toml:"sample_ratio"`
	SamplerProfile      string  `toml:"sampler_profile"`
	AcceptRemoteContext bool    `toml:"accept_remote_context"`
}

type TelemetryMetricsConfig struct {
	Enabled            bool                    `toml:"enabled"`
	TemporalityProfile string                  `toml:"temporality_profile"`
	Exemplars          TelemetryExemplarConfig `toml:"exemplars"`
}

type TelemetryExemplarConfig struct {
	Enabled bool `toml:"enabled"`
}

type TelemetryLogsConfig struct {
	BridgeEnabled bool  `toml:"bridge_enabled"`
	BodyMaxChars  int64 `toml:"body_max_chars"`
}

type TelemetryProcessorConfig struct {
	MaxQueueSize       int64                          `toml:"max_queue_size"`
	MaxExportBatchSize int64                          `toml:"max_export_batch_size"`
	Traces             TelemetryProcessorSignalConfig `toml:"traces"`
	Metrics            TelemetryProcessorSignalConfig `toml:"metrics"`
	Logs               TelemetryProcessorSignalConfig `toml:"logs"`
	ExportTimeoutMS    int64                          `toml:"export_timeout_ms"`
	OverflowPolicy     string                         `toml:"overflow_policy"`
}

type TelemetryProcessorSignalConfig struct {
	ScheduleDelayMS int64 `toml:"schedule_delay_ms"`
}

type TelemetryShutdownConfig struct {
	FlushTimeoutMS int64 `toml:"flush_timeout_ms"`
}

type TelemetrySelfDiagnosticsConfig struct {
	Enabled        bool   `toml:"enabled"`
	RecursionGuard string `toml:"recursion_guard"`
}

type TelemetryResourceConfig struct {
	ServiceName               string `toml:"service_name"`
	ServiceNamespace          string `toml:"service_namespace"`
	ServiceVersion            string `toml:"service_version"`
	ServiceInstanceID         string `toml:"service_instance_id"`
	DeploymentEnvironmentName string `toml:"deployment_environment_name"`
}

type TelemetryAttributeConfig struct {
	IncidentCorrelation string    `toml:"incident_correlation"`
	HMACSecretRef       SecretRef `toml:"hmac_secret_ref"`
}

type SecretRef struct {
	Kind string `toml:"kind"`
	Name string `toml:"name"`
}

func (r SecretRef) Empty() bool {
	return r.Kind == "" && r.Name == ""
}

type configPresence struct {
	fileMeta     toml.MetaData
	overlayPaths map[string]struct{}
}

func (e *DiagnosticsError) Error() string {
	parts := make([]string, 0, len(e.Diagnostics))
	for _, diagnostic := range e.Diagnostics {
		if diagnostic.Path == "" {
			parts = append(parts, fmt.Sprintf("%s: %s", diagnostic.ReasonCode, diagnostic.Message))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (%s): %s", diagnostic.Path, diagnostic.ReasonCode, diagnostic.Message))
	}

	return strings.Join(parts, "; ")
}

func (e *DiagnosticsError) JSON() string {
	items := append([]Diagnostic(nil), e.Diagnostics...)
	sortDiagnostics(items)

	payload := struct {
		Error struct {
			Code    string `json:"code"`
			Details struct {
				Items []Diagnostic `json:"items"`
			} `json:"details"`
		} `json:"error"`
	}{
		Error: struct {
			Code    string `json:"code"`
			Details struct {
				Items []Diagnostic `json:"items"`
			} `json:"details"`
		}{
			Code: e.Code,
			Details: struct {
				Items []Diagnostic `json:"items"`
			}{
				Items: items,
			},
		},
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return `{"error":{"code":"invalid_deployment_config","details":{"items":[{"reason_code":"diagnostic_render_failed","message":"failed to render diagnostics"}]}}}`
	}

	return string(data)
}

func DiagnosticsFromError(err error) ([]Diagnostic, bool) {
	var diagnosticsErr *DiagnosticsError
	if !errors.As(err, &diagnosticsErr) {
		return nil, false
	}

	return diagnosticsErr.Diagnostics, true
}

func ResolvePath() (string, error) {
	return ResolvePathWithOptions(LoadOptions{})
}

func ResolvePathWithOptions(options LoadOptions) (string, error) {
	if options.Path != "" {
		return options.Path, nil
	}

	if override, ok := lookupEnv(options.Env, ConfigFileEnv); ok && override != "" {
		if !isPOSIXAbsolutePath(override) {
			return "", fmt.Errorf("%s must be an absolute path", ConfigFileEnv)
		}

		return override, nil
	}

	return DefaultConfigPath, nil
}

func Load() (Config, error) {
	return LoadWithOptions(LoadOptions{})
}

func LoadWithEnv(env map[string]string) (Config, error) {
	return LoadWithOptions(LoadOptions{Env: env})
}

func LoadWithOptions(options LoadOptions) (Config, error) {
	path, err := ResolvePathWithOptions(options)
	if err != nil {
		return Config{}, err
	}

	options.Path = path
	return loadFromOptions(options)
}

func LoadFromPath(path string) (Config, error) {
	return LoadWithOptions(LoadOptions{Path: path})
}

func loadFromOptions(options LoadOptions) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(options.Path)
	if err != nil {
		reasonCode := "config_parse_error"
		if errors.Is(err, os.ErrNotExist) {
			reasonCode = "config_file_not_found"
		}
		return Config{}, newDiagnosticsError([]Diagnostic{{
			Path:       options.Path,
			ReasonCode: reasonCode,
			Message:    fmt.Sprintf("read config: %v", err),
		}})
	}

	diagnostics := make([]Diagnostic, 0)
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{
			ReasonCode: "config_parse_error",
			Message:    err.Error(),
		})
		return Config{}, newDiagnosticsError(diagnostics)
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		for _, key := range undecoded {
			diagnostics = append(diagnostics, Diagnostic{
				Path:       strings.Join(key, "."),
				ReasonCode: "unknown_key",
				Message:    "unknown config key",
			})
		}
	}

	for _, key := range sortedOverlayKeys(options.Env) {
		if diagnostic := applyOverlay(&cfg, key, lookupEnvValue(options.Env, key)); diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
	}

	diagnostics = append(diagnostics, validateConfigStructure(&cfg, newConfigPresence(md, options.Env))...)
	if len(diagnostics) > 0 {
		return Config{}, newDiagnosticsError(diagnostics)
	}

	return cfg, nil
}

func applyOverlay(cfg *Config, envKey string, raw string) *Diagnostic {
	segments := overlaySegments(envKey)
	if len(segments) > 0 && segments[0] == "telemetry" {
		return applyTelemetryOverlay(cfg, segments, raw)
	}

	value := reflect.ValueOf(cfg).Elem()
	for _, segment := range segments[:len(segments)-1] {
		field, ok := findTaggedField(value, segment)
		if !ok {
			return &Diagnostic{
				Path:       strings.Join(segments, "."),
				ReasonCode: "unknown_key",
				Message:    fmt.Sprintf("unknown overlay path segment %q", segment),
			}
		}
		value = field
	}

	field, ok := findTaggedField(value, segments[len(segments)-1])
	if !ok {
		return &Diagnostic{
			Path:       strings.Join(segments, "."),
			ReasonCode: "unknown_key",
			Message:    fmt.Sprintf("unknown overlay target %q", segments[len(segments)-1]),
		}
	}

	reasonCode, err := assignOverlayValue(field, raw)
	if err != nil {
		return &Diagnostic{
			Path:       strings.Join(segments, "."),
			ReasonCode: reasonCode,
			Message:    err.Error(),
		}
	}

	return nil
}

func assignOverlayValue(field reflect.Value, raw string) (string, error) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
		return "", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return "type_mismatch", fmt.Errorf("parse integer overlay: %w", err)
		}
		field.SetInt(value)
		return "", nil
	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return "type_mismatch", fmt.Errorf("parse boolean overlay: %w", err)
		}
		field.SetBool(value)
		return "", nil
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(raw, field.Type().Bits())
		if err != nil {
			return "type_mismatch", fmt.Errorf("parse decimal overlay: %w", err)
		}
		field.SetFloat(value)
		return "", nil
	default:
		return "type_mismatch", fmt.Errorf("unsupported overlay target kind %s", field.Kind())
	}
}

func applyTelemetryOverlay(cfg *Config, segments []string, raw string) *Diagnostic {
	path := strings.Join(segments, ".")
	if raw == "" {
		return nil
	}
	if raw == "null" {
		return telemetryDiagnostic(path, "explicit null is not accepted in server-side environment bindings")
	}
	if path == "telemetry.exporter.headers" {
		headers, diagnostic := parseTelemetryHeaderOverlay(raw, path)
		if diagnostic != nil {
			return diagnostic
		}
		cfg.Telemetry.Exporter.Headers = headers
		return nil
	}
	if path == "telemetry.attribute.hmac_secret_ref" {
		if !isValidSecretRefName(raw) {
			return telemetryDiagnostic(path, "telemetry secret reference name is invalid")
		}
		cfg.Telemetry.Attribute.HMACSecretRef = SecretRef{Kind: "env", Name: raw}
		return nil
	}

	value := reflect.ValueOf(cfg).Elem()
	for _, segment := range segments[:len(segments)-1] {
		field, ok := findTaggedField(value, segment)
		if !ok {
			return &Diagnostic{
				Path:       path,
				ReasonCode: "unknown_key",
				Message:    fmt.Sprintf("unknown overlay path segment %q", segment),
			}
		}
		value = field
	}

	field, ok := findTaggedField(value, segments[len(segments)-1])
	if !ok {
		return &Diagnostic{
			Path:       path,
			ReasonCode: "unknown_key",
			Message:    fmt.Sprintf("unknown overlay target %q", segments[len(segments)-1]),
		}
	}
	if field.Kind() == reflect.Int || field.Kind() == reflect.Int8 || field.Kind() == reflect.Int16 || field.Kind() == reflect.Int32 || field.Kind() == reflect.Int64 {
		if !asciiDigits(raw) {
			return telemetryDiagnostic(path, "telemetry integer overlay must use unsigned base-10 ASCII digits")
		}
	}
	if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
		if !validTelemetryDecimalToken(raw) {
			return telemetryDiagnostic(path, "telemetry decimal overlay must be finite decimal notation")
		}
	}

	reasonCode, err := assignOverlayValue(field, raw)
	if err != nil {
		if reasonCode == "unknown_key" {
			return &Diagnostic{Path: path, ReasonCode: reasonCode, Message: err.Error()}
		}
		return telemetryDiagnostic(path, err.Error())
	}
	return nil
}

func parseTelemetryHeaderOverlay(raw string, path string) (map[string]SecretRef, *Diagnostic) {
	headers := make(map[string]SecretRef)
	seen := make(map[string]struct{})
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			return nil, telemetryDiagnostic(path, "telemetry exporter header overlay contains an empty pair")
		}
		name, refName, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, telemetryDiagnostic(path, "telemetry exporter header overlay must use key=secret_ref_name pairs")
		}
		name = strings.TrimSpace(name)
		refName = strings.TrimSpace(refName)
		canonicalName := strings.ToLower(name)
		if !isValidTelemetryHeaderName(name) || !isValidSecretRefName(refName) {
			return nil, telemetryDiagnostic(path, "telemetry exporter header overlay contains an invalid header or secret reference")
		}
		if _, exists := seen[canonicalName]; exists {
			return nil, telemetryDiagnostic(path, "telemetry exporter header overlay contains a duplicate header")
		}
		seen[canonicalName] = struct{}{}
		headers[canonicalName] = SecretRef{Kind: "env", Name: refName}
	}
	return headers, nil
}

func telemetryDiagnostic(path string, message string) *Diagnostic {
	return &Diagnostic{
		Path:       path,
		ReasonCode: "invalid_telemetry_config",
		Message:    message,
	}
}

var telemetryDecimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

func asciiDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func validTelemetryDecimalToken(value string) bool {
	return telemetryDecimalPattern.MatchString(value)
}

func findTaggedField(value reflect.Value, segment string) (reflect.Value, bool) {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		tag := structField.Tag.Get("toml")
		if tag == "" {
			tag = strings.ToLower(structField.Name)
		}
		if tag != segment {
			continue
		}
		return value.Field(i), true
	}

	return reflect.Value{}, false
}

func sortedOverlayKeys(env map[string]string) []string {
	keys := make([]string, 0)
	if env != nil {
		for key := range env {
			if strings.HasPrefix(key, overlayPrefix) {
				keys = append(keys, key)
			}
		}
	} else {
		for _, entry := range os.Environ() {
			key := entry
			if idx := strings.IndexByte(entry, '='); idx >= 0 {
				key = entry[:idx]
			}
			if strings.HasPrefix(key, overlayPrefix) {
				keys = append(keys, key)
			}
		}
	}

	sort.Strings(keys)
	return keys
}

func newConfigPresence(md toml.MetaData, env map[string]string) configPresence {
	overlayPaths := make(map[string]struct{})
	for _, key := range sortedOverlayKeys(env) {
		if strings.TrimPrefix(key, overlayPrefix) == "" {
			continue
		}
		if strings.HasPrefix(key, overlayPrefix+"TELEMETRY__") && lookupEnvValue(env, key) == "" {
			continue
		}
		overlayPaths[strings.Join(overlaySegments(key), ".")] = struct{}{}
	}

	return configPresence{
		fileMeta:     md,
		overlayPaths: overlayPaths,
	}
}

func (p configPresence) isDefined(path ...string) bool {
	if len(path) > 0 && p.fileMeta.IsDefined(path...) {
		return true
	}

	_, ok := p.overlayPaths[strings.Join(path, ".")]
	return ok
}

func overlaySegments(envKey string) []string {
	path := strings.TrimPrefix(envKey, overlayPrefix)
	segments := strings.Split(path, "__")
	for i := range segments {
		segments[i] = strings.ToLower(segments[i])
	}
	return segments
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Path != diagnostics[j].Path {
			return diagnostics[i].Path < diagnostics[j].Path
		}
		if diagnostics[i].ReasonCode != diagnostics[j].ReasonCode {
			return diagnostics[i].ReasonCode < diagnostics[j].ReasonCode
		}
		return diagnostics[i].Message < diagnostics[j].Message
	})
}

func NewDiagnosticsError(diagnostics ...Diagnostic) *DiagnosticsError {
	normalized := append([]Diagnostic(nil), diagnostics...)
	sortDiagnostics(normalized)
	return &DiagnosticsError{
		Code:        InvalidDeploymentConfigCode,
		Diagnostics: normalized,
	}
}

func newDiagnosticsError(diagnostics []Diagnostic) *DiagnosticsError {
	return NewDiagnosticsError(diagnostics...)
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}

	return os.LookupEnv(key)
}

func lookupEnvValue(env map[string]string, key string) string {
	value, _ := lookupEnv(env, key)
	return value
}
