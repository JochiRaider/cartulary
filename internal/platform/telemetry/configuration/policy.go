package configuration

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const invalidReason = "invalid_telemetry_config"

var (
	decimalPattern      = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	tokenPattern        = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	headerNamePattern   = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
	endpointHostPattern = regexp.MustCompile(`^[A-Za-z0-9.-]+$`)
	pathSegmentPattern  = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,64}$`)
	secretRefPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	semverPattern       = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
)

var protocolOwnedHeaders = map[string]struct{}{
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

// NormalizeAndValidate applies telemetry-owned defaults and returns all pure
// structural findings without performing secret, network, or filesystem work.
func NormalizeAndValidate(source Config, presence Presence) (Config, []Finding) {
	cfg := Clone(source)
	applyDefaults(&cfg, presence)
	findings := make([]Finding, 0)

	validateEnum(cfg.Exporter.Kind, "telemetry.exporter.kind", []string{"none", "otlp_http", "otlp_grpc"}, &findings)
	validateEnum(cfg.Exporter.Compression, "telemetry.exporter.compression", []string{"none", "gzip"}, &findings)
	validateFalse(cfg.OTelEnvPassthrough, "telemetry.otel_env_passthrough", &findings)
	validateFalse(cfg.Traces.AcceptRemoteContext, "telemetry.traces.accept_remote_context", &findings)
	validateFalse(cfg.Metrics.Exemplars.Enabled, "telemetry.metrics.exemplars.enabled", &findings)
	validateEnum(cfg.Metrics.TemporalityProfile, "telemetry.metrics.temporality_profile", []string{"cartulary.metrics.temporality.cumulative.v1"}, &findings)
	validateEnum(cfg.Processor.OverflowPolicy, "telemetry.processor.overflow_policy", []string{"drop_new"}, &findings)
	validateEnum(cfg.SelfDiagnostics.RecursionGuard, "telemetry.self_diagnostics.recursion_guard", []string{"drop_telemetry_about_telemetry"}, &findings)
	validateEnum(cfg.Attribute.IncidentCorrelation, "telemetry.attribute.incident_correlation", []string{"none", "hmac_64bit"}, &findings)
	validateEnum(cfg.Traces.SamplerProfile, "telemetry.traces.sampler_profile", []string{"auto", "cartulary.sampler.always_on.v1", "cartulary.sampler.always_off.v1", "cartulary.sampler.traceidratio_compat.v1"}, &findings)
	validateFloat(cfg.Traces.SampleRatio, "telemetry.traces.sample_ratio", 0, 1, &findings)
	validateFloat(cfg.Exporter.Retry.Multiplier, "telemetry.exporter.retry.multiplier", 1, 5, &findings)
	validateInt(cfg.Exporter.Retry.MaxElapsedMS, "telemetry.exporter.retry.max_elapsed_ms", 0, 300000, &findings)
	validateInt(cfg.Exporter.Retry.InitialIntervalMS, "telemetry.exporter.retry.initial_interval_ms", 50, 30000, &findings)
	validateInt(cfg.Exporter.Retry.MaxIntervalMS, "telemetry.exporter.retry.max_interval_ms", 100, 60000, &findings)
	validateInt(cfg.Logs.BodyMaxChars, "telemetry.logs.body_max_chars", 0, 8192, &findings)
	validateInt(cfg.Processor.MaxQueueSize, "telemetry.processor.max_queue_size", 1, 65536, &findings)
	validateInt(cfg.Processor.MaxExportBatchSize, "telemetry.processor.max_export_batch_size", 1, cfg.Processor.MaxQueueSize, &findings)
	validateInt(cfg.Processor.Traces.ScheduleDelayMS, "telemetry.processor.traces.schedule_delay_ms", 100, 300000, &findings)
	validateInt(cfg.Processor.Metrics.ScheduleDelayMS, "telemetry.processor.metrics.schedule_delay_ms", 100, 300000, &findings)
	validateInt(cfg.Processor.Logs.ScheduleDelayMS, "telemetry.processor.logs.schedule_delay_ms", 100, 300000, &findings)
	validateInt(cfg.Processor.ExportTimeoutMS, "telemetry.processor.export_timeout_ms", 100, 10000, &findings)
	validateInt(cfg.Shutdown.FlushTimeoutMS, "telemetry.shutdown.flush_timeout_ms", 100, 30000, &findings)
	validateToken(cfg.Resource.ServiceName, "telemetry.resource.service_name", &findings)
	validateToken(cfg.Resource.ServiceNamespace, "telemetry.resource.service_namespace", &findings)
	validateVersion(cfg.Resource.ServiceVersion, &findings)
	if cfg.Resource.DeploymentEnvironmentName != "" {
		validateToken(cfg.Resource.DeploymentEnvironmentName, "telemetry.resource.deployment_environment_name", &findings)
	}
	validateInstanceID(cfg.Resource.ServiceInstanceID, presence, &findings)
	validateHeaders(cfg.Exporter.Headers, &findings)
	if !cfg.Attribute.HMACSecretRef.Empty() {
		validateSecretRef(cfg.Attribute.HMACSecretRef, "telemetry.attribute.hmac_secret_ref", &findings)
	}

	endpointConfigured := cfg.Exporter.Endpoint != ""
	endpointDefined := defined(presence, "telemetry", "exporter", "endpoint")
	switch cfg.Exporter.Kind {
	case "none":
		if endpointConfigured {
			appendFinding(&findings, "telemetry.exporter.endpoint", "endpoint must be omitted when telemetry.exporter.kind is none")
		}
		if cfg.Exporter.Protocol != "" || defined(presence, "telemetry", "exporter", "protocol") {
			appendFinding(&findings, "telemetry.exporter.protocol", "protocol must be derived only when telemetry export is enabled")
		}
	case "otlp_http":
		if !endpointConfigured {
			appendFinding(&findings, "telemetry.exporter.endpoint", "endpoint is required when telemetry.exporter.kind is otlp_http")
		} else {
			validateHTTPEndpoint(cfg.Exporter.Endpoint, "telemetry.exporter.endpoint", &findings)
		}
		if cfg.Exporter.Protocol != "http/protobuf" {
			appendFinding(&findings, "telemetry.exporter.protocol", "otlp_http requires protocol http/protobuf")
		}
	case "otlp_grpc":
		if !endpointConfigured {
			appendFinding(&findings, "telemetry.exporter.endpoint", "endpoint is required when telemetry.exporter.kind is otlp_grpc")
		} else {
			validateGRPCEndpoint(cfg.Exporter.Endpoint, "telemetry.exporter.endpoint", &findings)
		}
		if cfg.Exporter.Protocol != "grpc" {
			appendFinding(&findings, "telemetry.exporter.protocol", "otlp_grpc requires protocol grpc")
		}
	}
	if endpointDefined && cfg.Exporter.Endpoint == "" {
		appendFinding(&findings, "telemetry.exporter.endpoint", "endpoint must not be empty")
	}
	if cfg.Exporter.Retry.MaxIntervalMS < cfg.Exporter.Retry.InitialIntervalMS {
		appendFinding(&findings, "telemetry.exporter.retry.max_interval_ms", "max_interval_ms must be greater than or equal to initial_interval_ms")
	}
	if cfg.Processor.MaxExportBatchSize > cfg.Processor.MaxQueueSize {
		appendFinding(&findings, "telemetry.processor.max_export_batch_size", "max_export_batch_size must be less than or equal to max_queue_size")
	}
	validateSamplerConsistency(cfg.Traces.SamplerProfile, cfg.Traces.SampleRatio, &findings)
	if cfg.Attribute.IncidentCorrelation == "hmac_64bit" && cfg.Attribute.HMACSecretRef.Empty() {
		appendFinding(&findings, "telemetry.attribute.hmac_secret_ref", "hmac_secret_ref is required when incident correlation is hmac_64bit")
	}

	return cfg, findings
}

func applyDefaults(cfg *Config, presence Presence) {
	defaultBool(&cfg.Enabled, true, presence, "telemetry", "enabled")
	defaultString(&cfg.Exporter.Kind, "none", presence, "telemetry", "exporter", "kind")
	defaultString(&cfg.Exporter.Compression, "none", presence, "telemetry", "exporter", "compression")
	defaultBool(&cfg.Exporter.Retry.Enabled, true, presence, "telemetry", "exporter", "retry", "enabled")
	defaultInt(&cfg.Exporter.Retry.MaxElapsedMS, 30000, presence, "telemetry", "exporter", "retry", "max_elapsed_ms")
	defaultInt(&cfg.Exporter.Retry.InitialIntervalMS, 100, presence, "telemetry", "exporter", "retry", "initial_interval_ms")
	defaultInt(&cfg.Exporter.Retry.MaxIntervalMS, 5000, presence, "telemetry", "exporter", "retry", "max_interval_ms")
	defaultFloat(&cfg.Exporter.Retry.Multiplier, 2, presence, "telemetry", "exporter", "retry", "multiplier")
	defaultBool(&cfg.Traces.Enabled, true, presence, "telemetry", "traces", "enabled")
	defaultFloat(&cfg.Traces.SampleRatio, 0.10, presence, "telemetry", "traces", "sample_ratio")
	defaultString(&cfg.Traces.SamplerProfile, "auto", presence, "telemetry", "traces", "sampler_profile")
	defaultBool(&cfg.Metrics.Enabled, true, presence, "telemetry", "metrics", "enabled")
	defaultString(&cfg.Metrics.TemporalityProfile, "cartulary.metrics.temporality.cumulative.v1", presence, "telemetry", "metrics", "temporality_profile")
	defaultInt(&cfg.Logs.BodyMaxChars, 2048, presence, "telemetry", "logs", "body_max_chars")
	defaultInt(&cfg.Processor.MaxQueueSize, 2048, presence, "telemetry", "processor", "max_queue_size")
	defaultInt(&cfg.Processor.MaxExportBatchSize, 512, presence, "telemetry", "processor", "max_export_batch_size")
	defaultInt(&cfg.Processor.Traces.ScheduleDelayMS, 5000, presence, "telemetry", "processor", "traces", "schedule_delay_ms")
	defaultInt(&cfg.Processor.Metrics.ScheduleDelayMS, 60000, presence, "telemetry", "processor", "metrics", "schedule_delay_ms")
	defaultInt(&cfg.Processor.Logs.ScheduleDelayMS, 1000, presence, "telemetry", "processor", "logs", "schedule_delay_ms")
	defaultInt(&cfg.Processor.ExportTimeoutMS, 2000, presence, "telemetry", "processor", "export_timeout_ms")
	defaultString(&cfg.Processor.OverflowPolicy, "drop_new", presence, "telemetry", "processor", "overflow_policy")
	defaultInt(&cfg.Shutdown.FlushTimeoutMS, 5000, presence, "telemetry", "shutdown", "flush_timeout_ms")
	defaultBool(&cfg.SelfDiagnostics.Enabled, true, presence, "telemetry", "self_diagnostics", "enabled")
	defaultString(&cfg.SelfDiagnostics.RecursionGuard, "drop_telemetry_about_telemetry", presence, "telemetry", "self_diagnostics", "recursion_guard")
	defaultString(&cfg.Resource.ServiceName, "cartulary.app", presence, "telemetry", "resource", "service_name")
	defaultString(&cfg.Resource.ServiceNamespace, "cartulary", presence, "telemetry", "resource", "service_namespace")
	defaultString(&cfg.Resource.ServiceVersion, "0.0.0+unknown", presence, "telemetry", "resource", "service_version")
	if cfg.Resource.ServiceInstanceID == "" && !defined(presence, "telemetry", "resource", "service_instance_id") {
		cfg.Resource.ServiceInstanceID = uuid.NewString()
	}
	defaultString(&cfg.Attribute.IncidentCorrelation, "none", presence, "telemetry", "attribute", "incident_correlation")
	switch cfg.Exporter.Kind {
	case "otlp_http":
		defaultString(&cfg.Exporter.Protocol, "http/protobuf", presence, "telemetry", "exporter", "protocol")
	case "otlp_grpc":
		defaultString(&cfg.Exporter.Protocol, "grpc", presence, "telemetry", "exporter", "protocol")
	}
	if cfg.Exporter.Headers == nil {
		cfg.Exporter.Headers = map[string]SecretRef{}
	}
}

func defaultInt(target *int64, value int64, presence Presence, path ...string) {
	if *target == 0 && !defined(presence, path...) {
		*target = value
	}
}

func defaultBool(target *bool, value bool, presence Presence, path ...string) {
	if !*target && !defined(presence, path...) {
		*target = value
	}
}

func defaultString(target *string, value string, presence Presence, path ...string) {
	if *target == "" && !defined(presence, path...) {
		*target = value
	}
}

func defaultFloat(target *float64, value float64, presence Presence, path ...string) {
	if *target == 0 && !defined(presence, path...) {
		*target = value
	}
}

// ApplyOverlay parses one telemetry overlay using owner-defined grammar.
func ApplyOverlay(cfg *Config, segments []string, raw string) *Finding {
	path := strings.Join(segments, ".")
	if cfg == nil || len(segments) < 2 || segments[0] != "telemetry" {
		return &Finding{Path: path, ReasonCode: "unknown_key", Message: "unknown telemetry overlay"}
	}
	if raw == "" {
		return nil
	}
	if raw == "null" {
		return invalid(path, "explicit null is not accepted in server-side environment bindings")
	}
	if path == "telemetry.exporter.headers" {
		headers, finding := parseHeaderOverlay(raw, path)
		if finding != nil {
			return finding
		}
		cfg.Exporter.Headers = headers
		return nil
	}
	if path == "telemetry.attribute.hmac_secret_ref" {
		if !validSecretRefName(raw) {
			return invalid(path, "telemetry secret reference name is invalid")
		}
		cfg.Attribute.HMACSecretRef = SecretRef{Kind: "env", Name: raw}
		return nil
	}

	value := reflect.ValueOf(cfg).Elem()
	for _, segment := range segments[1 : len(segments)-1] {
		field, ok := taggedField(value, segment)
		if !ok {
			return &Finding{Path: path, ReasonCode: "unknown_key", Message: fmt.Sprintf("unknown overlay path segment %q", segment)}
		}
		value = field
	}
	field, ok := taggedField(value, segments[len(segments)-1])
	if !ok {
		return &Finding{Path: path, ReasonCode: "unknown_key", Message: fmt.Sprintf("unknown overlay target %q", segments[len(segments)-1])}
	}
	if isInteger(field.Kind()) && !asciiDigits(raw) {
		return invalid(path, "telemetry integer overlay must use unsigned base-10 ASCII digits")
	}
	if isFloat(field.Kind()) && !decimalPattern.MatchString(raw) {
		return invalid(path, "telemetry decimal overlay must be finite decimal notation")
	}
	if err := assignOverlay(field, raw); err != nil {
		return invalid(path, err.Error())
	}
	return nil
}

func parseHeaderOverlay(raw string, path string) (map[string]SecretRef, *Finding) {
	headers := make(map[string]SecretRef)
	seen := make(map[string]struct{})
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			return nil, invalid(path, "telemetry exporter header overlay contains an empty pair")
		}
		name, refName, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, invalid(path, "telemetry exporter header overlay must use key=secret_ref_name pairs")
		}
		name = strings.TrimSpace(name)
		refName = strings.TrimSpace(refName)
		canonical := strings.ToLower(name)
		if !validHeaderName(name) || !validSecretRefName(refName) {
			return nil, invalid(path, "telemetry exporter header overlay contains an invalid header or secret reference")
		}
		if _, duplicate := seen[canonical]; duplicate {
			return nil, invalid(path, "telemetry exporter header overlay contains a duplicate header")
		}
		seen[canonical] = struct{}{}
		headers[canonical] = SecretRef{Kind: "env", Name: refName}
	}
	return headers, nil
}

func taggedField(value reflect.Value, segment string) (reflect.Value, bool) {
	valueType := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := valueType.Field(index)
		tag := strings.Split(field.Tag.Get("toml"), ",")[0]
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		if tag == segment {
			return value.Field(index), true
		}
	}
	return reflect.Value{}, false
}

func assignOverlay(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parse integer overlay: %w", err)
		}
		field.SetInt(value)
	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("parse boolean overlay: %w", err)
		}
		field.SetBool(value)
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(raw, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse decimal overlay: %w", err)
		}
		field.SetFloat(value)
	default:
		return fmt.Errorf("unsupported overlay target kind %s", field.Kind())
	}
	return nil
}

// ResolvedSecret is one owner-purpose-bound immutable secret value.
type ResolvedSecret struct {
	Name    string
	Purpose string
	Value   []byte
}

// ResolvedSecrets contains preflight-only values and never enters a snapshot.
type ResolvedSecrets struct {
	ExporterHeaders map[string]string
	Secrets         []ResolvedSecret
}

// ResolveSecrets performs the explicit telemetry secret preflight.
func ResolveSecrets(cfg Config, env map[string]string) (ResolvedSecrets, []Finding) {
	result := ResolvedSecrets{ExporterHeaders: make(map[string]string, len(cfg.Exporter.Headers))}
	findings := make([]Finding, 0)
	headerBytes := 0
	for name, ref := range cfg.Exporter.Headers {
		path := "telemetry.exporter.headers." + name
		validateSecretRef(ref, path, &findings)
		if value, ok := resolvedSecret(ref, path, env, &findings); ok {
			canonical := strings.ToLower(name)
			headerBytes += len(canonical) + len(": ") + len(value)
			result.ExporterHeaders[canonical] = value
			result.Secrets = append(result.Secrets, ResolvedSecret{Name: ref.Name, Purpose: path, Value: []byte(value)})
		}
	}
	if headerBytes > 8192 {
		appendFinding(&findings, "telemetry.exporter.headers", "configured telemetry exporter header block must be at most 8192 bytes")
	}
	if !cfg.Attribute.HMACSecretRef.Empty() {
		path := "telemetry.attribute.hmac_secret_ref"
		validateSecretRef(cfg.Attribute.HMACSecretRef, path, &findings)
		if value, ok := resolvedSecret(cfg.Attribute.HMACSecretRef, path, env, &findings); ok {
			result.Secrets = append(result.Secrets, ResolvedSecret{
				Name: cfg.Attribute.HMACSecretRef.Name, Purpose: path, Value: []byte(value),
			})
		}
	}
	return result, findings
}

func resolvedSecret(ref SecretRef, path string, env map[string]string, findings *[]Finding) (string, bool) {
	value, ok := lookupEnv(env, secretEnvName(ref.Name))
	if !ok || value == "" || !validResolvedSecret(value) {
		appendFinding(findings, path, "telemetry secret reference could not be resolved to a safe value")
		return "", false
	}
	return value, true
}

func validateEnum(value string, path string, allowed []string, findings *[]Finding) {
	for _, candidate := range allowed {
		if value == candidate {
			return
		}
	}
	appendFinding(findings, path, "value is outside the adopted telemetry enum")
}

func validateFalse(value bool, path string, findings *[]Finding) {
	if value {
		appendFinding(findings, path, "value must be false in the adopted telemetry profile")
	}
}

func validateInt(value int64, path string, min int64, max int64, findings *[]Finding) {
	if value < min {
		appendFinding(findings, path, fmt.Sprintf("value must be at least %d", min))
	} else if value > max {
		appendFinding(findings, path, fmt.Sprintf("value must be at most %d", max))
	}
}

func validateFloat(value float64, path string, min float64, max float64, findings *[]Finding) {
	if value < min || value > max {
		appendFinding(findings, path, fmt.Sprintf("value must be between %.2f and %.2f", min, max))
	}
}

func validateToken(value string, path string, findings *[]Finding) {
	if !tokenPattern.MatchString(value) {
		appendFinding(findings, path, "value must be 1..128 ASCII letters, digits, '.', '_', or '-'")
	}
}

func validateVersion(value string, findings *[]Finding) {
	if value != "0.0.0+unknown" && (len(value) > 128 || !semverPattern.MatchString(value)) {
		appendFinding(findings, "telemetry.resource.service_version", "service_version must be SemVer 2.0.0 or 0.0.0+unknown")
	}
}

func validateInstanceID(value string, presence Presence, findings *[]Finding) {
	if value == "" {
		if defined(presence, "telemetry", "resource", "service_instance_id") {
			appendFinding(findings, "telemetry.resource.service_instance_id", "service_instance_id must be a canonical lowercase UUID v4 when configured")
		}
		return
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.Version() != 4 || value != strings.ToLower(parsed.String()) || parsed == uuid.Nil {
		appendFinding(findings, "telemetry.resource.service_instance_id", "service_instance_id must be a canonical lowercase non-nil UUID v4")
	}
}

func validateHeaders(headers map[string]SecretRef, findings *[]Finding) {
	if len(headers) > 16 {
		appendFinding(findings, "telemetry.exporter.headers", "at most 16 telemetry exporter headers are allowed")
	}
	seen := make(map[string]string)
	for name, ref := range headers {
		canonical := strings.ToLower(name)
		path := "telemetry.exporter.headers." + name
		if !validHeaderName(name) {
			appendFinding(findings, path, "telemetry exporter header name is invalid")
		}
		if _, forbidden := protocolOwnedHeaders[canonical]; forbidden {
			appendFinding(findings, path, "telemetry exporter header must not override a protocol-owned header")
		}
		if previous, duplicate := seen[canonical]; duplicate && previous != name {
			appendFinding(findings, path, "telemetry exporter header duplicates another header after lowercase canonicalization")
		}
		seen[canonical] = name
		validateSecretRef(ref, path, findings)
	}
}

func validateSecretRef(ref SecretRef, path string, findings *[]Finding) {
	if ref.Kind != "env" || !validSecretRefName(ref.Name) {
		appendFinding(findings, path, "telemetry secret references must use secret_ref_v1 kind env and a safe name")
	}
}

func validateSamplerConsistency(profile string, ratio float64, findings *[]Finding) {
	switch profile {
	case "cartulary.sampler.always_off.v1":
		if ratio != 0 {
			appendFinding(findings, "telemetry.traces.sampler_profile", "always_off sampler requires sample_ratio 0.0")
		}
	case "cartulary.sampler.always_on.v1":
		if ratio != 1 {
			appendFinding(findings, "telemetry.traces.sampler_profile", "always_on sampler requires sample_ratio 1.0")
		}
	case "cartulary.sampler.traceidratio_compat.v1":
		if ratio <= 0 || ratio >= 1 {
			appendFinding(findings, "telemetry.traces.sampler_profile", "traceidratio sampler requires 0.0 < sample_ratio < 1.0")
		}
	}
}

func validateHTTPEndpoint(raw string, path string, findings *[]Finding) {
	parsed, err := url.Parse(raw)
	if err != nil || !validEndpointBase(parsed) || parsed.Path == "" && strings.HasSuffix(raw, "//") {
		appendFinding(findings, path, "OTLP/HTTP endpoint must be an absolute http(s) URL with explicit port and no userinfo, query, or fragment")
		return
	}
	if parsed.Path != "" && parsed.Path != "/" {
		if strings.Contains(parsed.EscapedPath(), "%") || strings.Contains(parsed.Path, "//") {
			appendFinding(findings, path, "OTLP/HTTP endpoint path must not contain encoded or duplicate slash segments")
			return
		}
		for _, segment := range strings.Split(strings.Trim(parsed.Path, "/"), "/") {
			if segment == "" || segment == "." || segment == ".." || !pathSegmentPattern.MatchString(segment) {
				appendFinding(findings, path, "OTLP/HTTP endpoint path contains an unsupported segment")
				return
			}
		}
	}
}

func validateGRPCEndpoint(raw string, path string, findings *[]Finding) {
	parsed, err := url.Parse(raw)
	if err != nil || !validEndpointBase(parsed) {
		appendFinding(findings, path, "OTLP/gRPC endpoint must be an absolute http(s) URL with explicit port and no userinfo, query, fragment, or path")
		return
	}
	if parsed.Path != "" && parsed.Path != "/" {
		appendFinding(findings, path, "OTLP/gRPC endpoint must not include a non-root path")
	}
}

func validEndpointBase(parsed *url.URL) bool {
	if parsed == nil || parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	host := parsed.Hostname()
	if host == "" || parsed.Port() == "" || strings.ContainsAny(host, "\t\n\r ") || strings.Contains(host, "%") {
		return false
	}
	if strings.Contains(host, ":") {
		return net.ParseIP(host) != nil
	}
	if strings.Contains(strings.ToLower(host), "xn--") {
		return false
	}
	for _, character := range host {
		if character > 127 {
			return false
		}
	}
	return endpointHostPattern.MatchString(host)
}

func validHeaderName(value string) bool {
	return headerNamePattern.MatchString(value)
}

func validSecretRefName(value string) bool {
	return secretRefPattern.MatchString(value) && normalizedSecretSuffix(value) != ""
}

func secretEnvName(name string) string {
	return "CARTULARY_SECRET_" + normalizedSecretSuffix(name)
}

func normalizedSecretSuffix(name string) string {
	var builder strings.Builder
	previousUnderscore := false
	for _, character := range name {
		var next rune
		switch {
		case character >= 'a' && character <= 'z':
			next = character - ('a' - 'A')
		case character >= 'A' && character <= 'Z', character >= '0' && character <= '9':
			next = character
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

func validResolvedSecret(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || !utf8.ValidString(value) || len(value) > 4096 {
		return false
	}
	for _, character := range value {
		if character == '\n' || character == '\r' || character == '\x00' || character == 0x7f || character < 0x20 || character > 0x7e {
			return false
		}
	}
	return true
}

func lookupEnv(env map[string]string, name string) (string, bool) {
	if env == nil {
		return os.LookupEnv(name)
	}
	value, ok := env[name]
	return value, ok
}

func appendFinding(findings *[]Finding, path string, message string) {
	*findings = append(*findings, Finding{Path: path, ReasonCode: invalidReason, Message: message})
}

func invalid(path string, message string) *Finding {
	return &Finding{Path: path, ReasonCode: invalidReason, Message: message}
}

func defined(presence Presence, path ...string) bool {
	return presence != nil && presence.Defined(path...)
}

func isInteger(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Int64
}

func isFloat(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}

func asciiDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
