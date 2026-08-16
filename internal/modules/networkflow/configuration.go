package networkflow

import (
	"encoding/json"
	"errors"
	"fmt"
	pathpkg "path"
	"sort"
	"strconv"
	"strings"
)

type Configuration struct {
	Claimed             bool                    `toml:"claimed,omitempty"`
	KeyRingManifestPath string                  `toml:"key_ring_manifest_path,omitempty"`
	ResourceLimits      *ResourceLimitOverrides `toml:"resource_limits,omitempty"`
	effectiveLimits     EffectiveLimits
}

// ResourceLimitOverrides is the closed configuration-major-2 input. Pointers
// distinguish omission from owner-approved zero values.
type ResourceLimitOverrides struct {
	MaxActiveTablesPerIncident         *int64 `toml:"max_active_tables_per_incident,omitempty" json:"max_active_tables_per_incident,omitempty"`
	MaxRetainedTablesPerIncident       *int64 `toml:"max_retained_tables_per_incident,omitempty" json:"max_retained_tables_per_incident,omitempty"`
	MaxSelectedTablesPerQuery          *int64 `toml:"max_selected_tables_per_query,omitempty" json:"max_selected_tables_per_query,omitempty"`
	MaxColumnsPerCSV                   *int64 `toml:"max_columns_per_csv,omitempty" json:"max_columns_per_csv,omitempty"`
	MaxHeaderScalarLength              *int64 `toml:"max_header_scalar_length,omitempty" json:"max_header_scalar_length,omitempty"`
	MaxRawCellScalarLength             *int64 `toml:"max_raw_cell_scalar_length,omitempty" json:"max_raw_cell_scalar_length,omitempty"`
	MaxRowsPerCSV                      *int64 `toml:"max_rows_per_csv,omitempty" json:"max_rows_per_csv,omitempty"`
	MaxAcceptedRowsPerTable            *int64 `toml:"max_accepted_rows_per_table,omitempty" json:"max_accepted_rows_per_table,omitempty"`
	MaxRejectedRowDiagnostics          *int64 `toml:"max_rejected_row_diagnostics,omitempty" json:"max_rejected_row_diagnostics,omitempty"`
	MaxFiltersPerQuery                 *int64 `toml:"max_filters_per_query,omitempty" json:"max_filters_per_query,omitempty"`
	MaxSortsPerQuery                   *int64 `toml:"max_sorts_per_query,omitempty" json:"max_sorts_per_query,omitempty"`
	MaxQueryLimit                      *int64 `toml:"max_query_limit,omitempty" json:"max_query_limit,omitempty"`
	MaxGraphVertices                   *int64 `toml:"max_graph_vertices,omitempty" json:"max_graph_vertices,omitempty"`
	MaxGraphEdges                      *int64 `toml:"max_graph_edges,omitempty" json:"max_graph_edges,omitempty"`
	MaxActiveGraphViewsPerIncident     *int64 `toml:"max_active_graph_views_per_incident,omitempty" json:"max_active_graph_views_per_incident,omitempty"`
	MaxRetainedGraphViewsPerIncident   *int64 `toml:"max_retained_graph_views_per_incident,omitempty" json:"max_retained_graph_views_per_incident,omitempty"`
	MaxNonterminalGraphJobsPerIncident *int64 `toml:"max_nonterminal_graph_jobs_per_incident,omitempty" json:"max_nonterminal_graph_jobs_per_incident,omitempty"`
	MaxExampleRowRefsPerEdge           *int64 `toml:"max_example_row_refs_per_edge,omitempty" json:"max_example_row_refs_per_edge,omitempty"`
	MaxBindingSourceRowRefs            *int64 `toml:"max_binding_source_row_refs,omitempty" json:"max_binding_source_row_refs,omitempty"`
	MaxAggregateCounterDigits          *int64 `toml:"max_aggregate_counter_digits,omitempty" json:"max_aggregate_counter_digits,omitempty"`
	MaxContributingRowsPerGraph        *int64 `toml:"max_contributing_rows_per_graph,omitempty" json:"max_contributing_rows_per_graph,omitempty"`
	MaxTimeBucketsPerGraph             *int64 `toml:"max_time_buckets_per_graph,omitempty" json:"max_time_buckets_per_graph,omitempty"`
	GraphMaterializationTimeoutSeconds *int64 `toml:"graph_materialization_timeout_seconds,omitempty" json:"graph_materialization_timeout_seconds,omitempty"`
}

