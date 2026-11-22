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
	// A facade-only classification must never cross the route writer.
	for _, code := range []string{"network_flow_source_changed", "network_flow_mapping_conflict", "network_flow_csv_empty_file", "network_flow_invalid_utf8"} {
		got := completeRouteError(&httpapi.APIError{Status: 400, Code: code, Details: map[string]any{"filename": "private"}}, "")
		if got.Status != http.StatusInternalServerError || got.Code != "internal_error" || len(got.Details) != 0 {
			t.Fatalf("import-only classification %s escaped the route writer: %#v", code, got)
		}
	}
	var registry struct {
		Errors []struct {
			Code        string `json:"code"`
			RetryAction string `json:"retry_action"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(contractnetworkflow.Index["contracts/network-flow/errors.v2.json"].JSON), &registry); err != nil {
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
		semanticHTTPError(invalidSortAt("unknown_field", nil, nil, nil)),
		semanticHTTPError(invalidQueryLimit("above_maximum", nil, EffectiveLimits{MaxQueryLimit: 1000})),
		semanticHTTPError(invalidLimitOverride("limits", "unknown_limit_key", "", nil, nil, nil)),
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
	t.Run("a reason never changes its code's retry action", func(t *testing.T) {
		details := write(t, semanticHTTPError(invalidTableScope(tableScope{Mode: "selected_tables"}, "selected_table_limit_exceeded")))["details"].(map[string]any)
		if details["retry_action"] != registryActions["network_flow_invalid_table_scope"] {
			t.Fatalf("selected-table limit retry_action = %v, want the registry's %q", details["retry_action"], registryActions["network_flow_invalid_table_scope"])
		}
	})
	// NF-REQ-175: a member the emitting stage determined is never null.
	operationWrite := func(t *testing.T, operationID string, apiErr *httpapi.APIError) map[string]any {
		t.Helper()
		recorder := httptest.NewRecorder()
		writeOperationError(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/incidents/x/network-flow/rows/query", nil), apiErr, operationID)
		var envelope struct {
			Error struct {
				Details map[string]any `json:"details"`
			} `json:"error"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Error.Details
	}
	t.Run("determined members are non-null", func(t *testing.T) {
		index, field, direction := 1, "packets", "sideways"
		sort := write(t, semanticHTTPError(invalidSortAt("invalid_direction", &index, &field, &direction)))["details"].(map[string]any)
		if sort["sort_index"] != float64(1) || sort["field_key"] != "packets" || sort["direction"] != "sideways" {
			t.Fatalf("invalid sort details = %v", sort)
		}
		if _, present := sort["field"]; present {
			t.Fatalf("invalid sort carries a member outside Table 21-B: %v", sort)
		}
		zero := 0
		limit := write(t, semanticHTTPError(invalidQueryLimit("below_minimum", &zero, EffectiveLimits{MaxQueryLimit: 500})))["details"].(map[string]any)
		if limit["limit_key"] != "network_flow.max_query_limit" || limit["limit"] != float64(0) || limit["minimum"] != float64(1) || limit["maximum"] != float64(500) {
			t.Fatalf("invalid limit details = %v", limit)
		}
		notInteger := write(t, semanticHTTPError(invalidQueryLimit("not_integer", nil, EffectiveLimits{MaxQueryLimit: 500})))["details"].(map[string]any)
		if notInteger["limit"] != nil || notInteger["minimum"] != float64(1) {
			t.Fatalf("non-integer limit details = %v", notInteger)
		}
		unknown := write(t, semanticHTTPError(invalidLimitOverride("max_widgets", "unknown_limit_key", "", nil, nil, nil)))["details"].(map[string]any)
		for _, member := range []string{"limit_key", "limit", "minimum", "maximum"} {
			if value, present := unknown[member]; !present || value != nil {
				t.Fatalf("unknown override member %s = %v, want present null", member, value)
			}
		}
		_, overrideErr := decodeLowerableGraphLimit(json.RawMessage(`0`), "max_vertices", 1, 100)
		override := write(t, semanticHTTPError(overrideErr))["details"].(map[string]any)
		if override["limit_key"] != "network_flow.max_graph_vertices" || override["limit"] != float64(0) || override["minimum"] != float64(1) || override["maximum"] != float64(100) {
			t.Fatalf("override details = %v", override)
		}
		_, mismatch := decodePositiveInt(json.RawMessage(`"seven"`), "limit")
		request := operationWrite(t, "nf.rows.query", semanticHTTPError(mismatch))
		if request["actual_kind"] != "string" || request["expected_contract"] != "cartulary.network_flow.rows_query_request.v2" {
			t.Fatalf("invalid request details = %v", request)
		}
		if bodyless := operationWrite(t, "nf.tables.list", semanticHTTPError(invalidNetworkFlowRequest("query", "explicit_null"))); bodyless["expected_contract"] != nil || bodyless["actual_kind"] != "null" {
			t.Fatalf("bodyless invalid request details = %v", bodyless)
		}
		for operation, scope := range cursorScopes {
			if cursor := operationWrite(t, operation, semanticHTTPError(cursorInvalid("expired"))); cursor["cursor_scope"] != scope {
				t.Fatalf("%s cursor_scope = %v, want %s", operation, cursor["cursor_scope"], scope)
			}
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
