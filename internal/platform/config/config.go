package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	DefaultConfigPath = "/etc/cartulary/config.toml"
	ConfigFileEnv     = "CARTULARY_CONFIG_FILE"
	overlayPrefix     = "CARTULARY__"
)

type LoadOptions struct {
	Path string
	Env  map[string]string
}

type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type DiagnosticsError struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Config struct {
	ConfigSchemaID    string          `toml:"config_schema_id"`
	DeploymentProfile string          `toml:"deployment_profile"`
	Roots             RootBindings    `toml:"roots"`
	Bootstrap         BootstrapConfig `toml:"bootstrap"`
	Limits            LimitConfig     `toml:"limits"`
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

func (e *DiagnosticsError) Error() string {
	parts := make([]string, 0, len(e.Diagnostics))
	for _, diagnostic := range e.Diagnostics {
		if diagnostic.Path == "" {
			parts = append(parts, diagnostic.Message)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", diagnostic.Path, diagnostic.Message))
	}

	return strings.Join(parts, "; ")
}

func (e *DiagnosticsError) JSON() string {
	payload := struct {
		Diagnostics []Diagnostic `json:"diagnostics"`
	}{
		Diagnostics: e.Diagnostics,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return `{"diagnostics":[{"code":"diagnostic_render_failed","message":"failed to render diagnostics"}]}`
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
		if !filepath.IsAbs(override) {
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
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	diagnostics := make([]Diagnostic, 0)
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{
			Code:    "invalid_toml",
			Message: err.Error(),
		})
		return Config{}, &DiagnosticsError{Diagnostics: diagnostics}
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		for _, key := range undecoded {
			diagnostics = append(diagnostics, Diagnostic{
				Code:    "unknown_config_key",
				Path:    strings.Join(key, "."),
				Message: "unknown config key",
			})
		}
	}

	for _, key := range sortedOverlayKeys(options.Env) {
		if diagnostic := applyOverlay(&cfg, key, lookupEnvValue(options.Env, key)); diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
	}

	if cfg.ConfigSchemaID == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Code:    "missing_required",
			Path:    "config_schema_id",
			Message: "config_schema_id is required",
		})
	}

	if cfg.DeploymentProfile == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Code:    "missing_required",
			Path:    "deployment_profile",
			Message: "deployment_profile is required",
		})
	}

	// TODO: enforce full Core 04 runtime-root, binding-kind, and limit validation.
	if len(diagnostics) > 0 {
		sortDiagnostics(diagnostics)
		return Config{}, &DiagnosticsError{Diagnostics: diagnostics}
	}

	return cfg, nil
}

func applyOverlay(cfg *Config, envKey string, raw string) *Diagnostic {
	path := strings.TrimPrefix(envKey, overlayPrefix)
	segments := strings.Split(path, "__")
	for i := range segments {
		segments[i] = strings.ToLower(segments[i])
	}

	value := reflect.ValueOf(cfg).Elem()
	for _, segment := range segments[:len(segments)-1] {
		field, ok := findTaggedField(value, segment)
		if !ok {
			return &Diagnostic{
				Code:    "invalid_env_overlay",
				Path:    strings.Join(segments, "."),
				Message: fmt.Sprintf("unknown overlay path segment %q", segment),
			}
		}
		value = field
	}

	field, ok := findTaggedField(value, segments[len(segments)-1])
	if !ok {
		return &Diagnostic{
			Code:    "invalid_env_overlay",
			Path:    strings.Join(segments, "."),
			Message: fmt.Sprintf("unknown overlay target %q", segments[len(segments)-1]),
		}
	}

	if err := assignOverlayValue(field, raw); err != nil {
		return &Diagnostic{
			Code:    "invalid_env_overlay",
			Path:    strings.Join(segments, "."),
			Message: err.Error(),
		}
	}

	return nil
}

func assignOverlayValue(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parse integer overlay: %w", err)
		}
		field.SetInt(value)
		return nil
	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("parse boolean overlay: %w", err)
		}
		field.SetBool(value)
		return nil
	default:
		return fmt.Errorf("unsupported overlay target kind %s", field.Kind())
	}
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

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Path != diagnostics[j].Path {
			return diagnostics[i].Path < diagnostics[j].Path
		}
		if diagnostics[i].Code != diagnostics[j].Code {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return diagnostics[i].Message < diagnostics[j].Message
	})
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
