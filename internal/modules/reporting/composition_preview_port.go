package reporting

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type compositionPreviewJobPort struct {
	runner       reportingJobRunner
	transactions reportingJobAdmission
}

func NewCompositionPreviewJobPort(runner reportingJobRunner, transactions reportingJobAdmission) (reportcomposition.PreviewJobPort, error) {
	if runner == nil || transactions == nil {
		return nil, errors.New("reporting composition preview port requires job runner and transaction service")
	}
	return compositionPreviewJobPort{runner: runner, transactions: transactions}, nil
}

func (port compositionPreviewJobPort) AdmitPreviewJob(
	ctx context.Context,
	tx pgx.Tx,
	request reportcomposition.PreviewJobAdmission,
) (jobs.Resource, error) {
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		jobs.NewRouteIdempotencyKey(request.IdempotencyKey.RouteKey, request.IdempotencyKey.ActorUserID, request.IdempotencyKey.ScopeKey, request.IdempotencyKey.ClientTxnID),
		request.Scope,
		request.Normalized,
	)
	if err != nil {
		return jobs.Resource{}, err
	}
	return port.transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{
		JobKind:           CompositionPreviewJobKind,
		Scope:             request.Scope,
		SubmittedByUserID: request.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		Extension:         admission,
	}, request.Now.UTC())
}

func (port compositionPreviewJobPort) DispatchPreviewJob(jobID string) error {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}
	port.runner.Notify(parsed)
	return nil
}
