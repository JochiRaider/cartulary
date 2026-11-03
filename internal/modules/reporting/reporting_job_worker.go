package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	// Backstop for dispatch/payload loading. Render attempts additionally use
	// their operative manifest-owned total, which may be shorter.
	defaultReportingJobTimeout            = 30 * time.Minute
	reportingJobTerminalTransitionTimeout = 5 * time.Second
	reportingJobTimeoutCode               = "reporting_job_timeout"
	reportingJobUnknownKindCode           = "internal_error"
	reportingJobInternalErrorCode         = "internal_error"
	reportingJobKindSnapshotCreate        = "snapshot_create"
	reportingJobKindReleaseCreate         = "release_create"
	reportingJobKindCompositionPreview    = "composition_preview"
	reportingJobRenderFailedCode          = "release_render_failed"
	reportingJobRenderFailedMessage       = "Release render failed."
	reportingJobSnapshotCreatedCode       = "snapshot_created"
	reportingJobSnapshotCreatedMessage    = "Snapshot created."
	reportingJobReleaseCreatedCode        = "release_created"
	reportingJobReleaseCreatedMessage     = "Release rendered."
	reportingJobPreviewCreatedCode        = "composition_preview_rendered"
	reportingJobPreviewCreatedMessage     = "Composition preview rendered."
	reportingJobTimeoutMessage            = "Reporting job timed out."
)

type reportingJobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	Notify(uuid.UUID)
}

type reportingJobAdmission interface {
	jobs.RetainedJobReader
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
}

type reportingJobTransactions interface {
	reportingJobAdmission
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	ExtensionFinalizationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error)
	ExtensionCancellationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error)
}

type reportingJobStore interface {
	authorizeExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	PrepareSnapshotJob(context.Context, jobs.Execution) error
	SnapshotCandidateForJob(context.Context, jobs.Execution) (SnapshotCandidate, error)
	ReportingJobKind(context.Context, jobs.Execution) (string, error)
	CompleteSnapshotCreateJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, SnapshotCandidate) error
	ReleasePayloadForJob(context.Context, jobs.Execution) (releaseCreateJobPayload, error)
	CompositionPreviewPayloadForJob(context.Context, jobs.Execution) (compositionPreviewJobPayload, error)
	prepareRenderFailureCandidate(context.Context, jobs.Execution, RenderFailure) (renderFailureCandidate, error)
	completeRenderFailureTx(context.Context, pgx.Tx, uuid.UUID, renderFailureCandidate, time.Time) (*uuid.UUID, error)
	PrepareRenderCandidate(context.Context, jobs.Execution, renderInput, RenderedRelease) (RenderCandidate, error)
	CompleteReleaseCreateJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderCandidate, time.Time) error
	CompleteCompositionPreviewJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderCandidate, time.Time) error
}

type reportingJobManager interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
	UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error)
}

type candidateRenderer interface {
	Render(context.Context, renderInput) (RenderedRelease, string, error)
}

type reportingJobWorker struct {
	store                  reportingJobStore
	jobManager             reportingJobManager
	renderer               candidateRenderer
	jobFinalizer           JobSuccessFinalizer
	now                    func() time.Time
	timeout                time.Duration
	terminalTimeout        time.Duration
	renderObservationTicks func() (<-chan time.Time, func())
}

func newReportingJobWorker(
	store reportingJobStore,
	jobManager reportingJobManager,
	jobFinalizer JobSuccessFinalizer,
	renderer candidateRenderer,
	now func() time.Time,
) *reportingJobWorker {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &reportingJobWorker{
		store:           store,
		jobManager:      jobManager,
		jobFinalizer:    jobFinalizer,
		renderer:        renderer,
		now:             now,
		timeout:         defaultReportingJobTimeout,
		terminalTimeout: reportingJobTerminalTransitionTimeout,
	}
}

