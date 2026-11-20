package reporting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

// RegisterRoutes binds the Snapshot and Reporting routes for central admission
// (Core 01 REQ-01-707): the registry resolves each operation, authenticates by
// its declared security and admits its path parameters before a handler runs.
func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.reporting", map[string]http.HandlerFunc{
			"approveRelease":                   service.handleApproveRelease,
			"createRelease":                    service.handleCreateRelease,
			"createSnapshot":                   service.handleCreateSnapshot,
			"getRelease":                       service.handleGetRelease,
			"getSnapshot":                      service.handleGetSnapshot,
			"getRetainedReportArtifacts":       service.handleRetainedReportArtifacts,
			"getRetainedReportArtifactContent": service.handleRetainedReportArtifactContent,
			"invalidateRelease":                service.handleInvalidateRelease,
			"publishRelease":                   service.handlePublishRelease,
		})
	}
}

func newService(deps httpapi.DependencySet, options RouteOptions) (*Service, error) {
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
		options.Renderer,
		now,
	)
	if err != nil {
		return nil, err
	}
	return &Service{
		app:       app,
		authStore: authn.NewStore(deps.PostgresHandle()),
		now:       now,
	}, nil
}

// admittedPrincipal returns the principal central admission established; its
// absence is a binding error.
func admittedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	if !authenticated {
		return httpauth.Principal{}, internalAPIError(errors.New("reporting route was not admitted centrally"))
	}
	return principal, nil
}

// admittedPathUUID returns a uuid path parameter central admission admitted;
// its absence is a binding error.
func admittedPathUUID(r *http.Request, name string) (uuid.UUID, *httpapi.APIError) {
	value, identified := httpapi.PathUUID(r, name)
	if !identified {
		return uuid.UUID{}, internalAPIError(errors.New("reporting route was not admitted centrally"))
	}
	return value, nil
}

// handleCreateSnapshot decodes the body before visibility: its incident_id
// addresses the incident whose visibility is decided (Core 01 §17.3).
func (s *Service) handleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
}

func (s *Service) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	snapshotID, apiErr := admittedPathUUID(r, "snapshot_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, apiErr := s.app.GetSnapshot(r.Context(), principal.User.ID, snapshotID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-read query check is route handling,
	// after the snapshot's visibility.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

// handleCreateRelease validates the body structurally before visibility,
// because its snapshot_id addresses the snapshot whose visibility is decided;
// selector semantics follow visibility (Core 01 §17.3, POST /api/v1/releases).
func (s *Service) handleCreateRelease(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, job)
}

func (s *Service) handleGetRelease(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	releaseID, apiErr := admittedPathUUID(r, "release_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, apiErr := s.app.GetRelease(r.Context(), principal.User.ID, releaseID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-read query check is route handling,
	// after the release's visibility.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleApproveRelease(w http.ResponseWriter, r *http.Request) {
	s.serveReleaseAction(w, r, s.app.ApproveRelease)
}

func (s *Service) handlePublishRelease(w http.ResponseWriter, r *http.Request) {
	s.serveReleaseAction(w, r, s.app.PublishRelease)
}

func (s *Service) handleInvalidateRelease(w http.ResponseWriter, r *http.Request) {
	s.serveReleaseAction(w, r, s.app.InvalidateRelease)
}

// releaseAction runs one release action. It decides the release's visibility
// and the action's role gate before it reads the request (REQ-01-704).
type releaseAction func(ctx context.Context, actorUserID uuid.UUID, releaseID uuid.UUID, request func() (ReleaseActionRequest, *httpapi.APIError)) (map[string]any, *httpapi.APIError)

func (s *Service) serveReleaseAction(w http.ResponseWriter, r *http.Request, action releaseAction) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	releaseID, apiErr := admittedPathUUID(r, "release_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the query and body checks are route handling, after
	// the release's visibility and role gate.
	request := func() (ReleaseActionRequest, *httpapi.APIError) {
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			return ReleaseActionRequest{}, apiErr
		}
		return DecodeReleaseActionRequest(r.Body)
	}
	payload, apiErr := action(r.Context(), principal.User.ID, releaseID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
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

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
}
