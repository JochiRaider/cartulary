package reporting

import (
	"context"
	"errors"
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
	jobNotifier    reportingJobRunner
	now            func() time.Time
}

func NewApplicationService(
	store *Store,
	incidentAccess incidents.Access,
	jobManager reportingJobManager,
	jobRunner reportingJobRunner,
	jobFinalizer JobSuccessFinalizer,
	renderExportInvoker RenderExportInvoker,
	now func() time.Time,
) (*ApplicationService, error) {
	if store == nil || jobManager == nil || jobRunner == nil || jobFinalizer == nil || renderExportInvoker == nil {
		return nil, errors.New("reporting application service requires store, jobs, runner, finalizer, and admitted render participant")
	}
	service := &ApplicationService{
		store:          store,
		incidentAccess: incidentAccess,
		jobNotifier:    jobRunner,
		now:            now,
	}
	worker := newReportingJobWorker(store, jobManager, jobFinalizer, renderExportInvoker, now)
	if err := jobRunner.RegisterHandler(JobWorkerKind, worker.Handle); err != nil {
		return nil, err
	}
	return service, nil
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
	s.notifyReportingJob(result.JobID)
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
	s.notifyReportingJob(result.JobID)
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

func (s *ApplicationService) notifyReportingJob(jobID uuid.UUID) {
	if s.jobNotifier != nil {
		s.jobNotifier.Notify(jobID)
	}
}

func (s *ApplicationService) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *ApplicationService) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}