func (w *reportingJobWorker) Handle(parentCtx context.Context, execution jobs.Execution) error {
	if w == nil || w.store == nil || w.jobManager == nil || w.renderer == nil || w.jobFinalizer == nil {
		return errors.New("reporting: incomplete worker dependencies")
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
	return w.execute(runCtx, parentCtx, execution)
}

func (w *reportingJobWorker) execute(ctx context.Context, parentCtx context.Context, execution jobs.Execution) error {
	progress := jobs.Progress{Completed: 0, Total: intPtr(1)}
	if stop, err := w.stopIfInterrupted(ctx, parentCtx, execution, progress); stop || err != nil {
		return err
	}
	if _, err := w.jobManager.UpdateProgress(ctx, execution, progress, nil); err != nil {
		return err
	}
	kind, err := w.store.ReportingJobKind(ctx, execution)
	if err != nil {
		return w.failOrObserveContext(ctx, parentCtx, execution, reportingJobInternalErrorCode, err, progress)
	}
	switch kind {
	case reportingJobKindSnapshotCreate:
		return w.executeSnapshotCreateJob(ctx, parentCtx, execution)
	case reportingJobKindReleaseCreate:
		return w.executeReleaseCreateJob(ctx, parentCtx, execution)
	case reportingJobKindCompositionPreview:
		return w.executeCompositionPreviewJob(ctx, parentCtx, execution)
	default:
		return w.failReportingJob(parentCtx, execution, reportingJobUnknownKindCode, nil, false, progress)
	}
}

func (w *reportingJobWorker) executeCompositionPreviewJob(ctx context.Context, parentCtx context.Context, execution jobs.Execution) error {
	started := time.Now()
	jobID := execution.JobID()
	total := 1
	progress := jobs.Progress{Completed: 0, Total: &total}
	var payload compositionPreviewJobPayload
	workErr, controlErr := w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		payload, err = w.store.CompositionPreviewPayloadForJob(work, execution)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	err := workErr
	if err != nil {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	limits, err := resolveRenderLimits(payload.TemplateContract.DeclaredLimits)
	if err != nil {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	ctx, cancelAttempt := withRenderDeadlineFrom(ctx, started, limits.timeout("total_attempt"), "release_state", "timeout.total_attempt")
	defer cancelAttempt()
	workErr, controlErr = w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		payload.renderInput, err = admitFrozenRenderInput(work, payload.renderInput)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	if workErr != nil {
		return w.finishRenderFailure(parentCtx, execution, workErr)
	}
	renderResult, controlErr := w.renderWhileCurrent(ctx, execution, payload.renderInput)
	if controlErr != nil {
		if parentCtx.Err() == nil && errors.Is(controlErr, context.DeadlineExceeded) {
			if renderResult.err != nil {
				return w.finishRenderFailure(parentCtx, execution, errors.Join(renderResult.err, controlErr))
			}
			return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, controlErr))
		}
		if errors.Is(controlErr, jobs.ErrCancellationRequested) {
			return w.finishReportingFailureOutcome(parentCtx, execution, errors.Join(context.Canceled, renderResult.err), true)
		}
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	if parentCtx.Err() != nil {
		return parentCtx.Err()
	}
	if renderResult.reason == "partial_output_cleanup_failed" && renderResult.err != nil {
		return w.finishRenderFailure(parentCtx, execution, renderResult.err)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, ctx.Err()))
	}
	observationCtx, finishObservation := w.terminalContext(parentCtx)
	stop, observationErr := w.stopIfInterrupted(observationCtx, parentCtx, execution, progress)
	finishObservation()
	if stop || observationErr != nil {
		return observationErr
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, ctx.Err()))
	}
	if renderResult.err != nil {
		return w.finishRenderFailure(parentCtx, execution, renderResult.err)
	}
	var candidate RenderCandidate
	workErr, controlErr = w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		candidate, err = w.store.PrepareRenderCandidate(work, execution, payload.renderInput, renderResult.rendered)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	if workErr != nil {
		return w.finishRenderFailure(parentCtx, execution, workErr)
	}
	_, err = w.jobFinalizer.FinalizeReportingJobSuccess(ctx, JobSuccessFinalization{
		Execution: execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:         reportingJobPreviewCreatedCode,
				Message:      reportingJobPreviewCreatedMessage,
				ResourceRefs: []jobs.ResourceRef{},
			},
		},
		FinalCommitID: CompositionPreviewOperationKind + ":" + jobID.String(),
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			if err := w.store.authorizeExecutionTx(ctx, tx, execution); err != nil {
				return err
			}
			if err := w.store.CompleteCompositionPreviewJobTx(
				ctx,
				tx,
				jobID,
				payload.PreviewAttemptID,
				candidate,
				w.now().UTC(),
			); err != nil {
				return err
			}
			return nil
		},
	})
	if errors.Is(err, jobs.ErrCancellationRequested) {
		return w.completeCanceled(parentCtx, execution, progress)
	}
	if errors.Is(err, errReportingAuthorityUnavailable) {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, err))
	}
	return err
}

