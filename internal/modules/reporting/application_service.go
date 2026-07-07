package reporting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type ApplicationService struct {
	store          *Store
	incidentAccess incidents.Access
	jobManager     *jobs.Manager
	now            func() time.Time
}

func NewApplicationService(store *Store, incidentAccess incidents.Access, jobManager *jobs.Manager, now func() time.Time) *ApplicationService {
	return &ApplicationService{
		store:          store,
		incidentAccess: incidentAccess,
		jobManager:     jobManager,
		now:            now,
	}
}

func (s *ApplicationService) CreateSnapshot(ctx context.Context, actorUserID uuid.UUID, request CreateSnapshotRequest) (jobs.Resource, *httpapi.APIError) {
	if _, apiErr := s.requireIncidentRole(ctx, request.IncidentID, actorUserID, "editor", "reviewer", "admin"); apiErr != nil {
		return jobs.Resource{}, apiErr
	}
	result, err := s.store.CreateSnapshot(ctx, CreateSnapshotParams{
		ActorUserID: actorUserID,
		Request:     request,
		Now:         s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return jobs.Resource{}, clientTxnConflict(request.ClientTxnID)
	}
	var boundaryConflict *SnapshotBoundaryConflictError
	if errors.As(err, &boundaryConflict) {
		return jobs.Resource{}, snapshotBoundaryMismatch(boundaryConflict.Expected, boundaryConflict.Actual)
	}
	if err != nil {
		return jobs.Resource{}, internalAPIError(err)
	}
	s.dispatchReportingJob(result.Job.JobID)
	return result.Job, nil
}

func (s *ApplicationService) GetSnapshot(ctx context.Context, actorUserID uuid.UUID, snapshotID uuid.UUID) (map[string]any, *httpapi.APIError) {
	resource, incidentID, err := s.store.GetSnapshot(ctx, snapshotID)
	if errors.Is(err, ErrNotFound) {
		return nil, &httpapi.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return nil, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentMembership(ctx, incidentID, actorUserID); apiErr != nil {
		return nil, snapshotVisibilityError(apiErr)
	}
	return resource, nil
}

func (s *ApplicationService) CreateRelease(ctx context.Context, actorUserID uuid.UUID, request CreateReleaseRequest) (jobs.Resource, *httpapi.APIError) {
	snapshot, model, err := s.store.GetSnapshotForRender(ctx, request.SnapshotID)
	if errors.Is(err, ErrNotFound) {
		return jobs.Resource{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}}
	}
	var unsupportedDerivation *UnsupportedSnapshotDerivationError
	if errors.As(err, &unsupportedDerivation) {
		return jobs.Resource{}, invalidReleaseRequest("derivation_version", "unsupported_derivation_version")
	}
	if err != nil {
		return jobs.Resource{}, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentRole(ctx, snapshot.IncidentID, actorUserID, "editor", "reviewer", "admin"); apiErr != nil {
		return jobs.Resource{}, snapshotVisibilityError(apiErr)
	}
	templateContract, apiErr := validateCreateReleaseRequestSemantics(request)
	if apiErr != nil {
		return jobs.Resource{}, apiErr
	}
	if apiErr := validateCreateReleaseRecipientPartitions(request, model); apiErr != nil {
		return jobs.Resource{}, apiErr
	}
	result, err := s.store.CreateRelease(ctx, CreateReleaseParams{
		ActorUserID:      actorUserID,
		Request:          request,
		TemplateContract: templateContract,
		Now:              s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return jobs.Resource{}, clientTxnConflict(request.ClientTxnID)
	}
	var invalidRelease *InvalidReleaseRequestError
	if errors.As(err, &invalidRelease) {
		return jobs.Resource{}, invalidReleaseRequest(invalidRelease.Field, invalidRelease.ReasonCode)
	}
	if err != nil {
		return jobs.Resource{}, internalAPIError(err)
	}
	s.dispatchReportingJob(result.Job.JobID)
	return result.Job, nil
}

func (s *ApplicationService) GetRelease(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID) (map[string]any, *httpapi.APIError) {
	resource, incidentID, err := s.store.GetRelease(ctx, releaseID)
	if errors.Is(err, ErrNotFound) {
		return nil, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return nil, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentMembership(ctx, incidentID, actorUserID); apiErr != nil {
		return nil, releaseVisibilityError(apiErr)
	}
	return resource, nil
}

func (s *ApplicationService) ApproveRelease(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID, request ReleaseActionRequest) (map[string]any, *httpapi.APIError) {
	_, incidentID, err := s.store.GetRelease(ctx, releaseID)
	if errors.Is(err, ErrNotFound) {
		return nil, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return nil, internalAPIError(err)
	}
	membership, apiErr := s.requireIncidentMembership(ctx, incidentID, actorUserID)
	if apiErr != nil {
		return nil, releaseVisibilityError(apiErr)
	}
	result, err := s.store.ApproveRelease(ctx, ApproveReleaseParams{
		ActorUserID:       actorUserID,
		ActorIncidentRole: membership.Role,
		ReleaseID:         releaseID,
		Request:           request,
		Now:               s.now(),
	})
	return s.releaseActionPayload(request.ClientTxnID, result, err)
}

func (s *ApplicationService) PublishRelease(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID, request ReleaseActionRequest) (map[string]any, *httpapi.APIError) {
	return s.adminReleaseAction(ctx, actorUserID, releaseID, request, s.store.PublishRelease)
}

func (s *ApplicationService) InvalidateRelease(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID, request ReleaseActionRequest) (map[string]any, *httpapi.APIError) {
	return s.adminReleaseAction(ctx, actorUserID, releaseID, request, s.store.InvalidateRelease)
}

func (s *ApplicationService) adminReleaseAction(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID, request ReleaseActionRequest, action func(context.Context, ReleaseActionParams) (ReleaseActionResult, error)) (map[string]any, *httpapi.APIError) {
	_, incidentID, err := s.store.GetRelease(ctx, releaseID)
	if errors.Is(err, ErrNotFound) {
		return nil, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return nil, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentRole(ctx, incidentID, actorUserID, "admin"); apiErr != nil {
		return nil, releaseVisibilityError(apiErr)
	}
	result, err := action(ctx, ReleaseActionParams{
		ActorUserID: actorUserID,
		ReleaseID:   releaseID,
		Request:     request,
		Now:         s.now(),
	})
	return s.releaseActionPayload(request.ClientTxnID, result, err)
}

func (s *ApplicationService) releaseActionPayload(clientTxnID string, result ReleaseActionResult, err error) (map[string]any, *httpapi.APIError) {
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return nil, clientTxnConflict(clientTxnID)
	}
	var stateConflict *StateConflictError
	if errors.As(err, &stateConflict) {
		return nil, releaseStateConflict(stateConflict.ReasonCode)
	}
	var approvalRejected *ApprovalRejectedError
	if errors.As(err, &approvalRejected) {
		return nil, releaseApprovalRejected(approvalRejected.ReasonCode)
	}
	if err != nil {
		return nil, internalAPIError(err)
	}
	return result.Payload, nil
}

func (s *ApplicationService) dispatchReportingJob(jobID string) {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	go func() {
		time.Sleep(25 * time.Millisecond)
		s.executeReportingJob(context.Background(), parsed)
	}()
}

func (s *ApplicationService) executeReportingJob(ctx context.Context, jobID uuid.UUID) {
	kind, err := s.store.ReportingJobKind(ctx, jobID)
	if err != nil {
		s.failReportingJob(ctx, jobID, "internal_error", err)
		return
	}
	switch kind {
	case "snapshot_create":
		s.executeSnapshotCreateJob(ctx, jobID)
	case "release_create":
		s.executeReleaseCreateJob(ctx, jobID)
	default:
		s.failReportingJob(ctx, jobID, "internal_error", fmt.Errorf("unknown reporting job kind %q", kind))
	}
}

func (s *ApplicationService) executeSnapshotCreateJob(ctx context.Context, jobID uuid.UUID) {
	total := 1
	if !s.markReportingJobRunning(ctx, jobID, total) {
		return
	}
	if s.cancelReportingJobIfRequested(ctx, jobID, total) {
		return
	}
	snapshotID, err := s.store.CompleteSnapshotCreateJob(ctx, jobID)
	if err != nil {
		s.failReportingJob(ctx, jobID, "internal_error", err)
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:    "snapshot_created",
			Message: "Snapshot created.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "snapshot",
				ID:    snapshotID.String(),
				Route: "/api/v1/snapshots/" + snapshotID.String(),
			}},
		},
	})
}

