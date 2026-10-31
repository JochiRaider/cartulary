package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"testing"
)

func TestQueryAuthoringAdmission_Unit(t *testing.T) {
	t.Run("shared paginated limit admission", assertPaginatedLimitAdmission)
	t.Run("safe table context and continuation failures", assertTableScopeFailurePolicy)
	t.Run("graph selection and contributor cursor scope failures", assertGraphContributorScopeFailurePolicy)
	t.Run("shared scope order and admission", assertSharedTableScope)
	t.Run("direct table admission", assertDirectTableReadAdmission)
	t.Run("production facade", assertProductionSurface)
	t.Run("semantic dependency and error boundary", assertSemanticBoundary)
	for _, body := range []string{
		`{"field_keys":["unknown"]}`, `{"error_codes":["network_flow_unknown"]}`,
		`{"field_keys":["network_flow.src_ip","network_flow.src_ip"]}`,
		`{"source_row_range":{}}`, `{"source_row_range":{"gte":null,"lte":null}}`,
		`{"source_row_range":{"gte":3,"lte":2}}`, `{"source_row_range":{"gte":1.5}}`,
		`{"source_row_range":null}`, `{"field_keys":null}`,
	} {
		request := `{"schema_id":"cartulary.network_flow.rejected_rows_query_request.v1",` + body[1:]
		_, err := decodeRejectedRowsQueryRequest(strings.NewReader(request), defaultLimits())
		if err == nil || string(err.kind) != "network_flow_invalid_filter" {
			t.Fatalf("%s: %v", request, err)
		}
	}
	for _, body := range []string{
		`[{"field_key":"network_flow.ip_protocol","op":"in","value":[6,17]}]`,
		`[{"field_key":"network_flow.bytes_count","op":"eq","value":"18446744073709551615"}]`,
		`[{"field_key":"network_flow.bytes_count","op":"range","value":{"gte":null,"lte":"18446744073709551615"}}]`,
		`[{"field_key":"network_flow.bytes_count","op":"range","value":{"gte":"0","lte":"18446744073709551615"}}]`,
		`[{"field_key":"network_flow.input_interface","op":"in","value":["01","1","a,b"]}]`,
		`[{"field_key":"network_flow.src_port","op":"is_null"}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"range","value":{"gte":"2026-07-10T12:00:00.000001Z","lt":"2026-07-10T12:00:00.000002Z"}}]`,
	} {
		if _, err := decodeFilters(json.RawMessage(body), defaultLimits()); err != nil {
			t.Fatalf("%s: %v", body, err)
		}
	}
	for _, body := range []string{
		`[{"field_key":"network_flow.ip_protocol","op":"in","value":[6,6]}]`,
		`[{"field_key":"network_flow.bytes_count","op":"eq","value":"18446744073709551616"}]`,
		`[{"field_key":"network_flow.src_ip","op":"in","value":["2001:db8::1","2001:0db8:0:0:0:0:0:1"]}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"eq","value":"2026-07-10T12:00:00Z"}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"range","value":{"gte":"2026-07-10T12:00:00Z","lt":"2026-07-10T12:00:00Z"}}]`,
		`[{"field_key":"network_flow.src_port","op":"is_null","value":null}]`,
	} {
		_, err := decodeFilters(json.RawMessage(body), defaultLimits())
		if err == nil || string(err.kind) != "network_flow_invalid_filter" || (err.details.FilterIndex == nil || *err.details.FilterIndex != 0) || err.details.FieldKey == nil || err.details.Op == nil {
			t.Fatalf("%s: %v", body, err)
		}
	}
	if compareFilterValues(fieldInputInterface, "01", "1") == 0 {
		t.Fatal("text was compared as an integer")
	}
	request, err := decodeRejectedRowsQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.rejected_rows_query_request.v1","field_keys":["network_flow.src_ip","network_flow.dst_ip"],"source_row_range":{"gte":1,"lte":1}}`), defaultLimits())
	if err != nil || len(request.FieldKeys) != 2 {
		t.Fatalf("diagnostic round trip: %+v %v", request, err)
	}
}

