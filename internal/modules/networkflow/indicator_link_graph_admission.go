package networkflow

import (
	"encoding/json"
)

// The link boundary validates the full echoed query without changing graph or
// contributor admission. Current effective limits apply only after replay.
func decodeLinkGraphQuery(raw json.RawMessage, limits EffectiveLimits) (graphSemanticRequest, *semanticFailure) {
	object, apiErr := linkObject(raw, "graph_query")
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	if apiErr = linkMembers(object, map[string]string{"schema_id": "string", "selected_table_ids": "array", "filters": "array", "time_range": "object", "aggregation": "object"}); apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	var tables []string
	if json.Unmarshal(object["selected_table_ids"], &tables) != nil || len(tables) < 1 || len(tables) > 64 {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("selected_table_ids", "type_mismatch")
	}
	seen := map[string]bool{}
	for _, id := range tables {
		if !linkTableIDPattern.MatchString(id) {
			return graphSemanticRequest{}, invalidNetworkFlowRequest("selected_table_ids", "type_mismatch")
		}
		if seen[id] {
			return graphSemanticRequest{}, &semanticFailure{kind: failureInvalidTableScope, reason: "duplicate_table_id", details: failureDetails{TableIDs: tables, LinkGraph: true}}
		}
		seen[id] = true
	}
	var timeRange map[string]json.RawMessage
	_ = json.Unmarshal(object["time_range"], &timeRange)
	if apiErr = linkMembers(timeRange, map[string]string{"start_utc": "nullable:string", "end_utc": "nullable:string"}); apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	var aggregation map[string]json.RawMessage
	_ = json.Unmarshal(object["aggregation"], &aggregation)
	members := map[string]string{"mode": "string", "include_example_row_refs": "boolean"}
	if linkString(aggregation, "mode") == "time_bucket_v1" {
		members["bucket_width_seconds"] = "number"
	}
	if apiErr = linkMembers(aggregation, members); apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	var filters []json.RawMessage
	_ = json.Unmarshal(object["filters"], &filters)
	if len(filters) > 16 {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("filters", "type_mismatch")
	}
	for _, filter := range filters {
		entry, err := linkObject(filter, "filters")
		if err != nil {
			return graphSemanticRequest{}, err
		}
		members := map[string]string{"field_key": "string", "op": "string"}
		op := linkString(entry, "op")
		if op != "is_null" && op != "not_null" {
			members["value"] = "any"
		}
		if err = linkMembers(entry, members); err != nil {
			return graphSemanticRequest{}, err
		}
	}
	admissionLimits := limits
	admissionLimits.MaxSelectedTablesPerQuery = 64
	admissionLimits.MaxFiltersPerQuery = 16
	query, apiErr := decodeGraphSemanticRequest(raw, admissionLimits)
	if apiErr != nil && apiErr.kind == failureInvalidFilter {
		// Preserve filter input ordering and safe structural context, never values.
		for index := range filters {
			prefix, _ := json.Marshal(filters[:index+1])
			if _, err := decodeFilters(prefix, admissionLimits); err != nil {
				var entry map[string]json.RawMessage
				_ = json.Unmarshal(filters[index], &entry)
				apiErr.details.FieldKey = optionalStringPtr(linkString(entry, "field_key"))
				apiErr.details.Op = optionalStringPtr(linkString(entry, "op"))
				apiErr.details.FilterIndex = &index
				if apiErr.reason == "value_forbidden" {
					apiErr.reason = "invalid_value"
				}
				break
			}
		}
	}
	if apiErr != nil {
		return graphSemanticRequest{}, completeLinkGraphError(apiErr, tables)
	}
	// Derived execution limits are current facts, excluded from semantic identity.
	query.ResultLimits = effectiveGraphResultLimits(limits)
	return query, nil
}

func completeLinkGraphError(f *semanticFailure, tables []string) *semanticFailure {
	f.details.LinkGraph = true
	if f.details.TableIDs == nil {
		f.details.TableIDs = append([]string{}, tables...)
	}
	if f.kind == failureInvalidRequest && f.reason == "invalid_timestamp" {
		f.reason = "type_mismatch"
	}
	return f
}
