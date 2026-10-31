package networkflow

import (
	"encoding/json"
	generatedmapping "github.com/JochiRaider/cartulary/internal/gen/networkflowmapping"
	"slices"
	"strings"
	"time"
)

func tableResource(table tableRecord) map[string]any {
	return map[string]any{
		"network_flow_table_id":         table.TableID,
		"incident_id":                   table.IncidentID.String(),
		"display_name":                  table.DisplayName,
		"table_version":                 table.TableVersion,
		"table_status":                  table.TableStatus,
		"source_import_session_id":      table.SourceImportSessionID.String(),
		"source_import_unit_id":         table.SourceImportUnitID.String(),
		"source_content_sha256":         table.SourceContentSHA256,
		"source_filename_display":       table.SourceFilenameDisplay,
		"source_filename_digest":        table.SourceFilenameDigest,
		"source_filename_digest_key_id": table.SourceFilenameDigestKeyID,
		"mapping_fingerprint":           table.MappingFingerprint,
		"source_profile_id":             table.SourceProfileID,
		"parser_profile_id":             table.ParserProfileID,
		"row_count_accepted":            table.RowCountAccepted,
		"row_count_rejected":            table.RowCountRejected,
		"diagnostics_truncated":         table.DiagnosticsTruncated,
		"created_by_user_id":            table.CreatedByUserID.String(),
		"created_at":                    timestamp(table.CreatedAt),
		"updated_at":                    timestamp(table.UpdatedAt),
		"deleted_at":                    nullableTimestamp(table.DeletedAt),
	}
}

func rowResource(row flowRow) map[string]any {
	return map[string]any{
		"network_flow_row_id":          row.RowID,
		"network_flow_table_id":        row.NetworkFlowTableID,
		"incident_id":                  row.IncidentID.String(),
		"source_row_number":            row.SourceRowNumber,
		"source_row_digest_sha256":     row.SourceRowDigestSHA256,
		"normalized_row_digest_sha256": row.NormalizedRowDigestSHA256,
		"mapping_fingerprint":          row.MappingFingerprint,
		fieldFlowStartUTC:              timestamp(row.FlowStartUTC),
		fieldFlowEndUTC:                timestamp(row.FlowEndUTC),
		fieldSrcIP:                     row.SrcIP,
		fieldDstIP:                     row.DstIP,
		fieldSrcPort:                   nullableInt32Value(row.SrcPort),
		fieldDstPort:                   nullableInt32Value(row.DstPort),
		fieldIPProtocol:                row.IPProtocol,
		fieldBytesCount:                row.BytesCount,
		fieldPacketsCount:              row.PacketsCount,
		fieldExporterID:                nullableStringValue(row.ExporterID),
		fieldInputInterface:            nullableStringValue(row.InputInterface),
		fieldOutputInterface:           nullableStringValue(row.OutputInterface),
		fieldTCPFlags:                  nullableInt32Value(row.TCPFlags),
		fieldApplicationLabel:          nullableStringValue(row.ApplicationLabel),
		"unmapped_raw":                 rawJSONValue(row.UnmappedRaw),
		fieldObservationSourceRef:      rawJSONValue(row.ObservationSourceRef),
		"created_at":                   timestamp(row.CreatedAt),
		"created_by_user_id":           row.CreatedByUserID.String(),
	}
}

func rowRefResource(row flowRow) map[string]any {
	return map[string]any{
		"network_flow_table_id": row.NetworkFlowTableID,
		"network_flow_row_id":   row.RowID,
		"source_row_number":     row.SourceRowNumber,
		"mapping_fingerprint":   row.MappingFingerprint,
	}
}

func storedRowRefResource(ref networkFlowRowRef) map[string]any {
	return map[string]any{
		"network_flow_table_id": ref.NetworkFlowTableID,
		"network_flow_row_id":   ref.NetworkFlowRowID,
		"source_row_number":     ref.SourceRowNumber,
		"mapping_fingerprint":   ref.MappingFingerprint,
	}
}

func indicatorBindingResource(binding indicatorBindingRecord) map[string]any {
	refs := make([]any, 0, len(binding.SourceRowRefs))
	for _, ref := range binding.SourceRowRefs {
		refs = append(refs, storedRowRefResource(ref))
	}
	normalized := ""
	if binding.TargetIndicator.NormalizedValue != nil {
		normalized = *binding.TargetIndicator.NormalizedValue
	}
	return map[string]any{
		"network_flow_indicator_binding_id": binding.BindingID,
		"incident_id":                       binding.IncidentID.String(),
		"target_indicator_ref": map[string]any{
			"indicator_id":     binding.TargetIndicator.RecordID.String(),
			"indicator_type":   binding.TargetIndicator.IndicatorType,
			"value_kind":       binding.TargetIndicator.ValueKind,
			"normalized_value": normalized,
		},
		"selector_kind":               binding.SelectorKind,
		"candidate_value":             binding.CandidateValue,
		"source_row_refs":             refs,
		"source_row_refs_truncated":   binding.SourceRowRefsTruncated,
		"source_row_refs_total_count": binding.SourceRowRefsTotalCount,
		"created_observation_refs":    []any{},
		"created_by_user_id":          binding.CreatedByUserID.String(),
		"created_at":                  timestamp(binding.CreatedAt),
	}
}

