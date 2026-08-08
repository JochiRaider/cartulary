package reporting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	defaultReportingJobTimeout             = 2 * time.Minute
	reportingJobTerminalTransitionTimeout  = 5 * time.Second
	reportingJobTimeoutCode                = "reporting_job_timeout"
	reportingJobWorkerContextCanceledCode  = "reporting_job_context_canceled"
	reportingJobUnknownKindCode            = "internal_error"
	reportingJobInternalErrorCode          = "internal_error"
	reportingJobKindSnapshotCreate         = "snapshot_create"
	reportingJobKindReleaseCreate          = "release_create"
	reportingJobKindCompositionPreview     = "composition_preview"
	reportingJobRenderFailedCode           = "release_render_failed"
	reportingJobRenderFailedMessage        = "Release render failed."
	reportingJobSnapshotCreatedCode        = "snapshot_created"
	reportingJobSnapshotCreatedMessage     = "Snapshot created."
	reportingJobReleaseCreatedCode         = "release_created"
	reportingJobReleaseCreatedMessage      = "Release rendered."
	reportingJobPreviewCreatedCode         = "composition_preview_rendered"
	reportingJobPreviewCreatedMessage      = "Composition preview rendered."
	reportingJobTimeoutMessage             = "Reporting job timed out."
	reportingJobWorkerContextCanceledError = "reporting job worker context canceled"
)

type reportingJobDispatcher interface {
	Dispatch(jobID string)
}

type reportingJobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	RecoverHandler(context.Context, string) error
	DispatchJob(string, string) error
}

type reportingJobAdmission interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.CreateParams, time.Time) (jobs.Resource, error)
}

type durableReportingJobDispatcher struct {
	runner reportingJobRunner
}

func newDurableReportingJobDispatcher(runner reportingJobRunner) reportingJobDispatcher {
	return durableReportingJobDispatcher{runner: runner}
}

func (d durableReportingJobDispatcher) Dispatch(jobID string) {
	if d.runner == nil {
		return
	}
	_ = d.runner.DispatchJob(JobWorkerKind, jobID)
}

type reportingJobStore interface {
	ReportingJobKind(context.Context, uuid.UUID) (string, error)
	CompleteSnapshotCreateJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) error
	ReleasePayloadForJob(context.Context, uuid.UUID) (releaseCreateJobPayload, error)
	CompositionPreviewPayloadForJob(context.Context, uuid.UUID) (compositionPreviewJobPayload, error)
	CompleteReleaseRenderFailedJob(context.Context, uuid.UUID, RedactionProfile, string, string, time.Time) (uuid.UUID, error)
	CompleteReleaseCreateJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderedRelease, time.Time) error
	CompleteCompositionPreviewJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderedRelease, time.Time) error
}

type reportingJobManager interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	MarkRunning(context.Context, uuid.UUID, jobs.Progress, *string) (jobs.Resource, error)
	CompleteFailed(context.Context, jobs.TransitionParams) (jobs.Resource, error)
	CompleteCanceled(context.Context, jobs.TransitionParams) (jobs.Resource, error)
}

type releaseCandidateRenderer interface {
	Render(context.Context, releaseCreateJobPayload) (RenderedRelease, string, error)
}

type defaultReleaseCandidateRenderer struct{}

func (defaultReleaseCandidateRenderer) Render(ctx context.Context, payload releaseCreateJobPayload) (RenderedRelease, string, error) {
	if err := ctx.Err(); err != nil {
		return RenderedRelease{}, "", err
	}
	rendered, reasonCode, err := renderReleaseCandidate(payload.Request, payload.TemplateContract, payload.ExportModel, payload.ExportModelSHA256)
	if err != nil {
		return rendered, reasonCode, err
	}
	if err := ctx.Err(); err != nil {
		return RenderedRelease{}, "", err
	}
	return rendered, reasonCode, nil
}

type reportingJobWorker struct {
	store           reportingJobStore
	jobManager      reportingJobManager
	renderer        releaseCandidateRenderer
	jobFinalizer    JobSuccessFinalizer
	renderExport    RenderExportInvoker
	now             func() time.Time
	timeout         time.Duration
	terminalTimeout time.Duration
}

