package reporting

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	app       *ApplicationService
	authStore *authn.Store
	keys      authn.MasterKeys
	now       func() time.Time
}

type SourceBoundaryResolver interface {
	ResolveCurrentTx(context.Context, pgx.Tx, sourceboundary.ResolveInput) (sourceboundary.Boundary, error)
}

type RouteOptions struct {
	Renderer             *Renderer
	TemplatePacks        *TemplatePackRegistry
	ReferencePacks       reference_data.Retention
	JobSuccessFinalizer  JobSuccessFinalizer
	RenderExportInvoker  RenderExportInvoker
	ExportFieldProviders []exportprovider.FieldProvider
	SupportRefProvider   exportprovider.SupportReferenceProvider
	GraphSourceProviders []GraphSourceProvider
	SourceBoundary       SourceBoundaryResolver
	jobAdmission         reportingJobTransactions
	jobOperations        reportingJobManager
	jobRunner            reportingJobRunner
}

func WithJobs(options RouteOptions, admission reportingJobTransactions, operations reportingJobManager, runner reportingJobRunner) RouteOptions {
	options.jobAdmission = admission
	options.jobOperations = operations
	options.jobRunner = runner
	return options
}

func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.reporting", map[string]http.HandlerFunc{
			"approveRelease":    service.handleReleasesMember,
			"createRelease":     service.handleReleasesCollection,
			"createSnapshot":    service.handleSnapshotsCollection,
			"getRelease":        service.handleReleasesMember,
			"getSnapshot":       service.handleSnapshotsMember,
			"invalidateRelease": service.handleReleasesMember,
			"publishRelease":    service.handleReleasesMember,
		})
	}
}

func newService(deps httpapi.DependencySet, options RouteOptions) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if options.jobOperations != nil && options.jobAdmission == nil {
		return nil, fmt.Errorf("reporting admitted route requires the Jobs transaction service")
	}
	if options.jobOperations != nil && options.ReferencePacks == nil {
		return nil, fmt.Errorf("reporting admitted route requires Reference Pack retention")
	}
	exportMaterializer, err := newReportingExportMaterializer(options.SourceBoundary, options.SupportRefProvider, options.ExportFieldProviders...)
	if err != nil {
		return nil, fmt.Errorf("compose reporting export materializer: %w", err)
	}
	graphSources, err := NewGraphSourceRegistry(options.GraphSourceProviders...)
	if err != nil {
		return nil, fmt.Errorf("compose Reporting graph source registry: %w", err)
	}
	if options.TemplatePacks == nil {
		options.TemplatePacks, err = NewTemplatePackRegistry()
		if err != nil {
			return nil, err
		}
	}
	store := newStore(deps.Postgres, options.jobAdmission, exportMaterializer, graphSources, options.ReferencePacks)
	store.templatePacks = options.TemplatePacks
	app, err := NewApplicationService(
		store,
		admission.NewChecker(deps.PostgresHandle()),
		options.jobOperations,
		options.jobRunner,
		options.JobSuccessFinalizer,
		options.RenderExportInvoker,
		graphSources,
		options.Renderer,
		now,
	)
	if err != nil {
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
