// Package configuration owns the immutable OpenTelemetry deployment settings
// and their pure normalization and structural validation policy.
package configuration

// Config is the closed telemetry.* deployment namespace.
type Config struct {
	Enabled            bool                  `toml:"enabled"`
	OTelEnvPassthrough bool                  `toml:"otel_env_passthrough"`
	Exporter           ExporterConfig        `toml:"exporter"`
	Traces             TracesConfig          `toml:"traces"`
	Metrics            MetricsConfig         `toml:"metrics"`
	Logs               LogsConfig            `toml:"logs"`
	Processor          ProcessorConfig       `toml:"processor"`
	Shutdown           ShutdownConfig        `toml:"shutdown"`
	SelfDiagnostics    SelfDiagnosticsConfig `toml:"self_diagnostics"`
	Resource           ResourceConfig        `toml:"resource"`
	Attribute          AttributeConfig       `toml:"attribute"`
}

type ExporterConfig struct {
	Kind        string               `toml:"kind"`
	Endpoint    string               `toml:"endpoint,omitempty"`
	Protocol    string               `toml:"protocol,omitempty"`
	Compression string               `toml:"compression"`
	Headers     map[string]SecretRef `toml:"headers"`
	Retry       ExporterRetryConfig  `toml:"retry"`
}

type ExporterRetryConfig struct {
	Enabled           bool    `toml:"enabled"`
	MaxElapsedMS      int64   `toml:"max_elapsed_ms"`
	InitialIntervalMS int64   `toml:"initial_interval_ms"`
	MaxIntervalMS     int64   `toml:"max_interval_ms"`
	Multiplier        float64 `toml:"multiplier"`
}

type TracesConfig struct {
	Enabled             bool    `toml:"enabled"`
	SampleRatio         float64 `toml:"sample_ratio"`
	SamplerProfile      string  `toml:"sampler_profile"`
	AcceptRemoteContext bool    `toml:"accept_remote_context"`
}

type MetricsConfig struct {
	Enabled            bool           `toml:"enabled"`
	TemporalityProfile string         `toml:"temporality_profile"`
	Exemplars          ExemplarConfig `toml:"exemplars"`
}

type ExemplarConfig struct {
	Enabled bool `toml:"enabled"`
}

type LogsConfig struct {
	BridgeEnabled bool  `toml:"bridge_enabled"`
	BodyMaxChars  int64 `toml:"body_max_chars"`
}

type ProcessorConfig struct {
	MaxQueueSize       int64                 `toml:"max_queue_size"`
	MaxExportBatchSize int64                 `toml:"max_export_batch_size"`
	Traces             ProcessorSignalConfig `toml:"traces"`
	Metrics            ProcessorSignalConfig `toml:"metrics"`
	Logs               ProcessorSignalConfig `toml:"logs"`
	ExportTimeoutMS    int64                 `toml:"export_timeout_ms"`
	OverflowPolicy     string                `toml:"overflow_policy"`
}

type ProcessorSignalConfig struct {
	ScheduleDelayMS int64 `toml:"schedule_delay_ms"`
}

type ShutdownConfig struct {
	FlushTimeoutMS int64 `toml:"flush_timeout_ms"`
}

type SelfDiagnosticsConfig struct {
	Enabled        bool   `toml:"enabled"`
	RecursionGuard string `toml:"recursion_guard"`
}

type ResourceConfig struct {
	ServiceName               string `toml:"service_name"`
	ServiceNamespace          string `toml:"service_namespace"`
	ServiceVersion            string `toml:"service_version"`
	ServiceInstanceID         string `toml:"service_instance_id"`
	DeploymentEnvironmentName string `toml:"deployment_environment_name"`
}

type AttributeConfig struct {
	IncidentCorrelation string    `toml:"incident_correlation"`
	HMACSecretRef       SecretRef `toml:"hmac_secret_ref"`
}

type SecretRef struct {
	Kind string `toml:"kind"`
	Name string `toml:"name"`
}

func (ref SecretRef) Empty() bool {
	return ref.Kind == "" && ref.Name == ""
}

// Clone makes all map-bearing owner settings independent.
func Clone(source Config) Config {
	cloned := source
	if source.Exporter.Headers == nil {
		return cloned
	}
	cloned.Exporter.Headers = make(map[string]SecretRef, len(source.Exporter.Headers))
	for name, ref := range source.Exporter.Headers {
		cloned.Exporter.Headers[name] = ref
	}
	return cloned
}

// Finding is a pure owner-local structural diagnostic.
type Finding struct {
	Path       string
	ReasonCode string
	Message    string
}

// Presence reports whether an exact deployment path was explicitly supplied.
type Presence interface {
	Defined(path ...string) bool
}
