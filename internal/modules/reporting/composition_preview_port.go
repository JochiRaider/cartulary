package reporting

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type compositionPreviewJobPort struct {
	runner *jobs.Runner
}

func NewCompositionPreviewJobPort(runner *jobs.Runner) (reportcomposition.PreviewJobPort, error) {
	if runner == nil {
		return nil, errors.New("reporting composition preview port requires job runner")
	}
	return compositionPreviewJobPort{runner: runner}, nil
}

func (port compositionPreviewJobPort) AdmitPreviewJob(
	ctx context.Context,
	tx pgx.Tx,
	request reportcomposition.PreviewJobAdmission,
) (jobs.Resource, error) {
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		CompositionPreviewJobKind,
		request.IdempotencyKey,
		request.Scope,
		request.Normalized,
	)
	if err != nil {
		return jobs.Resource{}, err
	}
	return jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             request.Scope,
		SubmittedByUserID: request.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       JobWorkerKind,
		Extension:         admission,
	}, request.Now.UTC())
}

func (port compositionPreviewJobPort) DispatchPreviewJob(jobID string) error {
	return port.runner.DispatchJob(JobWorkerKind, jobID)
}
