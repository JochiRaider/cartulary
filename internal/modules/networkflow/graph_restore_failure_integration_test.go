package networkflow_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/app/graphassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestNetworkFlowGraphReconstructionBindingFailureRollsBack_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "graph-reconstruction-binding-failure")
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "restore-failure-incident", "incident_key": "IR-GRAPH-ROLLBACK", "title": "Graph reconstruction rollback",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	table := createTestTable(t, harness.Pool, store, actorID, incidentID, "restore-failure.csv", nil, 1)
	collection := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graph-views"
	response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, collection, map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_create_request.v5", "refresh_scope": "selected_tables", "client_txn_id": "restore-failure-create", "display_name": "Rollback binding",
		"semantic_query": map[string]any{
			"schema_id": "cartulary.network_flow.graph_semantic_query.v3", "selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true},
		},
	}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	created := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	graphViewID := created["graph_view"].(map[string]any)["graph_view_id"].(string)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, login, created["job"].(map[string]any)["job_id"].(string), "succeeded")

	// Real materialization establishes the source, selected binding and derived
	// content. Stop workers before corrupting only the isolated fixture's binding.
	harness.Server.Close()
	declaration, err := store.GetGraphViewDeclaration(t.Context(), incidentID, graphViewID)
	if err != nil || declaration.SelectedResult == nil {
		t.Fatalf("materialized declaration: %#v, %v", declaration, err)
	}
	wrongDigest := strings.Repeat("f", 64)
	if declaration.SelectedResult.CanonicalOutputSHA256 == wrongDigest {
		wrongDigest = strings.Repeat("e", 64)
	}
	if _, err := harness.Pool.Exec(t.Context(), `UPDATE network_flow_graph_views SET selected_canonical_output_sha256=$2, selected_source_binding_json=jsonb_set(selected_source_binding_json,'{expected_result_binding,canonical_output_sha256}',to_jsonb($2::text)) WHERE graph_view_id=$1`, graphViewID, wrongDigest); err != nil {
		t.Fatal(err)
	}
	var results, vertices, edges int
	if err := harness.Pool.QueryRow(t.Context(), `SELECT
 (SELECT count(*) FROM graph_projection_results),
 (SELECT count(*) FROM graph_projection_result_vertices),
 (SELECT count(*) FROM graph_projection_result_edges)`).Scan(&results, &vertices, &edges); err != nil || results != 1 || vertices == 0 || edges == 0 {
		t.Fatalf("fixture must populate all derived tables: %d/%d/%d, %v", results, vertices, edges, err)
	}
	before := networkFlowReconstructionState(t, harness.Pool)
	dsn, err := harness.Database.DSNForPurpose(postgres.PurposeRuntime)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRuntime, ExpectedRole: "cartulary_runtime"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(handle.Close)
	observed := &networkFlowReconstructionWitness{DB: handle.Pool()}
	registry := graphrestore.CurrentRestoreSourceRegistry()
	binding := graphrestore.CurrentRestoreImplementationBinding()
	result, err := graphassembly.RebuildClaimedGraph(t.Context(), observed, graphrestore.RestoreRebuildRequest{
		Context: t.Context(), RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000009201"),
		TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000009203"), PublishedAt: time.Date(2026, 7, 10, 13, 0, 0, 0, time.UTC),
		RecoveryStateCatalog: graphrestore.RestoreRecoveryCatalogRef{DigestSHA256: binding.Binding.RecoveryStateCatalogSHA256, AlgorithmID: graphrestore.RestoreAlgorithmID, GraphTableIDs: graphrestore.RestoreGraphTableIDs()},
		SourceRegistry:       graphrestore.RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()}, ImplementationBinding: binding,
	})
	if code, ok := graphrestore.RestoreErrorCodeOf(err); !ok || code != graphrestore.RestoreErrorInvalidCandidate {
		t.Fatalf("real derived binding mismatch did not reject after enumeration: code=%q err=%v", code, err)
	}
	if result.Validate() != nil || result.Status != graphrestore.RestoreStatusFailed || result.ReadinessOutcome != graphrestore.RestoreReadinessIncomplete ||
		len(result.ClearedTableIDs) != 0 || len(result.RebuiltViews) != 0 || result.ReconciledNonterminalJobCount != 0 || result.PostconditionSHA256 != nil {
		t.Fatalf("binding failure claimed committed reconstruction: %#v", result)
	}
	if observed.transactions != 1 || observed.commits != 0 || observed.rollbacks != 1 || !observed.rollbackSucceeded || !observed.cleared || !observed.enumeratedAfterClear ||
		observed.sourceOutside != 0 || observed.isolation != "repeatable read" || observed.role != "cartulary_runtime" {
		t.Fatalf("binding failure transaction witness: %#v", observed)
	}
	// This independent pool read runs before cleanup and uses the test context,
	// not a failed reconstruction transaction. Include full job/input rows.
	if after := networkFlowReconstructionState(t, harness.Pool); after != before {
		t.Fatal("binding failure changed committed Graph, declaration, job or retained input state")
	}
}