func (w *reportingJobWorker) executeSnapshotCreateJob(ctx context.Context, parentCtx context.Context, execution jobs.Execution) error {
	jobID := execution.JobID()
	total := 1
	if stop, err := w.stopIfInterrupted(ctx, parentCtx, execution, jobs.Progress{Completed: 0, Total: &total}); stop || err != nil {
		return err
	}
	var candidate SnapshotCandidate
	workErr, controlErr := w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		if err := w.store.PrepareSnapshotJob(work, execution); err != nil {
			return err
		}
		var err error
		candidate, err = w.store.SnapshotCandidateForJob(work, execution)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderControl(ctx, parentCtx, execution, jobs.Progress{Completed: 0, Total: &total}, controlErr)
	}
	if workErr != nil {
		code := reportingJobInternalErrorCode
		if errors.Is(workErr, errReportingAuthorityUnavailable) {
			code = reportingJobAuthorizationDeniedCode
		}
		var conflict *SnapshotBoundaryConflictError
		if errors.As(workErr, &conflict) {
			code = "snapshot_source_boundary_conflict"
		}
		return w.failOrObserveContext(ctx, parentCtx, execution, code, workErr, jobs.Progress{Completed: 0, Total: &total})
	}
	snapshotID := reportingResourceID(jobID, "snapshot")
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	_, err := w.jobFinalizer.FinalizeReportingJobSuccess(ctxTerminal, JobSuccessFinalization{
		Execution: execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
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
			if err := w.store.authorizeExecutionTx(ctx, tx, execution); err != nil {
				return err
			}
			return w.store.CompleteSnapshotCreateJobTx(ctx, tx, jobID, snapshotID, candidate)
		},
	})
	if errors.Is(err, errReportingAuthorityUnavailable) {
		return w.failOrObserveContext(ctx, parentCtx, execution, reportingJobAuthorizationDeniedCode, err, jobs.Progress{Completed: 0, Total: &total})
	}
	return err
}

