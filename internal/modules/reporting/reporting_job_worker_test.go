package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestReportingJobWorkerCancelsBeforeWork(t *testing.T) {
	manager := &fakeReportingJobManager{status: jobs.StatusCancelRequested}
	store := &fakeReportingJobStore{kind: reportingJobKindSnapshotCreate}
	worker := newFakeReportingJobWorker(store, manager, nil)

	worker.Run(context.Background(), jobs.Execution{})

	if manager.status != jobs.StatusCanceled {
		t.Fatalf("job status = %q want canceled", manager.status)
	}
	if store.completeSnapshotCalls != 0 {
		t.Fatalf("snapshot completion ran after cancel request: %d", store.completeSnapshotCalls)
	}
	if manager.canceled == nil || manager.canceled.ResultSummary.Code != "job_canceled" {
		t.Fatalf("cancel summary not recorded: %#v", manager.canceled)
	}
}

func TestReportingJobWorkerTimeoutFailsBeforePersistingRelease(t *testing.T) {
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
	worker := newFakeReportingJobWorker(store, manager, blockingReleaseRenderer{})
	worker.timeout = time.Millisecond

	worker.Run(context.Background(), jobs.Execution{})

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if manager.failed == nil || manager.failed.ErrorSummary.Code == "" {
		t.Fatalf("failure summary missing: %#v", manager.failed)
	}
	if got := manager.failed.ErrorSummary.Code; got != reportingJobTimeoutCode {
		t.Fatalf("failure code = %q want %q", got, reportingJobTimeoutCode)
	}
	if !manager.failed.ErrorSummary.Retryable {
		t.Fatalf("timeout failure must be retryable: %#v", manager.failed.ErrorSummary)
	}
	if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 0 {
		t.Fatalf("release persisted after timeout: complete=%d render_failed=%d", store.completeReleaseCalls, store.renderFailedReleaseCalls)
	}
}

func TestReportingJobWorkerRenderFailureSummaryIsDeterministic(t *testing.T) {
	releaseID := uuid.MustParse("00000000-0000-0000-0000-000000000304")
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, renderFailedReleaseID: releaseID}
	worker := newFakeReportingJobWorker(store, manager, failingReleaseRenderer{reasonCode: "remote_asset_runtime_ref"})

	worker.Run(context.Background(), jobs.Execution{})

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if store.renderFailedReleaseCalls != 1 || store.completeReleaseCalls != 0 {
		t.Fatalf("unexpected release persistence calls: render_failed=%d complete=%d", store.renderFailedReleaseCalls, store.completeReleaseCalls)
	}
	if manager.failed == nil || manager.failed.ErrorSummary.Code == "" {
		t.Fatalf("failure summary missing: %#v", manager.failed)
	}
	summary := manager.failed.ErrorSummary
	if summary.Code != reportingJobRenderFailedCode || summary.Message != reportingJobRenderFailedMessage || summary.Retryable {
		t.Fatalf("unexpected render failure summary: %#v", summary)
	}
	if summary.Details["reason_code"] != "remote_asset_runtime_ref" || summary.Details["release_id"] != releaseID.String() {
		t.Fatalf("unexpected render failure details: %#v", summary.Details)
	}
}

func TestReportingJobWorkerParticipantFailurePublishesNoRenderOutput(t *testing.T) {
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
	worker := newFakeReportingJobWorker(store, manager, succeedingReleaseRenderer{})
	worker.renderExport = failingRenderExportInvoker{}

	worker.Run(context.Background(), jobs.Execution{})

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 0 {
		t.Fatalf("participant failure persisted render output: complete=%d failed=%d", store.completeReleaseCalls, store.renderFailedReleaseCalls)
	}
}

