package imports_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
)

// The object-store barrier pauses the first export member after the incident
// source has fixed its database snapshot. All concurrent mutations remain public
// owner operations; no production test hook or alternate export path is used.
type exportSnapshotBarrier struct {
	objectstore.Store
	objectstore.TypedStore
	armed   atomic.Bool
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (s *exportSnapshotBarrier) resume() { s.once.Do(func() { close(s.release) }) }

func (s *exportSnapshotBarrier) PutObject(ctx context.Context, key string, reader io.Reader, size int64, media string) error {
	if s.armed.CompareAndSwap(true, false) {
		close(s.entered)
		select {
		case <-s.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.Store.PutObject(ctx, key, reader, size, media)
}

func TestNetworkFlowExportUsesOneSnapshotDuringConcurrentMutations_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	barrier := &exportSnapshotBarrier{entered: make(chan struct{}), release: make(chan struct{})}
	defer barrier.resume()
	options := appsupport.ServerOptions{Prefix: "nf-export-snapshot", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}, ConfigureRuntime: func(options *server.Options) {
		barrier.Store = options.ObjectStore
		barrier.TypedStore = options.ObjectStore.(objectstore.TypedStore)
		options.ObjectStore = barrier
	}}
	h := runtime.StartServer(t, options)
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	id := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "snapshot-incident", "incident_key": "NF-SNAPSHOT", "title": "Coherent export"})["incident_id"].(string)
	c := irPublicClient{t: t, origin: h.Server.HTTP.URL, login: login, incident: id}
	table := networkflowsupport.SeedGraphSource(t, http.DefaultClient, c.origin, login, id)
	base := "/api/v1/incidents/" + id + "/network-flow"
	created := c.request(http.MethodPost, base+"/graph-views", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_create_request.v5", "refresh_scope": "all_active_tables", "client_txn_id": "snapshot-graph", "display_name": "Selected before concurrent changes",
		"semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v3", "selected_table_ids": []string{table}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}},
	}, http.StatusAccepted)
	if waitIRApply(t, c.origin, login, created["job"].(map[string]any)["job_id"].(string))["status"] != "succeeded" {
		t.Fatal("initial graph failed")
	}
	graphPath := base + "/graph-views/" + created["graph_view"].(map[string]any)["graph_view_id"].(string)
	before := readPublicBundle(t, c.origin, login, id, "snapshot-before")
	barrier.armed.Store(true)
	export := c.request(http.MethodPost, "/api/v1/incident-bundles/export", map[string]any{"incident_id": id, "client_txn_id": "snapshot-concurrent", "optional_sections": []string{}}, http.StatusAccepted)
	select {
	case <-barrier.entered:
	case <-time.After(20 * time.Second):
		t.Fatal("export did not reach member snapshot barrier")
	case <-t.Context().Done():
		t.Fatal(t.Context().Err())
	}
	// Import a distinct successful source, then resolve all-active refresh intent.
	sourceID := irCommittedSources(t, c.origin, login, id)[0]["committed_source_id"].(string)
	sourcePath := "/api/v1/incidents/" + id + "/committed-sources/" + sourceID
	runRoot, err := networkflowfixtures.SelectedRunRoot()
	if err != nil {
		t.Fatal(err)
	}
	var mapping json.RawMessage
	_, err = networkflowfixtures.Run(networkflowfixtures.Config{CorpusRoot: fixtures.Path("../../..", "fixtures", "network-flow"), RunRoot: runRoot, FixtureID: "NF-FIX-001-cisco-sna-minimal", Revision: 11}, func(bundle *networkflowfixtures.Bundle) error {
		raw, err := bundle.Read("source/cases.json")
		if err != nil {
			return err
		}
		var cases []struct {
			Mapping json.RawMessage `json:"mapping"`
		}
		if err := json.Unmarshal(raw, &cases); err != nil {
			return err
		}
		if len(cases) != 1 {
			t.Fatal("minimal fixture requires exactly one mapping")
		}
		mapping = cases[0].Mapping
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	content := doImportJSON(t, c.origin, login, http.MethodGet, sourcePath+"/content", nil)
	input, err := io.ReadAll(content.Body)
	content.Body.Close()
	if err != nil || content.StatusCode != http.StatusOK {
		t.Fatal("committed source unavailable", err)
	}
	second := networkflowsupport.ImportCSV(t, http.DefaultClient, c.origin, login, id, "concurrent-source", "concurrent.csv", input, mapping)
	if second == table {
		t.Fatal("fresh import reused an original flow table")
	}
	c.request(http.MethodPatch, base+"/tables/"+table, map[string]any{"client_txn_id": "snapshot-rename", "base_table_version": 1, "display_name": "Concurrent rename"}, http.StatusOK)
	current := c.request(http.MethodGet, graphPath, nil, http.StatusOK)
	refresh := c.request(http.MethodPost, graphPath+"/refresh", map[string]any{"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "snapshot-refresh", "base_graph_view_version": current["graph_view"].(map[string]any)["graph_view_version"]}, http.StatusAccepted)
	if waitIRApply(t, c.origin, login, refresh["job"].(map[string]any)["job_id"].(string))["status"] != "succeeded" {
		t.Fatal("concurrent graph refresh failed")
	}
	c.request(http.MethodDelete, base+"/tables/"+table, map[string]any{"client_txn_id": "snapshot-delete", "base_table_version": 2}, http.StatusOK)
	current = c.request(http.MethodGet, graphPath, nil, http.StatusOK)
	httptestx.RequireNoContent(t, doImportJSON(t, c.origin, login, http.MethodDelete, graphPath, map[string]any{"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "snapshot-retire", "base_graph_view_version": current["graph_view"].(map[string]any)["graph_view_version"]}))
	barrier.resume()
	terminal := waitIRApply(t, c.origin, login, export["job_id"].(string))
	if terminal["status"] != "succeeded" {
		t.Fatal("concurrent export failed")
	}
	ref := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	response := doImportJSON(t, c.origin, login, http.MethodGet, "/api/v1/incident-bundles/"+ref["id"].(string)+"/content", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("concurrent bundle unavailable")
	}
	after, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"data/incident.json", "data/committed_sources.ndjson", "ext/extensions/network_flow_activity/members/tables.ndjson", "ext/extensions/network_flow_activity/members/observations.ndjson", "ext/extensions/network_flow_activity/members/diagnostics.ndjson", "ext/extensions/network_flow_activity/members/indicator_bindings.ndjson", "ext/extensions/network_flow_activity/members/graph_views.ndjson"} {
		if !bytes.Equal(sourceBundleMember(t, before, path), sourceBundleMember(t, after, path)) {
			t.Fatalf("export mixed database snapshots: %s", path)
		}
	}
}