// EffectiveResourceLimits returns the resolved immutable runtime value.
func (configuration Configuration) EffectiveResourceLimits() EffectiveLimits {
	return configuration.effectiveLimits
}

func CloneConfiguration(configuration Configuration) Configuration {
	cloned := configuration
	cloned.ResourceLimits = cloneResourceLimitOverrides(configuration.ResourceLimits)
	return cloned
}

// ApplyConfigurationOverlay parses one owner-scoped environment binding.
func ApplyConfigurationOverlay(configuration Configuration, path []string, raw string) (Configuration, *ConfigurationFinding) {
	joined := strings.Join(path, ".")
	switch joined {
	case "network_flow_activity.claimed":
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return configuration, &ConfigurationFinding{Path: joined, ReasonCode: "type_mismatch", Message: fmt.Sprintf("parse boolean overlay: %v", err)}
		}
		configuration.Claimed = value
	case keyRingConfigPath:
		configuration.KeyRingManifestPath = raw
	case "network_flow_activity.resource_limits":
		limits, finding := decodeResourceLimitOverlay(raw)
		if finding != nil {
			return configuration, finding
		}
		configuration.ResourceLimits = limits
	default:
		const prefix = "network_flow_activity.resource_limits."
		if strings.HasPrefix(joined, prefix) {
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return configuration, &ConfigurationFinding{Path: joined, ReasonCode: "type_mismatch", Message: "resource limit must be an integer"}
			}
			if configuration.ResourceLimits == nil {
				configuration.ResourceLimits = &ResourceLimitOverrides{}
			}
			if !setResourceLimitOverride(configuration.ResourceLimits, strings.TrimPrefix(joined, prefix), value) {
				return configuration, &ConfigurationFinding{Path: joined, ReasonCode: "unknown_key", Message: "unknown Network Flow resource limit"}
			}
			break
		}
		return configuration, &ConfigurationFinding{Path: joined, ReasonCode: "unknown_key", Message: "unknown Network Flow configuration overlay"}
	}
	return configuration, nil
}

type ConfigurationFinding struct {
	Path       string
	ReasonCode string
	Message    string
}

type ConfigurationError struct {
	Finding ConfigurationFinding
}

func (err *ConfigurationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", err.Finding.Path, err.Finding.ReasonCode, err.Finding.Message)
}

func ConfigurationFindingFromError(err error) (ConfigurationFinding, bool) {
	var configurationError *ConfigurationError
	if !errors.As(err, &configurationError) {
		return ConfigurationFinding{}, false
	}
	return configurationError.Finding, true
}

// NormalizeAndValidateConfiguration is pure owner policy. It performs no file,
// secret, network, database, process, or other external access.
func NormalizeAndValidateConfiguration(configuration Configuration) (Configuration, []ConfigurationFinding) {
	configuration.effectiveLimits = DefaultEffectiveLimits()
	if !configuration.Claimed {
		configuration.KeyRingManifestPath = ""
		configuration.ResourceLimits = nil
		return configuration, nil
	}
	findings := make([]ConfigurationFinding, 0)
	if configuration.KeyRingManifestPath == "" {
		findings = append(findings,
			ConfigurationFinding{
				Path:       keyRingConfigPath,
				ReasonCode: "network_flow_cursor_key_missing",
				Message:    "Network Flow cursor key-ring configuration is required when Network Flow Activity is claimed",
			},
			ConfigurationFinding{
				Path:       keyRingConfigPath,
				ReasonCode: "network_flow_safe_digest_key_missing",
				Message:    "Network Flow safe-digest key-ring configuration is required when Network Flow Activity is claimed",
			},
		)
	} else {
		normalized, finding := normalizeManifestPath(configuration.KeyRingManifestPath)
		if finding != nil {
			findings = append(findings, *finding)
		} else {
			configuration.KeyRingManifestPath = normalized
		}
	}

	effective, limitFindings := resolveResourceLimits(configuration.ResourceLimits)
	configuration.effectiveLimits = effective
	findings = append(findings, limitFindings...)
	return configuration, findings
}

