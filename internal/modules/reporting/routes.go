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
	"github.com/google/uuid"
)

type Service struct {
	app       *ApplicationService
	authStore *authn.Store
	keys      authn.MasterKeys
	now       func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
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
	store := NewStore(deps.Postgres)
	app := NewApplicationService(store, incidents.NewAccess(deps.PostgresHandle()), deps.Jobs, deps.JobRunner, now)
	if err := app.recoverReportingJobs(context.Background()); err != nil {
		return nil, err
	}
	return &Service{
		app:       app,
		authStore: authn.NewStore(deps.PostgresHandle()),
		keys:      keys,
		now:       now,
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
	job, apiErr := s.app.CreateSnapshot(r.Context(), principal.User.ID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
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
	resource, apiErr := s.app.GetSnapshot(r.Context(), principal.User.ID, snapshotID)
	if apiErr != nil {
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
	job, apiErr := s.app.CreateRelease(r.Context(), principal.User.ID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
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
		resource, apiErr := s.app.GetRelease(r.Context(), principal.User.ID, route.ReleaseID)
		if apiErr != nil {
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
	payload, apiErr := s.app.ApproveRelease(r.Context(), principal.User.ID, releaseID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
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
	payload, apiErr := s.app.PublishRelease(r.Context(), principal.User.ID, releaseID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
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
	payload, apiErr := s.app.InvalidateRelease(r.Context(), principal.User.ID, releaseID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
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
	bundle, err := renderReportBundle(contract, request.OutputKind, redaction.Model, redaction.ManifestSHA256, request.ReleaseScope, request.OutputOptions, request.GraphProjectionRefs, request.CompositionJSON, RedactionBundleArtifacts{
		RedactionManifestJSON: redaction.ManifestJSON,
		ProfileViewJSON:       redaction.ProfileViewJSON,
		TokenManifestJSON:     redaction.TokenManifestJSON,
		TokenManifestSHA256:   redaction.TokenManifestSHA256,
		RevealMapJSON:         redaction.RevealMapJSON,
	})
	if errors.Is(err, ErrUndeclaredTemplateBinding) {
		return partial, "undeclared_template_binding", err
	}
	if errors.Is(err, ErrMissingRequiredField) {
		return partial, "missing_required_field", err
	}
	var validationErr *renderValidationError
	if errors.As(err, &validationErr) {
		return partial, validationErr.ReasonCode, err
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
