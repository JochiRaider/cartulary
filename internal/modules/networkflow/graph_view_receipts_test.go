package networkflow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestGraphViewNameByteContract_Unit(t *testing.T) {
	for _, test := range []struct{ input, want, reason string }{
		{"  e\u0301\u2000", "é", ""},
		{strings.Repeat("é", 32), strings.Repeat("é", 32), ""},
		{strings.Repeat("é", 32) + "x", "", "display_name_too_long"},
		{strings.Repeat("e\u0301", 32), strings.Repeat("é", 32), ""},
		{"\u2000\u202f\u3000", "", "empty_display_name"},
		{"\ufeffgraph\ufeff", "\ufeffgraph\ufeff", ""},
		{"\tgraph", "", "forbidden_control"},
		{"\u0085", "", "forbidden_control"},
		{"\u009fgraph", "", "forbidden_control"},
	} {
		got, err := NormalizeGraphViewDisplayName(test.input)
		if test.reason == "" {
			if err != nil || got != test.want {
				t.Fatalf("normalize %q = %q, %v", test.input, got, err)
			}
		} else if invalid, ok := err.(*InvalidDisplayNameError); !ok || invalid.ReasonCode != test.reason {
			t.Fatalf("normalize %q error = %v; want %s", test.input, err, test.reason)
		}
	}
}

func TestGraphViewReceiptIntegrityAndComparison_Unit(t *testing.T) {
	incident, actor, jobID := uuid.New(), uuid.New(), uuid.New()
	query := canonicalJSON(map[string]any{"schema_id": schemaGraphSemanticQueryV2, "selected_table_ids": []string{"nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}})
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	declaration := GraphViewDeclaration{
		GraphViewID: "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", IncidentID: incident,
		DisplayName: "Repeated name", NormalizedDisplayName: "repeated name", DeclarationState: GraphViewDeclarationStateActive,
		SemanticQueryJSON: query, SemanticQuerySHA256: GraphViewSemanticQuerySHA256(query), DesiredSourceSnapshotID: "nfsnap_example",
		GraphViewVersion: 1, MaterializationGeneration: 1, CreatedByUserID: actor, CreatedAt: now, UpdatedAt: now, LatestJobID: &jobID,
	}
	receipt := graphViewAcceptedPayload(declaration, jobID)
	key := graphViewIdempotencyKey(routeKeyGraphViewsCreate, actor, incident, "graph-views", "one-attempt")
	job := jobs.Resource{JobID: jobID.String(), StatusRoute: "/api/v1/jobs/" + jobID.String(), Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incident}, Status: jobs.StatusQueued}
	original := canonicalJSON(receipt)
	if err := validateGraphViewReceipt(key, http.StatusAccepted, receipt, &job); err != nil {
		t.Fatal(err)
	}
	job.Status = jobs.StatusSucceeded
	job.ResultSummary = &jobs.ResultSummary{Code: "network_flow_graph_view_materialized", Message: "Saved graph materialized.", ResourceRefs: []jobs.ResourceRef{{Kind: graphViewResultResourceKind, ID: declaration.GraphViewID, Route: graphViewRoute(incident, declaration.GraphViewID)}}}
	if err := validateGraphViewReceipt(key, http.StatusAccepted, receipt, &job); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, canonicalJSON(receipt)) {
		t.Fatal("finalization validation altered the admission receipt")
	}
	for _, terminal := range []string{jobs.StatusFailed, jobs.StatusCanceled} {
		job.Status = terminal
		job.ResultSummary = nil
		if err := validateGraphViewReceipt(key, http.StatusAccepted, receipt, &job); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(original, canonicalJSON(receipt)) {
			t.Fatal("terminal failure or cancellation changed original receipt")
		}
	}
	// Current state is intentionally not an input to exact historical replay.
	declaration.DisplayName = "Renamed"
	declaration.GraphViewVersion++
	if err := validateGraphViewReceipt(key, http.StatusAccepted, receipt, nil); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(map[string]any){
		func(p map[string]any) {
			p["job"].(map[string]any)["status_route"] = "https://example.invalid/jobs/" + jobID.String()
		},
		func(p map[string]any) { p["job"].(map[string]any)["job_id"] = uuid.NewString() },
		func(p map[string]any) { p["graph_view"].(map[string]any)["incident_id"] = uuid.NewString() },
		func(p map[string]any) { delete(p["graph_view"].(map[string]any), "selected_result_binding") },
		func(p map[string]any) {
			p["graph_view"].(map[string]any)["last_failure_code"] = "network_flow_graph_materialization_timeout"
		},
		func(p map[string]any) {
			p["graph_view"].(map[string]any)["selected_result_binding"] = map[string]any{"projection_result_id": "gpres_" + strings.Repeat("a", 64)}
		},
		func(p map[string]any) { p["job_kind"] = GraphViewMaterializationJobKind },
	} {
		var invalid map[string]any
		if err := json.Unmarshal(original, &invalid); err != nil {
			t.Fatal(err)
		}
		mutate(invalid)
		if err := validateGraphViewReceipt(key, http.StatusAccepted, invalid, nil); err == nil {
			t.Fatalf("accepted malformed receipt: %#v", invalid)
		}
	}
	retireKey := graphViewIdempotencyKey(routeKeyGraphViewsDelete, actor, incident, declaration.GraphViewID, "retire")
	if err := validateGraphViewReceipt(retireKey, http.StatusNoContent, map[string]any{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := validateGraphViewReceipt(retireKey, http.StatusOK, map[string]any{}, nil); err == nil {
		t.Fatal("accepted old retirement status")
	}
	body := map[string]any{"base_graph_view_version": int64(1)}
	first := graphViewMutationBytes(routeKeyGraphViewsRefresh, "graph_view_id:"+declaration.GraphViewID, body)
	if !bytes.Equal(first, graphViewMutationBytes(routeKeyGraphViewsRefresh, "graph_view_id:"+declaration.GraphViewID, body)) {
		t.Fatal("immutable attempt comparison changed")
	}
	if bytes.Equal(first, graphViewMutationBytes(routeKeyGraphViewsDelete, "graph_view_id:"+declaration.GraphViewID, body)) {
		t.Fatal("comparison omitted route identity")
	}
	if bytes.Equal(first, graphViewMutationBytes(routeKeyGraphViewsRefresh, "graph_view_id:nfgv_other", body)) {
		t.Fatal("comparison omitted target identity")
	}
}
