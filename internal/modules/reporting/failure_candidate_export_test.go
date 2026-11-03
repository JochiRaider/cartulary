package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// External service fixtures use the shipped preparation/commit boundary without
// retaining a fixture-only publication method in the production Store API.
func PrepareRenderFailureForTest(ctx context.Context, pool *pgxpool.Pool, transactions *jobs.TransactionService, execution jobs.Execution, failure RenderFailure) (renderFailureCandidate, error) {
	store := &Store{pool: pool, jobTransactions: transactions}
	return store.prepareRenderFailureCandidate(ctx, execution, failure)
}

func CommitRenderFailureForTest(ctx context.Context, tx pgx.Tx, candidate renderFailureCandidate, now time.Time) (*uuid.UUID, error) {
	return candidate.owner.completeRenderFailureTx(ctx, tx, candidate.jobID, candidate, now)
}

// Replay publication reads the actual composed Core contract. These fixtures
// deliberately provide no Jobs mutation port: that belongs to the finalizer.
func ReplayStoreForTest(reader jobs.RetainedJobReader) *Store {
	return &Store{jobTransactions: replayJobReads{RetainedJobReader: reader}}
}

type replayJobReads struct{ jobs.RetainedJobReader }

var errReplayJobMutation = errors.New("unexpected Jobs mutation during owner replay publication")

func (replayJobReads) CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error) {
	return jobs.Resource{}, errReplayJobMutation
}
func (replayJobReads) ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error {
	return errReplayJobMutation
}
func (replayJobReads) ExtensionFinalizationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error) {
	return jobs.ExtensionFinalizationContext{}, errReplayJobMutation
}
func (replayJobReads) ExtensionCancellationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error) {
	return jobs.ExtensionFinalizationContext{}, errReplayJobMutation
}
