package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

func TestReportingJobWorkerCancelsBeforeWork(t *testing.T) {
	t.Run("snapshot diagnostics exclude raw storage and source values", func(t *testing.T) {
		cause := &snapshotWorkError{stage: "capture_persistence", cause: &pgconn.PgError{Code: "54000", Message: "private SQL and source payload", Detail: "private row"}}
		details := snapshotFailureDetails(cause)
		raw, err := json.Marshal(details)
		if err != nil || string(raw) != `{"failure_kind":"storage_limit","snapshot_stage":"capture_persistence"}` {
			t.Fatalf("unexpected diagnostic: %s %v", raw, err)
		}
		if len(snapshotFailureDetails(&snapshotWorkError{stage: "private stage", cause: cause})) != 0 {
			t.Fatal("unadmitted stage escaped")
		}
	})

	manager := &fakeReportingJobManager{status: jobs.StatusCancelRequested}
	store := &fakeReportingJobStore{kind: reportingJobKindSnapshotCreate}
	worker := newFakeReportingJobWorker(store, manager, nil)

	if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}

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

func TestReportingJobWorkerTimeoutRetainsFailedRelease(t *testing.T) {
	t.Run("cleanup failure wins over timeout", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
		worker := newFakeReportingJobWorker(store, manager, releaseRendererFunc(func(ctx context.Context, _ renderInput) (RenderedRelease, string, error) {
			<-ctx.Done()
			return RenderedRelease{ProfileSHA256: "must-not-publish"}, "partial_output_cleanup_failed", errors.Join(ctx.Err(), renderprocess.ErrCleanup)
		}))
		worker.timeout = 50 * time.Millisecond
		if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
			t.Fatal(err)
		}
		if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 1 || manager.failed == nil || manager.failed.ErrorSummary.Details["reason_code"] != "partial_output_cleanup_failed" || !manager.renderTimedOut {
			t.Fatalf("cleanup failure lost or output admitted: store=%+v summary=%+v", store, manager.failed)
		}
	})
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
	worker := newFakeReportingJobWorker(store, manager, blockingReleaseRenderer{})
	worker.timeout = 50 * time.Millisecond

	if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if manager.failed == nil || manager.failed.ErrorSummary.Code == "" {
		t.Fatalf("failure summary missing: %#v", manager.failed)
	}
	if got := manager.failed.ErrorSummary.Code; got != reportingJobRenderFailedCode {
		t.Fatalf("failure code = %q want %q", got, reportingJobRenderFailedCode)
	}
	if manager.failed.ErrorSummary.Retryable || manager.failed.ErrorSummary.Details["reason_code"] != "render_timeout" || !manager.renderTimedOut {
		t.Fatalf("timeout failure lacks terminal identity: %#v", manager.failed.ErrorSummary)
	}
	if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 1 {
		t.Fatalf("timeout must retain failure without output: complete=%d render_failed=%d", store.completeReleaseCalls, store.renderFailedReleaseCalls)
	}
}

func TestReportingJobWorkerRenderFailureSummaryIsDeterministic(t *testing.T) {
	t.Run("missing retained Reference Pack keeps its owner outcome", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, prepare: func(context.Context) error {
			return renderFailureAt(ErrRequiredReferencePackUnavailable, "export_model_materialization", "export_model_invalid", "required_reference_pack_unavailable")
		}}
		worker := newFakeReportingJobWorker(store, manager, nil)
		if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
			t.Fatal(err)
		}
		if manager.failed == nil || manager.failed.ErrorSummary.Code != "required_reference_pack_unavailable" || store.failure == nil || store.failure.ReasonCode != "required_reference_pack_unavailable" || store.completeReleaseCalls != 0 {
			t.Fatal("preparation erased the Reference Data outcome")
		}
	})

	releaseID := uuid.MustParse("00000000-0000-0000-0000-000000000304")
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, renderFailedReleaseID: releaseID}
	worker := newFakeReportingJobWorker(store, manager, failingReleaseRenderer{reasonCode: "remote_asset_reference_invalid"})

	if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}

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
	if summary.Details["reason_code"] != "remote_asset_reference_invalid" || summary.Details["release_id"] != releaseID.String() {
		t.Fatalf("unexpected render failure details: %#v", summary.Details)
	}
}