func (w *reportingJobWorker) executeReleaseCreateJob(ctx context.Context, parentCtx context.Context, execution jobs.Execution) error {
	started := time.Now()
	jobID := execution.JobID()
	total := 1
	progress := jobs.Progress{Completed: 0, Total: &total}
	var payload releaseCreateJobPayload
	workErr, controlErr := w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		payload, err = w.store.ReleasePayloadForJob(work, execution)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	err := workErr
	if err != nil {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	limits, err := resolveRenderLimits(payload.TemplateContract.DeclaredLimits)
	if err != nil {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	ctx, cancelAttempt := withRenderDeadlineFrom(ctx, started, limits.timeout("total_attempt"), "release_state", "timeout.total_attempt")
	defer cancelAttempt()
	workErr, controlErr = w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		payload.renderInput, err = admitFrozenRenderInput(work, payload.renderInput)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	if workErr != nil {
		return w.finishRenderFailure(parentCtx, execution, workErr)
	}
	renderResult, controlErr := w.renderWhileCurrent(ctx, execution, payload.renderInput)
	if controlErr != nil {
		if parentCtx.Err() == nil && errors.Is(controlErr, context.DeadlineExceeded) {
			if renderResult.err != nil {
				return w.finishRenderFailure(parentCtx, execution, errors.Join(renderResult.err, controlErr))
			}
			return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, controlErr))
		}
		if errors.Is(controlErr, jobs.ErrCancellationRequested) {
			return w.finishReportingFailureOutcome(parentCtx, execution, errors.Join(context.Canceled, renderResult.err), true)
		}
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	rendered, reasonCode, renderErr := renderResult.rendered, renderResult.reason, renderResult.err
	if parentCtx.Err() != nil {
		return parentCtx.Err()
	}
	if renderErr != nil && reasonCode == "partial_output_cleanup_failed" {
		return w.finishRenderFailure(parentCtx, execution, renderErr)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, ctx.Err()))
	}
	// Observe with a bounded control context, not the expiring work context.
	// Finalization still validates the opaque execution under the Jobs row lock.
	observationCtx, finishObservation := w.terminalContext(parentCtx)
	stop, observationErr := w.stopIfInterrupted(observationCtx, parentCtx, execution, progress)
	finishObservation()
	if stop || observationErr != nil {
		return observationErr
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, ctx.Err()))
	}
	if renderErr != nil {
		return w.finishRenderFailure(parentCtx, execution, renderErr)
	}
	var candidate RenderCandidate
	workErr, controlErr = w.whileExecutionCurrent(ctx, execution, func(work context.Context) error {
		var err error
		candidate, err = w.store.PrepareRenderCandidate(work, execution, payload.renderInput, rendered)
		return err
	})
	if controlErr != nil {
		return w.stopForRenderPreparationControl(ctx, parentCtx, execution, progress, controlErr)
	}
	if workErr != nil {
		return w.finishRenderFailure(parentCtx, execution, workErr)
	}
	releaseID := reportingResourceID(jobID, "release")
	// Publication is work inside the total-attempt deadline. The finalizer's
	// independent commit proof may confirm an already committed success; an
	// indeterminate acknowledgement must never become a second terminal write.
	_, err = w.jobFinalizer.FinalizeReportingJobSuccess(ctx, JobSuccessFinalization{
		Execution: execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
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
			if err := w.store.authorizeExecutionTx(ctx, tx, execution); err != nil {
				return err
			}
			if err := w.store.CompleteReleaseCreateJobTx(ctx, tx, jobID, releaseID, candidate, w.now().UTC()); err != nil {
				return err
			}
			return nil
		},
	})
	if errors.Is(err, jobs.ErrCancellationRequested) {
		return w.completeCanceled(parentCtx, execution, progress)
	}
	if errors.Is(err, errReportingAuthorityUnavailable) {
		return w.finishRenderFailure(parentCtx, execution, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return w.finishRenderFailure(parentCtx, execution, classifyRenderContextError(ctx, err))
	}
	return err
}

func (w *reportingJobWorker) finishRenderFailure(parentCtx context.Context, execution jobs.Execution, cause error) error {
	if errors.Is(cause, jobs.ErrCancellationRequested) {
		return w.finishReportingFailureOutcome(parentCtx, execution, context.Canceled, true)
	}
	return w.finishReportingFailureOutcome(parentCtx, execution, cause, false)
}

