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

func (s *Store) InvalidateGraphView(ctx context.Context, request graphprojection.RetainedInvalidation) (graphprojection.InvalidationSummary, error) {
	return s.invalidateGraphViewAt(ctx, request.GraphViewID, request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey, request.InvalidatedAt)
}

func (s *Store) invalidateGraphViewAt(ctx context.Context, graphViewID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (graphprojection.InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("begin graph projection invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockGraphViewTx(ctx, tx, graphViewID); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	fingerprint, err := invalidationReplayFingerprint("invalidate_graph_view", graphViewID, "", reasonCode, requestedBy)
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	var selectedRunID string
	var currentViewState string
	var retentionPolicyJSON []byte
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT graph_view.state,
       COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id),
       run.retention_policy_json,
       graph_view.updated_at
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS run ON run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, run
`, graphViewID).Scan(&currentViewState, &selectedRunID, &retentionPolicyJSON, &currentUpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, graphprojection.ErrGraphViewNotFound
		}
		return graphprojection.InvalidationSummary{}, err
	}
	if currentViewState != string(graphprojection.GraphViewStateAvailable) && currentViewState != string(graphprojection.GraphViewStateRefreshing) {
		return graphprojection.InvalidationSummary{}, &graphprojection.LifecycleError{Code: "invalid_operation", ReasonCode: "invalid_invalidation_target", Details: map[string]any{"operation": "invalidate_graph_view", "reason_code": "invalid_invalidation_target"}}
	}
	var retentionPolicy graphprojection.RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
	if !invalidatedAt.After(currentUpdatedAt) {
		invalidatedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	rows, err := tx.Query(ctx, `SELECT projection_run_id FROM graph_projection_runs WHERE graph_view_id = $1 AND state IN ('available', 'replaced') ORDER BY projection_run_id ASC`, graphViewID)
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("list invalidation targets: %w", err)
	}
	runIDs := []string{}
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			rows.Close()
			return graphprojection.InvalidationSummary{}, err
		}
		runIDs = append(runIDs, runID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if len(runIDs) == 0 {
		return graphprojection.InvalidationSummary{}, &graphprojection.LifecycleError{Code: "invalid_operation", ReasonCode: "invalid_invalidation_target", Details: map[string]any{"operation": "invalidate_graph_view", "reason_code": "invalid_invalidation_target"}}
	}
	summary := graphprojection.InvalidationSummary{GraphViewID: graphViewID, TargetScope: "graph_view", InvalidatedRunIDs: runIDs, GraphViewStateAfter: graphprojection.GraphViewStateInvalidated, InvalidatedAt: graphprojection.FormatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := graphprojection.FormatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	invalidationJSON, _ := json.Marshal(graphprojection.Invalidation{InvalidatedAt: summary.InvalidatedAt, ReasonCode: reasonCode, RequestedBy: requestedBy, TargetScope: "graph_view"})
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', invalidated_at = $2::timestamptz, retention_expires_at = CASE WHEN projection_run_id = $4 THEN NULL ELSE $2::timestamptz + make_interval(secs => $5::integer) END, invalidation_json = $3::jsonb WHERE graph_view_id = $1 AND state IN ('available', 'replaced')`, graphViewID, invalidatedAt, string(invalidationJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("invalidate graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(invalidationJSON)); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("invalidate graph projection view: %w", err)
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("commit graph projection invalidation: %w", err)
	}
	return summary, nil
}

func (s *Store) InvalidateProjectionRun(ctx context.Context, request graphprojection.RetainedInvalidation) (graphprojection.InvalidationSummary, error) {
	return s.invalidateProjectionRunAt(ctx, request.GraphViewID, request.ProjectionRunID, request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey, request.InvalidatedAt)
}

func (s *Store) invalidateProjectionRunAt(ctx context.Context, graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (graphprojection.InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("begin graph projection run invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockGraphViewTx(ctx, tx, graphViewID); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	fingerprint, err := invalidationReplayFingerprint("invalidate_projection_run", graphViewID, projectionRunID, reasonCode, requestedBy)
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		scopeKey := graphViewID + "\n" + projectionRunID
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", scopeKey, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	if expired, err := expireRunIfNeededInGraphTx(ctx, tx, graphViewID, projectionRunID, invalidatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("expire graph projection invalidation target: %w", err)
	} else if expired {
		return graphprojection.InvalidationSummary{}, graphprojection.ErrProjectionRunNotFound
	}
	var state string
	if err := tx.QueryRow(ctx, `SELECT state FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2 FOR UPDATE`, graphViewID, projectionRunID).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, graphprojection.ErrProjectionRunNotFound
		}
		return graphprojection.InvalidationSummary{}, err
	}
	if state != string(graphprojection.RunStateAvailable) && state != string(graphprojection.RunStateReplaced) {
		return graphprojection.InvalidationSummary{}, &graphprojection.LifecycleError{Code: "invalid_operation", ReasonCode: "invalid_invalidation_target", Details: map[string]any{"operation": "invalidate_projection_run", "reason_code": "invalid_invalidation_target"}}
	}
	graphViewStateAfter := graphprojection.GraphViewStateAvailable
	var selectedRunID string
	var currentViewState string
	var retentionPolicyJSON []byte
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id, ''),
       graph_view.state,
       selected_run.retention_policy_json,
       graph_view.updated_at
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS selected_run
    ON selected_run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, selected_run
`, graphViewID).Scan(&selectedRunID, &currentViewState, &retentionPolicyJSON, &currentUpdatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	var retentionPolicy graphprojection.RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
	if !invalidatedAt.After(currentUpdatedAt) {
		invalidatedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	if currentViewState == string(graphprojection.GraphViewStateInvalidated) || selectedRunID == projectionRunID {
		graphViewStateAfter = graphprojection.GraphViewStateInvalidated
	}
	summary := graphprojection.InvalidationSummary{GraphViewID: graphViewID, TargetScope: "projection_run", TargetProjectionRunID: &projectionRunID, InvalidatedRunIDs: []string{projectionRunID}, GraphViewStateAfter: graphViewStateAfter, InvalidatedAt: graphprojection.FormatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := graphprojection.FormatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	invalidationJSON, err := json.Marshal(graphprojection.Invalidation{InvalidatedAt: summary.InvalidatedAt, ReasonCode: reasonCode, RequestedBy: requestedBy, TargetScope: "projection_run", TargetProjectionRunID: &projectionRunID})
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', completed_at = COALESCE(completed_at, $3), invalidated_at = $3, retention_expires_at = CASE WHEN $2 = $5 THEN NULL ELSE $3::timestamptz + make_interval(secs => $6::integer) END, invalidation_json = $4::jsonb WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID, invalidatedAt, string(invalidationJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if graphViewStateAfter == graphprojection.GraphViewStateInvalidated {
		if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(invalidationJSON)); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		scopeKey := graphViewID + "\n" + projectionRunID
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", scopeKey, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	return summary, nil
}
