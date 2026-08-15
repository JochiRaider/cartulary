package postgresstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) GetProjectionRun(ctx context.Context, projectionRunID string) (graphprojection.ProjectionRun, error) {
	queryReceivedAt := s.now().UTC()
	if expired, err := s.expireRunIfNeeded(ctx, projectionRunID, queryReceivedAt); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("expire graph projection run: %w", err)
	} else if expired {
		return graphprojection.ProjectionRun{}, graphprojection.ErrProjectionRunNotFound
	}
	row := s.pool.QueryRow(ctx, `
SELECT graph_view_id,
       source_snapshot_id,
       projection_version,
       state,
       projection_run_nonce,
       projection_config_digest,
       projection_source_digest,
       projection_output_digest,
       accepted_at,
	   started_at,
	   generated_at,
       completed_at,
	   replaced_at,
	   invalidated_at,
       validation_summary_json,
       COALESCE(failure_reason, ''),
	       graph_view_json,
	   invalidation_json,
	       retention_expires_at
  FROM graph_projection_runs
 WHERE projection_run_id = $1
`, projectionRunID)
	var run graphprojection.ProjectionRun
	var state string
	var summaryJSON []byte
	var graphJSON []byte
	var invalidationJSON []byte
	var failureReason string
	var projectionOutputDigest *string
	var completedAt *time.Time
	var retentionExpiresAt *time.Time
	if err := row.Scan(&run.GraphViewID, &run.Request.SourceSnapshotID, &run.Request.ProjectionConfig.ProjectionVersion, &state, &run.ProjectionRunNonce, &run.ProjectionConfigDigest, &run.ProjectionSourceDigest, &projectionOutputDigest, &run.AcceptedAt, &run.StartedAt, &run.GeneratedAt, &completedAt, &run.ReplacedAt, &run.InvalidatedAt, &summaryJSON, &failureReason, &graphJSON, &invalidationJSON, &retentionExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.ProjectionRun{}, graphprojection.ErrProjectionRunNotFound
		}
		return graphprojection.ProjectionRun{}, fmt.Errorf("get graph projection run: %w", err)
	}
	run.ProjectionRunID = projectionRunID
	run.State = graphprojection.RunState(state)
	run.CompletedAt = completedAt
	if projectionOutputDigest != nil {
		run.ProjectionOutputDigest = *projectionOutputDigest
	}
	run.FailureReason = failureReason
	run.RetentionExpiresAt = retentionExpiresAt
	if len(summaryJSON) > 0 {
		if err := json.Unmarshal(summaryJSON, &run.ValidationSummary); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode validation summary: %w", err)
		}
	}
	if len(graphJSON) > 0 {
		var graphView graphprojection.GraphView
		if err := json.Unmarshal(graphJSON, &graphView); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode graph view: %w", err)
		}
		run.GraphView = &graphView
	}
	if len(invalidationJSON) > 0 {
		var invalidation graphprojection.Invalidation
		if err := json.Unmarshal(invalidationJSON, &invalidation); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode graph projection invalidation: %w", err)
		}
		run.Invalidation = &invalidation
	}
	return run, nil
}

func (s *Store) GetGraphView(ctx context.Context, graphViewID string, projectionRunID string) (graphprojection.GraphView, error) {
	resolvedRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.GraphView{}, err
	}
	run, err := s.GetProjectionRun(ctx, resolvedRunID)
	if err != nil {
		return graphprojection.GraphView{}, err
	}
	if run.GraphView == nil || (run.State != graphprojection.RunStateAvailable && run.State != graphprojection.RunStateReplaced) {
		return graphprojection.GraphView{}, graphprojection.ErrGraphViewUnavailable
	}
	if run.GraphViewID != graphViewID {
		return graphprojection.GraphView{}, graphprojection.ErrProjectionRunNotFound
	}
	return *run.GraphView, nil
}