func TestReportingPreviewWorkerHonorsCancellationAndParticipantFailure(t *testing.T) {
	canceledManager := &fakeReportingJobManager{status: jobs.StatusCancelRequested}
	canceledStore := &fakeReportingJobStore{kind: reportingJobKindCompositionPreview}
	canceledWorker := newFakeReportingJobWorker(canceledStore, canceledManager, succeedingReleaseRenderer{})
	canceledWorker.Run(context.Background(), jobs.Execution{})
	if canceledManager.status != jobs.StatusCanceled || canceledStore.completePreviewCalls != 0 {
		t.Fatalf("canceled preview executed: status=%q complete=%d", canceledManager.status, canceledStore.completePreviewCalls)
	}

	failedManager := &fakeReportingJobManager{status: jobs.StatusRunning}
	failedStore := &fakeReportingJobStore{kind: reportingJobKindCompositionPreview}
	failedWorker := newFakeReportingJobWorker(failedStore, failedManager, succeedingReleaseRenderer{})
	failedWorker.renderExport = failingRenderExportInvoker{}
	failedWorker.Run(context.Background(), jobs.Execution{})
	if failedManager.status != jobs.StatusFailed || failedStore.completePreviewCalls != 0 {
		t.Fatalf("failed preview published output: status=%q complete=%d", failedManager.status, failedStore.completePreviewCalls)
	}
}

func newFakeReportingJobWorker(store *fakeReportingJobStore, manager *fakeReportingJobManager, renderer releaseCandidateRenderer) *reportingJobWorker {
	if renderer == nil {
		renderer = succeedingReleaseRenderer{releaseID: uuid.MustParse("00000000-0000-0000-0000-000000000399")}
	}
	return &reportingJobWorker{
		store:           store,
		jobManager:      manager,
		jobFinalizer:    fakeReportingJobFinalizer{manager: manager},
		renderExport:    fakeRenderExportInvoker{},
		renderer:        renderer,
		now:             func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) },
		timeout:         time.Second,
		terminalTimeout: time.Second,
	}
}

type fakeReportingJobStore struct {
	kind                     string
	snapshotID               uuid.UUID
	releaseID                uuid.UUID
	renderFailedReleaseID    uuid.UUID
	completeSnapshotCalls    int
	releasePayloadCalls      int
	renderFailedReleaseCalls int
	completeReleaseCalls     int
	completePreviewCalls     int
}

func (s *fakeReportingJobStore) ReportingJobKind(ctx context.Context, _ uuid.UUID) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.kind, nil
}

func (s *fakeReportingJobStore) CompleteSnapshotCreateJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, snapshotID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.completeSnapshotCalls++
	s.snapshotID = snapshotID
	return nil
}

func (s *fakeReportingJobStore) ReleasePayloadForJob(ctx context.Context, _ uuid.UUID) (releaseCreateJobPayload, error) {
	if err := ctx.Err(); err != nil {
		return releaseCreateJobPayload{}, err
	}
	s.releasePayloadCalls++
	model := ExportModel{
		SchemaID:          ExportModelSchemaID,
		IncidentID:        "00000000-0000-0000-0000-000000000311",
		SnapshotID:        "00000000-0000-0000-0000-000000000312",
		DerivationVersion: DerivationVersion,
	}
	modelJSON, _ := canonicalJSON(model)
	return releaseCreateJobPayload{
		ActorUserID:             "00000000-0000-0000-0000-000000000313",
		IncidentID:              model.IncidentID,
		SnapshotID:              model.SnapshotID,
		ExportModel:             model,
		ExportModelSHA256:       hashHex(modelJSON),
		RedactionProfileID:      InternalRedactionProfileID,
		RedactionProfileVersion: "1",
	}, nil
}

func (s *fakeReportingJobStore) CompositionPreviewPayloadForJob(ctx context.Context, _ uuid.UUID) (compositionPreviewJobPayload, error) {
	payload, err := s.ReleasePayloadForJob(ctx, uuid.Nil)
	return compositionPreviewJobPayload{
		PreviewAttemptID: uuid.MustParse("00000000-0000-0000-0000-000000000314"),
		Release:          payload,
	}, err
}

func (s *fakeReportingJobStore) CompleteReleaseRenderFailedJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, _ RedactionProfile, _ string, _ string, _ time.Time) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.UUID{}, err
	}
	s.renderFailedReleaseCalls++
	if s.renderFailedReleaseID == uuid.Nil {
		s.renderFailedReleaseID = uuid.MustParse("00000000-0000-0000-0000-000000000306")
	}
	return s.renderFailedReleaseID, nil
}

func (s *fakeReportingJobStore) CompleteReleaseCreateJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, releaseID uuid.UUID, _ RenderedRelease, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.completeReleaseCalls++
	s.releaseID = releaseID
	return nil
}

func (s *fakeReportingJobStore) CompleteCompositionPreviewJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderedRelease, time.Time) error {
	s.completePreviewCalls++
	return nil
}

