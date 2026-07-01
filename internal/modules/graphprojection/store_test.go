package graphprojection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestStoreLifecycleRetentionAndInvalidation(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-lifecycle")
	store := NewStore(db)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))

	first, err := store.CreateProjection(ctx, input, StoreProjectionOptions{
		ProjectionRunNonce: "store-run-1",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime().Add(time.Second),
	})
	if err != nil {
		t.Fatalf("create first projection: %v", err)
	}
	if first.State != RunStateAvailable {
		t.Fatalf("first state = %s", first.State)
	}
	graph, err := store.GetGraphView(ctx, first.GraphViewID, "")
	if err != nil {
		t.Fatalf("load graph view: %v", err)
	}
	if graph.ProjectionRunID != first.ProjectionRunID {
		t.Fatalf("latest run = %s want %s", graph.ProjectionRunID, first.ProjectionRunID)
	}

	second, err := store.RefreshProjection(ctx, input, StoreProjectionOptions{
		ProjectionRunNonce: "store-run-2",
		AcceptedAt:         fixedTime().Add(2 * time.Second),
		GeneratedAt:        fixedTime().Add(3 * time.Second),
	})
	if err != nil {
		t.Fatalf("refresh projection: %v", err)
	}
	if second.PreviousProjectionRunID == nil || *second.PreviousProjectionRunID != first.ProjectionRunID {
		t.Fatalf("previous run not recorded: %#v", second.PreviousProjectionRunID)
	}
	reloadedFirst, err := store.GetProjectionRun(ctx, first.ProjectionRunID)
	if err != nil {
		t.Fatalf("reload first run: %v", err)
	}
	if reloadedFirst.State != RunStateReplaced || reloadedFirst.RetentionExpiresAt == nil {
		t.Fatalf("first run replacement state = %s retention=%v", reloadedFirst.State, reloadedFirst.RetentionExpiresAt)
	}

	summary, err := store.InvalidateGraphView(ctx, first.GraphViewID, "source_snapshot_expired", "2026-05-30T00:00:05Z", "fixture")
	if err != nil {
		t.Fatalf("invalidate graph view: %v", err)
	}
	if len(summary.InvalidatedRunIDs) != 2 {
		t.Fatalf("invalidated runs = %#v", summary.InvalidatedRunIDs)
	}
	if _, err := store.GetGraphView(ctx, first.GraphViewID, ""); !errors.Is(err, ErrGraphViewUnavailable) {
		t.Fatalf("get invalidated graph view err = %v", err)
	}
}