func ValidateEffectiveLimits(limits EffectiveLimits) error {
	findings := validateEffectiveLimits(limits)
	if len(findings) == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", findings[0].Path, findings[0].ReasonCode)
}

func resolveResourceLimits(overrides *ResourceLimitOverrides) (EffectiveLimits, []ConfigurationFinding) {
	effective := DefaultEffectiveLimits()
	if overrides != nil {
		applyLimitOverride(&effective.MaxActiveTablesPerIncident, overrides.MaxActiveTablesPerIncident)
		applyLimitOverride(&effective.MaxRetainedTablesPerIncident, overrides.MaxRetainedTablesPerIncident)
		applyLimitOverride(&effective.MaxSelectedTablesPerQuery, overrides.MaxSelectedTablesPerQuery)
		applyLimitOverride(&effective.MaxColumnsPerCSV, overrides.MaxColumnsPerCSV)
		applyLimitOverride(&effective.MaxHeaderScalarLength, overrides.MaxHeaderScalarLength)
		applyLimitOverride(&effective.MaxRawCellScalarLength, overrides.MaxRawCellScalarLength)
		applyLimitOverride(&effective.MaxRowsPerCSV, overrides.MaxRowsPerCSV)
		applyLimitOverride(&effective.MaxAcceptedRowsPerTable, overrides.MaxAcceptedRowsPerTable)
		applyLimitOverride(&effective.MaxRejectedRowDiagnostics, overrides.MaxRejectedRowDiagnostics)
		applyLimitOverride(&effective.MaxFiltersPerQuery, overrides.MaxFiltersPerQuery)
		applyLimitOverride(&effective.MaxSortsPerQuery, overrides.MaxSortsPerQuery)
		applyLimitOverride(&effective.MaxQueryLimit, overrides.MaxQueryLimit)
		applyLimitOverride(&effective.MaxGraphVertices, overrides.MaxGraphVertices)
		applyLimitOverride(&effective.MaxGraphEdges, overrides.MaxGraphEdges)
		applyLimitOverride(&effective.MaxActiveGraphViewsPerIncident, overrides.MaxActiveGraphViewsPerIncident)
		applyLimitOverride(&effective.MaxRetainedGraphViewsPerIncident, overrides.MaxRetainedGraphViewsPerIncident)
		applyLimitOverride(&effective.MaxNonterminalGraphJobsPerIncident, overrides.MaxNonterminalGraphJobsPerIncident)
		applyLimitOverride(&effective.MaxExampleRowRefsPerEdge, overrides.MaxExampleRowRefsPerEdge)
		applyLimitOverride(&effective.MaxBindingSourceRowRefs, overrides.MaxBindingSourceRowRefs)
		applyLimitOverride(&effective.MaxAggregateCounterDigits, overrides.MaxAggregateCounterDigits)
		applyLimitOverride(&effective.MaxContributingRowsPerGraph, overrides.MaxContributingRowsPerGraph)
		applyLimitOverride(&effective.MaxTimeBucketsPerGraph, overrides.MaxTimeBucketsPerGraph)
		applyLimitOverride(&effective.GraphMaterializationTimeoutSeconds, overrides.GraphMaterializationTimeoutSeconds)
	}
	return effective, validateEffectiveLimits(effective)
}