func newReportingJobWorker(
	store reportingJobStore,
	jobManager reportingJobManager,
	jobFinalizer JobSuccessFinalizer,
	renderExport RenderExportInvoker,
	now func() time.Time,
) *reportingJobWorker {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &reportingJobWorker{
		store:           store,
		jobManager:      jobManager,
		jobFinalizer:    jobFinalizer,
		renderExport:    renderExport,
		renderer:        defaultReleaseCandidateRenderer{},
		now:             now,
		timeout:         defaultReportingJobTimeout,
		terminalTimeout: reportingJobTerminalTransitionTimeout,
	}
}

func (w *reportingJobWorker) Run(parentCtx context.Context, jobID uuid.UUID) {
	if w == nil || w.store == nil || w.jobManager == nil || w.renderer == nil || w.jobFinalizer == nil || w.renderExport == nil {
		return
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	timeout := w.timeout
	if timeout <= 0 {
		timeout = defaultReportingJobTimeout
	}
	runCtx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	w.execute(runCtx, parentCtx, jobID)
}

func (w *reportingJobWorker) Handle(ctx context.Context, jobID uuid.UUID) error {
	w.Run(ctx, jobID)
	return nil
}

func (w *reportingJobWorker) execute(ctx context.Context, parentCtx context.Context, jobID uuid.UUID) {
	kind, err := w.store.ReportingJobKind(ctx, jobID)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, jobs.Progress{Completed: 0, Total: intPtr(1)}) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, jobs.Progress{Completed: 0, Total: intPtr(1)})
		return
	}
	switch kind {
	case reportingJobKindSnapshotCreate:
		w.executeSnapshotCreateJob(ctx, parentCtx, jobID)
	case reportingJobKindReleaseCreate:
		w.executeReleaseCreateJob(ctx, parentCtx, jobID)
	case reportingJobKindCompositionPreview:
		w.executeCompositionPreviewJob(ctx, parentCtx, jobID)
	default:
		w.failReportingJob(parentCtx, jobID, reportingJobUnknownKindCode, fmt.Errorf("unknown reporting job kind %q", kind), false, jobs.Progress{Completed: 0, Total: intPtr(1)})
	}
}

func (w *reportingJobWorker) executeCompositionPreviewJob(ctx context.Context, parentCtx context.Context, jobID uuid.UUID) {
	total := 1
	progress := jobs.Progress{Completed: 0, Total: &total}
	if !w.markReportingJobRunning(ctx, parentCtx, jobID, total) {
		return
	}
	if w.cancelReportingJobIfRequested(ctx, parentCtx, jobID, total) || w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	payload, err := w.store.CompositionPreviewPayloadForJob(ctx, jobID)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, progress)
		return
	}
	payload.Release, err = w.invokeRenderExport(ctx, payload.Release)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobRenderFailedCode, err, false, progress)
		return
	}
	rendered, reasonCode, err := w.renderer.Render(ctx, payload.Release)
	if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	if err != nil {
		w.failReportingJob(parentCtx, jobID, reportingJobRenderFailedCode,
			fmt.Errorf("%s: %w", reasonCode, err), false, progress)
		return
	}
	if w.cancelReportingJobIfRequested(ctx, parentCtx, jobID, total) || w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, _ = w.jobFinalizer.FinalizeReportingJobSuccess(ctxTerminal, JobSuccessFinalization{
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:         reportingJobPreviewCreatedCode,
				Message:      reportingJobPreviewCreatedMessage,
				ResourceRefs: []jobs.ResourceRef{},
			},
		},
		FinalCommitID: CompositionPreviewOperationKind + ":" + jobID.String(),
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return w.store.CompleteCompositionPreviewJobTx(
				ctx,
				tx,
				jobID,
				payload.PreviewAttemptID,
				rendered,
				w.now().UTC(),
			)
		},
	})
}

