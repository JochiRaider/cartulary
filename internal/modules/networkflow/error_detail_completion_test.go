package networkflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// routeErrorRequiredDetails is NF NLSpec Table 21-B as test expectations: the
// details members a route-local error must carry, JSON null when unknown
// (NF-REQ-175). Every route-local error also carries reason_code and
// retry_action.
var routeErrorRequiredDetails = map[string][]string{
	"network_flow_invalid_request":                      {"field", "expected_contract", "actual_kind"},
	"network_flow_table_name_exhausted":                 {"base_display_name_digest", "base_display_name_digest_key_id", "attempted_suffix_min", "attempted_suffix_max"},
	"network_flow_table_not_found":                      {"network_flow_table_id", "table_status", "allowed_states"},
	"network_flow_table_not_active":                     {"network_flow_table_id", "table_status", "allowed_states"},
	"network_flow_table_version_conflict":               {"network_flow_table_id", "base_table_version", "current_table_version"},
	"network_flow_invalid_display_name":                 {"field", "max_length", "normalized_length"},
	"network_flow_invalid_table_scope":                  {"mode", "table_ids", "limit_key"},
	"network_flow_invalid_filter":                       {"field_key", "op", "filter_index"},
	"network_flow_invalid_sort":                         {"field_key", "direction", "sort_index"},
	"network_flow_invalid_limit":                        {"limit_key", "limit", "minimum", "maximum"},
	"network_flow_invalid_limit_override":               {"limit_key", "limit", "minimum", "maximum"},
	"network_flow_cursor_invalid":                       {"cursor_scope"},
	"network_flow_invalid_time_range":                   {"field", "start_utc", "end_utc"},
	"network_flow_graph_projection_failed":              {"projection_contract_version"},
	"network_flow_graph_query_stale":                    {"graph_query_digest"},
	"network_flow_graph_limit_exceeded":                 {"limit_key", "limit", "actual", "phase"},
	"network_flow_counter_sum_limit_exceeded":           {"limit_key", "limit", "actual", "phase"},
	"network_flow_resource_limit_exceeded":              {"limit_key", "limit", "actual", "phase"},
	"network_flow_indicator_link_ambiguous":             {"selector_kind", "field_key", "target_mode", "resolved_candidate_value"},
	"network_flow_invalid_indicator_selector":           {"selector_kind", "field_key", "target_mode", "resolved_candidate_value"},
	"network_flow_invalid_indicator_target":             {"selector_kind", "field_key", "target_mode", "resolved_candidate_value"},
	"network_flow_indicator_link_forbidden":             {"selector_kind", "field_key", "target_mode", "resolved_candidate_value"},
	"network_flow_graph_view_not_found":                 {},
	"network_flow_graph_view_not_materialized":          {},
	"network_flow_graph_view_version_conflict":          {},
	"network_flow_graph_view_limit_exceeded":            {},
	"network_flow_graph_materialization_failed":         {},
	"network_flow_graph_materialization_limit_exceeded": {},
	"network_flow_invalid_graph_aggregation":            {},
}

