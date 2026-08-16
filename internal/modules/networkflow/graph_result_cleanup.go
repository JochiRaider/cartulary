package networkflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	graphResultCleanupExpiredLeaseLimit = 1000
	graphResultCleanupMaximumResults    = 8
	graphResultCleanupMaximumDuration   = 30 * time.Second
)

type graphResultCleanupSweepResult struct {
	DeletedLeases           int
	Examined                int
	DeletedResults          int
	HasMore                 bool
	Exhausted               bool
	NextCursor              *graphprojection.ResultCleanupCandidateV2
	HealthSnapshotValid     bool
	EligibleResultBacklog   int64
	OldestEligibleResultAge *time.Duration
}

type graphResultCleanupSweeper interface {
	SweepGraphResults(
		context.Context,
		time.Time,
		*graphprojection.ResultCleanupCandidateV2,
	) (graphResultCleanupSweepResult, error)
}

type graphResultCleanupService struct {
	db              postgres.DB
	declarations    *Store
	maximumResults  int
	maximumDuration time.Duration
}

func newGraphResultCleanupService(db postgres.DB, declarations *Store) (*graphResultCleanupService, error) {
	if db == nil || declarations == nil {
		return nil, errors.New("compose Network Flow graph-result cleanup: persistence is required")
	}
	return &graphResultCleanupService{
		db:              db,
		declarations:    declarations,
		maximumResults:  graphResultCleanupMaximumResults,
		maximumDuration: graphResultCleanupMaximumDuration,
	}, nil
}

func (service *graphResultCleanupService) SweepGraphResults(
	ctx context.Context,
	observedAt time.Time,
	after *graphprojection.ResultCleanupCandidateV2,
) (graphResultCleanupSweepResult, error) {
	if service == nil || service.db == nil || service.declarations == nil || observedAt.IsZero() ||
		service.maximumResults < 1 || service.maximumResults > graphResultCleanupMaximumResults || service.maximumDuration <= 0 {
		return graphResultCleanupSweepResult{}, errors.New("network flow graph-result cleanup is not configured")
	}
	observedAt = observedAt.UTC()
	deletedLeases, leasesHaveMore, err := service.deleteExpiredLeases(ctx, observedAt)
	result := graphResultCleanupSweepResult{DeletedLeases: deletedLeases, HasMore: leasesHaveMore}
	if err != nil {
		return result, err
	}
	started := time.Now()
	cursor := cloneCleanupCandidate(after)
	for result.Examined < service.maximumResults {
		if time.Since(started) >= service.maximumDuration {
			result.HasMore = true
			result.NextCursor = cursor
			return service.withCleanupHealth(ctx, observedAt, result), nil
		}
		candidate, deleted, err := service.processNextCandidate(ctx, observedAt, cursor)
		if err != nil {
			result.NextCursor = cursor
			return result, err
		}
		if candidate == nil {
			result.Exhausted = true
			result.NextCursor = cursor
			return service.withCleanupHealth(ctx, observedAt, result), nil
		}
		cursor = candidate
		result.NextCursor = cloneCleanupCandidate(candidate)
		result.Examined++
		if deleted {
			result.DeletedResults++
		}
	}
	result.HasMore = true
	return service.withCleanupHealth(ctx, observedAt, result), nil
}

func (service *graphResultCleanupService) withCleanupHealth(
	ctx context.Context,
	observedAt time.Time,
	result graphResultCleanupSweepResult,
) graphResultCleanupSweepResult {
	backlog, oldestAge, err := service.cleanupHealthSnapshot(ctx, observedAt)
	if err != nil {
		return result
	}
	result.HealthSnapshotValid = true
	result.EligibleResultBacklog = backlog
	result.OldestEligibleResultAge = oldestAge
	return result
}

func (service *graphResultCleanupService) cleanupHealthSnapshot(
	ctx context.Context,
	observedAt time.Time,
) (int64, *time.Duration, error) {
	var backlog int64
	var oldestAgeSeconds float64
	if err := service.db.QueryRow(ctx, `
SELECT COUNT(*)::bigint,
       COALESCE(MAX(EXTRACT(EPOCH FROM ($2 - result.published_at))), 0)
  FROM graph_projection_results result
 WHERE result.source_owner_id = $1
   AND NOT EXISTS (
       SELECT 1
         FROM network_flow_graph_views declaration
        WHERE declaration.selected_projection_result_id = result.projection_result_id
   )
   AND NOT EXISTS (
       SELECT 1
         FROM graph_projection_result_leases lease
        WHERE lease.projection_result_id = result.projection_result_id
   )
`, ProfileID, observedAt.UTC()).Scan(&backlog, &oldestAgeSeconds); err != nil {
		return 0, nil, fmt.Errorf("load Network Flow graph-result cleanup health: %w", err)
	}
	if backlog < 0 || oldestAgeSeconds < 0 {
		return 0, nil, errors.New("network flow graph-result cleanup health violates persisted time invariants")
	}
	if backlog == 0 {
		return 0, nil, nil
	}
	age := time.Duration(oldestAgeSeconds * float64(time.Second))
	return backlog, &age, nil
}

func (service *graphResultCleanupService) deleteExpiredLeases(ctx context.Context, observedAt time.Time) (int, bool, error) {
	tx, err := service.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, false, fmt.Errorf("begin Network Flow expired graph-result lease cleanup: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	cleaner, err := postgresresult.NewCleaner(tx)
	if err != nil {
		return 0, false, err
	}
	deleted, hasMore, err := cleaner.DeleteExpiredLeases(ctx, observedAt, graphResultCleanupExpiredLeaseLimit)
	if err != nil {
		return 0, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, false, fmt.Errorf("commit Network Flow expired graph-result lease cleanup: %w", err)
	}
	return deleted, hasMore, nil
}

func (service *graphResultCleanupService) processNextCandidate(
	ctx context.Context,
	observedAt time.Time,
	after *graphprojection.ResultCleanupCandidateV2,
) (*graphprojection.ResultCleanupCandidateV2, bool, error) {
	tx, err := service.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("begin Network Flow graph-result cleanup candidate: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	cleaner, err := postgresresult.NewCleaner(tx)
	if err != nil {
		return nil, false, err
	}
	candidate, err := cleaner.LockCleanupCandidate(ctx, ProfileID, after)
	if err != nil {
		return nil, false, err
	}
	if candidate == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, false, fmt.Errorf("commit empty Network Flow graph-result cleanup candidate: %w", err)
		}
		return nil, false, nil
	}
	selected, err := service.declarations.LockGraphViewDeclarationsSelectingResultTx(ctx, tx, candidate.ProjectionResultID)
	if err != nil {
		return nil, false, err
	}
	leased, err := cleaner.HasUnexpiredLease(ctx, candidate.ProjectionResultID, observedAt)
	if err != nil {
		return nil, false, err
	}
	deleted := false
	if !selected && !leased {
		deleted, err = cleaner.DeleteLockedResult(ctx, candidate.ProjectionResultID)
		if err != nil {
			return nil, false, err
		}
		if !deleted {
			return nil, false, graphprojection.ErrResultV2IdentityConflict
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit Network Flow graph-result cleanup candidate: %w", err)
	}
	return cloneCleanupCandidate(candidate), deleted, nil
}

func cloneCleanupCandidate(candidate *graphprojection.ResultCleanupCandidateV2) *graphprojection.ResultCleanupCandidateV2 {
	if candidate == nil {
		return nil
	}
	cloned := *candidate
	return &cloned
}