func applyLimitOverride(target *int64, override *int64) {
	if override != nil {
		*target = *override
	}
}

type effectiveLimitCheck struct {
	key     string
	value   int64
	minimum int64
	maximum int64
}

func validateEffectiveLimits(limits EffectiveLimits) []ConfigurationFinding {
	checks := []effectiveLimitCheck{
		{"max_active_tables_per_incident", limits.MaxActiveTablesPerIncident, 1, 128},
		{"max_retained_tables_per_incident", limits.MaxRetainedTablesPerIncident, 1, 512},
		{"max_selected_tables_per_query", limits.MaxSelectedTablesPerQuery, 1, 64},
		{"max_columns_per_csv", limits.MaxColumnsPerCSV, 1, 512},
		{"max_header_scalar_length", limits.MaxHeaderScalarLength, 1, 256},
		{"max_raw_cell_scalar_length", limits.MaxRawCellScalarLength, 1, 16384},
		{"max_rows_per_csv", limits.MaxRowsPerCSV, 1, 5000000},
		{"max_accepted_rows_per_table", limits.MaxAcceptedRowsPerTable, 1, 5000000},
		{"max_rejected_row_diagnostics", limits.MaxRejectedRowDiagnostics, 0, 100000},
		{"max_filters_per_query", limits.MaxFiltersPerQuery, 0, 16},
		{"max_sorts_per_query", limits.MaxSortsPerQuery, 0, 8},
		{"max_query_limit", limits.MaxQueryLimit, 1, 1000},
		{"max_graph_vertices", limits.MaxGraphVertices, 1, 100000},
		{"max_graph_edges", limits.MaxGraphEdges, 0, 250000},
		{"max_active_graph_views_per_incident", limits.MaxActiveGraphViewsPerIncident, 1, 32},
		{"max_retained_graph_views_per_incident", limits.MaxRetainedGraphViewsPerIncident, 1, 128},
		{"max_nonterminal_graph_jobs_per_incident", limits.MaxNonterminalGraphJobsPerIncident, 1, 4},
		{"max_example_row_refs_per_edge", limits.MaxExampleRowRefsPerEdge, 0, 100},
		{"max_binding_source_row_refs", limits.MaxBindingSourceRowRefs, 1, 1000},
		{"max_aggregate_counter_digits", limits.MaxAggregateCounterDigits, 1, 128},
		{"max_contributing_rows_per_graph", limits.MaxContributingRowsPerGraph, 1, 5000000},
		{"max_time_buckets_per_graph", limits.MaxTimeBucketsPerGraph, 1, 1024},
		{"graph_materialization_timeout_seconds", limits.GraphMaterializationTimeoutSeconds, 30, 3600},
	}
	findings := make([]ConfigurationFinding, 0)
	for _, check := range checks {
		reason := ""
		switch {
		case check.value < check.minimum:
			reason = "below_minimum"
		case check.value > check.maximum:
			reason = "above_maximum"
		}
		if reason != "" {
			findings = append(findings, ConfigurationFinding{
				Path:       "network_flow_activity.resource_limits." + check.key,
				ReasonCode: reason,
				Message:    "Network Flow resource limit is outside its adopted range",
			})
		}
	}
	if limits.MaxActiveTablesPerIncident > limits.MaxRetainedTablesPerIncident {
		findings = append(findings, ConfigurationFinding{
			Path:       "network_flow_activity.resource_limits.max_active_tables_per_incident",
			ReasonCode: "invalid_limit_relationship",
			Message:    "active table limit must not exceed retained table limit",
		})
	}
	if limits.MaxActiveGraphViewsPerIncident > limits.MaxRetainedGraphViewsPerIncident {
		findings = append(findings, ConfigurationFinding{
			Path:       "network_flow_activity.resource_limits.max_active_graph_views_per_incident",
			ReasonCode: "invalid_limit_relationship",
			Message:    "active graph-view limit must not exceed retained graph-view limit",
		})
	}
	return findings
}