func (w *reportingJobWorker) executeSnapshotCreateJob(ctx context.Context, parentCtx context.Context, jobID uuid.UUID) {
	total := 1
	progress := jobs.Progress{Completed: 0, Total: &total}
	if !w.markReportingJobRunning(ctx, parentCtx, jobID, total) {
		return
	}
	if w.cancelReportingJobIfRequested(ctx, parentCtx, jobID, total) || w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	snapshotID := reportingResourceID(jobID, "snapshot")
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, _ = w.jobFinalizer.FinalizeReportingJobSuccess(ctxTerminal, JobSuccessFinalization{
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    reportingJobSnapshotCreatedCode,
				Message: reportingJobSnapshotCreatedMessage,
				ResourceRefs: []jobs.ResourceRef{{
					Kind:  "snapshot",
					ID:    snapshotID.String(),
					Route: "/api/v1/snapshots/" + snapshotID.String(),
				}},
			},
		},
		FinalCommitID: SnapshotCreateOperationKind + ":" + jobID.String(),
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return w.store.CompleteSnapshotCreateJobTx(ctx, tx, jobID, snapshotID)
		},
	})
}

func (w *reportingJobWorker) executeReleaseCreateJob(ctx context.Context, parentCtx context.Context, jobID uuid.UUID) {
	total := 1
	progress := jobs.Progress{Completed: 0, Total: &total}
	if !w.markReportingJobRunning(ctx, parentCtx, jobID, total) {
		return
	}
	if w.cancelReportingJobIfRequested(ctx, parentCtx, jobID, total) || w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	payload, err := w.store.ReleasePayloadForJob(ctx, jobID)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, progress)
		return
	}
	payload, err = w.invokeRenderExport(ctx, payload)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobRenderFailedCode, err, false, progress)
		return
	}
	rendered, reasonCode, renderErr := w.renderer.Render(ctx, payload)
	if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	if renderErr != nil {
		releaseID, err := w.store.CompleteReleaseRenderFailedJob(ctx, jobID, rendered.Profile, rendered.ProfileSHA256, reasonCode, w.now().UTC())
		if err != nil {
			w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, progress)
			return
		}
		ctxTerminal, cancel := w.terminalContext(parentCtx)
		defer cancel()
		_, _ = w.jobManager.CompleteFailed(ctxTerminal, jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ErrorSummary: &jobs.ErrorSummary{
				Code:      reportingJobRenderFailedCode,
				Message:   reportingJobRenderFailedMessage,
				Retryable: false,
				Details: map[string]any{
					"reason_code": reasonCode,
					"release_id":  releaseID.String(),
				},
			},
		})
		return
	}
	if w.cancelReportingJobIfRequested(ctx, parentCtx, jobID, total) || w.completeContextFailure(ctx, parentCtx, jobID, progress) {
		return
	}
	releaseID := reportingResourceID(jobID, "release")
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, _ = w.jobFinalizer.FinalizeReportingJobSuccess(ctxTerminal, JobSuccessFinalization{
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    reportingJobReleaseCreatedCode,
				Message: reportingJobReleaseCreatedMessage,
				ResourceRefs: []jobs.ResourceRef{{
					Kind:  "release",
					ID:    releaseID.String(),
					Route: "/api/v1/releases/" + releaseID.String(),
				}},
			},
		},
		FinalCommitID: ReleaseCreateOperationKind + ":" + jobID.String(),
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return w.store.CompleteReleaseCreateJobTx(ctx, tx, jobID, releaseID, rendered, w.now().UTC())
		},
	})
}

func (w *reportingJobWorker) invokeRenderExport(ctx context.Context, payload releaseCreateJobPayload) (releaseCreateJobPayload, error) {
	_, redactionSHA, err := ResolveRedactionProfile(
		payload.RedactionProfileID,
		payload.RedactionProfileVersion,
		payload.RecipientPartitionRefs,
	)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	authorizationView, err := canonicalJSON(map[string]any{
		"actor_user_id": payload.ActorUserID,
		"incident_id":   payload.IncidentID,
		"snapshot_id":   payload.SnapshotID,
	})
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	invocation := RenderExportInvocation{
		Context: RenderExportContext{
			SchemaID:                RenderExportContextSchemaID,
			Operation:               RenderExportOperationKind,
			ProfileID:               ProfileID,
			ContractMajor:           1,
			ClaimState:              "claimed",
			StatePresent:            false,
			StateVersion:            nil,
			SnapshotRef:             "snapshot:" + payload.SnapshotID,
			AuthorizationViewSHA256: hashHex(authorizationView),
			RedactionProfileSHA256:  redactionSHA,
			TimeoutSeconds:          renderExportTimeoutSeconds(w.timeout),
		},
		ImmutableModel:    payload.ExportModel,
		ImmutableModelSHA: payload.ExportModelSHA256,
	}
	result, err := w.renderExport.Invoke(ctx, invocation)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	model, digest, err := AdmitRenderExportResult(invocation, result)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	payload.ExportModel = model
	payload.ExportModelSHA256 = digest
	return payload, nil
}

