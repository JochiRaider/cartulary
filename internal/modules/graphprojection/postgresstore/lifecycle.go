package postgresstore

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) CreateProjection(ctx context.Context, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	return s.projectRetained(ctx, "create_projection", data, options)
}

func (s *Store) RefreshProjection(ctx context.Context, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	return s.projectRetained(ctx, "refresh_projection", data, options)
}

func (s *Store) projectRetained(ctx context.Context, operation string, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	run, err := graphprojection.AdmitRetainedProjection(data, options.ProjectionRunNonce, options.AcceptedAt)
	if err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("begin graph projection admission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockGraphViewTx(ctx, tx, run.GraphViewID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}

	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		replayed, replay, err := s.checkIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run.AcceptedAt)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		if replay {
			return replayed, nil
		}
	}

	var currentViewState string
	var currentLatestRunID string
	var currentSelectedRunID string
	var currentUpdatedAt time.Time
	viewErr := tx.QueryRow(ctx, `SELECT state, COALESCE(latest_projection_run_id, ''), COALESCE(selected_projection_run_id, latest_projection_run_id, ''), updated_at FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, run.GraphViewID).Scan(&currentViewState, &currentLatestRunID, &currentSelectedRunID, &currentUpdatedAt)
	viewExists := viewErr == nil
	if viewErr != nil && !errors.Is(viewErr, pgx.ErrNoRows) {
		return graphprojection.ProjectionRun{}, fmt.Errorf("load graph projection lifecycle state: %w", viewErr)
	}
	if viewExists && !run.AcceptedAt.After(currentUpdatedAt) {
		run.AcceptedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	if operation == "create_projection" && viewExists {
		switch graphprojection.GraphViewState(currentViewState) {
		case graphprojection.GraphViewStateCreating, graphprojection.GraphViewStateRefreshing:
			return graphprojection.ProjectionRun{}, &graphprojection.LifecycleError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case graphprojection.GraphViewStateAvailable, graphprojection.GraphViewStateInvalidated:
			return graphprojection.ProjectionRun{}, &graphprojection.LifecycleError{Code: "invalid_operation", ReasonCode: "graph_view_already_exists", Details: map[string]any{"operation": operation, "reason_code": "graph_view_already_exists"}}
		}
	}
	if operation == "refresh_projection" {
		if !viewExists {
			return graphprojection.ProjectionRun{}, graphprojection.ErrGraphViewNotFound
		}
		switch graphprojection.GraphViewState(currentViewState) {
		case graphprojection.GraphViewStateCreating, graphprojection.GraphViewStateRefreshing:
			return graphprojection.ProjectionRun{}, &graphprojection.LifecycleError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case graphprojection.GraphViewStateFailed:
			return graphprojection.ProjectionRun{}, &graphprojection.LifecycleError{Code: "invalid_operation", ReasonCode: "no_consumable_prior_run", Details: map[string]any{"operation": operation, "reason_code": "no_consumable_prior_run"}}
		}
	}

	var previousRunID *string
	if operation == "refresh_projection" && currentSelectedRunID != "" {
		previousRunID = &currentSelectedRunID
		run.PreviousProjectionRunID = previousRunID
	}
	admissionViewState := graphprojection.GraphViewStateCreating
	if operation == "refresh_projection" {
		admissionViewState = graphprojection.GraphViewStateRefreshing
	}
	if err := s.persistAcceptedRunTx(ctx, tx, run, admissionViewState, currentSelectedRunID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		if err := s.recordIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run); err != nil {
			return graphprojection.ProjectionRun{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("commit graph projection admission: %w", err)
	}

	startedAt := options.GeneratedAt.UTC()
	if !startedAt.After(run.AcceptedAt) {
		startedAt = run.AcceptedAt.Add(time.Microsecond)
	}
	run.StartedAt = &startedAt
	run.State = graphprojection.RunStateComputing
	if tag, err := s.pool.Exec(ctx, `UPDATE graph_projection_runs SET state = 'computing', started_at = $2 WHERE projection_run_id = $1 AND state = 'accepted'`, run.ProjectionRunID, startedAt); err != nil {
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", startedAt.Add(time.Microsecond))
	} else if tag.RowsAffected() != 1 {
		return s.finalizeRetainedFailure(run, operation, "implementation_invariant_failed", startedAt.Add(time.Microsecond))
	}

	generatedAt := options.GeneratedAt.UTC()
	if !generatedAt.After(startedAt) {
		generatedAt = startedAt.Add(time.Microsecond)
	}
	result, err := graphprojection.ProjectAdmittedRetainedProjection(ctx, run, generatedAt, previousRunID)
	if err != nil {
		reason := "internal_exception"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			reason = "timeout"
		}
		return s.finalizeRetainedFailure(run, operation, reason, generatedAt.Add(time.Microsecond))
	}
	completedAt := generatedAt.Add(time.Microsecond)
	result.CompletedAt = &completedAt
	if s.hooks.BeforePublication != nil {
		if err := s.hooks.BeforePublication(ctx, result); err != nil {
			return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
		}
	}
	terminalTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
	}
	defer func() { _ = terminalTx.Rollback(ctx) }()
	if err := lockGraphViewTx(ctx, terminalTx, result.GraphViewID); err != nil {
		_ = terminalTx.Rollback(ctx)
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
	}
	var terminalViewState string
	var terminalSelectedRunID string
	var invalidationJSON []byte
	if err := terminalTx.QueryRow(ctx, `SELECT state, COALESCE(selected_projection_run_id, ''), COALESCE(invalidation_json, 'null'::jsonb) FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, result.GraphViewID).Scan(&terminalViewState, &terminalSelectedRunID, &invalidationJSON); err != nil {
		_ = terminalTx.Rollback(ctx)
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
	}
	wasInvalidated := graphprojection.GraphViewState(terminalViewState) == graphprojection.GraphViewStateInvalidated || string(invalidationJSON) != "null"
	if wasInvalidated && result.State == graphprojection.RunStateAvailable {
		result.State = graphprojection.RunStateInvalidated
		result.PreviousProjectionRunID = nil
		var invalidation graphprojection.Invalidation
		if string(invalidationJSON) != "null" && json.Unmarshal(invalidationJSON, &invalidation) == nil {
			result.Invalidation = &invalidation
			if timestamp, parseErr := time.Parse("2006-01-02T15:04:05.999999Z", invalidation.InvalidatedAt); parseErr == nil {
				result.InvalidatedAt = &timestamp
			}
		}
		result.GraphView = nil
		result.ProjectionOutputDigest = ""
	}
	publicationPriorState := graphprojection.GraphViewState(terminalViewState)
	if wasInvalidated {
		publicationPriorState = graphprojection.GraphViewStateInvalidated
	}
	if err := s.publishProjectionRunTx(ctx, terminalTx, result, operation, publicationPriorState, terminalSelectedRunID); err != nil {
		_ = terminalTx.Rollback(ctx)
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
	}
	queryReceivedAt := result.AcceptedAt
	if result.CompletedAt != nil {
		queryReceivedAt = *result.CompletedAt
	}
	if err := s.pruneRetentionTx(ctx, terminalTx, result.GraphViewID, result.Request.ProjectionConfig.RetentionPolicy, queryReceivedAt); err != nil {
		_ = terminalTx.Rollback(ctx)
		return s.finalizeRetainedFailure(run, operation, "dependency_unavailable", completedAt)
	}
	if err := terminalTx.Commit(ctx); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("commit graph projection publication: %w", err)
	}
	return result, nil
}

