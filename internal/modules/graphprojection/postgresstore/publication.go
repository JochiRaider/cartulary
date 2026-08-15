package postgresstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) publishProjectionRunTx(ctx context.Context, tx pgx.Tx, run graphprojection.ProjectionRun, operation string, priorViewState graphprojection.GraphViewState, selectedRunID string) error {
	summaryJSON, err := json.Marshal(run.ValidationSummary)
	if err != nil {
		return fmt.Errorf("marshal validation summary: %w", err)
	}
	var graphJSON []byte
	if run.GraphView != nil {
		graphJSON, err = json.Marshal(run.GraphView)
		if err != nil {
			return fmt.Errorf("marshal graph view: %w", err)
		}
	}
	var runInvalidationJSON []byte
	if run.Invalidation != nil {
		runInvalidationJSON, err = json.Marshal(run.Invalidation)
		if err != nil {
			return fmt.Errorf("marshal graph projection invalidation: %w", err)
		}
	}
	var retentionExpiresAt any
	if run.State == graphprojection.RunStateFailed && operation == "refresh_projection" && run.CompletedAt != nil {
		retentionExpiresAt = run.CompletedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.FailedRetentionDurationSecs) * time.Second)
	}
	if tag, err := tx.Exec(ctx, `
UPDATE graph_projection_runs
   SET state = $2,
       projection_output_digest = $3,
       generated_at = $4,
       completed_at = $5,
       invalidated_at = $6,
       validation_summary_json = $7::jsonb,
       failure_reason = $8,
	       graph_view_json = $9::jsonb,
	   retention_expires_at = $10,
	   invalidation_json = $11::jsonb
 WHERE projection_run_id = $1
   AND state IN ('accepted', 'computing')
`, run.ProjectionRunID, string(run.State), nullString(run.ProjectionOutputDigest), run.GeneratedAt, run.CompletedAt, run.InvalidatedAt, string(summaryJSON), nullString(run.FailureReason), nullJSON(graphJSON), retentionExpiresAt, nullJSON(runInvalidationJSON)); err != nil {
		return fmt.Errorf("publish graph projection run: %w", err)
	} else if tag.RowsAffected() != 1 {
		return fmt.Errorf("publish graph projection run: expected 1 row, updated %d", tag.RowsAffected())
	}

	updatedAt := run.AcceptedAt
	if run.CompletedAt != nil {
		updatedAt = *run.CompletedAt
	}
	viewState := graphprojection.GraphViewStateFailed
	viewLatestRunID := run.ProjectionRunID
	viewSelectedRunID := ""
	validationStatus := run.ValidationSummary.Status
	if validationStatus == "" {
		validationStatus = "failed"
	}
	preserveInvalidation := false
	switch run.State {
	case graphprojection.RunStateAvailable:
		viewState = graphprojection.GraphViewStateAvailable
		viewSelectedRunID = run.ProjectionRunID
		if selectedRunID != "" && selectedRunID != run.ProjectionRunID {
			retentionExpiresAt := updatedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.RetentionDurationSeconds) * time.Second)
			if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'replaced', replaced_at = $2, retention_expires_at = $3 WHERE projection_run_id = $1 AND state = 'available'`, selectedRunID, updatedAt, retentionExpiresAt); err != nil {
				return fmt.Errorf("replace prior graph projection run: %w", err)
			}
		}
	case graphprojection.RunStateInvalidated:
		viewState = graphprojection.GraphViewStateInvalidated
		viewSelectedRunID = run.ProjectionRunID
		validationStatus = "invalidated"
		preserveInvalidation = true
	case graphprojection.RunStateFailed:
		if operation == "refresh_projection" && selectedRunID != "" {
			viewSelectedRunID = selectedRunID
			if priorViewState == graphprojection.GraphViewStateInvalidated {
				viewState = graphprojection.GraphViewStateInvalidated
				validationStatus = "invalidated"
				preserveInvalidation = true
			} else {
				viewState = graphprojection.GraphViewStateAvailable
				validationStatus = "passed"
			}
		}
	}
	if tag, err := tx.Exec(ctx, `
UPDATE graph_projection_views
   SET state = $2,
       latest_projection_run_id = $3,
       selected_projection_run_id = $4,
       latest_source_snapshot_id = CASE WHEN $4::text IS NULL THEN $5 ELSE (SELECT source_snapshot_id FROM graph_projection_runs WHERE projection_run_id = $4) END,
       projection_version = CASE WHEN $4::text IS NULL THEN $6 ELSE (SELECT projection_version FROM graph_projection_runs WHERE projection_run_id = $4) END,
       updated_at = $7,
       validation_status = $8,
       invalidation_json = CASE WHEN $9 THEN invalidation_json ELSE NULL END
 WHERE graph_view_id = $1
`, run.GraphViewID, string(viewState), viewLatestRunID, nullString(viewSelectedRunID), run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, updatedAt, validationStatus, preserveInvalidation); err != nil {
		return fmt.Errorf("publish graph projection view: %w", err)
	} else if tag.RowsAffected() != 1 {
		return fmt.Errorf("publish graph projection view: expected 1 row, updated %d", tag.RowsAffected())
	}

	if run.GraphView != nil {
		for _, vertex := range run.GraphView.Vertices {
			vertexJSON, err := json.Marshal(vertex)
			if err != nil {
				return fmt.Errorf("marshal graph vertex: %w", err)
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_vertices (
    projection_run_id,
    graph_view_id,
    vertex_id,
    vertex_kind,
    sort_key,
    vertex_json
) VALUES ($1, $2, $3, $4, $5, $6::jsonb)
`, run.ProjectionRunID, run.GraphViewID, vertex.VertexID, vertex.VertexKind, vertex.SortKey, string(vertexJSON)); err != nil {
				return fmt.Errorf("insert graph projection vertex: %w", err)
			}
		}
		for _, edge := range run.GraphView.Edges {
			edgeJSON, err := json.Marshal(edge)
			if err != nil {
				return fmt.Errorf("marshal graph edge: %w", err)
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_edges (
    projection_run_id,
    graph_view_id,
    edge_id,
    edge_kind,
    src_vertex_id,
    dst_vertex_id,
    direction,
    sort_key,
    edge_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
`, run.ProjectionRunID, run.GraphViewID, edge.EdgeID, edge.EdgeKind, edge.SrcVertexID, edge.DstVertexID, edge.Direction, edge.SortKey, string(edgeJSON)); err != nil {
				return fmt.Errorf("insert graph projection edge: %w", err)
			}
		}
	}
	return nil
}