func TestReportingJobWorkerParticipantFailurePublishesNoRenderOutput(t *testing.T) {
	manager := &fakeReportingJobManager{status: jobs.StatusRunning}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
	worker := newFakeReportingJobWorker(store, manager, succeedingReleaseRenderer{})
	worker.renderExport = failingRenderExportInvoker{}

	if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 1 {
		t.Fatalf("participant failure must retain failed admission without output: complete=%d failed=%d", store.completeReleaseCalls, store.renderFailedReleaseCalls)
	}
}

func TestReportingPreviewWorkerHonorsCancellationAndParticipantFailure(t *testing.T) {
	canceledManager := &fakeReportingJobManager{status: jobs.StatusCancelRequested}
	canceledStore := &fakeReportingJobStore{kind: reportingJobKindCompositionPreview}
	canceledWorker := newFakeReportingJobWorker(canceledStore, canceledManager, succeedingReleaseRenderer{})
	if err := canceledWorker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}
	if canceledManager.status != jobs.StatusCanceled || canceledStore.completePreviewCalls != 0 {
		t.Fatalf("canceled preview executed: status=%q complete=%d", canceledManager.status, canceledStore.completePreviewCalls)
	}

	failedManager := &fakeReportingJobManager{status: jobs.StatusRunning}
	failedStore := &fakeReportingJobStore{kind: reportingJobKindCompositionPreview}
	failedWorker := newFakeReportingJobWorker(failedStore, failedManager, succeedingReleaseRenderer{})
	failedWorker.renderExport = failingRenderExportInvoker{}
	if err := failedWorker.Handle(context.Background(), jobs.Execution{}); err != nil {
		t.Fatal(err)
	}
	if failedManager.status != jobs.StatusFailed || failedStore.completePreviewCalls != 0 {
		t.Fatalf("failed preview published output: status=%q complete=%d", failedManager.status, failedStore.completePreviewCalls)
	}
}

func newFakeReportingJobWorker(store *fakeReportingJobStore, manager *fakeReportingJobManager, renderer candidateRenderer) *reportingJobWorker {
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
	prepare                  func(context.Context) error
	failure                  *RenderFailure
	declaredLimits           map[string]json.RawMessage
	kind                     string
	kindErr                  error
	writeErr                 error
	snapshotID               uuid.UUID
	releaseID                uuid.UUID
	renderFailedReleaseID    uuid.UUID
	completeSnapshotCalls    int
	releasePayloadCalls      int
	renderFailedReleaseCalls int
	completeReleaseCalls     int
	completePreviewCalls     int
}

func (s *fakeReportingJobStore) ReportingJobKind(ctx context.Context, _ jobs.Execution) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.kind, s.kindErr
}

func (s *fakeReportingJobStore) PrepareSnapshotJob(ctx context.Context, _ jobs.Execution) error {
	return ctx.Err()
}

func (s *fakeReportingJobStore) SnapshotCandidateForJob(ctx context.Context, _ jobs.Execution) (SnapshotCandidate, error) {
	if s.prepare != nil {
		return SnapshotCandidate{}, s.prepare(ctx)
	}
	return SnapshotCandidate{}, ctx.Err()
}

func (s *fakeReportingJobStore) CompleteSnapshotCreateJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, snapshotID uuid.UUID, _ SnapshotCandidate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.completeSnapshotCalls++
	if s.writeErr != nil {
		return s.writeErr
	}
	s.snapshotID = snapshotID
	return nil
}

