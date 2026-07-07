package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestReportingJobWorkerCancelsBeforeWork(t *testing.T) {
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000301")
	manager := &fakeReportingJobManager{status: jobs.StatusCancelRequested}
	store := &fakeReportingJobStore{kind: reportingJobKindSnapshotCreate}
	worker := newFakeReportingJobWorker(store, manager, nil)

	worker.Run(context.Background(), jobID)

	if manager.status != jobs.StatusCanceled {
		t.Fatalf("job status = %q want canceled", manager.status)
	}
	if store.completeSnapshotCalls != 0 {
		t.Fatalf("snapshot completion ran after cancel request: %d", store.completeSnapshotCalls)
	}
	if manager.canceled == nil || manager.canceled.ResultSummary == nil || manager.canceled.ResultSummary.Code != "job_canceled" {
		t.Fatalf("cancel summary not recorded: %#v", manager.canceled)
	}
}

func TestReportingJobWorkerTimeoutFailsBeforePersistingRelease(t *testing.T) {
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000302")
	manager := &fakeReportingJobManager{status: jobs.StatusQueued}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate}
	worker := newFakeReportingJobWorker(store, manager, blockingReleaseRenderer{})
	worker.timeout = time.Millisecond

	worker.Run(context.Background(), jobID)

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if manager.failed == nil || manager.failed.ErrorSummary == nil {
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
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000303")
	releaseID := uuid.MustParse("00000000-0000-0000-0000-000000000304")
	manager := &fakeReportingJobManager{status: jobs.StatusQueued}
	store := &fakeReportingJobStore{kind: reportingJobKindReleaseCreate, renderFailedReleaseID: releaseID}
	worker := newFakeReportingJobWorker(store, manager, failingReleaseRenderer{reasonCode: "remote_asset_runtime_ref"})

	worker.Run(context.Background(), jobID)

	if manager.status != jobs.StatusFailed {
		t.Fatalf("job status = %q want failed", manager.status)
	}
	if store.renderFailedReleaseCalls != 1 || store.completeReleaseCalls != 0 {
		t.Fatalf("unexpected release persistence calls: render_failed=%d complete=%d", store.renderFailedReleaseCalls, store.completeReleaseCalls)
	}
	if manager.failed == nil || manager.failed.ErrorSummary == nil {
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

func newFakeReportingJobWorker(store *fakeReportingJobStore, manager *fakeReportingJobManager, renderer releaseCandidateRenderer) *reportingJobWorker {
	if renderer == nil {
		renderer = succeedingReleaseRenderer{releaseID: uuid.MustParse("00000000-0000-0000-0000-000000000399")}
	}
	return &reportingJobWorker{
		store:           store,
		jobManager:      manager,
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
}

func (s *fakeReportingJobStore) ReportingJobKind(ctx context.Context, _ uuid.UUID) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.kind, nil
}

func (s *fakeReportingJobStore) CompleteSnapshotCreateJob(ctx context.Context, _ uuid.UUID) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.UUID{}, err
	}
	s.completeSnapshotCalls++
	if s.snapshotID == uuid.Nil {
		s.snapshotID = uuid.MustParse("00000000-0000-0000-0000-000000000305")
	}
	return s.snapshotID, nil
}

func (s *fakeReportingJobStore) ReleasePayloadForJob(ctx context.Context, _ uuid.UUID) (releaseCreateJobPayload, error) {
	if err := ctx.Err(); err != nil {
		return releaseCreateJobPayload{}, err
	}
	s.releasePayloadCalls++
	return releaseCreateJobPayload{}, nil
}

func (s *fakeReportingJobStore) CompleteReleaseRenderFailedJob(ctx context.Context, _ uuid.UUID, _ RedactionProfile, _ string, _ string, _ time.Time) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.UUID{}, err
	}
	s.renderFailedReleaseCalls++
	if s.renderFailedReleaseID == uuid.Nil {
		s.renderFailedReleaseID = uuid.MustParse("00000000-0000-0000-0000-000000000306")
	}
	return s.renderFailedReleaseID, nil
}

func (s *fakeReportingJobStore) CompleteReleaseCreateJob(ctx context.Context, _ uuid.UUID, _ RenderedRelease, _ time.Time) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.UUID{}, err
	}
	s.completeReleaseCalls++
	if s.releaseID == uuid.Nil {
		s.releaseID = uuid.MustParse("00000000-0000-0000-0000-000000000307")
	}
	return s.releaseID, nil
}

type fakeReportingJobManager struct {
	status    string
	failed    *jobs.TransitionParams
	succeeded *jobs.TransitionParams
	canceled  *jobs.TransitionParams
}

func (m *fakeReportingJobManager) Get(ctx context.Context, _ uuid.UUID) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	return jobs.Resource{Status: m.status}, nil
}

func (m *fakeReportingJobManager) MarkRunning(ctx context.Context, _ uuid.UUID, progress jobs.Progress, message *string) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	if m.status == jobs.StatusCancelRequested {
		return jobs.Resource{}, jobs.ErrInvalidTransition
	}
	m.status = jobs.StatusRunning
	return jobs.Resource{Status: m.status, Progress: progress, Message: message}, nil
}

func (m *fakeReportingJobManager) CompleteSucceeded(ctx context.Context, params jobs.TransitionParams) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusSucceeded
	copied := params
	m.succeeded = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ResultSummary: params.ResultSummary}, nil
}

func (m *fakeReportingJobManager) CompleteFailed(ctx context.Context, params jobs.TransitionParams) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusFailed
	copied := params
	m.failed = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ErrorSummary: params.ErrorSummary}, nil
}

func (m *fakeReportingJobManager) CompleteCanceled(ctx context.Context, params jobs.TransitionParams) (jobs.Resource, error) {
	if err := ctx.Err(); err != nil {
		return jobs.Resource{}, err
	}
	m.status = jobs.StatusCanceled
	copied := params
	if copied.ResultSummary == nil {
		copied.ResultSummary = &jobs.ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	m.canceled = &copied
	return jobs.Resource{Status: m.status, Progress: params.Progress, ResultSummary: copied.ResultSummary}, nil
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