func TestStorePreAdmissionFailureTouchesNoState(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-fail-before-touch")
	store := NewStore(db)
	ctx := context.Background()

	_, err := store.CreateProjection(ctx, []byte(`{"projection_schema_id":"graph_projection.v1","projection_schema_id":"graph_projection.v1"}`), StoreProjectionOptions{})
	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected operation error, got %T %v", err, err)
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_views`).Scan(&count); err != nil {
		t.Fatalf("count graph views: %v", err)
	}
	if count != 0 {
		t.Fatalf("pre-admission failure touched graph state: count=%d", count)
	}
}

func TestQueryLifecycleStateMatrix(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-query-lifecycle")
	store := NewStore(db)
	ctx := context.Background()

	available, err := store.CreateProjection(ctx, mustJSON(t, retargetGraphInput(t, incidentGraphInput(t), "lifecycle_available")), StoreProjectionOptions{
		ProjectionRunNonce: "query-lifecycle-available",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime().Add(time.Second),
	})
	if err != nil {
		t.Fatalf("create available projection: %v", err)
	}
	if _, err := store.GetGraphView(ctx, available.GraphViewID, ""); err != nil {
		t.Fatalf("available graph should be readable: %v", err)
	}

	failedInput := retargetGraphInput(t, incidentGraphInput(t), "lifecycle_failed")
	config := failedInput["projection_config"].(map[string]any)
	relMappings := config["relationship_mappings"].([]any)
	relMappings[0].(map[string]any)["direction_policy"] = "preserve"
	relMappings[0].(map[string]any)["emit_reverse_edge"] = true
	failed, err := store.CreateProjection(ctx, mustJSON(t, failedInput), StoreProjectionOptions{
		ProjectionRunNonce: "query-lifecycle-failed",
		AcceptedAt:         fixedTime().Add(2 * time.Second),
		GeneratedAt:        fixedTime().Add(3 * time.Second),
	})
	if err != nil {
		t.Fatalf("create failed projection: %v", err)
	}
	reloadedFailed, err := store.GetProjectionRun(ctx, failed.ProjectionRunID)
	if err != nil {
		t.Fatalf("inspect failed run: %v", err)
	}
	if reloadedFailed.State != RunStateFailed || reloadedFailed.GraphView != nil {
		t.Fatalf("failed run state=%s graph=%#v", reloadedFailed.State, reloadedFailed.GraphView)
	}
	if _, err := store.GetGraphView(ctx, failed.GraphViewID, ""); !errors.Is(err, ErrGraphViewUnavailable) {
		t.Fatalf("failed graph view err = %v", err)
	}

	computingGraphViewID, err := DeriveGraphViewID("lifecycle_computing")
	if err != nil {
		t.Fatalf("derive computing graph id: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id,
    graph_view_key,
    state,
    latest_projection_run_id,
    latest_source_snapshot_id,
    projection_version,
    updated_at,
    validation_status
) VALUES ($1, 'lifecycle_computing', 'creating', 'gr_lifecycle_computing', 'snap_lifecycle_computing', 'v1', $2, 'pending')
`, computingGraphViewID, fixedTime()); err != nil {
		t.Fatalf("insert computing graph view: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO graph_projection_runs (
    projection_run_id,
    graph_view_id,
    source_snapshot_id,
    projection_version,
    state,
    projection_run_nonce,
    projection_config_digest,
    projection_source_digest,
    accepted_at,
    validation_summary_json
) VALUES ('gr_lifecycle_computing', $1, 'snap_lifecycle_computing', 'v1', 'computing', 'query-lifecycle-computing', 'sha256:config', 'sha256:source', $2, '{"Status":"pending"}'::jsonb)
`, computingGraphViewID, fixedTime()); err != nil {
		t.Fatalf("insert computing run: %v", err)
	}
	computing, err := store.GetProjectionRun(ctx, "gr_lifecycle_computing")
	if err != nil {
		t.Fatalf("inspect computing run: %v", err)
	}
	if computing.State != RunStateComputing {
		t.Fatalf("computing state = %s", computing.State)
	}
	if _, err := store.GetGraphView(ctx, computingGraphViewID, ""); !errors.Is(err, ErrGraphViewUnavailable) {
		t.Fatalf("computing graph view err = %v", err)
	}
}

func TestListGraphViewsPagination(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-idempotency")
	store := NewStore(db)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))

	first, err := store.CreateProjection(ctx, input, StoreProjectionOptions{
		ProjectionRunNonce: "idem-run-1",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime(),
		IdempotencyKey:     "idem-key-1",
	})
	if err != nil {
		t.Fatalf("create projection: %v", err)
	}
	replayed, err := store.CreateProjection(ctx, input, StoreProjectionOptions{
		ProjectionRunNonce: "idem-run-1",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime(),
		IdempotencyKey:     "idem-key-1",
	})
	if err != nil {
		t.Fatalf("replay projection: %v", err)
	}
	if replayed.ProjectionRunID != first.ProjectionRunID {
		t.Fatalf("replayed run = %s want %s", replayed.ProjectionRunID, first.ProjectionRunID)
	}

	other := minimalInput(t, "empty")
	if _, err := store.CreateProjection(ctx, mustJSON(t, other), StoreProjectionOptions{
		ProjectionRunNonce: "idem-run-2",
		AcceptedAt:         fixedTime().Add(time.Second),
		GeneratedAt:        fixedTime().Add(time.Second),
	}); err != nil {
		t.Fatalf("create second graph view: %v", err)
	}

	page, cursor, err := store.ListGraphViews(ctx, ListGraphViewsOptions{Limit: 1, Now: fixedTime()})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(page) != 1 || cursor == "" {
		t.Fatalf("page len=%d cursor=%q", len(page), cursor)
	}
	if _, err := db.Exec(ctx, `DELETE FROM graph_projection_views WHERE graph_view_id = $1`, page[0].GraphViewID); err != nil {
		t.Fatalf("delete cursor anchor: %v", err)
	}
	next, _, err := store.ListGraphViews(ctx, ListGraphViewsOptions{Limit: 1, CursorToken: cursor, Now: fixedTime().Add(time.Minute)})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(next) != 1 || next[0].GraphViewID == page[0].GraphViewID {
		t.Fatalf("bad second page: first=%#v next=%#v", page, next)
	}
	if _, _, err := store.ListGraphViews(ctx, ListGraphViewsOptions{CursorToken: cursor, Now: fixedTime().Add(16 * time.Minute)}); !errors.Is(err, ErrCursorInvalid) {
		t.Fatalf("expired cursor err = %v", err)
	}
}

func TestTraverseDeterminism(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-traverse")
	store := NewStore(db)
	ctx := context.Background()
	input := retargetGraphInput(t, incidentGraphInput(t), "traverse_graph")
	relationships := input["source_relationships"].([]any)
	relationships = append(relationships,
		map[string]any{"source_relationship_id": "logon_self", "source_relationship_kind": "logon", "src_source_entity_id": "host1", "dst_source_entity_id": "host1", "direction": "forward", "properties": map[string]any{"site": "hq", "src_site": "hq", "dst_site": "hq"}},
		map[string]any{"source_relationship_id": "logon2", "source_relationship_kind": "logon", "src_source_entity_id": "host1", "dst_source_entity_id": "host2", "direction": "forward", "properties": map[string]any{"site": "hq", "src_site": "hq", "dst_site": "hq"}},
	)
	input["source_relationships"] = relationships
	run, err := store.CreateProjection(ctx, mustJSON(t, input), StoreProjectionOptions{
		ProjectionRunNonce: "traverse-run",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime(),
	})
	if err != nil {
		t.Fatalf("create projection: %v", err)
	}
	var seed string
	for _, vertex := range run.GraphView.Vertices {
		if vertex.VertexFamily == "direct" {
			seed = vertex.VertexID
			break
		}
	}
	if seed == "" {
		t.Fatal("missing direct seed vertex")
	}
	traversal, err := store.Traverse(ctx, TraverseRequest{
		GraphViewID:   run.GraphViewID,
		SeedVertexIDs: []string{seed},
		MaxDepth:      2,
		Direction:     "out",
		EdgeKinds:     []string{"logon_edge"},
	})
	if err != nil {
		t.Fatalf("traverse: %v", err)
	}
	if len(traversal.Vertices) < 2 || len(traversal.Edges) == 0 {
		t.Fatalf("unexpected traversal result: %#v", traversal)
	}
	var selfLoopFound bool
	seenPairEdges := map[string]int{}
	for i, edge := range traversal.Edges {
		if i > 0 && traversal.Edges[i-1].SortKey > edge.SortKey {
			t.Fatalf("edges are not sorted: previous=%s current=%s", traversal.Edges[i-1].SortKey, edge.SortKey)
		}
		if edge.SrcVertexID == edge.DstVertexID {
			selfLoopFound = true
		}
		seenPairEdges[edge.SrcVertexID+"->"+edge.DstVertexID]++
	}
	if !selfLoopFound {
		t.Fatal("expected self-loop in traversal")
	}
	var multiEdgeFound bool
	for _, count := range seenPairEdges {
		if count > 1 {
			multiEdgeFound = true
			break
		}
	}
	if !multiEdgeFound {
		t.Fatalf("expected multi-edge traversal, got counts %#v", seenPairEdges)
	}
	if len(traversal.Metadata) != 0 {
		t.Fatalf("traversal metadata = %#v", traversal.Metadata)
	}

	emptyTraversal, err := store.Traverse(ctx, TraverseRequest{
		GraphViewID:   run.GraphViewID,
		SeedVertexIDs: []string{"missing_vertex"},
		MaxDepth:      1,
		Direction:     "both",
	})
	if err != nil {
		t.Fatalf("traverse unknown seed: %v", err)
	}
	if len(emptyTraversal.Vertices) != 0 || len(emptyTraversal.Edges) != 0 {
		t.Fatalf("unknown seed should produce empty traversal: %#v", emptyTraversal)
	}
}

func retargetGraphInput(t *testing.T, input map[string]any, key string) map[string]any {
	t.Helper()
	graphViewID, err := DeriveGraphViewID(key)
	if err != nil {
		t.Fatalf("derive graph view id for %q: %v", key, err)
	}
	config := input["projection_config"].(map[string]any)
	config["graph_view_key"] = key
	input["graph_view_id"] = graphViewID
	input["source_snapshot_id"] = "snap_" + key
	return input
}