// assertRouteErrorDetailsComplete covers NF-REQ-175 at the route's single
// error write: every route-local error the routes emit carries retry_action,
// by default its Table 21-A registry value, and every Table 21-B member, while
// Core-registered codes pass through unchanged and keep their retryable flag.
func assertRouteErrorDetailsComplete(t *testing.T) {
	t.Helper()
	var registry struct {
		Errors []struct {
			Code        string `json:"code"`
			Scope       string `json:"scope"`
			RetryAction string `json:"retry_action"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(contractnetworkflow.Index["contracts/network-flow/errors.v1.json"].JSON), &registry); err != nil {
		t.Fatal(err)
	}
	registryActions := map[string]string{}
	for _, entry := range registry.Errors {
		registryActions[entry.Code] = entry.RetryAction
	}
	write := func(t *testing.T, apiErr *httpapi.APIError) map[string]any {
		t.Helper()
		recorder := httptest.NewRecorder()
		writeAPIError(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/incidents/x/network-flow/tables", nil), apiErr)
		var envelope struct {
			Error map[string]any `json:"error"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Error
	}
	versionConflict := &graphViewVersionConflictError{Current: 3, Base: 2}
	emissions := []*httpapi.APIError{
		semanticHTTPError(invalidNetworkFlowRequest("query", "unknown_member")),
		semanticHTTPError(invalidNetworkFlowRequest("body", "variant_member_conflict")),
		semanticHTTPError(invalidSort("sort", "unknown_field")),
		semanticHTTPError(invalidLimit("limit", "above_maximum")),
		semanticHTTPError(invalidLimitOverride("limits", "unknown_limit_key", "", 0, 0, 0)),
		semanticHTTPError(invalidFilter("value", "duplicate_in_value")),
		semanticHTTPError(cursorInvalid("expired")),
		semanticHTTPError(invalidTableScope(tableScope{Mode: "selected_tables"}, "unknown_mode")),
		semanticHTTPError(invalidTimeRange("time_range", "empty_range", nil, nil)),
		semanticHTTPError(invalidGraphAggregation("aggregation", "unknown_mode")),
		semanticHTTPError(graphQueryStale("digest_mismatch", "digest")),
		semanticHTTPError(graphLimitExceeded("vertex_limit_exceeded", "max_vertices", 10, 11)),
		semanticHTTPError(&semanticFailure{kind: failureGraphProjectionFailed, reason: "adapter_contract_rejected"}),
		semanticHTTPError(&semanticFailure{kind: failureGraphMaterializationFailed, reason: "timeout"}),
		semanticHTTPError(newSemanticFailure(failureGraphViewNotFound, "graph_view_id", "not_found")),
		semanticHTTPError(newSemanticFailure(failureGraphViewNotMaterialized, "graph_view_id", "unmaterialized")),
		semanticHTTPError(tableReadFailure(errTableNotFound, "nft_"+fixtureHex32())),
		semanticHTTPError(tableReadFailure(errTableNotActive, "nft_"+fixtureHex32())),
		semanticHTTPError(invalidIndicatorSelector("selector", "unknown_selector_kind")),
		semanticHTTPError(indicatorLinkForbidden("row_field_value", "network_flow.endpoint_a", "existing_indicator", "target_not_visible")),
		graphViewMutationError(errGraphViewDeclarationNotFound),
		graphViewMutationError(errGraphViewDeclarationNotActive),
		graphViewMutationError(versionConflict),
		graphViewMutationError(errGraphViewActiveLimit),
		graphViewMutationError(errGraphViewRetainedLimit),
		graphViewMutationError(errGraphViewJobLimit),
		graphViewMutationError(&invalidDisplayNameError{ReasonCode: "display_name_too_long", NormalizedLength: 65}),
		tableMutationError(errTableNameExhausted, "nft_"+fixtureHex32()),
		tableMutationError(&invalidDisplayNameError{ReasonCode: "empty_display_name"}, "nft_"+fixtureHex32()),
		tableMutationError(&tableVersionConflictError{TableID: "nft_" + fixtureHex32(), BaseTableVersion: 1, CurrentTableVersion: 2}, "nft_"+fixtureHex32()),
	}
	retryActions := map[string]bool{"correct_request": true, "refresh_resource": true, "restart_query": true, "reduce_scope_or_limits": true, "retry_with_backoff": true, "do_not_retry": true}
	for _, apiErr := range emissions {
		t.Run(apiErr.Code+"/"+apiErr.Details["reason_code"].(string), func(t *testing.T) {
			payload := write(t, apiErr)
			code, _ := payload["code"].(string)
			details, _ := payload["details"].(map[string]any)
			want, routeLocal := registryActions[code]
			required, listed := routeErrorRequiredDetails[code]
			if !routeLocal || !listed {
				t.Fatalf("%s is not a listed route-local error", code)
			}
			if action, _ := details["retry_action"].(string); !retryActions[action] || action != want {
				t.Fatalf("%s retry_action = %v, want registry %q", code, details["retry_action"], want)
			}
			for _, member := range append([]string{"reason_code"}, required...) {
				if _, present := details[member]; !present {
					t.Fatalf("%s details %v lack Table 21-B member %q", code, details, member)
				}
			}
		})
	}
	t.Run("an explicit retry action is kept", func(t *testing.T) {
		details := write(t, semanticHTTPError(invalidTableScope(tableScope{Mode: "selected_tables"}, "selected_table_limit_exceeded")))["details"].(map[string]any)
		if details["retry_action"] != "reduce_scope_or_limits" {
			t.Fatalf("selected-table limit retry_action = %v", details["retry_action"])
		}
	})
	t.Run("the graph-view display name carries its bounds", func(t *testing.T) {
		details := write(t, graphViewMutationError(&invalidDisplayNameError{ReasonCode: "display_name_too_long", NormalizedLength: 65}))["details"].(map[string]any)
		if details["max_length"] != float64(64) || details["normalized_length"] != float64(65) {
			t.Fatalf("graph-view display name details = %v", details)
		}
	})
	t.Run("Core-registered codes pass through and stay retryable", func(t *testing.T) {
		for _, apiErr := range []*httpapi.APIError{
			semanticHTTPError(&semanticFailure{kind: failureTransactionConflict, reason: "deadlock_detected"}),
			semanticHTTPError(&semanticFailure{kind: failureTransactionTimeout, reason: "extension_transaction_timeout", details: failureDetails{OperationID: "network-flow-indicator-link:x", TimeoutSeconds: 30}}),
		} {
			payload := write(t, apiErr)
			details, _ := payload["details"].(map[string]any)
			if _, present := details["retry_action"]; present || payload["retryable"] != true {
				t.Fatalf("%s payload = %v", apiErr.Code, payload)
			}
		}
	})
}

func fixtureHex32() string {
	return "0123456789abcdef0123456789abcdef"
}