func (w *reportingJobWorker) finishReportingFailureOutcome(parentCtx context.Context, execution jobs.Execution, cause error, canceled bool) error {
	if err := parentCtx.Err(); err != nil {
		return err
	}
	if errors.Is(cause, jobs.ErrExecutionLost) {
		return cause
	}
	control, finishControl := w.terminalContext(parentCtx)
	kind, kindErr := w.store.ReportingJobKind(control, execution)
	finishControl()
	if kindErr != nil {
		return kindErr
	}
	if kind == reportingJobKindSnapshotCreate {
		if canceled {
			ctxTerminal, cancel := w.terminalContext(parentCtx)
			defer cancel()
			_, err := w.jobFinalizer.FinalizeReportingJobCancellation(ctxTerminal, JobCancellationFinalization{Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}}})
			return err
		}
		code := reportingJobInternalErrorCode
		if errors.Is(cause, context.DeadlineExceeded) {
			code = reportingJobTimeoutCode
		}
		return w.failReportingJob(parentCtx, execution, code, cause, code == reportingJobTimeoutCode, jobs.Progress{Completed: 0, Total: intPtr(1)})
	}
	failure := renderFailureDetails(cause)
	if err := validateRenderFailure(failure); err != nil {
		return err
	}
	jobID, total := execution.JobID(), 1
	ctxTerminal, cancel := w.terminalContext(parentCtx)
	defer cancel()
	candidate, err := w.store.prepareRenderFailureCandidate(ctxTerminal, execution, failure)
	if err != nil {
		return err
	}
	failure = candidate.failure
	details := map[string]any{"reason_code": failure.ReasonCode}
	mutate := func(ctx context.Context, tx pgx.Tx) error {
		releaseID, err := w.store.completeRenderFailureTx(ctx, tx, jobID, candidate, w.now().UTC())
		if err != nil {
			return err
		}
		if releaseID != nil {
			details["release_id"] = releaseID.String()
		}
		return nil
	}
	if canceled {
		_, err := w.jobFinalizer.FinalizeReportingJobCancellation(ctxTerminal, JobCancellationFinalization{
			Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 1, Total: &total}}, Mutate: mutate,
		})
		return err
	}
	_, err = w.jobFinalizer.FinalizeReportingJobFailure(ctxTerminal, JobFailureFinalization{
		Execution: execution,
		TimedOut:  failure.ReasonCode == "render_timeout" || errors.Is(cause, context.DeadlineExceeded),
		Completion: jobs.FailureCompletion{
			Progress:     jobs.Progress{Completed: 1, Total: &total},
			ErrorSummary: jobs.ErrorSummary{Code: renderFailureJobCode(failure.ReasonCode), Message: reportingJobRenderFailedMessage, Retryable: false, Details: details},
		},
		Mutate: mutate,
	})
	if errors.Is(err, jobs.ErrCancellationRequested) {
		cancelCause := error(context.Canceled)
		if failure.ReasonCode == "partial_output_cleanup_failed" {
			cancelCause = cause
		}
		return w.finishReportingFailureOutcome(parentCtx, execution, cancelCause, true)
	}
	return err
}

func admitFrozenRenderInput(ctx context.Context, payload renderInput) (renderInput, error) {
	limits, err := resolveRenderLimits(payload.TemplateContract.DeclaredLimits)
	if err != nil {
		return renderInput{}, err
	}
	ctx, cancel := withRenderDeadline(ctx, limits.timeout("export_model_materialization"), "export_model_materialization", "timeout.export_model_materialization")
	defer cancel()
	if ctx.Err() != nil {
		return renderInput{}, classifyRenderContextError(ctx, ctx.Err())
	}
	if err := validateFrozenRenderProfile(payload); err != nil {
		return renderInput{}, renderFailureAt(err, "redaction", "redaction_manifest_invalid", "redaction_profile_invalid")
	}
	model, err := admitFrozenExportModel(ctx, payload.ExportModel, payload.ExportModelSHA256)
	if ctx.Err() != nil {
		return renderInput{}, classifyRenderContextError(ctx, ctx.Err())
	}
	if err != nil {
		return renderInput{}, renderFailureAt(err, "export_model_materialization", "export_model_invalid", "invalid_type")
	}
	payload.ExportModel = model
	return payload, nil
}

