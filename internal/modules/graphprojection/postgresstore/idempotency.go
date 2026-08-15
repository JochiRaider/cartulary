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

func invalidationReplayFingerprint(operation, graphViewID, projectionRunID, reasonCode, requestedBy string) (string, error) {
	comparison := map[string]any{"operation": operation, "graph_view_id": graphViewID, "reason_code": reasonCode, "requested_by": requestedBy}
	if projectionRunID != "" {
		comparison["projection_run_id"] = projectionRunID
	}
	bytes, err := graphprojection.CanonicalJSON(comparison)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Store) checkInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (graphprojection.InvalidationSummary, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, response_json, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var responseJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &responseJSON, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, false, nil
		}
		return graphprojection.InvalidationSummary{}, false, fmt.Errorf("load graph projection invalidation idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return graphprojection.InvalidationSummary{}, false, err
		}
		return graphprojection.InvalidationSummary{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return graphprojection.InvalidationSummary{}, false, &graphprojection.LifecycleError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	var summary graphprojection.InvalidationSummary
	if err := json.Unmarshal(responseJSON, &summary); err != nil {
		return graphprojection.InvalidationSummary{}, false, fmt.Errorf("decode graph projection invalidation replay: %w", err)
	}
	return summary, true, nil
}

func (s *Store) recordInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, summary graphprojection.InvalidationSummary, createdAt time.Time) error {
	responseJSON, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	var targetProjectionRunID any
	if summary.TargetProjectionRunID != nil {
		targetProjectionRunID = *summary.TargetProjectionRunID
	}
	_, err = tx.Exec(ctx, `
INSERT INTO graph_projection_idempotency (
    operation,
    scope_key,
    idempotency_key,
    request_fingerprint,
    graph_view_id,
    projection_run_id,
    response_json,
    created_at,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
`, operation, scopeKey, key, fingerprint, summary.GraphViewID, targetProjectionRunID, string(responseJSON), createdAt, createdAt.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("record graph projection invalidation idempotency: %w", err)
	}
	return nil
}

func (s *Store) checkIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (graphprojection.ProjectionRun, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, response_json, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var responseJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &responseJSON, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.ProjectionRun{}, false, nil
		}
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("load graph projection idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return graphprojection.ProjectionRun{}, false, err
		}
		return graphprojection.ProjectionRun{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return graphprojection.ProjectionRun{}, false, &graphprojection.LifecycleError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	var response graphprojection.AcceptedRunSummary
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("decode graph projection accepted replay: %w", err)
	}
	acceptedAt, err := time.Parse("2006-01-02T15:04:05.999999Z", response.AcceptedAt)
	if err != nil {
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("decode graph projection accepted replay timestamp: %w", err)
	}
	run := graphprojection.ProjectionRun{GraphViewID: response.GraphViewID, ProjectionRunID: response.ProjectionRunID, State: graphprojection.RunStateAccepted, AcceptedAt: acceptedAt, AcceptedReplay: &response}
	run.Request.SourceSnapshotID = response.SourceSnapshotID
	run.Request.ProjectionConfig.ProjectionVersion = response.ProjectionVersion
	return run, true, nil
}

func retainedReplayFingerprint(operation string, run graphprojection.ProjectionRun) (string, error) {
	bytes, err := graphprojection.CanonicalJSON(map[string]any{
		"operation":                operation,
		"graph_view_id":            run.GraphViewID,
		"projection_config_digest": run.ProjectionConfigDigest,
		"projection_source_digest": run.ProjectionSourceDigest,
	})
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Store) recordIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, run graphprojection.ProjectionRun) error {
	expiresAt := graphprojection.FormatLifecycleTimestamp(run.AcceptedAt.Add(24 * time.Hour))
	responseJSON, err := json.Marshal(graphprojection.AcceptedRunSummary{
		GraphViewID:          run.GraphViewID,
		ProjectionRunID:      run.ProjectionRunID,
		State:                graphprojection.RunStateAccepted,
		SourceSnapshotID:     run.Request.SourceSnapshotID,
		ProjectionVersion:    run.Request.ProjectionConfig.ProjectionVersion,
		AcceptedAt:           graphprojection.FormatLifecycleTimestamp(run.AcceptedAt),
		IdempotencyExpiresAt: &expiresAt,
	})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO graph_projection_idempotency (
    operation,
	scope_key,
    idempotency_key,
    request_fingerprint,
    graph_view_id,
    projection_run_id,
    response_json,
    created_at,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
`, operation, scopeKey, key, fingerprint, run.GraphViewID, run.ProjectionRunID, string(responseJSON), run.AcceptedAt, run.AcceptedAt.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("record graph projection idempotency: %w", err)
	}
	return nil
}