func decodeResourceLimitOverlay(raw string) (*ResourceLimitOverrides, *ConfigurationFinding) {
	path := "network_flow_activity.resource_limits"
	if strings.TrimSpace(raw) == "null" {
		return nil, &ConfigurationFinding{Path: path, ReasonCode: "explicit_null", Message: "resource limits must be an object"}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil || object == nil {
		return nil, &ConfigurationFinding{Path: path, ReasonCode: "type_mismatch", Message: "resource limits must be a JSON object"}
	}
	limits := &ResourceLimitOverrides{}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		memberPath := path + "." + key
		if string(object[key]) == "null" {
			return nil, &ConfigurationFinding{Path: memberPath, ReasonCode: "explicit_null", Message: "resource limit must be an integer"}
		}
		var value int64
		if err := json.Unmarshal(object[key], &value); err != nil {
			return nil, &ConfigurationFinding{Path: memberPath, ReasonCode: "type_mismatch", Message: "resource limit must be an integer"}
		}
		if !setResourceLimitOverride(limits, key, value) {
			return nil, &ConfigurationFinding{Path: memberPath, ReasonCode: "unknown_key", Message: "unknown Network Flow resource limit"}
		}
	}
	return limits, nil
}

func setResourceLimitOverride(limits *ResourceLimitOverrides, key string, value int64) bool {
	if limits == nil {
		return false
	}
	switch key {
	case "max_active_tables_per_incident":
		limits.MaxActiveTablesPerIncident = &value
	case "max_retained_tables_per_incident":
		limits.MaxRetainedTablesPerIncident = &value
	case "max_selected_tables_per_query":
		limits.MaxSelectedTablesPerQuery = &value
	case "max_columns_per_csv":
		limits.MaxColumnsPerCSV = &value
	case "max_header_scalar_length":
		limits.MaxHeaderScalarLength = &value
	case "max_raw_cell_scalar_length":
		limits.MaxRawCellScalarLength = &value
	case "max_rows_per_csv":
		limits.MaxRowsPerCSV = &value
	case "max_accepted_rows_per_table":
		limits.MaxAcceptedRowsPerTable = &value
	case "max_rejected_row_diagnostics":
		limits.MaxRejectedRowDiagnostics = &value
	case "max_filters_per_query":
		limits.MaxFiltersPerQuery = &value
	case "max_sorts_per_query":
		limits.MaxSortsPerQuery = &value
	case "max_query_limit":
		limits.MaxQueryLimit = &value
	case "max_graph_vertices":
		limits.MaxGraphVertices = &value
	case "max_graph_edges":
		limits.MaxGraphEdges = &value
	case "max_active_graph_views_per_incident":
		limits.MaxActiveGraphViewsPerIncident = &value
	case "max_retained_graph_views_per_incident":
		limits.MaxRetainedGraphViewsPerIncident = &value
	case "max_nonterminal_graph_jobs_per_incident":
		limits.MaxNonterminalGraphJobsPerIncident = &value
	case "max_example_row_refs_per_edge":
		limits.MaxExampleRowRefsPerEdge = &value
	case "max_binding_source_row_refs":
		limits.MaxBindingSourceRowRefs = &value
	case "max_aggregate_counter_digits":
		limits.MaxAggregateCounterDigits = &value
	case "max_contributing_rows_per_graph":
		limits.MaxContributingRowsPerGraph = &value
	case "max_time_buckets_per_graph":
		limits.MaxTimeBucketsPerGraph = &value
	case "graph_materialization_timeout_seconds":
		limits.GraphMaterializationTimeoutSeconds = &value
	default:
		return false
	}
	return true
}

