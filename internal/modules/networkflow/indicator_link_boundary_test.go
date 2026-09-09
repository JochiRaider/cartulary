package networkflow

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndicatorLinkPublicBoundary(t *testing.T) {
	t.Run("nested graph admission and required error context", func(t *testing.T) {
		query := map[string]any{"schema_id": schemaGraphSemanticQueryV2, "selected_table_ids": []string{"nft_" + strings.Repeat("a", 32)}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}}
		limits := DefaultEffectiveLimits()
		limits.MaxSelectedTablesPerQuery = 0
		raw, _ := json.Marshal(query)
		if _, err := decodeLinkGraphQuery(raw, limits); err != nil {
			t.Fatal("current limits must follow exact replay", err)
		}
		query["filters"] = []any{map[string]any{"field_key": "network_flow.not_authorized", "op": "eq", "value": "protected"}}
		raw, _ = json.Marshal(query)
		_, apiErr := decodeLinkGraphQuery(raw, DefaultEffectiveLimits())
		if apiErr == nil || apiErr.Code != "network_flow_invalid_filter" || apiErr.Details["filter_index"] != 0 || apiErr.Details["op"] != "eq" || apiErr.Details["retry_action"] == nil {
			t.Fatalf("missing filter context: %v", apiErr)
		}
		query["filters"] = []any{map[string]any{"field_key": FieldSrcIP, "op": "eq", "value": "192.0.2.1", "extra": true}}
		raw, _ = json.Marshal(query)
		_, apiErr = decodeLinkGraphQuery(raw, DefaultEffectiveLimits())
		if apiErr == nil || apiErr.Code != "network_flow_invalid_request" || apiErr.Details["reason_code"] != "unknown_member" {
			t.Fatalf("nested unknown member: %v", apiErr)
		}
		apiErr = linkMembers(map[string]json.RawMessage{"a": json.RawMessage(`null`)}, map[string]string{"a": "string", "z": "string"})
		if apiErr == nil || apiErr.Details["reason_code"] != "missing_member" || apiErr.Details["field"] != "z" {
			t.Fatal("missing members must precede explicit null")
		}
	})
	t.Run("result identifier", func(t *testing.T) {
		if schemaIndicatorLinkResult != "cartulary.network_flow_indicator_link_result.v1" {
			t.Fatal("indicator-link result identifier does not match the public boundary")
		}
	})
	t.Run("selector variants and canonical admission order", func(t *testing.T) {
		for _, id := range []string{"urn:uuid:a3333333-3333-4333-8333-333333333333", "a3333333333343338333333333333333", "{a3333333-3333-4333-8333-333333333333}"} {
			raw, _ := json.Marshal(map[string]any{"mode": "existing_indicator", "indicator_id": id})
			if _, apiErr := decodeIndicatorTarget(raw); apiErr == nil {
				t.Fatal("non-UUID wire spelling accepted")
			}
		}
		query := map[string]any{"schema_id": schemaGraphSemanticQueryV2, "selected_table_ids": []string{"nft_" + strings.Repeat("a", 32)}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}}
		for _, kind := range []string{"graph_vertex", "graph_edge"} {
			selector := map[string]any{"kind": kind, "graph_query": query, "graph_query_digest": strings.Repeat("b", 64)}
			if kind == "graph_vertex" {
				selector["vertex_id"] = "nfe_" + strings.Repeat("c", 64)
			} else {
				selector["edge_id"] = "nff_" + strings.Repeat("c", 64)
				selector["field_key"] = FieldDstIP
			}
			raw, _ := json.Marshal(selector)
			if _, apiErr := decodeIndicatorSelector(raw, DefaultEffectiveLimits()); apiErr != nil {
				t.Fatalf("authorized %s rejected: %v", kind, apiErr)
			}
			if kind == "graph_edge" {
				selector["edge_id"] = "nfbe_" + strings.Repeat("c", 64)
				raw, _ = json.Marshal(selector)
				if _, apiErr := decodeIndicatorSelector(raw, DefaultEffectiveLimits()); apiErr == nil {
					t.Fatal("bucket edge admitted for linking")
				}
			}
		}
		for _, body := range []string{`{"z":true,"a":true}`, `{"schema_id":null}`, `{"schema_id":"first","schema_id":"second"}`} {
			_, apiErr := decodeIndicatorLinkRequest(httptest.NewRequest("POST", "/", strings.NewReader(body)), DefaultEffectiveLimits())
			if apiErr == nil {
				t.Fatal("malformed request admitted")
			}
			for _, key := range []string{"field", "reason_code", "expected_contract", "actual_kind", "retry_action"} {
				if _, exists := apiErr.Details[key]; !exists {
					t.Fatalf("required admission detail %s absent", key)
				}
			}
			if strings.Contains(body, `"z"`) && apiErr.Details["field"] != "a" {
				t.Fatal("unknown members not checked in canonical order")
			}
		}
	})
	t.Run("domain separated immutable intent", func(t *testing.T) {
		request := indicatorLinkRequest{ClientTxnID: "first", Selector: indicatorLinkSelector{
			Kind: "row_field_value", TableID: "nft_" + strings.Repeat("a", 32),
			RowID: "nfr_" + strings.Repeat("b", 64), FieldKey: FieldSrcIP,
		}, Target: indicatorLinkTarget{Mode: "create_indicator", IndicatorType: "ipv4_addr"}, ConfirmExactValue: "192.0.2.1"}
		body := canonicalJSON(map[string]any{"selector": indicatorSelectorHashResource(request.Selector),
			"target": indicatorTargetHashResource(request.Target), "observation_mode": "binding_only", "confirm_exact_value": request.ConfirmExactValue})
		transcript := append([]byte("cartulary.network_flow.mutation_request_digest.v1\x00nf.indicator_links.create\x00indicator-links\x00"), body...)
		want := sha256.Sum256(append(transcript, 0))
		if !bytes.Equal(indicatorLinkRequestHash(request), want[:]) {
			t.Fatal("link digest does not bind the prescribed domain, route, path and comparison body")
		}
		request.ClientTxnID = "second"
		if !bytes.Equal(indicatorLinkRequestHash(request), want[:]) {
			t.Fatal("transaction identity leaked into the comparison body")
		}
		request.ConfirmExactValue += " "
		if bytes.Equal(indicatorLinkRequestHash(request), want[:]) {
			t.Fatal("changed confirmation must be a different intent")
		}
	})
	t.Run("closed source references", func(t *testing.T) {
		selector := map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": []any{map[string]any{
			"network_flow_table_id": "nft_" + strings.Repeat("a", 32), "network_flow_row_id": "nfr_" + strings.Repeat("b", 64),
			"source_row_number": 1, "mapping_fingerprint": strings.Repeat("c", 64), "label": "forbidden",
		}}}
		raw, _ := json.Marshal(selector)
		if _, apiErr := decodeIndicatorSelector(raw, DefaultEffectiveLimits()); apiErr == nil {
			t.Fatal("unknown source-reference members were accepted")
		}
	})
}
