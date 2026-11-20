package networkflow

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// routeErrorRetryActions maps each route-local Network Flow error code to its
// retry action in the packaged NF error registry (NF NLSpec Tables 21-A and
// 21-B1). Core-registered codes are not route-local and are absent.
var routeErrorRetryActions = func() map[string]string {
	var registry struct {
		Errors []struct {
			Code        string `json:"code"`
			Scope       string `json:"scope"`
			RetryAction string `json:"retry_action"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(contractnetworkflow.Index["contracts/network-flow/errors.v1.json"].JSON), &registry); err != nil {
		panic(err)
	}
	actions := make(map[string]string, len(registry.Errors))
	for _, entry := range registry.Errors {
		if strings.Contains(entry.Scope, "route") && entry.RetryAction != "" {
			actions[entry.Code] = entry.RetryAction
		}
	}
	return actions
}()

// routeErrorDetailMembers is NF NLSpec Table 21-B: the details members, beyond
// reason_code and retry_action, that a route-local error carries. A member the
// emitting path does not know is JSON null (NF-REQ-175).
var routeErrorDetailMembers = map[string][]string{
	"network_flow_invalid_request":               {"field", "expected_contract", "actual_kind"},
	"network_flow_table_name_exhausted":          {"base_display_name_digest", "base_display_name_digest_key_id", "attempted_suffix_min", "attempted_suffix_max"},
	"network_flow_table_not_found":               {"network_flow_table_id", "table_status", "allowed_states"},
	"network_flow_table_not_active":              {"network_flow_table_id", "table_status", "allowed_states"},
	"network_flow_table_version_conflict":        {"network_flow_table_id", "base_table_version", "current_table_version"},
	"network_flow_invalid_display_name":          {"field", "max_length", "normalized_length"},
	"network_flow_invalid_table_scope":           {"mode", "table_ids", "limit_key"},
	"network_flow_invalid_filter":                {"field_key", "op", "filter_index"},
	"network_flow_invalid_sort":                  {"field_key", "direction", "sort_index"},
	"network_flow_invalid_limit":                 {"limit_key", "limit", "minimum", "maximum"},
	"network_flow_invalid_limit_override":        {"limit_key", "limit", "minimum", "maximum"},
	"network_flow_cursor_invalid":                {"cursor_scope"},
	"network_flow_invalid_time_range":            {"field", "start_utc", "end_utc"},
	"network_flow_graph_projection_failed":       {"projection_contract_version"},
	"network_flow_graph_query_stale":             {"graph_query_digest"},
	"network_flow_table_limit_exceeded":          {"limit_key", "limit", "actual", "phase"},
	"network_flow_resource_limit_exceeded":       {"limit_key", "limit", "actual", "phase"},
	"network_flow_graph_limit_exceeded":          {"limit_key", "limit", "actual", "phase"},
	"network_flow_counter_sum_limit_exceeded":    {"limit_key", "limit", "actual", "phase"},
	"network_flow_indicator_link_ambiguous":      {"selector_kind", "field_key", "target_mode"},
	"network_flow_invalid_indicator_selector":    {"selector_kind", "field_key", "target_mode"},
	"network_flow_invalid_indicator_target":      {"selector_kind", "field_key", "target_mode"},
	"network_flow_indicator_link_forbidden":      {"selector_kind", "field_key", "target_mode"},
	"network_flow_id_generation_failed":          {"attempt_count"},
	"network_flow_external_enrichment_forbidden": {"requested_capability"},
}

// completeRouteError is the single completion of every Network Flow route
// error (NF-REQ-175): a route-local error gains its registry retry_action when
// the emitting path chose none, and every Table 21-B member it lacks as JSON
// null. An indicator-link error on this incident-authorized route that carries
// no candidate value or safe digest carries resolved_candidate_value null.
// Core-registered codes pass unchanged.
func completeRouteError(apiErr *httpapi.APIError) *httpapi.APIError {
	if apiErr == nil {
		return nil
	}
	action, routeLocal := routeErrorRetryActions[apiErr.Code]
	if !routeLocal {
		return apiErr
	}
	completed := *apiErr
	completed.Details = make(map[string]any, len(apiErr.Details)+4)
	for member, value := range apiErr.Details {
		completed.Details[member] = value
	}
	if _, chosen := completed.Details["retry_action"]; !chosen {
		completed.Details["retry_action"] = action
	}
	for _, member := range routeErrorDetailMembers[apiErr.Code] {
		if _, present := completed.Details[member]; !present {
			completed.Details[member] = nil
		}
	}
	if _, linkError := routeErrorDetailMembers[apiErr.Code]; linkError && strings.Contains(apiErr.Code, "indicator") {
		_, value := completed.Details["resolved_candidate_value"]
		_, digest := completed.Details["resolved_candidate_safe_digest"]
		if !value && !digest {
			completed.Details["resolved_candidate_value"] = nil
		}
	}
	return &completed
}

// pathTableNotFound reports a not-found table the request path addresses with
// network_flow_table_id null, so a well-formed missing identifier receives the
// malformed identifier's response (NF NLSpec Table 21-B; Core 01 REQ-01-707
// step 3). Every other error passes unchanged.
func pathTableNotFound(apiErr *httpapi.APIError) *httpapi.APIError {
	if apiErr == nil || apiErr.Code != string(failureTableNotFound) {
		return apiErr
	}
	addressed := *apiErr
	addressed.Details = make(map[string]any, len(apiErr.Details))
	for member, value := range apiErr.Details {
		addressed.Details[member] = value
	}
	addressed.Details["network_flow_table_id"] = nil
	return &addressed
}

// writeAPIError writes every Network Flow route error through
// completeRouteError, keeping its retryable flag (Extensions EXT-REQ-219).
func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	apiErr = completeRouteError(apiErr)
	if apiErr == nil {
		httpapi.WriteAPIError(w, r, nil)
		return
	}
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithOptions(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, httpapi.ErrorOptions{Retryable: apiErr.Retryable})
}