func (s *fakeReportingJobStore) ReleasePayloadForJob(ctx context.Context, _ jobs.Execution) (releaseCreateJobPayload, error) {
	if err := ctx.Err(); err != nil {
		return releaseCreateJobPayload{}, err
	}
	s.releasePayloadCalls++
	if s.prepare != nil {
		if err := s.prepare(ctx); err != nil {
			return releaseCreateJobPayload{}, err
		}
	}
	snapshot, _, err := BuildSnapshotModel(IncidentMetadataSnapshot{
		ID: "00000000-0000-0000-0000-000000000311", Title: "Synthetic worker fixture", Status: "active",
	}, "00000000-0000-0000-0000-000000000312", time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC), "fixture-boundary", fixtureSnapshotSource(nil))
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	model, _, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	modelJSON, _ := canonicalJSON(model)
	profile, profileSHA, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	return releaseCreateJobPayload{renderInput: renderInput{
		SourceFacts: fixtureSourceFactsForContent(model.SnapshotContent), SourceFactsSHA256: fixtureSourceFactsDigest(model.SnapshotContent),
		TemplateContract:        TemplateContract{DeclaredLimits: s.declaredLimits},
		ActorUserID:             "00000000-0000-0000-0000-000000000313",
		IncidentID:              model.IncidentID,
		SnapshotID:              model.SnapshotID,
		ExportModel:             model,
		ReleaseScope:            model.ReleaseScope,
		ExportModelSHA256:       reportingObjectDigest(ExportModelSchemaID, modelJSON),
		RedactionProfileID:      InternalRedactionProfileID,
		RedactionProfileVersion: DefaultRedactionProfileVersion,
		RedactionProfile:        profile,
		RedactionProfileSHA256:  profileSHA,
	},
	}, nil
}

func (s *fakeReportingJobStore) CompositionPreviewPayloadForJob(ctx context.Context, _ jobs.Execution) (compositionPreviewJobPayload, error) {
	payload, err := s.ReleasePayloadForJob(ctx, jobs.Execution{})
	return compositionPreviewJobPayload{
		PreviewAttemptID: uuid.MustParse("00000000-0000-0000-0000-000000000314"),
		renderInput:      payload.renderInput,
	}, err
}

func (s *fakeReportingJobStore) CompleteRenderFailureJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, failure RenderFailure, _ time.Time) (*uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.failure = &failure
	if s.kind != reportingJobKindReleaseCreate {
		return nil, nil
	}
	s.renderFailedReleaseCalls++
	if s.renderFailedReleaseID == uuid.Nil {
		s.renderFailedReleaseID = uuid.MustParse("00000000-0000-0000-0000-000000000306")
	}
	return &s.renderFailedReleaseID, nil
}

func (s *fakeReportingJobStore) CompleteReleaseCreateJobTx(ctx context.Context, _ pgx.Tx, _ uuid.UUID, releaseID uuid.UUID, _ RenderCandidate, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.completeReleaseCalls++
	s.releaseID = releaseID
	return nil
}

func (s *fakeReportingJobStore) CompleteCompositionPreviewJobTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, RenderCandidate, time.Time) error {
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
	if f.manager.completionErr != nil {
		return jobs.Resource{}, f.manager.completionErr
	}
	f.manager.renderTimedOut = request.TimedOut
	if request.Mutate != nil {
		if err := request.Mutate(ctx, nil); err != nil {
			return jobs.Resource{}, err
		}
	}
	return f.manager.CompleteFailed(ctx, request.Execution, request.Completion)
}

func (f fakeReportingJobFinalizer) FinalizeReportingJobCancellation(ctx context.Context, request JobCancellationFinalization) (jobs.Resource, error) {
	if f.manager.completionErr != nil {
		return jobs.Resource{}, f.manager.completionErr
	}
	if request.Mutate != nil {
		if err := request.Mutate(ctx, nil); err != nil {
			return jobs.Resource{}, err
		}
	}
	return f.manager.CompleteCanceled(ctx, request.Execution, request.Completion)
}

type fakeReportingJobManager struct {
	observe        func() error
	renderTimedOut bool
	status         string
	observeErr     error
	observeCalls   int
	failObserveAt  int
	progressErr    error
	completionErr  error
	failed         *jobs.FailureCompletion
	succeeded      *jobs.SuccessCompletion
	canceled       *jobs.CancellationCompletion
}

func (m *fakeReportingJobManager) Get(ctx context.Context, _ uuid.UUID) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	return jobs.Resource{Status: m.status}, nil
}