func reportingResourceID(jobID uuid.UUID, kind string) uuid.UUID {
	return uuid.NewSHA1(jobID, []byte("snapshot_reporting:"+kind))
}

// Every checkpoint observes the live execution before continuing. Observation
// failure is an error, never permission to render or finalize.
func (w *reportingJobWorker) stopIfInterrupted(ctx, parentCtx context.Context, execution jobs.Execution, progress jobs.Progress) (bool, error) {
	if stop, err := w.stopForContext(ctx, parentCtx, execution, progress); stop {
		return true, err
	}
	resource, err := w.jobManager.ObserveExecution(ctx, execution)
	if err != nil {
		return true, err
	}
	switch resource.Status {
	case jobs.StatusRunning:
		return false, nil
	case jobs.StatusCancelRequested:
		return true, w.completeCanceled(parentCtx, execution, progress)
	case jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled:
		return true, nil
	default:
		return true, jobs.ErrInvalidTransition
	}
}

func (w *reportingJobWorker) stopForContext(ctx, parentCtx context.Context, execution jobs.Execution, progress jobs.Progress) (bool, error) {
	// Runner shutdown owns conditional release. Do not replace it with a new
	// failed Job or retry policy in this handler.
	if err := parentCtx.Err(); err != nil {
		return true, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return true, w.failReportingJob(parentCtx, execution, reportingJobTimeoutCode, ctx.Err(), true, progress)
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}
	return false, nil
}

func (w *reportingJobWorker) failOrObserveContext(ctx, parentCtx context.Context, execution jobs.Execution, code string, cause error, progress jobs.Progress) error {
	if stop, err := w.stopForContext(ctx, parentCtx, execution, progress); stop {
		return err
	}
	if errors.Is(cause, jobs.ErrExecutionLost) {
		return cause
	}
	if errors.Is(cause, jobs.ErrCancellationRequested) {
		return w.completeCanceled(parentCtx, execution, progress)
	}
	return w.failReportingJob(parentCtx, execution, code, cause, false, progress)
}

func (w *reportingJobWorker) failReportingJob(ctx context.Context, execution jobs.Execution, code string, err error, retryable bool, progress jobs.Progress) error {
	if errors.Is(err, ErrRequiredReferencePackUnavailable) {
		code = "required_reference_pack_unavailable"
		retryable = false
	}

	// Raw persistence/participant errors can contain paths, SQL or source data.
	// Only closed owner summaries cross the public Jobs boundary.
	message := "Reporting job failed."
	details := snapshotFailureDetails(err)
	switch code {
	case reportingJobAuthorizationDeniedCode:
		message = "Initiating actor is no longer authorized."
		retryable = false
		details = nil
	case reportingJobTimeoutCode:
		message = reportingJobTimeoutMessage
	case reportingJobRenderFailedCode:
		message = reportingJobRenderFailedMessage
	case "required_reference_pack_unavailable":
		message = "A required reference pack is unavailable."
	}
	ctxTerminal, cancel := w.terminalContext(ctx)
	defer cancel()
	_, completionErr := w.jobFinalizer.FinalizeReportingJobFailure(ctxTerminal, JobFailureFinalization{
		Execution: execution,
		TimedOut:  code == reportingJobTimeoutCode,
		Completion: jobs.FailureCompletion{
			Progress: progress,
			ErrorSummary: jobs.ErrorSummary{
				Code: code, Message: message, Retryable: retryable, Details: details,
			},
		},
	})
	if errors.Is(completionErr, jobs.ErrCancellationRequested) {
		return w.completeCanceled(ctx, execution, progress)
	}
	return completionErr
}

func (w *reportingJobWorker) completeCanceled(ctx context.Context, execution jobs.Execution, _ jobs.Progress) error {
	return w.finishReportingFailureOutcome(ctx, execution, context.Canceled, true)
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