func cloneResourceLimitOverrides(source *ResourceLimitOverrides) *ResourceLimitOverrides {
	if source == nil {
		return nil
	}
	return &ResourceLimitOverrides{
		MaxActiveTablesPerIncident:         cloneInt64(source.MaxActiveTablesPerIncident),
		MaxRetainedTablesPerIncident:       cloneInt64(source.MaxRetainedTablesPerIncident),
		MaxSelectedTablesPerQuery:          cloneInt64(source.MaxSelectedTablesPerQuery),
		MaxColumnsPerCSV:                   cloneInt64(source.MaxColumnsPerCSV),
		MaxHeaderScalarLength:              cloneInt64(source.MaxHeaderScalarLength),
		MaxRawCellScalarLength:             cloneInt64(source.MaxRawCellScalarLength),
		MaxRowsPerCSV:                      cloneInt64(source.MaxRowsPerCSV),
		MaxAcceptedRowsPerTable:            cloneInt64(source.MaxAcceptedRowsPerTable),
		MaxRejectedRowDiagnostics:          cloneInt64(source.MaxRejectedRowDiagnostics),
		MaxFiltersPerQuery:                 cloneInt64(source.MaxFiltersPerQuery),
		MaxSortsPerQuery:                   cloneInt64(source.MaxSortsPerQuery),
		MaxQueryLimit:                      cloneInt64(source.MaxQueryLimit),
		MaxGraphVertices:                   cloneInt64(source.MaxGraphVertices),
		MaxGraphEdges:                      cloneInt64(source.MaxGraphEdges),
		MaxActiveGraphViewsPerIncident:     cloneInt64(source.MaxActiveGraphViewsPerIncident),
		MaxRetainedGraphViewsPerIncident:   cloneInt64(source.MaxRetainedGraphViewsPerIncident),
		MaxNonterminalGraphJobsPerIncident: cloneInt64(source.MaxNonterminalGraphJobsPerIncident),
		MaxExampleRowRefsPerEdge:           cloneInt64(source.MaxExampleRowRefsPerEdge),
		MaxBindingSourceRowRefs:            cloneInt64(source.MaxBindingSourceRowRefs),
		MaxAggregateCounterDigits:          cloneInt64(source.MaxAggregateCounterDigits),
		MaxContributingRowsPerGraph:        cloneInt64(source.MaxContributingRowsPerGraph),
		MaxTimeBucketsPerGraph:             cloneInt64(source.MaxTimeBucketsPerGraph),
		GraphMaterializationTimeoutSeconds: cloneInt64(source.GraphMaterializationTimeoutSeconds),
	}
}

func cloneInt64(source *int64) *int64 {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

func normalizeManifestPath(raw string) (string, *ConfigurationFinding) {
	if !strings.HasPrefix(raw, "/") {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_not_absolute",
			Message:    "bootstrap manifest path must be an absolute POSIX path",
		}
	}
	if strings.ContainsRune(raw, '\x00') {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_forbidden_segment",
			Message:    "bootstrap manifest path must not contain NUL",
		}
	}
	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_forbidden_segment",
			Message:    "bootstrap manifest path must not use shell expansion segments",
		}
	}
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &ConfigurationFinding{
				Path:       keyRingConfigPath,
				ReasonCode: "path_forbidden_segment",
				Message:    "bootstrap manifest path must not contain . or .. segments",
			}
		}
	}
	return pathpkg.Clean(raw), nil
}

type ManifestReadFailure string

const (
	ManifestUnreadable ManifestReadFailure = "unreadable"
	ManifestUnsafe     ManifestReadFailure = "unsafe_object"
	ManifestTooLarge   ManifestReadFailure = "too_large"
)

func KeyRingManifestReadError(failure ManifestReadFailure) error {
	switch failure {
	case ManifestTooLarge:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_invalid",
			"Network Flow key-ring manifest exceeds 65536 bytes",
		)
	case ManifestUnsafe:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_invalid",
			"Network Flow key-ring manifest path must reference one regular file",
		)
	default:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_missing",
			"read Network Flow key-ring manifest: file is unavailable",
		)
	}
}