func (m *fakeReportingJobManager) ObserveExecution(ctx context.Context, _ jobs.Execution) (jobs.Resource, error) {
	if m.observe != nil {
		if err := m.observe(); err != nil {
			return jobs.Resource{}, err
		}
	}
	m.observeCalls++
	if m.observeErr != nil && (m.failObserveAt == 0 || m.observeCalls == m.failObserveAt) {
		return jobs.Resource{}, m.observeErr
	}
	return m.Get(ctx, uuid.Nil)
}

func (m *fakeReportingJobManager) UpdateProgress(ctx context.Context, _ jobs.Execution, progress jobs.Progress, message *string) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	if m.progressErr != nil {
		return jobs.Resource{}, m.progressErr
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
	if m.completionErr != nil {
		return jobs.Resource{}, m.completionErr
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
	if m.completionErr != nil {
		return jobs.Resource{}, m.completionErr
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

func (blockingReleaseRenderer) Render(ctx context.Context, _ renderInput) (RenderedRelease, string, error) {
	<-ctx.Done()
	return RenderedRelease{}, "", ctx.Err()
}

type failingReleaseRenderer struct {
	reasonCode string
}

func (r failingReleaseRenderer) Render(context.Context, renderInput) (RenderedRelease, string, error) {
	return RenderedRelease{
		Profile:       RedactionProfile{ProfileID: ExternalRedactionProfileID, Version: DefaultRedactionProfileVersion},
		ProfileSHA256: "profile-sha",
	}, r.reasonCode, newRenderValidationError("remote_asset_reference", r.reasonCode)
}

type succeedingReleaseRenderer struct {
	releaseID uuid.UUID
}

func (succeedingReleaseRenderer) Render(context.Context, renderInput) (RenderedRelease, string, error) {
	return RenderedRelease{
		Profile:       RedactionProfile{ProfileID: ExternalRedactionProfileID, Version: DefaultRedactionProfileVersion},
		ProfileSHA256: "profile-sha",
	}, "", nil
}

// The handler returns persistence errors to the shared Jobs supervisor. It must
// never infer a rollback, fabricate a second terminal write, or expose raw errors.
func TestReportingJobWorkerExecutionBoundary(t *testing.T) {
	persistenceErr := errors.New("private persistence detail /incident/source")
	for _, kind := range []string{reportingJobKindReleaseCreate, reportingJobKindCompositionPreview} {
		for _, outcome := range []string{"success", "failure", "cleanup failure"} {
			t.Run(kind+"/cancellation at "+outcome+" commit", func(t *testing.T) {
				manager := &fakeReportingJobManager{status: jobs.StatusRunning}
				store := &fakeReportingJobStore{kind: kind}
				worker := newFakeReportingJobWorker(store, manager, nil)
				if outcome != "success" {
					worker.renderer = releaseRendererFunc(func(context.Context, renderInput) (RenderedRelease, string, error) {
						if outcome == "cleanup failure" {
							return RenderedRelease{}, "partial_output_cleanup_failed", errors.Join(newRenderValidationError("render_timeout", "render_timeout"), renderprocess.ErrCleanup)
						}
						return RenderedRelease{}, "remote_asset_reference_invalid", newRenderValidationError("remote_asset_reference", "remote_asset_reference_invalid")
					})
				}
				finalizer := &cancellationRaceReportingFinalizer{fakeReportingJobFinalizer: fakeReportingJobFinalizer{manager: manager}}
				worker.jobFinalizer = finalizer
				if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
					t.Fatal(err)
				}
				want := "render_canceled"
				if outcome == "cleanup failure" {
					want = "partial_output_cleanup_failed"
				}
				if finalizer.rejected != 1 || manager.status != jobs.StatusCanceled || manager.failed != nil || store.failure == nil || store.failure.ReasonCode != want || store.completeReleaseCalls != 0 || store.completePreviewCalls != 0 {
					t.Fatal("cancellation race lost owner diagnostics or published output")
				}
			})
		}
	}
	t.Run("execution lost before payload read", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, kindErr: jobs.ErrExecutionLost}
		worker := newFakeReportingJobWorker(store, manager, nil)
		if err := worker.Handle(context.Background(), jobs.Execution{}); !errors.Is(err, jobs.ErrExecutionLost) {
			t.Fatal(err)
		}
		if manager.failed != nil || manager.canceled != nil || store.releasePayloadCalls != 0 {
			t.Fatal("lost execution continued or completed")
		}
	})
	t.Run("cancellation between checkpoint and payload read", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, kindErr: jobs.ErrCancellationRequested}
		worker := newFakeReportingJobWorker(store, manager, nil)
		if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
			t.Fatal(err)
		}
		if manager.failed != nil || manager.canceled == nil || store.releasePayloadCalls != 0 {
			t.Fatal("cancellation became failure or continued reading")
		}
	})

	for _, kind := range []string{reportingJobKindSnapshotCreate, reportingJobKindReleaseCreate, reportingJobKindCompositionPreview} {
		for _, renderFails := range []bool{false, true} {
			if renderFails && kind != reportingJobKindReleaseCreate {
				continue
			}
			name := kind + "/success finalizer"
			if renderFails {
				name = kind + "/failure finalizer"
			}
			t.Run(name, func(t *testing.T) {
				manager := &fakeReportingJobManager{status: jobs.StatusRunning}
				store := &fakeReportingJobStore{kind: kind}
				worker := newFakeReportingJobWorker(store, manager, nil)
				if renderFails {
					worker.renderer = failingReleaseRenderer{reasonCode: "remote_asset_reference_invalid"}
				}
				finalizer := &faultReportingJobFinalizer{err: persistenceErr}
				worker.jobFinalizer = finalizer
				if err := worker.Handle(context.Background(), jobs.Execution{}); !errors.Is(err, persistenceErr) {
					t.Fatalf("finalizer error lost: %v", err)
				}
				if finalizer.calls != 1 || manager.status != jobs.StatusRunning || manager.failed != nil || manager.succeeded != nil {
					t.Fatal("handler retried or invented a terminal outcome")
				}
			})
		}
	}
	t.Run("mutation error survives finalizer", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindSnapshotCreate, writeErr: persistenceErr}
		worker := newFakeReportingJobWorker(store, manager, nil)
		if err := worker.Handle(context.Background(), jobs.Execution{}); !errors.Is(err, persistenceErr) || store.completeSnapshotCalls != 1 || manager.status != jobs.StatusRunning {
			t.Fatalf("mutation failure was not returned unchanged: %v", err)
		}
	})
	for _, stage := range []string{"initial observation", "progress", "final observation", "failure completion", "cancellation completion"} {
		t.Run(stage, func(t *testing.T) {
			manager := &fakeReportingJobManager{status: jobs.StatusRunning}
			store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
			worker := newFakeReportingJobWorker(store, manager, nil)
			switch stage {
			case "initial observation":
				manager.observeErr = persistenceErr
			case "progress":
				manager.progressErr = persistenceErr
			case "final observation":
				manager.observeErr, manager.failObserveAt = persistenceErr, 2
			case "failure completion":
				store.kindErr, manager.completionErr = persistenceErr, persistenceErr
			case "cancellation completion":
				manager.status, manager.completionErr = jobs.StatusCancelRequested, persistenceErr
			}
			if err := worker.Handle(context.Background(), jobs.Execution{}); !errors.Is(err, persistenceErr) {
				t.Fatalf("%s error lost: %v", stage, err)
			}
			if store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 0 || manager.failed != nil || manager.canceled != nil {
				t.Fatal("failed boundary allowed a terminal write")
			}
			if stage != "final observation" && store.releasePayloadCalls != 0 {
				t.Fatal("execution continued after failed boundary")
			}
		})
	}
	for _, duringRender := range []bool{false, true} {
		name := "shutdown before work"
		if duringRender {
			name = "shutdown during render"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			manager := &fakeReportingJobManager{status: jobs.StatusRunning}
			store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
			worker := newFakeReportingJobWorker(store, manager, nil)
			if duringRender {
				worker.renderer = releaseRendererFunc(func(context.Context, renderInput) (RenderedRelease, string, error) {
					cancel()
					return RenderedRelease{}, "", context.Canceled
				})
			} else {
				cancel()
			}
			if err := worker.Handle(ctx, jobs.Execution{}); !errors.Is(err, context.Canceled) {
				t.Fatalf("shutdown error = %v", err)
			}
			if manager.status != jobs.StatusRunning || manager.failed != nil || store.completeReleaseCalls != 0 || store.renderFailedReleaseCalls != 0 {
				t.Fatal("handler replaced supervisor shutdown with a terminal write")
			}
		})
	}
	t.Run("cancellation after rendering wins", func(t *testing.T) {
		manager := &fakeReportingJobManager{status: jobs.StatusRunning}
		store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
		worker := newFakeReportingJobWorker(store, manager, releaseRendererFunc(func(context.Context, renderInput) (RenderedRelease, string, error) {
			manager.status = jobs.StatusCancelRequested
			return RenderedRelease{}, "remote_asset_reference_invalid", persistenceErr
		}))
		if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
			t.Fatal(err)
		}
		if manager.status != jobs.StatusCanceled || manager.failed != nil || store.renderFailedReleaseCalls != 1 || store.completeReleaseCalls != 0 || store.failure.ReasonCode != "render_canceled" {
			t.Fatal("render outcome replaced cancellation")
		}
	})
	for _, kind := range []string{reportingJobKindSnapshotCreate, reportingJobKindCompositionPreview} {
		t.Run(kind+"/safe failure summary", func(t *testing.T) {
			manager := &fakeReportingJobManager{status: jobs.StatusRunning}
			store := &fakeReportingJobStore{kind: kind}
			worker := newFakeReportingJobWorker(store, manager, failingReleaseRenderer{reasonCode: "remote_asset_reference_invalid"})
			if kind == reportingJobKindSnapshotCreate {
				store.kindErr = persistenceErr
			}
			if err := worker.Handle(context.Background(), jobs.Execution{}); err != nil {
				t.Fatal(err)
			}
			want := "Reporting job failed."
			if kind == reportingJobKindCompositionPreview {
				want = reportingJobRenderFailedMessage
			}
			if manager.failed == nil || manager.failed.ErrorSummary.Message != want {
				t.Fatal("raw error crossed public Jobs boundary")
			}
			if kind == reportingJobKindCompositionPreview {
				if len(manager.failed.ErrorSummary.Details) != 1 || manager.failed.ErrorSummary.Details["reason_code"] != "remote_asset_reference_invalid" {
					t.Fatal("preview must retain only the closed failure identity")
				}
			} else if len(manager.failed.ErrorSummary.Details) != 0 {
				t.Fatal("raw persistence detail crossed public Jobs boundary")
			}
		})
	}
}

