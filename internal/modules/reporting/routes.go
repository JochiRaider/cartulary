package reporting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store         *Store
	incidentStore *incidents.Store
	authStore     *authn.Store
	jobManager    *jobs.Manager
	hub           *platformws.Hub
	keys          authn.MasterKeys
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimed(ProfileID) {
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
		store:         NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		jobManager:    deps.Jobs,
		hub:           deps.WSHub,
		keys:          keys,
		now:           now,
	}, nil
}

func (s *Service) handleSnapshotsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
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
	if !result.Replayed {
		s.completeSnapshotJob(r.Context(), result)
	}
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
	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, incidentID, err := s.store.GetSnapshot(r.Context(), snapshotID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeCreateReleaseRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if request.TemplateID != DefaultTemplateID || request.TemplateVersion != DefaultTemplateVersion {
		writeAPIError(w, r, unsupportedTemplateError(request.TemplateID, request.TemplateVersion))
		return
	}
	snapshot, model, err := s.store.GetSnapshotForRender(r.Context(), request.SnapshotID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "snapshot_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), snapshot.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	rendered, apiErr := renderReleaseCandidate(request, model, snapshot.ExportModelSHA256)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreateRelease(r.Context(), CreateReleaseParams{
		ActorUserID: principal.User.ID,
		Request:     request,
		Rendered:    rendered,
		Now:         s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.completeReleaseJob(r.Context(), result)
	}
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
	principal, apiErr := s.authenticateSessionRequest(r, stateChanging)
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
		if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		resource, incidentID, err := s.store.GetRelease(r.Context(), route.ReleaseID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
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

func (s *Service) handleApproveRelease(w http.ResponseWriter, r *http.Request, principal auth.SessionPrincipal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	release, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = release
	membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "reviewer", "admin")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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

func (s *Service) handlePublishRelease(w http.ResponseWriter, r *http.Request, principal auth.SessionPrincipal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
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

func (s *Service) handleInvalidateRelease(w http.ResponseWriter, r *http.Request, principal auth.SessionPrincipal, releaseID uuid.UUID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	request, apiErr := DecodeReleaseActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetRelease(r.Context(), releaseID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &auth.APIError{Status: http.StatusNotFound, Code: "release_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
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

func renderReleaseCandidate(request CreateReleaseRequest, model ExportModel, exportModelSHA string) (RenderedRelease, *auth.APIError) {
	profile, profileSHA, err := ResolveRedactionProfile(request.RedactionProfileID, request.RedactionProfileVersion)
	if errors.Is(err, ErrInvalidRedactionProfile) {
		return RenderedRelease{}, unsupportedRedactionProfileError(request.RedactionProfileID, request.RedactionProfileVersion)
	}
	if err != nil {
		return RenderedRelease{}, releaseRenderFailed("invalid_redaction_profile", err)
	}
	redaction, err := RedactExportModel(model, profile, profileSHA, exportModelSHA, request.ReleaseScope)
	if errors.Is(err, ErrInvalidRedactionProfile) {
		return RenderedRelease{}, releaseRenderFailed("invalid_redaction_profile", err)
	}
	if errors.Is(err, ErrRedactionValidation) {
		return RenderedRelease{}, releaseRenderFailed("post_redaction_validation_failed", err)
	}
	if err != nil {
		return RenderedRelease{}, releaseRenderFailed("redaction_failed", err)
	}
	output, mediaType, err := RenderOutput(request.OutputKind, redaction.Model, redaction.Manifest, request.ReleaseScope)
	if err != nil {
		return RenderedRelease{}, releaseRenderFailed("template_render_failed", err)
	}
	manifestJSON, err := canonicalJSON(redaction.Manifest)
	if err != nil {
		return RenderedRelease{}, releaseRenderFailed("manifest_encoding_failed", err)
	}
	return RenderedRelease{
		Profile:                 profile,
		ProfileSHA256:           profileSHA,
		Redaction:               redaction,
		Output:                  output,
		OutputMediaType:         mediaType,
		OutputSHA256:            hashHex(output),
		RedactionManifestSHA256: redaction.ManifestSHA256,
		RedactionManifestJSON:   manifestJSON,
	}, nil
}

func (s *Service) writeActionResult(w http.ResponseWriter, r *http.Request, principal *auth.SessionPrincipal, clientTxnID string, result ReleaseActionResult, err error) {
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

func (s *Service) completeSnapshotJob(ctx context.Context, result CreateSnapshotResult) {
	jobID, err := uuid.Parse(result.Job.JobID)
	if err != nil {
		return
	}
	total := 1
	_, _ = s.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil)
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:    "snapshot_created",
			Message: "Snapshot created.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "snapshot",
				ID:    result.SnapshotID.String(),
				Route: "/api/v1/snapshots/" + result.SnapshotID.String(),
			}},
		},
	})
}

func (s *Service) completeReleaseJob(ctx context.Context, result CreateReleaseResult) {
	jobID, err := uuid.Parse(result.Job.JobID)
	if err != nil {
		return
	}
	total := 1
	_, _ = s.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil)
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:    "release_created",
			Message: "Release rendered.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "release",
				ID:    result.ReleaseID.String(),
				Route: "/api/v1/releases/" + result.ReleaseID.String(),
			}},
		},
	})
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *auth.APIError) {
	record, err := s.incidentStore.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *auth.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	for _, role := range roles {
		if record.Role == role {
			return record, nil
		}
	}
	return incidents.MembershipRecord{}, &auth.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": strings.Join(roles, "|")}}
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	return auth.AuthenticateSessionRequest(r, auth.SessionAuthOptions{
		Store:         s.authStore,
		Keys:          s.keys,
		Hub:           s.hub,
		Now:           s.now,
		StateChanging: stateChanging,
	})
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if !auth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
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
