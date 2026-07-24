package reporting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	reportingJobHandlerName                = "snapshot_reporting.job_worker_v1"
	reportingJobDispatchDelay              = 25 * time.Millisecond
	defaultReportingJobTimeout             = 2 * time.Minute
	reportingJobTerminalTransitionTimeout  = 5 * time.Second
	reportingJobTimeoutCode                = "reporting_job_timeout"
	reportingJobWorkerContextCanceledCode  = "reporting_job_context_canceled"
	reportingJobUnknownKindCode            = "internal_error"
	reportingJobInternalErrorCode          = "internal_error"
	reportingJobKindSnapshotCreate         = "snapshot_create"
	reportingJobKindReleaseCreate          = "release_create"
	reportingJobRenderFailedCode           = "release_render_failed"
	reportingJobRenderFailedMessage        = "Release render failed."
	reportingJobSnapshotCreatedCode        = "snapshot_created"
	reportingJobSnapshotCreatedMessage     = "Snapshot created."
	reportingJobReleaseCreatedCode         = "release_created"
	reportingJobReleaseCreatedMessage      = "Release rendered."
	reportingJobTimeoutMessage             = "Reporting job timed out."
	reportingJobWorkerContextCanceledError = "reporting job worker context canceled"
)

type reportingJobDispatcher interface {
	Dispatch(jobID string)
}

type asyncReportingJobDispatcher struct {
	worker *reportingJobWorker
	delay  time.Duration
}

type durableReportingJobDispatcher struct {
	runner *jobs.Runner
}

func newDurableReportingJobDispatcher(runner *jobs.Runner) reportingJobDispatcher {
	return durableReportingJobDispatcher{runner: runner}
}

func (d durableReportingJobDispatcher) Dispatch(jobID string) {
	if d.runner == nil {
		return
	}
	_ = d.runner.DispatchJob(reportingJobHandlerName, jobID)
}

func newAsyncReportingJobDispatcher(worker *reportingJobWorker, delay time.Duration) reportingJobDispatcher {
	return &asyncReportingJobDispatcher{worker: worker, delay: delay}
}

func (d *asyncReportingJobDispatcher) Dispatch(jobID string) {
	if d == nil || d.worker == nil {
		return
	}
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	go func() {
		if d.delay > 0 {
			timer := time.NewTimer(d.delay)
			<-timer.C
		}
		d.worker.Run(context.Background(), parsed)
	}()
}

type reportingJobStore interface {
	ReportingJobKind(context.Context, uuid.UUID) (string, error)
	CompleteSnapshotCreateJob(context.Context, uuid.UUID) (uuid.UUID, error)
	ReleasePayloadForJob(context.Context, uuid.UUID) (releaseCreateJobPayload, error)
	CompleteReleaseRenderFailedJob(context.Context, uuid.UUID, RedactionProfile, string, string, time.Time) (uuid.UUID, error)
	CompleteReleaseCreateJob(context.Context, uuid.UUID, RenderedRelease, time.Time) (uuid.UUID, error)
}

type reportingJobManager interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	MarkRunning(context.Context, uuid.UUID, jobs.Progress, *string) (jobs.Resource, error)
	CompleteSucceeded(context.Context, jobs.TransitionParams) (jobs.Resource, error)
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
	now             func() time.Time
	timeout         time.Duration
	terminalTimeout time.Duration
}

func newReportingJobWorker(store reportingJobStore, jobManager reportingJobManager, now func() time.Time) *reportingJobWorker {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &reportingJobWorker{
		store:           store,
		jobManager:      jobManager,
		renderer:        defaultReleaseCandidateRenderer{},
		now:             now,
		timeout:         defaultReportingJobTimeout,
		terminalTimeout: reportingJobTerminalTransitionTimeout,
	}
}

func (w *reportingJobWorker) Run(parentCtx context.Context, jobID uuid.UUID) {
	if w == nil || w.store == nil || w.jobManager == nil || w.renderer == nil {
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
	default:
		w.failReportingJob(parentCtx, jobID, reportingJobUnknownKindCode, fmt.Errorf("unknown reporting job kind %q", kind), false, jobs.Progress{Completed: 0, Total: intPtr(1)})
	}
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
	snapshotID, err := w.store.CompleteSnapshotCreateJob(ctx, jobID)
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, progress)
		return
	}
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, _ = w.jobManager.CompleteSucceeded(ctxTerminal, jobs.TransitionParams{
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
	releaseID, err := w.store.CompleteReleaseCreateJob(ctx, jobID, rendered, w.now().UTC())
	if err != nil {
		if w.completeContextFailure(ctx, parentCtx, jobID, progress) {
			return
		}
		w.failReportingJob(parentCtx, jobID, reportingJobInternalErrorCode, err, false, progress)
		return
	}
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, _ = w.jobManager.CompleteSucceeded(ctxTerminal, jobs.TransitionParams{
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
	})
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