type faultReportingJobFinalizer struct {
	err   error
	calls int
}

// Models cancellation winning the Jobs row lock before an uncommitted owner
// mutation. The succeeding cancellation uses the ordinary atomic finalizer.
type cancellationRaceReportingFinalizer struct {
	fakeReportingJobFinalizer
	rejected int
}

func (f *cancellationRaceReportingFinalizer) reject() (jobs.Resource, error) {
	f.rejected++
	f.manager.status = jobs.StatusCancelRequested
	return jobs.Resource{}, jobs.ErrCancellationRequested
}
func (f *cancellationRaceReportingFinalizer) FinalizeReportingJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error) {
	return f.reject()
}
func (f *cancellationRaceReportingFinalizer) FinalizeReportingJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	return f.reject()
}

func (f *faultReportingJobFinalizer) FinalizeReportingJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error) {
	f.calls++
	return jobs.Resource{}, f.err
}
func (f *faultReportingJobFinalizer) FinalizeReportingJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	f.calls++
	return jobs.Resource{}, f.err
}

func (f *faultReportingJobFinalizer) FinalizeReportingJobCancellation(context.Context, JobCancellationFinalization) (jobs.Resource, error) {
	f.calls++
	return jobs.Resource{}, f.err
}

type releaseRendererFunc func(context.Context, renderInput) (RenderedRelease, string, error)

func (f releaseRendererFunc) Render(ctx context.Context, payload renderInput) (RenderedRelease, string, error) {
	return f(ctx, payload)
}

func (s *fakeReportingJobStore) PrepareRenderCandidate(ctx context.Context, execution jobs.Execution, input renderInput, rendered RenderedRelease) (RenderCandidate, error) {
	return RenderCandidate{jobID: execution.JobID(), rendered: rendered}, ctx.Err()
}