func diagnosticResource(diagnostic rejectedRowDiagnostic) map[string]any {
	return map[string]any{
		"diagnostic_id":         diagnostic.DiagnosticID,
		"source_row_number":     diagnostic.SourceRowNumber,
		"source_column_ordinal": nullableInt64Value(diagnostic.SourceColumnOrdinal),
		"raw_header_sha256":     nullableStringValue(diagnostic.RawHeaderSHA256),
		"field_key":             nullableStringValue(diagnostic.FieldKey),
		"error_code":            diagnostic.ErrorCode,
		"reason_code":           diagnostic.ReasonCode,
		"safe_sample":           nullableStringValue(diagnostic.SafeSample),
		"raw_value_sha256":      nullableStringValue(diagnostic.RawValueSHA256),
		"message_key":           diagnostic.MessageKey,
		"message_args":          rawJSONValue(diagnostic.MessageArgs),
		"message":               diagnostic.Message,
		"limit_name":            nullableStringValue(diagnostic.LimitName),
		"limit_value":           nullableInt64Value(diagnostic.LimitValue),
		"actual_value":          nullableInt64Value(diagnostic.ActualValue),
	}
}

func sourceProfileResources() []any {
	profiles := []any{}
	registry := slices.Clone(generatedmapping.Registry.SourceProfiles)
	slices.SortFunc(registry, func(a, b generatedmapping.SourceProfile) int {
		return strings.Compare(a.SourceProfileID, b.SourceProfileID)
	})
	for _, profile := range registry {
		if profile.ConformanceStatus == "required_v1" {
			profiles = append(profiles, sourceProfileResource(profile))
		}
	}
	return profiles
}

func sourceProfileResource(profile generatedmapping.SourceProfile) map[string]any {
	optional := []string{}
	for _, field := range profile.Fields {
		if field.Requirement == "optional_map_when_present" {
			optional = append(optional, field.FieldKey)
		}
	}
	return map[string]any{
		"source_profile_id":         profile.SourceProfileID,
		"display_name":              profile.DisplayName,
		"conformance_status":        profile.ConformanceStatus,
		"default_parser_profile_id": profile.ParserProfileID,
		"required_field_keys":       requiredFields(profile.SourceProfileID),
		"optional_field_keys":       optional,
		"system_derived_field_keys": []string{fieldObservationSourceRef},
		"supported_timestamp_modes": profile.SupportedTimestampModes,
	}
}

func effectiveLimitsResource(l EffectiveLimits) map[string]any {
	return map[string]any{
		"network_flow.max_active_tables_per_incident":          l.MaxActiveTablesPerIncident,
		"network_flow.max_retained_tables_per_incident":        l.MaxRetainedTablesPerIncident,
		"network_flow.max_selected_tables_per_query":           l.MaxSelectedTablesPerQuery,
		"network_flow.max_columns_per_csv":                     l.MaxColumnsPerCSV,
		"network_flow.max_header_scalar_length":                l.MaxHeaderScalarLength,
		"network_flow.max_raw_cell_scalar_length":              l.MaxRawCellScalarLength,
		"network_flow.max_rows_per_csv":                        l.MaxRowsPerCSV,
		"network_flow.max_accepted_rows_per_table":             l.MaxAcceptedRowsPerTable,
		"network_flow.max_rejected_row_diagnostics":            l.MaxRejectedRowDiagnostics,
		"network_flow.max_filters_per_query":                   l.MaxFiltersPerQuery,
		"network_flow.max_sorts_per_query":                     l.MaxSortsPerQuery,
		"network_flow.max_query_limit":                         l.MaxQueryLimit,
		"network_flow.max_graph_vertices":                      l.MaxGraphVertices,
		"network_flow.max_graph_edges":                         l.MaxGraphEdges,
		"network_flow.max_active_graph_views_per_incident":     l.MaxActiveGraphViewsPerIncident,
		"network_flow.max_retained_graph_views_per_incident":   l.MaxRetainedGraphViewsPerIncident,
		"network_flow.max_nonterminal_graph_jobs_per_incident": l.MaxNonterminalGraphJobsPerIncident,
		"network_flow.max_example_row_refs_per_edge":           l.MaxExampleRowRefsPerEdge,
		"network_flow.max_binding_source_row_refs":             l.MaxBindingSourceRowRefs,
		"network_flow.max_aggregate_counter_digits":            l.MaxAggregateCounterDigits,
		"network_flow.max_contributing_rows_per_graph":         l.MaxContributingRowsPerGraph,
		"network_flow.max_time_buckets_per_graph":              l.MaxTimeBucketsPerGraph,
		"network_flow.graph_materialization_timeout_seconds":   l.GraphMaterializationTimeoutSeconds,
	}
}

func timestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func nullableTimestamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return timestamp(*value)
}

func nullableInt64Value(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func rawJSONValue(raw json.RawMessage) any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return map[string]any{}
	}
	if value == nil {
		return map[string]any{}
	}
	return value
}