type fakeReportingJobFinalizer struct {
	manager *fakeReportingJobManager
}

type fakeRenderExportInvoker struct{}

func (fakeRenderExportInvoker) Invoke(ctx context.Context, invocation RenderExportInvocation) (RenderExportResult, error) {
	return (BuiltInRenderExportParticipant{}).Emit(ctx, invocation)
}

type failingRenderExportInvoker struct{}

func (failingRenderExportInvoker) Invoke(context.Context, RenderExportInvocation) (RenderExportResult, error) {
	return RenderExportResult{}, errors.New("participant failed")
}

func (f fakeReportingJobFinalizer) FinalizeReportingJobSuccess(ctx context.Context, request JobSuccessFinalization) (jobs.Resource, error) {
	if request.Mutate != nil {
		if err := request.Mutate(ctx, nil); err != nil {
			return jobs.Resource{}, err
		}
	}
	return f.manager.CompleteSucceeded(ctx, request.Execution, request.Completion)
}

func (f fakeReportingJobFinalizer) FinalizeReportingJobFailure(ctx context.Context, request JobFailureFinalization) (jobs.Resource, error) {
	if request.Mutate != nil {
		if err := request.Mutate(ctx, nil); err != nil {
			return jobs.Resource{}, err
		}
	}
	return f.manager.CompleteFailed(ctx, request.Execution, request.Completion)
}

type fakeReportingJobManager struct {
	status    string
	failed    *jobs.FailureCompletion
	succeeded *jobs.SuccessCompletion
	canceled  *jobs.CancellationCompletion
}

func (m *fakeReportingJobManager) Get(ctx context.Context, _ uuid.UUID) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	return jobs.Resource{Status: m.status}, nil
}

func (m *fakeReportingJobManager) ObserveExecution(ctx context.Context, _ jobs.Execution) (jobs.Resource, error) {
	return m.Get(ctx, uuid.Nil)
}

func (m *fakeReportingJobManager) UpdateProgress(ctx context.Context, _ jobs.Execution, progress jobs.Progress, message *string) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	if m.status == jobs.StatusCancelRequested {
		return jobs.Resource{}, jobs.ErrInvalidTransition
	}
	m.status = jobs.StatusRunning
	return jobs.Resource{Status: m.status, Progress: progress, Message: message}, nil
}

func (m *fakeReportingJobManager) CompleteSucceeded(ctx context.Context, _ jobs.Execution, params jobs.SuccessCompletion) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusSucceeded
	copied := params
	m.succeeded = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ResultSummary: &params.ResultSummary}, nil
}

func (m *fakeReportingJobManager) CompleteFailed(ctx context.Context, _ jobs.Execution, params jobs.FailureCompletion) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusFailed
	copied := params
	m.failed = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ErrorSummary: &params.ErrorSummary}, nil
}

func (m *fakeReportingJobManager) CompleteCanceled(ctx context.Context, _ jobs.Execution, params jobs.CancellationCompletion) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusCanceled
	copied := params
	if copied.ResultSummary.Code == "" {
		copied.ResultSummary = jobs.ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	m.canceled = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ResultSummary: &copied.ResultSummary}, nil
}

type blockingReleaseRenderer struct{}

func (blockingReleaseRenderer) Render(ctx context.Context, _ releaseCreateJobPayload) (RenderedRelease, string, error) {
	<-ctx.Done()
	return RenderedRelease{}, "", ctx.Err()
}

type failingReleaseRenderer struct {
	reasonCode string
}

func (r failingReleaseRenderer) Render(context.Context, releaseCreateJobPayload) (RenderedRelease, string, error) {
	return RenderedRelease{
		Profile:       RedactionProfile{ProfileID: ExternalRedactionProfileID, Version: "1"},
		ProfileSHA256: "profile-sha",
	}, r.reasonCode, errors.New("render failed")
}

type succeedingReleaseRenderer struct {
	releaseID uuid.UUID
}

func (succeedingReleaseRenderer) Render(context.Context, releaseCreateJobPayload) (RenderedRelease, string, error) {
	return RenderedRelease{
		Profile:       RedactionProfile{ProfileID: ExternalRedactionProfileID, Version: "1"},
		ProfileSHA256: "profile-sha",
	}, "", nil
}
