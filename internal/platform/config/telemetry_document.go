package config

// The deployment document retains only a wire-shaped representation of owner
// namespaces. Policy, defaults, validation, overlays, and runtime types belong
// to the registered owner contribution.
type telemetryDocument struct {
	Enabled            bool                             `toml:"enabled"`
	OTelEnvPassthrough bool                             `toml:"otel_env_passthrough"`
	Exporter           telemetryExporterDocument        `toml:"exporter"`
	Traces             telemetryTracesDocument          `toml:"traces"`
	Metrics            telemetryMetricsDocument         `toml:"metrics"`
	Logs               telemetryLogsDocument            `toml:"logs"`
	Processor          telemetryProcessorDocument       `toml:"processor"`
	Shutdown           telemetryShutdownDocument        `toml:"shutdown"`
	SelfDiagnostics    telemetrySelfDiagnosticsDocument `toml:"self_diagnostics"`
	Resource           telemetryResourceDocument        `toml:"resource"`
	Attribute          telemetryAttributeDocument       `toml:"attribute"`
}

type telemetryExporterDocument struct {
	Kind        string                             `toml:"kind"`
	Endpoint    string                             `toml:"endpoint"`
	Protocol    string                             `toml:"protocol"`
	Compression string                             `toml:"compression"`
	Headers     map[string]telemetrySecretDocument `toml:"headers"`
	Retry       telemetryExporterRetryDocument     `toml:"retry"`
}

type telemetryExporterRetryDocument struct {
	Enabled           bool    `toml:"enabled"`
	MaxElapsedMS      int64   `toml:"max_elapsed_ms"`
	InitialIntervalMS int64   `toml:"initial_interval_ms"`
	MaxIntervalMS     int64   `toml:"max_interval_ms"`
	Multiplier        float64 `toml:"multiplier"`
}

type telemetryTracesDocument struct {
	Enabled             bool    `toml:"enabled"`
	SampleRatio         float64 `toml:"sample_ratio"`
	SamplerProfile      string  `toml:"sampler_profile"`
	AcceptRemoteContext bool    `toml:"accept_remote_context"`
}

type telemetryMetricsDocument struct {
	Enabled            bool                      `toml:"enabled"`
	TemporalityProfile string                    `toml:"temporality_profile"`
	Exemplars          telemetryExemplarDocument `toml:"exemplars"`
}

type telemetryExemplarDocument struct {
	Enabled bool `toml:"enabled"`
}

type telemetryLogsDocument struct {
	BridgeEnabled bool  `toml:"bridge_enabled"`
	BodyMaxChars  int64 `toml:"body_max_chars"`
}

type telemetryProcessorDocument struct {
	MaxQueueSize       int64                            `toml:"max_queue_size"`
	MaxExportBatchSize int64                            `toml:"max_export_batch_size"`
	Traces             telemetryProcessorSignalDocument `toml:"traces"`
	Metrics            telemetryProcessorSignalDocument `toml:"metrics"`
	Logs               telemetryProcessorSignalDocument `toml:"logs"`
	ExportTimeoutMS    int64                            `toml:"export_timeout_ms"`
	OverflowPolicy     string                           `toml:"overflow_policy"`
}

type telemetryProcessorSignalDocument struct {
	ScheduleDelayMS int64 `toml:"schedule_delay_ms"`
}

type telemetryShutdownDocument struct {
	FlushTimeoutMS int64 `toml:"flush_timeout_ms"`
}

type telemetrySelfDiagnosticsDocument struct {
	Enabled        bool   `toml:"enabled"`
	RecursionGuard string `toml:"recursion_guard"`
}

type telemetryResourceDocument struct {
	ServiceName               string `toml:"service_name"`
	ServiceNamespace          string `toml:"service_namespace"`
	ServiceVersion            string `toml:"service_version"`
	ServiceInstanceID         string `toml:"service_instance_id"`
	DeploymentEnvironmentName string `toml:"deployment_environment_name"`
}

type telemetryAttributeDocument struct {
	IncidentCorrelation string                  `toml:"incident_correlation"`
	HMACSecretRef       telemetrySecretDocument `toml:"hmac_secret_ref"`
}

type telemetrySecretDocument struct {
	Kind string `toml:"kind"`
	Name string `toml:"name"`
}