func assertPaginatedLimitAdmission(t *testing.T) {
	selector := `{"kind":"vertex","source_vertex_id":"nfe_` + strings.Repeat("a", 64) + `","endpoint_value":"192.0.2.10"}`
	semantic := `{"schema_id":"cartulary.network_flow.graph_semantic_query.v2","selected_table_ids":["nft_00000000000040008000000000000005"],"filters":[],"time_range":{"start_utc":null,"end_utc":null},"aggregation":{"mode":"default_flow_edge_v1","include_example_row_refs":true}}`
	for _, route := range []struct {
		name, body string
		decode     func(string, EffectiveLimits) (int, *semanticFailure)
	}{
		{"table", `{"schema_id":"cartulary.network_flow.table_query_request.v1"}`, func(body string, limits EffectiveLimits) (int, *semanticFailure) {
			r, f := decodeAcceptedRowQueryRequest(strings.NewReader(body), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
			return r.Limit, f
		}},
		{"rows", `{"schema_id":"cartulary.network_flow.rows_query_request.v1","table_scope":{"mode":"all_active_tables"}}`, func(body string, limits EffectiveLimits) (int, *semanticFailure) {
			r, f := decodeAcceptedRowQueryRequest(strings.NewReader(body), schemaRowsQueryRequest, schemaRowsQueryContinuation, limits)
			return r.Limit, f
		}},
		{"diagnostics", `{"schema_id":"cartulary.network_flow.rejected_rows_query_request.v1"}`, func(body string, limits EffectiveLimits) (int, *semanticFailure) {
			r, f := decodeRejectedRowsQueryRequest(strings.NewReader(body), limits)
			return r.Limit, f
		}},
		{"ephemeral contributors", `{"schema_id":"cartulary.network_flow.graph_contributor_query_request.v2","graph_query":` + semantic + `,"graph_query_digest":"` + strings.Repeat("b", 64) + `","selector":` + selector + `}`, func(body string, limits EffectiveLimits) (int, *semanticFailure) {
			r, f := decodeGraphContributorQueryRequestValue(strings.NewReader(body), limits)
			return r.Limit, f
		}},
		{"saved contributors", `{"schema_id":"cartulary.network_flow.graph_view_contributor_query_request.v2","projection_result_id":"gpres_` + strings.Repeat("c", 64) + `","selector":` + selector + `}`, func(body string, limits EffectiveLimits) (int, *semanticFailure) {
			r, f := decodeSavedGraphContributorRequestValue(strings.NewReader(body), limits)
			return r.Limit, f
		}},
	} {
		t.Run(route.name, func(t *testing.T) {
			for _, c := range []struct {
				name, raw string
				max       int64
				want      int
				reason    string
			}{
				{"bounded default", "", 32, 32, ""}, {"ordinary default", "", 500, 200, ""}, {"maximum deployment default", "", 1000, 200, ""},
				{"minimum", "1", 32, 1, ""}, {"deployment maximum", "32", 32, 32, ""},
				{"zero", "0", 32, 0, "below_minimum"}, {"negative", "-1", 32, 0, "below_minimum"}, {"above deployment", "33", 32, 0, "above_maximum"},
				{"null", "null", 32, 0, "not_integer"}, {"fraction", "1.5", 32, 0, "not_integer"}, {"string", `"1"`, 32, 0, "not_integer"}, {"boolean", "true", 32, 0, "not_integer"},
			} {
				t.Run(c.name, func(t *testing.T) {
					body := route.body
					if c.raw != "" {
						body = strings.TrimSuffix(body, "}") + `,"limit":` + c.raw + `}`
					}
					limits := defaultLimits()
					limits.MaxQueryLimit = c.max
					limit, failure := route.decode(body, limits)
					if c.reason == "" {
						if failure != nil || limit != c.want {
							t.Fatalf("limit=%d failure=%v; want %d", limit, failure, c.want)
						}
					} else if failure == nil || failure.kind != failureInvalidLimit || failure.reason != c.reason {
						t.Fatalf("failure=%v; want invalid limit/%s", failure, c.reason)
					}
				})
			}
		})
	}
}

func assertDirectTableReadAdmission(t *testing.T) {
	access := &authorizationAccess{err: &admission.Denied{Code: admission.DenialNotVisible}}
	app := &tableQueryApplication{incidentAccess: access}
	who := readIdentity{ActorID: uuid.New(), IncidentID: uuid.New(), SessionID: uuid.New()}
	_, a := app.profiles(context.Background(), who)
	_, b := app.list(context.Background(), who)
	_, c := app.get(context.Background(), who, "invalid")
	_, d := app.rows(context.Background(), who, "invalid", rowQueryRequest{Continuation: true, CursorToken: "invalid"})
	_, e := app.diagnostics(context.Background(), who, "invalid", rejectedRowsQueryRequest{})
	for _, f := range []*semanticFailure{a, b, c, d, e} {
		if f == nil || f.kind != failureAdmission {
			t.Fatalf("direct read bypassed current authority: %v", f)
		}
	}
}

func assertSharedTableScope(t *testing.T) {
	active := []tableRecord{{TableID: "nft_ffffffffffffffffffffffffffffffff"}, {TableID: "nft_00000000000000000000000000000001"}, {TableID: "nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	limits := defaultLimits()
	limits.MaxSelectedTablesPerQuery = 2
	selected, missing, failure := selectActiveScopeTables(active, tableScope{Mode: "selected_tables", SelectedTableIDs: []string{active[2].TableID, active[0].TableID}}, limits)
	if failure != nil || len(missing) != 0 || !reflect.DeepEqual(selected, []tableRecord{active[0], active[2]}) {
		t.Fatalf("workspace order was lost: %v %v %v", selected, missing, failure)
	}
	_, _, failure = selectActiveScopeTables(active, tableScope{Mode: "all_active_tables"}, limits)
	if failure == nil || failure.reason != "selected_table_limit_exceeded" || !reflect.DeepEqual(failure.details.TableIDs, []string{active[0].TableID, active[1].TableID, active[2].TableID}) {
		t.Fatal("all-active scope bypassed the shared bound")
	}
	_, _, failure = selectActiveScopeTables(active, tableScope{Mode: "selected_tables", SelectedTableIDs: []string{active[0].TableID, active[0].TableID}}, limits)
	if failure == nil || failure.reason != "duplicate_table_id" {
		t.Fatal("explicit duplicate was removed")
	}
	_, missing, failure = selectActiveScopeTables(active, tableScope{Mode: "selected_tables", SelectedTableIDs: []string{active[0].TableID, "nft_00000000000000000000000000000002"}}, limits)
	if failure != nil || !reflect.DeepEqual(missing, []string{"nft_00000000000000000000000000000002"}) {
		t.Fatal("missing explicit selection was removed")
	}
}

func assertTableScopeFailurePolicy(t *testing.T) {
	tableID := "nft_00000000000040008000000000000005"
	for _, err := range []error{errTableNotFound, errTableNotActive} {
		failure := tableReadFailure(err, tableID)
		for _, continuation := range []bool{false, true} {
			presented := semanticHTTPError(queryScopeFailure(failure, continuation))
			if continuation {
				if presented.Code != "network_flow_cursor_invalid" || !reflect.DeepEqual(presented.Details, map[string]any{"reason_code": "scope_stale", "cursor_scope": nil, "retry_action": "restart_query"}) {
					t.Fatal(presented)
				}
			} else if presented.Details["network_flow_table_id"] != tableID || presented.Details["retry_action"] != "refresh_resource" {
				t.Fatal(presented)
			}
		}
	}
	malformed := semanticHTTPError(tableReadFailure(errTableNotFound, "private-path-like-input"))
	if id, present := malformed.Details["network_flow_table_id"]; !present || id != nil {
		t.Fatal("malformed table identity escaped safe boundary")
	}
	internal := internalSemanticFailure(context.Canceled)
	if queryScopeFailure(internal, true) != internal {
		t.Fatal("internal failure was relabeled as stale cursor")
	}
}

func assertGraphContributorScopeFailurePolicy(t *testing.T) {
	digest := strings.Repeat("a", 64)
	for _, cause := range []error{errTableNotFound, errTableNotActive} {
		failure := tableReadFailure(cause, "nft_00000000000040008000000000000005")
		for _, continuation := range []bool{false, true} {
			presented := semanticHTTPError(graphContributorScopeFailure(failure, continuation, digest))
			code, status := "network_flow_graph_query_stale", 409
			details := map[string]any{"reason_code": "scope_stale", "graph_query_digest": digest, "retry_action": "refresh_resource"}
			if continuation {
				code, status = "network_flow_cursor_invalid", 400
				details = map[string]any{"reason_code": "scope_stale", "cursor_scope": nil, "retry_action": "restart_query"}
			}
			if presented.Code != code || presented.Status != status || !reflect.DeepEqual(presented.Details, details) {
				t.Fatal(presented)
			}
		}
	}
	for _, failure := range []*semanticFailure{nil, internalSemanticFailure(context.Canceled), graphProjectionFailed("projection_cancelled"), cursorInvalid("actor_mismatch"), graphQueryStale("digest_mismatch", digest)} {
		for _, continuation := range []bool{false, true} {
			if graphContributorScopeFailure(failure, continuation, digest) != failure {
				t.Fatal("unrelated contributor failure was reclassified")
			}
		}
	}
}
