package reporting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/google/uuid"
)

type Service struct {
	store          *Store
	incidentAccess incidents.Access
	authStore      *authn.Store
	jobManager     *jobs.Manager
	hub            *platformws.Hub
	keys           authn.MasterKeys
	now            func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimedIn(deps.ExtensionProfiles, ProfileID) {
			return nil
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/snapshots", service.handleSnapshotsCollection)
		mux.HandleFunc("/api/v1/snapshots/", service.handleSnapshotsMember)
		mux.HandleFunc("/api/v1/releases", service.handleReleasesCollection)
		mux.HandleFunc("/api/v1/releases/", service.handleReleasesMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		store:          NewStore(deps.Postgres),
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		jobManager:     deps.Jobs,
		hub:            deps.WSHub,
		keys:           keys,
		now:            now,
	}, nil
}

func (s *Service) handleSnapshotsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeCreateSnapshotRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreateSnapshot(r.Context(), CreateSnapshotParams{
		ActorUserID: principal.User.ID,
		Request:     request,
		Now:         s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	var boundaryConflict *SnapshotBoundaryConflictError
	if errors.As(err, &boundaryConflict) {
		writeAPIError(w, r, snapshotBoundaryMismatch(boundaryConflict.Expected, boundaryConflict.Actual))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.dispatchReportingJob(result.Job.JobID)
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleSnapshotsMember(w http.ResponseWriter, r *http.Request) {
	snapshotID, ok := parseCollectionMemberPath(r.URL.Path, "/api/v1/snapshots/")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, incidentID, err := s.store.GetSnapshot(r.Context(), snapshotID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, snapshotVisibilityError(apiErr))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleReleasesCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeCreateReleaseRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	snapshot, model, err := s.store.GetSnapshotForRender(r.Context(), request.SnapshotID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), snapshot.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, snapshotVisibilityError(apiErr))
		return
	}
	templateContract, apiErr := validateCreateReleaseRequestSemantics(request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := validateCreateReleaseRecipientPartitions(request, model); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreateRelease(r.Context(), CreateReleaseParams{
		ActorUserID:      principal.User.ID,
		Request:          request,
		TemplateContract: templateContract,
		Now:              s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	var invalidRelease *InvalidReleaseRequestError
	if errors.As(err, &invalidRelease) {
		writeAPIError(w, r, invalidReleaseRequest(invalidRelease.Field, invalidRelease.ReasonCode))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.dispatchReportingJob(result.Job.JobID)
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleReleasesMember(w http.ResponseWriter, r *http.Request) {
	route, ok := parseReleasePath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	stateChanging := route.Action != ""
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch route.Action {
	case "":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		resource, incidentID, err := s.store.GetRelease(r.Context(), route.ReleaseID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, releaseVisibilityError(apiErr))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
	case "approve":
		s.handleApproveRelease(w, r, principal, route.ReleaseID)
	case "publish":
		s.handlePublishRelease(w, r, principal, route.ReleaseID)
	case "invalidate":
		s.handleInvalidateRelease(w, r, principal, route.ReleaseID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Service) handleApproveRelease(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	release, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = release
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, releaseVisibilityError(apiErr))
		return
	}
	result, err := s.store.ApproveRelease(r.Context(), ApproveReleaseParams{
		ActorUserID:       principal.User.ID,
		ActorIncidentRole: membership.Role,
		ReleaseID:         releaseID,
		Request:           request,
		Now:               s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *Service) handlePublishRelease(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
		writeAPIError(w, r, releaseVisibilityError(apiErr))
		return
	}
	result, err := s.store.PublishRelease(r.Context(), ReleaseActionParams{
		ActorUserID: principal.User.ID,
		ReleaseID:   releaseID,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *Service) handleInvalidateRelease(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
		writeAPIError(w, r, releaseVisibilityError(apiErr))
		return
	}
	result, err := s.store.InvalidateRelease(r.Context(), ReleaseActionParams{
		ActorUserID: principal.User.ID,
		ReleaseID:   releaseID,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func renderReleaseCandidate(request CreateReleaseRequest, contract TemplateContract, model ExportModel, exportModelSHA string) (RenderedRelease, string, error) {
	profile, profileSHA, err := ResolveRedactionProfile(request.RedactionProfileID, request.RedactionProfileVersion, request.RecipientPartitionRefs)
	if errors.Is(err, ErrInvalidRedactionProfile) {
		return RenderedRelease{}, "invalid_redaction_profile", err
	}
	if err != nil {
		return RenderedRelease{}, "invalid_redaction_profile", err
	}
	partial := RenderedRelease{Profile: profile, ProfileSHA256: profileSHA}
	redaction, err := RedactExportModel(model, profile, profileSHA, exportModelSHA, request.ReleaseScope, request.RecipientPartitionRefs)
	if errors.Is(err, ErrInvalidRedactionProfile) {
		return partial, "invalid_redaction_profile", err
	}
	if errors.Is(err, ErrRedactionValidation) {
		return partial, "post_redaction_validation_failed", err
	}
	if err != nil {
		return partial, "post_redaction_validation_failed", err
	}
	partial.Redaction = redaction
	bundle, err := renderReportBundle(contract, request.OutputKind, redaction.Model, redaction.ManifestSHA256, request.ReleaseScope)
	if errors.Is(err, ErrUndeclaredTemplateBinding) {
		return partial, "undeclared_template_binding", err
	}
	if errors.Is(err, ErrMissingRequiredField) {
		return partial, "missing_required_field", err
	}
	if err != nil {
		return partial, "template_render_failed", err
	}
	manifestJSON, err := canonicalJSON(redaction.Manifest)
	if err != nil {
		return partial, "manifest_encoding_failed", err
	}
	return RenderedRelease{
		Profile:                 profile,
		ProfileSHA256:           profileSHA,
		Redaction:               redaction,
		OutputMediaType:         bundle.PrimaryMedia,
		OutputSHA256:            bundle.ManifestSHA256,
		RenderBundle:            bundle,
		RedactionManifestSHA256: redaction.ManifestSHA256,
		RedactionManifestJSON:   manifestJSON,
	}, "", nil
}

func (s *Service) writeActionResult(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, clientTxnID string, result ReleaseActionResult, err error) {
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
		return
	}
	var stateConflict *StateConflictError
	if errors.As(err, &stateConflict) {
		writeAPIError(w, r, releaseStateConflict(stateConflict.ReasonCode))
		return
	}
	var approvalRejected *ApprovalRejectedError
	if errors.As(err, &approvalRejected) {
		writeAPIError(w, r, releaseApprovalRejected(approvalRejected.ReasonCode))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) dispatchReportingJob(jobID string) {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	go func() {
		time.Sleep(25 * time.Millisecond)
		s.executeReportingJob(context.Background(), parsed)
	}()
}

func (s *Service) executeReportingJob(ctx context.Context, jobID uuid.UUID) {
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

func (s *Service) executeSnapshotCreateJob(ctx context.Context, jobID uuid.UUID) {
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

func (s *Service) executeReleaseCreateJob(ctx context.Context, jobID uuid.UUID) {
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

func (s *Service) markReportingJobRunning(ctx context.Context, jobID uuid.UUID, total int) bool {
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

func (s *Service) cancelReportingJobIfRequested(ctx context.Context, jobID uuid.UUID, total int) bool {
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

func (s *Service) failReportingJob(ctx context.Context, jobID uuid.UUID, code string, err error) {
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

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func snapshotVisibilityError(apiErr *httpapi.APIError) *httpapi.APIError {
	if apiErr != nil && apiErr.Status == http.StatusNotFound && apiErr.Code == "incident_not_found" {
		return &httpapi.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}}
	}
	return apiErr
}

func releaseVisibilityError(apiErr *httpapi.APIError) *httpapi.APIError {
	if apiErr != nil && apiErr.Status == http.StatusNotFound && apiErr.Code == "incident_not_found" {
		return &httpapi.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}}
	}
	return apiErr
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

type releaseRoute struct {
	ReleaseID uuid.UUID
	Action    string
}

func parseReleasePath(path string) (releaseRoute, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/releases/")
	if rest == path || rest == "" {
		return releaseRoute{}, false
	}
	parts := strings.Split(rest, "/")
	releaseID, err := uuid.Parse(parts[0])
	if err != nil {
		return releaseRoute{}, false
	}
	if len(parts) == 1 {
		return releaseRoute{ReleaseID: releaseID}, true
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "approve", "publish", "invalidate":
			return releaseRoute{ReleaseID: releaseID, Action: parts[1]}, true
		}
	}
	return releaseRoute{}, false
}

func parseCollectionMemberPath(path string, prefix string) (uuid.UUID, bool) {
	rest := strings.TrimPrefix(path, prefix)
	if rest == path || rest == "" || strings.Contains(rest, "/") {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(rest)
	if err != nil {
		return uuid.UUID{}, false
	}
	return id, true
}
