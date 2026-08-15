package postgresstore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) expireRunIfNeeded(ctx context.Context, projectionRunID string, queryReceivedAt time.Time) (bool, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM graph_projection_runs WHERE projection_run_id = $1 AND retention_expires_at IS NOT NULL AND $2::timestamptz >= retention_expires_at`, projectionRunID, queryReceivedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func expireRunIfNeededInGraphTx(ctx context.Context, tx pgx.Tx, graphViewID, projectionRunID string, queryReceivedAt time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `DELETE FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2 AND retention_expires_at IS NOT NULL AND $3::timestamptz >= retention_expires_at`, graphViewID, projectionRunID, queryReceivedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) pruneInvalidatedRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID, selectedRunID string, policy graphprojection.RetentionPolicy, queryReceivedAt time.Time) error {
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY invalidated_at DESC, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'invalidated' AND projection_run_id <> $2
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND (
       $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.invalidated_at + make_interval(secs => $4)
   )
`, graphViewID, selectedRunID, policy.RetentionCount, policy.RetentionDurationSeconds, queryReceivedAt); err != nil {
		return fmt.Errorf("prune invalidated graph projection runs: %w", err)
	}
	return nil
}

func (s *Store) pruneRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID string, policy graphprojection.RetentionPolicy, queryReceivedAt time.Time) error {
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY replaced_at DESC NULLS LAST, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'replaced'
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND (
       $2::boolean = false
       OR $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.replaced_at + make_interval(secs => $4)
   )
`, graphViewID, policy.RetainReplacedResults, policy.RetentionCount, policy.RetentionDurationSeconds, queryReceivedAt); err != nil {
		return fmt.Errorf("prune replaced graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY completed_at DESC NULLS LAST, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'failed'
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND run.projection_run_id <> COALESCE((SELECT latest_projection_run_id FROM graph_projection_views WHERE graph_view_id = $1 AND state = 'failed'), '')
   AND (
       $2::boolean = false
       OR $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.completed_at + make_interval(secs => $4)
   )
`, graphViewID, policy.RetainFailedResults, policy.FailedRetentionCount, policy.FailedRetentionDurationSecs, queryReceivedAt); err != nil {
		return fmt.Errorf("prune failed graph projection runs: %w", err)
	}
	return nil
}
