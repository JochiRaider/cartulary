package networkflow_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSavedGraphMaterializationRevalidatesSubmitterRole_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "saved-graph-publication-role")
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	actorID := uuid.MustParse(actorText)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "authority-incident", "incident_key": "IR-SG-AUTHORITY", "title": "Saved graph authority",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	table := createTestTable(t, harness.Pool, store, actorID, incidentID, "authority.csv", nil, 1)
	collection := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graph-views"
	response := httptestx.DoJSON(t, http.MethodPost, collection, map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": "authority-create", "display_name": "Authorized result",
		"semantic_query": map[string]any{
			"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true},
		},
	}, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	created := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	graphID := created["graph_view"].(map[string]any)["graph_view_id"].(string)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, login, created["job"].(map[string]any)["job_id"].(string), "succeeded")
	ctx := context.Background()
	initial, err := store.GetGraphViewDeclaration(ctx, incidentID, graphID)
	if err != nil {
		t.Fatal(err)
	}
	catalog := collaborationsupport.NewJobCatalog()
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: harness.Pool, Catalog: catalog, Transactions: transactions, Policy: jobs.ProductionRuntimePolicy()})
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name, role                 string
		beforePublication, allowed bool
	}{
		{"viewer before computation", "viewer", false, false},
		{"reviewer before publication", "reviewer", true, false},
		{"editor publication", "editor", true, true},
		{"admin publication", "admin", true, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			setRole := func(role string) {
				t.Helper()
				if _, err := harness.Pool.Exec(ctx, "UPDATE incident_memberships SET role=$3 WHERE incident_id=$1 AND user_id=$2", incidentID, actorID, role); err != nil {
					t.Fatal(err)
				}
			}
			setRole("admin")
			current, err := store.GetGraphViewDeclaration(ctx, incidentID, graphID)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			job, err := transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: collaborationsupport.TestJobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, SubmittedByUserID: actorID, Progress: jobs.Progress{Completed: 0}}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			current, err = store.RefreshGraphViewDeclarationTx(ctx, tx, incidentID, graphID, current.GraphViewVersion, current.DesiredSourceSnapshotID, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			jobID := uuid.MustParse(job.JobID)
			if _, err = store.SetGraphViewLatestJobTx(ctx, tx, incidentID, graphID, jobID); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			execution, claimed, err := manager.Claim(ctx, jobID)
			if err != nil || !claimed {
				t.Fatalf("claim controlled attempt: %v %v", claimed, err)
			}
			payload, err := json.Marshal(map[string]any{"schema_id": "cartulary.network_flow.graph_view_materialization_payload.v1", "incident_id": incidentID, "graph_view_id": graphID, "materialization_generation": current.MaterializationGeneration, "source_snapshot_id": current.DesiredSourceSnapshotID})
			if err != nil {
				t.Fatal(err)
			}
			finalizer := &authorityGraphFinalizer{pool: harness.Pool}
			if scenario.beforePublication {
				finalizer.beforePublication = func() { setRole(scenario.role) }
			} else {
				setRole(scenario.role)
			}
			handler := NewGraphViewAuthorityTestHandler(store, authorityGraphManager{Manager: manager, payload: payload}, finalizer)
			if err := handler(ctx, execution); err != nil {
				t.Fatal(err)
			}
			if finalizer.published != scenario.allowed || finalizer.failed == scenario.allowed {
				t.Fatalf("published=%v failed=%v, allowed=%v", finalizer.published, finalizer.failed, scenario.allowed)
			}
			after, err := store.GetGraphViewDeclaration(ctx, incidentID, graphID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(after.SelectedResult, initial.SelectedResult) {
				t.Fatal("authority transition changed the last safe binding")
			}
			if !scenario.allowed && (after.LastFailureCode == nil || *after.LastFailureCode != "network_flow_graph_materialization_publication_conflict") {
				t.Fatalf("failure = %v", after.LastFailureCode)
			}
		})
	}
}

type authorityGraphManager struct {
	*jobs.Manager
	payload json.RawMessage
}

func (m authorityGraphManager) HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error) {
	return m.payload, nil
}

type authorityGraphFinalizer struct {
	pool              *pgxpool.Pool
	beforePublication func()
	published, failed bool
}

func (f *authorityGraphFinalizer) FinalizeGraphViewJobSuccess(ctx context.Context, request GraphViewJobSuccessFinalization) (jobs.Resource, error) {
	if f.beforePublication != nil {
		f.beforePublication()
	}
	err := f.mutate(ctx, request.Mutate)
	f.published = err == nil
	return jobs.Resource{}, err
}
func (f *authorityGraphFinalizer) FinalizeGraphViewJobFailure(ctx context.Context, request GraphViewJobFailureFinalization) (jobs.Resource, error) {
	f.failed = true
	return jobs.Resource{}, f.mutate(ctx, request.Mutate)
}
func (f *authorityGraphFinalizer) mutate(ctx context.Context, mutate GraphViewJobMutation) error {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := mutate(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
