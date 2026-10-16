package reporting

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type compositionPreviewJobPort struct {
	runner       reportingJobRunner
	transactions reportingJobAdmission
	references   reference_data.RetainedBindingValidator
	graphs       *GraphSourceRegistry
	templates    *TemplatePackRegistry
}

func NewCompositionPreviewJobPort(runner reportingJobRunner, transactions reportingJobAdmission, references reference_data.RetainedBindingValidator, templates *TemplatePackRegistry, providers ...GraphSourceProvider) (reportcomposition.PreviewJobPort, error) {
	if runner == nil || transactions == nil || references == nil || templates == nil {
		return nil, errors.New("reporting composition preview port requires job runner and transaction service")
	}
	graphs, err := NewGraphSourceRegistry(providers...)
	if err != nil {
		return nil, err
	}
	return compositionPreviewJobPort{runner: runner, transactions: transactions, references: references, graphs: graphs, templates: templates}, nil
}

func (port compositionPreviewJobPort) AdmitPreviewJob(
	ctx context.Context,
	tx pgx.Tx,
	request reportcomposition.PreviewJobAdmission,
) (jobs.Resource, error) {
	normalized, apiErr := reportcomposition.DecodeNormalizedPreviewRequest(request.Normalized)
	if apiErr != nil || request.IdempotencyKey.RouteKey != "report_compositions.preview" || request.IdempotencyKey.ActorUserID != request.ActorUserID || request.IdempotencyKey.ScopeKey != request.Source.CompositionID.String() || request.IdempotencyKey.ClientTxnID != normalized.ClientTxnID {
		return jobs.Resource{}, errors.New("reporting: inconsistent preview idempotency")
	}
	if request.Source.CreatedByUserID != request.ActorUserID || request.Scope.Kind != jobs.ScopeKindIncident || request.Scope.IncidentID == nil || *request.Scope.IncidentID != request.Source.IncidentID || !request.Now.Equal(request.Source.CreatedAt) {
		return jobs.Resource{}, errors.New("reporting: inconsistent preview admission")
	}
	payload, err := buildPreviewAdmissionTx(ctx, tx, request.Source, request.Normalized, port.templates)
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := port.references.ValidateBindingTx(ctx, tx, payload.ReferencePacks); err != nil {
		return jobs.Resource{}, err
	}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		jobs.NewRouteIdempotencyKey(request.IdempotencyKey.RouteKey, request.IdempotencyKey.ActorUserID, request.IdempotencyKey.ScopeKey, request.IdempotencyKey.ClientTxnID),
		request.Scope,
		request.Normalized,
	)
	if err != nil {
		return jobs.Resource{}, err
	}
	resource, err := port.transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{
		JobKind:           CompositionPreviewJobKind,
		Scope:             request.Scope,
		SubmittedByUserID: request.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		Extension:         admission,
	}, request.Now.UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	jobID, err := uuid.Parse(resource.JobID)
	if err != nil {
		return jobs.Resource{}, err
	}
	source := request.Source
	source.RenderAttemptID = jobID
	refs, err := decodeSourceProjectionRefs(payload.GraphProjectionRefs)
	if err != nil {
		return jobs.Resource{}, err
	}
	payload.GraphSources, err = port.graphs.FreezeTx(ctx, tx, source.IncidentID, jobID, refs, request.Now)
	if err != nil {
		var rejected *InvalidReleaseRequestError
		if errors.As(err, &rejected) {
			return jobs.Resource{}, reportcomposition.ErrPreviewInputRejected
		}
		return jobs.Resource{}, err
	}
	payload.GraphSourcesSHA256, err = graphSourcesDigest(payload.GraphSources)
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := insertRenderAdmissionTx(ctx, tx, jobID, reportingJobKindCompositionPreview, request.Now, payload); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil

}

func (port compositionPreviewJobPort) DispatchPreviewJob(jobID string) error {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}
	port.runner.Notify(parsed)
	return nil
}
