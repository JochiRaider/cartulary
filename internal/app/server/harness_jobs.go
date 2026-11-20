package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// harnessJobPorts gives a contributed harness job kind the Jobs capabilities
// it needs (TH-HARNESS-REQ-450): a deployment-scoped enqueue that commits and
// wakes the runner, and the handler payload of a running attempt.
func harnessJobPorts(pool *pgxpool.Pool, transactions *jobs.TransactionService, manager *jobs.Manager, runner *jobs.Runner, jobKind string, now func() time.Time) jobs.ContributionPorts {
	return jobs.ContributionPorts{
		Enqueue: func(ctx context.Context, submittedBy uuid.UUID, payload json.RawMessage) (uuid.UUID, error) {
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return uuid.Nil, fmt.Errorf("begin harness job enqueue: %w", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			job, err := transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{
				JobKind:           jobKind,
				Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
				SubmittedByUserID: submittedBy,
				Cancelable:        true,
				HandlerPayload:    payload,
			}, now().UTC())
			if err != nil {
				return uuid.Nil, err
			}
			if err := tx.Commit(ctx); err != nil {
				return uuid.Nil, err
			}
			jobID, err := uuid.Parse(job.JobID)
			if err != nil {
				return uuid.Nil, err
			}
			runner.Notify(jobID)
			return jobID, nil
		},
		Payload: manager.HandlerPayload,
	}
}
