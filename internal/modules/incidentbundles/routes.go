package incidentbundles

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

type service struct {
	store          *store
	authStore      *authn.Store
	incidentAccess incidents.Access
	storage        BundleStorage
	worker         *incidentBundleWorker
	keys           authn.MasterKeys
	now            func() time.Time
}

// RegisterRoutes returns the Incident Bundles registrar only after the module
// facade has completed its coordinator and worker lifecycle.
func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if err := m.routeDependenciesReady(); err != nil {
			return err
		}
		service, err := newService(deps, m)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.incidentbundles", map[string]http.HandlerFunc{
			"exportIncidentBundle":        service.handleExport,
			"getIncidentBundleDescriptor": service.handleBundleMember,
			"importIncidentBundle":        service.handleImport,
		})
	}
}

func newService(deps httpapi.DependencySet, module *Module) (*service, error) {
	if module == nil || module.pool == nil || module.store == nil || module.worker == nil || module.storage == nil || module.now == nil {
		return nil, fmt.Errorf("incident bundle module is incomplete")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	return &service{
		store:          module.store,
		authStore:      authn.NewStore(module.pool),
		incidentAccess: incidents.NewAccess(module.pool),
		storage:        module.storage,
		worker:         module.worker,
		keys:           keys,
		now:            module.now,
	}, nil
}

func (s *service) handleBundleMember(w http.ResponseWriter, r *http.Request) {
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
	record, err := s.store.getDescriptor(r.Context(), bundleID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if _, err := s.incidentAccess.GetIncidentMembershipForUser(r.Context(), record.IncidentID, principal.User.ID); s.incidentAccess.IsMembershipNotFound(err) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.resource())
}

func (s *service) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeExportRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := s.incidentAccess.GetIncidentMembershipForUser(r.Context(), request.IncidentID, principal.User.ID); s.incidentAccess.IsMembershipNotFound(err) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if request.CapabilityActivationRequested {
		writeAPIError(w, r, extensionCapabilityNotSupported())
		return
	}
	result, err := s.store.acceptExport(r.Context(), exportAcceptedParams{
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
		writeAPIError(w, r, internalAPIError())
		return
	}
	if !result.Replayed {
		_ = s.worker.dispatch(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: acceptedIncidentBundleFileContentTypes()})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := decodeImportMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	stagingReference, err := s.storage.Stage(r.Context(), envelope.FileSHA256Hex, envelope.File)
	if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	result, err := s.store.acceptImport(r.Context(), importAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		UploadedSHA256:    envelope.FileSHA256Hex,
		BundleStagingRef:  stagingReference,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		_ = s.storage.RemoveStaged(stagingReference)
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		_ = s.storage.RemoveStaged(stagingReference)
		writeAPIError(w, r, internalAPIError())
		return
	}
	if !result.Replayed {
		_ = s.worker.dispatch(result.Job.JobID)
	} else {
		_ = s.storage.RemoveStaged(stagingReference)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) requireDeploymentAdmin(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	return principal, nil
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}