func (s *ApplicationService) executeReleaseCreateJob(ctx context.Context, jobID uuid.UUID) {
	total := 1
	if !s.markReportingJobRunning(ctx, jobID, total) {
		return
	}
	if s.cancelReportingJobIfRequested(ctx, jobID, total) {
		return
	}
	payload, err := s.store.ReleasePayloadForJob(ctx, jobID)
	if err != nil {
		s.failReportingJob(ctx, jobID, "internal_error", err)
		return
	}
	rendered, reasonCode, renderErr := renderReleaseCandidate(payload.Request, payload.TemplateContract, payload.ExportModel, payload.ExportModelSHA256)
	if renderErr != nil {
		releaseID, err := s.store.CompleteReleaseRenderFailedJob(ctx, jobID, rendered.Profile, rendered.ProfileSHA256, reasonCode, s.now())
		if err != nil {
			s.failReportingJob(ctx, jobID, "internal_error", err)
			return
		}
		_, _ = s.jobManager.CompleteFailed(ctx, jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ErrorSummary: &jobs.ErrorSummary{
				Code:      "release_render_failed",
				Message:   "Release render failed.",
				Retryable: false,
				Details: map[string]any{
					"reason_code": reasonCode,
					"release_id":  releaseID.String(),
				},
			},
		})
		return
	}
	if s.cancelReportingJobIfRequested(ctx, jobID, total) {
		return
	}
	releaseID, err := s.store.CompleteReleaseCreateJob(ctx, jobID, rendered, s.now())
	if err != nil {
		s.failReportingJob(ctx, jobID, "internal_error", err)
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:    "release_created",
			Message: "Release rendered.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "release",
				ID:    releaseID.String(),
				Route: "/api/v1/releases/" + releaseID.String(),
			}},
		},
	})
}

