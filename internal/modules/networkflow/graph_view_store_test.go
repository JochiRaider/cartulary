package networkflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

const graphViewResultID = "gpres_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestNetworkFlowGraphViewDeclarationPersistence_Integration(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-graph-views")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 20, 0, 0, 0, time.UTC)

	first := graphViewDeclarationFixture("nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", incidentID, actor.ID, now)
	second := graphViewDeclarationFixture("nfgv_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", incidentID, actor.ID, now.Add(time.Second))
	second.SelectedResult = &GraphViewSelectedResultBinding{
		ProjectionResultID:            graphViewResultID,
		SourceSnapshotID:              "network-flow-source-1",
		ProjectionSchemaID:            "graph_projection.v2",
		ProjectionVersion:             "network-flow-graph-v1",
		NormalizedConfigurationSHA256: testSHA1,
		NormalizedSourceSHA256:        testSHA2,
		CanonicalOutputSHA256:         testSHA3,
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin graph declaration transaction: %v", err)
	}
	for _, declaration := range []GraphViewDeclaration{second, first} {
		if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("insert graph declaration: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit graph declarations: %v", err)
	}

	// The selected binding may be restored before its derived result. This is
	// the intentional absence of an authoritative-to-derived foreign key.
	loaded, err := store.GetGraphViewDeclaration(ctx, incidentID, second.GraphViewID)
	if err != nil || loaded.SelectedResult == nil || loaded.SelectedResult.ProjectionResultID != graphViewResultID {
		t.Fatalf("read selected graph declaration got %#v err=%v", loaded, err)
	}
	listed, err := store.ListActiveGraphViewDeclarations(ctx, incidentID)
	if err != nil {
		t.Fatalf("list active graph declarations: %v", err)
	}
	if len(listed) != 2 || listed[0].GraphViewID != first.GraphViewID || listed[1].GraphViewID != second.GraphViewID {
		t.Fatalf("duplicate-name declaration order = %#v", listed)
	}

	invalid := first
	invalid.GraphViewID = "nfgv_cccccccccccccccccccccccccccccccc"
	invalid.SelectedResult = &GraphViewSelectedResultBinding{ProjectionResultID: graphViewResultID}
	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin invalid graph declaration transaction: %v", err)
	}
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, invalid); !errors.Is(err, ErrGraphViewDeclarationInvalid) {
		t.Fatalf("incomplete selected binding got %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback invalid graph declaration: %v", err)
	}
}

