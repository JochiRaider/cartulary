package reference_data

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

type referenceJobAdmission interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
}

type referenceJobOperations interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
	UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error)
	CompleteFailed(context.Context, jobs.Execution, jobs.FailureCompletion) (jobs.Resource, error)
	CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error)
}

func failedCompletion(code string, details map[string]any) jobs.FailureCompletion {
	return jobs.FailureCompletion{
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ErrorSummary: jobs.ErrorSummary{
			Code:      code,
			Message:   code,
			Retryable: false,
			Details:   details,
		},
	}
}