func (s *Store) finalizeRetainedFailure(run graphprojection.ProjectionRun, operation, reasonCode string, completedAt time.Time) (graphprojection.ProjectionRun, error) {
	finalizationCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	failed := graphprojection.FailAdmittedRetainedProjection(run, completedAt, reasonCode)
	tx, err := s.pool.BeginTx(finalizationCtx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("begin graph projection failure finalization: %w", err)
	}
	defer func() { _ = tx.Rollback(finalizationCtx) }()
	if err := lockGraphViewTx(finalizationCtx, tx, run.GraphViewID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	var priorState string
	var selectedRunID string
	if err := tx.QueryRow(finalizationCtx, `SELECT state, COALESCE(selected_projection_run_id, '') FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, run.GraphViewID).Scan(&priorState, &selectedRunID); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("load graph projection failure state: %w", err)
	}
	if err := s.publishProjectionRunTx(finalizationCtx, tx, failed, operation, graphprojection.GraphViewState(priorState), selectedRunID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	if err := s.pruneRetentionTx(finalizationCtx, tx, run.GraphViewID, run.Request.ProjectionConfig.RetentionPolicy, completedAt); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	if err := tx.Commit(finalizationCtx); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("commit graph projection failure finalization: %w", err)
	}
	return failed, nil
}

func (s *Store) persistAcceptedRunTx(ctx context.Context, tx pgx.Tx, run graphprojection.ProjectionRun, viewState graphprojection.GraphViewState, selectedRunID string) error {
	retentionPolicyJSON, err := json.Marshal(run.Request.ProjectionConfig.RetentionPolicy)
	if err != nil {
		return fmt.Errorf("marshal graph projection retention policy: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id,
    graph_view_key,
    state,
    latest_projection_run_id,
    selected_projection_run_id,
    latest_source_snapshot_id,
    projection_version,
    updated_at,
    validation_status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
ON CONFLICT (graph_view_id) DO UPDATE
SET graph_view_key = EXCLUDED.graph_view_key,
    state = EXCLUDED.state,
    latest_projection_run_id = EXCLUDED.latest_projection_run_id,
    selected_projection_run_id = EXCLUDED.selected_projection_run_id,
    latest_source_snapshot_id = EXCLUDED.latest_source_snapshot_id,
    projection_version = EXCLUDED.projection_version,
    updated_at = EXCLUDED.updated_at,
    validation_status = EXCLUDED.validation_status
`, run.GraphViewID, run.Request.ProjectionConfig.GraphViewKey, string(viewState), run.ProjectionRunID, nullString(selectedRunID), run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, run.AcceptedAt); err != nil {
		return fmt.Errorf("upsert accepted graph projection view: %w", err)
	}
	if _, err := tx.Exec(ctx, `
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
    validation_summary_json,
    retention_policy_json
) VALUES ($1, $2, $3, $4, 'accepted', $5, $6, $7, $8, 'null'::jsonb, $9::jsonb)
`, run.ProjectionRunID, run.GraphViewID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, run.ProjectionRunNonce, run.ProjectionConfigDigest, run.ProjectionSourceDigest, run.AcceptedAt, string(retentionPolicyJSON)); err != nil {
		return fmt.Errorf("insert accepted graph projection run: %w", err)
	}
	return nil
}

func lockGraphViewTx(ctx context.Context, tx pgx.Tx, graphViewID string) error {
	digest := sha256.Sum256([]byte("cartulary.graph_projection.graph_view_lock.v1\n" + graphViewID))
	lockID := int64(binary.BigEndian.Uint64(digest[:8]))
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID); err != nil {
		return fmt.Errorf("lock graph projection view %s: %w", graphViewID, err)
	}
	return nil
}