func (s *Store) GetVertex(ctx context.Context, graphViewID, projectionRunID, vertexID string) (graphprojection.Vertex, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.Vertex{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT vertex_json FROM graph_projection_vertices WHERE graph_view_id = $1 AND projection_run_id = $2 AND vertex_id = $3`, graphViewID, projectionRunID, vertexID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.Vertex{}, graphprojection.NewQueryError("vertex_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "vertex_id": vertexID}, graphprojection.ErrVertexNotFound)
		}
		return graphprojection.Vertex{}, fmt.Errorf("get graph projection vertex: %w", err)
	}
	var vertex graphprojection.Vertex
	if err := json.Unmarshal(payload, &vertex); err != nil {
		return graphprojection.Vertex{}, fmt.Errorf("decode graph projection vertex: %w", err)
	}
	return vertex, nil
}

func (s *Store) GetEdge(ctx context.Context, graphViewID, projectionRunID, edgeID string) (graphprojection.Edge, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.Edge{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT edge_json FROM graph_projection_edges WHERE graph_view_id = $1 AND projection_run_id = $2 AND edge_id = $3`, graphViewID, projectionRunID, edgeID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.Edge{}, graphprojection.NewQueryError("edge_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "edge_id": edgeID}, graphprojection.ErrEdgeNotFound)
		}
		return graphprojection.Edge{}, fmt.Errorf("get graph projection edge: %w", err)
	}
	var edge graphprojection.Edge
	if err := json.Unmarshal(payload, &edge); err != nil {
		return graphprojection.Edge{}, fmt.Errorf("decode graph projection edge: %w", err)
	}
	return edge, nil
}

func (s *Store) resolveReadableRunID(ctx context.Context, graphViewID, projectionRunID string) (string, error) {
	suppliedRun := projectionRunID != ""
	if projectionRunID == "" {
		var viewState string
		if err := s.pool.QueryRow(ctx, `SELECT state, COALESCE(selected_projection_run_id, '') FROM graph_projection_views WHERE graph_view_id = $1`, graphViewID).Scan(&viewState, &projectionRunID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", graphprojection.NewQueryError("graph_view_not_found", "", map[string]any{"graph_view_id": graphViewID}, graphprojection.ErrGraphViewNotFound)
			}
			return "", fmt.Errorf("resolve selected graph projection run: %w", err)
		}
		if projectionRunID == "" || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateCreating || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateRefreshing || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateFailed || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateInvalidated {
			return "", graphprojection.NewQueryError("projection_not_available", viewState, map[string]any{"graph_view_id": graphViewID, "state": viewState}, graphprojection.ErrGraphViewUnavailable)
		}
	}
	if expired, err := s.expireRunIfNeeded(ctx, projectionRunID, s.now().UTC()); err != nil {
		return "", fmt.Errorf("expire selected graph projection run: %w", err)
	} else if expired {
		return "", graphprojection.NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrProjectionRunNotFound)
	}
	var state string
	var invalidationJSON []byte
	if err := s.pool.QueryRow(ctx, `SELECT state, COALESCE(invalidation_json, 'null'::jsonb) FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID).Scan(&state, &invalidationJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", graphprojection.NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrProjectionRunNotFound)
		}
		return "", fmt.Errorf("resolve graph projection run: %w", err)
	}
	switch graphprojection.RunState(state) {
	case graphprojection.RunStateAvailable, graphprojection.RunStateReplaced:
		return projectionRunID, nil
	case graphprojection.RunStateAccepted, graphprojection.RunStateComputing:
		return "", graphprojection.NewQueryError("projection_not_available", state, map[string]any{"graph_view_id": graphViewID, "state": state}, graphprojection.ErrGraphViewUnavailable)
	case graphprojection.RunStateFailed:
		return "", graphprojection.NewQueryError("projection_run_failed", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrGraphViewUnavailable)
	case graphprojection.RunStateInvalidated:
		if !suppliedRun {
			return "", graphprojection.NewQueryError("projection_not_available", string(graphprojection.GraphViewStateInvalidated), map[string]any{"graph_view_id": graphViewID, "state": graphprojection.GraphViewStateInvalidated}, graphprojection.ErrGraphViewUnavailable)
		}
		invalidationDetails := map[string]any{"reason_code": nil, "invalidated_at": nil}
		details := map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "invalidation": invalidationDetails}
		var invalidation graphprojection.Invalidation
		if string(invalidationJSON) != "null" && json.Unmarshal(invalidationJSON, &invalidation) == nil {
			invalidationDetails["reason_code"] = invalidation.ReasonCode
			invalidationDetails["invalidated_at"] = invalidation.InvalidatedAt
		}
		return "", graphprojection.NewQueryError("projection_run_invalidated", "", details, graphprojection.ErrGraphViewUnavailable)
	default:
		return "", graphprojection.NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrProjectionRunNotFound)
	}
}