func TestNetworkFlowGraphViewPublicationRejectsStaleGenerationAndAllowsRename_Integration(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-graph-publication-cas")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 20, 30, 0, 0, time.UTC)
	graphViewID := "nfgv_dddddddddddddddddddddddddddddddd"
	jobID := uuid.New()
	declaration := graphViewDeclarationFixture(graphViewID, incidentID, actor.ID, now)
	declaration.LatestJobID = &jobID
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy, handler_name,
    job_kind, progress_unit_id
) VALUES (
    $1, 'incident', $2, 'queued', true, $3, $4, $4, 0,
    'incident_membership', $5, $6,
    'network_flow_activity.graph_view_materialize.projection_result.v1'
)
`, jobID, incidentID, actor.ID, now, GraphViewWorkerKind, GraphViewMaterializationJobKind); err != nil {
		t.Fatalf("seed graph materialization job: %v", err)
	}
	result := graphResultForReportingSource(graphViewID, now)
	selected := GraphViewSelectedResultBinding{
		ProjectionResultID: result.Binding.ProjectionResultID, SourceSnapshotID: result.Binding.SourceSnapshotID,
		ProjectionSchemaID: result.Binding.ProjectionSchemaID, ProjectionVersion: result.Binding.ProjectionVersion,
		NormalizedConfigurationSHA256: result.Binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        result.Binding.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         result.Binding.CanonicalOutputSHA256,
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin declaration transaction: %v", err)
	}
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		t.Fatalf("insert graph declaration: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit graph declaration: %v", err)
	}

	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin stale publication transaction: %v", err)
	}
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct stale publication adapter: %v", err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		t.Fatalf("stage stale immutable result: %v", err)
	}
	if _, err := store.PublishGraphViewResultTx(ctx, tx, incidentID, graphViewID, 2, declaration.DesiredSourceSnapshotID, jobID, selected, now); !errors.Is(err, ErrGraphViewPublicationStale) {
		t.Fatalf("stale generation publication error = %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback stale publication: %v", err)
	}
	var resultCount int
	if err := harness.DB.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, result.Binding.ProjectionResultID).Scan(&resultCount); err != nil || resultCount != 0 {
		t.Fatalf("stale generation leaked immutable result count=%d err=%v", resultCount, err)
	}

	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin concurrent rename transaction: %v", err)
	}
	renamed, err := store.RenameGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, 1, "Renamed graph", "renamed graph", now.Add(time.Second))
	if err != nil {
		t.Fatalf("rename graph declaration: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit graph rename: %v", err)
	}
	if renamed.GraphViewVersion != 2 || renamed.MaterializationGeneration != 1 {
		t.Fatalf("rename changed materialization generation: %#v", renamed)
	}

	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin renamed publication transaction: %v", err)
	}
	publisher, err = postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct renamed publication adapter: %v", err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		t.Fatalf("publish immutable result after rename: %v", err)
	}
	published, err := store.PublishGraphViewResultTx(ctx, tx, incidentID, graphViewID, 1, declaration.DesiredSourceSnapshotID, jobID, selected, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("generation-stable rename blocked publication: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit renamed publication: %v", err)
	}
	if published.GraphViewVersion != 2 || published.MaterializationGeneration != 1 || published.SelectedResult == nil || published.SelectedResult.ProjectionResultID != result.Binding.ProjectionResultID {
		t.Fatalf("renamed publication drifted: %#v", published)
	}

	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin timeout failure transaction: %v", err)
	}
	if err := store.RecordGraphViewMaterializationFailureTx(
		ctx, tx, incidentID, graphViewID, 1, jobID,
		"network_flow_graph_materialization_timeout", now.Add(3*time.Second),
	); err != nil {
		t.Fatalf("record terminal timeout failure: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit terminal timeout failure: %v", err)
	}
	afterTimeout, err := store.GetGraphViewDeclaration(ctx, incidentID, graphViewID)
	if err != nil || afterTimeout.SelectedResult == nil || afterTimeout.SelectedResult.ProjectionResultID != result.Binding.ProjectionResultID ||
		afterTimeout.LastFailureCode == nil || *afterTimeout.LastFailureCode != "network_flow_graph_materialization_timeout" {
		t.Fatalf("terminal timeout did not preserve prior selected result: declaration=%#v err=%v", afterTimeout, err)
	}
}

func graphViewDeclarationFixture(graphViewID string, incidentID, actorID uuid.UUID, now time.Time) GraphViewDeclaration {
	semanticQuery := []byte(`{"aggregation":{"include_example_row_refs":true,"mode":"default_flow_edge_v1"},"filters":[],"result_limits":{"max_aggregate_counter_digits":39,"max_edges":10000,"max_example_row_refs_per_edge":10,"max_vertices":5000},"schema_id":"cartulary.network_flow.graph_semantic_query.v1","selected_table_ids":["nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],"time_range":{"end_utc":null,"start_utc":null}}`)
	return GraphViewDeclaration{
		GraphViewID:               graphViewID,
		IncidentID:                incidentID,
		DisplayName:               "Shared graph",
		NormalizedDisplayName:     "shared graph",
		DeclarationState:          GraphViewDeclarationStateActive,
		SemanticQueryJSON:         semanticQuery,
		SemanticQuerySHA256:       GraphViewSemanticQuerySHA256(semanticQuery),
		DesiredSourceSnapshotID:   "network-flow-source-1",
		GraphViewVersion:          1,
		MaterializationGeneration: 1,
		CreatedByUserID:           actorID,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
}
