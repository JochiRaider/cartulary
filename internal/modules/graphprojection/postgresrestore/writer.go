package postgresrestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const truncateGraphProjectionTablesSQL = `TRUNCATE TABLE
  graph_projection_edges,
  graph_projection_idempotency,
  graph_projection_runs,
  graph_projection_vertices,
  graph_projection_views
RESTRICT`

const committedProofTimeout = 30 * time.Second

// Writer is the narrow, borrowed-Postgres publication adapter used only by
// Recovery. It deliberately does not embed or construct postgresstore.Store.
type Writer struct {
	db postgres.DB
}

var _ graphprojection.RestorePublisher = (*Writer)(nil)

func New(db postgres.DB) (*Writer, error) {
	if db == nil {
		return nil, graphprojection.NewRestoreError(graphprojection.RestoreErrorInvalidRequest)
	}
	return &Writer{db: db}, nil
}

func (writer *Writer) ReplaceAll(ctx context.Context, plan graphprojection.RestorePublicationPlan) (graphprojection.RestorePublicationProof, error) {
	if writer == nil || writer.db == nil || ctx == nil || ctx.Err() != nil || !validPlan(plan) {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Cause: context.Canceled}
	}
	tx, err := writer.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Cause: err}
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, `SET CONSTRAINTS graph_projection_views_selected_run_fkey DEFERRED`); err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	if _, err := tx.Exec(ctx, truncateGraphProjectionTablesSQL); err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	for _, staged := range plan.Projections {
		if err := insertProjection(ctx, tx, staged.Run); err != nil {
			return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
		}
	}
	if err := verifyPublishedState(ctx, tx, plan); err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Indeterminate: true, Cause: err}
	}
	proofContext, cancelProof := context.WithTimeout(context.WithoutCancel(ctx), committedProofTimeout)
	defer cancelProof()
	if err := verifyPublishedState(proofContext, writer.db, plan); err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Indeterminate: true, Cause: err}
	}
	return graphprojection.RestorePublicationProof{
		RebuiltViews:        append([]graphprojection.RestoreRebuiltView{}, plan.RebuiltViews...),
		PostconditionSHA256: plan.PostconditionSHA256,
	}, nil
}

func publicationError(ctx context.Context, cause error) error {
	return &graphprojection.RestorePublicationError{Indeterminate: ctx != nil && ctx.Err() != nil, Cause: cause}
}

func validPlan(plan graphprojection.RestorePublicationPlan) bool {
	if plan.RestoreOperationID.String() == "00000000-0000-0000-0000-000000000000" ||
		plan.TargetGenerationID.String() == "00000000-0000-0000-0000-000000000000" ||
		len(plan.ClearedTableIDs) != len(graphprojection.RestoreGraphTableIDs()) ||
		len(plan.Projections) != len(plan.RebuiltViews) || len(plan.PostconditionSHA256) != 64 {
		return false
	}
	for index, tableID := range graphprojection.RestoreGraphTableIDs() {
		if plan.ClearedTableIDs[index] != tableID {
			return false
		}
	}
	return true
}

func insertProjection(ctx context.Context, tx pgx.Tx, run graphprojection.ProjectionRun) error {
	if run.State != graphprojection.RunStateAvailable || run.GraphView == nil || run.GeneratedAt == nil || run.CompletedAt == nil {
		return fmt.Errorf("staged projection is not a terminal available run")
	}
	validationJSON, err := json.Marshal(run.ValidationSummary)
	if err != nil {
		return err
	}
	graphJSON, err := json.Marshal(run.GraphView)
	if err != nil {
		return err
	}
	retentionJSON, err := json.Marshal(run.Request.ProjectionConfig.RetentionPolicy)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id, graph_view_key, state, latest_projection_run_id,
    latest_source_snapshot_id, projection_version, selected_projection_run_id,
    updated_at, validation_status, invalidation_json
) VALUES ($1, $2, 'available', $3, $4, $5, $3, $6, 'passed', NULL)
`, run.GraphViewID, run.Request.ProjectionConfig.GraphViewKey, run.ProjectionRunID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, *run.CompletedAt); err != nil {
		return fmt.Errorf("insert Graph restore view: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_runs (
    projection_run_id, graph_view_id, source_snapshot_id, projection_version,
    state, projection_run_nonce, projection_config_digest, projection_source_digest,
    accepted_at, started_at, generated_at, completed_at, validation_summary_json,
    graph_view_json, retention_policy_json, projection_output_digest
) VALUES ($1, $2, $3, $4, 'available', $5, $6, $7, $8, $9, $10, $11, $12::jsonb, $13::jsonb, $14::jsonb, $15)
`, run.ProjectionRunID, run.GraphViewID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion,
		run.ProjectionRunNonce, run.ProjectionConfigDigest, run.ProjectionSourceDigest, run.AcceptedAt,
		run.StartedAt, run.GeneratedAt, run.CompletedAt, string(validationJSON), string(graphJSON), string(retentionJSON), run.ProjectionOutputDigest); err != nil {
		return fmt.Errorf("insert Graph restore run: %w", err)
	}
	for _, vertex := range run.GraphView.Vertices {
		body, err := json.Marshal(vertex)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_vertices (
    projection_run_id, graph_view_id, vertex_id, vertex_kind, sort_key, vertex_json
) VALUES ($1, $2, $3, $4, $5, $6::jsonb)
`, run.ProjectionRunID, run.GraphViewID, vertex.VertexID, vertex.VertexKind, vertex.SortKey, string(body)); err != nil {
			return fmt.Errorf("insert Graph restore vertex: %w", err)
		}
	}
	for _, edge := range run.GraphView.Edges {
		body, err := json.Marshal(edge)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_edges (
    projection_run_id, graph_view_id, edge_id, edge_kind, src_vertex_id,
    dst_vertex_id, direction, sort_key, edge_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
`, run.ProjectionRunID, run.GraphViewID, edge.EdgeID, edge.EdgeKind, edge.SrcVertexID, edge.DstVertexID, edge.Direction, edge.SortKey, string(body)); err != nil {
			return fmt.Errorf("insert Graph restore edge: %w", err)
		}
	}
	return nil
}

type restoreQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func verifyPublishedState(ctx context.Context, queryer restoreQueryer, plan graphprojection.RestorePublicationPlan) error {
	var views, runs, vertices, edges, idempotency, invalidRuns, invalidViews, invalidSelected int
	if err := queryer.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM graph_projection_views),
    (SELECT COUNT(*) FROM graph_projection_runs),
    (SELECT COUNT(*) FROM graph_projection_vertices),
    (SELECT COUNT(*) FROM graph_projection_edges),
    (SELECT COUNT(*) FROM graph_projection_idempotency),
    (SELECT COUNT(*) FROM graph_projection_runs WHERE state <> 'available' OR projection_output_digest IS NULL),
    (SELECT COUNT(*) FROM graph_projection_views WHERE state <> 'available' OR validation_status <> 'passed' OR invalidation_json IS NOT NULL),
    (SELECT COUNT(*) FROM graph_projection_views AS view_state
       LEFT JOIN graph_projection_runs AS selected
         ON selected.projection_run_id = view_state.selected_projection_run_id
        AND selected.graph_view_id = view_state.graph_view_id
      WHERE selected.projection_run_id IS NULL)
`).Scan(&views, &runs, &vertices, &edges, &idempotency, &invalidRuns, &invalidViews, &invalidSelected); err != nil {
		return fmt.Errorf("verify Graph restore aggregate state: %w", err)
	}
	wantVertices := 0
	wantEdges := 0
	storedOutputDigests := make(map[string]string, len(plan.Projections))
	for _, staged := range plan.Projections {
		if staged.Run.ProjectionRunID == "" || staged.Run.ProjectionOutputDigest == "" {
			return fmt.Errorf("graph restore staged projection proof is incomplete")
		}
		if _, duplicate := storedOutputDigests[staged.Run.ProjectionRunID]; duplicate {
			return fmt.Errorf("graph restore staged projection proof is duplicated")
		}
		storedOutputDigests[staged.Run.ProjectionRunID] = staged.Run.ProjectionOutputDigest
	}
	for _, rebuilt := range plan.RebuiltViews {
		wantVertices += rebuilt.VertexCount
		wantEdges += rebuilt.EdgeCount
	}
	if views != len(plan.RebuiltViews) || runs != len(plan.RebuiltViews) || vertices != wantVertices || edges != wantEdges ||
		idempotency != 0 || invalidRuns != 0 || invalidViews != 0 || invalidSelected != 0 {
		return fmt.Errorf("graph restore aggregate postcondition mismatch")
	}
	for _, rebuilt := range plan.RebuiltViews {
		var configDigest, sourceDigest, outputDigest, sourceSnapshotID, projectionVersion string
		var vertexCount, edgeCount int
		if err := queryer.QueryRow(ctx, `
SELECT run.projection_config_digest,
       run.projection_source_digest,
       run.projection_output_digest,
       run.source_snapshot_id,
       run.projection_version,
       (SELECT COUNT(*) FROM graph_projection_vertices WHERE projection_run_id = run.projection_run_id),
       (SELECT COUNT(*) FROM graph_projection_edges WHERE projection_run_id = run.projection_run_id)
  FROM graph_projection_runs AS run
  JOIN graph_projection_views AS view_state
    ON view_state.graph_view_id = run.graph_view_id
   AND view_state.selected_projection_run_id = run.projection_run_id
 WHERE run.graph_view_id = $1
   AND run.projection_run_id = $2
`, rebuilt.GraphViewID, rebuilt.ProjectionRunID).Scan(
			&configDigest, &sourceDigest, &outputDigest, &sourceSnapshotID, &projectionVersion, &vertexCount, &edgeCount,
		); err != nil {
			return fmt.Errorf("verify Graph restore projection: %w", err)
		}
		storedOutputDigest, ok := storedOutputDigests[rebuilt.ProjectionRunID]
		if !ok || configDigest != rebuilt.NormalizedConfigurationSHA256 || sourceDigest != rebuilt.NormalizedSourceSHA256 ||
			outputDigest != storedOutputDigest || sourceSnapshotID != rebuilt.SourceSnapshotID ||
			projectionVersion != rebuilt.ProjectionVersion || vertexCount != rebuilt.VertexCount || edgeCount != rebuilt.EdgeCount {
			return fmt.Errorf("graph restore projection postcondition mismatch")
		}
	}
	return nil
}