func networkFlowReconstructionState(t *testing.T, db postgres.DB) string {
	t.Helper()
	var state string
	if err := db.QueryRow(t.Context(), `SELECT jsonb_build_array(
 (SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY projection_result_id), '[]') FROM graph_projection_results r),
 (SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY projection_result_id, vertex_id), '[]') FROM graph_projection_result_vertices v),
 (SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY projection_result_id, edge_id), '[]') FROM graph_projection_result_edges e),
 (SELECT coalesce(jsonb_agg(to_jsonb(d) ORDER BY graph_view_id), '[]') FROM network_flow_graph_views d),
 (SELECT coalesce(jsonb_agg(to_jsonb(j) ORDER BY job_id), '[]') FROM jobs j),
 (SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY job_id), '[]') FROM reporting_job_payloads p)
)::text`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

type networkFlowReconstructionWitness struct {
	postgres.DB
	transactions, commits, rollbacks, sourceOutside  int
	cleared, enumeratedAfterClear, rollbackSucceeded bool
	isolation, role                                  string
}

func (db *networkFlowReconstructionWitness) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM network_flow_graph_views") {
		db.sourceOutside++
	}
	return db.DB.Query(ctx, sql, args...)
}

func (db *networkFlowReconstructionWitness) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	if err := tx.QueryRow(ctx, "SELECT current_setting('transaction_isolation'), current_user").Scan(&db.isolation, &db.role); err != nil {
		_ = tx.Rollback(context.WithoutCancel(ctx))
		return nil, err
	}
	db.transactions++
	return &networkFlowReconstructionTx{Tx: tx, owner: db}, nil
}

type networkFlowReconstructionTx struct {
	pgx.Tx
	owner *networkFlowReconstructionWitness
}

func (tx *networkFlowReconstructionTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := tx.Tx.Exec(ctx, sql, args...)
	if err == nil && strings.Contains(sql, "DELETE FROM graph_projection_results") {
		var remaining int
		if err := tx.Tx.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM graph_projection_results) +
 (SELECT count(*) FROM graph_projection_result_vertices) +
 (SELECT count(*) FROM graph_projection_result_edges)`).Scan(&remaining); err != nil {
			return tag, err
		}
		if remaining != 0 {
			return tag, fmt.Errorf("clear left %d derived rows", remaining)
		}
		tx.owner.cleared = true
	}
	return tag, err
}

func (tx *networkFlowReconstructionTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM network_flow_graph_views") {
		tx.owner.enumeratedAfterClear = tx.owner.cleared
	}
	return tx.Tx.Query(ctx, sql, args...)
}

func (tx *networkFlowReconstructionTx) Commit(ctx context.Context) error {
	tx.owner.commits++
	return tx.Tx.Commit(ctx)
}

func (tx *networkFlowReconstructionTx) Rollback(ctx context.Context) error {
	tx.owner.rollbacks++
	err := tx.Tx.Rollback(ctx)
	tx.owner.rollbackSucceeded = err == nil
	return err
}
