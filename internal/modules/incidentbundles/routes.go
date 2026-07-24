package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type Service struct {
	store          *Store
	authStore      *authn.Store
	incidentAccess incidents.Access
	files          bundleFileStore
	worker         *incidentBundleWorker
	keys           authn.MasterKeys
	now            func() time.Time
}

type RouteOption func(*routeOptions)

type routeOptions struct {
	importFinalizer incidents.IncidentBundleImportFinalizer
	portability     *PortabilityOrchestrator
	transactions    *crossownertransaction.Coordinator
}

func WithPortability(orchestrator *PortabilityOrchestrator, transactions *crossownertransaction.Coordinator) RouteOption {
	return func(options *routeOptions) {
		options.portability = orchestrator
		options.transactions = transactions
	}
}

func WithImportFinalizer(finalizer incidents.IncidentBundleImportFinalizer) RouteOption {
	return func(options *routeOptions) {
		options.importFinalizer = finalizer
	}
}

func RegisterRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	resolved := routeOptions{}
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimedBy(deps.ExtensionEpoch, ProfileID) {
			return nil
		}
		service, err := newService(deps, resolved)
		if err != nil {
			return err
		}
		if err := service.recoverIncidentBundleJobs(context.Background()); err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/incident-bundles/export", service.handleExport)
		mux.HandleFunc("/api/v1/incident-bundles/import", service.handleImport)
		mux.HandleFunc("/api/v1/incident-bundles/", service.handleBundleMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet, options routeOptions) (*Service, error) {
	if options.importFinalizer == nil {
		return nil, fmt.Errorf("incident bundle import finalizer is required")
	}
	if options.portability == nil || options.transactions == nil {
		return nil, fmt.Errorf("incident bundle portability composition is required")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	store := NewStore(deps.Postgres)
	files := newBundleFileStore(deps.Config.Roots.TemporaryWork.Path, deps.Config.Roots.ExportOutputs.Path)
	workerStartHook, err := workerStartHookFromDependencies(deps)
	if err != nil {
		return nil, err
	}
	worker := newIncidentBundleWorker(store, deps, files, options.importFinalizer, options.portability, options.transactions, now, workerStartHook)
	if err := worker.registerJobHandler(); err != nil {
		return nil, err
	}
	return &Service{
		store:          store,
		authStore:      authn.NewStore(deps.PostgresHandle()),
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		files:          files,
		worker:         worker,
		keys:           keys,
		now:            now,
	}, nil
}

func (s *Service) handleBundleMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bundleIDText := strings.TrimPrefix(r.URL.Path, "/api/v1/incident-bundles/")
	if bundleIDText == "" || strings.Contains(bundleIDText, "/") {
		http.NotFound(w, r)
		return
	}
	bundleID, err := uuid.Parse(bundleIDText)
	if err != nil {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.GetDescriptor(r.Context(), bundleID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, err := s.incidentAccess.GetIncidentMembershipForUser(r.Context(), record.IncidentID, principal.User.ID); s.incidentAccess.IsMembershipNotFound(err) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.Resource())
}

func (s *Service) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeExportRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := s.incidentAccess.GetIncidentMembershipForUser(r.Context(), request.IncidentID, principal.User.ID); s.incidentAccess.IsMembershipNotFound(err) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.AcceptExport(r.Context(), ExportAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
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
		s.worker.dispatch(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: IncidentBundleFileContentTypes})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := DecodeImportMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	stagingPath, err := s.files.stageBundle(envelope.FileSHA256Hex, envelope.File)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.AcceptImport(r.Context(), ImportAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		UploadedSHA256:    envelope.FileSHA256Hex,
		BundleStagingPath: stagingPath,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		s.files.remove(stagingPath)
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		s.files.remove(stagingPath)
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.worker.dispatch(result.Job.JobID)
	} else {
		s.files.remove(stagingPath)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) recoverIncidentBundleJobs(ctx context.Context) error {
	return s.worker.recoverJobs(ctx)
}

func (s *Service) requireDeploymentAdmin(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	return principal, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}