func reportingResourceID(jobID uuid.UUID, kind string) uuid.UUID {
	return uuid.NewSHA1(jobID, []byte("snapshot_reporting:"+kind))
}

func (w *reportingJobWorker) markReportingJobRunning(ctx context.Context, parentCtx context.Context, jobID uuid.UUID, total int) bool {
	resource, err := w.jobManager.Get(ctx, jobID)
	if err != nil {
		w.completeContextFailure(ctx, parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total})
		return false
	}
	switch resource.Status {
	case jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled:
		return false
	case jobs.StatusRunning:
		return true
	case jobs.StatusCancelRequested:
		w.completeCanceled(parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total})
		return false
	}
	if _, err := w.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total}) {
			return false
		}
		resource, getErr := w.jobManager.Get(ctx, jobID)
		if getErr == nil && resource.Status == jobs.StatusCancelRequested {
			w.completeCanceled(parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total})
		}
		return false
	}
	return true
}

func (w *reportingJobWorker) cancelReportingJobIfRequested(ctx context.Context, parentCtx context.Context, jobID uuid.UUID, total int) bool {
	resource, err := w.jobManager.Get(ctx, jobID)
	if err != nil {
		w.completeContextFailure(ctx, parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total})
		return false
	}
	if resource.Status != jobs.StatusCancelRequested {
		return false
	}
	w.completeCanceled(parentCtx, jobID, jobs.Progress{Completed: 0, Total: &total})
	return true
}

func (w *reportingJobWorker) completeContextFailure(runCtx context.Context, terminalCtx context.Context, jobID uuid.UUID, progress jobs.Progress) bool {
	err := runCtx.Err()
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		w.failReportingJob(terminalCtx, jobID, reportingJobTimeoutCode, err, true, progress)
		return true
	case errors.Is(err, context.Canceled):
		w.failReportingJob(terminalCtx, jobID, reportingJobWorkerContextCanceledCode, errors.New(reportingJobWorkerContextCanceledError), true, progress)
		return true
	}
	return false
}

func (w *reportingJobWorker) failReportingJob(ctx context.Context, jobID uuid.UUID, code string, err error, retryable bool, progress jobs.Progress) {
	if err == nil {
		err = errors.New(code)
	}
	message := err.Error()
	if code == reportingJobTimeoutCode {
		message = reportingJobTimeoutMessage
	}
	ctxTerminal, cancel := w.terminalContext(ctx)
	defer cancel()
	_, _ = w.jobManager.CompleteFailed(ctxTerminal, jobs.TransitionParams{
		JobID:    jobID,
		Progress: progress,
		ErrorSummary: &jobs.ErrorSummary{
			Code:      code,
			Message:   message,
			Retryable: retryable,
			Details:   map[string]any{},
		},
	})
}

func (w *reportingJobWorker) completeCanceled(ctx context.Context, jobID uuid.UUID, progress jobs.Progress) {
	ctxTerminal, cancel := w.terminalContext(ctx)
	defer cancel()
	_, _ = w.jobManager.CompleteCanceled(ctxTerminal, jobs.TransitionParams{
		JobID:    jobID,
		Progress: progress,
	})
}

func (w *reportingJobWorker) terminalContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := w.terminalTimeout
	if timeout <= 0 {
		timeout = reportingJobTerminalTransitionTimeout
	}
	return context.WithTimeout(context.WithoutCancel(ctx), timeout)
}

func intPtr(value int) *int {
	return &value
}