func (s *ApplicationService) markReportingJobRunning(ctx context.Context, jobID uuid.UUID, total int) bool {
	resource, err := s.jobManager.Get(ctx, jobID)
	if err != nil {
		return false
	}
	switch resource.Status {
	case jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled:
		return false
	case jobs.StatusCancelRequested:
		_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 0, Total: &total},
		})
		return false
	}
	if _, err := s.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err != nil {
		resource, getErr := s.jobManager.Get(ctx, jobID)
		if getErr == nil && resource.Status == jobs.StatusCancelRequested {
			_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
				JobID:    jobID,
				Progress: jobs.Progress{Completed: 0, Total: &total},
			})
		}
		return false
	}
	return true
}

func (s *ApplicationService) cancelReportingJobIfRequested(ctx context.Context, jobID uuid.UUID, total int) bool {
	resource, err := s.jobManager.Get(ctx, jobID)
	if err != nil || resource.Status != jobs.StatusCancelRequested {
		return false
	}
	_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 0, Total: &total},
	})
	return true
}

func (s *ApplicationService) failReportingJob(ctx context.Context, jobID uuid.UUID, code string, err error) {
	total := 1
	_, _ = s.jobManager.CompleteFailed(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 0, Total: &total},
		ErrorSummary: &jobs.ErrorSummary{
			Code:      code,
			Message:   err.Error(),
			Retryable: false,
			Details:   map[string]any{},
		},
	})
}

func (s *ApplicationService) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *ApplicationService) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}
