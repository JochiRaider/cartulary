package networkflow

import (
	"encoding/json"
	"time"
)

func tableResource(table TableRecord) map[string]any {
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

func rowResource(row FlowRow) map[string]any {
	return map[string]any{
		"network_flow_row_id":          row.RowID,
		"network_flow_table_id":        row.NetworkFlowTableID,
		"incident_id":                  row.IncidentID.String(),
		"source_row_number":            row.SourceRowNumber,
		"source_row_digest_sha256":     row.SourceRowDigestSHA256,
		"normalized_row_digest_sha256": row.NormalizedRowDigestSHA256,
		"mapping_fingerprint":          row.MappingFingerprint,
		FieldFlowStartUTC:              timestamp(row.FlowStartUTC),
		FieldFlowEndUTC:                timestamp(row.FlowEndUTC),
		FieldSrcIP:                     row.SrcIP,
		FieldDstIP:                     row.DstIP,
		FieldSrcPort:                   nullableInt32Value(row.SrcPort),
		FieldDstPort:                   nullableInt32Value(row.DstPort),
		FieldIPProtocol:                row.IPProtocol,
		FieldBytesCount:                row.BytesCount,
		FieldPacketsCount:              row.PacketsCount,
		FieldExporterID:                nullableStringValue(row.ExporterID),
		FieldInputInterface:            nullableStringValue(row.InputInterface),
		FieldOutputInterface:           nullableStringValue(row.OutputInterface),
		FieldTCPFlags:                  nullableInt32Value(row.TCPFlags),
		FieldApplicationLabel:          nullableStringValue(row.ApplicationLabel),
		"unmapped_raw":                 rawJSONValue(row.UnmappedRaw),
		FieldObservationSourceRef:      rawJSONValue(row.ObservationSourceRef),
		"created_at":                   timestamp(row.CreatedAt),
		"created_by_user_id":           row.CreatedByUserID.String(),
	}
}

func rowRefResource(row FlowRow) map[string]any {
	return map[string]any{
		"network_flow_table_id": row.NetworkFlowTableID,
		"network_flow_row_id":   row.RowID,
		"source_row_number":     row.SourceRowNumber,
		"mapping_fingerprint":   row.MappingFingerprint,
	}
}

func storedRowRefResource(ref NetworkFlowRowRef) map[string]any {
	return map[string]any{
		"network_flow_table_id": ref.NetworkFlowTableID,
		"network_flow_row_id":   ref.NetworkFlowRowID,
		"source_row_number":     ref.SourceRowNumber,
		"mapping_fingerprint":   ref.MappingFingerprint,
	}
}

func indicatorBindingResource(binding IndicatorBindingRecord) map[string]any {
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

func diagnosticResource(diagnostic RejectedRowDiagnostic) map[string]any {
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

func sourceProfileResource() map[string]any {
	return map[string]any{
		"source_profile_id":         SourceProfileCiscoSNANetFlowCSV,
		"display_name":              "Cisco Secure Network Analytics NetFlow CSV",
		"conformance_status":        "required_v1",
		"default_parser_profile_id": ParserProfileRFC4180HeaderedCSV,
		"required_field_keys": []string{
			FieldFlowStartUTC,
			FieldFlowEndUTC,
			FieldSrcIP,
			FieldDstIP,
			FieldIPProtocol,
			FieldBytesCount,
			FieldPacketsCount,
		},
		"optional_field_keys": []string{
			FieldSrcPort,
			FieldDstPort,
			FieldExporterID,
			FieldInputInterface,
			FieldOutputInterface,
			FieldTCPFlags,
			FieldApplicationLabel,
		},
		"system_derived_field_keys": []string{FieldObservationSourceRef},
		"supported_timestamp_modes": []string{"rfc3339", "epoch_seconds", "epoch_milliseconds", "netflow_sys_uptime_milliseconds"},
	}
}

func effectiveLimitsResource(limits Limits) map[string]any {
	l := limits.normalized()
	return map[string]any{
		"network_flow.max_active_tables_per_incident":   l.MaxActiveTablesPerIncident,
		"network_flow.max_retained_tables_per_incident": l.MaxRetainedTablesPerIncident,
		"network_flow.max_selected_tables_per_query":    l.MaxSelectedTablesPerQuery,
		"network_flow.max_columns_per_csv":              l.MaxColumnsPerCSV,
		"network_flow.max_header_scalar_length":         l.MaxHeaderScalarLength,
		"network_flow.max_raw_cell_scalar_length":       l.MaxRawCellScalarLength,
		"network_flow.max_rows_per_csv":                 l.MaxRowsPerCSV,
		"network_flow.max_accepted_rows_per_table":      l.MaxAcceptedRowsPerTable,
		"network_flow.max_rejected_row_diagnostics":     l.MaxRejectedRowDiagnostics,
		"network_flow.max_filters_per_query":            l.MaxFiltersPerQuery,
		"network_flow.max_sorts_per_query":              l.MaxSortsPerQuery,
		"network_flow.max_query_limit":                  l.MaxQueryLimit,
		"network_flow.max_graph_vertices":               l.MaxGraphVertices,
		"network_flow.max_graph_edges":                  l.MaxGraphEdges,
		"network_flow.max_example_row_refs_per_edge":    l.MaxExampleRowRefsPerEdge,
		"network_flow.max_binding_source_row_refs":      l.MaxBindingSourceRowRefs,
		"network_flow.max_aggregate_counter_digits":     l.MaxAggregateCounterDigits,
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
