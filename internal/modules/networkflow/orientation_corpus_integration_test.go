package networkflow_test

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Exercise the incident corpus through the actual Imports/Jobs/SQL/query/graph
// boundaries. Expectations are the independently retained per-observation and
// per-connection JSON, never output captured from the application.
func testOrientationIncidentCorpusService(t *testing.T) {
	const root = "../../../fixtures/network-flow-orientation/v1"
	var manifest struct {
		Datasets []struct {
			Name, CSV, Expected string
			CSVHash             string `json:"csv_sha256"`
			ExpectedHash        string `json:"expected_sha256"`
		} `json:"datasets"`
	}
	read := func(path string) []byte {
		t.Helper()
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	if err := json.Unmarshal(read(filepath.Join(root, "manifest.json")), &manifest); err != nil {
		t.Fatal(err)
	}
	var cases []publicationFixtureCase
	if err := json.Unmarshal(read("../../../fixtures/network-flow/NF-FIX-029-cisco-sna-flow-search/source/cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	var runtime *publicationFixtureRuntime
	type total struct {
		Members        map[string]bool
		Count          int
		Bytes, Packets *big.Int
	}
	combined := map[string]*total{}
	hash := func(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
	key := func(kind string, a, b map[string]any, protocol any) string {
		raw, _ := json.Marshal([]any{kind, a, b, protocol})
		return string(raw)
	}
	for index, dataset := range manifest.Datasets {
		if dataset.Name == "example1200" {
			continue
		}
		source, expectedBytes := read(filepath.Join(root, dataset.CSV)), read(filepath.Join(root, dataset.Expected))
		if hash(source) != dataset.CSVHash || hash(expectedBytes) != dataset.ExpectedHash {
			t.Fatal("incident corpus bytes changed")
		}
		var expected struct{ Rows, Edges []map[string]any }
		if err := json.Unmarshal(expectedBytes, &expected); err != nil {
			t.Fatal(err)
		}
		input := cases[0]
		input.Name = dataset.Name
		input.Filename = dataset.CSV
		input.TxnPrefix = "corpus-" + dataset.Name
		input.QueryLimit = 200
		input.SessionID = fmt.Sprintf("10000000-0000-4000-8000-%012d", index+1)
		input.UnitID = fmt.Sprintf("20000000-0000-4000-8000-%012d", index+1)
		input.TableUUID = fmt.Sprintf("30000000-0000-4000-8000-%012d", index+1)
		input.ResourceLimits = map[string]int64{"max_graph_vertices": 1000, "max_graph_edges": 5000, "max_contributing_rows_per_graph": 5000}
		var mapping map[string]any
		if err := json.Unmarshal(input.Mapping, &mapping); err != nil {
			t.Fatal(err)
		}
		mapping["unknown_column_policy"] = "ignore_unmapped_columns"
		fields := mapping["field_mappings"].([]any)
		mapped := map[int]bool{}
		for _, raw := range fields {
			f := raw.(map[string]any)
			if f["mapping_kind"] == "source_column" {
				mapped[int(f["source_column_ordinal"].(float64))] = true
			}
		}
		header, err := csv.NewReader(strings.NewReader(string(source))).Read()
		if err != nil {
			t.Fatal(err)
		}
		for ordinal := range header {
			if !mapped[ordinal+1] {
				fields = append(fields, map[string]any{"mapping_kind": "ignored_source_column", "source_column_ordinal": ordinal + 1, "ignore_reason": "policy_accounting"})
			}
		}
		mapping["field_mappings"] = fields
		input.Mapping, err = json.Marshal(mapping)
		if err != nil {
			t.Fatal(err)
		}
		_, err = runPublicationBytes(t, nil, input, source, runtime, func(r *publicationFixtureRuntime, out map[string]any) error {
			runtime = r
			if out["job_status"] != "succeeded" {
				return fmt.Errorf("%s failed: %v", dataset.Name, out["job_error"])
			}
			byNumber := map[int]map[string]any{}
			for _, row := range expected.Rows {
				byNumber[int(row["source_row_number"].(float64))] = row
			}
			page := out["rows"].(map[string]any)
			seen := map[string]bool{}
			count := 0
			for {
				for _, raw := range page["rows"].([]any) {
					row := raw.(map[string]any)
					want := byNumber[int(row["source_row_number"].(float64))]
					for actualKey, wantKey := range map[string]string{"network_flow.endpoint_a": "endpoint_a", "network_flow.endpoint_b": "endpoint_b", "network_flow.direction": "direction", "network_flow.ip_protocol": "protocol", "network_flow.bytes_count": "bytes", "network_flow.packets_count": "packets"} {
						if !reflect.DeepEqual(row[actualKey], want[wantKey]) {
							return fmt.Errorf("%s row %v field %s differs", dataset.Name, row["source_row_number"], actualKey)
						}
					}
					id := row["network_flow_row_id"].(string)
					if seen[id] {
						return fmt.Errorf("repeated query row")
					}
					seen[id] = true
					count++
					evidence := row["network_flow.direction_evidence"].(map[string]any)
					if len(evidence["inputs"].([]any)) != 6 || len(row["unmapped_raw"].(map[string]any)) != 0 {
						return fmt.Errorf("required evidence depended on optional raw retention")
					}
				}
				cursor, _ := page["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
				if cursor == "" {
					break
				}
				page = r.request(http.MethodPost, "/api/v1/incidents/"+input.IncidentID+"/network-flow/tables/nft_"+strings.ReplaceAll(input.TableUUID, "-", "")+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v2", "cursor_token": cursor}, http.StatusOK)
			}
			if count != len(expected.Rows) {
				return fmt.Errorf("%s lost observations: %d", dataset.Name, count)
			}
			g := out["graph"].(map[string]any)
			projection := g["graph_projection_result"].(map[string]any)
			edges := projection["edges"].([]any)
			if len(edges) != len(expected.Edges) {
				return fmt.Errorf("%s edge count differs", dataset.Name)
			}
			expectedByKey := map[string]map[string]any{}
			for _, e := range expected.Edges {
				k := key(e["direction_kind"].(string), e["first"].(map[string]any), e["second"].(map[string]any), e["protocol"])
				expectedByKey[k] = e
				if combined[k] == nil {
					combined[k] = &total{Bytes: new(big.Int), Packets: new(big.Int), Members: map[string]bool{}}
				}
				v := combined[k]
				v.Count += len(e["source_rows"].([]any))
				for _, number := range e["source_rows"].([]any) {
					v.Members[fmt.Sprintf("nft_%s:%d", strings.ReplaceAll(input.TableUUID, "-", ""), int(number.(float64)))] = true
				}
				b, _ := new(big.Int).SetString(e["bytes"].(string), 10)
				p, _ := new(big.Int).SetString(e["packets"].(string), 10)
				v.Bytes.Add(v.Bytes, b)
				v.Packets.Add(v.Packets, p)
			}
			values := map[string]string{}
			for _, raw := range projection["vertices"].([]any) {
				v := raw.(map[string]any)
				values[v["source_entity_ref"].(map[string]any)["source_entity_id"].(string)] = v["properties"].(map[string]any)["endpoint_value"].(string)
			}
			for _, raw := range edges {
				edge := raw.(map[string]any)
				p := edge["properties"].(map[string]any)
				k := key(p["direction_kind"].(string), map[string]any{"ip": values[p["src_endpoint_id"].(string)], "port": p["src_port"]}, map[string]any{"ip": values[p["dst_endpoint_id"].(string)], "port": p["dst_port"]}, p["ip_protocol"])
				want := expectedByKey[k]
				if want == nil || p["bytes_sum"] != want["bytes"] || p["packets_sum"] != want["packets"] || int(p["flow_row_count"].(float64)) != len(want["source_rows"].([]any)) {
					return fmt.Errorf("%s graph totals differ", dataset.Name)
				}
				direction := "directed"
				if p["direction_kind"] == "unknown" {
					direction = "undirected"
				}
				if edge["direction"] != direction || p["counter_semantics"] != "conversation_total" {
					return fmt.Errorf("graph meaning differs")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		runtime.transport = nil
	}
	graph := runtime.request(http.MethodPost, "/api/v1/incidents/"+cases[0].IncidentID+"/network-flow/graphs/query", map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v3", "table_scope": map[string]any{"mode": "all_active_tables"}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}}, http.StatusOK)
	projection := graph["graph_projection_result"].(map[string]any)
	values := map[string]string{}
	for _, raw := range projection["vertices"].([]any) {
		v := raw.(map[string]any)
		values[v["source_entity_ref"].(map[string]any)["source_entity_id"].(string)] = v["properties"].(map[string]any)["endpoint_value"].(string)
	}
	edges := projection["edges"].([]any)
	if len(edges) != len(combined) {
		t.Fatal("cross-table semantic grouping differs")
	}
	count := 0
	checkedKinds := map[string]bool{}
	annotations := map[string]any{}
	for _, raw := range graph["edge_annotations"].([]any) {
		a := raw.(map[string]any)
		annotations[a["projected_edge_id"].(string)] = a["selector"]
	}
	bytes, packets := new(big.Int), new(big.Int)
	for _, raw := range edges {
		p := raw.(map[string]any)["properties"].(map[string]any)
		k := key(p["direction_kind"].(string), map[string]any{"ip": values[p["src_endpoint_id"].(string)], "port": p["src_port"]}, map[string]any{"ip": values[p["dst_endpoint_id"].(string)], "port": p["dst_port"]}, p["ip_protocol"])
		want := combined[k]
		if want == nil || p["bytes_sum"] != want.Bytes.String() || p["packets_sum"] != want.Packets.String() || int(p["flow_row_count"].(float64)) != want.Count {
			t.Fatal("cross-table graph lost repeated observations")
		}
		kind := p["direction_kind"].(string)
		if !checkedKinds[kind] && len(p["contributing_table_ids"].([]any)) > 1 {
			checkedKinds[kind] = true
			seen := map[string]bool{}
			path := "/api/v1/incidents/" + cases[0].IncidentID + "/network-flow/graphs/contributors/query"
			request := graphContributorRequest(graph, annotations[raw.(map[string]any)["edge_id"].(string)], 2)
			for {
				page := runtime.request(http.MethodPost, path, request, http.StatusOK)
				for _, item := range page["contributors"].([]any) {
					ref := item.(map[string]any)["row_ref"].(map[string]any)
					id := fmt.Sprintf("%s:%d", ref["network_flow_table_id"], int(ref["source_row_number"].(float64)))
					if !want.Members[id] || seen[id] {
						t.Fatal("contributor page contains wrong or repeated observation")
					}
					seen[id] = true
				}
				token, _ := page["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
				if token == "" {
					break
				}
				request = map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_continuation.v2", "cursor_token": token}
			}
			if len(seen) != want.Count {
				t.Fatal("contributor page lost overlapping observations")
			}
		}
		count += want.Count
		bytes.Add(bytes, want.Bytes)
		packets.Add(packets, want.Packets)
	}
	if !checkedKinds["unknown"] || !checkedKinds["client_to_server"] {
		t.Fatal("corpus did not exercise cross-table directed and undirected contributors")
	}
	if count != 4500 || bytes.String() != "3903637992" || packets.String() != "4412353" {
		t.Fatal("incident corpus totals changed")
	}
}
